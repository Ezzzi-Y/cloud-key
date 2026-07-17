# Task 10: Database 连接管理

**Files:**
- Create: `internal/database/database.go`

**Interfaces:**
- Produces: `Connect(cfg config.DatabaseConfig) (*gorm.DB, error)` 函数，`Close(db *gorm.DB) error` 函数

## Steps

- [ ] **Step 1: 创建目录**

```bash
mkdir -p internal/database
```

- [ ] **Step 2: 编写 database.go**

```go
package database

import (
	"CloudKey/internal/config"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Connect 建立数据库连接
// 支持 MySQL 和 SQLite 两种类型
func Connect(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Type {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
		dialector = mysql.Open(dsn)

	case "sqlite":
		// SQLite 文件名固定为 cloudkey.db
		dbPath := cfg.SQLite.Path
		if dbPath == "" {
			dbPath = "cloudkey.db"
		}
		dialector = sqlite.Open(dbPath)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// Close 关闭数据库连接
func Close(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	return sqlDB.Close()
}
```

- [ ] **Step 3: 格式化代码**

```bash
gofmt -w internal/database/database.go
```

- [ ] **Step 4: 验证编译**

```bash
go build ./internal/database/
```

- [ ] **Step 5: 提交**

```bash
git add internal/database/database.go
git commit -m "feat(database): implement Connect and Close with GORM"
```

## Global Constraints

- SQLite 文件名固定为 `cloudkey.db`
- 支持 MySQL 和 SQLite 两种类型
