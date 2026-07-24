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

	zap.L().Info("test message", zap.String("key", "value"))
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

	zap.L().Debug("debug message")
	zap.L().Info("info message")
	zap.L().Warn("warn message")
	zap.L().Error("error message")
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

	zap.L().Info("file log test")
	Sync()

	// 关闭日志文件句柄，确保 TempDir 清理时不报错
	defer Close()

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
