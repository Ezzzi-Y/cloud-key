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
	keySvc := service.NewKeyService(db)
	usageLogSvc := service.NewUsageLogService(db)
	statsSvc := service.NewStatsService(db)
	adminSvc := service.NewAdminService(db, cfg.Auth.Secret, cfg.Auth.Expiration)
	serviceAccountSvc := service.NewServiceAccountService(db)
	configSvc := service.NewConfigService(db)
	loginLogSvc := service.NewLoginLogService(db)

	// Init defaults
	if err := configSvc.InitDefaultConfigs(); err != nil {
		log.Warn("初始化默认配置失败", zap.Error(err))
	}

	adminUser := os.Getenv("ADMIN_USERNAME")
	if adminUser == "" {
		adminUser = "admin"
	}
	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminPass == "" {
		log.Fatal("ADMIN_PASSWORD 环境变量未设置")
	}
	if err := adminSvc.SeedAdmin(adminUser, adminPass); err != nil {
		log.Warn("创建初始管理员失败", zap.Error(err))
	}

	// Handlers
	keyHandler := handler.NewKeyHandler(keySvc, usageLogSvc, false)
	usageLogHandler := handler.NewUsageLogHandler(usageLogSvc)
	statsHandler := handler.NewStatsHandler(statsSvc)
	adminHandler := handler.NewAdminHandler(adminSvc, loginLogSvc)
	serviceHandler := handler.NewServiceHandler(keySvc, serviceAccountSvc)
	configHandler := handler.NewConfigHandler(configSvc)

	// Middleware
	adminAuthMW := middleware.AuthMiddleware(cfg.Auth.Secret)
	serviceAuthMW := middleware.ServiceAuthMiddleware(serviceAccountSvc)

	// Gin mode
	if cfg.App.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Router
	r := router.SetupRouter(
		keyHandler, usageLogHandler, statsHandler,
		adminHandler, serviceHandler, configHandler,
		adminAuthMW, serviceAuthMW,
	)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Info("服务器启动", zap.String("address", addr))
	if err := r.Run(addr); err != nil {
		log.Fatal("服务器启动失败", zap.Error(err))
	}
}
