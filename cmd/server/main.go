package main

import (
	"log"
	"net/http"
	"time"

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
	router.Use(middleware.LoggerWithZap())

	metricController := controller.NewMetricController(
		storage.NewMemStorage[int](), storage.NewMemStorage[float64]())

	router.GET("/", metricController.ListMetrics)
	router.POST("/update/:metricType/:metricName/:metricValue", metricController.UpdateMetric)
	router.GET("/value/:metricType/:metricName", metricController.GetMetric)
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
