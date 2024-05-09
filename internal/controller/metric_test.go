package controller_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itallix/go-metrics/internal/controller"

	"github.com/gin-gonic/gin"
	"github.com/itallix/go-metrics/internal/storage"

	"github.com/stretchr/testify/assert"
)

func TestMetricHandler_Update(t *testing.T) {
	tests := []struct {
		name        string
		givePath    string
		giveMethod  string
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "MetricDoesntExist",
			givePath:    "/update/histogram/someMetric/123.0",
			giveMethod:  http.MethodPost,
			wantStatus:  http.StatusBadRequest,
			wantMessage: `{"error": "metric is not found"}`,
		},
		{
			name:        "CounterMetricTypeIsNotSupported",
			givePath:    "/update/counter/someMetric/123.0",
			giveMethod:  http.MethodPost,
			wantStatus:  http.StatusBadRequest,
			wantMessage: `{"error": "metric type is not supported"}`,
		},
		{
			name:        "GaugeMetricTypeIsNotSupported",
			givePath:    "/update/gauge/someMetric/abc",
			giveMethod:  http.MethodPost,
			wantStatus:  http.StatusBadRequest,
			wantMessage: `{"error": "metric type is not supported"}`,
		},
		{
			name:        "CanUpdateCounter",
			givePath:    "/update/counter/someMetric/123",
			giveMethod:  http.MethodPost,
			wantStatus:  http.StatusOK,
			wantMessage: `{"message": "OK"}`,
		},
		{
			name:        "CanUpdateGauge",
			givePath:    "/update/gauge/someMetric/123.0",
			giveMethod:  http.MethodPost,
			wantStatus:  http.StatusOK,
			wantMessage: `{"message": "OK"}`,
		},
	}

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	metricController := controller.NewMetricController(
		storage.NewMemStorage[int](), storage.NewMemStorage[float64]()) // FIXME: introduce mock

	router.POST("/update/:metricType/:metricName/:metricValue", metricController.UpdateMetric)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.giveMethod, tt.givePath, nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			assert.Equal(t, tt.wantStatus, resp.Code, "handler returned wrong status code")
			assert.JSONEq(t, tt.wantMessage, resp.Body.String(), "handler returned wrong message")
		})
	}
}
