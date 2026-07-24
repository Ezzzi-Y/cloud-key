package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// maxBodyLogSize 请求/响应体最大记录字节数，超出截断
const maxBodyLogSize = 4096

// bodyLogWriter 包装 gin.ResponseWriter，捕获响应体
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// RequestLogger 请求日志中间件，通过 zap 记录所有 HTTP 请求
//
// 所有请求记录 method、path、status、latency、client_ip；
// 4xx/5xx 额外记录请求体和响应体，便于排查。
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过健康检查等静态资源
		path := c.Request.URL.Path
		if path == "/swagger" || len(path) > 4 && path[:5] == "/swag" {
			c.Next()
			return
		}

		start := time.Now()

		// 读取并恢复请求体（POST/PUT/PATCH）
		var reqBody []byte
		if c.Request.Body != nil && c.Request.Method != "GET" && c.Request.Method != "DELETE" {
			reqBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		// 包装 ResponseWriter 捕获响应体
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 执行后续处理器
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		// 基础字段
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("client_ip", c.ClientIP()),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.Int("body_size", c.Writer.Size()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		// 附加身份信息（如已通过认证中间件）
		if uid, exists := c.Get("user_id"); exists {
			fields = append(fields, zap.Any("user_id", uid))
		}
		if tid, exists := c.Get("tenant_id"); exists {
			fields = append(fields, zap.Any("tenant_id", tid))
		}

		if status >= 400 {
			// 错误请求：记录请求体和响应体，便于排查
			if len(reqBody) > 0 {
				truncated := truncateBody(reqBody)
				fields = append(fields, zap.String("request_body", truncated))
			}
			respBody := blw.body.String()
			if len(respBody) > 0 {
				fields = append(fields, zap.String("response_body", truncateBody([]byte(respBody))))
			}
			if status >= 500 {
				zap.L().Error("HTTP Request", fields...)
			} else {
				zap.L().Warn("HTTP Request", fields...)
			}
		} else {
			zap.L().Info("HTTP Request", fields...)
		}
	}
}

// truncateBody 截断过大的 body，避免撑爆日志
func truncateBody(b []byte) string {
	if len(b) <= maxBodyLogSize {
		return string(b)
	}
	return string(b[:maxBodyLogSize]) + "...(truncated)"
}
