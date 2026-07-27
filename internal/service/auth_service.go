package service

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	pendingTOTPPrefix   = "totp:pending:"
	preAuthTokenPrefix  = "auth:pretoken:"
	lockoutPrefix       = "auth:lock:"
	failedAttemptPrefix = "auth:fail:"
	lockoutThreshold    = 5
	lockoutDuration     = 10 * time.Minute
	preAuthTokenTTL     = 5 * time.Minute
)

type AuthService struct {
	db          *gorm.DB
	rdb         *redis.Client
	jwtSecret   string
	jwtExpHours int
}

func NewAuthService(db *gorm.DB, rdb *redis.Client, jwtSecret string, jwtExpHours int) *AuthService {
	return &AuthService{db: db, rdb: rdb, jwtSecret: jwtSecret, jwtExpHours: jwtExpHours}
}

// ========== Lockout & Pre-auth Token Lua scripts ==========

var lockoutCheckScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 1 then
	return redis.call('PTTL', KEYS[1])
end
return -1
`)

var failRecordScript = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
redis.call('EXPIRE', KEYS[1], ARGV[1])
if count >= tonumber(ARGV[2]) then
	redis.call('SET', KEYS[2], '1', 'EX', ARGV[3])
	redis.call('DEL', KEYS[1])
	return 1
end
return 0
`)

// ========== Lockout helpers ==========

func (s *AuthService) checkLockout(userID uint64) error {
	ctx := context.Background()
	key := fmt.Sprintf("%s%d", lockoutPrefix, userID)

	result, err := lockoutCheckScript.Run(ctx, s.rdb, []string{key}).Int64()
	if err != nil && err != redis.Nil {
		return nil // Redis error, fail open
	}

	if result >= 0 {
		remainingSec := result / 1000
		if remainingSec <= 0 {
			remainingSec = 1
		}
		return fmt.Errorf("%s", errcode.GetMessage(errcode.CodeAccountLocked))
	}

	return nil
}

func (s *AuthService) recordFailure(userID uint64) {
	ctx := context.Background()
	failKey := fmt.Sprintf("%s%d", failedAttemptPrefix, userID)
	lockKey := fmt.Sprintf("%s%d", lockoutPrefix, userID)

	failRecordScript.Run(ctx, s.rdb, []string{failKey, lockKey},
		int(lockoutDuration.Seconds()), lockoutThreshold, int(lockoutDuration.Seconds()))
}

func (s *AuthService) clearFailedAttempts(userID uint64) {
	ctx := context.Background()
	key := fmt.Sprintf("%s%d", failedAttemptPrefix, userID)
	s.rdb.Del(ctx, key)
}

// ========== Pre-auth Token ==========

func (s *AuthService) GeneratePreAuthToken(userID uint64) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)

	ctx := context.Background()
	key := fmt.Sprintf("%s%d:%s", preAuthTokenPrefix, userID, token)
	if err := s.rdb.Set(ctx, key, userID, preAuthTokenTTL).Err(); err != nil {
		return "", fmt.Errorf("store pre-auth token: %w", err)
	}

	return token, nil
}

func (s *AuthService) ValidatePreAuthToken(userID uint64, token string) (uint64, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s%d:%s", preAuthTokenPrefix, userID, token)

	stored, err := s.rdb.GetDel(ctx, key).Result()
	if err == redis.Nil {
		return 0, fmt.Errorf("%s", errcode.GetMessage(errcode.CodePreAuthInvalid))
	}
	if err != nil {
		return 0, fmt.Errorf("validate pre-auth token: %w", err)
	}

	var storedUserID uint64
	if _, err := fmt.Sscan(stored, &storedUserID); err != nil {
		return 0, fmt.Errorf("invalid pre-auth token data")
	}

	if storedUserID != userID {
		return 0, fmt.Errorf("%s", errcode.GetMessage(errcode.CodePreAuthInvalid))
	}

	return storedUserID, nil
}

// GenerateLoginJWT loads user info and issues a JWT without requiring pre-auth token validation.
// Used after the pre-auth token has already been validated at the handler level.
func (s *AuthService) GenerateLoginJWT(userID uint64) (*LoginResponse, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
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
		return nil, fmt.Errorf("sign JWT: %w", err)
	}

	resp := &LoginResponse{
		Token:    tokenString,
		Role:     user.Role,
		TenantID: user.TenantID,
		Username: user.Username,
	}
	resp.fillTenantInfo(s.db)
	return resp, nil
}

type LoginResult struct {
	RequireTOTP  bool    `json:"require_totp"`
	UserID       uint64  `json:"user_id"`
	TenantID     *uint64 `json:"tenant_id"`
	PreAuthToken string  `json:"pre_auth_token"`
}

func (s *AuthService) Login(username, password string) (*LoginResult, error) {
	var user model.User
	if err := s.db.Where("username = ? AND is_active = ?", username, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	// Check account lockout
	if err := s.checkLockout(user.ID); err != nil {
		s.recordFailure(user.ID)
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.recordFailure(user.ID)
		return nil, nil
	}

	// Password correct — clear failed attempts and generate pre-auth token
	s.clearFailedAttempts(user.ID)

	preAuthToken, err := s.GeneratePreAuthToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate pre-auth token: %w", err)
	}

	return &LoginResult{
		RequireTOTP:  user.TotpSetup,
		UserID:       user.ID,
		TenantID:     user.TenantID,
		PreAuthToken: preAuthToken,
	}, nil
}

type AuthClaims struct {
	UserID   uint64         `json:"user_id"`
	Username string         `json:"username"`
	Role     model.UserRole `json:"role"`
	TenantID *uint64        `json:"tenant_id"`
	jwt.RegisteredClaims
}

func (s *AuthService) VerifyTOTP(userID uint64, code string, preAuthToken string) (string, *LoginResponse, error) {
	// Validate pre-auth token (single-use)
	if _, err := s.ValidatePreAuthToken(userID, preAuthToken); err != nil {
		return "", nil, nil
	}

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
	resp.fillTenantInfo(s.db)
	return tokenString, resp, nil
}

type LoginResponse struct {
	Token          string          `json:"token"`
	Role           model.UserRole  `json:"role"`
	TenantID       *uint64         `json:"tenant_id"`
	Username       string          `json:"username"`
	TenantStatus   *model.TenantStatus `json:"tenant_status,omitempty"`
	TenantExpireAt *time.Time      `json:"tenant_expire_at,omitempty"`
}

// fillTenantInfo looks up the tenant by TenantID and populates TenantStatus and TenantExpireAt.
func (r *LoginResponse) fillTenantInfo(db *gorm.DB) {
	if r.TenantID == nil {
		return
	}
	var tenant model.Tenant
	if err := db.Select("status", "expire_at").First(&tenant, *r.TenantID).Error; err != nil {
		return
	}
	r.TenantStatus = &tenant.Status
	r.TenantExpireAt = tenant.ExpireAt
}

// GenerateTOTPSecret generates a new TOTP key and stores it in Redis as pending.
// The existing DB secret remains valid until ConfirmTOTPSetup succeeds.
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

	secret := key.Secret()
	ctx := context.Background()
	if err := s.rdb.Set(ctx, pendingTOTPPrefix+fmt.Sprint(userID), secret, 10*time.Minute).Err(); err != nil {
		return "", "", fmt.Errorf("store pending TOTP: %w", err)
	}

	return secret, key.URL(), nil
}

// ConfirmTOTPSetup validates the code against the pending secret in Redis,
// then persists it to DB and removes the pending entry.
func (s *AuthService) ConfirmTOTPSetup(userID uint64, code string) error {
	ctx := context.Background()
	key := pendingTOTPPrefix + fmt.Sprint(userID)

	secret, err := s.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("no pending TOTP setup, please generate a new key first")
	}
	if err != nil {
		return fmt.Errorf("get pending TOTP: %w", err)
	}

	if !totp.Validate(code, secret) {
		return fmt.Errorf("TOTP code invalid")
	}

	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}
	if err := s.db.Model(&user).Updates(map[string]interface{}{
		"totp_secret": secret,
		"totp_setup":  true,
	}).Error; err != nil {
		return err
	}

	// Clean up pending entry
	s.rdb.Del(ctx, key)
	return nil
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
