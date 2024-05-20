package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/gzip"

	"github.com/itallix/go-metrics/internal/service"

	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/itallix/go-metrics/internal/controller"
	"github.com/itallix/go-metrics/internal/storage"
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

	addr, err := parseFlags()
	if err != nil {
		logger.Log().Errorf("Can't parse flags: %v", err.Error())
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LoggerWithZap(logger.Log()))
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(middleware.GzipDecompress())

	metricService := service.NewMetricService(
		storage.NewMemStorage[int64](), storage.NewMemStorage[float64]())
	metricController := controller.NewMetricController(
		metricService)

	router.GET("/", metricController.ListMetrics)
	router.POST("/update", metricController.UpdateMetric)
	router.POST("/value", metricController.GetMetric)
	router.POST("/update/:metricType/:metricName/:metricValue", metricController.UpdateMetricQuery)
	router.GET("/value/:metricType/:metricName", metricController.GetMetricQuery)
	router.GET("/healthcheck", func(c *gin.Context) {
		c.Status(http.StatusOK)
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
