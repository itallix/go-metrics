package controller_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itallix/go-metrics/internal/service"

	"github.com/itallix/go-metrics/internal/model"

	"github.com/itallix/go-metrics/internal/controller"

	"github.com/gin-gonic/gin"
	"github.com/itallix/go-metrics/internal/storage"

	"github.com/stretchr/testify/assert"
)

func TestMetricHandler_Update(t *testing.T) {
	const requestPath = "/update"
	var (
		floatValue = 123.0
		intValue   = int64(123)
	)

	tests := []struct {
		name        string
		givePayload *model.Metrics
		wantStatus  int
		wantJSON    string
	}{
		{
			name: "MetricDoesntExist",
			givePayload: &model.Metrics{
				ID:    "someMetric",
				MType: "hist",
				Value: &floatValue,
			},
			wantStatus: http.StatusBadRequest,
			wantJSON:   `{"error": "metric is not found"}`,
		},
		{
			name: "CounterMetricTypeIsNotSupported",
			givePayload: &model.Metrics{
				ID:    "someCounter",
				MType: model.Counter,
				Value: &floatValue,
			},
			wantStatus: http.StatusBadRequest,
			wantJSON:   `{"error": "metric type is not supported"}`,
		},
		{
			name: "GaugeMetricTypeIsNotSupported",
			givePayload: &model.Metrics{
				ID:    "someGauge",
				MType: model.Gauge,
				Delta: &intValue,
			},
			wantStatus: http.StatusBadRequest,
			wantJSON:   `{"error": "metric type is not supported"}`,
		},
		{
			name:        "CanUpdateCounter",
			givePayload: model.NewCounter("someCounter", &intValue),
			wantStatus:  http.StatusOK,
			wantJSON:    `{"id": "someCounter", "type": "counter", "delta": 123}`,
		},
		{
			name:        "CanUpdateGauge",
			givePayload: model.NewGauge("someGauge", &floatValue),
			wantStatus:  http.StatusOK,
			wantJSON:    `{"id": "someGauge", "type": "gauge", "value": 123.0}`,
		},
	}

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	metricService := service.NewMetricService(
		storage.NewMemStorage[int64](), storage.NewMemStorage[float64]())
	metricController := controller.NewMetricController(metricService)

	router.POST(requestPath, metricController.UpdateMetric)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			encoder := json.NewEncoder(&buf)
			if err := encoder.Encode(&tt.givePayload); err != nil {
				t.Fatalf("Issue encoding payload to json: %v", err)
				return
			}
			req := httptest.NewRequest(http.MethodPost, requestPath, &buf)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			assert.Equal(t, tt.wantStatus, resp.Code, "handler returned wrong status code")
			assert.JSONEq(t, tt.wantJSON, resp.Body.String(), "handler returned wrong message")
		})
	}
}

func TestMetricHandler_Value(t *testing.T) {
	const requestPath = "/value"
	tests := []struct {
		name        string
		givePayload *model.Metrics
		wantStatus  int
		wantJSON    string
	}{
		{
			name: "MetricNotFound",
			givePayload: &model.Metrics{
				ID:    "someMetric",
				MType: "hist",
			},
			wantStatus: http.StatusNotFound,
			wantJSON:   `{"error": "metric is not found"}`,
		},
		{
			name:        "CounterNotFound",
			givePayload: model.NewCounter("counter1", nil),
			wantStatus:  http.StatusNotFound,
			wantJSON:    `{"error": "metric is not found"}`,
		},
		{
			name:        "GaugeNotFound",
			givePayload: model.NewGauge("gauge1", nil),
			wantStatus:  http.StatusNotFound,
			wantJSON:    `{"error": "metric is not found"}`,
		},
		{
			name:        "CanGetCounter",
			givePayload: model.NewCounter("counter0", nil),
			wantStatus:  http.StatusOK,
			wantJSON:    `{"id": "counter0", "type": "counter", "delta": 10}`,
		},
		{
			name:        "CanGetGauge",
			givePayload: model.NewGauge("gauge0", nil),
			wantStatus:  http.StatusOK,
			wantJSON:    `{"id": "gauge0", "type": "gauge", "value": 25.0}`,
		},
	}

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	counters := storage.NewMemStorage[int64]()
	counters.Update("counter0", 5)
	counters.Update("counter0", 5)
	gauges := storage.NewMemStorage[float64]()
	gauges.Set("gauge0", 25.0)
	metricService := service.NewMetricService(counters, gauges)
	metricController := controller.NewMetricController(metricService)

	router.POST(requestPath, metricController.GetMetric)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			encoder := json.NewEncoder(&buf)
			if err := encoder.Encode(&tt.givePayload); err != nil {
				t.Fatalf("Issue encoding payload to json: %v", err)
				return
			}
			req := httptest.NewRequest(http.MethodPost, requestPath, &buf)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			assert.Equal(t, tt.wantStatus, resp.Code, "handler returned wrong status code")
			assert.JSONEq(t, tt.wantJSON, resp.Body.String(), "handler returned wrong message")
		})
	}
}
