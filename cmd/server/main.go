package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/itallix/go-metrics/internal/storage"
	"github.com/itallix/go-metrics/internal/storage/db"
	"github.com/itallix/go-metrics/internal/storage/memory"

	"github.com/gin-contrib/gzip"
	_ "github.com/jackc/pgx"

	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/itallix/go-metrics/internal/controller"
)

const (
	ReadTimeoutSeconds  = 5
	WriteTimeoutSeconds = 10
	IdleTimeoutSeconds  = 15
)

func main() {
	if err := logger.Initialize("debug"); err != nil {
		log.Fatalf("Cannot instantiate zap logger: %s", err)
	}
	defer func() {
		if deferErr := logger.Log().Sync(); deferErr != nil {
			logger.Log().Errorf("Failed to sync logger: %s", deferErr)
		}
	}()

	addr, serverConfig, err := parseFlags()
	if err != nil {
		logger.Log().Errorf("Can't parse flags: %v", err.Error())
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LoggerWithZap(logger.Log()))
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(middleware.GzipDecompress())

	ctx := context.Background()
	var mStorage storage.Storage
	if serverConfig.DatabaseDSN != "" {
		mStorage, err = db.NewPgStorage(ctx, serverConfig.DatabaseDSN)
		if err != nil {
			logger.Log().Errorf("Cannot instantiate DB: %v", err)
			mStorage = memory.NewMemStorage(ctx, memory.NewConfig(serverConfig.FilePath, serverConfig.StoreInterval,
				serverConfig.Restore))
		}
	}
	if mStorage == nil {
		mStorage = memory.NewMemStorage(ctx, memory.NewConfig(serverConfig.FilePath, serverConfig.StoreInterval,
			serverConfig.Restore))
	}
	defer mStorage.Close()
	metricController := controller.NewMetricController(mStorage)

	router.GET("/", metricController.ListMetrics)
	router.POST("/update", metricController.UpdateOne)
	router.POST("/updates", metricController.UpdateBatch)
	router.POST("/value", metricController.GetMetric)
	router.POST("/update/:metricType/:metricName/:metricValue", metricController.UpdateMetricQuery)
	router.GET("/value/:metricType/:metricName", metricController.GetMetricQuery)
	router.GET("/healthcheck", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/ping", func(c *gin.Context) {
		if mStorage.Ping(c.Request.Context()) {
			c.Status(http.StatusOK)
			return
		}
		_ = c.AbortWithError(http.StatusInternalServerError, errors.New("internal server error"))
	})

	server := &http.Server{
		Addr:         addr.String(),
		Handler:      router,
		ReadTimeout:  ReadTimeoutSeconds * time.Second,
		WriteTimeout: WriteTimeoutSeconds * time.Second,
		IdleTimeout:  IdleTimeoutSeconds * time.Second,
	}

	logger.Log().Infof("Server is starting on %s...", addr)
	if err = server.ListenAndServe(); err != nil {
		logger.Log().Fatalf("Error starting server: %v", err)
	}
}
