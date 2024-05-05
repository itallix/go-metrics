package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricHandler(t *testing.T) {
	tests := []struct {
		name        string
		givePath    string
		giveMethod  string
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "MethodIsNotSupported",
			givePath:    UpdatePath,
			giveMethod:  http.MethodGet,
			wantStatus:  http.StatusMethodNotAllowed,
			wantMessage: "Method is not supported\n",
		},
		{
			name:        "PathDoesntExist",
			givePath:    "/test/",
			giveMethod:  http.MethodPost,
			wantStatus:  http.StatusNotFound,
			wantMessage: "Invalid URL format\n",
		},
		{
			name:        "CounterMetricTypeIsNotSupported",
			givePath:    UpdatePath + "counter/someMetric/123.0",
			giveMethod:  http.MethodPost,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "Metric type is not supported\n",
		},
		{
			name:        "GaugeMetricTypeIsNotSupported",
			givePath:    UpdatePath + "gauge/someMetric/abc",
			giveMethod:  http.MethodPost,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "Metric type is not supported\n",
		},
		{
			name:        "CanUpdateCounter",
			givePath:    UpdatePath + "counter/someMetric/123",
			giveMethod:  http.MethodPost,
			wantStatus:  http.StatusOK,
			wantMessage: "",
		},
		{
			name:        "CanUpdateGauge",
			givePath:    UpdatePath + "gauge/someMetric/123.0",
			giveMethod:  http.MethodPost,
			wantStatus:  http.StatusOK,
			wantMessage: "",
		},
	}

	handler := http.HandlerFunc(metricHandler)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.giveMethod, tt.givePath, nil)
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)
			assert.Equal(t, tt.wantStatus, resp.Code, "handler returned wrong status code")
			assert.Equal(t, tt.wantMessage, resp.Body.String(), "handler returned wrong message")
		})
	}
}
