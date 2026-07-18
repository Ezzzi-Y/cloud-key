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

