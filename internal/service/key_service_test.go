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
