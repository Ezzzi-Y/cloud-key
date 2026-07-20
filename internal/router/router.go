package router

import (
	_ "CloudKey/docs"
	"CloudKey/internal/handler"
	"CloudKey/internal/middleware"
	"CloudKey/internal/service"
	"CloudKey/internal/web"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	rdb *redis.Client,
	debug bool,
) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	// Swagger UI (仅 debug 模式启用)
	if debug {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// 静态文件服务（React SPA）
	distFS, distErr := fs.Sub(web.FS, "dist")
	if distErr != nil {
		panic("failed to open embedded dist: " + distErr.Error())
	}
	fileServer := http.FileServer(http.FS(distFS))

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") {
			c.JSON(404, gin.H{"code": 404, "message": "接口不存在"})
			return
		}
		// 尝试提供静态资源，不存在则 fallback 到 index.html（SPA 路由）
		trimmed := strings.TrimPrefix(path, "/")
		if f, err := distFS.Open(trimmed); err == nil {
			f.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
		} else {
			c.FileFromFS("index.html", http.FS(distFS))
		}
	})

	api := r.Group("/api")

	// ========== 认证（公开） ==========
	{
		authLimiter := middleware.RateLimitMiddleware(rdb, middleware.DefaultRateLimitWindow, middleware.DefaultRateLimitMaxRequests)
		api.POST("/auth/login", authLimiter, authHandler.Login)
		api.POST("/auth/verify-2fa", authLimiter, authHandler.Verify2FA)
		api.POST("/auth/totp/setup-init", authLimiter, authHandler.SetupTOTPPublic)
		api.POST("/auth/totp/confirm-init", authLimiter, authHandler.ConfirmTOTPPublic)
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
		tenantKeys.GET("/status", tenantKeyHandler.Status)
		tenantKeys.POST("/consume", tenantKeyHandler.Consume)
		tenantKeys.GET("/export", tenantKeyHandler.ExportKeys)
		tenantKeys.GET("/export/json", tenantKeyHandler.ExportKeysJSON)
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
		tenant.GET("/login-logs", tenantUsageLogHandler.LoginLogs)
	}

	// ========== 服务账号 API ==========
	serviceAPI := api.Group("/service")
	serviceAPI.Use(middleware.ServiceAuthMiddleware(saSvc, db))
	{
		serviceAPI.POST("/keys", middleware.TenantBusinessGuard(db), tenantSAHandler.ServiceCreateKey)
		serviceAPI.GET("/keys", tenantSAHandler.ServiceListKeys)
		serviceAPI.GET("/keys/status", tenantSAHandler.ServiceGetKeyStatus)
		serviceAPI.POST("/keys/consume", tenantSAHandler.ServiceConsumeKey)
		serviceAPI.GET("/keys/export", tenantSAHandler.ServiceExportKeys)
		serviceAPI.GET("/keys/export/json", tenantSAHandler.ServiceExportKeysJSON)
		serviceAPI.GET("/keys/:id", tenantSAHandler.ServiceGetKey)
		serviceAPI.PATCH("/keys/:id", middleware.TenantBusinessGuard(db), tenantSAHandler.ServiceUpdateKey)
		serviceAPI.PATCH("/keys/:id/disable", middleware.TenantBusinessGuard(db), tenantSAHandler.ServiceDisableKey)
		serviceAPI.PATCH("/keys/:id/enable", middleware.TenantBusinessGuard(db), tenantSAHandler.ServiceEnableKey)
		serviceAPI.DELETE("/keys/:id", middleware.TenantBusinessGuard(db), tenantSAHandler.ServiceDeleteKey)
	}

	return r
}
