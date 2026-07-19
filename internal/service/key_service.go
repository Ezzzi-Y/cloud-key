package service

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type KeyService struct {
	db        *gorm.DB
	keyPrefix string
	keyLength int
	suffixLen int
}

func NewKeyService(db *gorm.DB) *KeyService {
	return &KeyService{
		db:        db,
		keyPrefix: "sk-",
		keyLength: 16,
		suffixLen: 4,
	}
}

func (s *KeyService) WithConfig(prefix string, keyLen, suffixLen int) *KeyService {
	s.keyPrefix = prefix
	s.keyLength = keyLen
	s.suffixLen = suffixLen
	return s
}

func (s *KeyService) generateRawKey() (string, error) {
	bytes := make([]byte, s.keyLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random: %w", err)
	}
	return s.keyPrefix + hex.EncodeToString(bytes), nil
}

func (s *KeyService) hashKey(rawKey string) string {
	h := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(h[:])
}

type CreateKeyRequest struct {
	Alias         string               `json:"alias"`
	BillingMode   model.KeyBillingMode `json:"billing_mode"`
	InitialAmount int64                `json:"initial_amount"`
	CreatedBy     string               `json:"created_by"`
	ExpireAt      *time.Time           `json:"expire_at"`
	MaxUsage      *int64               `json:"max_usage"`
}

type CreateKeyResult struct {
	RawKey string    `json:"raw_key"`
	Key    model.Key `json:"key"`
}

func (s *KeyService) CreateKey(req CreateKeyRequest) (*CreateKeyResult, error) {
	rawKey, err := s.generateRawKey()
	if err != nil {
		return nil, err
	}

	suffixLen := s.suffixLen
	if len(rawKey) < suffixLen {
		suffixLen = len(rawKey)
	}
	suffix := rawKey[len(rawKey)-suffixLen:]

	keyHash := s.hashKey(rawKey)

	key := model.Key{
		Alias:           req.Alias,
		KeyHash:         keyHash,
		KeyPrefix:       s.keyPrefix,
		KeySuffix:       suffix,
		BillingMode:     req.BillingMode,
		InitialAmount:   req.InitialAmount,
		RemainingAmount: req.InitialAmount,
		Version:         0,
		Status:          model.KeyStatusUnused,
		CreatedBy:       req.CreatedBy,
		ExpireAt:        req.ExpireAt,
		MaxUsage:        req.MaxUsage,
	}

	if err := s.db.Create(&key).Error; err != nil {
		return nil, fmt.Errorf("create key: %w", err)
	}

	return &CreateKeyResult{RawKey: rawKey, Key: key}, nil
}

func (s *KeyService) FindByRawKey(rawKey string) (*model.Key, error) {
	keyHash := s.hashKey(rawKey)
	var key model.Key
	if err := s.db.Where("key_hash = ?", keyHash).First(&key).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &key, nil
}

type KeyStatusResult struct {
	Alias           string               `json:"alias"`
	BillingMode     model.KeyBillingMode `json:"billing_mode"`
	RemainingAmount int64                `json:"remaining_amount"`
	Status          model.KeyStatus      `json:"status"`
	CreatedAt       string               `json:"created_at"`
	UsedAt          *string              `json:"used_at"`
}

func (s *KeyService) GetKeyStatus(rawKey string) (*KeyStatusResult, error) {
	key, err := s.FindByRawKey(rawKey)
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
		BillingMode:     key.BillingMode,
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

func (s *KeyService) ConsumeKey(rawKey string, amount int64) (*ConsumeResult, int, error) {
	if amount <= 0 {
		return nil, 0, fmt.Errorf("invalid amount: %d", amount)
	}

	const maxRetries = 3
	for retry := 0; retry < maxRetries; retry++ {
		// Re-fetch key each attempt to get the latest version
		key, err := s.FindByRawKey(rawKey)
		if err != nil {
			return nil, 0, err
		}
		if key == nil {
			return nil, errcode.CodeKeyNotFound, nil
		}
		if key.Status == model.KeyStatusDisabled {
			return nil, errcode.CodeKeyDisabled, nil
		}
		if key.Status == model.KeyStatusUsed {
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
			Where("id = ? AND version = ?", key.ID, key.Version).
			Updates(map[string]interface{}{
				"remaining_amount": gorm.Expr("remaining_amount - ?", amount),
				"version":          gorm.Expr("version + 1"),
				"status":           gorm.Expr("CASE WHEN remaining_amount - ? <= 0 THEN 'used' ELSE status END", amount),
				"used_at":          gorm.Expr("CASE WHEN remaining_amount - ? <= 0 THEN NOW() ELSE used_at END", amount),
			})

		if result.Error != nil {
			tx.Rollback()
			return nil, 0, result.Error
		}
		if result.RowsAffected == 0 {
			tx.Rollback()
			continue // optimistic lock conflict, retry with fresh key
		}

		var updatedKey model.Key
		if err := tx.Where("id = ?", key.ID).First(&updatedKey).Error; err != nil {
			tx.Rollback()
			return nil, 0, err
		}

		if err := tx.Commit().Error; err != nil {
			return nil, 0, err
		}

		return &ConsumeResult{
			RemainingAmount: updatedKey.RemainingAmount,
			Status:          updatedKey.Status,
			Exhausted:       updatedKey.Status == model.KeyStatusUsed,
		}, 0, nil
	}
	return nil, 0, fmt.Errorf("concurrency conflict after %d retries", maxRetries)
}

type KeyListQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"`
	Search   string `form:"search"`
}

func (s *KeyService) GetKeyDetail(id uint64) (*model.Key, error) {
	var key model.Key
	if err := s.db.First(&key, id).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (s *KeyService) ListKeys(query KeyListQuery) ([]model.Key, int64, error) {
	var keys []model.Key
	var total int64

	db := s.db.Model(&model.Key{})
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

func (s *KeyService) ListKeysByCreatedBy(createdBy string, page, pageSize int) ([]model.Key, int64, error) {
	var keys []model.Key
	var total int64

	db := s.db.Model(&model.Key{}).Where("created_by = ?", createdBy)
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
	RemainingAmount *int64  `json:"remaining_amount"`
}

func (s *KeyService) UpdateKey(id uint64, req UpdateKeyRequest) error {
	updates := map[string]interface{}{}
	if req.Alias != nil {
		updates["alias"] = *req.Alias
	}
	if req.RemainingAmount != nil {
		updates["remaining_amount"] = *req.RemainingAmount
		updates["status"] = gorm.Expr(
			"CASE WHEN ? > 0 AND status = 'used' THEN 'unused' ELSE status END",
			*req.RemainingAmount,
		)
	}
	if len(updates) == 0 {
		return nil
	}
	return s.db.Model(&model.Key{}).Where("id = ?", id).Updates(updates).Error
}

func (s *KeyService) DisableKey(id uint64) error {
	return s.db.Model(&model.Key{}).Where("id = ?", id).Update("status", model.KeyStatusDisabled).Error
}

func (s *KeyService) EnableKey(id uint64) error {
	return s.db.Model(&model.Key{}).Where("id = ?", id).Update("status", model.KeyStatusUnused).Error
}

func (s *KeyService) DeleteKey(id uint64) error {
	return s.db.Delete(&model.Key{}, id).Error
}

type ExportKeyItem struct {
	ID              uint64              `json:"id"`
	KeyPrefix       string              `json:"key_prefix"`
	KeySuffix       string              `json:"key_suffix"`
	Alias           string              `json:"alias"`
	BillingMode     model.KeyBillingMode `json:"billing_mode"`
	InitialAmount   int64               `json:"initial_amount"`
	RemainingAmount int64               `json:"remaining_amount"`
	Status          model.KeyStatus     `json:"status"`
	CreatedAt       time.Time           `json:"created_at"`
	ExpireAt        *time.Time          `json:"expire_at"`
	MaxUsage        *int64              `json:"max_usage"`
}

func (s *KeyService) ExportKeysJSON() ([]ExportKeyItem, error) {
	var keys []model.Key
	if err := s.db.Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, err
	}

	items := make([]ExportKeyItem, len(keys))
	for i, k := range keys {
		items[i] = ExportKeyItem{
			ID:              k.ID,
			KeyPrefix:       k.KeyPrefix,
			KeySuffix:       k.KeySuffix,
			Alias:           k.Alias,
			BillingMode:     k.BillingMode,
			InitialAmount:   k.InitialAmount,
			RemainingAmount: k.RemainingAmount,
			Status:          k.Status,
			CreatedAt:       k.CreatedAt,
			ExpireAt:        k.ExpireAt,
			MaxUsage:        k.MaxUsage,
		}
	}
	return items, nil
}

func (s *KeyService) ExportKeys() ([]model.Key, error) {
	var keys []model.Key
	if err := s.db.Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}
