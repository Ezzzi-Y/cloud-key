package service

import (
	"CloudKey/internal/model"
	cryptorand "crypto/rand"
	"fmt"
	mathrand "math/rand"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type TenantService struct {
	db *gorm.DB
}

func NewTenantService(db *gorm.DB) *TenantService {
	return &TenantService{db: db}
}

type CreateTenantRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateTenantResult struct {
	Tenant        model.Tenant `json:"tenant"`
	AdminUsername string       `json:"admin_username"`
	AdminPassword string       `json:"admin_password"`
}

func (s *TenantService) CreateTenant(req CreateTenantRequest) (*CreateTenantResult, error) {
	username := fmt.Sprintf("%s_admin", req.Name)

	var count int64
	s.db.Model(&model.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		username = fmt.Sprintf("%s_admin%d", req.Name, mathrand.Intn(9999))
	}

	password := generateRandomPassword(16)

	tx := s.db.Begin()

	tenant := model.Tenant{
		Name:            req.Name,
		Status:          model.TenantStatusActive,
		KeyPrefix:       "sk-",
		KeyLength:       32,
		KeySuffixLength: 4,
	}
	if err := tx.Create(&tenant).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create tenant: %w", err)
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := model.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         model.RoleTenantAdmin,
		TenantID:     &tenant.ID,
		IsActive:     true,
	}
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create tenant admin: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &CreateTenantResult{
		Tenant:        tenant,
		AdminUsername: username,
		AdminPassword: password,
	}, nil
}

type TenantListItem struct {
	model.Tenant
	KeyCount  int64 `json:"key_count"`
	UserCount int64 `json:"user_count"`
}

func (s *TenantService) ListTenants() ([]TenantListItem, error) {
	tenants := make([]TenantListItem, 0)
	rows, err := s.db.Model(&model.Tenant{}).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t TenantListItem
		s.db.ScanRows(rows, &t.Tenant)

		s.db.Model(&model.Key{}).Where("tenant_id = ?", t.ID).Count(&t.KeyCount)
		s.db.Model(&model.User{}).Where("tenant_id = ?", t.ID).Count(&t.UserCount)

		tenants = append(tenants, t)
	}
	return tenants, nil
}

func (s *TenantService) GetTenant(id uint64) (*TenantListItem, error) {
	var t TenantListItem
	if err := s.db.First(&t.Tenant, id).Error; err != nil {
		return nil, err
	}
	s.db.Model(&model.Key{}).Where("tenant_id = ?", id).Count(&t.KeyCount)
	s.db.Model(&model.User{}).Where("tenant_id = ?", id).Count(&t.UserCount)
	return &t, nil
}

type UpdateTenantRequest struct {
	Name            *string             `json:"name"`
	Status          *model.TenantStatus `json:"status"`
	ExpireAt        *string             `json:"expire_at"`
	KeyPrefix       *string             `json:"key_prefix"`
	KeyLength       *int                `json:"key_length"`
	KeySuffixLength *int                `json:"key_suffix_length"`
}

func (s *TenantService) UpdateTenant(id uint64, req UpdateTenantRequest) error {
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.KeyPrefix != nil {
		updates["key_prefix"] = *req.KeyPrefix
	}
	if req.KeyLength != nil {
		updates["key_length"] = *req.KeyLength
	}
	if req.KeySuffixLength != nil {
		updates["key_suffix_length"] = *req.KeySuffixLength
	}
	if req.ExpireAt != nil {
		if *req.ExpireAt == "" {
			updates["expire_at"] = nil
		} else {
			updates["expire_at"] = *req.ExpireAt
		}
	}
	if len(updates) == 0 {
		return nil
	}
	return s.db.Model(&model.Tenant{}).Where("id = ?", id).Updates(updates).Error
}

func (s *TenantService) ResetPassword(tenantID uint64) (string, error) {
	newPass := generateRandomPassword(16)
	hash, _ := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	err := s.db.Model(&model.User{}).Where("tenant_id = ? AND role = ?", tenantID, model.RoleTenantAdmin).
		Update("password_hash", string(hash)).Error
	if err != nil {
		return "", err
	}
	return newPass, nil
}

func generateRandomPassword(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$"
	bytes := make([]byte, length)
	if _, err := cryptorand.Read(bytes); err != nil {
		panic("failed to generate random password: " + err.Error())
	}
	for i, b := range bytes {
		bytes[i] = chars[int(b)%len(chars)]
	}
	return string(bytes)
}
