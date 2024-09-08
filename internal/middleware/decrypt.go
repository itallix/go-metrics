package middleware

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/service"
)

func DecryptMiddleware(privateKeyPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		encryptedData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, errors.New("failed to read request body"))
			return
		}

		if len(encryptedData) > 0 {
			decryptedData, err := service.DecryptData(encryptedData, privateKeyPath)
			if err != nil {
				logger.Log().Errorf("Error decrypting the request payload %v", err)
				c.AbortWithError(http.StatusInternalServerError, errors.New("failed to decrypt data"))
				return
			}

			c.Request.Body = io.NopCloser(bytes.NewBuffer(decryptedData))
		}

		c.Next()
	}
}
