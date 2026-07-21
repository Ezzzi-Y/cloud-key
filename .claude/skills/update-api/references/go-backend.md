# Go 后端开发流程

新增或修改 API 端点时，按以下层次顺序操作。每步完成后运行 `go build ./...` 验证编译。

## 1. Model (`internal/model/`)

定义 GORM 结构体，实现 `TableName()` 方法：

```go
type MyModel struct {
    ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
    TenantID  uint64    `gorm:"type:bigint;index;not null" json:"tenant_id"`
    // ... 其他字段
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (MyModel) TableName() string { return "my_models" }
```

新模型须在 `internal/model/migrate.go` 的 `AutoMigrate()` 中注册。

## 2. Errcode (`internal/errcode/errcode.go`)

按类别分段新增错误码常量，在 `codeMessages` map 中添加中文消息。

## 3. Service (`internal/service/`)

- 注入 `*gorm.DB`，不依赖 Gin
- 并发安全操作使用乐观锁：`version` 字段 + 重试循环（参考 `KeyService.ConsumeKey`）
- 返回业务结果 struct

## 4. Handler (`internal/handler/`)

**JWT 认证（租户管理员）**：
```go
func (h *TenantKeyHandler) MyEndpoint(c *gin.Context) {
    tenantID := getTenantID(c)
    // 参数绑定、校验、调用 service、返回响应
    Success(c, result)
}
```

**X-Service-Key 认证（服务账号）**：
```go
func (h *TenantServiceAccountHandler) ServiceMyEndpoint(c *gin.Context) {
    tenantID, ok := getServiceTenantID(c)
    if !ok { return }
    // ...
}
```

响应辅助函数：`Success()`, `BadRequest()`, `InternalError()`, `NotFound()`, `Unauthorized()`, `SuccessPaginated()`

添加 swaggo 注释（参考已有 handler 格式）。

## 5. Router (`internal/router/router.go`)

在 `SetupRouter()` 中注册路由：

- 租户路由：`middleware.AuthMiddleware` + `middleware.RequireTenantAdmin`
- 写操作加 `middleware.TenantBusinessGuard(db)`
- 服务账号路由：`middleware.ServiceAuthMiddleware`

新增 handler 参数时更新 `SetupRouter()` 函数签名。

## 6. Main (`main.go`)

- 实例化新 Service 和 Handler
- 注入到 `router.SetupRouter()`

## 7. swag 重生成

```bash
swag init
```

更新 `docs/` 目录下的 Swagger spec。
