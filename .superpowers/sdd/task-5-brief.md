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

