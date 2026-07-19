package service

import (
	"CloudKey/internal/model"
	"testing"

	sqlite "github.com/glebarez/sqlite"
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
