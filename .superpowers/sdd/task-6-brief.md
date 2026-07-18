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

