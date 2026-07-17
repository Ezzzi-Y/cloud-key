package config

import "github.com/spf13/viper"

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
	Type     string       `yaml:"type"`
	Host     string       `yaml:"host"`
	Port     int          `yaml:"port"`
	User     string       `yaml:"user"`
	Password string       `yaml:"password"`
	DBName   string       `yaml:"dbname"`
	SSLMode  string       `yaml:"sslmode"`
	SQLite   SQLiteConfig `yaml:"sqlite"`
}

// SQLiteConfig SQLite 配置
type SQLiteConfig struct {
	Path string `yaml:"path"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string     `yaml:"level"`
	Format string     `yaml:"format"`
	Output string     `yaml:"output"`
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

// LoadConfig 加载配置文件
// path: 配置文件路径（支持 JSON、YAML、TOML 格式）
func LoadConfig(path string) (*AppConfig, error) {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var config AppConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
