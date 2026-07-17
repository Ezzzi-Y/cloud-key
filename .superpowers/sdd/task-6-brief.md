# Task 6: 安装 Log 依赖

**Files:**
- Modify: `go.mod`

**Interfaces:**
- Produces: `go.mod` 包含 Zap 和 Lumberjack 依赖

## Steps

- [ ] **Step 1: 添加 Zap 依赖**

```bash
go get go.uber.org/zap@v1.27.0
```

- [ ] **Step 2: 添加 Lumberjack 依赖**

```bash
go get gopkg.in/natefinish/lumberjack.v2@v2.2.1
```

- [ ] **Step 3: 运行 go mod tidy**

```bash
go mod tidy
```

- [ ] **Step 4: 验证 go.mod**

```bash
cat go.mod
```

Expected: 包含 zap 和 lumberjack 依赖

- [ ] **Step 5: 提交**

```bash
git add go.mod go.sum
git commit -m "feat(deps): add Zap and Lumberjack for logging"
```
