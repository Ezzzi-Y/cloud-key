package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_JSON(t *testing.T) {
	// 创建临时 JSON 配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configJSON := `{
		"server": {
			"port": 8080,
			"host": "localhost"
		},
		"database": {
			"type": "sqlite",
			"sqlite": {
				"path": "cloudkey.db"
			}
		},
		"log": {
			"level": "info",
			"format": "json"
		},
		"security": {
			"encryption": {
				"enabled": false
			}
		}
	}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatal(err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}

	if config.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", config.Server.Port)
	}
	if config.Database.Type != "sqlite" {
		t.Errorf("expected db type sqlite, got %s", config.Database.Type)
	}
	if config.Security.Encryption.Enabled != false {
		t.Error("expected encryption.enabled to be false")
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	// 创建最小配置，验证默认值
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configJSON := `{
		"server": {"port": 3000}
	}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatal(err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}

	// encryption.enabled 应为默认 false
	if config.Security.Encryption.Enabled != false {
		t.Error("expected encryption.enabled default to be false")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/config.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
