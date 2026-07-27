package service

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

// cacheKey 返回 Redis 缓存 key
func cacheKey(keyHash string) string {
	return "ck:" + keyHash
}

// BizError 业务错误码包装，实现 error 接口，用于 Redis 路径区分「业务错误」和「系统错误」
type BizError struct {
	Code    int
	Message string
}

func (e *BizError) Error() string { return e.Message }

func newBizError(code int) *BizError {
	return &BizError{Code: code, Message: fmt.Sprintf("biz error: %d", code)}
}

// rateLimitToInt 将 *int 转换为 int，nil 视为 -1（表示使用租户默认值）
func rateLimitToInt(v *int) int {
	if v == nil {
		return -1
	}
	return *v
}

// intToRateLimitPtr 将 int 转换为 *int，-1 视为 nil（使用租户默认值），0 视为不限速
func intToRateLimitPtr(v int) *int {
	if v <= 0 {
		return nil
	}
	return &v
}

type KeyService struct {
	db      *gorm.DB
	rdb     *redis.Client
	mqSvc   *MQService
	sfGroup singleflight.Group
}

func NewKeyService(db *gorm.DB, rdb *redis.Client, mqSvc *MQService) *KeyService {
	return &KeyService{
		db:    db,
		rdb:   rdb,
		mqSvc: mqSvc,
	}
}

// topKeysZSetKey returns the Redis ZSET key for a tenant's top-keys ranking.
func topKeysZSetKey(tenantID uint64) string {
	return "top_keys:" + strconv.FormatUint(tenantID, 10)
}

// topAmountZSetKey returns the Redis ZSET key for a tenant's top-amount ranking.
func topAmountZSetKey(tenantID uint64) string {
	return "top_amount:" + strconv.FormatUint(tenantID, 10)
}

// rateLimitKey returns the Redis ZSET key for per-key rate limiting.
func rateLimitKey(keyID uint64) string {
	return "rl:key:" + strconv.FormatUint(keyID, 10)
}

func (s *KeyService) generateRawKeyWithConfig(prefix string, keyLen int) (string, error) {
	bytes := make([]byte, keyLen)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random: %w", err)
	}
	return prefix + hex.EncodeToString(bytes), nil
}

func (s *KeyService) hashKey(rawKey string) string {
	h := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(h[:])
}

type CreateKeyRequest struct {
	Alias           string     `json:"alias"`
	RemainingAmount int64      `json:"remaining_amount"`
	CreatedBy       string     `json:"created_by"`
	ExpireAt        *time.Time `json:"expire_at"`
	MaxUsage        *int64     `json:"max_usage"`
	RateLimit       *int       `json:"rate_limit"`        // nil=使用租户默认，0=不限速
	RateLimitWindow *int       `json:"rate_limit_window"` // 窗口大小（秒）
}

type CreateKeyResult struct {
	RawKey string    `json:"raw_key"`
	Key    model.Key `json:"key"`
}

func (s *KeyService) CreateKey(req CreateKeyRequest, tenantID uint64, keyPrefix string, keyLen, suffixLen int) (*CreateKeyResult, error) {
	rawKey, err := s.generateRawKeyWithConfig(keyPrefix, keyLen)
	if err != nil {
		return nil, err
	}

	if len(rawKey) < suffixLen {
		suffixLen = len(rawKey)
	}
	suffix := rawKey[len(rawKey)-suffixLen:]

	keyHash := s.hashKey(rawKey)

	key := model.Key{
		TenantID:        tenantID,
		Alias:           req.Alias,
		KeyHash:         keyHash,
		KeyPrefix:       keyPrefix,
		KeySuffix:       suffix,
		RemainingAmount: req.RemainingAmount,
		Version:         0,
		Status:          model.KeyStatusActive,
		CreatedBy:       req.CreatedBy,
		ExpireAt:        req.ExpireAt,
		MaxUsage:        req.MaxUsage,
		RateLimit:       req.RateLimit,
		RateLimitWindow: req.RateLimitWindow,
	}

	if err := s.db.Create(&key).Error; err != nil {
		return nil, fmt.Errorf("create key: %w", err)
	}

	s.syncCacheOnCreate(&key)

	return &CreateKeyResult{RawKey: rawKey, Key: key}, nil
}

// syncCacheOnCreate 创建 key 后同步缓存
func (s *KeyService) syncCacheOnCreate(key *model.Key) {
	if s.rdb == nil {
		return
	}
	ctx := context.Background()
	ck := cacheKey(key.KeyHash)
	expireTs := int64(0)
	if key.ExpireAt != nil {
		expireTs = key.ExpireAt.UnixMilli()
	}
	if err := s.rdb.HSet(ctx, ck, map[string]interface{}{
		"id":               key.ID,
		"tenant_id":        key.TenantID,
		"alias":            key.Alias,
		"key_suffix":       key.KeySuffix,
		"remaining":        key.RemainingAmount,
		"status":           string(key.Status),
		"version":          key.Version,
		"expire_at":        expireTs,
		"created_at":       key.CreatedAt.UnixMilli(),
		"used_at":          0,
		"rate_limit":       rateLimitToInt(key.RateLimit),
		"rate_limit_window": rateLimitToInt(key.RateLimitWindow),
	}).Err(); err != nil {
		zap.L().Debug("syncCacheOnCreate failed", zap.Error(err))
	}
}

func (s *KeyService) FindByRawKeyTenant(rawKey string, tenantID uint64) (*model.Key, error) {
	keyHash := s.hashKey(rawKey)
	var key model.Key
	if err := s.db.Where("key_hash = ? AND tenant_id = ?", keyHash, tenantID).First(&key).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &key, nil
}

// LookupKey 统一查询 key：Redis 优先，miss 时回源 MySQL 并回填缓存（singleflight 防击穿）。
// key 不存在时返回 (nil, nil)。
func (s *KeyService) LookupKey(rawKey string, tenantID uint64) (*model.Key, error) {
	keyHash := s.hashKey(rawKey)
	log := zap.L().With(
		zap.String("key_hash", keyHash[:16]),
		zap.Uint64("tenant_id", tenantID),
	)

	// Redis 不可用时直接查 MySQL
	if s.rdb == nil {
		log.Warn("LookupKey: Redis 不可用，降级查 MySQL")
		return s.FindByRawKeyTenant(rawKey, tenantID)
	}

	ck := cacheKey(keyHash)
	ctx := context.Background()
	// 尝试从 Redis 缓存读取
	fields, err := s.rdb.HGetAll(ctx, ck).Result()
	if err != nil {
		log.Error("LookupKey: Redis HGetAll 失败", zap.Error(err))
	} else if len(fields) > 0 {
		key, ok := buildKeyFromHash(fields, tenantID)
		if ok {
			key.KeyHash = keyHash // ensure KeyHash is set for downstream (e.g. syncCacheOnCreate)
			log.Info("LookupKey: Redis 命中",
				zap.Uint64("key_id", key.ID),
				zap.String("status", string(key.Status)),
				zap.Int64("remaining", key.RemainingAmount),
			)
			return key, nil
		}
		log.Warn("LookupKey: Redis 数据校验失败（tenant 不匹配或 id 为空），忽略缓存",
			zap.Int("fields_count", len(fields)),
		)
	} else {
		log.Info("LookupKey: Redis 缓存为空")
	}

	// 缓存未命中，singleflight 回源 MySQL
	sfKey := fmt.Sprintf("%s:%d", ck, tenantID)
	log.Debug("LookupKey: 回源 MySQL（singleflight）")
	v, err, _ := s.sfGroup.Do(sfKey, func() (interface{}, error) {
		var key model.Key
		if err := s.db.Where("key_hash = ? AND tenant_id = ?", keyHash, tenantID).First(&key).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Debug("LookupKey: MySQL 未找到记录")
				return nil, nil
			}
			log.Error("LookupKey: MySQL 查询异常", zap.Error(err))
			return nil, err
		}

		log.Debug("LookupKey: MySQL 命中",
			zap.Uint64("key_id", key.ID),
			zap.String("status", string(key.Status)),
			zap.Int64("remaining", key.RemainingAmount),
			zap.Int64("version", key.Version),
		)

		// 回填 Redis 缓存
		expireTs := int64(0)
		if key.ExpireAt != nil {
			expireTs = key.ExpireAt.UnixMilli()
		}
		usedAtTs := int64(0)
		if key.UsedAt != nil {
			usedAtTs = key.UsedAt.UnixMilli()
		}
		if err := s.rdb.HSet(ctx, ck, map[string]interface{}{
			"id":               key.ID,
			"tenant_id":        key.TenantID,
			"alias":            key.Alias,
			"key_suffix":       key.KeySuffix,
			"remaining":        key.RemainingAmount,
			"status":           string(key.Status),
			"version":          key.Version,
			"expire_at":        expireTs,
			"created_at":       key.CreatedAt.UnixMilli(),
			"used_at":          usedAtTs,
			"rate_limit":       rateLimitToInt(key.RateLimit),
			"rate_limit_window": rateLimitToInt(key.RateLimitWindow),
		}).Err(); err != nil {
			log.Error("LookupKey: Redis HSet 回填失败", zap.Error(err))
		} else {
			log.Debug("LookupKey: Redis 缓存回填成功")
		}

		return &key, nil
	})

	if err != nil {
		log.Error("LookupKey: singleflight 执行失败", zap.Error(err))
		return nil, err
	}
	if v == nil {
		log.Debug("LookupKey: key 不存在")
		return nil, nil
	}
	return v.(*model.Key), nil
}

// buildKeyFromHash 从 Redis Hash 字段构造 model.Key。
// tenantID 用于校验归属。返回 (key, true) 表示成功。
func buildKeyFromHash(fields map[string]string, tenantID uint64) (*model.Key, bool) {
	id, _ := strconv.ParseUint(fields["id"], 10, 64)
	tid, _ := strconv.ParseUint(fields["tenant_id"], 10, 64)
	remaining, _ := strconv.ParseInt(fields["remaining"], 10, 64)
	version, _ := strconv.ParseInt(fields["version"], 10, 64)
	expireAtMs, _ := strconv.ParseInt(fields["expire_at"], 10, 64)
	createdAtMs, _ := strconv.ParseInt(fields["created_at"], 10, 64)

	if id == 0 || tid != tenantID {
		return nil, false
	}

	key := &model.Key{
		ID:              id,
		TenantID:        tid,
		Alias:           fields["alias"],
		KeySuffix:       fields["key_suffix"],
		RemainingAmount: remaining,
		Status:          model.KeyStatus(fields["status"]),
		Version:         version,
		CreatedAt:       time.UnixMilli(createdAtMs),
	}
	if expireAtMs > 0 {
		t := time.UnixMilli(expireAtMs)
		key.ExpireAt = &t
	}
	if usedAtMs, _ := strconv.ParseInt(fields["used_at"], 10, 64); usedAtMs > 0 {
		t := time.UnixMilli(usedAtMs)
		key.UsedAt = &t
	}

	// 解析限流配置：-1 表示使用租户默认值，0 表示不限速，正整数表示限流值
	if rl, err := strconv.Atoi(fields["rate_limit"]); err == nil && rl >= 0 {
		key.RateLimit = &rl
	} else if rl == -1 {
		key.RateLimit = nil // 使用租户默认值
	}
	if rlw, err := strconv.Atoi(fields["rate_limit_window"]); err == nil && rlw >= 0 {
		key.RateLimitWindow = &rlw
	} else if rlw == -1 {
		key.RateLimitWindow = nil
	}

	return key, true
}

type KeyStatusResult struct {
	Alias           string          `json:"alias"`
	RemainingAmount int64           `json:"remaining_amount"`
	Status          model.KeyStatus `json:"status"`
	CreatedAt       string          `json:"created_at"`
	UsedAt          *string         `json:"used_at"`
}

func (s *KeyService) GetKeyStatusByTenant(rawKey string, tenantID uint64) (*KeyStatusResult, error) {
	key, err := s.LookupKey(rawKey, tenantID)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, nil
	}

	var usedAt *string
	if key.UsedAt != nil {
		t := key.UsedAt.Format("2006-01-02 15:04:05")
		usedAt = &t
	}

	return &KeyStatusResult{
		Alias:           key.Alias,
		RemainingAmount: key.RemainingAmount,
		Status:          key.Status,
		CreatedAt:       key.CreatedAt.Format("2006-01-02 15:04:05"),
		UsedAt:          usedAt,
	}, nil
}

type ConsumeResult struct {
	RemainingAmount int64           `json:"remaining_amount"`
	Status          model.KeyStatus `json:"status"`
	Exhausted       bool            `json:"exhausted"`
}

// ConsumeMeta 请求上下文，由 handler 层注入
type ConsumeMeta struct {
	RequestID string
	IP        string
	UserAgent string
}

// consumeResult 从 Lua 脚本返回的 JSON 解析
type consumeResult struct {
	Code      int    `json:"code"`
	Remaining int64  `json:"remaining"`
	Status    string `json:"status"`
	KeyID     uint64 `json:"key_id"`
	TenantID  uint64 `json:"tenant_id"`
	Alias     string `json:"alias"`
}

// consumeViaRedis Redis 原子路径：LookupKey + Lua 扣减 + MQ 发布。
// LookupKey 在此函数内部调用，Lua 脚本保证「读 + 校验 + 扣减」原子性。
func (s *KeyService) consumeViaRedis(rawKey string, amount int64, tenantID uint64, meta *ConsumeMeta) (*ConsumeResult, error) {
	key, err := s.LookupKey(rawKey, tenantID)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, newBizError(errcode.CodeKeyNotFound)
	}

	ck := cacheKey(key.KeyHash)
	ctx := context.Background()

	raw, err := consumeLuaScript.Run(ctx, s.rdb, []string{ck}, amount, time.Now().UnixMilli()).Result()
	if err != nil {
		return nil, fmt.Errorf("lua consume: %w", err)
	}

	resStr, ok := raw.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected lua result type: %T", raw)
	}

	var res consumeResult
	if err := json.Unmarshal([]byte(resStr), &res); err != nil {
		return nil, fmt.Errorf("unmarshal lua result: %w", err)
	}

	// Lua 返回 -1 表示缓存数据丢失（可能被淘汰），重新回填后重试
	if res.Code == -1 {
		zap.L().Warn("consumeViaRedis: Lua cache miss，重新回填缓存",
			zap.Uint64("key_id", key.ID),
		)
		s.syncCacheOnCreate(key)

		raw, err = consumeLuaScript.Run(ctx, s.rdb, []string{ck}, amount, time.Now().UnixMilli()).Result()
		if err != nil {
			return nil, fmt.Errorf("lua consume retry: %w", err)
		}
		if err := json.Unmarshal([]byte(raw.(string)), &res); err != nil {
			return nil, fmt.Errorf("unmarshal lua result: %w", err)
		}
		if res.Code == -1 {
			// 回填后仍然 miss，返回系统错误触发 MySQL 降级
			return nil, fmt.Errorf("cache miss after backfill")
		}
	}

	// Lua 返回的非 0 code 表示业务错误，直接返回（不降级到 MySQL）
	if res.Code != 0 {
		return nil, newBizError(res.Code)
	}

	// 发布 MQ 事件
	if s.mqSvc != nil {
		now := time.Now().UnixMilli()
		event := ConsumeEvent{
			EventID:        uuid.New().String(),
			RequestID:      meta.RequestID,
			KeyID:          res.KeyID,
			KeyAlias:       res.Alias,
			KeySuffix:      key.KeySuffix,
			TenantID:       res.TenantID,
			Amount:         amount,
			RemainingAfter: res.Remaining,
			StatusAfter:    res.Status,
			Timestamp:      now,
			UsedAt:         now,
		}
		if meta != nil {
			event.IP = meta.IP
			event.UserAgent = meta.UserAgent
		}
		if err := s.mqSvc.PublishConsumeEvent(event); err != nil {
			zap.L().Error("发布 ConsumeEvent 失败", zap.Error(err))
		}
	}

	status := model.KeyStatus(res.Status)
	return &ConsumeResult{
		RemainingAmount: res.Remaining,
		Status:          status,
		Exhausted:       res.Remaining <= 0,
	}, nil
}

// consumeViaMySQL MySQL 降级路径：乐观锁重试
// 每次重试直接查 MySQL 拿最新 version，保证乐观锁正确。
func (s *KeyService) consumeViaMySQL(rawKey string, amount int64, tenantID uint64, meta *ConsumeMeta) (*ConsumeResult, int, error) {
	const maxRetries = 3
	for retry := 0; retry < maxRetries; retry++ {
		// 直接查 MySQL 获取最新状态和 version
		key, err := s.FindByRawKeyTenant(rawKey, tenantID)
		if err != nil {
			return nil, 0, err
		}
		if key == nil {
			return nil, errcode.CodeKeyNotFound, nil
		}
		if key.Status == model.KeyStatusDisabled {
			return nil, errcode.CodeKeyDisabled, nil
		}
		if key.Status == model.KeyStatusExpired {
			return nil, errcode.CodeKeyExpired, nil
		}
		if key.Status == model.KeyStatusExhausted {
			return nil, errcode.CodeKeyExhausted, nil
		}
		if !key.CanDeduct(amount) {
			if key.RemainingAmount <= 0 {
				return nil, errcode.CodeKeyExhausted, nil
			}
			return nil, errcode.CodeKeyInsufficient, nil
		}

		tx := s.db.Begin()
		if tx.Error != nil {
			return nil, 0, tx.Error
		}

		result := tx.Model(&model.Key{}).
			Where("id = ? AND version = ? AND remaining_amount >= ? AND status = ?",
				key.ID, key.Version, amount, model.KeyStatusActive).
			Updates(map[string]interface{}{
				"remaining_amount": gorm.Expr("remaining_amount - ?", amount),
				"version":          gorm.Expr("version + 1"),
				"used_at":          time.Now(),
			})

		if result.Error != nil {
			tx.Rollback()
			return nil, 0, result.Error
		}
		if result.RowsAffected == 0 {
			tx.Rollback()
			continue
		}

		newRemaining := key.RemainingAmount - amount

		if newRemaining <= 0 {
			if err := tx.Model(&model.Key{}).Where("id = ?", key.ID).
				Updates(map[string]interface{}{
					"status": model.KeyStatusExhausted,
				}).Error; err != nil {
				tx.Rollback()
				return nil, 0, err
			}
		}

		if err := tx.Commit().Error; err != nil {
			return nil, 0, err
		}

		status := key.Status
		if newRemaining <= 0 {
			status = model.KeyStatusExhausted
		}

		// 原子递增 Redis ZSET 中该 key 的消费计数和额度消耗
		if s.rdb != nil {
			ctx := context.Background()
			member := strconv.FormatUint(key.ID, 10)
			if err := s.rdb.ZIncrBy(ctx, topKeysZSetKey(tenantID), 1, member).Err(); err != nil {
				zap.L().Debug("top_keys ZINCRBY failed", zap.Error(err))
			}
			if err := s.rdb.ZIncrBy(ctx, topAmountZSetKey(tenantID), float64(amount), member).Err(); err != nil {
				zap.L().Debug("top_amount ZINCRBY failed", zap.Error(err))
			}

			// 更新单 key 使用统计 ZSET
			today := time.Now().Format("2006-01-02")
			callsKey := keyUsageCallsKey(tenantID, key.ID)
			amountKey := keyUsageAmountKey(tenantID, key.ID)
			if err := s.rdb.ZIncrBy(ctx, callsKey, 1, today).Err(); err != nil {
				zap.L().Debug("key_usage calls ZINCRBY failed", zap.Error(err))
			} else {
				s.rdb.Expire(ctx, callsKey, 31*24*time.Hour)
			}
			if err := s.rdb.ZIncrBy(ctx, amountKey, float64(amount), today).Err(); err != nil {
				zap.L().Debug("key_usage amount ZINCRBY failed", zap.Error(err))
			} else {
				s.rdb.Expire(ctx, amountKey, 31*24*time.Hour)
			}
		}

		// MySQL 降级路径直接写入使用日志（正常路径由 MQ Worker 写入）
		usageLog := model.UsageLog{
			TenantID:       tenantID,
			KeyID:          key.ID,
			KeyAlias:       key.Alias,
			KeySuffix:      key.KeySuffix,
			Amount:         amount,
			ResponseStatus: http.StatusOK,
		}
		if meta != nil {
			usageLog.IP = meta.IP
			usageLog.UserAgent = meta.UserAgent
			usageLog.RequestID = meta.RequestID
		}
		if err := s.db.Create(&usageLog).Error; err != nil {
			zap.L().Error("MySQL 降级路径写入使用日志失败", zap.Error(err))
		}

		return &ConsumeResult{
			RemainingAmount: newRemaining,
			Status:          status,
			Exhausted:       newRemaining <= 0,
		}, 0, nil
	}
	return nil, 0, fmt.Errorf("concurrency conflict after %d retries", maxRetries)
}

// CheckConsumeRateLimit 检查 key 的消费限流（供 handler 在 consume 前调用）。
// 返回 (allowed, retryAfter, bizCode, error)。
// allowed=false 时 retryAfter 表示建议等待秒数；bizCode 非 0 表示 key 异常。
func (s *KeyService) CheckConsumeRateLimit(rawKey string, tenantID uint64) (bool, int, int, error) {
	if s.rdb == nil {
		return true, 0, 0, nil // Redis 不可用时限流跳过
	}

	key, err := s.LookupKey(rawKey, tenantID)
	if err != nil {
		return false, 0, 0, err
	}
	if key == nil {
		return false, 0, errcode.CodeKeyNotFound, nil
	}

	allowed, retryAfter, err := s.checkKeyRateLimit(key, tenantID)
	if err != nil {
		return false, 0, 0, err
	}
	if !allowed {
		return false, retryAfter, 0, nil
	}
	return true, 0, 0, nil
}

func (s *KeyService) ConsumeKeyByTenant(rawKey string, amount int64, tenantID uint64, meta *ConsumeMeta) (*ConsumeResult, int, error) {
	if amount <= 0 {
		return nil, 0, fmt.Errorf("invalid amount: %d", amount)
	}

	// Redis Lua 原子路径：LookupKey + 校验 + 扣减在 Lua 内一次完成
	if s.rdb != nil {
		result, err := s.consumeViaRedis(rawKey, amount, tenantID, meta)
		if err == nil {
			return result, 0, nil
		}
		// 业务错误码直接返回，不降级
		if bizErr, ok := err.(*BizError); ok {
			return nil, bizErr.Code, nil
		}
		// 系统错误降级到 MySQL
		zap.L().Warn("Redis deduct failed, falling back to MySQL", zap.Error(err))
	}

	// MySQL 乐观锁降级
	return s.consumeViaMySQL(rawKey, amount, tenantID, meta)
}

type KeyListQuery struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	Status    string `form:"status"`
	Alias     string `form:"alias"`      // 别名前缀搜索
	KeySuffix string `form:"key_suffix"` // 后缀精准搜索
}

func (s *KeyService) GetKeyDetail(id, tenantID uint64) (*model.Key, error) {
	var key model.Key
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (s *KeyService) ListKeys(query KeyListQuery, tenantID uint64) ([]model.Key, int64, error) {
	var keys []model.Key
	var total int64

	db := s.db.Model(&model.Key{}).Where("tenant_id = ?", tenantID)
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Alias != "" {
		db = db.Where("alias LIKE ?", query.Alias+"%")
	}
	if query.KeySuffix != "" {
		db = db.Where("key_suffix = ?", query.KeySuffix)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&keys).Error; err != nil {
		return nil, 0, err
	}
	return keys, total, nil
}

func (s *KeyService) ListKeysByCreatedBy(createdBy string, tenantID uint64, page, pageSize int) ([]model.Key, int64, error) {
	var keys []model.Key
	var total int64

	db := s.db.Model(&model.Key{}).Where("created_by = ? AND tenant_id = ?", createdBy, tenantID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&keys).Error; err != nil {
		return nil, 0, err
	}
	return keys, total, nil
}

func (s *KeyService) ListKeysByTenant(tenantID uint64, page, pageSize int) ([]model.Key, int64, error) {
	var keys []model.Key
	var total int64

	db := s.db.Model(&model.Key{}).Where("tenant_id = ?", tenantID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&keys).Error; err != nil {
		return nil, 0, err
	}
	return keys, total, nil
}

type UpdateKeyRequest struct {
	Alias           *string `json:"alias"`
	RateLimit       *int    `json:"rate_limit"`        // nil=不更新，0=不限速
	RateLimitWindow *int    `json:"rate_limit_window"` // nil=不更新
}

func (s *KeyService) UpdateKey(id, tenantID uint64, req UpdateKeyRequest) error {
	updates := make(map[string]interface{})
	if req.Alias != nil {
		updates["alias"] = *req.Alias
	}
	if req.RateLimit != nil {
		updates["rate_limit"] = *req.RateLimit
	}
	if req.RateLimitWindow != nil {
		updates["rate_limit_window"] = *req.RateLimitWindow
	}
	if len(updates) == 0 {
		return nil
	}
	if err := s.db.Model(&model.Key{}).Where("id = ? AND tenant_id = ?", id, tenantID).Updates(updates).Error; err != nil {
		return err
	}
	// 更新缓存中的限流字段
	s.syncCacheRateLimit(id, tenantID, updates)
	return nil
}

type AdjustBalanceRequest struct {
	Delta     int64  `json:"delta"`     // 正=增加, 负=减少
	Operator  string `json:"operator"`  // 操作人标识
	Remark    string `json:"remark"`    // 可选备注
	RequestID string `json:"request_id"` // 请求 ID（幂等 + 可追溯）
}

type AdjustBalanceResult struct {
	BeforeAmount int64 `json:"before_amount"`
	AfterAmount  int64 `json:"after_amount"`
}

// adjustResult 从调额 Lua 脚本返回的 JSON 解析
type adjustResult struct {
	Code     int    `json:"code"`
	Before   int64  `json:"before"`
	After    int64  `json:"after"`
	Status   string `json:"status"`
	KeyID    uint64 `json:"key_id"`
	TenantID uint64 `json:"tenant_id"`
	Alias    string `json:"alias"`
	Error    string `json:"error"`
}

func (s *KeyService) AdjustBalance(id, tenantID uint64, req AdjustBalanceRequest) (*AdjustBalanceResult, error) {
	if req.Delta == 0 {
		return nil, fmt.Errorf("delta 不能为 0")
	}

	// 需要先拿到 keyHash（从 MySQL）
	var key model.Key
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&key).Error; err != nil {
		return nil, fmt.Errorf("卡密不存在")
	}

	// Redis 快速路径
	if s.rdb != nil {
		result, err := s.adjustViaRedis(key.KeyHash, id, tenantID, req, key)
		if err == nil {
			return result, nil
		}
		zap.L().Warn("Redis adjust failed, falling back to MySQL", zap.Error(err))
	}

	// 降级到 MySQL
	return s.adjustViaMySQL(id, tenantID, req)
}

func (s *KeyService) adjustViaRedis(keyHash string, id, tenantID uint64, req AdjustBalanceRequest, key model.Key) (*AdjustBalanceResult, error) {
	ck := cacheKey(keyHash)
	ctx := context.Background()

	raw, err := adjustLuaScript.Run(ctx, s.rdb, []string{ck}, req.Delta).Result()
	if err != nil {
		return nil, fmt.Errorf("lua adjust: %w", err)
	}

	var res adjustResult
	if err := json.Unmarshal([]byte(raw.(string)), &res); err != nil {
		return nil, fmt.Errorf("unmarshal lua result: %w", err)
	}

	// Cache miss → 加载后重试
	if res.Code == -1 {
		// 先把当前 MySQL 数据加载到缓存
		if s.rdb != nil {
			s.syncCacheOnCreate(&key)
		}
		raw, err = adjustLuaScript.Run(ctx, s.rdb, []string{ck}, req.Delta).Result()
		if err != nil {
			return nil, fmt.Errorf("lua adjust retry: %w", err)
		}
		if err := json.Unmarshal([]byte(raw.(string)), &res); err != nil {
			return nil, fmt.Errorf("unmarshal lua result: %w", err)
		}
		if res.Code == -1 {
			return nil, fmt.Errorf("缓存加载失败")
		}
	}

	if res.Code != 0 {
		return nil, fmt.Errorf("%s", res.Error)
	}

	// 发布 MQ 事件
	if s.mqSvc != nil {
		event := AdjustEvent{
			EventID:        uuid.New().String(),
			RequestID:      req.RequestID,
			KeyID:          id,
			KeyAlias:       key.Alias,
			KeySuffix:      key.KeySuffix,
			TenantID:       tenantID,
			Delta:          req.Delta,
			RemainingAfter: res.After,
			StatusAfter:    res.Status,
			Operator:       req.Operator,
			Remark:         req.Remark,
			Timestamp:      time.Now().UnixMilli(),
		}
		if err := s.mqSvc.PublishAdjustEvent(event); err != nil {
			zap.L().Error("发布 AdjustEvent 失败", zap.Error(err))
		}
	}

	return &AdjustBalanceResult{
		BeforeAmount: res.Before,
		AfterAmount:  res.After,
	}, nil
}

// adjustViaMySQL MySQL 降级路径：乐观锁重试
func (s *KeyService) adjustViaMySQL(id, tenantID uint64, req AdjustBalanceRequest) (*AdjustBalanceResult, error) {
	const maxRetries = 3
	for retry := 0; retry < maxRetries; retry++ {
		var key model.Key
		if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&key).Error; err != nil {
			return nil, fmt.Errorf("卡密不存在")
		}

		newAmount := key.RemainingAmount + req.Delta
		if newAmount < 0 {
			return nil, fmt.Errorf("调整后余额不能低于 0，当前余额: %d, 调整量: %d", key.RemainingAmount, req.Delta)
		}

		tx := s.db.Begin()
		if tx.Error != nil {
			return nil, tx.Error
		}

		result := tx.Model(&model.Key{}).
			Where("id = ? AND version = ?", key.ID, key.Version).
			Updates(map[string]interface{}{
				"remaining_amount": gorm.Expr("remaining_amount + ?", req.Delta),
				"version":          gorm.Expr("version + 1"),
			})

		if result.Error != nil {
			tx.Rollback()
			return nil, result.Error
		}
		if result.RowsAffected == 0 {
			tx.Rollback()
			continue // 乐观锁冲突，重试
		}

		// 如果是增加额度且 key 之前已耗尽，恢复为 active
		if req.Delta > 0 && newAmount > 0 && key.Status == model.KeyStatusExhausted {
			if err := tx.Model(&model.Key{}).Where("id = ?", key.ID).
				Update("status", model.KeyStatusActive).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}

		if err := tx.Commit().Error; err != nil {
			return nil, err
		}

		return &AdjustBalanceResult{
			BeforeAmount: key.RemainingAmount,
			AfterAmount:  newAmount,
		}, nil
	}
	return nil, fmt.Errorf("并发冲突，超过最大重试次数")
}

// syncCacheOnStatusChange 更新缓存中的 status 字段
func (s *KeyService) syncCacheOnStatusChange(keyHash string, status model.KeyStatus) {
	if s.rdb == nil {
		return
	}
	ctx := context.Background()
	ck := cacheKey(keyHash)
	if err := s.rdb.HSet(ctx, ck, "status", string(status)).Err(); err != nil {
		zap.L().Debug("syncCacheOnStatusChange failed", zap.Error(err))
	}
}

// syncCacheRateLimit 更新缓存中的限流字段
func (s *KeyService) syncCacheRateLimit(keyID, tenantID uint64, updates map[string]interface{}) {
	if s.rdb == nil {
		return
	}
	// 需要先获取 keyHash 才能定位缓存 key
	var key model.Key
	if err := s.db.Select("key_hash").Where("id = ? AND tenant_id = ?", keyID, tenantID).First(&key).Error; err != nil {
		return
	}
	ctx := context.Background()
	ck := cacheKey(key.KeyHash)
	cacheUpdates := make(map[string]interface{})
	if v, ok := updates["rate_limit"]; ok {
		if iv, ok := v.(int); ok {
			cacheUpdates["rate_limit"] = iv
		}
	}
	if v, ok := updates["rate_limit_window"]; ok {
		if iv, ok := v.(int); ok {
			cacheUpdates["rate_limit_window"] = iv
		}
	}
	if len(cacheUpdates) > 0 {
		if err := s.rdb.HSet(ctx, ck, cacheUpdates).Err(); err != nil {
			zap.L().Debug("syncCacheRateLimit failed", zap.Error(err))
		}
	}
}

// GetTenantDefaultRateLimit 从数据库获取租户的默认限流配置
func (s *KeyService) GetTenantDefaultRateLimit(tenantID uint64) (rateLimit *int, rateLimitWindow *int, err error) {
	var tenant model.Tenant
	if err := s.db.Select("default_rate_limit", "default_rate_limit_window").
		Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		return nil, nil, err
	}
	return tenant.DefaultRateLimit, tenant.DefaultRateLimitWindow, nil
}

// InvalidateTenantKeyCaches 清除租户下所有 key 的缓存（用于租户默认限流配置变更后）
func (s *KeyService) InvalidateTenantKeyCaches(tenantID uint64) {
	if s.rdb == nil {
		return
	}
	var keys []model.Key
	if err := s.db.Select("key_hash").Where("tenant_id = ?", tenantID).Find(&keys).Error; err != nil {
		zap.L().Debug("InvalidateTenantKeyCaches: query failed", zap.Error(err))
		return
	}
	ctx := context.Background()
	for _, k := range keys {
		s.rdb.Del(ctx, cacheKey(k.KeyHash))
	}
}

// checkKeyRateLimit 检查 key 是否触发限流。
// 需要 key 包含 ID、RateLimit、RateLimitWindow 字段。
// 返回 (allowed, retryAfter, error)。Redis 错误时 fail-open（放行）。
func (s *KeyService) checkKeyRateLimit(key *model.Key, tenantID uint64) (bool, int, error) {
	if s.rdb == nil {
		return true, 0, nil
	}

	// 解析 key 的限流配置：nil=使用租户默认，0=不限速
	rateLimit := 0
	rateLimitWindow := 0

	if key.RateLimit != nil && key.RateLimitWindow != nil {
		rateLimit = *key.RateLimit
		rateLimitWindow = *key.RateLimitWindow
	} else {
		// Key 未配置，使用租户默认值
		tenantRL, tenantRLW, err := s.GetTenantDefaultRateLimit(tenantID)
		if err != nil {
			zap.L().Warn("checkKeyRateLimit: 获取租户默认限流配置失败",
				zap.Uint64("tenant_id", tenantID), zap.Error(err))
			return true, 0, nil // fail-open
		}
		if tenantRL != nil && tenantRLW != nil {
			rateLimit = *tenantRL
			rateLimitWindow = *tenantRLW
		}
	}

	// 不限速
	if rateLimit <= 0 || rateLimitWindow <= 0 {
		return true, 0, nil
	}

	ctx := context.Background()
	now := time.Now().UnixNano()
	uid := fmt.Sprintf("%d:%d", now, key.ID)

	rlKey := rateLimitKey(key.ID)
	result, err := rateLimitCheckScript.Run(ctx, s.rdb, []string{rlKey},
		rateLimitWindow, rateLimit, now, uid).Int64Slice()
	if err != nil {
		zap.L().Warn("checkKeyRateLimit: Redis 脚本执行失败，fail-open",
			zap.Uint64("key_id", key.ID), zap.Error(err))
		return true, 0, nil
	}

	if len(result) >= 2 && result[0] == 0 {
		return false, int(result[1]), nil
	}
	return true, 0, nil
}

func (s *KeyService) DisableKey(id, tenantID uint64) error {
	var key model.Key
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&key).Error; err != nil {
		return err
	}
	if err := s.db.Model(&model.Key{}).Where("id = ? AND tenant_id = ?", id, tenantID).Update("status", model.KeyStatusDisabled).Error; err != nil {
		return err
	}
	s.syncCacheOnStatusChange(key.KeyHash, model.KeyStatusDisabled)
	return nil
}

func (s *KeyService) EnableKey(id, tenantID uint64) error {
	var key model.Key
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&key).Error; err != nil {
		return err
	}
	newStatus := model.KeyStatusActive
	if key.RemainingAmount <= 0 {
		newStatus = model.KeyStatusExhausted
	}
	if err := s.db.Model(&model.Key{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Update("status", newStatus).Error; err != nil {
		return err
	}
	s.syncCacheOnStatusChange(key.KeyHash, newStatus)
	return nil
}

// syncCacheOnDelete 删除缓存
func (s *KeyService) syncCacheOnDelete(keyHash string) {
	if s.rdb == nil {
		return
	}
	ctx := context.Background()
	ck := cacheKey(keyHash)
	if err := s.rdb.Del(ctx, ck).Err(); err != nil {
		zap.L().Debug("syncCacheOnDelete failed", zap.Error(err))
	}
}

func (s *KeyService) DeleteKey(id, tenantID uint64) error {
	var key model.Key
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&key).Error; err != nil {
		return err
	}
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Key{}).Error; err != nil {
		return err
	}
	s.syncCacheOnDelete(key.KeyHash)
	// 从 Redis ZSET 中移除该 key
	s.RemoveFromTopKeys(tenantID, id)
	s.RemoveFromTopAmount(tenantID, id)
	return nil
}

type ExportKeyItem struct {
	ID              uint64          `json:"id"`
	KeyPrefix       string          `json:"key_prefix"`
	KeySuffix       string          `json:"key_suffix"`
	Alias           string          `json:"alias"`
	RemainingAmount int64           `json:"remaining_amount"`
	Status          model.KeyStatus `json:"status"`
	CreatedAt       time.Time       `json:"created_at"`
	ExpireAt        *time.Time      `json:"expire_at"`
	MaxUsage        *int64          `json:"max_usage"`
}

func (s *KeyService) ExportKeysJSON(tenantID uint64) ([]ExportKeyItem, error) {
	var keys []model.Key
	if err := s.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, err
	}

	items := make([]ExportKeyItem, len(keys))
	for i, k := range keys {
		items[i] = ExportKeyItem{
			ID:              k.ID,
			KeyPrefix:       k.KeyPrefix,
			KeySuffix:       k.KeySuffix,
			Alias:           k.Alias,
			RemainingAmount: k.RemainingAmount,
			Status:          k.Status,
			CreatedAt:       k.CreatedAt,
			ExpireAt:        k.ExpireAt,
			MaxUsage:        k.MaxUsage,
		}
	}
	return items, nil
}

func (s *KeyService) ExportKeys(tenantID uint64) ([]model.Key, error) {
	var keys []model.Key
	if err := s.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

// FindKeysByTenant returns all keys for a given tenant, used by service accounts.
func (s *KeyService) FindKeysByTenant(tenantID uint64) ([]model.Key, error) {
	var keys []model.Key
	if err := s.db.Where("tenant_id = ?", tenantID).Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

// ExpireKeys 标记已过期的 key 为 expired 状态。
// 只处理 active 或 exhausted 状态的 key，disabled 优先级更高不会被覆盖。
func (s *KeyService) ExpireKeys() (int64, error) {
	// 先查出即将过期的 key
	var keys []model.Key
	s.db.Where("expire_at IS NOT NULL AND expire_at < NOW() AND status IN ?",
		[]string{string(model.KeyStatusActive), string(model.KeyStatusExhausted)}).
		Find(&keys)

	result := s.db.Model(&model.Key{}).
		Where("expire_at IS NOT NULL AND expire_at < NOW() AND status IN ?",
			[]string{string(model.KeyStatusActive), string(model.KeyStatusExhausted)}).
		Update("status", model.KeyStatusExpired)

	// 同步 Redis 缓存
	for _, k := range keys {
		s.syncCacheOnStatusChange(k.KeyHash, model.KeyStatusExpired)
	}

	return result.RowsAffected, result.Error
}

// RemoveFromTopKeys 从 Redis ZSET 中移除指定 key（用于卡密删除时清理）。
func (s *KeyService) RemoveFromTopKeys(tenantID, keyID uint64) {
	if s.rdb == nil {
		return
	}
	ctx := context.Background()
	member := strconv.FormatUint(keyID, 10)
	if err := s.rdb.ZRem(ctx, topKeysZSetKey(tenantID), member).Err(); err != nil {
		zap.L().Debug("top_keys ZREM failed", zap.Error(err))
	}
}

// RemoveFromTopAmount 从 Redis ZSET 中移除指定 key（用于卡密删除时清理）。
func (s *KeyService) RemoveFromTopAmount(tenantID, keyID uint64) {
	if s.rdb == nil {
		return
	}
	ctx := context.Background()
	member := strconv.FormatUint(keyID, 10)
	if err := s.rdb.ZRem(ctx, topAmountZSetKey(tenantID), member).Err(); err != nil {
		zap.L().Debug("top_amount ZREM failed", zap.Error(err))
	}
}

// BackfillTopKeys 从 usage_logs 聚合历史消费数据，回填 Redis ZSET。
// 应在服务启动时调用一次，确保 Redis 与数据库一致。
func (s *KeyService) BackfillTopKeys() {
	if s.rdb == nil {
		return
	}

	type keyCount struct {
		TenantID uint64
		KeyID    uint64
		Count    int64
	}
	var rows []keyCount

	if err := s.db.Model(&model.UsageLog{}).
		Select("tenant_id, key_id, COUNT(*) as count").
		Where("key_id > 0").
		Group("tenant_id, key_id").
		Scan(&rows).Error; err != nil {
		zap.L().Error("top_keys backfill query failed", zap.Error(err))
		return
	}

	// 按 tenantID 分组，批量 ZADD
	grouped := make(map[uint64][]redis.Z)
	for _, r := range rows {
		grouped[r.TenantID] = append(grouped[r.TenantID], redis.Z{
			Score:  float64(r.Count),
			Member: strconv.FormatUint(r.KeyID, 10),
		})
	}

	ctx := context.Background()
	var total int
	for tenantID, members := range grouped {
		if err := s.rdb.ZAdd(ctx, topKeysZSetKey(tenantID), members...).Err(); err != nil {
			zap.L().Error("top_keys ZADD failed",
				zap.Uint64("tenant_id", tenantID), zap.Error(err))
			continue
		}
		total += len(members)
	}

	zap.L().Info("top_keys 回填完成",
		zap.Int("tenants", len(grouped)),
		zap.Int("total_keys", total))
}

// BackfillTopAmount 从 usage_logs 聚合历史额度消耗，回填 Redis ZSET。
// 应在服务启动时调用一次，确保 Redis 与数据库一致。
func (s *KeyService) BackfillTopAmount() {
	if s.rdb == nil {
		return
	}

	type keyAmount struct {
		TenantID uint64
		KeyID    uint64
		Amount   int64
	}
	var rows []keyAmount

	if err := s.db.Model(&model.UsageLog{}).
		Select("tenant_id, key_id, SUM(amount) as amount").
		Where("key_id > 0").
		Group("tenant_id, key_id").
		Scan(&rows).Error; err != nil {
		zap.L().Error("top_amount backfill query failed", zap.Error(err))
		return
	}

	// 按 tenantID 分组，批量 ZADD
	grouped := make(map[uint64][]redis.Z)
	for _, r := range rows {
		grouped[r.TenantID] = append(grouped[r.TenantID], redis.Z{
			Score:  float64(r.Amount),
			Member: strconv.FormatUint(r.KeyID, 10),
		})
	}

	ctx := context.Background()
	var total int
	for tenantID, members := range grouped {
		if err := s.rdb.ZAdd(ctx, topAmountZSetKey(tenantID), members...).Err(); err != nil {
			zap.L().Error("top_amount ZADD failed",
				zap.Uint64("tenant_id", tenantID), zap.Error(err))
			continue
		}
		total += len(members)
	}

	zap.L().Info("top_amount 回填完成",
		zap.Int("tenants", len(grouped)),
		zap.Int("total_keys", total))
}

// refreshTopLockKey 返回限制刷新频率的 Redis key。
func refreshTopLockKey(tenantID uint64) string {
	return "top_refresh_lock:" + strconv.FormatUint(tenantID, 10)
}

// CanRefreshTop 检查该租户今天是否还能刷新 Top 统计。
func (s *KeyService) CanRefreshTop(tenantID uint64) (bool, error) {
	if s.rdb == nil {
		return true, nil
	}
	ctx := context.Background()
	exists, err := s.rdb.Exists(ctx, refreshTopLockKey(tenantID)).Result()
	if err != nil {
		return false, err
	}
	return exists == 0, nil
}

// RefreshTopStats 重新统计指定租户的 top_keys 和 top_amount Redis ZSET。
// 限频：同一天只能调用一次。
func (s *KeyService) RefreshTopStats(tenantID uint64) error {
	if s.rdb == nil {
		return fmt.Errorf("redis not available")
	}

	ctx := context.Background()
	lockKey := refreshTopLockKey(tenantID)

	// 尝试获取锁（SETNX + 24h TTL）
	ok, err := s.rdb.SetNX(ctx, lockKey, 1, 24*time.Hour).Result()
	if err != nil {
		return fmt.Errorf("setnx lock: %w", err)
	}
	if !ok {
		return fmt.Errorf("already refreshed")
	}

	// 统计调用次数 top_keys
	type keyCount struct {
		KeyID uint64
		Count int64
	}
	var callRows []keyCount
	if err := s.db.Model(&model.UsageLog{}).
		Select("key_id, COUNT(*) as count").
		Where("tenant_id = ? AND key_id > 0", tenantID).
		Group("key_id").Scan(&callRows).Error; err != nil {
		return fmt.Errorf("query top_keys: %w", err)
	}

	// 统计额度消耗 top_amount
	type keyAmount struct {
		KeyID  uint64
		Amount int64
	}
	var amountRows []keyAmount
	if err := s.db.Model(&model.UsageLog{}).
		Select("key_id, SUM(amount) as amount").
		Where("tenant_id = ? AND key_id > 0", tenantID).
		Group("key_id").Scan(&amountRows).Error; err != nil {
		return fmt.Errorf("query top_amount: %w", err)
	}

	// 构建 ZSET members
	callMembers := make([]redis.Z, 0, len(callRows))
	for _, r := range callRows {
		callMembers = append(callMembers, redis.Z{
			Score:  float64(r.Count),
			Member: strconv.FormatUint(r.KeyID, 10),
		})
	}
	amountMembers := make([]redis.Z, 0, len(amountRows))
	for _, r := range amountRows {
		amountMembers = append(amountMembers, redis.Z{
			Score:  float64(r.Amount),
			Member: strconv.FormatUint(r.KeyID, 10),
		})
	}

	// 原子替换：先删除旧数据，再写入新数据
	pipe := s.rdb.Pipeline()
	pipe.Del(ctx, topKeysZSetKey(tenantID))
	pipe.Del(ctx, topAmountZSetKey(tenantID))
	if len(callMembers) > 0 {
		pipe.ZAdd(ctx, topKeysZSetKey(tenantID), callMembers...)
	}
	if len(amountMembers) > 0 {
		pipe.ZAdd(ctx, topAmountZSetKey(tenantID), amountMembers...)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("pipeline exec: %w", err)
	}

	zap.L().Info("top 统计刷新完成",
		zap.Uint64("tenant_id", tenantID),
		zap.Int("top_keys", len(callMembers)),
		zap.Int("top_amount", len(amountMembers)))
	return nil
}
