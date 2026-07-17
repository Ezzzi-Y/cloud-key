# Task 5: Config 测试

**Files:**
- Create: `internal/config/config_test.go`

**Interfaces:**
- Tests: `LoadConfig` 函数的各种场景

## Steps

- [ ] **Step 1: 编写测试文件**

```go
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
```

- [ ] **Step 2: 运行测试**

```bash
go test ./internal/config/ -v
```

Expected: 全部 PASS

- [ ] **Step 3: 提交**

```bash
git add internal/config/config_test.go
git commit -m "test(config): add unit tests for LoadConfig"
```

## Global Constraints

- 测试使用 `t.TempDir()` 生成临时目录
