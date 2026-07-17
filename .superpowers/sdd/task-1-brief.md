# Task 1: 创建 Go 模块并安装 Config 依赖

**Files:**
- Modify: `go.mod`

**Interfaces:**
- Produces: `go.mod` 包含 Viper 依赖

## Steps

- [ ] **Step 1: 添加 Viper 依赖**

```bash
cd D:\MyGoProject\CloudKey
go get github.com/spf13/viper@v1.20.1
```

- [ ] **Step 2: 验证 go.mod 更新**

```bash
cat go.mod
```

Expected: 包含 `require github.com/spf13/viper v1.20.1`

- [ ] **Step 3: 运行 go mod tidy**

```bash
go mod tidy
```

- [ ] **Step 4: 提交**

```bash
git add go.mod go.sum
git commit -m "feat(deps): add Viper for config management"
```
