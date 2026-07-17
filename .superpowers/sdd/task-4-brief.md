# Task 4: Config SetDefaults 和验证

**Files:**
- Modify: `internal/config/config.go`

**Interfaces:**
- Produces: `setDefaults(v *viper.Viper)` 函数，集成到 LoadConfig

## Steps

- [x] **Step 1: 添加 setDefaults 函数**

在 `config.go` 中添加：

```go
// setDefaults 设置默认值
func setDefaults(v *viper.Viper) {
	// encryption.enabled 默认 false
	v.SetDefault("security.encryption.enabled", false)

	// app.debug 默认 false
	v.SetDefault("app.debug", false)
}
```

- [x] **Step 2: 在 LoadConfig 中调用 setDefaults**

修改 LoadConfig 函数，在 `v.SetConfigFile(path)` 之后添加：

```go
	setDefaults(v)
```

- [x] **Step 3: 格式化代码**

```bash
gofmt -w internal/config/config.go
```

- [x] **Step 4: 验证编译**

```bash
go build ./internal/config/
```

- [x] **Step 5: 提交**

```bash
git add internal/config/config.go
git commit -m "feat(config): add defaults for encryption.enabled and app.debug"
```

## Global Constraints

- `encryption.enabled` 默认 `false`
- `app.debug` 默认 `false`
