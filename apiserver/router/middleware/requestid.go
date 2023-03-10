package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/xnile/muxwaf/pkg/utils"
)

// RequestID 透传Request-ID，如果没有则生成一个
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for incoming header, use it if exists
		requestID := c.Request.Header.Get("X-Request-ID")

		// Create request id with UUID4
		if requestID == "" {
			requestID = utils.GenUUID()
		}

		// Expose it for use in the application
		c.Set("X-Request-ID", requestID)

		// Set X-Request-ID header
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}
