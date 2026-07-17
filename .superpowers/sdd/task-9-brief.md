# Task 9: 安装 Database 依赖

**Files:**
- Modify: `go.mod`

**Interfaces:**
- Produces: `go.mod` 包含 GORM、MySQL Driver、SQLite Driver 依赖

## Steps

- [ ] **Step 1: 添加 GORM 依赖**

```bash
go get gorm.io/gorm@v1.30.0
```

- [ ] **Step 2: 添加 MySQL Driver 依赖**

```bash
go get gorm.io/driver/mysql@v1.6.0
```

- [ ] **Step 3: 添加 SQLite Driver 依赖**

```bash
go get gorm.io/driver/sqlite@v1.6.0
```

- [ ] **Step 4: 运行 go mod tidy**

```bash
go mod tidy
```

- [ ] **Step 5: 验证 go.mod**

```bash
cat go.mod
```

Expected: 包含 gorm、mysql、sqlite 依赖

- [ ] **Step 6: 提交**

```bash
git add go.mod go.sum
git commit -m "feat(deps): add GORM with MySQL and SQLite drivers"
```
