package database

import (
	"CloudKey/internal/config"
	"testing"
)

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
