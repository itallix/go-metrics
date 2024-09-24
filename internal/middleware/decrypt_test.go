package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/itallix/go-metrics/internal/service"
)

func TestDecryptMiddleware_OK(t *testing.T) {
	message, err := service.EncryptData([]byte("secret"), "../../test_data/client.pem")
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(DecryptMiddleware("../../test_data/server.pem"))

	r.GET("/test", func(c *gin.Context) {
		read, _ := io.ReadAll(c.Request.Body)
		c.String(200, string(read))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", bytes.NewReader(message))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "secret", w.Body.String())
}
