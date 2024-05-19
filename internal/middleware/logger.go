package middleware

import (
	"time"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

func LoggerWithZap(logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Infow("Incoming request",
			"uri", c.Request.RequestURI,
			"method", c.Request.Method,
		)

		startTime := time.Now()
		c.Next()
		latency := time.Since(startTime)
		statusCode := c.Writer.Status()
		responseSize := c.Writer.Size()

		logger.Infow("Sending response",
			"latency", latency,
			"status", statusCode,
			"responseSize", responseSize,
		)
	}
}
