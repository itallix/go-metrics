package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itallix/go-metrics/internal/controller"
	"github.com/itallix/go-metrics/internal/storage"

	"go.uber.org/zap"
)

const (
	ReadTimeoutSeconds  = 5
	WriteTimeoutSeconds = 10
	IdleTimeoutSeconds  = 15
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Cannot instantiate zap logger: %s", err)
	}
	defer func() {
		if deferErr := logger.Sync(); deferErr != nil {
			logger.Error("Failed to sync logger", zap.Error(deferErr))
		}
	}()

	addr := parseFlags()

	router := gin.New()
	router.Use(gin.Recovery())

	metricController := controller.NewMetricController(
		storage.NewMemStorage[int](), storage.NewMemStorage[float64]())

	router.GET("/", metricController.ListMetrics)
	router.POST("/update/:metricType/:metricName/:metricValue", metricController.UpdateMetric)
	router.GET("/value/:metricType/:metricName", metricController.GetMetric)

	server := &http.Server{
		Addr:         addr.String(),
		Handler:      router,
		ReadTimeout:  ReadTimeoutSeconds * time.Second,
		WriteTimeout: WriteTimeoutSeconds * time.Second,
		IdleTimeout:  IdleTimeoutSeconds * time.Second,
	}

	logger.Info(fmt.Sprintf("Server is starting on %v...", addr))
	if err = server.ListenAndServe(); err != nil {
		logger.Fatal("Error starting server", zap.Error(err))
	}
}
