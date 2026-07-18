# CloudKey Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 CloudKey 卡密管理平台的 encryption、models、api 三层，完成从卡密创建到验证扣减的完整业务流程。

**Architecture:** 分层架构 — `internal/crypto` 提供加密原语（哈希、AES-GCM 加密、密钥派生），`internal/models` 定义 GORM 数据模型，`internal/api` 使用 Gin 框架实现 RESTful API（公开接口 + 管理接口 + 服务接口）。认证采用 JWT + TOTP 双因素验证。

**Tech Stack:** Go 1.25, Gin, GORM, golang.org/x/crypto (argon2/bcrypt), golang-jwt/v5, pquerna/otp

## File Structure

```
internal/
├── config/          # 已完成 - 配置加载
├── log/             # 已完成 - 日志初始化
├── database/        # 已完成 - 数据库连接
├── crypto/          # 新增 - 加密原语
│   ├── key.go       # 卡密生成（GenerateKey）
│   ├── hash.go      # 哈希函数（HashKey argon2id, HashPassword bcrypt）
│   ├── encrypt.go   # AES-256-GCM 加解密
│   └── kdf.go       # HKDF-SHA256 密钥派生
├── models/          # 新增 - GORM 数据模型
│   ├── key.go       # 卡密模型
│   ├── usage_log.go # 使用记录模型
│   ├── admin.go     # 管理员模型
│   ├── service_account.go # 服务账号模型
│   ├── login_log.go # 登录日志模型
│   └── config.go    # 系统配置模型
└── api/             # 新增 - HTTP API 层
    ├── response.go  # 统一响应结构 + 错误码
    ├── middleware/   # 中间件
    │   ├── auth.go   # JWT 管理员认证
    │   ├── service.go # 服务账号认证
    │   └── cors.go   # CORS 配置
    ├── handler/      # 请求处理器
    │   ├── key.go    # 公开卡密接口（status, consume）
    │   ├── admin_auth.go  # 管理员认证（login, 2FA）
    │   ├── admin_key.go   # 管理员卡密 CRUD
    │   ├── admin_log.go   # 使用记录查询
    │   ├── admin_stats.go # 数据统计
    │   └── service.go     # 服务账号接口
    └── router.go     # 路由注册 + AutoMigrate
cmd/
└── cloudkey/
    └── main.go      # 程序入口
config.yaml          # 示例配置文件
```

## Global Constraints

- Go module: `CloudKey`
- 仅支持 MySQL 8.0+，数据库类型字段固定为 `mysql`
- 卡密格式：可配置前缀（默认 `sk-`）+ 随机字符串（默认 32 字符）+ 3-5 位后缀
- 卡密存储：argon2id 哈希，不明文存储；后缀用于脱敏显示
- 管理员密码：bcrypt 哈希存储
- 加密：AES-256-GCM，密钥由 HKDF-SHA256 从 master key 派生
- 认证：管理员用 JWT + TOTP，服务账号用 X-Service-Key header
- 并发安全：扣减操作使用数据库事务 + 乐观锁（version 字段）
- 所有测试使用 `t.TempDir()` 或内存数据库，不依赖外部 MySQL

---

## Task 1: Encryption Package

**Files:**
- Create: `internal/crypto/key.go`
- Create: `internal/crypto/key_test.go`
- Create: `internal/crypto/hash.go`
- Create: `internal/crypto/hash_test.go`
- Create: `internal/crypto/encrypt.go`
- Create: `internal/crypto/encrypt_test.go`
- Create: `internal/crypto/kdf.go`
- Create: `internal/crypto/kdf_test.go`

**Dependencies to add:**
```
go get golang.org/x/crypto
```

**Interfaces:**
- Produces: `GenerateKey(prefix string, length int) (plainKey, hash, suffix string, err error)` — 供 Task 5 创建卡密使用
- Produces: `HashKey(key string) string` — argon2id 哈希，供验证卡密
- Produces: `HashPassword(password string) (string, error)` / `VerifyPassword(password, hash string) bool` — bcrypt，供 Task 6 管理员密码
- Produces: `Encrypt(plaintext []byte, key []byte) ([]byte, error)` / `Decrypt(ciphertext []byte, key []byte) ([]byte, error)` — AES-256-GCM
- Produces: `DeriveKey(masterKey []byte, context string) []byte` — HKDF-SHA256

### Step 1: Write kdf.go

```go
// internal/crypto/kdf.go
package crypto

import (
	"crypto/hkdf"
	"crypto/sha256"
	"io"
)

// DeriveKey 使用 HKDF-SHA256 从 master key 派生子密钥
func DeriveKey(masterKey []byte, context string) []byte {
	key := make([]byte, 32) // AES-256
	hkdf.New(sha256.New, masterKey, nil, []byte(context)).Read(key)
	return key
}
```

### Step 2: Write kdf_test.go

```go
// internal/crypto/kdf_test.go
package crypto

import (
	"bytes"
	"testing"
)

func TestDeriveKey_Deterministic(t *testing.T) {
	master := []byte("test-master-key")
	k1 := DeriveKey(master, "encrypt")
	k2 := DeriveKey(master, "encrypt")
	if !bytes.Equal(k1, k2) {
		t.Error("same input should produce same key")
	}
}

func TestDeriveKey_DifferentContext(t *testing.T) {
	master := []byte("test-master-key")
	k1 := DeriveKey(master, "encrypt")
	k2 := DeriveKey(master, "sign")
	if bytes.Equal(k1, k2) {
		t.Error("different context should produce different key")
	}
}

func TestDeriveKey_Length(t *testing.T) {
	key := DeriveKey([]byte("master"), "ctx")
	if len(key) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(key))
	}
}
```

### Step 3: Run tests to verify they pass

```bash
go test ./internal/crypto/ -run TestDerive -v
```

### Step 4: Write encrypt.go

```go
// internal/crypto/encrypt.go
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// Encrypt 使用 AES-256-GCM 加密
func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt 使用 AES-256-GCM 解密
func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
```

### Step 5: Write encrypt_test.go

```go
// internal/crypto/encrypt_test.go
package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := DeriveKey([]byte("master-key"), "encrypt")
	plaintext := []byte("hello cloudkey")

	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("decrypted text doesn't match original")
	}
}

func TestEncrypt_DifferentCiphertexts(t *testing.T) {
	key := DeriveKey([]byte("master"), "ctx")
	plaintext := []byte("same text")

	c1, _ := Encrypt(plaintext, key)
	c2, _ := Encrypt(plaintext, key)
	if bytes.Equal(c1, c2) {
		t.Error("same plaintext should produce different ciphertexts (random nonce)")
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	key1 := DeriveKey([]byte("master1"), "ctx")
	key2 := DeriveKey([]byte("master2"), "ctx")

	encrypted, _ := Encrypt([]byte("secret"), key1)
	_, err := Decrypt(encrypted, key2)
	if err == nil {
		t.Error("expected error with wrong key")
	}
}

func TestDecrypt_ShortCiphertext(t *testing.T) {
	key := DeriveKey([]byte("master"), "ctx")
	_, err := Decrypt([]byte("short"), key)
	if err == nil {
		t.Error("expected error for short ciphertext")
	}
}
```

### Step 6: Run encrypt tests

```bash
go test ./internal/crypto/ -run TestEncrypt -v
```

### Step 7: Write hash.go

```go
// internal/crypto/hash.go
package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// HashKey 使用 argon2id 对卡密进行哈希
func HashKey(key string) string {
	salt := make([]byte, 16)
	rand.Read(salt)
	hash := argon2.IDKey([]byte(key), salt, 1, 64*1024, 4, 32)
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=1,p=4$%s$%s",
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash))
}

// VerifyKey 验证卡密是否匹配 argon2id 哈希
func VerifyKey(key, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	salt, _ := base64.RawStdEncoding.DecodeString(parts[4])
	expectedHash, _ := base64.RawStdEncoding.DecodeString(parts[5])
	hash := argon2.IDKey([]byte(key), salt, 1, 64*1024, 4, 32)
	return subtle.ConstantTimeCompare(hash, expectedHash) == 1
}

// HashPassword 使用 bcrypt 对管理员密码进行哈希
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// VerifyPassword 验证管理员密码
func VerifyPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
```

### Step 8: Write hash_test.go

```go
// internal/crypto/hash_test.go
package crypto

import "testing"

func TestHashKey_Verify(t *testing.T) {
	key := "sk-abc123def456"
	hash := HashKey(key)
	if !VerifyKey(key, hash) {
		t.Error("valid key should verify")
	}
}

func TestHashKey_WrongKey(t *testing.T) {
	hash := HashKey("sk-abc123")
	if VerifyKey("sk-wrong", hash) {
		t.Error("wrong key should not verify")
	}
}

func TestHashPassword_Verify(t *testing.T) {
	hash, err := HashPassword("admin123")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword("admin123", hash) {
		t.Error("valid password should verify")
	}
}

func TestHashPassword_WrongPassword(t *testing.T) {
	hash, _ := HashPassword("admin123")
	if VerifyPassword("wrong", hash) {
		t.Error("wrong password should not verify")
	}
}
```

### Step 9: Run all hash tests

```bash
go test ./internal/crypto/ -run TestHash -v
```

### Step 10: Write key.go

```go
// internal/crypto/key.go
package crypto

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateKey 生成卡密
// prefix: 前缀（如 "sk-"）
// length: 随机部分长度
// 返回：明文卡密、argon2id 哈希、后缀（4位）
func GenerateKey(prefix string, length int) (plainKey, hash, suffix string, err error) {
	randomBytes := make([]byte, length/2+1)
	if _, err = rand.Read(randomBytes); err != nil {
		return
	}
	randomPart := hex.EncodeToString(randomBytes)[:length]
	plainKey = prefix + randomPart
	hash = HashKey(plainKey)

	suffixLen := 4
	if len(randomPart) < suffixLen {
		suffixLen = len(randomPart)
	}
	suffix = randomPart[len(randomPart)-suffixLen:]
	return
}
```

### Step 11: Write key_test.go

```go
// internal/crypto/key_test.go
package crypto

import (
	"strings"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	key, hash, suffix, err := GenerateKey("sk-", 32)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(key, "sk-") {
		t.Error("key should have prefix")
	}
	if len(key) != 35 { // "sk-" + 32 chars
		t.Errorf("expected length 35, got %d", len(key))
	}
	if len(suffix) != 4 {
		t.Errorf("suffix should be 4 chars, got %d", len(suffix))
	}
	if !VerifyKey(key, hash) {
		t.Error("generated key should match its hash")
	}
	if !strings.HasSuffix(key[len(key)-4:], suffix) {
		t.Error("suffix should be last 4 chars of random part")
	}
}

func TestGenerateKey_Unique(t *testing.T) {
	key1, _, _, _ := GenerateKey("sk-", 32)
	key2, _, _, _ := GenerateKey("sk-", 32)
	if key1 == key2 {
		t.Error("two keys should be different")
	}
}
```

### Step 12: Run all crypto tests

```bash
go test ./internal/crypto/ -v
```

Expected: All tests PASS.

### Step 13: Commit

```bash
git add internal/crypto/
git commit -m "feat(crypto): add key generation, argon2id/bcrypt hashing, AES-256-GCM encryption, HKDF key derivation"
```

---

## Task 2: GORM Models

**Files:**
- Create: `internal/models/key.go`
- Create: `internal/models/key_test.go`
- Create: `internal/models/usage_log.go`
- Create: `internal/models/admin.go`
- Create: `internal/models/admin_test.go`
- Create: `internal/models/service_account.go`
- Create: `internal/models/login_log.go`
- Create: `internal/models/config.go`

**Interfaces:**
- Produces: `Key`, `UsageLog`, `Admin`, `ServiceAccount`, `LoginLog`, `Config` 结构体 — 供 Task 4 AutoMigrate 和 Task 5-9 handler 使用

### Step 1: Write key.go

```go
// internal/models/key.go
package models

import "time"

// KeyStatus 卡密状态
type KeyStatus string

const (
	KeyStatusUnused   KeyStatus = "unused"
	KeyStatusUsed     KeyStatus = "used"
	KeyStatusDisabled KeyStatus = "disabled"
	KeyStatusExpired  KeyStatus = "expired"
)

// BillingMode 计费模式
type BillingMode string

const (
	BillingModeCount  BillingMode = "count"
	BillingModeCredit BillingMode = "credit"
)

// Key 卡密模型
type Key struct {
	ID              int64       `gorm:"primaryKey;autoIncrement" json:"id"`
	Alias           string      `gorm:"size:255" json:"alias"`
	KeyHash         string      `gorm:"size:255;uniqueIndex" json:"-"`
	KeyPrefix       string      `gorm:"size:50" json:"key_prefix"`
	KeySuffix       string      `gorm:"size:10" json:"key_suffix"`
	BillingMode     BillingMode `gorm:"size:20" json:"billing_mode"`
	InitialAmount   int64       `json:"initial_amount"`
	RemainingAmount int64       `json:"remaining_amount"`
	Status          KeyStatus   `gorm:"size:20;index" json:"status"`
	CreatedBy       string      `gorm:"size:100" json:"created_by"`
	Version         int64       `gorm:"default:0" json:"-"` // 乐观锁
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	UsedAt          *time.Time  `json:"used_at"`
}
```

### Step 2: Write key_test.go

```go
// internal/models/key_test.go
package models

import "testing"

func TestKeyConstants(t *testing.T) {
	if KeyStatusUnused != "unused" {
		t.Error("KeyStatusUnused should be 'unused'")
	}
	if BillingModeCount != "count" {
		t.Error("BillingModeCount should be 'count'")
	}
}

func TestKeyDefaultValues(t *testing.T) {
	k := Key{}
	if k.Status != "" {
		t.Error("default status should be empty")
	}
	if k.Version != 0 {
		t.Error("default version should be 0")
	}
}
```

### Step 3: Run key model tests

```bash
go test ./internal/models/ -run TestKey -v
```

### Step 4: Write usage_log.go

```go
// internal/models/usage_log.go
package models

import "time"

// UsageLog 使用记录模型
type UsageLog struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	KeyID          int64     `gorm:"index" json:"key_id"`
	KeyAlias       string    `gorm:"size:255" json:"key_alias"`
	Amount         int64     `json:"amount"`
	IP             string    `gorm:"size:50;index" json:"ip"`
	UserAgent      string    `gorm:"size:500" json:"user_agent"`
	RequestPath    string    `gorm:"size:500" json:"request_path"`
	RequestParams  string    `gorm:"type:text" json:"request_params,omitempty"`
	ResponseStatus int       `json:"response_status"`
	CreatedAt      time.Time `gorm:"index" json:"created_at"`
}
```

### Step 5: Write admin.go

```go
// internal/models/admin.go
package models

import "time"

// Admin 管理员模型
type Admin struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"size:100;uniqueIndex" json:"username"`
	PasswordHash string    `gorm:"size:255" json:"-"`
	TOTPSecret   string    `gorm:"size:255" json:"-"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
```

### Step 6: Write admin_test.go

```go
// internal/models/admin_test.go
package models

import "testing"

func TestAdminDefaults(t *testing.T) {
	a := Admin{}
	if a.IsActive != false {
		// bool zero value is false
		t.Error("default IsActive should be false")
	}
}
```

### Step 7: Write service_account.go

```go
// internal/models/service_account.go
package models

import "time"

// ServiceAccount 服务账号模型
type ServiceAccount struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:100" json:"name"`
	KeyHash   string    `gorm:"size:255;uniqueIndex" json:"-"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

### Step 8: Write login_log.go

```go
// internal/models/login_log.go
package models

import "time"

// LoginLog 登录日志模型
type LoginLog struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	AdminID   int64     `gorm:"index" json:"admin_id"`
	IP        string    `gorm:"size:50" json:"ip"`
	UserAgent string    `gorm:"size:500" json:"user_agent"`
	Status    string    `gorm:"size:20" json:"status"` // success / failed
	CreatedAt time.Time `json:"created_at"`
}
```

### Step 9: Write config.go

```go
// internal/models/config.go
package models

import "time"

// Config 系统配置模型
type Config struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Key         string    `gorm:"size:100;uniqueIndex" json:"key"`
	Value       string    `gorm:"size:500" json:"value"`
	Description string    `gorm:"size:500" json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}
```

### Step 10: Run all model tests

```bash
go test ./internal/models/ -v
```

Expected: All tests PASS.

### Step 11: Commit

```bash
git add internal/models/
git commit -m "feat(models): add GORM models for Key, UsageLog, Admin, ServiceAccount, LoginLog, Config"
```

---

## Task 3: API Foundation

**Files:**
- Create: `internal/api/response.go`
- Create: `internal/api/response_test.go`
- Create: `internal/api/router.go`
- Create: `internal/api/middleware/auth.go`
- Create: `internal/api/middleware/cors.go`

**Dependencies to add:**
```
go get github.com/gin-gonic/gin
go get github.com/golang-jwt/jwt/v5
```

**Interfaces:**
- Produces: `Success(c *gin.Context, data interface{})` — 统一成功响应
- Produces: `Error(c *gin.Context, code int, message string)` — 统一错误响应
- Produces: `PaginatedSuccess(c *gin.Context, list interface{}, total int64, page, pageSize int)` — 分页响应
- Produces: 错误码常量 `ErrCode*`
- Produces: `SetupRouter(db *gorm.DB) *gin.Engine` — 路由注册（含 AutoMigrate）
- Produces: `JWTAuthMiddleware(secret string) gin.HandlerFunc` — JWT 认证中间件

### Step 1: Write response.go

```go
// internal/api/response.go
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 错误码定义
const (
	ErrCodeSuccess       = 0
	ErrCodeKeyNotFound   = 1001
	ErrCodeKeyDisabled   = 1002
	ErrCodeKeyExhausted  = 1003
	ErrCodeAmountExceeds = 1004
	ErrCodeAuthFailed    = 2001
	ErrCodeTOTPFailed    = 2002
	ErrCodeTokenInvalid  = 2003
	ErrCodeNoPermission  = 2004
	ErrCodeServiceKeyInv = 3001
	ErrCodeInternal      = 9999
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// PaginatedData 分页数据
type PaginatedData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    ErrCodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// PaginatedSuccess 分页成功响应
func PaginatedSuccess(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	Success(c, PaginatedData{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
```

### Step 2: Write response_test.go

```go
// internal/api/response_test.go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Success(c, map[string]string{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp Response
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != ErrCodeSuccess {
		t.Errorf("expected code 0, got %d", resp.Code)
	}
	if resp.Message != "success" {
		t.Errorf("expected 'success', got %s", resp.Message)
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Error(c, ErrCodeKeyNotFound, "卡密不存在")

	var resp Response
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != ErrCodeKeyNotFound {
		t.Errorf("expected code 1001, got %d", resp.Code)
	}
	if resp.Data != nil {
		t.Error("error response data should be nil")
	}
}

func TestPaginatedSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	items := []string{"a", "b", "c"}
	PaginatedSuccess(c, items, 100, 1, 20)

	var resp Response
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != ErrCodeSuccess {
		t.Errorf("expected code 0, got %d", resp.Code)
	}
}
```

### Step 3: Run response tests

```bash
go test ./internal/api/ -run TestSuccess -v
```

### Step 4: Write middleware/auth.go

```go
// internal/api/middleware/auth.go
package middleware

import (
	"CloudKey/internal/api"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuthMiddleware JWT 认证中间件
func JWTAuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			api.Error(c, api.ErrCodeTokenInvalid, "缺少认证信息")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			api.Error(c, api.ErrCodeTokenInvalid, "Token 无效或已过期")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			api.Error(c, api.ErrCodeTokenInvalid, "Token 无效")
			c.Abort()
			return
		}

		adminID, ok := claims["admin_id"].(float64)
		if !ok {
			api.Error(c, api.ErrCodeTokenInvalid, "Token 无效")
			c.Abort()
			return
		}

		c.Set("admin_id", int64(adminID))
		c.Next()
	}
}
```

### Step 5: Write middleware/cors.go

```go
// internal/api/middleware/cors.go
package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Service-Key")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
```

### Step 6: Write router.go

```go
// internal/api/router.go
package api

import (
	"CloudKey/internal/api/handler"
	"CloudKey/internal/api/middleware"
	"CloudKey/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter 注册路由并执行 AutoMigrate
func SetupRouter(db *gorm.DB, jwtSecret string) *gin.Engine {
	// AutoMigrate
	db.AutoMigrate(
		&models.Key{},
		&models.UsageLog{},
		&models.Admin{},
		&models.ServiceAccount{},
		&models.LoginLog{},
		&models.Config{},
	)

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	// 公开接口（卡密持有者）
	keyH := handler.NewKeyHandler(db)
	api := r.Group("/api")
	{
		api.GET("/key/status", keyH.Status)
		api.POST("/key/consume", keyH.Consume)
	}

	// 管理接口
	admin := r.Group("/api/admin")
	{
		// 无需认证
		authH := handler.NewAdminAuthHandler(db, jwtSecret)
		admin.POST("/login", authH.Login)
		admin.POST("/login/verify-2fa", authH.Verify2FA)

		// 需要 JWT 认证
		protected := admin.Group("", middleware.JWTAuthMiddleware(jwtSecret))
		{
			protected.GET("/profile", authH.Profile)
			protected.PUT("/password", authH.ChangePassword)

			adminKeyH := handler.NewAdminKeyHandler(db)
			protected.POST("/keys", adminKeyH.Create)
			protected.GET("/keys", adminKeyH.List)
			protected.GET("/keys/:id", adminKeyH.Detail)
			protected.PATCH("/keys/:id", adminKeyH.Update)
			protected.PATCH("/keys/:id/disable", adminKeyH.Disable)
			protected.PATCH("/keys/:id/enable", adminKeyH.Enable)
			protected.DELETE("/keys/:id", adminKeyH.Delete)
			protected.GET("/keys/export", adminKeyH.Export)

			adminLogH := handler.NewAdminLogHandler(db)
			protected.GET("/usage-logs", adminLogH.List)

			adminStatsH := handler.NewAdminStatsHandler(db)
			protected.GET("/stats/overview", adminStatsH.Overview)
			protected.GET("/stats/trends", adminStatsH.Trends)
		}
	}

	// 服务接口
	svc := r.Group("/api/service")
	{
		svcH := handler.NewServiceHandler(db)
		svc.Use(middleware.ServiceAuthMiddleware(db))
		svc.POST("/keys", svcH.CreateKey)
		svc.GET("/keys", svcH.ListKeys)
	}

	return r
}
```

### Step 7: Write middleware/service.go

```go
// internal/api/middleware/service.go
package middleware

import (
	"CloudKey/internal/api"
	"CloudKey/internal/crypto"
	"CloudKey/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ServiceAuthMiddleware 服务账号认证中间件
func ServiceAuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceKey := c.GetHeader("X-Service-Key")
		if serviceKey == "" {
			api.Error(c, api.ErrCodeServiceKeyInv, "缺少服务账号密钥")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 遍历所有活跃服务账号验证密钥
		var accounts []models.ServiceAccount
		db.Where("is_active = ?", true).Find(&accounts)

		for _, acc := range accounts {
			if crypto.VerifyKey(serviceKey, acc.KeyHash) {
				c.Set("service_account_id", acc.ID)
				c.Set("service_account_name", acc.Name)
				c.Next()
				return
			}
		}

		api.Error(c, api.ErrCodeServiceKeyInv, "服务账号密钥无效")
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}
```

### Step 8: Write router_test.go

```go
// internal/api/router_test.go
package api

import "testing"

func TestErrorCodes(t *testing.T) {
	codes := map[int]string{
		ErrCodeSuccess:       "success",
		ErrCodeKeyNotFound:   "key not found",
		ErrCodeKeyDisabled:   "key disabled",
		ErrCodeKeyExhausted:  "key exhausted",
		ErrCodeAmountExceeds: "amount exceeds",
		ErrCodeAuthFailed:    "auth failed",
		ErrCodeTOTPFailed:    "totp failed",
		ErrCodeTokenInvalid:  "token invalid",
		ErrCodeNoPermission:  "no permission",
		ErrCodeServiceKeyInv: "service key invalid",
		ErrCodeInternal:      "internal error",
	}
	if len(codes) != 11 {
		t.Errorf("expected 11 error codes, got %d", len(codes))
	}
}
```

### Step 9: Run all API foundation tests

```bash
go test ./internal/api/... -v
```

Expected: All tests PASS. Note: router_test.go tests error code constants; handler tests come in later tasks.

### Step 10: Commit

```bash
git add internal/api/ go.mod go.sum
git commit -m "feat(api): add response helpers, error codes, CORS, JWT middleware, router with AutoMigrate"
```

---

## Task 4: Public Key Endpoints

**Files:**
- Create: `internal/api/handler/key.go`
- Create: `internal/api/handler/key_test.go`

**Interfaces:**
- Consumes: `models.Key`, `models.UsageLog` (Task 2)
- Consumes: `crypto.HashKey`, `crypto.VerifyKey` (Task 1)
- Consumes: `api.Success`, `api.Error`, `api.ErrCode*` (Task 3)
- Produces: `KeyHandler.Status` — `GET /api/key/status?sk=卡密`
- Produces: `KeyHandler.Consume` — `POST /api/key/consume`

### Step 1: Write handler/key.go

```go
// internal/api/handler/key.go
package handler

import (
	"CloudKey/internal/api"
	"CloudKey/internal/crypto"
	"CloudKey/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type KeyHandler struct {
	db *gorm.DB
}

func NewKeyHandler(db *gorm.DB) *KeyHandler {
	return &KeyHandler{db: db}
}

// StatusRequest 查询请求
type StatusRequest struct {
	SK string `form:"sk" binding:"required"`
}

// Status 查询卡密状态（不扣减）
func (h *KeyHandler) Status(c *gin.Context) {
	var req StatusRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		api.Error(c, http.StatusBadRequest, "缺少参数 sk")
		return
	}

	keyHash := crypto.HashKey(req.SK)
	var key models.Key
	if err := h.db.Where("key_hash = ?", keyHash).First(&key).Error; err != nil {
		api.Error(c, api.ErrCodeKeyNotFound, "卡密不存在")
		return
	}

	api.Success(c, gin.H{
		"alias":            key.Alias,
		"billing_mode":     key.BillingMode,
		"remaining_amount": key.RemainingAmount,
		"status":           key.Status,
		"created_at":       key.CreatedAt,
		"used_at":          key.UsedAt,
	})
}

// ConsumeRequest 扣减请求
type ConsumeRequest struct {
	Key    string `json:"key" binding:"required"`
	Amount int64  `json:"amount"`
}

// Consume 扣减卡密额度
func (h *KeyHandler) Consume(c *gin.Context) {
	var req ConsumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.Error(c, http.StatusBadRequest, "缺少参数 key")
		return
	}
	if req.Amount <= 0 {
		req.Amount = 1
	}

	keyHash := crypto.HashKey(req.SK)
	var key models.Key

	// 事务 + 乐观锁扣减
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("key_hash = ?", keyHash).First(&key).Error; err != nil {
			return err
		}
		if key.Status != models.KeyStatusUnused {
			return fmt.Errorf("key status: %s", key.Status)
		}
		if key.RemainingAmount < req.Amount {
			return fmt.Errorf("insufficient amount")
		}

		result := tx.Model(&models.Key{}).
			Where("id = ? AND version = ?", key.ID, key.Version).
			Updates(map[string]interface{}{
				"remaining_amount": key.RemainingAmount - req.Amount,
				"version":          key.Version + 1,
				"used_at":          time.Now(),
			})

		if result.RowsAffected == 0 {
			return fmt.Errorf("concurrent modification")
		}

		// 更新本地变量用于响应
		key.RemainingAmount -= req.Amount
		key.UsedAt = &time.Time{}
		*key.UsedAt = time.Now()
		return nil
	})

	if err != nil {
		errMsg := err.Error()
		switch {
		case errMsg == "record not found" || errMsg == "卡密不存在":
			api.Error(c, api.ErrCodeKeyNotFound, "卡密不存在")
		case errMsg == "卡密已禁用" || key.Status == models.KeyStatusDisabled:
			api.Error(c, api.ErrCodeKeyDisabled, "卡密已禁用")
		case errMsg == "卡密额度已用尽" || key.Status == models.KeyStatusUsed:
			api.Error(c, api.ErrCodeKeyExhausted, "卡密额度已用尽")
		case errMsg == "扣减数量超过剩余额度":
			api.Error(c, api.ErrCodeAmountExceeds, "扣减数量超过剩余额度")
		default:
			api.Error(c, api.ErrCodeInternal, "系统内部错误")
		}
		return
	}

	// 如果额度用尽，更新状态
	if key.RemainingAmount == 0 {
		h.db.Model(&key).Update("status", models.KeyStatusUsed)
	}

	// 记录使用日志
	log := models.UsageLog{
		KeyID:          key.ID,
		KeyAlias:       key.Alias,
		Amount:         req.Amount,
		IP:             c.ClientIP(),
		UserAgent:      c.GetHeader("User-Agent"),
		RequestPath:    c.Request.URL.Path,
		ResponseStatus: http.StatusOK,
		CreatedAt:      time.Now(),
	}
	h.db.Create(&log)

	api.Success(c, gin.H{
		"remaining_amount": key.RemainingAmount,
		"status":           key.Status,
		"is_exhausted":     key.RemainingAmount == 0,
	})
}
```

**注意：** 上面的 Consume 方法中有几个问题需要修正：
1. `req.SK` 应为 `req.Key`
2. 需要 `import "fmt"`
3. 错误匹配逻辑需要简化

### Step 2: 修正后的完整 handler/key.go

```go
// internal/api/handler/key.go
package handler

import (
	"CloudKey/internal/api"
	"CloudKey/internal/crypto"
	"CloudKey/internal/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type KeyHandler struct {
	db *gorm.DB
}

func NewKeyHandler(db *gorm.DB) *KeyHandler {
	return &KeyHandler{db: db}
}

type StatusRequest struct {
	SK string `form:"sk" binding:"required"`
}

func (h *KeyHandler) Status(c *gin.Context) {
	var req StatusRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		api.Error(c, http.StatusBadRequest, "缺少参数 sk")
		return
	}

	keyHash := crypto.HashKey(req.SK)
	var key models.Key
	if err := h.db.Where("key_hash = ?", keyHash).First(&key).Error; err != nil {
		api.Error(c, api.ErrCodeKeyNotFound, "卡密不存在")
		return
	}

	api.Success(c, gin.H{
		"alias":            key.Alias,
		"billing_mode":     key.BillingMode,
		"remaining_amount": key.RemainingAmount,
		"status":           key.Status,
		"created_at":       key.CreatedAt,
		"used_at":          key.UsedAt,
	})
}

type ConsumeRequest struct {
	Key    string `json:"key" binding:"required"`
	Amount int64  `json:"amount"`
}

func (h *KeyHandler) Consume(c *gin.Context) {
	var req ConsumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.Error(c, http.StatusBadRequest, "缺少参数 key")
		return
	}
	if req.Amount <= 0 {
		req.Amount = 1
	}

	keyHash := crypto.HashKey(req.Key)
	var key models.Key

	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("key_hash = ?", keyHash).First(&key).Error; err != nil {
			return fmt.Errorf("not_found")
		}
		if key.Status == models.KeyStatusDisabled {
			return fmt.Errorf("disabled")
		}
		if key.Status == models.KeyStatusUsed {
			return fmt.Errorf("exhausted")
		}
		if key.RemainingAmount < req.Amount {
			return fmt.Errorf("exceeds")
		}

		result := tx.Model(&models.Key{}).
			Where("id = ? AND version = ?", key.ID, key.Version).
			Updates(map[string]interface{}{
				"remaining_amount": gorm.Expr("remaining_amount - ?", req.Amount),
				"version":          gorm.Expr("version + 1"),
				"used_at":          time.Now(),
			})

		if result.RowsAffected == 0 {
			return fmt.Errorf("concurrent")
		}
		return nil
	})

	if err != nil {
		switch err.Error() {
		case "not_found":
			api.Error(c, api.ErrCodeKeyNotFound, "卡密不存在")
		case "disabled":
			api.Error(c, api.ErrCodeKeyDisabled, "卡密已禁用")
		case "exhausted":
			api.Error(c, api.ErrCodeKeyExhausted, "卡密额度已用尽")
		case "exceeds":
			api.Error(c, api.ErrCodeAmountExceeds, "扣减数量超过剩余额度")
		default:
			api.Error(c, api.ErrCodeInternal, "系统内部错误")
		}
		return
	}

	// 重新查询扣减后的状态
	h.db.Where("id = ?", key.ID).First(&key)

	// 更新状态为已用尽
	if key.RemainingAmount == 0 {
		h.db.Model(&key).Update("status", models.KeyStatusUsed)
		key.Status = models.KeyStatusUsed
	}

	// 记录使用日志
	h.db.Create(&models.UsageLog{
		KeyID:          key.ID,
		KeyAlias:       key.Alias,
		Amount:         req.Amount,
		IP:             c.ClientIP(),
		UserAgent:      c.GetHeader("User-Agent"),
		RequestPath:    c.Request.URL.Path,
		ResponseStatus: http.StatusOK,
		CreatedAt:      time.Now(),
	})

	api.Success(c, gin.H{
		"remaining_amount": key.RemainingAmount,
		"status":           key.Status,
		"is_exhausted":     key.RemainingAmount == 0,
	})
}
```

### Step 3: Write handler/key_test.go

```go
// internal/api/handler/key_test.go
package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"CloudKey/internal/api"
	"CloudKey/internal/crypto"
	"CloudKey/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(&models.Key{}, &models.UsageLog{})
	return db
}

func TestStatus_NotFound(t *testing.T) {
	db := setupTestDB(t)
	h := NewKeyHandler(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/key/status?sk=nonexistent", nil)

	h.Status(c)

	var resp api.Response
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != api.ErrCodeKeyNotFound {
		t.Errorf("expected code %d, got %d", api.ErrCodeKeyNotFound, resp.Code)
	}
}

func TestStatus_Success(t *testing.T) {
	db := setupTestDB(t)
	h := NewKeyHandler(db)

	// 插入测试数据
	plainKey := "sk-test1234567890abcdef"
	hash := crypto.HashKey(plainKey)
	db.Create(&models.Key{
		Alias:           "test-key",
		KeyHash:         hash,
		KeyPrefix:       "sk-",
		KeySuffix:       "cdef",
		BillingMode:     models.BillingModeCount,
		InitialAmount:   100,
		RemainingAmount: 100,
		Status:          models.KeyStatusUnused,
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/key/status?sk="+plainKey, nil)

	h.Status(c)

	var resp api.Response
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != api.ErrCodeSuccess {
		t.Errorf("expected code 0, got %d", resp.Code)
	}
}

func TestConsume_MissingKey(t *testing.T) {
	db := setupTestDB(t)
	h := NewKeyHandler(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/key/consume", bytes.NewBufferString("{}"))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Consume(c)

	var resp api.Response
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.Code)
	}
}

func TestConsume_Success(t *testing.T) {
	db := setupTestDB(t)
	h := NewKeyHandler(db)

	plainKey := "sk-testconsume000000000000000000"
	hash := crypto.HashKey(plainKey)
	db.Create(&models.Key{
		Alias:           "consume-test",
		KeyHash:         hash,
		KeyPrefix:       "sk-",
		KeySuffix:       "0000",
		BillingMode:     models.BillingModeCount,
		InitialAmount:   10,
		RemainingAmount: 10,
		Status:          models.KeyStatusUnused,
	})

	body, _ := json.Marshal(ConsumeRequest{Key: plainKey, Amount: 1})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/key/consume", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Consume(c)

	var resp api.Response
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != api.ErrCodeSuccess {
		t.Errorf("expected code 0, got %d: %v", resp.Code, resp.Message)
	}
}
```

**注意：** key_test.go 使用 SQLite 内存数据库做测试，需要额外依赖：
```
go get gorm.io/driver/sqlite
```

### Step 4: Run key handler tests

```bash
go test ./internal/api/handler/ -run TestStatus -v
go test ./internal/api/handler/ -run TestConsume -v
```

### Step 5: Commit

```bash
git add internal/api/handler/key.go internal/api/handler/key_test.go
git commit -m "feat(api): add public key endpoints - status query and consume with optimistic locking"
```
