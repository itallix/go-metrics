package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
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

	fmt.Printf("Updating metric '%s' of type '%s' with value '%s'\n", metricName, metricType, metricValue)
	w.WriteHeader(http.StatusOK)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", metricHandler)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal("Cannot start server")
	}
	fmt.Println("Server has started on port 8080...")
}
