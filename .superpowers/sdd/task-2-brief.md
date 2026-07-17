# Task 2: Config 定义

**Files:**
- Create: `internal/config/config.go`

**Interfaces:**
- Produces: `AppConfig` struct, `LoadConfig(path string) (*AppConfig, error)` 函数

## Steps

- [ ] **Step 1: 创建目录结构**

```bash
mkdir -p internal/config
```

- [ ] **Step 2: 编写 config.go**

```go
package config

// AppConfig 定义应用程序配置结构
// 字段名与 YAML tag 保持一致
type AppConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Log      LogConfig      `yaml:"log"`
	Auth     AuthConfig     `yaml:"auth"`
	Security SecurityConfig `yaml:"security"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
	SQLite   SQLiteConfig `yaml:"sqlite"`
}

// SQLiteConfig SQLite 配置
type SQLiteConfig struct {
	Path string `yaml:"path"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
	File   FileConfig `yaml:"file"`
}

// FileConfig 日志文件配置
type FileConfig struct {
	Path       string `yaml:"path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Secret     string `yaml:"secret"`
	Expiration int    `yaml:"expiration"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	Encryption EncryptionConfig `yaml:"encryption"`
}

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Algorithm string `yaml:"algorithm"`
	Key       string `yaml:"key"`
}
```

- [ ] **Step 3: 格式化代码**

```bash
gofmt -w internal/config/config.go
```

- [ ] **Step 4: 验证编译**

```bash
go build ./internal/config/
```

Expected: 无错误

- [ ] **Step 5: 提交**

```bash
git add internal/config/config.go
git commit -m "feat(config): define AppConfig struct with YAML tags"
```

## Global Constraints

- 配置文件字段名与结构体 YAML tag 必须一致
- `encryption.enabled` 默认 `false`
- `app.debug` 默认 `false`
