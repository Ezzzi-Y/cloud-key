package handler

import (
	"CloudKey/internal/errcode"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	// idempotentTTL 幂等缓存过期时间
	idempotentTTL = 24 * time.Hour
)

// idempotentKey 返回 Redis 幂等缓存 key
func idempotentKey(requestID string) string {
	return "idemp:" + requestID
}

// idempotentResult 缓存的响应结构
type idempotentResult struct {
	HTTPStatus int         `json:"s"` // HTTP 状态码
	Code       int         `json:"c"` // 业务 code
	Data       interface{} `json:"d"` // 响应数据
}

// idempotentClaimScript 原子 claim：SETNX + EXPIRE
// 返回 1 表示首次请求（已加锁），0 表示重复请求
var idempotentClaimScript = redis.NewScript(`
local key = KEYS[1]
local ttl = tonumber(ARGV[1])
local ok = redis.call('SETNX', key, '__claimed__')
if ok == 1 then
    redis.call('EXPIRE', key, ttl)
    return 1
end
return 0
`)

// IdempotentCheck 检查幂等性并尝试 claim
// 返回 (claimed, cachedResponse):
//   - claimed=true: 首次请求，调用方应继续执行业务逻辑
//   - claimed=false: 重复请求，cachedResponse 已写入响应（若非 nil），调用方应直接 return
func IdempotentCheck(c *gin.Context, rdb *redis.Client, requestID string) (claimed bool, cached *idempotentResult) {
	if rdb == nil || requestID == "" {
		return true, nil
	}

	ctx := context.Background()
	key := idempotentKey(requestID)

	// 尝试 claim
	result, err := idempotentClaimScript.Run(ctx, rdb, []string{key}, int(idempotentTTL.Seconds())).Int()
	if err != nil {
		zap.L().Error("idempotent claim failed", zap.Error(err))
		// Redis 异常时放行，不阻塞业务
		return true, nil
	}

	if result == 1 {
		// 首次请求，已 claim
		return true, nil
	}

	// 重复请求，读取缓存结果
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		zap.L().Error("idempotent get cached result failed", zap.Error(err))
		// 无法读取缓存，放行（宁可重复执行也不丢请求）
		return true, nil
	}

	// 缓存值仍为 __claimed__ 说明前一个请求还在处理中（并发竞态）
	if val == "__claimed__" {
		c.JSON(http.StatusConflict, Response{
			Code:    errcode.CodeIdempotentReplay,
			Message: "请求正在处理中，请稍后重试",
			Data:    nil,
		})
		return false, nil
	}

	// 解析缓存结果
	var cachedResult idempotentResult
	if err := json.Unmarshal([]byte(val), &cachedResult); err != nil {
		zap.L().Error("idempotent unmarshal cached result failed", zap.Error(err))
		return true, nil
	}

	// 返回缓存响应
	c.JSON(cachedResult.HTTPStatus, Response{
		Code:    cachedResult.Code,
		Message: "success",
		Data:    cachedResult.Data,
	})
	return false, &cachedResult
}

// CacheIdempotentResult 缓存幂等请求的结果
func CacheIdempotentResult(rdb *redis.Client, requestID string, httpStatus int, code int, data interface{}) {
	if rdb == nil || requestID == "" {
		return
	}

	ctx := context.Background()
	key := idempotentKey(requestID)

	result := idempotentResult{
		HTTPStatus: httpStatus,
		Code:       code,
		Data:       data,
	}

	body, err := json.Marshal(result)
	if err != nil {
		zap.L().Error("idempotent marshal result failed", zap.Error(err))
		return
	}

	// SET 覆盖 __claimed__ 占位，保留原有 TTL
	if err := rdb.Set(ctx, key, string(body), idempotentTTL).Err(); err != nil {
		zap.L().Error("idempotent cache result failed", zap.Error(err))
	}
}

// CacheIdempotentError 缓存幂等请求的错误结果（业务错误也应缓存，防止重试时重复执行）
func CacheIdempotentError(rdb *redis.Client, requestID string, httpStatus int, code int) {
	if rdb == nil || requestID == "" {
		return
	}

	ctx := context.Background()
	key := idempotentKey(requestID)

	result := idempotentResult{
		HTTPStatus: httpStatus,
		Code:       code,
		Data:       nil,
	}

	body, err := json.Marshal(result)
	if err != nil {
		return
	}

	rdb.Set(ctx, key, string(body), idempotentTTL)
}

// LookupIdempotentResult 根据 request_id 查询缓存的幂等结果
func LookupIdempotentResult(rdb *redis.Client, requestID string) (*idempotentResult, error) {
	if rdb == nil || requestID == "" {
		return nil, fmt.Errorf("redis not available or empty request_id")
	}

	ctx := context.Background()
	key := idempotentKey(requestID)

	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if val == "__claimed__" {
		return nil, fmt.Errorf("request still processing")
	}

	var result idempotentResult
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, err
	}

	return &result, nil
}
