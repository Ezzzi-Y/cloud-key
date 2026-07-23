package config

import "github.com/spf13/viper"

// AppConfig 定义应用程序配置结构
// 字段名与 YAML tag 保持一致
type AppConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Log      LogConfig      `yaml:"log"`
	Auth     AuthConfig     `yaml:"auth"`
	Security SecurityConfig `yaml:"security"`
	App      AppSettings    `yaml:"app"`
	MQ       MQConfig       `yaml:"rabbitmq" mapstructure:"rabbitmq"`
}

// AppSettings 应用级别设置
type AppSettings struct {
	Debug bool `yaml:"debug"`
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
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
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
	Secret             string `yaml:"secret" mapstructure:"secret"`
	Expiration         int    `yaml:"expiration" mapstructure:"expiration"`
	SuperAdminUsername string `yaml:"super_admin_username" mapstructure:"super_admin_username"`
	SuperAdminPassword string `yaml:"super_admin_password" mapstructure:"super_admin_password"`
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

// MQConfig RabbitMQ 配置
type MQConfig struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	Username string `yaml:"username" mapstructure:"username"`
	Password string `yaml:"password" mapstructure:"password"`
}

// setDefaults 设置默认值
func setDefaults(v *viper.Viper) {
	// encryption.enabled 默认 false
	v.SetDefault("security.encryption.enabled", false)

	// app.debug 默认 false
	v.SetDefault("app.debug", false)
}

// LoadConfig 加载配置文件
// path: 配置文件路径（支持 JSON、YAML、TOML 格式）
func LoadConfig(path string) (*AppConfig, error) {
	v := viper.New()
	v.SetConfigFile(path)
	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var config AppConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
