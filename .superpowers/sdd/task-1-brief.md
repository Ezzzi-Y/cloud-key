### Task 1: 安装依赖 + 补充配置

**Files:**
- Modify: `go.mod`（添加依赖）
- Modify: `internal/config/config.go`（添加 AppSettings 字段）

**Interfaces:**
- Produces: go.mod 包含 Gin, JWT, TOTP, bcrypt, SQLite(测试) 依赖
- Produces: `config.AppSettings` struct, `config.AppConfig.App` field

- [ ] **Step 1: 添加 Gin 依赖**

```bash
go get github.com/gin-gonic/gin@latest
```

- [ ] **Step 2: 添加 JWT 依赖**

```bash
go get github.com/golang-jwt/jwt/v5@latest
```

- [ ] **Step 3: 添加 TOTP 依赖**

```bash
go get github.com/pquerna/otp@latest
```

- [ ] **Step 4: 添加 bcrypt 依赖**

```bash
go get golang.org/x/crypto@latest
```

- [ ] **Step 5: 添加 SQLite 测试依赖**

```bash
go get gorm.io/driver/sqlite@latest
```

- [ ] **Step 6: 运行 go mod tidy**

```bash
go mod tidy
```

- [ ] **Step 7: 补充 config.go — 添加 AppSettings**

在 `internal/config/config.go` 的 `AppConfig` struct 中添加 `App` 字段：

```go
type AppConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Log      LogConfig      `yaml:"log"`
	Auth     AuthConfig     `yaml:"auth"`
	Security SecurityConfig `yaml:"security"`
	App      AppSettings    `yaml:"app"`
}

// AppSettings 应用级别设置
type AppSettings struct {
	Debug bool `yaml:"debug"`
}
```

- [ ] **Step 8: 验证编译**

```bash
go build ./...
```

- [ ] **Step 9: 提交**

```bash
git add go.mod go.sum internal/config/config.go
git commit -m "feat(deps): add Gin, JWT, TOTP, bcrypt, SQLite(test) deps; add AppSettings config"
```

---

