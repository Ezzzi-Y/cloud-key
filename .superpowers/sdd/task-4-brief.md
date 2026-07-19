### Task 4: 创建 AuthService（统一登录）

**Files:**
- Create: `internal/service/auth_service.go`
- Rewrite: `internal/service/admin_service.go` → `internal/service/user_service.go`

**Interfaces:**
- Produces: `AuthService` with `Login(username, password)`, `VerifyTOTP(userID, code)`, `GenerateTOTPSecret(userID)`, `ConfirmTOTPSetup(userID, code)`, `ChangePassword(userID, oldPass, newPass)`, `GetProfile(userID)`, `SeedSuperAdmin(username, password)`
- JWT Claims: `{user_id, username, role, tenant_id, exp, iat}`

- [ ] **Step 1: 创建 `internal/service/auth_service.go`**

```go
package service

import (
	"CloudKey/internal/model"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db          *gorm.DB
	jwtSecret   string
	jwtExpHours int
}

func NewAuthService(db *gorm.DB, jwtSecret string, jwtExpHours int) *AuthService {
	return &AuthService{db: db, jwtSecret: jwtSecret, jwtExpHours: jwtExpHours}
}

type LoginResult struct {
	RequireTOTP bool   `json:"require_totp"`
	UserID      uint64 `json:"user_id"`
}

func (s *AuthService) Login(username, password string) (*LoginResult, error) {
	var user model.User
	if err := s.db.Where("username = ? AND is_active = ?", username, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil
	}

	return &LoginResult{RequireTOTP: user.TotpSetup, UserID: user.ID}, nil
}

type AuthClaims struct {
	UserID   uint64         `json:"user_id"`
	Username string         `json:"username"`
	Role     model.UserRole `json:"role"`
	TenantID *uint64        `json:"tenant_id"`
	jwt.RegisteredClaims
}

func (s *AuthService) VerifyTOTP(userID uint64, code string) (string, *LoginResponse, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return "", nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.TotpSetup {
		return "", nil, nil
	}

	if !totp.Validate(code, user.TotpSecret) {
		return "", nil, nil
	}

	claims := AuthClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		TenantID: user.TenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwtExpHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, fmt.Errorf("sign JWT: %w", err)
	}

	resp := &LoginResponse{
		Token:    tokenString,
		Role:     user.Role,
		TenantID: user.TenantID,
		Username: user.Username,
	}
	return tokenString, resp, nil
}

type LoginResponse struct {
	Token    string         `json:"token"`
	Role     model.UserRole `json:"role"`
	TenantID *uint64        `json:"tenant_id"`
	Username string         `json:"username"`
}

func (s *AuthService) GenerateTOTPSecret(userID uint64) (string, string, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return "", "", err
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "CloudKey",
		AccountName: user.Username,
	})
	if err != nil {
		return "", "", fmt.Errorf("generate TOTP: %w", err)
	}

	if err := s.db.Model(&user).Updates(map[string]interface{}{
		"totp_secret": key.Secret(),
		"totp_setup":  false,
	}).Error; err != nil {
		return "", "", err
	}

	return key.Secret(), key.URL(), nil
}

func (s *AuthService) ConfirmTOTPSetup(userID uint64, code string) error {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}
	if !totp.Validate(code, user.TotpSecret) {
		return fmt.Errorf("TOTP code invalid")
	}
	return s.db.Model(&user).Update("totp_setup", true).Error
}

func (s *AuthService) ChangePassword(userID uint64, oldPass, newPass string) error {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPass)); err != nil {
		return fmt.Errorf("old password incorrect")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return s.db.Model(&user).Update("password_hash", string(hash)).Error
}

func (s *AuthService) GetProfile(userID uint64) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) SeedSuperAdmin(username, password string) error {
	var count int64
	s.db.Model(&model.User{}).Where("role = ?", model.RoleSuperAdmin).Count(&count)
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.db.Create(&model.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         model.RoleSuperAdmin,
		TotpSetup:    false,
		IsActive:     true,
	}).Error
}

func generateTempToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
```

- [ ] **Step 2: 保留旧 admin_service.go 文件（后续 task 删除或归档），无需操作。此 task 只创建新文件 auth_service.go。**

- [ ] **Step 3: 验证编译通过**

Run: `cd D:/MyGoProject/CloudKey && go build ./...`
Expected: 编译通过

- [ ] **Step 4: Commit**

```bash
git add internal/service/auth_service.go
git commit -m "feat: add AuthService with unified login + JWT with role+tenant_id"
```

---

