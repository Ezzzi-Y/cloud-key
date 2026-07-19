### Task 9: 重构 router.go + main.go（全部组装）

**Files:**
- Rewrite: `internal/router/router.go`
- Rewrite: `main.go`

- [ ] **Step 1: 删除旧的 handler/admin/service 文件**

```bash
rm internal/model/admin.go
rm internal/service/admin_service.go
rm internal/handler/admin_handler.go
```

- [ ] **Step 2: 重写 `internal/router/router.go`**

```go
package router

import (
	"CloudKey/internal/handler"
	"CloudKey/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(
	authHandler *handler.AuthHandler,
	superHandler *handler.SuperHandler,
	tenantKeyHandler *handler.TenantKeyHandler,
	tenantSAHandler *handler.TenantServiceAccountHandler,
	tenantStatsHandler *handler.TenantStatsHandler,
	tenantUsageLogHandler *handler.TenantUsageLogHandler,
	jwtSecret string,
	db *gorm.DB,
	saSvc *service.ServiceAccountService,
) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	// 静态文件
	r.StaticFile("/", "web/admin.html")
	r.NoRoute(func(c *gin.Context) {
		if c.Request.URL.Path != "/" {
			if !strings.HasPrefix(c.Request.URL.Path, "/api") {
				http.ServeFile(c.Writer, c.Request, "web/admin.html")
				return
			}
		}
	})

	api := r.Group("/api")

	// ========== 认证（公开） ==========
	{
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/verify-2fa", authHandler.Verify2FA)
		api.POST("/auth/totp/setup-init", authHandler.SetupTOTPPublic)
		api.POST("/auth/totp/confirm-init", authHandler.ConfirmTOTPPublic)
	}

	// ========== 公共 API ==========
	{
		api.GET("/key/status", tenantKeyHandler.Status)   // 无 auth
		api.POST("/key/consume", tenantKeyHandler.Consume) // 无 auth
	}

	// ========== 系统管理员 ==========
	super := api.Group("/super")
	super.Use(middleware.AuthMiddleware(jwtSecret))
	super.Use(middleware.RequireSuperAdmin())
	{
		super.GET("/tenants", superHandler.ListTenants)
		super.POST("/tenants", superHandler.CreateTenant)
		super.GET("/tenants/:id", superHandler.GetTenant)
		super.PATCH("/tenants/:id", superHandler.UpdateTenant)
		super.PATCH("/tenants/:id/reset-password", superHandler.ResetPassword)

		super.GET("/configs", superHandler.GetConfigs)
		super.PUT("/configs", superHandler.UpdateConfigs)

		super.GET("/profile", authHandler.Profile)
		super.PUT("/password", authHandler.ChangePassword)
		super.POST("/totp/setup", authHandler.SetupTOTP)
		super.POST("/totp/confirm", authHandler.ConfirmTOTP)
		super.GET("/login-logs", superHandler.LoginLogs)
	}

	// ========== 租户管理员 ==========
	tenant := api.Group("/tenant")
	tenant.Use(middleware.AuthMiddleware(jwtSecret))
	tenant.Use(middleware.RequireTenantAdmin(db))
	{
		// Key 管理（业务操作加 BusinessGuard）
		tenantKeys := tenant.Group("/keys")
		tenantKeys.POST("", middleware.TenantBusinessGuard(db), tenantKeyHandler.CreateKey)
		tenantKeys.GET("", tenantKeyHandler.ListKeys)
		tenantKeys.GET("/export", tenantKeyHandler.ExportKeys)
		tenantKeys.GET("/:id", tenantKeyHandler.GetKey)
		tenantKeys.PATCH("/:id", middleware.TenantBusinessGuard(db), tenantKeyHandler.UpdateKey)
		tenantKeys.PATCH("/:id/disable", middleware.TenantBusinessGuard(db), tenantKeyHandler.DisableKey)
		tenantKeys.PATCH("/:id/enable", middleware.TenantBusinessGuard(db), tenantKeyHandler.EnableKey)
		tenantKeys.DELETE("/:id", middleware.TenantBusinessGuard(db), tenantKeyHandler.DeleteKey)

		// 服务账号（全部加 BusinessGuard）
		tenantSA := tenant.Group("/service-accounts")
		tenantSA.Use(middleware.TenantBusinessGuard(db))
		{
			tenantSA.GET("", tenantSAHandler.ListServiceAccounts)
			tenantSA.POST("", tenantSAHandler.CreateServiceAccount)
			tenantSA.PATCH("/:id/toggle", tenantSAHandler.ToggleServiceAccount)
			tenantSA.DELETE("/:id", tenantSAHandler.DeleteServiceAccount)
		}

		// 统计（不 guard，expired 可查看）
		tenantStats := tenant.Group("/stats")
		{
			tenantStats.GET("/dashboard", tenantStatsHandler.Dashboard)
			tenantStats.GET("/overview", tenantStatsHandler.Overview)
			tenantStats.GET("/trends", tenantStatsHandler.Trends)
			tenantStats.GET("/top-keys", tenantStatsHandler.TopKeys)
			tenantStats.GET("/top-ips", tenantStatsHandler.TopIPs)
		}

		// 使用日志（不 guard）
		tenantLogs := tenant.Group("/usage-logs")
		{
			tenantLogs.GET("", tenantUsageLogHandler.ListLogs)
			tenantLogs.GET("/export", tenantUsageLogHandler.ExportLogs)
		}

		// 个人设置
		tenant.GET("/profile", authHandler.Profile)
		tenant.PUT("/password", authHandler.ChangePassword)
		tenant.POST("/totp/setup", authHandler.SetupTOTP)
		tenant.POST("/totp/confirm", authHandler.ConfirmTOTP)
		tenant.GET("/login-logs", tenantUsageLogHandler.LoginLogs) // 或 superHandler
	}

	// ========== 服务账号 API ==========
	serviceAPI := api.Group("/service")
	serviceAPI.Use(middleware.ServiceAuthMiddleware(saSvc, db))
	{
		serviceAPI.POST("/keys", tenantSAHandler.ServiceCreateKey)
		serviceAPI.GET("/keys", tenantSAHandler.ServiceListKeys)
	}

	return r
}
```

- [ ] **Step 3: 重写 `main.go`**

```go
package main

import (
	"CloudKey/internal/config"
	"CloudKey/internal/database"
	"CloudKey/internal/handler"
	"CloudKey/internal/log"
	"CloudKey/internal/middleware"
	"CloudKey/internal/model"
	"CloudKey/internal/router"
	"CloudKey/internal/service"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := log.InitLogger(cfg.Log); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()
	defer log.Close()

	log.Info("CloudKey 启动中...")

	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatal("数据库连接失败", zap.Error(err))
	}
	defer database.Close(db)

	if err := model.AutoMigrate(db); err != nil {
		log.Fatal("数据库迁移失败", zap.Error(err))
	}
	log.Info("数据库迁移完成")

	// Services
	authSvc := service.NewAuthService(db, cfg.Auth.Secret, cfg.Auth.Expiration)
	keySvc := service.NewKeyService(db)
	usageLogSvc := service.NewUsageLogService(db)
	statsSvc := service.NewStatsService(db)
	serviceAccountSvc := service.NewServiceAccountService(db)
	configSvc := service.NewConfigService(db)
	loginLogSvc := service.NewLoginLogService(db)
	tenantSvc := service.NewTenantService(db)

	// Init defaults
	if err := configSvc.InitDefaultConfigs(); err != nil {
		log.Warn("初始化默认配置失败", zap.Error(err))
	}

	// Seed super admin
	superAdminUser := os.Getenv("SUPER_ADMIN_USERNAME")
	if superAdminUser == "" {
		superAdminUser = cfg.Auth.SuperAdminUsername
	}
	if superAdminUser == "" {
		superAdminUser = "admin"
	}
	superAdminPass := os.Getenv("SUPER_ADMIN_PASSWORD")
	if superAdminPass == "" {
		superAdminPass = cfg.Auth.SuperAdminPassword
	}
	if superAdminPass == "" {
		log.Fatal("请设置 SUPER_ADMIN_PASSWORD 环境变量或在 config.yaml 的 auth.super_admin_password 中配置")
	}
	if err := authSvc.SeedSuperAdmin(superAdminUser, superAdminPass); err != nil {
		log.Warn("创建超级管理员失败", zap.Error(err))
	}

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc, loginLogSvc)
	superHandler := handler.NewSuperHandler(tenantSvc, configSvc, statsSvc, loginLogSvc)
	tenantKeyHandler := handler.NewTenantKeyHandler(keySvc, usageLogSvc, false)
	tenantSAHandler := handler.NewTenantServiceAccountHandler(keySvc, serviceAccountSvc)
	tenantStatsHandler := handler.NewTenantStatsHandler(statsSvc)
	tenantUsageLogHandler := handler.NewTenantUsageLogHandler(usageLogSvc, loginLogSvc)

	// Gin mode
	if cfg.App.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Router
	r := router.SetupRouter(
		authHandler, superHandler,
		tenantKeyHandler, tenantSAHandler, tenantStatsHandler, tenantUsageLogHandler,
		cfg.Auth.Secret, db, serviceAccountSvc,
	)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Info("服务器启动", zap.String("address", addr))
	if err := r.Run(addr); err != nil {
		log.Fatal("服务器启动失败", zap.Error(err))
	}
}
```

- [ ] **Step 4: 更新 `config/config.go` — AuthConfig 加 SuperAdmin 字段**

```go
type AuthConfig struct {
	Secret             string `yaml:"secret" mapstructure:"secret"`
	Expiration         int    `yaml:"expiration" mapstructure:"expiration"`
	SuperAdminUsername string `yaml:"super_admin_username" mapstructure:"super_admin_username"`
	SuperAdminPassword string `yaml:"super_admin_password" mapstructure:"super_admin_password"`
}
```

- [ ] **Step 5: 更新 `config.yaml.example`**

```yaml
auth:
  secret: "change-me-to-a-random-string"
  expiration: 24
  super_admin_username: "admin"
  super_admin_password: "change-me"
```

- [ ] **Step 6: 验证编译**

Run: `cd D:/MyGoProject/CloudKey && go build ./...`
Expected: 编译通过，无错误

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: wire up new router and main.go for multi-tenant SaaS"
```

---

