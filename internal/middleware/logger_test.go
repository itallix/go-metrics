package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggerWithZapMiddleware(t *testing.T) {
	core, recorded := observer.New(zap.InfoLevel)
	mockLogger := zap.New(core).Sugar()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LoggerWithZap(mockLogger))

	r.GET("/test", func(c *gin.Context) {
		c.String(200, "Hello World")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if recorded.Len() != 2 {
		t.Errorf("Expected 2 log entries, got %d", recorded.Len())
	}

	incomingEntry := recorded.All()[0]
	assert.Equal(t, incomingEntry.ContextMap()["uri"], req.URL.RequestURI())
	assert.Equal(t, incomingEntry.ContextMap()["method"], req.Method)
	outEntry := recorded.All()[1]
	assert.NotNil(t, outEntry.ContextMap()["latency"])
	assert.Equal(t, outEntry.ContextMap()["status"], int64(w.Code))
	assert.Equal(t, outEntry.ContextMap()["responseSize"], int64(w.Body.Len()))
}
