# Task 8: Log 测试

**Files:**
- Create: `internal/log/logger_test.go`

**Interfaces:**
- Tests: `InitLogger` 函数的各种场景

## Steps

- [ ] **Step 1: 编写测试文件**

```go
package log

import (
	"CloudKey/internal/config"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
)

func TestInitLogger_Console(t *testing.T) {
	cfg := config.LogConfig{
		Level:  "info",
		Format: "console",
		Output: "stdout",
	}

	err := InitLogger(cfg)
	if err != nil {
		t.Fatal(err)
	}

	Info("test message", zap.String("key", "value"))
	Sync()
}

func TestInitLogger_JSON(t *testing.T) {
	cfg := config.LogConfig{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	}

	err := InitLogger(cfg)
	if err != nil {
		t.Fatal(err)
	}

	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")
	Sync()
}

func TestInitLogger_File(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := config.LogConfig{
		Level:  "info",
		Format: "json",
		Output: "file",
		File: config.FileConfig{
			Path:       logPath,
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   false,
		},
	}

	err := InitLogger(cfg)
	if err != nil {
		t.Fatal(err)
	}

	Info("file log test")
	Sync()

	// 验证日志文件已创建
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("expected log file to be created")
	}
}

func TestInitLogger_InvalidLevel(t *testing.T) {
	cfg := config.LogConfig{
		Level:  "invalid",
		Format: "console",
		Output: "stdout",
	}

	// 应该使用默认级别，不返回错误
	err := InitLogger(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
```

- [ ] **Step 2: 运行测试**

```bash
go test ./internal/log/ -v
```

Expected: 全部 PASS

- [ ] **Step 3: 提交**

```bash
git add internal/log/logger_test.go
git commit -m "test(log): add unit tests for InitLogger"
```

## Global Constraints

- 测试使用 `t.TempDir()` 生成临时目录
