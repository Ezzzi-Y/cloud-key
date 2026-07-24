package middleware

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/handler"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const DefaultRateLimitWindow = 1 * time.Minute
const DefaultRateLimitMaxRequests = 5

var rateLimitScript = redis.NewScript(`
local key = KEYS[1]
local window = tonumber(ARGV[1])
local maxRequests = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local uniqueId = ARGV[4]

redis.call('ZREMRANGEBYSCORE', key, 0, now - window)
local count = redis.call('ZCARD', key)

if count >= maxRequests then
	return 0
end

redis.call('ZADD', key, now, uniqueId)
redis.call('EXPIRE', key, math.ceil(window / 1000000000))
return 1
`)

// RateLimitMiddleware returns a gin.HandlerFunc that limits requests per IP
// using a Redis ZSet-based sliding window.
// window is the time window in nanoseconds, maxRequests is the maximum number of requests allowed.
func RateLimitMiddleware(rdb *redis.Client, window time.Duration, maxRequests int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("rl:ip:%s", ip)
		now := time.Now().UnixNano()
		uniqueId := fmt.Sprintf("%d:%s", now, ip)

		ctx := context.Background()
		allowed, err := rateLimitScript.Run(ctx, rdb, []string{key},
			window.Nanoseconds(), maxRequests, now, uniqueId).Int()
		if err != nil {
			// On Redis error, fail open to avoid blocking all users
			c.Next()
			return
		}

		if allowed == 0 {
			c.JSON(http.StatusTooManyRequests, handler.Response{
				Code:    errcode.CodeRateLimit,
				Message: errcode.GetMessage(errcode.CodeRateLimit),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
