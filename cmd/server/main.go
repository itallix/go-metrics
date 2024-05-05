package main

import (
	"log"
	"net/http"
	"time"

	"github.com/itallix/go-metrics/internal/handler"

	"go.uber.org/zap"
)

const (
	ReadTimeoutSeconds  = 5
	WriteTimeoutSeconds = 10
	IdleTimeoutSeconds  = 15
	UpdatePath          = "/update/"
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

	mux := http.NewServeMux()
	mux.HandleFunc(UpdatePath, handler.MetricHandler)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  ReadTimeoutSeconds * time.Second,
		WriteTimeout: WriteTimeoutSeconds * time.Second,
		IdleTimeout:  IdleTimeoutSeconds * time.Second,
	}

	logger.Info("Server is starting on port 8080...")
	if err = server.ListenAndServe(); err != nil {
		logger.Fatal("Error starting server", zap.Error(err))
	}
}
