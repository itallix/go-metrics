package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/realip"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/itallix/go-metrics/internal/controller"
	"github.com/itallix/go-metrics/internal/grpc/api"
	pb "github.com/itallix/go-metrics/internal/grpc/proto"
	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/middleware"
	"github.com/itallix/go-metrics/internal/model"
	"github.com/itallix/go-metrics/internal/service"
	"github.com/itallix/go-metrics/internal/storage"
	"github.com/itallix/go-metrics/internal/storage/db"
	"github.com/itallix/go-metrics/internal/storage/memory"

	_ "github.com/jackc/pgx"
	_ "google.golang.org/grpc/encoding/gzip"
)

const (
	ReadTimeoutSeconds     = 5
	WriteTimeoutSeconds    = 10
	IdleTimeoutSeconds     = 15
	ShutdownTimeoutSeconds = 10
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func startGrpcServer(grpcServer *grpc.Server, storage storage.Storage, hasher service.HashService) {
	grpcServerAddr := "localhost:" + model.GRPCPort
	lis, err := net.Listen("tcp", grpcServerAddr)
	if err != nil {
		logger.Log().Fatalf("failed to run gRPC server: %v", err)
	}
	pb.RegisterMetricsServer(grpcServer, api.NewServer(storage, hasher))
	reflection.Register(grpcServer)
	logger.Log().Infof("GRPC server is starting on %s...", grpcServerAddr)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Log().Fatalf("failed to run gRPC server: %v", err)
	}
}

func main() {
	if err := logger.Initialize("debug"); err != nil {
		log.Fatalf("Cannot instantiate zap logger: %s", err)
	}
	defer func() {
		if deferErr := logger.Log().Sync(); deferErr != nil {
			logger.Log().Errorf("Failed to sync logger: %s", deferErr)
		}
	}()

	service.PrintBuildInfo(buildVersion, buildDate, buildCommit, os.Stdout)

	serverConfig, err := parseConfig()
	if err != nil {
		logger.Log().Errorf("Can't parse flags: %v", err.Error())
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LoggerWithZap(logger.Log()))
	if serverConfig.TrustedSubnet != "" {
		router.Use(middleware.CheckIPAddr(serverConfig.TrustedSubnet))
	}
	var hashService service.HashService
	if serverConfig.Key != "" {
		hashService = service.NewHashService(serverConfig.Key)
		router.Use(middleware.VerifyHash(hashService))
	}
	if serverConfig.CryptoKey != "" {
		router.Use(middleware.DecryptMiddleware(serverConfig.CryptoKey))
	}
	router.Use(gzip.Gzip(gzip.BestCompression))
	router.Use(middleware.GzipDecompress())

	ctx, cancel := context.WithCancel(context.Background())
	var (
		mStorage storage.Storage
		wg       sync.WaitGroup
	)
	if serverConfig.DatabaseDSN != "" {
		mStorage, err = db.NewPgStorage(ctx, serverConfig.DatabaseDSN)
		if err != nil {
			logger.Log().Errorf("Cannot instantiate DB: %v", err)
			mStorage = memory.NewMemStorage(ctx, &wg, memory.NewConfig(serverConfig.FilePath, serverConfig.StoreInterval,
				serverConfig.Restore))
		}
	}
	if mStorage == nil {
		mStorage = memory.NewMemStorage(ctx, &wg, memory.NewConfig(serverConfig.FilePath, serverConfig.StoreInterval,
			serverConfig.Restore))
	}
	defer mStorage.Close()
	metricController := controller.NewMetricController(mStorage)

	router.GET("/", metricController.ListMetrics)
	router.POST("/update/", metricController.UpdateOne)
	router.POST("/updates/", metricController.UpdateBatch)
	router.POST("/value/", metricController.GetMetric)
	router.POST("/update/:metricType/:metricName/:metricValue", metricController.UpdateMetricQuery)
	router.GET("/value/:metricType/:metricName", metricController.GetMetricQuery)
	router.GET("/ping", func(c *gin.Context) {
		if mStorage.Ping(c.Request.Context()) {
			c.Status(http.StatusOK)
			return
		}
		_ = c.AbortWithError(http.StatusInternalServerError, errors.New("internal server error"))
	})
	pprof.Register(router)

	server := &http.Server{
		Addr:         serverConfig.Address,
		Handler:      router,
		ReadTimeout:  ReadTimeoutSeconds * time.Second,
		WriteTimeout: WriteTimeoutSeconds * time.Second,
		IdleTimeout:  IdleTimeoutSeconds * time.Second,
	}

	var grpcServer *grpc.Server
	if serverConfig.TrustedSubnet != "" {
		trustedPeers := []netip.Prefix{
			netip.MustParsePrefix(serverConfig.TrustedSubnet),
		}
		opts := []realip.Option{
			realip.WithTrustedPeers(trustedPeers),
			realip.WithHeaders([]string{model.XRealIPHeader}),
		}
		grpcServer = grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				realip.UnaryServerInterceptorOpts(opts...),
			),
		)
	} else {
		grpcServer = grpc.NewServer()
	}
	go startGrpcServer(grpcServer, mStorage, hashService)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		<-quit
		logger.Log().Info("Shutting down servers gracefully...")
		cancel()
		wg.Wait()

		logger.Log().Info("Stopping gRPC server...")
		grpcServer.GracefulStop()

		logger.Log().Info("Stopping HTTP server...")
		ctx, cancelTimeout := context.WithTimeout(context.Background(), ShutdownTimeoutSeconds*time.Second)
		defer cancelTimeout()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatal("Server Shutdown:", err)
		}
	}()

	logger.Log().Infof("Server is starting on %s...", serverConfig.Address)
	if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log().Fatalf("Error starting server: %v", err)
	}
}
