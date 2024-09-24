package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/itallix/go-metrics/internal/controller"
	"github.com/itallix/go-metrics/internal/model"
	"github.com/itallix/go-metrics/internal/storage/memory"
)

func TestMetricHandler_UpdateOne(t *testing.T) {
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
	router := gin.New()
	metricStorage := memory.NewMemStorage(context.Background(), nil, nil)
	metricController := controller.NewMetricController(metricStorage)

	router.POST(requestPath, metricController.UpdateOne)

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

func TestMetricHandler_UpdateBatch(t *testing.T) {
	const requestPath = "/updates"
	var (
		floatValue = 123.0
		intValue   = int64(123)
	)

	tests := []struct {
		name        string
		givePayload []model.Metrics
		wantStatus  int
		wantJSON    string
	}{
		{
			name: "CanUpdateBatch",
			givePayload: []model.Metrics{
				*model.NewGauge("someGauge", &floatValue),
				*model.NewCounter("someCounter", &intValue),
			},
			wantStatus: http.StatusOK,
			wantJSON: `[
				{"id": "someGauge", "type": "gauge", "value": 123.0}, 
				{"id":"someCounter", "type":"counter", "delta":123}]
			`,
		},
		{
			name: "UpdateBatch_400",
			givePayload: []model.Metrics{
				{
					ID:    "broken",
					MType: model.Counter,
					Value: nil,
				},
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	metricStorage := memory.NewMemStorage(context.Background(), nil, nil)
	metricController := controller.NewMetricController(metricStorage)

	router.POST(requestPath, metricController.UpdateBatch)

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
			if tt.wantJSON != "" {
				assert.JSONEq(t, tt.wantJSON, resp.Body.String(), "handler returned wrong message")
			}
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
	router := gin.New()
	ctx := context.Background()
	metricStorage := memory.NewMemStorage(ctx, nil, nil)
	var (
		counter int64 = 10
		gauge         = 25.0
	)
	_ = metricStorage.Update(ctx, model.NewCounter("counter0", &counter))
	_ = metricStorage.Update(ctx, model.NewGauge("gauge0", &gauge))
	metricController := controller.NewMetricController(metricStorage)

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

func TestMetricHandler_GetMetricQuery(t *testing.T) {
	const requestPath = "/value"
	tests := []struct {
		name        string
		queryParams string
		wantStatus  int
		wantResp    string
	}{
		{
			name:        "MetricNotFound",
			queryParams: "/someMetric/hist",
			wantStatus:  http.StatusNotFound,
			wantResp:    `{"error": "metric is not found"}`,
		},
		{
			name:        "CounterNotFound",
			queryParams: "/counter1/nil",
			wantStatus:  http.StatusNotFound,
			wantResp:    `{"error": "metric is not found"}`,
		},
		{
			name:        "GaugeNotFound",
			queryParams: "/gauge1/nil",
			wantStatus:  http.StatusNotFound,
			wantResp:    `{"error": "metric is not found"}`,
		},
		{
			name:        "CanGetCounter",
			queryParams: "/counter/counter0",
			wantStatus:  http.StatusOK,
			wantResp:    "10",
		},
		{
			name:        "CanGetGauge",
			queryParams: "/gauge/gauge0",
			wantStatus:  http.StatusOK,
			wantResp:    "25.0",
		},
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	ctx := context.Background()
	metricStorage := memory.NewMemStorage(ctx, nil, nil)
	var (
		counter int64 = 10
		gauge         = 25.0
	)
	_ = metricStorage.Update(ctx, model.NewCounter("counter0", &counter))
	_ = metricStorage.Update(ctx, model.NewGauge("gauge0", &gauge))
	metricController := controller.NewMetricController(metricStorage)

	router.POST(requestPath+"/:metricType/:metricName", metricController.GetMetricQuery)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, requestPath+tt.queryParams, nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			assert.Equal(t, tt.wantStatus, resp.Code, "handler returned wrong status code")
			assert.JSONEq(t, tt.wantResp, resp.Body.String(), "handler returned wrong message")
		})
	}
}

func TestMetricHandler_UpdateMetricQuery(t *testing.T) {
	const requestPath = "/update"
	var (
		floatValue = 123.0
		intValue   = int64(123)
	)

	tests := []struct {
		name        string
		queryParams string
		wantStatus  int
		wantResp    string
	}{
		{
			name:        "CounterMetricIsNotFound",
			queryParams: fmt.Sprintf("/unknown/someCounter/%d", intValue),
			wantStatus:  http.StatusBadRequest,
			wantResp:    `{"error": "metric is not found"}`,
		},
		{
			name:        "GaugeMetricIsNotFound",
			queryParams: fmt.Sprintf("/unknown/someGauge/%g", floatValue),
			wantStatus:  http.StatusBadRequest,
			wantResp:    `{"error": "metric is not found"}`,
		},
		{
			name:        "CanUpdateCounter",
			queryParams: fmt.Sprintf("/%s/someCounter/%d", model.Counter, intValue),
			wantStatus:  http.StatusOK,
			wantResp:    `{"message": "OK"}`,
		},
		{
			name:        "CanUpdateGauge",
			queryParams: fmt.Sprintf("/%s/someGauge/%g", model.Gauge, floatValue),
			wantStatus:  http.StatusOK,
			wantResp:    `{"message": "OK"}`,
		},
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	metricStorage := memory.NewMemStorage(context.Background(), nil, nil)
	metricController := controller.NewMetricController(metricStorage)

	router.POST(requestPath+"/:metricType/:metricName/:metricValue", metricController.UpdateMetricQuery)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, requestPath+tt.queryParams, nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			assert.Equal(t, tt.wantStatus, resp.Code, "handler returned wrong status code")
			assert.JSONEq(t, tt.wantResp, resp.Body.String(), "handler returned wrong message")
		})
	}
}

func TestMetricHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	ctx := context.Background()
	metricStorage := memory.NewMemStorage(ctx, nil, nil)
	var (
		counter int64 = 10
		gauge         = 25.0
	)
	_ = metricStorage.Update(ctx, model.NewCounter("counter0", &counter))
	_ = metricStorage.Update(ctx, model.NewGauge("gauge0", &gauge))
	metricController := controller.NewMetricController(metricStorage)

	router.GET("/", metricController.ListMetrics)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "handler returned wrong status code")
	wantHTML := `
<html>
<head>
	<title>Metrics</title>
</head>
<body>
	<ul>
		<li>counter0: 10</li>
		<li>gauge0: 25</li>
	</ul>
</body>
</html>`
	assert.Equal(t, wantHTML, resp.Body.String(), "handler returned wrong message")
}
