package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/itallix/go-metrics/internal/model"
	"github.com/itallix/go-metrics/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestHashMiddleware_OK(t *testing.T) {
	srv := service.NewHashService("secret")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(VerifyHash(srv))

	r.GET("/test", func(c *gin.Context) {
		c.String(200, "Hello World")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", strings.NewReader("Body text"))
	req.Header.Set(model.HashSha256Header, srv.Sha256sum([]byte("Body text")))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.NotNil(t, w.Header().Get(model.HashSha256Header))
	assert.Equal(t, w.Code, http.StatusOK)
}

func TestHashMiddleware_BadRequest(t *testing.T) {
	srv := service.NewHashService("secret")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(VerifyHash(srv))

	r.GET("/test", func(c *gin.Context) {
		c.String(200, "Hello World")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", strings.NewReader("Body hey"))
	req.Header.Set(model.HashSha256Header, srv.Sha256sum([]byte("Body text")))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.NotNil(t, w.Header().Get(model.HashSha256Header))
	assert.Equal(t, w.Code, http.StatusBadRequest)
}
