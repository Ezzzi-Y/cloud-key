# Task 3: Config LoadConfig 函数

**Files:**
- Modify: `internal/config/config.go`

**Interfaces:**
- Produces: `LoadConfig(path string) (*AppConfig, error)` 函数

## Steps

- [ ] **Step 1: 添加 LoadConfig 函数**

在 `config.go` 末尾添加：

```go
import (
	"github.com/spf13/viper"
)

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
```

- [ ] **Step 2: 格式化代码**

```bash
gofmt -w internal/config/config.go
```

- [ ] **Step 3: 验证编译**

```bash
go build ./internal/config/
```

- [ ] **Step 4: 提交**

```bash
git add internal/config/config.go
git commit -m "feat(config): implement LoadConfig with Viper"
```

## Global Constraints

- 配置文件字段名与结构体 YAML tag 必须一致
