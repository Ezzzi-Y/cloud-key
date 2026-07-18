package service

import (
	"CloudKey/internal/model"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

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
