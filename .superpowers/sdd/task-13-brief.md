# Task 13: 运行全部测试并验证

**Files:**
- None (验证步骤)

**Interfaces:**
- None

## Steps

- [ ] **Step 1: 运行全部测试**

```bash
go test ./internal/... -v
```

Expected: 所有三个模块的测试全部 PASS

- [ ] **Step 2: 运行 go vet**

```bash
go vet ./internal/...
```

Expected: 无错误

- [ ] **Step 3: 格式化检查**

```bash
gofmt -l ./internal/
```

Expected: 无输出（所有文件已格式化）

- [ ] **Step 4: 构建验证**

```bash
go build ./...
```

Expected: 无错误

- [ ] **Step 5: 最终提交**

```bash
git add -A
git commit -m "chore: verify all modules build and pass tests"
```

## Global Constraints

- 所有测试必须通过
- go vet 无错误
- 所有文件已格式化
