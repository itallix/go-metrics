package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/model"
	"github.com/itallix/go-metrics/internal/service"
)

func VerifyHash(hashSrv service.HashService) gin.HandlerFunc {
	return func(c *gin.Context) {
		hashSha256 := c.GetHeader(model.HashSha256Header)
		if hashSha256 != "" {
			b, err := io.ReadAll(c.Request.Body)
			if err != nil {
				logger.Log().Errorf("Cannot read data from the request body: %s", err)
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(b))
			if !hashSrv.Matches(b, hashSha256) {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			c.Header(model.HashSha256Header, hashSha256)
		}

		c.Next()
	}
}
