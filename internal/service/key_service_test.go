package service

import (
	"testing"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tenant_id INTEGER NOT NULL DEFAULT 1,
		alias TEXT NOT NULL,
		key_hash TEXT NOT NULL UNIQUE,
		key_prefix TEXT NOT NULL,
		key_suffix TEXT NOT NULL,
		remaining_amount INTEGER NOT NULL,
		version INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'unused',
		created_by TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		used_at DATETIME,
		expire_at DATETIME,
		max_usage INTEGER
	)`)
	return db
}

const testTenantID uint64 = 1

func TestCreateKey(t *testing.T) {
	db := setupTestDB(t)
	svc := NewKeyService(db, nil)

	result, err := svc.CreateKey(CreateKeyRequest{
		Alias: "test-key", RemainingAmount: 100, CreatedBy: "admin",
	}, testTenantID, "sk-", 32, 4)
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

func TestFindByRawKeyTenant(t *testing.T) {
	db := setupTestDB(t)
	svc := NewKeyService(db, nil)

	result, _ := svc.CreateKey(CreateKeyRequest{
		Alias: "find-test", RemainingAmount: 50, CreatedBy: "admin",
	}, testTenantID, "sk-", 32, 4)

	found, err := svc.FindByRawKeyTenant(result.RawKey, testTenantID)
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

func TestFindByRawKeyTenant_WrongTenant(t *testing.T) {
	db := setupTestDB(t)
	svc := NewKeyService(db, nil)

	result, _ := svc.CreateKey(CreateKeyRequest{
		Alias: "isolation-test", RemainingAmount: 50, CreatedBy: "admin",
	}, testTenantID, "sk-", 32, 4)

	found, err := svc.FindByRawKeyTenant(result.RawKey, 999)
	if err != nil {
		t.Fatal(err)
	}
	if found != nil {
		t.Error("expected nil for key belonging to different tenant")
	}
}

func TestFindByRawKeyTenant_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewKeyService(db, nil)

	found, err := svc.FindByRawKeyTenant("sk-nonexistent-key", testTenantID)
	if err != nil {
		t.Fatal(err)
	}
	if found != nil {
		t.Error("expected nil for nonexistent key")
	}
}

func TestGetKeyStatusByTenant(t *testing.T) {
	db := setupTestDB(t)
	svc := NewKeyService(db, nil)

	result, _ := svc.CreateKey(CreateKeyRequest{
		Alias: "status-test", RemainingAmount: 10, CreatedBy: "admin",
	}, testTenantID, "sk-", 32, 4)

	status, err := svc.GetKeyStatusByTenant(result.RawKey, testTenantID)
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
	svc := NewKeyService(db, nil)

	for i := 0; i < 5; i++ {
		svc.CreateKey(CreateKeyRequest{
			Alias: "key-" + string(rune('A'+i)), RemainingAmount: 10, CreatedBy: "admin",
		}, testTenantID, "sk-", 32, 4)
	}

	keys, total, err := svc.ListKeys(KeyListQuery{Page: 1, PageSize: 3}, testTenantID)
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
	svc := NewKeyService(db, nil)

	result, _ := svc.CreateKey(CreateKeyRequest{
		Alias: "toggle-test", RemainingAmount: 10, CreatedBy: "admin",
	}, testTenantID, "sk-", 32, 4)

	if err := svc.DisableKey(result.Key.ID, testTenantID); err != nil {
		t.Fatal(err)
	}

	key, _ := svc.GetKeyDetail(result.Key.ID, testTenantID)
	if key.Status != "disabled" {
		t.Errorf("expected disabled, got %s", key.Status)
	}

	if err := svc.EnableKey(result.Key.ID, testTenantID); err != nil {
		t.Fatal(err)
	}

	key, _ = svc.GetKeyDetail(result.Key.ID, testTenantID)
	if key.Status != "active" {
		t.Errorf("expected active, got %s", key.Status)
	}
}

func TestConsumeKeyByTenant(t *testing.T) {
	db := setupTestDB(t)
	svc := NewKeyService(db, nil)

	result, _ := svc.CreateKey(CreateKeyRequest{
		Alias: "consume-test", RemainingAmount: 100, CreatedBy: "admin",
	}, testTenantID, "sk-", 32, 4)

	consumeResult, code, err := svc.ConsumeKeyByTenant(result.RawKey, 30, testTenantID)
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Fatalf("expected code 0, got %d", code)
	}
	if consumeResult.RemainingAmount != 70 {
		t.Errorf("expected 70 remaining, got %d", consumeResult.RemainingAmount)
	}
}

func TestConsumeKeyByTenant_WrongTenant(t *testing.T) {
	db := setupTestDB(t)
	svc := NewKeyService(db, nil)

	result, _ := svc.CreateKey(CreateKeyRequest{
		Alias: "isolation-consume", RemainingAmount: 100, CreatedBy: "admin",
	}, testTenantID, "sk-", 32, 4)

	_, code, err := svc.ConsumeKeyByTenant(result.RawKey, 10, 999)
	if err != nil {
		t.Fatal(err)
	}
	if code == 0 {
		t.Error("expected non-zero code for wrong tenant, got 0 (key consumed across tenants!)")
	}
}
