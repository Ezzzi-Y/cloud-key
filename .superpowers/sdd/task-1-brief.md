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

