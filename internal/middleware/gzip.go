package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GzipDecompress() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			reader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				_ = c.AbortWithError(http.StatusBadRequest, err)
				return
			}
			c.Request.Body = io.NopCloser(reader)
		}

		c.Next()
	}
}
