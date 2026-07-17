# Task 11: Database 连接池配置

**Files:**
- Modify: `internal/database/database.go`

**Interfaces:**
- Produces: 集成 `sql.DB` 连接池设置到 `Connect` 函数

## Steps

- [ ] **Step 1: 添加连接池配置**

修改 `Connect` 函数，在 `return db, nil` 之前添加：

```go
	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 连接池设置
	sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生命周期
```

- [ ] **Step 2: 添加 time 导入**

在 import 块中添加：

```go
	"time"
```

- [ ] **Step 3: 格式化代码**

```bash
gofmt -w internal/database/database.go
```

- [ ] **Step 4: 验证编译**

```bash
go build ./internal/database/
```

- [ ] **Step 5: 提交**

```bash
git add internal/database/database.go
git commit -m "feat(database): add connection pool configuration"
```

## Global Constraints

- 最大空闲连接数: 10
- 最大打开连接数: 100
- 连接最大生命周期: 1 hour
