package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/itallix/go-metrics/internal/model"
)

func TestCheckIPAddrMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		givenIP    string
		wantStatus int
	}{
		{
			name:       "OK",
			givenIP:    "192.168.2.60",
			wantStatus: http.StatusOK,
		},
		{
			name:       "OK",
			givenIP:    "192.168.1.60",
			wantStatus: http.StatusForbidden,
		},
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CheckIPAddr("192.168.2.0/24"))

	r.GET("/test", func(c *gin.Context) {
		c.String(200, "Hello World")
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", strings.NewReader("Body hey"))
			req.Header.Set(model.XRealIPHeader, tt.givenIP)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
