package service

import (
	"CloudKey/internal/model"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminService struct {
	db          *gorm.DB
	jwtSecret   string
	jwtExpHours int
}

func NewAdminService(db *gorm.DB, jwtSecret string, jwtExpHours int) *AdminService {
	return &AdminService{db: db, jwtSecret: jwtSecret, jwtExpHours: jwtExpHours}
}

type LoginResult struct {
	RequireTOTP bool   `json:"require_totp"`
	AdminID     uint64 `json:"admin_id"`
}

func (s *AdminService) Login(username, password string) (*LoginResult, error) {
	var admin model.Admin
	if err := s.db.Where("username = ? AND is_active = ?", username, true).First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return nil, nil
	}

	return &LoginResult{RequireTOTP: admin.TotpSetup, AdminID: admin.ID}, nil
}

func (s *AdminService) VerifyTOTP(adminID uint64, code string) (string, error) {
	var admin model.Admin
	if err := s.db.First(&admin, adminID).Error; err != nil {
		return "", fmt.Errorf("admin not found: %w", err)
	}

	if !admin.TotpSetup {
		return "", fmt.Errorf("TOTP not set up")
	}

	if !totp.Validate(code, admin.TotpSecret) {
		return "", nil
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin_id": admin.ID,
		"username": admin.Username,
		"exp":      time.Now().Add(time.Duration(s.jwtExpHours) * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("sign JWT: %w", err)
	}
	return tokenString, nil
}

func (s *AdminService) GenerateTOTPSecret(adminID uint64) (string, string, error) {
	var admin model.Admin
	if err := s.db.First(&admin, adminID).Error; err != nil {
		return "", "", err
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "CloudKey",
		AccountName: admin.Username,
	})
	if err != nil {
		return "", "", fmt.Errorf("generate TOTP: %w", err)
	}

	if err := s.db.Model(&admin).Updates(map[string]interface{}{
		"totp_secret": key.Secret(),
		"totp_setup":  false,
	}).Error; err != nil {
		return "", "", err
	}

	return key.Secret(), key.URL(), nil
}

func (s *AdminService) ConfirmTOTPSetup(adminID uint64, code string) error {
	var admin model.Admin
	if err := s.db.First(&admin, adminID).Error; err != nil {
		return err
	}
	if !totp.Validate(code, admin.TotpSecret) {
		return fmt.Errorf("TOTP code invalid")
	}
	return s.db.Model(&admin).Update("totp_setup", true).Error
}

func (s *AdminService) ChangePassword(adminID uint64, oldPass, newPass string) error {
	var admin model.Admin
	if err := s.db.First(&admin, adminID).Error; err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(oldPass)); err != nil {
		return fmt.Errorf("old password incorrect")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return s.db.Model(&admin).Update("password_hash", string(hash)).Error
}

func (s *AdminService) GetAdminProfile(adminID uint64) (*model.Admin, error) {
	var admin model.Admin
	if err := s.db.First(&admin, adminID).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (s *AdminService) SeedAdmin(username, password string) error {
	var count int64
	s.db.Model(&model.Admin{}).Count(&count)
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.db.Create(&model.Admin{
		Username:     username,
		PasswordHash: string(hash),
		TotpSetup:    false,
		IsActive:     true,
	}).Error
}
