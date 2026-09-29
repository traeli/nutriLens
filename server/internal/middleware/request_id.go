package middleware

import (
	"strings"

	"shijibu/internal/platform/httpx"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if requestID == "" || len(requestID) > 128 {
			requestID = "req_" + uuid.NewString()
		}
		c.Set(httpx.RequestIDKey, requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}
