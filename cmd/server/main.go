package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type MemStorage struct {
	metrics map[string]float64
	mu      sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]float64),
	}
}

type Storage interface {
	Set(name string, value float64)
	Get(name string) (float64, bool)
	Delete(name string)
}

func (m *MemStorage) Set(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics[name] = value
}

func (m *MemStorage) Get(name string) (float64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.metrics[name]
	return value, ok
}

func (m *MemStorage) Delete(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.metrics, name)
}

func metricHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if parts[1] != "update" || len(parts) != 5 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	metricType := parts[2]
	metricName := parts[3]
	metricValue := parts[4]

	switch metricType {
	case "counter":
		if _, err := strconv.ParseInt(metricValue, 10, 64); err != nil {
			http.Error(w, "Metric type is not supported", http.StatusBadRequest)
			return
		}
	case "gauge":
		if _, err := strconv.ParseFloat(metricValue, 64); err != nil {
			http.Error(w, "Metric type is not supported", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "Metric is not found", http.StatusNotFound)
		return
	}

	log.Printf("Updating metric '%s' of type '%s' with value '%s'\n", metricName, metricType, metricValue)
	w.WriteHeader(http.StatusOK)
}

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

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", metricHandler)

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
