package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipMiddleware_OK(t *testing.T) {
	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	require.NoError(t, err)
	_, err = gz.Write([]byte("test message"))
	require.NoError(t, err)
	require.NoError(t, gz.Close())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(GzipDecompress())

	r.GET("/gzip", func(c *gin.Context) {
		read, _ := io.ReadAll(c.Request.Body)
		c.String(200, string(read))
	})

	req := httptest.NewRequest(http.MethodGet, "/gzip", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test message", w.Body.String())
}
