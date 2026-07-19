# CloudKey SaaS 多租户架构改造 — 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 CloudKey 从单租户服务账号系统改造为多租户 SaaS 平台，支持系统管理员管理租户、租户管理员管理 Key 和服务账号。

**Architecture:** 共享数据库 + tenant_id 列隔离 + 统一 User 模型（role 区分 super_admin/tenant_admin）。JWT Claims 新增 role 和 tenant_id 字段。中间件链：AuthMiddleware → RequireSuperAdmin/RequireTenantAdmin → TenantBusinessGuard（守卫业务操作）。

**Tech Stack:** Go 1.25, Gin v1.12, GORM v1.30, MySQL 8.0, JWT (golang-jwt/v5), TOTP (pquerna/otp), bcrypt

**Spec:** `docs/superpowers/specs/2026-07-19-saas-multi-tenant-design.md`

## Global Constraints

- 所有业务表必须包含 `tenant_id BIGINT NOT NULL INDEX`
- `tenant_id` 由中间件注入 Gin context，service 层通过 `c.GetInt64("tenant_id")` 获取
- 系统管理员不能访问租户业务数据（Key、服务账号、详细日志）
- 统一登录 POST /api/auth/login，前端根据 role 加载对应页面
- 错误码扩展：4001-租户已过期, 4002-租户已被禁用, 4003-租户不存在, 5001-系统管理员权限不足, 5002-租户管理员权限不足

---

### Task 1: 新增 Tenant 模型 + 改造 LoginLog 模型

**Files:**
- Create: `internal/model/tenant.go`
- Modify: `internal/model/login_log.go`
- Modify: `internal/model/migrate.go`

**Interfaces:**
- Produces: `model.Tenant` struct with fields `ID`, `Name`, `Status`, `ExpireAt`, `KeyPrefix`, `KeyLength`, `KeySuffixLength`, `CreatedAt`, `UpdatedAt`; `TableName() "tenants"`
- Produces: `model.LoginLog.UserID uint64` (重命名 AdminID), `model.LoginLog.TenantID *uint64` (nullable for super_admin)

- [ ] **Step 1: 创建 `internal/model/tenant.go`**

```go
package model

import "time"

type TenantStatus string

const (
	TenantStatusActive   TenantStatus = "active"
	TenantStatusExpired  TenantStatus = "expired"
	TenantStatusDisabled TenantStatus = "disabled"
)

type Tenant struct {
	ID              uint64       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string       `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Status          TenantStatus `gorm:"type:varchar(20);not null;default:active" json:"status"`
	ExpireAt        *time.Time   `gorm:"default:null" json:"expire_at"`
	KeyPrefix       string       `gorm:"type:varchar(20);not null;default:sk-" json:"key_prefix"`
	KeyLength       int          `gorm:"type:int;not null;default:32" json:"key_length"`
	KeySuffixLength int          `gorm:"type:int;not null;default:4" json:"key_suffix_length"`
	CreatedAt       time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Tenant) TableName() string { return "tenants" }
```

- [ ] **Step 2: 改造 `internal/model/login_log.go` — AdminID → UserID + 加 TenantID**

```go
package model

import "time"

type LoginStatus string

const (
	LoginStatusSuccess LoginStatus = "success"
	LoginStatusFailed  LoginStatus = "failed"
)

type LoginLog struct {
	ID        uint64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64      `gorm:"type:bigint;index;not null" json:"user_id"`
	TenantID  *uint64     `gorm:"type:bigint;index;default:null" json:"tenant_id"`
	IP        string      `gorm:"type:varchar(50);not null" json:"ip"`
	UserAgent string      `gorm:"type:varchar(500)" json:"user_agent"`
	Status    LoginStatus `gorm:"type:varchar(20);not null" json:"status"`
	CreatedAt time.Time   `gorm:"autoCreateTime" json:"created_at"`
}

func (LoginLog) TableName() string { return "login_logs" }
```

- [ ] **Step 3: 更新 `internal/model/migrate.go` — 注册 Tenant**

```go
package model

import "gorm.io/gorm"

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Tenant{},
		&User{},
		&Key{},
		&UsageLog{},
		&ServiceAccount{},
		&LoginLog{},
		&SysConfig{},
	)
}
```

- [ ] **Step 4: 验证编译通过**

Run: `cd D:/MyGoProject/CloudKey && go build ./...`
Expected: 编译成功（可能有未引用警告，后续 task 会引用）

- [ ] **Step 5: Commit**

```bash
git add internal/model/tenant.go internal/model/login_log.go internal/model/migrate.go
git commit -m "feat: add Tenant model, update LoginLog with UserID+TenantID"
```

---

### Task 2: 创建 User 模型（替代 Admin）+ 改造 Key/ServiceAccount/UsageLog 加 tenant_id

**Files:**
- Create: `internal/model/user.go`
- Modify: `internal/model/key.go`
- Modify: `internal/model/service_account.go`
- Modify: `internal/model/usage_log.go`
- Modify: `internal/model/migrate.go`（已在 Task 1 更新，此 task 加入 User）

**Interfaces:**
- Produces: `model.User` struct with `ID`, `Username`, `PasswordHash`, `TotpSecret`, `TotpSetup`, `Role`, `TenantID`, `IsActive`, `CreatedAt`, `UpdatedAt`
- Produces: `model.Key.TenantID uint64` (新增)
- Produces: `model.ServiceAccount.TenantID uint64` (新增)
- Produces: `model.UsageLog.TenantID uint64` (新增)

- [ ] **Step 1: 创建 `internal/model/user.go`**

```go
package model

import "time"

type UserRole string

const (
	RoleSuperAdmin  UserRole = "super_admin"
	RoleTenantAdmin UserRole = "tenant_admin"
)

type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	TotpSecret   string    `gorm:"type:varchar(255)" json:"-"`
	TotpSetup    bool      `gorm:"default:false" json:"totp_setup"`
	Role         UserRole  `gorm:"type:varchar(20);not null" json:"role"`
	TenantID     *uint64   `gorm:"type:bigint;index;default:null" json:"tenant_id"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string { return "users" }
```

- [ ] **Step 2: 改造 `internal/model/key.go` — Key 结构体加 `TenantID`**

在 Key struct 字段中新增（放在 ID 之后）:
```go
TenantID        uint64         `gorm:"type:bigint;index;not null" json:"tenant_id"`
```

完整 Key struct:
```go
type Key struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID        uint64         `gorm:"type:bigint;index;not null" json:"tenant_id"`
	Alias           string         `gorm:"type:varchar(255);not null" json:"alias"`
	KeyHash         string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"`
	KeyPrefix       string         `gorm:"type:varchar(50);not null" json:"key_prefix"`
	KeySuffix       string         `gorm:"type:varchar(10);not null" json:"key_suffix"`
	BillingMode     KeyBillingMode `gorm:"type:varchar(20);not null" json:"billing_mode"`
	InitialAmount   int64          `gorm:"type:bigint;not null" json:"initial_amount"`
	RemainingAmount int64          `gorm:"type:bigint;not null" json:"remaining_amount"`
	Version         int64          `gorm:"type:bigint;not null;default:0" json:"-"`
	Status          KeyStatus      `gorm:"type:varchar(20);not null;default:unused" json:"status"`
	CreatedBy       string         `gorm:"type:varchar(100);not null" json:"created_by"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	UsedAt          *time.Time     `gorm:"default:null" json:"used_at"`
	ExpireAt        *time.Time     `gorm:"default:null" json:"expire_at"`
	MaxUsage        *int64         `gorm:"default:null" json:"max_usage"`
}
```

- [ ] **Step 3: 改造 `internal/model/service_account.go` — 加 `TenantID`**

在 ServiceAccount struct 中新增（放在 ID 之后）:
```go
TenantID  uint64    `gorm:"type:bigint;index;not null" json:"tenant_id"`
```

完整:
```go
type ServiceAccount struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID  uint64    `gorm:"type:bigint;index;not null" json:"tenant_id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	KeyHash   string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
```

- [ ] **Step 4: 改造 `internal/model/usage_log.go` — 加 `TenantID`**

在 UsageLog struct 中新增（放在 ID 之后）:
```go
TenantID       uint64    `gorm:"type:bigint;index;not null" json:"tenant_id"`
```

完整:
```go
type UsageLog struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID       uint64    `gorm:"type:bigint;index;not null" json:"tenant_id"`
	KeyID          uint64    `gorm:"type:bigint;index;not null" json:"key_id"`
	KeyAlias       string    `gorm:"type:varchar(255);not null" json:"key_alias"`
	Amount         int64     `gorm:"type:bigint;not null" json:"amount"`
	IP             string    `gorm:"type:varchar(50);not null" json:"ip"`
	UserAgent      string    `gorm:"type:varchar(500)" json:"user_agent"`
	RequestPath    string    `gorm:"type:varchar(500)" json:"request_path"`
	RequestParams  string    `gorm:"type:text" json:"request_params"`
	ResponseStatus int       `gorm:"type:int;not null" json:"response_status"`
	CreatedAt      time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}
```

- [ ] **Step 5: 验证编译通过**

Run: `cd D:/MyGoProject/CloudKey && go build ./...`
Expected: 编译通过（model 层变更完成）

- [ ] **Step 6: Commit**

```bash
git add internal/model/user.go internal/model/key.go internal/model/service_account.go internal/model/usage_log.go internal/model/migrate.go
git commit -m "feat: add User model, add TenantID to Key/ServiceAccount/UsageLog models"
```

---

### Task 3: 更新错误码 + 改造 JWT Claims（中间件层）

**Files:**
- Modify: `internal/errcode/errcode.go`
- Rewrite: `internal/middleware/auth.go`
- Create: `internal/middleware/super_admin.go`
- Create: `internal/middleware/tenant_admin.go`
- Create: `internal/middleware/tenant_business.go`
- Rewrite: `internal/middleware/service_auth.go`

**Interfaces:**
- Consumes: `model.User.Role`, `model.Tenant.Status`
- Produces: `AuthMiddleware(jwtSecret) -> gin.HandlerFunc`; `RequireSuperAdmin() -> gin.HandlerFunc`; `RequireTenantAdmin() -> gin.HandlerFunc`; `TenantBusinessGuard() -> gin.HandlerFunc`; `ServiceAuthMiddleware(svc) -> gin.HandlerFunc`
- Context keys: `"user_id"` (uint64), `"username"` (string), `"role"` (model.UserRole), `"tenant_id"` (uint64)

- [ ] **Step 1: 更新 `internal/errcode/errcode.go` — 新增租户相关错误码**

在 const block 末尾新增:
```go
// 租户相关 4001~4999
CodeTenantExpired  = 4001
CodeTenantDisabled = 4002
CodeTenantNotFound = 4003

// 权限相关 5001~5999
CodeSuperAdminRequired = 5001
CodeTenantAdminRequired = 5002
```

在 codeMessages map 中新增:
```go
CodeTenantExpired:         "租户已到期，仅可查看统计数据",
CodeTenantDisabled:        "租户已被禁用",
CodeTenantNotFound:        "租户不存在",
CodeSuperAdminRequired:    "需要系统管理员权限",
CodeTenantAdminRequired:   "需要租户管理员权限",
```

- [ ] **Step 2: 重写 `internal/middleware/auth.go` — 新的 JWT Claims + AuthMiddleware**

```go
package middleware

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint64         `json:"user_id"`
	Username string         `json:"username"`
	Role     model.UserRole `json:"role"`
	TenantID *uint64        `json:"tenant_id"`
	jwt.RegisteredClaims
}

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": errcode.CodeJWTInvalid, "message": errcode.GetMessage(errcode.CodeJWTInvalid), "data": nil})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": errcode.CodeJWTInvalid, "message": errcode.GetMessage(errcode.CodeJWTInvalid), "data": nil})
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"code": errcode.CodeJWTInvalid, "message": errcode.GetMessage(errcode.CodeJWTInvalid), "data": nil})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		if claims.TenantID != nil {
			c.Set("tenant_id", *claims.TenantID)
		}
		c.Next()
	}
}
```

- [ ] **Step 3: 创建 `internal/middleware/super_admin.go`**

```go
package middleware

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleI, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeSuperAdminRequired, "message": errcode.GetMessage(errcode.CodeSuperAdminRequired), "data": nil})
			c.Abort()
			return
		}
		role, ok := roleI.(model.UserRole)
		if !ok || role != model.RoleSuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeSuperAdminRequired, "message": errcode.GetMessage(errcode.CodeSuperAdminRequired), "data": nil})
			c.Abort()
			return
		}
		c.Next()
	}
}
```

- [ ] **Step 4: 创建 `internal/middleware/tenant_admin.go`**

```go
package middleware

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RequireTenantAdmin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleI, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeTenantAdminRequired, "message": errcode.GetMessage(errcode.CodeTenantAdminRequired), "data": nil})
			c.Abort()
			return
		}
		role, ok := roleI.(model.UserRole)
		if !ok || role != model.RoleTenantAdmin {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeTenantAdminRequired, "message": errcode.GetMessage(errcode.CodeTenantAdminRequired), "data": nil})
			c.Abort()
			return
		}

		tenantIDI, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeTenantAdminRequired, "message": "租户信息缺失", "data": nil})
			c.Abort()
			return
		}
		tenantID := tenantIDI.(uint64)

		// 检查租户是否 disabled
		var tenant model.Tenant
		if err := db.First(&tenant, tenantID).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeTenantNotFound, "message": errcode.GetMessage(errcode.CodeTenantNotFound), "data": nil})
			c.Abort()
			return
		}
		if tenant.Status == model.TenantStatusDisabled {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeTenantDisabled, "message": errcode.GetMessage(errcode.CodeTenantDisabled), "data": nil})
			c.Abort()
			return
		}
		// expired 仍放行，由 TenantBusinessGuard 控制业务操作

		c.Next()
	}
}
```

- [ ] **Step 5: 创建 `internal/middleware/tenant_business.go` — 业务操作守卫**

```go
package middleware

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TenantBusinessGuard(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantIDI, exists := c.Get("tenant_id")
		if !exists {
			// 非租户管理员，跳过（可能是 super_admin，但它不应调用租户业务接口）
			c.Next()
			return
		}
		tenantID := tenantIDI.(uint64)

		var tenant model.Tenant
		if err := db.First(&tenant, tenantID).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeTenantNotFound, "message": errcode.GetMessage(errcode.CodeTenantNotFound), "data": nil})
			c.Abort()
			return
		}

		if tenant.Status == model.TenantStatusExpired {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeTenantExpired, "message": errcode.GetMessage(errcode.CodeTenantExpired), "data": nil})
			c.Abort()
			return
		}
		if tenant.Status == model.TenantStatusDisabled {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeTenantDisabled, "message": errcode.GetMessage(errcode.CodeTenantDisabled), "data": nil})
			c.Abort()
			return
		}

		c.Next()
	}
}
```

- [ ] **Step 6: 重写 `internal/middleware/service_auth.go` — 加租户状态检查**

```go
package middleware

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"CloudKey/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ServiceAuthMiddleware(svc *service.ServiceAccountService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceKey := c.GetHeader("X-Service-Key")
		if serviceKey == "" {
			c.JSON(http.StatusOK, gin.H{"code": errcode.CodeServiceKeyInvalid, "message": errcode.GetMessage(errcode.CodeServiceKeyInvalid), "data": nil})
			c.Abort()
			return
		}

		account, err := svc.ValidateServiceKey(serviceKey)
		if err != nil || account == nil {
			c.JSON(http.StatusOK, gin.H{"code": errcode.CodeServiceKeyInvalid, "message": errcode.GetMessage(errcode.CodeServiceKeyInvalid), "data": nil})
			c.Abort()
			return
		}

		// 检查租户状态
		var tenant model.Tenant
		if err := db.First(&tenant, account.TenantID).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": errcode.CodeTenantNotFound, "message": errcode.GetMessage(errcode.CodeTenantNotFound), "data": nil})
			c.Abort()
			return
		}
		if tenant.Status != model.TenantStatusActive {
			c.JSON(http.StatusOK, gin.H{"code": errcode.CodeTenantExpired, "message": errcode.GetMessage(errcode.CodeTenantExpired), "data": nil})
			c.Abort()
			return
		}

		c.Set("service_account", account)
		c.Set("tenant_id", account.TenantID)
		c.Next()
	}
}
```

- [ ] **Step 7: 验证编译通过**

Run: `cd D:/MyGoProject/CloudKey && go build ./...`
Expected: 编译通过

- [ ] **Step 8: Commit**

```bash
git add internal/errcode/errcode.go internal/middleware/
git commit -m "feat: add new JWT claims with role+tenant_id, super/tenant/business middlewares"
```

---

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

### Task 5: 创建 AuthHandler（统一登录入口）

**Files:**
- Create: `internal/handler/auth_handler.go`

**Interfaces:**
- Consumes: `AuthService.Login`, `AuthService.VerifyTOTP`, `AuthService.GenerateTOTPSecret`, `AuthService.ConfirmTOTPSetup`, `AuthService.ChangePassword`, `AuthService.GetProfile`
- Produces: `AuthHandler` with `Login`, `Verify2FA`, `SetupTOTP`, `ConfirmTOTP`, `SetupTOTPPublic`, `ConfirmTOTPPublic`, `Profile`, `ChangePassword`

- [ ] **Step 1: 创建 `internal/handler/auth_handler.go`**

```go
package handler

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"CloudKey/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc     *service.AuthService
	loginLogSvc *service.LoginLogService
}

func NewAuthHandler(authSvc *service.AuthService, loginLogSvc *service.LoginLogService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, loginLogSvc: loginLogSvc}
}

// ========== 登录流程 ==========

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeInvalidCredentials, "参数错误")
		return
	}

	result, err := h.authSvc.Login(req.Username, req.Password)
	if err != nil {
		InternalError(c)
		return
	}
	if result == nil {
		h.loginLogSvc.RecordLogin(0, nil, c.ClientIP(), c.GetHeader("User-Agent"), false)
		Unauthorized(c, errcode.CodeInvalidCredentials, errcode.GetMessage(errcode.CodeInvalidCredentials))
		return
	}

	h.loginLogSvc.RecordLogin(result.UserID, nil, c.ClientIP(), c.GetHeader("User-Agent"), !result.RequireTOTP)

	if result.RequireTOTP {
		Success(c, gin.H{"require_totp": true, "user_id": result.UserID})
		return
	}

	Success(c, gin.H{
		"require_totp": false, "need_setup": true,
		"user_id": result.UserID, "message": "请设置 TOTP 两步验证",
	})
}

// ========== 2FA ==========

type Verify2FARequest struct {
	UserID uint64 `json:"user_id" binding:"required"`
	Code   string `json:"code" binding:"required"`
}

func (h *AuthHandler) Verify2FA(c *gin.Context) {
	var req Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeTOTPFailed, "参数错误")
		return
	}

	_, resp, err := h.authSvc.VerifyTOTP(req.UserID, req.Code)
	if err != nil {
		InternalError(c)
		return
	}
	if resp == nil {
		Unauthorized(c, errcode.CodeTOTPFailed, errcode.GetMessage(errcode.CodeTOTPFailed))
		return
	}

	Success(c, gin.H{
		"token":     resp.Token,
		"token_type": "Bearer",
		"role":      resp.Role,
		"tenant_id": resp.TenantID,
		"username":  resp.Username,
	})
}

// ========== TOTP 设置（已认证用户） ==========

func (h *AuthHandler) SetupTOTP(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		Unauthorized(c, errcode.CodeJWTInvalid, errcode.GetMessage(errcode.CodeJWTInvalid))
		return
	}

	secret, url, err := h.authSvc.GenerateTOTPSecret(userID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, gin.H{"secret": secret, "url": url})
}

func (h *AuthHandler) ConfirmTOTP(c *gin.Context) {
	var req Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeTOTPFailed, "参数错误")
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		Unauthorized(c, errcode.CodeJWTInvalid, errcode.GetMessage(errcode.CodeJWTInvalid))
		return
	}

	if err := h.authSvc.ConfirmTOTPSetup(userID, req.Code); err != nil {
		BadRequest(c, errcode.CodeTOTPFailed, err.Error())
		return
	}
	Success(c, nil)
}

// ========== TOTP 设置（公开接口：首次设置，无需 JWT） ==========

func (h *AuthHandler) SetupTOTPPublic(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeTOTPFailed, "参数错误")
		return
	}

	profile, err := h.authSvc.GetProfile(req.UserID)
	if err != nil || profile == nil {
		BadRequest(c, errcode.CodeInvalidCredentials, "用户不存在")
		return
	}
	if profile.TotpSetup {
		BadRequest(c, errcode.CodeTOTPFailed, "TOTP 已设置，请直接登录")
		return
	}

	secret, url, err := h.authSvc.GenerateTOTPSecret(req.UserID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, gin.H{"secret": secret, "url": url})
}

func (h *AuthHandler) ConfirmTOTPPublic(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
		Code   string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeTOTPFailed, "参数错误")
		return
	}

	profile, err := h.authSvc.GetProfile(req.UserID)
	if err != nil || profile == nil {
		BadRequest(c, errcode.CodeInvalidCredentials, "用户不存在")
		return
	}
	if profile.TotpSetup {
		BadRequest(c, errcode.CodeTOTPFailed, "TOTP 已设置，请直接登录")
		return
	}

	if err := h.authSvc.ConfirmTOTPSetup(req.UserID, req.Code); err != nil {
		BadRequest(c, errcode.CodeTOTPFailed, "验证码错误")
		return
	}

	_, resp, _ := h.authSvc.VerifyTOTP(req.UserID, req.Code)
	if resp == nil {
		Success(c, gin.H{"message": "TOTP 设置成功，请重新登录"})
		return
	}

	Success(c, gin.H{
		"token":      resp.Token,
		"token_type": "Bearer",
		"role":       resp.Role,
		"tenant_id":  resp.TenantID,
		"username":   resp.Username,
	})
}

// ========== 个人设置 ==========

func (h *AuthHandler) Profile(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		Unauthorized(c, errcode.CodeJWTInvalid, errcode.GetMessage(errcode.CodeJWTInvalid))
		return
	}
	user, err := h.authSvc.GetProfile(userID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, user)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeForbidden, "参数错误")
		return
	}

	userID := getUserID(c)
	if err := h.authSvc.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		BadRequest(c, errcode.CodeForbidden, err.Error())
		return
	}
	Success(c, nil)
}

// ========== helpers ==========

func getUserID(c *gin.Context) uint64 {
	idI, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	id, ok := idI.(uint64)
	if !ok {
		return 0
	}
	return id
}

func getTenantID(c *gin.Context) uint64 {
	idI, exists := c.Get("tenant_id")
	if !exists {
		return 0
	}
	id, ok := idI.(uint64)
	if !ok {
		return 0
	}
	return id
}

func getRole(c *gin.Context) model.UserRole {
	r, _ := c.Get("role")
	role, _ := r.(model.UserRole)
	return role
}
```

- [ ] **Step 2: 验证编译通过**

Run: `cd D:/MyGoProject/CloudKey && go build ./...`
Expected: 编译通过

- [ ] **Step 3: Commit**

```bash
git add internal/handler/auth_handler.go
git commit -m "feat: add AuthHandler with unified login, 2FA, and profile endpoints"
```

---

### Task 6: 改造 service 层 — 加 tenant_id 过滤

**Files:**
- Modify: `internal/service/key_service.go`
- Modify: `internal/service/service_account_service.go`
- Modify: `internal/service/usage_log_service.go`
- Modify: `internal/service/login_log_service.go`
- Modify: `internal/service/stats_service.go`

**Interfaces:**
- 所有写操作从 `gin.Context` 或参数获取 `tenantID`
- Key CRUD/service/stats/usage log service 的所有 DB 查询加 `WHERE tenant_id = ?`

- [ ] **Step 1: 改造 `key_service.go` — 所有查询加 tenant_id 过滤**

修改 KeyService struct 的函数签名，全部加 `tenantID uint64` 参数：

**CreateKey** — 设置 `key.TenantID = tenantID`:
```go
func (s *KeyService) CreateKey(req CreateKeyRequest, tenantID uint64, keyPrefix string, keyLen, suffixLen int) (*CreateKeyResult, error) {
	// 使用传入的租户配置而非 service 默认值
	rawKey, err := s.generateRawKeyWithConfig(keyPrefix, keyLen)
	// ...
	key := model.Key{
		TenantID:        tenantID,
		Alias:           req.Alias,
		// ...
	}
	// ...
}
```

需要新增 `generateRawKeyWithConfig`:
```go
func (s *KeyService) generateRawKeyWithConfig(prefix string, keyLen int) (string, error) {
	bytes := make([]byte, keyLen)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random: %w", err)
	}
	return prefix + hex.EncodeToString(bytes), nil
}
```

**ListKeys** — 加 `tenantID`:
```go
func (s *KeyService) ListKeys(query KeyListQuery, tenantID uint64) ([]model.Key, int64, error) {
	// ...
	db := s.db.Model(&model.Key{}).Where("tenant_id = ?", tenantID)
	// ...
}
```

**GetKeyDetail, UpdateKey, DisableKey, EnableKey, DeleteKey** — 加 `tenantID`:
```go
func (s *KeyService) GetKeyDetail(id, tenantID uint64) (*model.Key, error) {
	var key model.Key
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}
```

类似更新所有 CRUD 方法。

**FindByRawKey — 公开接口不由 tenant admin 调用，暂不加 tenant 过滤（后续 ConsumeKey 关联时通过 Key 的 tenant_id 来隔离）**

**ConsumeKey** — 公开接口，消耗后记录 usage_log 时需写 tenant_id。通过 key 获取 tenant_id：
```go
// 在 record usage log 时:
key, _ := s.FindByRawKey(rawKey)
tenantID := uint64(0)
if key != nil { tenantID = key.TenantID }

// usageLogSvc.Record 传入 tenantID
```

改动 `RecordUsageParams`:
```go
type RecordUsageParams struct {
	TenantID       uint64
	KeyID          uint64
	KeyAlias       string
	Amount         int64
	IP             string
	UserAgent      string
	RequestPath    string
	RequestParams  string
	ResponseStatus int
}
```

**ExportKeys, ExportKeysJSON** — 加 `tenantID`:
```go
func (s *KeyService) ExportKeys(tenantID uint64) ([]model.Key, error) {
	var keys []model.Key
	if err := s.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}
```

- [ ] **Step 2: 改造 `service_account_service.go` — 加 tenant_id 过滤**

所有方法加 `tenantID` 参数:

**CreateServiceAccount**:
```go
func (s *ServiceAccountService) CreateServiceAccount(name string, tenantID uint64) (*model.ServiceAccount, string, error) {
	rawKey, err := s.GenerateServiceKey()
	// ...
	account := model.ServiceAccount{Name: name, KeyHash: hashServiceKey(rawKey), TenantID: tenantID}
	// ...
}
```

**ListServiceAccounts**:
```go
func (s *ServiceAccountService) ListServiceAccounts(tenantID uint64) ([]model.ServiceAccount, error) {
	var accounts []model.ServiceAccount
	if err := s.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}
```

**ToggleServiceAccount, DeleteServiceAccount** — 加 `tenantID`:
```go
func (s *ServiceAccountService) ToggleServiceAccount(id, tenantID uint64, isActive bool) error {
	return s.db.Model(&model.ServiceAccount{}).Where("id = ? AND tenant_id = ?", id, tenantID).Update("is_active", isActive).Error
}
```

- [ ] **Step 3: 改造 `usage_log_service.go` — Record 加 TenantID + 查询加 tenant_id**

**RecordUsageParams** 已有 TenantID 字段（Step 1 中已添加）:

**Record**:
```go
func (s *UsageLogService) Record(params RecordUsageParams) error {
	return s.db.Create(&model.UsageLog{
		TenantID:       params.TenantID,
		KeyID:          params.KeyID,
		// ...
	}).Error
}
```

**ListLogs, ExportLogs** — 加 `tenantID`:
```go
func (s *UsageLogService) ListLogs(query UsageLogQuery, tenantID uint64) ([]model.UsageLog, int64, error) {
	// ...
	db := s.db.Model(&model.UsageLog{}).Where("tenant_id = ?", tenantID)
	// ...
}
```

- [ ] **Step 4: 改造 `login_log_service.go` — 加 TenantID**

**RecordLogin**:
```go
func (s *LoginLogService) RecordLogin(userID uint64, tenantID *uint64, ip, userAgent string, success bool) error {
	status := model.LoginStatusFailed
	if success { status = model.LoginStatusSuccess }
	return s.db.Create(&model.LoginLog{
		UserID: userID, TenantID: tenantID, IP: ip, UserAgent: userAgent, Status: status,
	}).Error
}
```

**ListLoginLogs** — 加 `tenantID`（nullable for super_admin 查全部）:
```go
func (s *LoginLogService) ListLoginLogs(page, pageSize int, tenantID *uint64) ([]model.LoginLog, int64, error) {
	var logs []model.LoginLog
	var total int64
	db := s.db.Model(&model.LoginLog{})
	if tenantID != nil {
		db = db.Where("tenant_id = ?", *tenantID)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}
```

- [ ] **Step 5: 改造 `stats_service.go` — 加 tenant_id 过滤**

所有聚合查询加 `WHERE tenant_id = ?`:

**GetKeyOverview**:
```go
func (s *StatsService) GetKeyOverview(dateRange *DateRange, tenantID uint64) (*KeyOverview, error) {
	// ...
	keyDB := applyDateFilter(s.db.Model(&model.Key{}), dateRange).Where("tenant_id = ?", tenantID)
	// ...
}
```

类似更新 **GetTrends, GetTopKeys, GetTopIPs, GetDashboard** 全部加 `tenantID` 参数。

GetDashboard 和 GetTrends 中的 UsageLog 查询也需要加 `tenant_id` 过滤。

- [ ] ***Step 6: Commit earlier, then verify compilation. Update router.go to remove broken import paths and rerun: `go build ./...` (repeat until clean).***

Run: `cd D:/MyGoProject/CloudKey && go build ./...`
Expected: 编译通过（如果旧 handler 文件调用旧的 service 签名，会有编译错误 — 这是预期的，后续 task 会解决）

保留旧的 handler 文件暂时不删，它们会有编译错误。后续 task 创建新的 handler 后会删除旧文件。

- [ ] **Step 7: Commit**

```bash
git add internal/service/
git commit -m "feat: add tenant_id filtering to all service layer queries"
```

---

### Task 7: 创建 SuperHandler（系统管理员接口）+ 改造 KeyHandler 为 TenantHandler

**Files:**
- Rewrite: `internal/handler/key_handler.go` → 面向 tenant admin
- Create: `internal/handler/super_handler.go`
- Create: `internal/service/tenant_service.go`
- Modify: `internal/handler/response.go`（如需要）
- Modify: `internal/router/router.go`（暂时不在这里更新，留到 Task 9）

**Interfaces:**
- Produces: `SuperHandler` with `ListTenants`, `CreateTenant`, `GetTenant`, `UpdateTenant`, `ResetPassword`, `GetConfigs`, `UpdateConfigs`
- Produces: `TenantHandler` with `CreateKey`, `ListKeys`, `GetKey`, ..., `ListServiceAccounts`, `CreateServiceAccount`, ..., `Dashboard`, `Overview`, ...
- Produces: `TenantService` with `CreateTenant`, `GetTenant`, `ListTenants`, `UpdateTenant`, `ResetPassword`

- [ ] **Step 1: 创建 `internal/service/tenant_service.go`**

```go
package service

import (
	"CloudKey/internal/model"
	"fmt"
	"math/rand"

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
	Name   string `json:"name" binding:"required"`
}

type CreateTenantResult struct {
	Tenant         model.Tenant `json:"tenant"`
	AdminUsername  string       `json:"admin_username"`
	AdminPassword  string       `json:"admin_password"`
}

func (s *TenantService) CreateTenant(req CreateTenantRequest) (*CreateTenantResult, error) {
	// 生成租户管理员账号
	username := fmt.Sprintf("%s_admin", req.Name)

	// 检查用户名冲突
	var count int64
	s.db.Model(&model.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		username = fmt.Sprintf("%s_admin%d", req.Name, rand.Intn(9999))
	}

	password := generateRandomPassword(16)

	// 事务
	tx := s.db.Begin()

	tenant := model.Tenant{
		Name:      req.Name,
		Status:    model.TenantStatusActive,
		KeyPrefix: "sk-",
		KeyLength: 32,
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
```

(文件较长，此处省略完整内容，包含 `GetTenant`, `UpdateTenant`, `ResetPassword`, `generateRandomPassword`)

```go
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
	Name      *string             `json:"name"`
	Status    *model.TenantStatus `json:"status"`
	ExpireAt  *string             `json:"expire_at"`  // "2006-01-02 15:04:05" or "" to clear
	KeyPrefix *string             `json:"key_prefix"`
	KeyLength *int                `json:"key_length"`
	KeySuffixLength *int          `json:"key_suffix_length"`
}

func (s *TenantService) UpdateTenant(id uint64, req UpdateTenantRequest) error {
	updates := map[string]interface{}{}
	if req.Name != nil { updates["name"] = *req.Name }
	if req.Status != nil { updates["status"] = *req.Status }
	if req.KeyPrefix != nil { updates["key_prefix"] = *req.KeyPrefix }
	if req.KeyLength != nil { updates["key_length"] = *req.KeyLength }
	if req.KeySuffixLength != nil { updates["key_suffix_length"] = *req.KeySuffixLength }
	if req.ExpireAt != nil {
		if *req.ExpireAt == "" {
			updates["expire_at"] = nil
		} else {
			updates["expire_at"] = *req.ExpireAt
		}
	}
	if len(updates) == 0 { return nil }
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
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
```

- [ ] **Step 2: 创建 `internal/handler/super_handler.go`**

```go
package handler

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"CloudKey/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SuperHandler struct {
	tenantSvc    *service.TenantService
	configSvc    *service.ConfigService
	statsSvc     *service.StatsService
	loginLogSvc  *service.LoginLogService
}

func NewSuperHandler(tenantSvc *service.TenantService, configSvc *service.ConfigService, statsSvc *service.StatsService, loginLogSvc *service.LoginLogService) *SuperHandler {
	return &SuperHandler{tenantSvc: tenantSvc, configSvc: configSvc, statsSvc: statsSvc, loginLogSvc: loginLogSvc}
}

// GET /api/super/tenants
func (h *SuperHandler) ListTenants(c *gin.Context) {
	tenants, err := h.tenantSvc.ListTenants()
	if err != nil { InternalError(c); return }
	Success(c, tenants)
}

// POST /api/super/tenants
// 创建租户 + 自动生成管理员账号密码
func (h *SuperHandler) CreateTenant(c *gin.Context) {
	var req service.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeForbidden, "参数错误")
		return
	}
	result, err := h.tenantSvc.CreateTenant(req)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, gin.H{
		"tenant":         result.Tenant,
		"admin_username": result.AdminUsername,
		"admin_password": result.AdminPassword,
	})
}

// GET /api/super/tenants/:id
func (h *SuperHandler) GetTenant(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil { BadRequest(c, errcode.CodeTenantNotFound, "无效的ID"); return }
	tenant, err := h.tenantSvc.GetTenant(id)
	if err != nil { NotFound(c, errcode.CodeTenantNotFound, "租户不存在"); return }
	Success(c, tenant)
}

// PATCH /api/super/tenants/:id
func (h *SuperHandler) UpdateTenant(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil { BadRequest(c, errcode.CodeTenantNotFound, "无效的ID"); return }
	var req service.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil { BadRequest(c, errcode.CodeForbidden, "参数错误"); return }
	if err := h.tenantSvc.UpdateTenant(id, req); err != nil { InternalError(c); return }
	Success(c, nil)
}

// PATCH /api/super/tenants/:id/reset-password
func (h *SuperHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil { BadRequest(c, errcode.CodeTenantNotFound, "无效的ID"); return }
	newPass, err := h.tenantSvc.ResetPassword(id)
	if err != nil { InternalError(c); return }
	Success(c, gin.H{"new_password": newPass})
}

// GET /api/super/configs
func (h *SuperHandler) GetConfigs(c *gin.Context) {
	configs, err := h.configSvc.GetAllConfigs()
	if err != nil { InternalError(c); return }
	Success(c, configs)
}

// PUT /api/super/configs
func (h *SuperHandler) UpdateConfigs(c *gin.Context) {
	var req []struct {
		Key         string `json:"key" binding:"required"`
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil { BadRequest(c, errcode.CodeForbidden, "参数错误"); return }
	for _, item := range req {
		if err := h.configSvc.SetConfig(item.Key, item.Value, item.Description); err != nil {
			InternalError(c); return
		}
	}
	Success(c, nil)
}

// GET /api/super/login-logs
func (h *SuperHandler) LoginLogs(c *gin.Context) {
	page, pageSize := pageParams(c)
	logs, total, err := h.loginLogSvc.ListLoginLogs(page, pageSize, nil) // nil = 全部
	if err != nil { InternalError(c); return }
	SuccessPaginated(c, logs, total, page, pageSize)
}
```

- [ ] **Step 3: 重写 `internal/handler/key_handler.go` → 面向 tenant admin**

将 `KeyHandler` 重命名为 `TenantKeyHandler`，增加 tenant scope:

```go
type TenantKeyHandler struct {
	keySvc       *service.KeyService
	usageLogSvc  *service.UsageLogService
	recordParams bool
}

func NewTenantKeyHandler(keySvc *service.KeyService, usageLogSvc *service.UsageLogService, recordParams bool) *TenantKeyHandler {
	return &TenantKeyHandler{keySvc: keySvc, usageLogSvc: usageLogSvc, recordParams: recordParams}
}
```

所有方法加 tenant scope。例如 **CreateKey**:
```go
func (h *TenantKeyHandler) CreateKey(c *gin.Context) {
	var req CreateKeyJSON
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, http.StatusBadRequest, "参数错误")
		return
	}

	tenantID := getTenantID(c)

	// 从 context 获取租户的 Key 配置（可通过 getTenantPrefix 等获取）
	// 简化：先用硬编码默认值，后续从 middleware 注入
	createdBy := "tenant_admin"

	expireAt, _ := parseExpireAt(req.ExpireAt)
	result, err := h.keySvc.CreateKey(service.CreateKeyRequest{
		Alias: req.Alias, BillingMode: model.KeyBillingMode(req.BillingMode),
		InitialAmount: req.InitialAmount, CreatedBy: createdBy,
		ExpireAt: expireAt, MaxUsage: req.MaxUsage,
	}, tenantID, "sk-", 32, 4) // 后续从 tenant 获取
	if err != nil { InternalError(c); return }

	Success(c, gin.H{ ... })
}
```

类似更新 **ListKeys, GetKey, UpdateKey, DisableKey, EnableKey, DeleteKey, ExportKeys, ExportKeysJSON** 全部加 `tenantID := getTenantID(c)` 并传入 service。

公共接口 **Status** 和 **Consume** 保留在 `key_handler.go` 中（无需认证），但 Consume 时需 recording usage_log 传入 key 的 tenant_id。

- [ ] **Step 4: 验证编译**

Run: `cd D:/MyGoProject/CloudKey && go build ./...`
Expected: 部分编译错误（旧 handler 文件未删）。**先忽略，后续 Task 9 router 改造后统一清理。**

- [ ] **Step 5: Commit**

```bash
git add internal/service/tenant_service.go internal/handler/super_handler.go internal/handler/key_handler.go
git commit -m "feat: add SuperHandler, TenantService, refactor KeyHandler for tenant scope"
```

---

### Task 8: 改造 ServiceHandler + StatsHandler + UsageLogHandler（面向 tenant）

**Files:**
- Rewrite: `internal/handler/service_handler.go`
- Rewrite: `internal/handler/stats_handler.go`
- Rewrite: `internal/handler/usage_log_handler.go`
- Rewrite: `internal/handler/config_handler.go`

- [ ] **Step 1: 重写 `service_handler.go` — 面向 tenant admin**

改造所有方法加 tenant scope:

```go
type TenantServiceAccountHandler struct {
	keySvc            *service.KeyService
	serviceAccountSvc *service.ServiceAccountService
}

func NewTenantServiceAccountHandler(keySvc *service.KeyService, saSvc *service.ServiceAccountService) *TenantServiceAccountHandler {
	return &TenantServiceAccountHandler{keySvc: keySvc, serviceAccountSvc: saSvc}
}

// ListServiceAccounts
func (h *TenantServiceAccountHandler) ListServiceAccounts(c *gin.Context) {
	tenantID := getTenantID(c)
	accounts, err := h.serviceAccountSvc.ListServiceAccounts(tenantID)
	// ...
}

// CreateServiceAccount
func (h *TenantServiceAccountHandler) CreateServiceAccount(c *gin.Context) {
	var req struct { Name string `json:"name" binding:"required"` }
	// ...
	tenantID := getTenantID(c)
	account, rawKey, err := h.serviceAccountSvc.CreateServiceAccount(req.Name, tenantID)
	// ...
}

// ToggleServiceAccount, DeleteServiceAccount — 类似加 tenantID
```

Service account 的 key create/list 方法 **ServiceCreateKey, ServiceListKeys** — 保留，使用 `c.Get("service_account")` 获取 sa，从中取 `sa.TenantID`:

```go
func (h *TenantServiceAccountHandler) ServiceCreateKey(c *gin.Context) {
	saI, _ := c.Get("service_account")
	sa := saI.(*model.ServiceAccount)
	tenantID := sa.TenantID
	// ...
	result, err := h.keySvc.CreateKey(req, tenantID, "sk-", 32, 4)
	// ...
}

func (h *TenantServiceAccountHandler) ServiceListKeys(c *gin.Context) {
	saI, _ := c.Get("service_account")
	sa := saI.(*model.ServiceAccount)
	tenantID := sa.TenantID
	// 使用 tenantID 过滤 keys
	keys, total, err := h.keySvc.ListKeysByTenant(tenantID, page, pageSize)
	// ...
}
```

需要在 KeyService 新增 `ListKeysByTenant`:
```go
func (s *KeyService) ListKeysByTenant(tenantID uint64, page, pageSize int) ([]model.Key, int64, error) {
	var keys []model.Key
	var total int64
	db := s.db.Model(&model.Key{}).Where("tenant_id = ?", tenantID)
	db.Count(&total)
	offset := (page - 1) * pageSize
	db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&keys)
	return keys, total, nil
}
```

- [ ] **Step 2: 重写 `stats_handler.go` — 面向 tenant admin**

```go
type TenantStatsHandler struct {
	statsSvc *service.StatsService
}

func NewTenantStatsHandler(svc *service.StatsService) *TenantStatsHandler {
	return &TenantStatsHandler{statsSvc: svc}
}

func (h *TenantStatsHandler) Dashboard(c *gin.Context) {
	tenantID := getTenantID(c)
	dash, err := h.statsSvc.GetDashboard(tenantID)
	// ...
}

// Overview, Trends, TopKeys, TopIPs — 全部加 tenantID
```

- [ ] **Step 3: 重写 `usage_log_handler.go` — 面向 tenant admin**

```go
type TenantUsageLogHandler struct {
	usageLogSvc *service.UsageLogService
}

func (h *TenantUsageLogHandler) ListLogs(c *gin.Context) {
	tenantID := getTenantID(c)
	page, pageSize := pageParams(c)
	logs, total, err := h.usageLogSvc.ListLogs(service.UsageLogQuery{...}, tenantID)
	// ...
}

func (h *TenantUsageLogHandler) ExportLogs(c *gin.Context) {
	tenantID := getTenantID(c)
	logs, err := h.usageLogSvc.ExportLogs(service.UsageLogQuery{...}, tenantID)
	// ...
}
```

- [ ] **Step 4: 保留 config_handler.go（仅 super_admin 使用，已在 SuperHandler 中重新实现）**

旧 config_handler.go 可保留或等 Task 9 删除。

- [ ] **Step 5: 验证编译**

Run: `cd D:/MyGoProject/CloudKey && go build ./...`
Expected: 如果有旧文件引用错误，忽略（Task 9 统一处理）

- [ ] **Step 6: Commit**

```bash
git add internal/handler/
git commit -m "feat: refactor all handlers for tenant-scoped access"
```

---

### Task 9: 重构 router.go + main.go（全部组装）

**Files:**
- Rewrite: `internal/router/router.go`
- Rewrite: `main.go`

- [ ] **Step 1: 删除旧的 handler/admin/service 文件**

```bash
rm internal/model/admin.go
rm internal/service/admin_service.go
rm internal/handler/admin_handler.go
```

- [ ] **Step 2: 重写 `internal/router/router.go`**

```go
package router

import (
	"CloudKey/internal/handler"
	"CloudKey/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(
	authHandler *handler.AuthHandler,
	superHandler *handler.SuperHandler,
	tenantKeyHandler *handler.TenantKeyHandler,
	tenantSAHandler *handler.TenantServiceAccountHandler,
	tenantStatsHandler *handler.TenantStatsHandler,
	tenantUsageLogHandler *handler.TenantUsageLogHandler,
	jwtSecret string,
	db *gorm.DB,
	saSvc *service.ServiceAccountService,
) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	// 静态文件
	r.StaticFile("/", "web/admin.html")
	r.NoRoute(func(c *gin.Context) {
		if c.Request.URL.Path != "/" {
			if !strings.HasPrefix(c.Request.URL.Path, "/api") {
				http.ServeFile(c.Writer, c.Request, "web/admin.html")
				return
			}
		}
	})

	api := r.Group("/api")

	// ========== 认证（公开） ==========
	{
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/verify-2fa", authHandler.Verify2FA)
		api.POST("/auth/totp/setup-init", authHandler.SetupTOTPPublic)
		api.POST("/auth/totp/confirm-init", authHandler.ConfirmTOTPPublic)
	}

	// ========== 公共 API ==========
	{
		api.GET("/key/status", tenantKeyHandler.Status)   // 无 auth
		api.POST("/key/consume", tenantKeyHandler.Consume) // 无 auth
	}

	// ========== 系统管理员 ==========
	super := api.Group("/super")
	super.Use(middleware.AuthMiddleware(jwtSecret))
	super.Use(middleware.RequireSuperAdmin())
	{
		super.GET("/tenants", superHandler.ListTenants)
		super.POST("/tenants", superHandler.CreateTenant)
		super.GET("/tenants/:id", superHandler.GetTenant)
		super.PATCH("/tenants/:id", superHandler.UpdateTenant)
		super.PATCH("/tenants/:id/reset-password", superHandler.ResetPassword)

		super.GET("/configs", superHandler.GetConfigs)
		super.PUT("/configs", superHandler.UpdateConfigs)

		super.GET("/profile", authHandler.Profile)
		super.PUT("/password", authHandler.ChangePassword)
		super.POST("/totp/setup", authHandler.SetupTOTP)
		super.POST("/totp/confirm", authHandler.ConfirmTOTP)
		super.GET("/login-logs", superHandler.LoginLogs)
	}

	// ========== 租户管理员 ==========
	tenant := api.Group("/tenant")
	tenant.Use(middleware.AuthMiddleware(jwtSecret))
	tenant.Use(middleware.RequireTenantAdmin(db))
	{
		// Key 管理（业务操作加 BusinessGuard）
		tenantKeys := tenant.Group("/keys")
		tenantKeys.POST("", middleware.TenantBusinessGuard(db), tenantKeyHandler.CreateKey)
		tenantKeys.GET("", tenantKeyHandler.ListKeys)
		tenantKeys.GET("/export", tenantKeyHandler.ExportKeys)
		tenantKeys.GET("/:id", tenantKeyHandler.GetKey)
		tenantKeys.PATCH("/:id", middleware.TenantBusinessGuard(db), tenantKeyHandler.UpdateKey)
		tenantKeys.PATCH("/:id/disable", middleware.TenantBusinessGuard(db), tenantKeyHandler.DisableKey)
		tenantKeys.PATCH("/:id/enable", middleware.TenantBusinessGuard(db), tenantKeyHandler.EnableKey)
		tenantKeys.DELETE("/:id", middleware.TenantBusinessGuard(db), tenantKeyHandler.DeleteKey)

		// 服务账号（全部加 BusinessGuard）
		tenantSA := tenant.Group("/service-accounts")
		tenantSA.Use(middleware.TenantBusinessGuard(db))
		{
			tenantSA.GET("", tenantSAHandler.ListServiceAccounts)
			tenantSA.POST("", tenantSAHandler.CreateServiceAccount)
			tenantSA.PATCH("/:id/toggle", tenantSAHandler.ToggleServiceAccount)
			tenantSA.DELETE("/:id", tenantSAHandler.DeleteServiceAccount)
		}

		// 统计（不 guard，expired 可查看）
		tenantStats := tenant.Group("/stats")
		{
			tenantStats.GET("/dashboard", tenantStatsHandler.Dashboard)
			tenantStats.GET("/overview", tenantStatsHandler.Overview)
			tenantStats.GET("/trends", tenantStatsHandler.Trends)
			tenantStats.GET("/top-keys", tenantStatsHandler.TopKeys)
			tenantStats.GET("/top-ips", tenantStatsHandler.TopIPs)
		}

		// 使用日志（不 guard）
		tenantLogs := tenant.Group("/usage-logs")
		{
			tenantLogs.GET("", tenantUsageLogHandler.ListLogs)
			tenantLogs.GET("/export", tenantUsageLogHandler.ExportLogs)
		}

		// 个人设置
		tenant.GET("/profile", authHandler.Profile)
		tenant.PUT("/password", authHandler.ChangePassword)
		tenant.POST("/totp/setup", authHandler.SetupTOTP)
		tenant.POST("/totp/confirm", authHandler.ConfirmTOTP)
		tenant.GET("/login-logs", tenantUsageLogHandler.LoginLogs) // 或 superHandler
	}

	// ========== 服务账号 API ==========
	serviceAPI := api.Group("/service")
	serviceAPI.Use(middleware.ServiceAuthMiddleware(saSvc, db))
	{
		serviceAPI.POST("/keys", tenantSAHandler.ServiceCreateKey)
		serviceAPI.GET("/keys", tenantSAHandler.ServiceListKeys)
	}

	return r
}
```

- [ ] **Step 3: 重写 `main.go`**

```go
package main

import (
	"CloudKey/internal/config"
	"CloudKey/internal/database"
	"CloudKey/internal/handler"
	"CloudKey/internal/log"
	"CloudKey/internal/middleware"
	"CloudKey/internal/model"
	"CloudKey/internal/router"
	"CloudKey/internal/service"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := log.InitLogger(cfg.Log); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()
	defer log.Close()

	log.Info("CloudKey 启动中...")

	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatal("数据库连接失败", zap.Error(err))
	}
	defer database.Close(db)

	if err := model.AutoMigrate(db); err != nil {
		log.Fatal("数据库迁移失败", zap.Error(err))
	}
	log.Info("数据库迁移完成")

	// Services
	authSvc := service.NewAuthService(db, cfg.Auth.Secret, cfg.Auth.Expiration)
	keySvc := service.NewKeyService(db)
	usageLogSvc := service.NewUsageLogService(db)
	statsSvc := service.NewStatsService(db)
	serviceAccountSvc := service.NewServiceAccountService(db)
	configSvc := service.NewConfigService(db)
	loginLogSvc := service.NewLoginLogService(db)
	tenantSvc := service.NewTenantService(db)

	// Init defaults
	if err := configSvc.InitDefaultConfigs(); err != nil {
		log.Warn("初始化默认配置失败", zap.Error(err))
	}

	// Seed super admin
	superAdminUser := os.Getenv("SUPER_ADMIN_USERNAME")
	if superAdminUser == "" {
		superAdminUser = cfg.Auth.SuperAdminUsername
	}
	if superAdminUser == "" {
		superAdminUser = "admin"
	}
	superAdminPass := os.Getenv("SUPER_ADMIN_PASSWORD")
	if superAdminPass == "" {
		superAdminPass = cfg.Auth.SuperAdminPassword
	}
	if superAdminPass == "" {
		log.Fatal("请设置 SUPER_ADMIN_PASSWORD 环境变量或在 config.yaml 的 auth.super_admin_password 中配置")
	}
	if err := authSvc.SeedSuperAdmin(superAdminUser, superAdminPass); err != nil {
		log.Warn("创建超级管理员失败", zap.Error(err))
	}

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc, loginLogSvc)
	superHandler := handler.NewSuperHandler(tenantSvc, configSvc, statsSvc, loginLogSvc)
	tenantKeyHandler := handler.NewTenantKeyHandler(keySvc, usageLogSvc, false)
	tenantSAHandler := handler.NewTenantServiceAccountHandler(keySvc, serviceAccountSvc)
	tenantStatsHandler := handler.NewTenantStatsHandler(statsSvc)
	tenantUsageLogHandler := handler.NewTenantUsageLogHandler(usageLogSvc, loginLogSvc)

	// Gin mode
	if cfg.App.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Router
	r := router.SetupRouter(
		authHandler, superHandler,
		tenantKeyHandler, tenantSAHandler, tenantStatsHandler, tenantUsageLogHandler,
		cfg.Auth.Secret, db, serviceAccountSvc,
	)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Info("服务器启动", zap.String("address", addr))
	if err := r.Run(addr); err != nil {
		log.Fatal("服务器启动失败", zap.Error(err))
	}
}
```

- [ ] **Step 4: 更新 `config/config.go` — AuthConfig 加 SuperAdmin 字段**

```go
type AuthConfig struct {
	Secret             string `yaml:"secret" mapstructure:"secret"`
	Expiration         int    `yaml:"expiration" mapstructure:"expiration"`
	SuperAdminUsername string `yaml:"super_admin_username" mapstructure:"super_admin_username"`
	SuperAdminPassword string `yaml:"super_admin_password" mapstructure:"super_admin_password"`
}
```

- [ ] **Step 5: 更新 `config.yaml.example`**

```yaml
auth:
  secret: "change-me-to-a-random-string"
  expiration: 24
  super_admin_username: "admin"
  super_admin_password: "change-me"
```

- [ ] **Step 6: 验证编译**

Run: `cd D:/MyGoProject/CloudKey && go build ./...`
Expected: 编译通过，无错误

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: wire up new router and main.go for multi-tenant SaaS"
```

---

### Task 10: 编译 + 启动验证 + 测试

**Files:** 无新文件创建

- [ ] **Step 1: 完整编译**

Run: `cd D:/MyGoProject/CloudKey && go build -o cloudkey.exe .`
Expected: 编译成功，生成 cloudkey.exe

- [ ] **Step 2: 启动服务验证**（需要 MySQL 运行）

Run: `cd D:/MyGoProject/CloudKey && set SUPER_ADMIN_PASSWORD=test123 && ./cloudkey.exe`
Expected: "服务器启动" 日志，无 panic

- [ ] **Step 3: 测试统一登录 API**

```bash
# 首次登录（TOTP 未设置）
curl -X POST http://localhost:8080/api/auth/login -H "Content-Type: application/json" -d '{"username":"admin","password":"test123"}'
# Expected: {"code":0,"data":{"need_setup":true,"user_id":1,...}}

# 设置 TOTP
curl -X POST http://localhost:8080/api/auth/totp/setup-init -H "Content-Type: application/json" -d '{"user_id":1}'
# Expected: {"code":0,"data":{"secret":"...","url":"otpauth://..."}}
```

- [ ] **Step 4: 测试创建租户**

```bash
# 先用 TOTP 完成登录获取 JWT
curl -X POST http://localhost:8080/api/super/tenants -H "Authorization: Bearer <jwt>" -H "Content-Type: application/json" -d '{"name":"myapp"}'
# Expected: {"code":0,"data":{"tenant":{...},"admin_username":"myapp_admin","admin_password":"..."}}
```

- [ ] **Step 5: 测试租户管理员登录 + Key CRUD**

创建租户后，用返回的账号密码登录租户，测试 Key 创建/查看/消耗。

- [ ] **Step 6: 处理发现的问题并修复**

- [ ] **Step 7: Final commit**

```bash
git add -A && git commit -m "feat: complete multi-tenant SaaS transformation"
```

---

## Self-Review Checklist

在交付前需确认:
1. 所有模型加 `tenant_id` (Task 2) ✅
2. 所有 service 查询加 tenant 过滤 (Task 6) ✅
3. JWT Claims 含 role + tenant_id (Task 3 + 4) ✅
4. 中间件链: Auth → RequireSuperAdmin/RequireTenantAdmin → TenantBusinessGuard (Task 3) ✅
5. 系统管理员不能访问业务数据 (Task 7: SuperHandler 无 key/service 接口) ✅
6. 统一登录入口 /api/auth/login (Task 5) ✅
7. 错误码扩展 (Task 3 Step 1) ✅
8. 编译通过 (Task 9) ✅
