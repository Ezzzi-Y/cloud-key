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

