package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader 客户端传入的请求 ID header
	RequestIDHeader = "X-Request-Id"
	// ContextKeyRequestID Gin context 中的 key
	ContextKeyRequestID = "request_id"
)

// RequestID 请求 ID 中间件
// - 从 X-Request-Id 请求头提取
// - 若客户端未传，自动生成 UUID
// - 存入 Gin context，写入响应头
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set(ContextKeyRequestID, requestID)
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}
