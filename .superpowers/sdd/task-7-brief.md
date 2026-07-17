# Task 7: Log 创建入口

**Files:**
- Create: `internal/log/logger.go`

**Interfaces:**
- Produces: `InitLogger(cfg config.LogConfig) error` 函数，`Sync() error` 函数，`Debug/Info/Warn/Error` 函数

## Steps

- [ ] **Step 1: 创建目录**

```bash
mkdir -p internal/log
```

- [ ] **Step 2: 编写 logger.go**

```go
package log

import (
	"CloudKey/internal/config"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.Logger

// InitLogger 初始化日志系统
func InitLogger(cfg config.LogConfig) error {
	// 解析日志级别
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// 配置编码器
	var encoderConfig zapcore.EncoderConfig
	if cfg.Format == "json" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 配置输出
	var writeSyncer zapcore.WriteSyncer
	if cfg.Output == "file" && cfg.File.Path != "" {
		lumberjackLogger := &lumberjack.Logger{
			Filename:   cfg.File.Path,
			MaxSize:    cfg.File.MaxSize,
			MaxBackups: cfg.File.MaxBackups,
			MaxAge:     cfg.File.MaxAge,
			Compress:   cfg.File.Compress,
		}
		writeSyncer = zapcore.AddSync(lumberjackLogger)
	} else {
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// 创建核心
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewCore(encoder, writeSyncer, level)
	logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

// Sync 刷新缓冲区
func Sync() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}

// Debug 记录 DEBUG 级别日志
func Debug(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Debug(msg, fields...)
	}
}

// Info 记录 INFO 级别日志
func Info(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Info(msg, fields...)
	}
}

// Warn 记录 WARN 级别日志
func Warn(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Warn(msg, fields...)
	}
}

// Error 记录 ERROR 级别日志
func Error(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Error(msg, fields...)
	}
}
```

- [ ] **Step 3: 格式化代码**

```bash
gofmt -w internal/log/logger.go
```

- [ ] **Step 4: 验证编译**

```bash
go build ./internal/log/
```

- [ ] **Step 5: 提交**

```bash
git add internal/log/logger.go
git commit -m "feat(log): implement InitLogger with Zap and Lumberjack"
```

## Global Constraints

- 日志级别通过 `cfg.Level` 配置
- 支持 console 和 file 两种输出
- 支持 json 和 console 两种格式
- 文件输出使用 Lumberjack 进行轮转
