package router

import (
	"CloudKey/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	keyHandler *handler.KeyHandler,
	usageLogHandler *handler.UsageLogHandler,
	statsHandler *handler.StatsHandler,
	adminHandler *handler.AdminHandler,
	serviceHandler *handler.ServiceHandler,
	configHandler *handler.ConfigHandler,
	adminAuthMW gin.HandlerFunc,
	serviceAuthMW gin.HandlerFunc,
) *gin.Engine {
	r := gin.Default()

	// 管理页面
	r.StaticFile("/", "web/admin.html")

	// 公开接口（无需认证）
	api := r.Group("/api")
	{
		api.GET("/key/status", keyHandler.Status)
		api.POST("/key/consume", keyHandler.Consume)
	}

	// 管理后台登录（无需认证）
	adminPublic := r.Group("/api/admin")
	{
		adminPublic.POST("/login", adminHandler.Login)
		adminPublic.POST("/login/verify-2fa", adminHandler.Verify2FA)
		adminPublic.POST("/totp/setup-init", adminHandler.SetupTOTPPublic)
		adminPublic.POST("/totp/confirm-init", adminHandler.ConfirmTOTPPublic)
	}

	// 管理后台（需 JWT 认证）
	adminAuth := r.Group("/api/admin")
	adminAuth.Use(adminAuthMW)
	{
		// 管理员自身
		adminAuth.GET("/profile", adminHandler.Profile)
		adminAuth.PUT("/password", adminHandler.ChangePassword)
		adminAuth.POST("/totp/setup", adminHandler.SetupTOTP)
		adminAuth.POST("/totp/confirm", adminHandler.ConfirmTOTP)
		adminAuth.GET("/login-logs", adminHandler.LoginLogs)

		// 卡密管理
		adminAuth.POST("/keys", keyHandler.CreateKey)
		adminAuth.GET("/keys", keyHandler.ListKeys)
		adminAuth.GET("/keys/export", keyHandler.ExportKeys)
		adminAuth.GET("/keys/:id", keyHandler.GetKey)
		adminAuth.PATCH("/keys/:id", keyHandler.UpdateKey)
		adminAuth.PATCH("/keys/:id/disable", keyHandler.DisableKey)
		adminAuth.PATCH("/keys/:id/enable", keyHandler.EnableKey)
		adminAuth.DELETE("/keys/:id", keyHandler.DeleteKey)

		// 使用记录
		adminAuth.GET("/usage-logs", usageLogHandler.ListLogs)
		adminAuth.GET("/usage-logs/export", usageLogHandler.ExportLogs)

		// 数据统计
		adminAuth.GET("/stats/dashboard", statsHandler.Dashboard)
		adminAuth.GET("/stats/overview", statsHandler.Overview)
		adminAuth.GET("/stats/trends", statsHandler.Trends)
		adminAuth.GET("/stats/top-keys", statsHandler.TopKeys)
		adminAuth.GET("/stats/top-ips", statsHandler.TopIPs)

		// 系统管理
		adminAuth.GET("/configs", configHandler.GetConfigs)
		adminAuth.PUT("/configs", configHandler.UpdateConfigs)

		// 服务账号管理
		adminAuth.GET("/service-accounts", serviceHandler.ListServiceAccounts)
		adminAuth.POST("/service-accounts", serviceHandler.CreateServiceAccount)
		adminAuth.PATCH("/service-accounts/:id/toggle", serviceHandler.ToggleServiceAccount)
		adminAuth.DELETE("/service-accounts/:id", serviceHandler.DeleteServiceAccount)
	}

	// 服务账号接口（需服务密钥认证）
	svc := r.Group("/api/service")
	svc.Use(serviceAuthMW)
	{
		svc.POST("/keys", serviceHandler.ServiceCreateKey)
		svc.GET("/keys", serviceHandler.ServiceListKeys)
	}

	return r
}
