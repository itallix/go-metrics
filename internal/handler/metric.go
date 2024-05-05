package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"
)

func MetricHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if parts[1] != "update" || len(parts) != 5 {
		http.Error(w, "Invalid URL format", http.StatusNotFound)
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
		http.Error(w, "Metric is not found", http.StatusBadRequest)
		return
	}

	log.Printf("Updating metric '%s' of type '%s' with value '%s'\n", metricName, metricType, metricValue)
	w.WriteHeader(http.StatusOK)
}
