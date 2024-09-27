package middleware

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/itallix/go-metrics/internal/logger"
	"github.com/itallix/go-metrics/internal/model"
)

func CheckIPAddr(trustedSubnet string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIPStr := c.GetHeader(model.XRealIPHeader)
		clientIP := net.ParseIP(clientIPStr)
		if clientIP == nil {
			logger.Log().Errorf("Error converting client IP addr from: %s", clientIPStr)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		_, ipNet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			logger.Log().Errorf("Error parsing CIDR address: %v", err)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		if !ipNet.Contains(clientIP) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}
