package service

import (
	"CloudKey/internal/model"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"gorm.io/gorm"
)

type ServiceAccountService struct {
	db *gorm.DB
}

func NewServiceAccountService(db *gorm.DB) *ServiceAccountService {
	return &ServiceAccountService{db: db}
}

var hmacServiceKeySecret = []byte("cloudkey-service-account-hmac-secret")

func hashServiceKey(key string) string {
	h := hmac.New(sha256.New, hmacServiceKeySecret)
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *ServiceAccountService) GenerateServiceKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "svc-" + hex.EncodeToString(bytes), nil
}

func (s *ServiceAccountService) CreateServiceAccount(name string, tenantID uint64) (*model.ServiceAccount, string, error) {
	rawKey, err := s.GenerateServiceKey()
	if err != nil {
		return nil, "", err
	}

	account := model.ServiceAccount{Name: name, KeyHash: hashServiceKey(rawKey), TenantID: tenantID}
	if err := s.db.Create(&account).Error; err != nil {
		return nil, "", fmt.Errorf("create service account: %w", err)
	}
	return &account, rawKey, nil
}

func (s *ServiceAccountService) ValidateServiceKey(serviceKey string) (*model.ServiceAccount, error) {
	if serviceKey == "" {
		return nil, nil
	}

	keyHash := hashServiceKey(serviceKey)
	var account model.ServiceAccount
	if err := s.db.Where("key_hash = ? AND is_active = ?", keyHash, true).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (s *ServiceAccountService) ListServiceAccounts(tenantID uint64) ([]model.ServiceAccount, error) {
	var accounts []model.ServiceAccount
	if err := s.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (s *ServiceAccountService) ToggleServiceAccount(id, tenantID uint64, isActive bool) error {
	return s.db.Model(&model.ServiceAccount{}).Where("id = ? AND tenant_id = ?", id, tenantID).Update("is_active", isActive).Error
}

func (s *ServiceAccountService) DeleteServiceAccount(id, tenantID uint64) error {
	return s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.ServiceAccount{}).Error
}
