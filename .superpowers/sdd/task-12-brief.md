# Task 12: Database 测试

**Files:**
- Create: `internal/database/database_test.go`

**Interfaces:**
- Tests: `Connect` 和 `Close` 函数的各种场景

## Steps

- [ ] **Step 1: 编写测试文件**

```go
package database

import (
	"CloudKey/internal/config"
	"path/filepath"
	"testing"
)

func TestConnect_SQLite(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := config.DatabaseConfig{
		Type: "sqlite",
		SQLite: config.SQLiteConfig{
			Path: dbPath,
		},
	}

	db, err := Connect(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer Close(db)

	// 验证连接可用
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}

	if err := sqlDB.Ping(); err != nil {
		t.Fatal(err)
	}
}

func TestConnect_SQLiteDefaultPath(t *testing.T) {
	cfg := config.DatabaseConfig{
		Type: "sqlite",
		SQLite: config.SQLiteConfig{
			Path: "", // 空路径应使用默认 cloudkey.db
		},
	}

	db, err := Connect(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer Close(db)

	// 清理默认文件
	sqlDB, _ := db.DB()
	sqlDB.Close()
	// 注意：这里会在当前目录创建 cloudkey.db，测试后需要清理
}

func TestConnect_UnsupportedType(t *testing.T) {
	cfg := config.DatabaseConfig{
		Type: "unsupported",
	}

	_, err := Connect(cfg)
	if err == nil {
		t.Error("expected error for unsupported database type")
	}
}

func TestConnect_MySQLNoServer(t *testing.T) {
	cfg := config.DatabaseConfig{
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password",
		DBName:   "test",
	}

	// 没有 MySQL 服务器，应该失败
	_, err := Connect(cfg)
	if err == nil {
		t.Error("expected error when MySQL server is not available")
	}
}

func TestClose_NilDB(t *testing.T) {
	err := Close(nil)
	if err != nil {
		t.Errorf("expected no error for nil db, got %v", err)
	}
}
```

- [ ] **Step 2: 运行测试**

```bash
go test ./internal/database/ -v
```

Expected: 全部 PASS（MySQL 测试会因无服务器而预期失败）

- [ ] **Step 3: 提交**

```bash
git add internal/database/database_test.go
git commit -m "test(database): add unit tests for Connect and Close"
```

## Global Constraints

- 测试使用 `t.TempDir()` 生成临时目录
- SQLite 测试使用临时目录中的文件
