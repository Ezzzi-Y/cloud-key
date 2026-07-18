# CloudKey 卡密发放与验证平台 — 完整实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 CloudKey 卡密发放与验证平台的完整后端服务，从数据模型到 HTTP API 到 Docker 部署。

**Architecture:** 三层架构 — `internal/model` 纯 GORM 数据模型（无业务逻辑），`internal/service` 业务逻辑层（依赖注入 DB），`internal/handler` Gin HTTP 处理器（薄层，委托给 service）。`internal/middleware` 处理 JWT 认证和服务账号认证。`internal/errcode` 集中定义错误码常量供 handler 和 service 共用。

**Tech Stack:** Go 1.25, Gin, GORM, MySQL 8.0+, golang.org/x/crypto (bcrypt), golang-jwt/v5, pquerna/otp, swaggo

## 已完成模块

- `internal/config/` — 配置加载（Viper），需补充 `App` 字段
- `internal/log/` — 日志（Zap + Lumberjack）
- `internal/database/` — GORM MySQL 连接 + 连接池

## 全局约束

- 卡密明文仅在创建时返回一次，数据库中只存储 SHA256 哈希值（确定性哈希，支持按哈希查找）
- 扣减操作必须使用数据库事务 + 乐观锁（version 字段）保证并发安全
- 所有 API 返回统一响应格式 `{ "code": 0, "message": "success", "data": {...} }`
- 错误码严格遵循规格文档定义（1001~9999）
- 日志中不记录完整卡密，只记录后缀和别名
- 格式化使用 `gofmt`，每个任务提交一次
- 提交信息格式：`feat(<scope>): <description>`
- 开发顺序严格：models → services → handlers → middleware → router → main → deploy
- 测试使用 SQLite 内存数据库（仅测试依赖，不影响生产环境 MySQL-only）

---

## 文件结构总览

```
CloudKey/
├── main.go                              # 应用入口
├── internal/
│   ├── config/
│   │   ├── config.go                    # ✅ 已有，需补充 App 字段
│   │   └── config_test.go               # ✅ 已有
│   ├── log/
│   │   ├── logger.go                    # ✅ 已有
│   │   └── logger_test.go               # ✅ 已有
│   ├── database/
│   │   ├── database.go                  # ✅ 已有（MySQL only）
│   │   └── database_test.go             # ✅ 已有
│   ├── errcode/                         # 新增 — 错误码常量
│   │   └── errcode.go
│   ├── model/                           # 新增 — GORM 数据模型
│   │   ├── key.go                       # 卡密模型
│   │   ├── usage_log.go                 # 使用记录模型
│   │   ├── admin.go                     # 管理员模型
│   │   ├── service_account.go           # 服务账号模型
│   │   ├── login_log.go                 # 登录日志模型
│   │   ├── config.go                    # 系统配置模型
│   │   └── migrate.go                   # AutoMigrate 函数
│   ├── service/                         # 新增 — 业务逻辑层
│   │   ├── key_service.go               # 卡密生成/查询/扣减/管理
│   │   ├── key_service_test.go
│   │   ├── usage_log_service.go         # 使用记录
│   │   ├── stats_service.go             # 数据统计
│   │   ├── admin_service.go             # 管理员认证(JWT+TOTP)
│   │   ├── admin_service_test.go
│   │   ├── service_account_service.go   # 服务账号
│   │   ├── login_log_service.go         # 登录日志
│   │   └── config_service.go            # 系统配置
│   ├── handler/                         # 新增 — HTTP 处理器
│   │   ├── response.go                  # 统一响应 + 分页
│   │   ├── key_handler.go               # 公开 + 管理接口
│   │   ├── usage_log_handler.go         # 使用记录
│   │   ├── stats_handler.go             # 统计
│   │   ├── admin_handler.go             # 管理员认证
│   │   ├── service_handler.go           # 服务账号
│   │   └── config_handler.go            # 系统配置
│   ├── middleware/                       # 新增 — 中间件
│   │   ├── auth.go                      # JWT 认证
│   │   ├── service_auth.go              # 服务账号认证
│   │   └── cors.go                      # CORS
│   └── router/
│       └── router.go                    # 路由注册
├── Dockerfile
├── docker-compose.yml
├── config.yaml.example
├── go.mod                               # ✅ 已有
└── go.sum                               # ✅ 已有
```

---

## 阶段一：基础设施 + 数据模型层 (Task 1-6)

---

### Task 1: 安装依赖 + 补充配置

**Files:**
- Modify: `go.mod`（添加依赖）
- Modify: `internal/config/config.go`（添加 AppSettings 字段）

**Interfaces:**
- Produces: go.mod 包含 Gin, JWT, TOTP, bcrypt, SQLite(测试) 依赖
- Produces: `config.AppSettings` struct, `config.AppConfig.App` field

- [ ] **Step 1: 添加 Gin 依赖**

```bash
go get github.com/gin-gonic/gin@latest
```

- [ ] **Step 2: 添加 JWT 依赖**

```bash
go get github.com/golang-jwt/jwt/v5@latest
```

- [ ] **Step 3: 添加 TOTP 依赖**

```bash
go get github.com/pquerna/otp@latest
```

- [ ] **Step 4: 添加 bcrypt 依赖**

```bash
go get golang.org/x/crypto@latest
```

- [ ] **Step 5: 添加 SQLite 测试依赖**

```bash
go get gorm.io/driver/sqlite@latest
```

- [ ] **Step 6: 运行 go mod tidy**

```bash
go mod tidy
```

- [ ] **Step 7: 补充 config.go — 添加 AppSettings**

在 `internal/config/config.go` 的 `AppConfig` struct 中添加 `App` 字段：

```go
type AppConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Log      LogConfig      `yaml:"log"`
	Auth     AuthConfig     `yaml:"auth"`
	Security SecurityConfig `yaml:"security"`
	App      AppSettings    `yaml:"app"`
}

// AppSettings 应用级别设置
type AppSettings struct {
	Debug bool `yaml:"debug"`
}
```

- [ ] **Step 8: 验证编译**

```bash
go build ./...
```

- [ ] **Step 9: 提交**

```bash
git add go.mod go.sum internal/config/config.go
git commit -m "feat(deps): add Gin, JWT, TOTP, bcrypt, SQLite(test) deps; add AppSettings config"
```

---

### Task 2: 错误码包

**Files:**
- Create: `internal/errcode/errcode.go`

**Interfaces:**
- Produces: 错误码常量 `CodeSuccess`, `CodeKeyNotFound` 等，供 handler 和 service 共用

- [ ] **Step 1: 创建 errcode 目录**

```bash
mkdir -p internal/errcode
```

- [ ] **Step 2: 编写 errcode.go**

```go
package errcode

const (
	CodeSuccess = 0

	// 卡密相关 1001~1999
	CodeKeyNotFound     = 1001
	CodeKeyDisabled     = 1002
	CodeKeyExhausted    = 1003
	CodeKeyInsufficient = 1004

	// 管理员认证相关 2001~2999
	CodeInvalidCredentials = 2001
	CodeTOTPFailed         = 2002
	CodeJWTInvalid         = 2003
	CodeForbidden          = 2004

	// 服务账号相关 3001~3999
	CodeServiceKeyInvalid = 3001

	// 系统 9999
	CodeInternalError = 9999
)

var codeMessages = map[int]string{
	CodeSuccess:            "success",
	CodeKeyNotFound:        "卡密不存在",
	CodeKeyDisabled:        "卡密已禁用",
	CodeKeyExhausted:       "卡密额度已用尽",
	CodeKeyInsufficient:    "扣减数量超过剩余额度",
	CodeInvalidCredentials: "管理员账号或密码错误",
	CodeTOTPFailed:         "TOTP 验证失败",
	CodeJWTInvalid:         "JWT Token 无效或已过期",
	CodeForbidden:          "无权限执行此操作",
	CodeServiceKeyInvalid:  "服务账号密钥无效",
	CodeInternalError:      "系统内部错误",
}

func GetMessage(code int) string {
	if msg, ok := codeMessages[code]; ok {
		return msg
	}
	return "未知错误"
}
```

- [ ] **Step 3: 格式化并编译**

```bash
gofmt -w internal/errcode/errcode.go
go build ./internal/errcode/
```

- [ ] **Step 4: 提交**

```bash
git add internal/errcode/
git commit -m "feat(errcode): add shared error code constants"
```

---

### Task 3: 卡密数据模型

**Files:**
- Create: `internal/model/key.go`

**Interfaces:**
- Produces: `model.Key` struct (GORM model, 表名 `keys`)
- Produces: `model.KeyBillingMode`, `model.KeyStatus` 类型和常量

- [ ] **Step 1: 创建 model 目录**

```bash
mkdir -p internal/model
```

- [ ] **Step 2: 编写 key.go**

```go
package model

import "time"

type KeyBillingMode string

const (
	BillingModeCount  KeyBillingMode = "count"
	BillingModeCredit KeyBillingMode = "credit"
)

type KeyStatus string

const (
	KeyStatusUnused   KeyStatus = "unused"
	KeyStatusUsed     KeyStatus = "used"
	KeyStatusDisabled KeyStatus = "disabled"
	KeyStatusExpired  KeyStatus = "expired"
)

type Key struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
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
}

func (Key) TableName() string { return "keys" }

func (k *Key) IsUsable() bool {
	return k.Status == KeyStatusUnused && k.RemainingAmount > 0
}

func (k *Key) CanDeduct(amount int64) bool {
	return k.IsUsable() && k.RemainingAmount >= amount
}
```

- [ ] **Step 3: 格式化并编译**

```bash
gofmt -w internal/model/key.go
go build ./internal/model/
```

- [ ] **Step 4: 提交**

```bash
git add internal/model/key.go
git commit -m "feat(model): add Key GORM model with optimistic lock"
```

---

### Task 4: 使用记录数据模型

**Files:**
- Create: `internal/model/usage_log.go`

**Interfaces:**
- Produces: `model.UsageLog` struct (表名 `usage_logs`)

- [ ] **Step 1: 编写 usage_log.go**

```go
package model

import "time"

type UsageLog struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
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

func (UsageLog) TableName() string { return "usage_logs" }
```

- [ ] **Step 2: 格式化并编译**

```bash
gofmt -w internal/model/usage_log.go
go build ./internal/model/
```

- [ ] **Step 3: 提交**

```bash
git add internal/model/usage_log.go
git commit -m "feat(model): add UsageLog GORM model"
```

---

### Task 5: 管理员 + 服务账号 + 登录日志 + 系统配置数据模型

**Files:**
- Create: `internal/model/admin.go`
- Create: `internal/model/service_account.go`
- Create: `internal/model/login_log.go`
- Create: `internal/model/config.go`

**Interfaces:**
- Produces: `model.Admin`, `model.ServiceAccount`, `model.LoginLog`, `model.SysConfig` structs

- [ ] **Step 1: 编写 admin.go**

```go
package model

import "time"

type Admin struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	TotpSecret   string    `gorm:"type:varchar(255)" json:"-"`
	TotpSetup    bool      `gorm:"default:false" json:"totp_setup"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Admin) TableName() string { return "admins" }
```

- [ ] **Step 2: 编写 service_account.go**

```go
package model

import "time"

type ServiceAccount struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	KeyHash   string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ServiceAccount) TableName() string { return "service_accounts" }
```

- [ ] **Step 3: 编写 login_log.go**

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
	AdminID   uint64      `gorm:"type:bigint;index;not null" json:"admin_id"`
	IP        string      `gorm:"type:varchar(50);not null" json:"ip"`
	UserAgent string      `gorm:"type:varchar(500)" json:"user_agent"`
	Status    LoginStatus `gorm:"type:varchar(20);not null" json:"status"`
	CreatedAt time.Time   `gorm:"autoCreateTime" json:"created_at"`
}

func (LoginLog) TableName() string { return "login_logs" }
```

- [ ] **Step 4: 编写 config.go (系统配置模型)**

```go
package model

import "time"

type SysConfig struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Key         string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"key"`
	Value       string    `gorm:"type:varchar(500);not null" json:"value"`
	Description string    `gorm:"type:varchar(500)" json:"description"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SysConfig) TableName() string { return "configs" }
```

- [ ] **Step 5: 格式化并编译**

```bash
gofmt -w internal/model/admin.go internal/model/service_account.go internal/model/login_log.go internal/model/config.go
go build ./internal/model/
```

- [ ] **Step 6: 提交**

```bash
git add internal/model/
git commit -m "feat(model): add Admin, ServiceAccount, LoginLog, SysConfig GORM models"
```

---

### Task 6: 数据库自动迁移

**Files:**
- Create: `internal/model/migrate.go`

**Interfaces:**
- Produces: `model.AutoMigrate(db *gorm.DB) error`

- [ ] **Step 1: 编写 migrate.go**

```go
package model

import "gorm.io/gorm"

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Key{},
		&UsageLog{},
		&Admin{},
		&ServiceAccount{},
		&LoginLog{},
		&SysConfig{},
	)
}
```

- [ ] **Step 2: 格式化并编译**

```bash
gofmt -w internal/model/migrate.go
go build ./internal/model/
```

- [ ] **Step 3: 提交**

```bash
git add internal/model/migrate.go
git commit -m "feat(model): add AutoMigrate for all models"
```

---

## 阶段二：统一响应 + 服务层 (Task 7-11)

---

### Task 7: 统一响应格式

**Files:**
- Create: `internal/handler/response.go`

**Interfaces:**
- Produces: `handler.Success()`, `handler.Error()`, `handler.SuccessPaginated()`, `handler.BadRequest()` 等响应函数
- Consumes: `errcode.CodeSuccess` (Task 2)

- [ ] **Step 1: 创建 handler 目录**

```bash
mkdir -p internal/handler
```

- [ ] **Step 2: 编写 response.go**

```go
package handler

import (
	"CloudKey/internal/errcode"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

func SuccessPaginated(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	Success(c, PageData{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{Code: code, Message: message, Data: nil})
}

func BadRequest(c *gin.Context, code int, message string) {
	Error(c, http.StatusBadRequest, code, message)
}

func Unauthorized(c *gin.Context, code int, message string) {
	Error(c, http.StatusUnauthorized, code, message)
}

func NotFound(c *gin.Context, code int, message string) {
	Error(c, http.StatusNotFound, code, message)
}

func InternalError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, errcode.CodeInternalError, errcode.GetMessage(errcode.CodeInternalError))
}
```

- [ ] **Step 3: 格式化并编译**

```bash
gofmt -w internal/handler/response.go
go build ./internal/handler/
```

- [ ] **Step 4: 提交**

```bash
git add internal/handler/response.go
git commit -m "feat(handler): add unified response format with error helpers"
```

---

### Task 8: 卡密服务 — 生成、创建、查询

**Files:**
- Create: `internal/service/key_service.go`

**Interfaces:**
- Produces: `service.KeyService`, `service.NewKeyService(db)`, `service.CreateKeyRequest`, `service.CreateKeyResult`
- Produces: `GenerateRawKey()`, `CreateKey()`, `FindByRawKey()`, `GetKeyStatus()` 方法
- Consumes: `model.Key`, `model.KeyBillingMode`, `model.KeyStatus`

- [ ] **Step 1: 创建 service 目录**

```bash
mkdir -p internal/service
```

- [ ] **Step 2: 编写 key_service.go（第一部分：生成 + 创建 + 查询）**

```go
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
	Alias         string              `json:"alias"`
	BillingMode   model.KeyBillingMode `json:"billing_mode"`
	InitialAmount int64               `json:"initial_amount"`
	CreatedBy     string              `json:"created_by"`
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
	Alias           string              `json:"alias"`
	BillingMode     model.KeyBillingMode `json:"billing_mode"`
	RemainingAmount int64               `json:"remaining_amount"`
	Status          model.KeyStatus     `json:"status"`
	CreatedAt       string              `json:"created_at"`
	UsedAt          *string             `json:"used_at"`
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
```

- [ ] **Step 3: 格式化并编译**

```bash
gofmt -w internal/service/key_service.go
go build ./internal/service/
```

- [ ] **Step 4: 提交**

```bash
git add internal/service/key_service.go
git commit -m "feat(service): add KeyService with key generation, creation, and status query"
```

---

### Task 9: 卡密服务 — 扣减（并发安全）

**Files:**
- Modify: `internal/service/key_service.go`

**Interfaces:**
- Produces: `ConsumeKey(rawKey, amount) (*ConsumeResult, int, error)` — 使用事务 + 乐观锁
- Consumes: `errcode.CodeKeyNotFound` 等 (Task 2)

- [ ] **Step 1: 在 key_service.go 末尾追加扣减方法**

```go
import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	// ... 保持其他 import
)

type ConsumeResult struct {
	RemainingAmount int64           `json:"remaining_amount"`
	Status          model.KeyStatus `json:"status"`
	Exhausted       bool            `json:"exhausted"`
}

func (s *KeyService) ConsumeKey(rawKey string, amount int64) (*ConsumeResult, int, error) {
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
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	result := tx.Model(&model.Key{}).
		Where("id = ? AND version = ?", key.ID, key.Version).
		Updates(map[string]interface{}{
			"remaining_amount": gorm.Expr("remaining_amount - ?", amount),
			"version":          gorm.Expr("version + 1"),
			"status":           gorm.Expr("CASE WHEN remaining_amount - ? <= 0 THEN 'used' ELSE status END", amount),
			"used_at":          gorm.Expr("NOW()"),
		})

	if result.Error != nil {
		tx.Rollback()
		return nil, 0, result.Error
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return s.ConsumeKey(rawKey, amount) // 乐观锁冲突，重试一次
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
```

- [ ] **Step 2: 格式化并编译**

```bash
gofmt -w internal/service/key_service.go
go build ./internal/service/
```

- [ ] **Step 3: 提交**

```bash
git add internal/service/key_service.go
git commit -m "feat(service): add ConsumeKey with transaction + optimistic locking"
```

---

### Task 10: 卡密服务 — 管理操作

**Files:**
- Modify: `internal/service/key_service.go`

**Interfaces:**
- Produces: `ListKeys()`, `ListKeysByCreatedBy()`, `GetKeyDetail()`, `UpdateKey()`, `DisableKey()`, `EnableKey()`, `DeleteKey()`, `ExportKeys()`

- [ ] **Step 1: 在 key_service.go 末尾追加管理操作**

```go
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

func (s *KeyService) ExportKeys() ([]model.Key, error) {
	var keys []model.Key
	if err := s.db.Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}
```

- [ ] **Step 2: 格式化并编译**

```bash
gofmt -w internal/service/key_service.go
go build ./internal/service/
```

- [ ] **Step 3: 提交**

```bash
git add internal/service/key_service.go
git commit -m "feat(service): add key management operations (list, update, disable, enable, delete, export)"
```

---

### Task 11: 使用记录 + 统计 + 管理员 + 服务账号 + 登录日志 + 配置服务

**Files:**
- Create: `internal/service/usage_log_service.go`
- Create: `internal/service/stats_service.go`
- Create: `internal/service/admin_service.go`
- Create: `internal/service/service_account_service.go`
- Create: `internal/service/login_log_service.go`
- Create: `internal/service/config_service.go`

**Interfaces:**
- Produces: `UsageLogService`, `StatsService`, `AdminService`, `ServiceAccountService`, `LoginLogService`, `ConfigService`

- [ ] **Step 1: 编写 usage_log_service.go**

```go
package service

import (
	"CloudKey/internal/model"
	"time"

	"gorm.io/gorm"
)

type UsageLogService struct {
	db *gorm.DB
}

func NewUsageLogService(db *gorm.DB) *UsageLogService {
	return &UsageLogService{db: db}
}

type RecordUsageParams struct {
	KeyID          uint64
	KeyAlias       string
	Amount         int64
	IP             string
	UserAgent      string
	RequestPath    string
	RequestParams  string
	ResponseStatus int
}

func (s *UsageLogService) Record(params RecordUsageParams) error {
	return s.db.Create(&model.UsageLog{
		KeyID:          params.KeyID,
		KeyAlias:       params.KeyAlias,
		Amount:         params.Amount,
		IP:             params.IP,
		UserAgent:      params.UserAgent,
		RequestPath:    params.RequestPath,
		RequestParams:  params.RequestParams,
		ResponseStatus: params.ResponseStatus,
		CreatedAt:      time.Now(),
	}).Error
}

type UsageLogQuery struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	KeyAlias  string `form:"key_alias"`
	IP        string `form:"ip"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
}

func (s *UsageLogService) ListLogs(query UsageLogQuery) ([]model.UsageLog, int64, error) {
	var logs []model.UsageLog
	var total int64

	db := s.db.Model(&model.UsageLog{})
	if query.KeyAlias != "" {
		db = db.Where("key_alias = ?", query.KeyAlias)
	}
	if query.IP != "" {
		db = db.Where("ip = ?", query.IP)
	}
	if query.StartTime != "" {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != "" {
		db = db.Where("created_at <= ?", query.EndTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (s *UsageLogService) ExportLogs(query UsageLogQuery) ([]model.UsageLog, error) {
	var logs []model.UsageLog
	db := s.db.Model(&model.UsageLog{})
	if query.KeyAlias != "" {
		db = db.Where("key_alias = ?", query.KeyAlias)
	}
	if query.IP != "" {
		db = db.Where("ip = ?", query.IP)
	}
	if query.StartTime != "" {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != "" {
		db = db.Where("created_at <= ?", query.EndTime)
	}
	if err := db.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
```

- [ ] **Step 2: 编写 stats_service.go**

```go
package service

import (
	"CloudKey/internal/model"
	"time"

	"gorm.io/gorm"
)

type StatsService struct {
	db *gorm.DB
}

func NewStatsService(db *gorm.DB) *StatsService {
	return &StatsService{db: db}
}

type KeyOverview struct {
	TotalKeys    int64            `json:"total_keys"`
	StatusCounts map[string]int64 `json:"status_counts"`
	TotalInitial int64            `json:"total_initial"`
	TotalRemain  int64            `json:"total_remaining"`
}

func (s *StatsService) GetKeyOverview() (*KeyOverview, error) {
	ov := &KeyOverview{StatusCounts: make(map[string]int64)}

	s.db.Model(&model.Key{}).Count(&ov.TotalKeys)

	var rows []struct {
		Status string
		Count  int64
	}
	s.db.Model(&model.Key{}).Select("status, COUNT(*) as count").Group("status").Scan(&rows)
	for _, r := range rows {
		ov.StatusCounts[r.Status] = r.Count
	}

	s.db.Model(&model.Key{}).Select("COALESCE(SUM(initial_amount), 0)").Scan(&ov.TotalInitial)
	s.db.Model(&model.Key{}).Select("COALESCE(SUM(remaining_amount), 0)").Scan(&ov.TotalRemain)

	return ov, nil
}

type TrendPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

func (s *StatsService) GetTrends(period string) ([]TrendPoint, error) {
	var dateFormat string
	var startTime time.Time
	now := time.Now()

	switch period {
	case "week":
		dateFormat = "%Y-%m-%d"
		startTime = now.AddDate(0, 0, -7)
	case "month":
		dateFormat = "%Y-%m-%d"
		startTime = now.AddDate(0, -1, 0)
	default:
		dateFormat = "%Y-%m-%d %H"
		startTime = now.AddDate(0, 0, -1)
	}

	var points []TrendPoint
	s.db.Model(&model.UsageLog{}).
		Select("DATE_FORMAT(created_at, ?) as date, COUNT(*) as count", dateFormat).
		Where("created_at >= ?", startTime).
		Group("date").Order("date ASC").Scan(&points)

	return points, nil
}

type TopItem struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

func (s *StatsService) GetTopKeys() ([]TopItem, error) {
	var items []TopItem
	s.db.Model(&model.UsageLog{}).
		Select("key_alias as name, COUNT(*) as count").
		Group("key_alias").Order("count DESC").Limit(10).Scan(&items)
	return items, nil
}

func (s *StatsService) GetTopIPs() ([]TopItem, error) {
	var items []TopItem
	s.db.Model(&model.UsageLog{}).
		Select("ip as name, COUNT(*) as count").
		Group("ip").Order("count DESC").Limit(10).Scan(&items)
	return items, nil
}

type DashboardStats struct {
	Overview   *KeyOverview     `json:"overview"`
	TodayCalls int64            `json:"today_calls"`
	WeekCalls  int64            `json:"week_calls"`
	MonthCalls int64            `json:"month_calls"`
	RecentLogs []model.UsageLog `json:"recent_logs"`
}

func (s *StatsService) GetDashboard() (*DashboardStats, error) {
	overview, err := s.GetKeyOverview()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := now.AddDate(0, 0, -7)
	monthStart := now.AddDate(0, -1, 0)

	var todayCalls, weekCalls, monthCalls int64
	s.db.Model(&model.UsageLog{}).Where("created_at >= ?", todayStart).Count(&todayCalls)
	s.db.Model(&model.UsageLog{}).Where("created_at >= ?", weekStart).Count(&weekCalls)
	s.db.Model(&model.UsageLog{}).Where("created_at >= ?", monthStart).Count(&monthCalls)

	var recentLogs []model.UsageLog
	s.db.Order("created_at DESC").Limit(20).Find(&recentLogs)

	return &DashboardStats{
		Overview:   overview,
		TodayCalls: todayCalls,
		WeekCalls:  weekCalls,
		MonthCalls: monthCalls,
		RecentLogs: recentLogs,
	}, nil
}
```

- [ ] **Step 3: 编写 admin_service.go**

```go
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
```

- [ ] **Step 4: 编写 service_account_service.go**

```go
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

func hashServiceKey(key string) string {
	h := hmac.New(sha256.New, []byte(key))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *ServiceAccountService) GenerateServiceKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "svc-" + hex.EncodeToString(bytes), nil
}

func (s *ServiceAccountService) CreateServiceAccount(name string) (*model.ServiceAccount, string, error) {
	rawKey, err := s.GenerateServiceKey()
	if err != nil {
		return nil, "", err
	}

	account := model.ServiceAccount{Name: name, KeyHash: hashServiceKey(rawKey)}
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

func (s *ServiceAccountService) ListServiceAccounts() ([]model.ServiceAccount, error) {
	var accounts []model.ServiceAccount
	if err := s.db.Order("created_at DESC").Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (s *ServiceAccountService) ToggleServiceAccount(id uint64, isActive bool) error {
	return s.db.Model(&model.ServiceAccount{}).Where("id = ?", id).Update("is_active", isActive).Error
}

func (s *ServiceAccountService) DeleteServiceAccount(id uint64) error {
	return s.db.Delete(&model.ServiceAccount{}, id).Error
}
```

- [ ] **Step 5: 编写 login_log_service.go**

```go
package service

import (
	"CloudKey/internal/model"

	"gorm.io/gorm"
)

type LoginLogService struct {
	db *gorm.DB
}

func NewLoginLogService(db *gorm.DB) *LoginLogService {
	return &LoginLogService{db: db}
}

func (s *LoginLogService) RecordLogin(adminID uint64, ip, userAgent string, success bool) error {
	status := model.LoginStatusFailed
	if success {
		status = model.LoginStatusSuccess
	}
	return s.db.Create(&model.LoginLog{
		AdminID: adminID, IP: ip, UserAgent: userAgent, Status: status,
	}).Error
}

func (s *LoginLogService) ListLoginLogs(page, pageSize int) ([]model.LoginLog, int64, error) {
	var logs []model.LoginLog
	var total int64

	s.db.Model(&model.LoginLog{}).Count(&total)
	offset := (page - 1) * pageSize
	s.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs)

	return logs, total, nil
}
```

- [ ] **Step 6: 编写 config_service.go**

```go
package service

import (
	"CloudKey/internal/model"

	"gorm.io/gorm"
)

type ConfigService struct {
	db *gorm.DB
}

func NewConfigService(db *gorm.DB) *ConfigService {
	return &ConfigService{db: db}
}

func (s *ConfigService) GetConfig(key string) (string, error) {
	var cfg model.SysConfig
	if err := s.db.Where("`key` = ?", key).First(&cfg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return cfg.Value, nil
}

func (s *ConfigService) GetAllConfigs() ([]model.SysConfig, error) {
	var configs []model.SysConfig
	if err := s.db.Order("id ASC").Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

func (s *ConfigService) SetConfig(key, value, description string) error {
	var cfg model.SysConfig
	result := s.db.Where("`key` = ?", key).First(&cfg)
	if result.Error == gorm.ErrRecordNotFound {
		return s.db.Create(&model.SysConfig{Key: key, Value: value, Description: description}).Error
	}
	if result.Error != nil {
		return result.Error
	}
	updates := map[string]interface{}{"value": value}
	if description != "" {
		updates["description"] = description
	}
	return s.db.Model(&cfg).Updates(updates).Error
}

func (s *ConfigService) InitDefaultConfigs() error {
	defaults := []model.SysConfig{
		{Key: "key_prefix", Value: "sk-", Description: "卡密默认前缀"},
		{Key: "key_length", Value: "32", Description: "卡密随机部分长度"},
		{Key: "key_suffix_length", Value: "4", Description: "卡密后缀长度"},
		{Key: "record_request_params", Value: "false", Description: "是否记录请求参数"},
		{Key: "jwt_expire_hours", Value: "24", Description: "JWT 过期时间（小时）"},
	}

	for _, d := range defaults {
		var count int64
		s.db.Model(&model.SysConfig{}).Where("`key` = ?", d.Key).Count(&count)
		if count == 0 {
			if err := s.db.Create(&d).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
```

- [ ] **Step 7: 格式化并编译全部服务**

```bash
gofmt -w internal/service/
go build ./internal/service/
```

- [ ] **Step 8: 提交**

```bash
git add internal/service/
git commit -m "feat(service): add all service layer (usage_log, stats, admin, service_account, login_log, config)"
```

---

## 阶段三：服务层测试 (Task 12)

---

### Task 12: 服务层单元测试

**Files:**
- Create: `internal/service/key_service_test.go`
- Create: `internal/service/admin_service_test.go`

**Interfaces:**
- Tests: KeyService 生成/创建/查询/扣减；AdminService 密码验证/JWT

**注意:** 测试使用 SQLite 内存数据库，Task 1 已安装 `gorm.io/driver/sqlite`。SQLite 不支持 MySQL 特有的 `CASE WHEN ... THEN ... ELSE ... END` SQL 表达式（部分语法兼容），因此扣减测试可能需要适配。

- [ ] **Step 1: 编写 key_service_test.go**

```go
package service

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		alias TEXT NOT NULL,
		key_hash TEXT NOT NULL UNIQUE,
		key_prefix TEXT NOT NULL,
		key_suffix TEXT NOT NULL,
		billing_mode TEXT NOT NULL,
		initial_amount INTEGER NOT NULL,
		remaining_amount INTEGER NOT NULL,
		version INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'unused',
		created_by TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		used_at DATETIME
	)`)
	return db
}

func TestCreateKey(t *testing.T) {
	db := setupTestDB(t)
	svc := NewKeyService(db)

	result, err := svc.CreateKey(CreateKeyRequest{
		Alias: "test-key", BillingMode: "count", InitialAmount: 100, CreatedBy: "admin",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.RawKey == "" {
		t.Error("expected raw key")
	}
	if result.Key.ID == 0 {
		t.Error("expected key ID > 0")
	}
	if result.Key.RemainingAmount != 100 {
		t.Errorf("expected 100 remaining, got %d", result.Key.RemainingAmount)
	}
}

func TestFindByRawKey(t *testing.T) {
	db := setupTestDB(t)
	svc := NewKeyService(db)

	result, _ := svc.CreateKey(CreateKeyRequest{
		Alias: "find-test", BillingMode: "count", InitialAmount: 50, CreatedBy: "admin",
	})

	found, err := svc.FindByRawKey(result.RawKey)
	if err != nil {
		t.Fatal(err)
	}
	if found == nil {
		t.Fatal("expected to find key")
	}
	if found.Alias != "find-test" {
		t.Errorf("expected alias 'find-test', got %s", found.Alias)
	}
}

func TestFindByRawKey_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewKeyService(db)

	found, err := svc.FindByRawKey("sk-nonexistent-key")
	if err != nil {
		t.Fatal(err)
	}
	if found != nil {
		t.Error("expected nil for nonexistent key")
	}
}

func TestGetKeyStatus(t *testing.T) {
	db := setupTestDB(t)
	svc := NewKeyService(db)

	result, _ := svc.CreateKey(CreateKeyRequest{
		Alias: "status-test", BillingMode: "count", InitialAmount: 10, CreatedBy: "admin",
	})

	status, err := svc.GetKeyStatus(result.RawKey)
	if err != nil {
		t.Fatal(err)
	}
	if status == nil {
		t.Fatal("expected status")
	}
	if status.Alias != "status-test" {
		t.Errorf("expected alias 'status-test', got %s", status.Alias)
	}
	if status.RemainingAmount != 10 {
		t.Errorf("expected 10 remaining, got %d", status.RemainingAmount)
	}
}

func TestListKeys(t *testing.T) {
	db := setupTestDB(t)
	svc := NewKeyService(db)

	for i := 0; i < 5; i++ {
		svc.CreateKey(CreateKeyRequest{
			Alias: "key-" + string(rune('A'+i)), BillingMode: "count", InitialAmount: 10, CreatedBy: "admin",
		})
	}

	keys, total, err := svc.ListKeys(KeyListQuery{Page: 1, PageSize: 3})
	if err != nil {
		t.Fatal(err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(keys) != 3 {
		t.Errorf("expected 3 keys on page 1, got %d", len(keys))
	}
}

func TestDisableEnableKey(t *testing.T) {
	db := setupTestDB(t)
	svc := NewKeyService(db)

	result, _ := svc.CreateKey(CreateKeyRequest{
		Alias: "toggle-test", BillingMode: "count", InitialAmount: 10, CreatedBy: "admin",
	})

	if err := svc.DisableKey(result.Key.ID); err != nil {
		t.Fatal(err)
	}

	key, _ := svc.GetKeyDetail(result.Key.ID)
	if key.Status != "disabled" {
		t.Errorf("expected disabled, got %s", key.Status)
	}

	if err := svc.EnableKey(result.Key.ID); err != nil {
		t.Fatal(err)
	}

	key, _ = svc.GetKeyDetail(result.Key.ID)
	if key.Status != "unused" {
		t.Errorf("expected unused, got %s", key.Status)
	}
}
```

- [ ] **Step 2: 编写 admin_service_test.go**

```go
package service

import (
	"CloudKey/internal/model"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAdminTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS admins (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		totp_secret TEXT,
		totp_setup INTEGER DEFAULT 0,
		is_active INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	return db
}

func TestSeedAdmin(t *testing.T) {
	db := setupAdminTestDB(t)
	svc := NewAdminService(db, "test-secret", 24)

	if err := svc.SeedAdmin("admin", "admin123"); err != nil {
		t.Fatal(err)
	}

	// 再次 seed 不应重复
	if err := svc.SeedAdmin("admin", "admin123"); err != nil {
		t.Fatal(err)
	}

	var count int64
	db.Model(&model.Admin{}).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 admin, got %d", count)
	}
}

func TestLogin_Success(t *testing.T) {
	db := setupAdminTestDB(t)
	svc := NewAdminService(db, "test-secret", 24)
	svc.SeedAdmin("admin", "admin123")

	result, err := svc.Login("admin", "admin123")
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected login result")
	}
	if result.RequireTOTP {
		t.Error("new admin should not require TOTP")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	db := setupAdminTestDB(t)
	svc := NewAdminService(db, "test-secret", 24)
	svc.SeedAdmin("admin", "admin123")

	result, _ := svc.Login("admin", "wrongpass")
	if result != nil {
		t.Error("expected nil for wrong password")
	}
}

func TestChangePassword(t *testing.T) {
	db := setupAdminTestDB(t)
	svc := NewAdminService(db, "test-secret", 24)
	svc.SeedAdmin("admin", "admin123")

	if err := svc.ChangePassword(1, "admin123", "newpass456"); err != nil {
		t.Fatal(err)
	}

	result, _ := svc.Login("admin", "admin123")
	if result != nil {
		t.Error("old password should not work")
	}

	result, _ = svc.Login("admin", "newpass456")
	if result == nil {
		t.Error("new password should work")
	}
}
```

- [ ] **Step 3: 运行测试**

```bash
go test ./internal/service/ -v -count=1
```

Expected: 全部 PASS

- [ ] **Step 4: 提交**

```bash
git add internal/service/key_service_test.go internal/service/admin_service_test.go
git commit -m "test(service): add unit tests for KeyService and AdminService"
```

---

## 阶段四：Handler 层 (Task 13-17)

---

### Task 13: 公开接口 Handler（卡密查询与扣减）

**Files:**
- Create: `internal/handler/key_handler.go`

**Interfaces:**
- Consumes: `service.KeyService`, `service.UsageLogService`
- Produces: `KeyHandler.Status()`, `KeyHandler.Consume()`

- [ ] **Step 1: 编写 key_handler.go**

```go
package handler

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"CloudKey/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type KeyHandler struct {
	keySvc       *service.KeyService
	usageLogSvc  *service.UsageLogService
	recordParams bool
}

func NewKeyHandler(keySvc *service.KeyService, usageLogSvc *service.UsageLogService, recordParams bool) *KeyHandler {
	return &KeyHandler{keySvc: keySvc, usageLogSvc: usageLogSvc, recordParams: recordParams}
}

// Status 查询卡密状态（不扣减）
func (h *KeyHandler) Status(c *gin.Context) {
	rawKey := c.Query("sk")
	if rawKey == "" {
		BadRequest(c, errcode.CodeKeyNotFound, "缺少卡密参数")
		return
	}

	result, err := h.keySvc.GetKeyStatus(rawKey)
	if err != nil {
		InternalError(c)
		return
	}
	if result == nil {
		NotFound(c, errcode.CodeKeyNotFound, errcode.GetMessage(errcode.CodeKeyNotFound))
		return
	}

	Success(c, result)
}

type ConsumeRequest struct {
	Key    string `json:"key" binding:"required"`
	Amount int64  `json:"amount"`
}

// Consume 扣减卡密额度
func (h *KeyHandler) Consume(c *gin.Context) {
	var req ConsumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeKeyNotFound, "参数错误")
		return
	}
	if req.Amount <= 0 {
		req.Amount = 1
	}

	result, code, err := h.keySvc.ConsumeKey(req.Key, req.Amount)
	if err != nil {
		InternalError(c)
		return
	}
	if code != 0 {
		key, _ := h.keySvc.FindByRawKey(req.Key)
		keyID, keyAlias := uint64(0), ""
		if key != nil {
			keyID, keyAlias = key.ID, key.Alias
		}
		h.usageLogSvc.Record(service.RecordUsageParams{
			KeyID: keyID, KeyAlias: keyAlias, Amount: req.Amount,
			IP: c.ClientIP(), UserAgent: c.GetHeader("User-Agent"),
			RequestPath: c.Request.URL.Path, ResponseStatus: code,
		})
		BadRequest(c, code, errcode.GetMessage(code))
		return
	}

	key, _ := h.keySvc.FindByRawKey(req.Key)
	keyID, keyAlias := uint64(0), ""
	if key != nil {
		keyID, keyAlias = key.ID, key.Alias
	}
	h.usageLogSvc.Record(service.RecordUsageParams{
		KeyID: keyID, KeyAlias: keyAlias, Amount: req.Amount,
		IP: c.ClientIP(), UserAgent: c.GetHeader("User-Agent"),
		RequestPath: c.Request.URL.Path, ResponseStatus: http.StatusOK,
	})

	Success(c, result)
}

// ========== 管理员接口 ==========

type CreateKeyJSON struct {
	Alias         string `json:"alias" binding:"required"`
	BillingMode   string `json:"billing_mode" binding:"required"`
	InitialAmount int64  `json:"initial_amount" binding:"required"`
}

// CreateKey 管理员创建卡密
func (h *KeyHandler) CreateKey(c *gin.Context) {
	var req CreateKeyJSON
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeKeyNotFound, "参数错误")
		return
	}

	adminID, _ := c.Get("admin_id")
	createdBy := ""
	if adminID != nil {
		createdBy = "admin"
	}

	result, err := h.keySvc.CreateKey(service.CreateKeyRequest{
		Alias: req.Alias, BillingMode: model.KeyBillingMode(req.BillingMode),
		InitialAmount: req.InitialAmount, CreatedBy: createdBy,
	})
	if err != nil {
		InternalError(c)
		return
	}

	Success(c, gin.H{
		"id": result.Key.ID, "raw_key": result.RawKey, "alias": result.Key.Alias,
		"key_prefix": result.Key.KeyPrefix, "key_suffix": result.Key.KeySuffix,
		"billing_mode": result.Key.BillingMode, "initial_amount": result.Key.InitialAmount,
		"remaining_amount": result.Key.RemainingAmount, "status": result.Key.Status,
		"created_by": result.Key.CreatedBy, "created_at": result.Key.CreatedAt,
	})
}

// ListKeys 管理员查询卡密列表
func (h *KeyHandler) ListKeys(c *gin.Context) {
	page, pageSize := pageParams(c)
	keys, total, err := h.keySvc.ListKeys(service.KeyListQuery{
		Page: page, PageSize: pageSize,
		Status: c.Query("status"), Search: c.Query("search"),
	})
	if err != nil {
		InternalError(c)
		return
	}
	SuccessPaginated(c, keys, total, page, pageSize)
}

// GetKey 管理员查看卡密详情
func (h *KeyHandler) GetKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, errcode.CodeKeyNotFound, "无效的卡密 ID")
		return
	}

	key, err := h.keySvc.GetKeyDetail(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, errcode.CodeKeyNotFound, errcode.GetMessage(errcode.CodeKeyNotFound))
			return
		}
		InternalError(c)
		return
	}
	Success(c, key)
}

// UpdateKey 管理员修改卡密
func (h *KeyHandler) UpdateKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, errcode.CodeKeyNotFound, "无效的卡密 ID")
		return
	}

	var req service.UpdateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeKeyNotFound, "参数错误")
		return
	}

	if err := h.keySvc.UpdateKey(id, req); err != nil {
		InternalError(c)
		return
	}
	Success(c, nil)
}

// DisableKey 管理员禁用卡密
func (h *KeyHandler) DisableKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, errcode.CodeKeyNotFound, "无效的卡密 ID")
		return
	}
	if err := h.keySvc.DisableKey(id); err != nil {
		InternalError(c)
		return
	}
	Success(c, nil)
}

// EnableKey 管理员启用卡密
func (h *KeyHandler) EnableKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, errcode.CodeKeyNotFound, "无效的卡密 ID")
		return
	}
	if err := h.keySvc.EnableKey(id); err != nil {
		InternalError(c)
		return
	}
	Success(c, nil)
}

// DeleteKey 管理员删除卡密
func (h *KeyHandler) DeleteKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, errcode.CodeKeyNotFound, "无效的卡密 ID")
		return
	}
	if err := h.keySvc.DeleteKey(id); err != nil {
		InternalError(c)
		return
	}
	Success(c, nil)
}

// ExportKeys 管理员导出卡密
func (h *KeyHandler) ExportKeys(c *gin.Context) {
	keys, err := h.keySvc.ExportKeys()
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, keys)
}

func pageParams(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}
```

- [ ] **Step 2: 格式化并编译**

```bash
gofmt -w internal/handler/key_handler.go
go build ./internal/handler/
```

- [ ] **Step 3: 提交**

```bash
git add internal/handler/key_handler.go
git commit -m "feat(handler): add public + admin key handlers (status, consume, CRUD)"
```

---

### Task 14: 使用记录 + 统计 Handler

**Files:**
- Create: `internal/handler/usage_log_handler.go`
- Create: `internal/handler/stats_handler.go`

- [ ] **Step 1: 编写 usage_log_handler.go**

```go
package handler

import (
	"CloudKey/internal/service"

	"github.com/gin-gonic/gin"
)

type UsageLogHandler struct {
	usageLogSvc *service.UsageLogService
}

func NewUsageLogHandler(svc *service.UsageLogService) *UsageLogHandler {
	return &UsageLogHandler{usageLogSvc: svc}
}

func (h *UsageLogHandler) ListLogs(c *gin.Context) {
	page, pageSize := pageParams(c)

	logs, total, err := h.usageLogSvc.ListLogs(service.UsageLogQuery{
		Page: page, PageSize: pageSize,
		KeyAlias: c.Query("key_alias"), IP: c.Query("ip"),
		StartTime: c.Query("start_time"), EndTime: c.Query("end_time"),
	})
	if err != nil {
		InternalError(c)
		return
	}
	SuccessPaginated(c, logs, total, page, pageSize)
}

func (h *UsageLogHandler) ExportLogs(c *gin.Context) {
	logs, err := h.usageLogSvc.ExportLogs(service.UsageLogQuery{
		KeyAlias: c.Query("key_alias"), IP: c.Query("ip"),
		StartTime: c.Query("start_time"), EndTime: c.Query("end_time"),
	})
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, logs)
}
```

- [ ] **Step 2: 编写 stats_handler.go**

```go
package handler

import (
	"CloudKey/internal/service"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	statsSvc *service.StatsService
}

func NewStatsHandler(svc *service.StatsService) *StatsHandler {
	return &StatsHandler{statsSvc: svc}
}

func (h *StatsHandler) Dashboard(c *gin.Context) {
	dash, err := h.statsSvc.GetDashboard()
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, dash)
}

func (h *StatsHandler) Overview(c *gin.Context) {
	overview, err := h.statsSvc.GetKeyOverview()
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, overview)
}

func (h *StatsHandler) Trends(c *gin.Context) {
	period := c.DefaultQuery("period", "today")
	points, err := h.statsSvc.GetTrends(period)
	if err != nil {
		InternalError(c)
		return
	}
	if points == nil {
		points = make([]service.TrendPoint, 0)
	}
	Success(c, points)
}

func (h *StatsHandler) TopKeys(c *gin.Context) {
	items, err := h.statsSvc.GetTopKeys()
	if err != nil {
		InternalError(c)
		return
	}
	if items == nil {
		items = make([]service.TopItem, 0)
	}
	Success(c, items)
}

func (h *StatsHandler) TopIPs(c *gin.Context) {
	items, err := h.statsSvc.GetTopIPs()
	if err != nil {
		InternalError(c)
		return
	}
	if items == nil {
		items = make([]service.TopItem, 0)
	}
	Success(c, items)
}
```

- [ ] **Step 3: 格式化并编译**

```bash
gofmt -w internal/handler/usage_log_handler.go internal/handler/stats_handler.go
go build ./internal/handler/
```

- [ ] **Step 4: 提交**

```bash
git add internal/handler/usage_log_handler.go internal/handler/stats_handler.go
git commit -m "feat(handler): add usage log and stats handlers"
```

---

### Task 15: 管理员认证 Handler

**Files:**
- Create: `internal/handler/admin_handler.go`

**Interfaces:**
- Consumes: `service.AdminService`, `service.LoginLogService`
- Produces: `AdminHandler` — Login, Verify2FA, Profile, ChangePassword, SetupTOTP, ConfirmTOTP, LoginLogs

- [ ] **Step 1: 编写 admin_handler.go**

```go
package handler

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminSvc    *service.AdminService
	loginLogSvc *service.LoginLogService
}

func NewAdminHandler(adminSvc *service.AdminService, loginLogSvc *service.LoginLogService) *AdminHandler {
	return &AdminHandler{adminSvc: adminSvc, loginLogSvc: loginLogSvc}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AdminHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeInvalidCredentials, "参数错误")
		return
	}

	result, err := h.adminSvc.Login(req.Username, req.Password)
	if err != nil {
		InternalError(c)
		return
	}
	if result == nil {
		h.loginLogSvc.RecordLogin(0, c.ClientIP(), c.GetHeader("User-Agent"), false)
		Unauthorized(c, errcode.CodeInvalidCredentials, errcode.GetMessage(errcode.CodeInvalidCredentials))
		return
	}

	h.loginLogSvc.RecordLogin(result.AdminID, c.ClientIP(), c.GetHeader("User-Agent"), !result.RequireTOTP)

	if result.RequireTOTP {
		Success(c, gin.H{"require_totp": true, "admin_id": result.AdminID})
		return
	}

	Success(c, gin.H{
		"require_totp": false, "need_setup": true,
		"admin_id": result.AdminID, "message": "请设置 TOTP 两步验证",
	})
}

type Verify2FARequest struct {
	AdminID uint64 `json:"admin_id" binding:"required"`
	Code    string `json:"code" binding:"required"`
}

func (h *AdminHandler) Verify2FA(c *gin.Context) {
	var req Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeTOTPFailed, "参数错误")
		return
	}

	token, err := h.adminSvc.VerifyTOTP(req.AdminID, req.Code)
	if err != nil {
		InternalError(c)
		return
	}
	if token == "" {
		Unauthorized(c, errcode.CodeTOTPFailed, errcode.GetMessage(errcode.CodeTOTPFailed))
		return
	}

	Success(c, gin.H{"token": token, "token_type": "Bearer"})
}

func (h *AdminHandler) Profile(c *gin.Context) {
	adminID, _ := c.Get("admin_id")
	id, ok := adminID.(uint64)
	if !ok {
		Unauthorized(c, errcode.CodeJWTInvalid, errcode.GetMessage(errcode.CodeJWTInvalid))
		return
	}

	admin, err := h.adminSvc.GetAdminProfile(id)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, admin)
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (h *AdminHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeForbidden, "参数错误")
		return
	}

	adminID, _ := c.Get("admin_id")
	id, _ := adminID.(uint64)

	if err := h.adminSvc.ChangePassword(id, req.OldPassword, req.NewPassword); err != nil {
		BadRequest(c, errcode.CodeForbidden, err.Error())
		return
	}
	Success(c, nil)
}

func (h *AdminHandler) SetupTOTP(c *gin.Context) {
	adminID, _ := c.Get("admin_id")
	id, _ := adminID.(uint64)

	secret, url, err := h.adminSvc.GenerateTOTPSecret(id)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, gin.H{"secret": secret, "url": url})
}

func (h *AdminHandler) ConfirmTOTP(c *gin.Context) {
	var req Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeTOTPFailed, "参数错误")
		return
	}

	adminID, _ := c.Get("admin_id")
	id, _ := adminID.(uint64)

	if err := h.adminSvc.ConfirmTOTPSetup(id, req.Code); err != nil {
		BadRequest(c, errcode.CodeTOTPFailed, err.Error())
		return
	}
	Success(c, nil)
}

func (h *AdminHandler) LoginLogs(c *gin.Context) {
	page, pageSize := pageParams(c)
	logs, total, err := h.loginLogSvc.ListLoginLogs(page, pageSize)
	if err != nil {
		InternalError(c)
		return
	}
	SuccessPaginated(c, logs, total, page, pageSize)
}
```

- [ ] **Step 2: 格式化并编译**

```bash
gofmt -w internal/handler/admin_handler.go
go build ./internal/handler/
```

- [ ] **Step 3: 提交**

```bash
git add internal/handler/admin_handler.go
git commit -m "feat(handler): add admin auth handlers (login, TOTP, change password)"
```

---

### Task 16: 服务账号 + 系统配置 Handler

**Files:**
- Create: `internal/handler/service_handler.go`
- Create: `internal/handler/config_handler.go`

- [ ] **Step 1: 编写 service_handler.go**

```go
package handler

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"CloudKey/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ServiceHandler struct {
	keySvc            *service.KeyService
	serviceAccountSvc *service.ServiceAccountService
}

func NewServiceHandler(keySvc *service.KeyService, saSvc *service.ServiceAccountService) *ServiceHandler {
	return &ServiceHandler{keySvc: keySvc, serviceAccountSvc: saSvc}
}

func (h *ServiceHandler) ServiceCreateKey(c *gin.Context) {
	sa, _ := c.Get("service_account")

	var req struct {
		Alias         string `json:"alias" binding:"required"`
		BillingMode   string `json:"billing_mode" binding:"required"`
		InitialAmount int64  `json:"initial_amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeServiceKeyInvalid, "参数错误")
		return
	}

	createdBy := "sa:" + sa.(*model.ServiceAccount).Name

	result, err := h.keySvc.CreateKey(service.CreateKeyRequest{
		Alias: req.Alias, BillingMode: model.KeyBillingMode(req.BillingMode),
		InitialAmount: req.InitialAmount, CreatedBy: createdBy,
	})
	if err != nil {
		InternalError(c)
		return
	}

	Success(c, gin.H{
		"id": result.Key.ID, "raw_key": result.RawKey, "alias": result.Key.Alias,
		"key_suffix": result.Key.KeySuffix, "billing_mode": result.Key.BillingMode,
		"initial_amount": result.Key.InitialAmount, "remaining_amount": result.Key.RemainingAmount,
		"status": result.Key.Status,
	})
}

func (h *ServiceHandler) ServiceListKeys(c *gin.Context) {
	sa, _ := c.Get("service_account")
	createdBy := "sa:" + sa.(*model.ServiceAccount).Name
	page, pageSize := pageParams(c)

	keys, total, err := h.keySvc.ListKeysByCreatedBy(createdBy, page, pageSize)
	if err != nil {
		InternalError(c)
		return
	}
	SuccessPaginated(c, keys, total, page, pageSize)
}

func (h *ServiceHandler) ListServiceAccounts(c *gin.Context) {
	accounts, err := h.serviceAccountSvc.ListServiceAccounts()
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, accounts)
}

func (h *ServiceHandler) CreateServiceAccount(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeServiceKeyInvalid, "参数错误")
		return
	}

	account, rawKey, err := h.serviceAccountSvc.CreateServiceAccount(req.Name)
	if err != nil {
		InternalError(c)
		return
	}

	Success(c, gin.H{
		"id": account.ID, "name": account.Name, "raw_key": rawKey,
		"is_active": account.IsActive, "created_at": account.CreatedAt,
	})
}

func (h *ServiceHandler) ToggleServiceAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, errcode.CodeServiceKeyInvalid, "无效的服务账号 ID")
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeServiceKeyInvalid, "参数错误")
		return
	}

	if err := h.serviceAccountSvc.ToggleServiceAccount(id, req.IsActive); err != nil {
		InternalError(c)
		return
	}
	Success(c, nil)
}

func (h *ServiceHandler) DeleteServiceAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, errcode.CodeServiceKeyInvalid, "无效的服务账号 ID")
		return
	}

	if err := h.serviceAccountSvc.DeleteServiceAccount(id); err != nil {
		InternalError(c)
		return
	}
	Success(c, nil)
}
```

- [ ] **Step 2: 编写 config_handler.go**

```go
package handler

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/service"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	configSvc *service.ConfigService
}

func NewConfigHandler(svc *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{configSvc: svc}
}

func (h *ConfigHandler) GetConfigs(c *gin.Context) {
	configs, err := h.configSvc.GetAllConfigs()
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, configs)
}

func (h *ConfigHandler) UpdateConfigs(c *gin.Context) {
	var req []struct {
		Key         string `json:"key" binding:"required"`
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeForbidden, "参数错误")
		return
	}

	for _, item := range req {
		if err := h.configSvc.SetConfig(item.Key, item.Value, item.Description); err != nil {
			InternalError(c)
			return
		}
	}
	Success(c, nil)
}
```

- [ ] **Step 3: 格式化并编译**

```bash
gofmt -w internal/handler/service_handler.go internal/handler/config_handler.go
go build ./internal/handler/
```

- [ ] **Step 4: 提交**

```bash
git add internal/handler/service_handler.go internal/handler/config_handler.go
git commit -m "feat(handler): add service account and system config handlers"
```

---

## 阶段五：中间件 + 路由 + 入口 (Task 17-19)

---

### Task 17: 中间件层

**Files:**
- Create: `internal/middleware/auth.go`
- Create: `internal/middleware/service_auth.go`
- Create: `internal/middleware/cors.go`

- [ ] **Step 1: 创建 middleware 目录**

```bash
mkdir -p internal/middleware
```

- [ ] **Step 2: 编写 auth.go**

```go
package middleware

import (
	"CloudKey/internal/errcode"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	AdminID  uint64 `json:"admin_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusOK, gin.H{"code": errcode.CodeJWTInvalid, "message": errcode.GetMessage(errcode.CodeJWTInvalid), "data": nil})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusOK, gin.H{"code": errcode.CodeJWTInvalid, "message": errcode.GetMessage(errcode.CodeJWTInvalid), "data": nil})
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
			c.JSON(http.StatusOK, gin.H{"code": errcode.CodeJWTInvalid, "message": errcode.GetMessage(errcode.CodeJWTInvalid), "data": nil})
			c.Abort()
			return
		}

		c.Set("admin_id", claims.AdminID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
```

- [ ] **Step 3: 编写 service_auth.go**

```go
package middleware

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ServiceAuthMiddleware(svc *service.ServiceAccountService) gin.HandlerFunc {
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

		c.Set("service_account", account)
		c.Next()
	}
}
```

- [ ] **Step 4: 编写 cors.go**

```go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Service-Key")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
```

- [ ] **Step 5: 格式化并编译**

```bash
gofmt -w internal/middleware/
go build ./internal/middleware/
```

- [ ] **Step 6: 提交**

```bash
git add internal/middleware/
git commit -m "feat(middleware): add JWT auth, service auth, and CORS middleware"
```

---

### Task 18: 路由注册

**Files:**
- Create: `internal/router/router.go`

**Interfaces:**
- Consumes: 所有 handler + middleware
- Produces: `router.SetupRouter(...)` 函数

- [ ] **Step 1: 创建 router 目录**

```bash
mkdir -p internal/router
```

- [ ] **Step 2: 编写 router.go**

```go
package router

import (
	"CloudKey/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	keyHandler *handler.KeyHandler,
	usageLogHandler *handler.UsageLogHandler,
	statsHandler *handler.StatsHandler,
	adminHandler *handler.AdminHandler,
	serviceHandler *handler.ServiceHandler,
	configHandler *handler.ConfigHandler,
	adminAuthMW gin.HandlerFunc,
	serviceAuthMW gin.HandlerFunc,
) *gin.Engine {
	r := gin.Default()

	// 公开接口（无需认证）
	api := r.Group("/api")
	{
		api.GET("/key/status", keyHandler.Status)
		api.POST("/key/consume", keyHandler.Consume)
	}

	// 管理后台登录（无需认证）
	adminPublic := r.Group("/api/admin")
	{
		adminPublic.POST("/login", adminHandler.Login)
		adminPublic.POST("/login/verify-2fa", adminHandler.Verify2FA)
	}

	// 管理后台（需 JWT 认证）
	adminAuth := r.Group("/api/admin")
	adminAuth.Use(adminAuthMW)
	{
		// 管理员自身
		adminAuth.GET("/profile", adminHandler.Profile)
		adminAuth.PUT("/password", adminHandler.ChangePassword)
		adminAuth.POST("/totp/setup", adminHandler.SetupTOTP)
		adminAuth.POST("/totp/confirm", adminHandler.ConfirmTOTP)
		adminAuth.GET("/login-logs", adminHandler.LoginLogs)

		// 卡密管理
		adminAuth.POST("/keys", keyHandler.CreateKey)
		adminAuth.GET("/keys", keyHandler.ListKeys)
		adminAuth.GET("/keys/export", keyHandler.ExportKeys)
		adminAuth.GET("/keys/:id", keyHandler.GetKey)
		adminAuth.PATCH("/keys/:id", keyHandler.UpdateKey)
		adminAuth.PATCH("/keys/:id/disable", keyHandler.DisableKey)
		adminAuth.PATCH("/keys/:id/enable", keyHandler.EnableKey)
		adminAuth.DELETE("/keys/:id", keyHandler.DeleteKey)

		// 使用记录
		adminAuth.GET("/usage-logs", usageLogHandler.ListLogs)
		adminAuth.GET("/usage-logs/export", usageLogHandler.ExportLogs)

		// 数据统计
		adminAuth.GET("/stats/dashboard", statsHandler.Dashboard)
		adminAuth.GET("/stats/overview", statsHandler.Overview)
		adminAuth.GET("/stats/trends", statsHandler.Trends)
		adminAuth.GET("/stats/top-keys", statsHandler.TopKeys)
		adminAuth.GET("/stats/top-ips", statsHandler.TopIPs)

		// 系统管理
		adminAuth.GET("/configs", configHandler.GetConfigs)
		adminAuth.PUT("/configs", configHandler.UpdateConfigs)

		// 服务账号管理
		adminAuth.GET("/service-accounts", serviceHandler.ListServiceAccounts)
		adminAuth.POST("/service-accounts", serviceHandler.CreateServiceAccount)
		adminAuth.PATCH("/service-accounts/:id/toggle", serviceHandler.ToggleServiceAccount)
		adminAuth.DELETE("/service-accounts/:id", serviceHandler.DeleteServiceAccount)
	}

	// 服务账号接口（需服务密钥认证）
	svc := r.Group("/api/service")
	svc.Use(serviceAuthMW)
	{
		svc.POST("/keys", serviceHandler.ServiceCreateKey)
		svc.GET("/keys", serviceHandler.ServiceListKeys)
	}

	return r
}
```

- [ ] **Step 3: 格式化并编译**

```bash
gofmt -w internal/router/router.go
go build ./internal/router/
```

- [ ] **Step 4: 提交**

```bash
git add internal/router/router.go
git commit -m "feat(router): add route registration for all endpoints"
```

---

### Task 19: 应用入口 main.go

**Files:**
- Create: `main.go`

**Interfaces:**
- Consumes: config, log, database, model, service, handler, middleware, router

- [ ] **Step 1: 编写 main.go**

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
	keySvc := service.NewKeyService(db)
	usageLogSvc := service.NewUsageLogService(db)
	statsSvc := service.NewStatsService(db)
	adminSvc := service.NewAdminService(db, cfg.Auth.Secret, cfg.Auth.Expiration)
	serviceAccountSvc := service.NewServiceAccountService(db)
	configSvc := service.NewConfigService(db)
	loginLogSvc := service.NewLoginLogService(db)

	// Init defaults
	if err := configSvc.InitDefaultConfigs(); err != nil {
		log.Warn("初始化默认配置失败", zap.Error(err))
	}

	adminUser := os.Getenv("ADMIN_USERNAME")
	if adminUser == "" {
		adminUser = "admin"
	}
	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminPass == "" {
		log.Fatal("ADMIN_PASSWORD 环境变量未设置")
	}
	if err := adminSvc.SeedAdmin(adminUser, adminPass); err != nil {
		log.Warn("创建初始管理员失败", zap.Error(err))
	}

	// Handlers
	keyHandler := handler.NewKeyHandler(keySvc, usageLogSvc, false)
	usageLogHandler := handler.NewUsageLogHandler(usageLogSvc)
	statsHandler := handler.NewStatsHandler(statsSvc)
	adminHandler := handler.NewAdminHandler(adminSvc, loginLogSvc)
	serviceHandler := handler.NewServiceHandler(keySvc, serviceAccountSvc)
	configHandler := handler.NewConfigHandler(configSvc)

	// Middleware
	adminAuthMW := middleware.AuthMiddleware(cfg.Auth.Secret)
	serviceAuthMW := middleware.ServiceAuthMiddleware(serviceAccountSvc)

	// Router
	r := router.SetupRouter(
		keyHandler, usageLogHandler, statsHandler,
		adminHandler, serviceHandler, configHandler,
		adminAuthMW, serviceAuthMW,
	)

	if cfg.App.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Info("服务器启动", zap.String("address", addr))
	if err := r.Run(addr); err != nil {
		log.Fatal("服务器启动失败", zap.Error(err))
	}
}
```

- [ ] **Step 2: 格式化并编译**

```bash
gofmt -w main.go
go build -o cloudkey.exe .
```

- [ ] **Step 3: 提交**

```bash
git add main.go
git commit -m "feat(core): add main.go application entry point"
```

---

## 阶段六：部署配置 + 验证 (Task 20-21)

---

### Task 20: Docker + 配置文件

**Files:**
- Create: `Dockerfile`
- Create: `docker-compose.yml`
- Create: `config.yaml.example`

- [ ] **Step 1: 编写 Dockerfile**

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cloudkey .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /app/cloudkey .
EXPOSE 8080
ENTRYPOINT ["./cloudkey"]
```

- [ ] **Step 2: 编写 docker-compose.yml**

```yaml
version: '3.8'

services:
  cloudkey:
    build: .
    container_name: cloudkey
    ports:
      - "8080:8080"
    environment:
      - CONFIG_PATH=/app/config.yaml
      - ADMIN_USERNAME=admin
      - ADMIN_PASSWORD=${ADMIN_PASSWORD:-changeme}
    volumes:
      - ./config.yaml:/app/config.yaml:ro
      - ./logs:/app/logs
    restart: unless-stopped
    depends_on:
      mysql:
        condition: service_healthy

  mysql:
    image: mysql:8.0
    container_name: cloudkey-mysql
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD:-rootpassword}
      MYSQL_DATABASE: cloudkey
      MYSQL_USER: cloudkey
      MYSQL_PASSWORD: ${MYSQL_PASSWORD:-cloudkey123}
    volumes:
      - mysql_data:/var/lib/mysql
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  mysql_data:
```

- [ ] **Step 3: 编写 config.yaml.example**

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  type: "mysql"
  host: "mysql"
  port: 3306
  user: "cloudkey"
  password: "cloudkey123"
  dbname: "cloudkey"
  sslmode: "disable"

log:
  level: "info"
  format: "json"
  output: "stdout"
  file:
    path: "logs/cloudkey.log"
    max_size: 100
    max_backups: 10
    max_age: 30
    compress: true

auth:
  secret: "your-jwt-secret-change-me"
  expiration: 24

app:
  debug: false

security:
  encryption:
    enabled: false
    algorithm: ""
    key: ""
```

- [ ] **Step 4: 提交**

```bash
git add Dockerfile docker-compose.yml config.yaml.example
git commit -m "feat(deploy): add Dockerfile, docker-compose, and config example"
```

---

### Task 21: 最终验证

**Files:**
- None（验证步骤）

- [ ] **Step 1: 运行全部测试**

```bash
go test ./... -v -count=1
```

Expected: 全部 PASS

- [ ] **Step 2: 运行 go vet**

```bash
go vet ./...
```

- [ ] **Step 3: 格式化检查**

```bash
gofmt -l .
```

Expected: 无输出（全部已格式化）

- [ ] **Step 4: 完整构建**

```bash
go build -o cloudkey.exe .
```

- [ ] **Step 5: 提交**

```bash
git add -A
git commit -m "chore: final verification — all tests pass, go vet clean"
```

---

## 规格文档覆盖检查

| 规格章节 | 覆盖状态 | 对应 Task |
|----------|----------|-----------|
| 2.1 卡密管理 | ✅ | Task 8-10 (KeyService) + Task 13 (handler) |
| 2.2 验证服务 | ✅ | Task 9 (ConsumeKey) + Task 13 (handler) |
| 2.3 使用记录 | ✅ | Task 11 (UsageLogService) + Task 14 (handler) |
| 2.4 数据统计 | ✅ | Task 11 (StatsService) + Task 14 (handler) |
| 2.5 管理后台 | ✅ | Task 15 (AdminHandler) |
| 2.6 服务账号 | ✅ | Task 11 (ServiceAccountService) + Task 16 (handler) |
| 3. API 接口设计 | ✅ | Task 18 (router.go 所有路由) |
| 4. 业务流程 | ✅ | Service 层实现 |
| 5. 数据模型 | ✅ | Task 3-6 (6 个 GORM 模型) |
| 6. 响应格式 | ✅ | Task 7 (统一 response) + Task 2 (errcode) |
| 7. 部署方案 | ✅ | Task 20 (Docker) |
| 8. 安全设计 | ✅ | Task 17 (JWT + 中间件) |
| 9. 配置项 | ✅ | Task 1 (config 补充) + Task 11 (ConfigService) |

## 已修复的问题（相比旧计划）

1. **错误码循环引用** — 错误码移至 `internal/errcode` 独立包
2. **SQLite 测试** — 明确在 Task 1 安装测试依赖
3. **Config 缺少 App 字段** — Task 1 补充 `AppSettings`
4. **ListKeysByCreatedBy 遗漏** — Task 10 已包含
5. **Service 构造函数签名不一致** — 统一为 `NewKeyService(db)` 无额外参数
6. **admin_id 类型** — middleware 设置为 `uint64`，handler 统一使用 `uint64` 类型断言
