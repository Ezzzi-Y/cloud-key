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

type KeyService struct {
	db        *gorm.DB
	rdb       *redis.Client
	mqSvc     *MQService
	keyPrefix string
	keyLength int
	suffixLen int
	sfGroup   singleflight.Group
}

func NewKeyService(db *gorm.DB, rdb *redis.Client, mqSvc *MQService) *KeyService {
	return &KeyService{
		db:        db,
		rdb:       rdb,
		mqSvc:     mqSvc,
		keyPrefix: "sk-",
		keyLength: 16,
		suffixLen: 4,
	}
}

// topKeysZSetKey returns the Redis ZSET key for a tenant's top-keys ranking.
func topKeysZSetKey(tenantID uint64) string {
	return "top_keys:" + strconv.FormatUint(tenantID, 10)
}

func (s *KeyService) WithConfig(prefix string, keyLen, suffixLen int) *KeyService {
	s.keyPrefix = prefix
	s.keyLength = keyLen
	s.suffixLen = suffixLen
	return s
}

func (s *KeyService) generateRawKey() (string, error) {
	return s.generateRawKeyWithConfig(s.keyPrefix, s.keyLength)
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
		"id":         key.ID,
		"tenant_id":  key.TenantID,
		"alias":      key.Alias,
		"remaining":  key.RemainingAmount,
		"status":     key.Status,
		"version":    key.Version,
		"expire_at":  expireTs,
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

type KeyStatusResult struct {
	Alias           string          `json:"alias"`
	RemainingAmount int64           `json:"remaining_amount"`
	Status          model.KeyStatus `json:"status"`
	CreatedAt       string          `json:"created_at"`
	UsedAt          *string         `json:"used_at"`
}

func (s *KeyService) GetKeyStatusByTenant(rawKey string, tenantID uint64) (*KeyStatusResult, error) {
	key, err := s.FindByRawKeyTenant(rawKey, tenantID)
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

// consumeResult 从 Lua 脚本返回的 JSON 解析
type consumeResult struct {
	Code      int    `json:"code"`
	Remaining int64  `json:"remaining"`
	Status    string `json:"status"`
	KeyID     uint64 `json:"key_id"`
	TenantID  uint64 `json:"tenant_id"`
	Alias     string `json:"alias"`
}

// loadKeyToCache 从 MySQL 加载 key 到 Redis 缓存（带 singleflight 防击穿）
// 返回 model.Key 指针和 error。key 不存在时返回 (nil, nil)。
func (s *KeyService) loadKeyToCache(rawKey string, tenantID uint64) (*model.Key, error) {
	keyHash := s.hashKey(rawKey)
	ck := cacheKey(keyHash)

	v, err, _ := s.sfGroup.Do(ck, func() (interface{}, error) {
		// 查 MySQL
		var key model.Key
		if err := s.db.Where("key_hash = ? AND tenant_id = ?", keyHash, tenantID).First(&key).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, nil
			}
			return nil, err
		}

		// 写入 Redis 缓存
		if s.rdb != nil {
			ctx := context.Background()
			expireTs := int64(0)
			if key.ExpireAt != nil {
				expireTs = key.ExpireAt.UnixMilli()
			}
			s.rdb.HSet(ctx, ck, map[string]interface{}{
				"id":         key.ID,
				"tenant_id":  key.TenantID,
				"alias":      key.Alias,
				"remaining":  key.RemainingAmount,
				"status":     key.Status,
				"version":    key.Version,
				"expire_at":  expireTs,
			})
		}

		return &key, nil
	})

	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	return v.(*model.Key), nil
}

// consumeViaRedis Redis 快速路径：Lua 扣减 + MQ 发布
func (s *KeyService) consumeViaRedis(rawKey string, amount int64, tenantID uint64) (*ConsumeResult, int, error) {
	keyHash := s.hashKey(rawKey)
	ck := cacheKey(keyHash)
	ctx := context.Background()

	// 执行 Lua 脚本
	raw, err := consumeLuaScript.Run(ctx, s.rdb, []string{ck}, amount, time.Now().UnixMilli()).Result()
	if err != nil {
		return nil, 0, fmt.Errorf("lua consume: %w", err)
	}

	var res consumeResult
	if err := json.Unmarshal([]byte(raw.(string)), &res); err != nil {
		return nil, 0, fmt.Errorf("unmarshal lua result: %w", err)
	}

	// Cache miss → 回源加载后重试
	if res.Code == -1 {
		_, err := s.loadKeyToCache(rawKey, tenantID)
		if err != nil {
			return nil, 0, err
		}
		// 重试 Lua
		raw, err = consumeLuaScript.Run(ctx, s.rdb, []string{ck}, amount, time.Now().UnixMilli()).Result()
		if err != nil {
			return nil, 0, fmt.Errorf("lua consume retry: %w", err)
		}
		if err := json.Unmarshal([]byte(raw.(string)), &res); err != nil {
			return nil, 0, fmt.Errorf("unmarshal lua result: %w", err)
		}
		if res.Code == -1 {
			// 回源后仍然 miss → key 不存在
			return nil, errcode.CodeKeyNotFound, nil
		}
	}

	// 业务错误
	if res.Code != 0 {
		return nil, res.Code, nil
	}

	// 发布 MQ 事件
	if s.mqSvc != nil {
		event := ConsumeEvent{
			EventID:        uuid.New().String(),
			KeyID:          res.KeyID,
			KeyAlias:       res.Alias,
			TenantID:       res.TenantID,
			Amount:         amount,
			RemainingAfter: res.Remaining,
			StatusAfter:    res.Status,
			Timestamp:      time.Now().UnixMilli(),
		}
		if err := s.mqSvc.PublishConsumeEvent(event); err != nil {
			zap.L().Error("发布 ConsumeEvent 失败", zap.Error(err))
			// MQ 失败不影响消费结果，Redis 已扣减
		}
	}

	status := model.KeyStatus(res.Status)
	return &ConsumeResult{
		RemainingAmount: res.Remaining,
		Status:          status,
		Exhausted:       res.Remaining <= 0,
	}, 0, nil
}

// consumeViaMySQL MySQL 降级路径：乐观锁重试
func (s *KeyService) consumeViaMySQL(rawKey string, amount int64, tenantID uint64) (*ConsumeResult, int, error) {
	const maxRetries = 3
	for retry := 0; retry < maxRetries; retry++ {
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
					"used_at": time.Now(),
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

		// 原子递增 Redis ZSET 中该 key 的消费计数
		if s.rdb != nil {
			ctx := context.Background()
			member := strconv.FormatUint(key.ID, 10)
			if err := s.rdb.ZIncrBy(ctx, topKeysZSetKey(tenantID), 1, member).Err(); err != nil {
				// Redis 故障不影响主流程，降级到 SQL 查询即可
				zap.L().Debug("top_keys ZINCRBY failed", zap.Error(err))
			}
		}

		return &ConsumeResult{
			RemainingAmount: newRemaining,
			Status:          status,
			Exhausted:       newRemaining <= 0,
		}, 0, nil
	}
	return nil, 0, fmt.Errorf("concurrency conflict after %d retries", maxRetries)
}

func (s *KeyService) ConsumeKeyByTenant(rawKey string, amount int64, tenantID uint64) (*ConsumeResult, int, error) {
	if amount <= 0 {
		return nil, 0, fmt.Errorf("invalid amount: %d", amount)
	}

	// Redis 可用时走快速路径
	if s.rdb != nil {
		result, code, err := s.consumeViaRedis(rawKey, amount, tenantID)
		if err == nil {
			return result, code, nil
		}
		zap.L().Warn("Redis consume failed, falling back to MySQL", zap.Error(err))
	}

	// 降级到 MySQL
	return s.consumeViaMySQL(rawKey, amount, tenantID)
}

type KeyListQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"`
	Search   string `form:"search"`
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
	if query.Search != "" {
		db = db.Where("alias LIKE ? OR key_suffix LIKE ?", "%"+query.Search+"%", "%"+query.Search+"%")
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
	Alias *string `json:"alias"`
}

func (s *KeyService) UpdateKey(id, tenantID uint64, req UpdateKeyRequest) error {
	if req.Alias == nil {
		return nil
	}
	return s.db.Model(&model.Key{}).Where("id = ? AND tenant_id = ?", id, tenantID).Update("alias", *req.Alias).Error
}

type AdjustBalanceRequest struct {
	Delta    int64  `json:"delta"`    // 正=增加, 负=减少
	Operator string `json:"operator"` // 操作人标识
	Remark   string `json:"remark"`   // 可选备注
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
			KeyID:          id,
			KeyAlias:       key.Alias,
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
	return nil
}

type ExportKeyItem struct {
	ID              uint64         `json:"id"`
	KeyPrefix       string         `json:"key_prefix"`
	KeySuffix       string         `json:"key_suffix"`
	Alias           string         `json:"alias"`
	RemainingAmount int64          `json:"remaining_amount"`
	Status          model.KeyStatus `json:"status"`
	CreatedAt       time.Time      `json:"created_at"`
	ExpireAt        *time.Time     `json:"expire_at"`
	MaxUsage        *int64         `json:"max_usage"`
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

// syncCacheOnBalanceChange 更新缓存中的 remaining 和 status（管理员调额后）
func (s *KeyService) syncCacheOnBalanceChange(keyHash string, remaining int64, status model.KeyStatus) {
	if s.rdb == nil {
		return
	}
	ctx := context.Background()
	ck := cacheKey(keyHash)
	if err := s.rdb.HSet(ctx, ck, map[string]interface{}{
		"remaining": remaining,
		"status":    string(status),
	}).Err(); err != nil {
		zap.L().Debug("syncCacheOnBalanceChange failed", zap.Error(err))
	}
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
