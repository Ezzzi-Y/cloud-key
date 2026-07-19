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

