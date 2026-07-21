package main

import (
	"CloudKey/internal/config"
	"CloudKey/internal/database"
	"CloudKey/internal/handler"
	"CloudKey/internal/log"
	"CloudKey/internal/model"
	"CloudKey/internal/router"
	"CloudKey/internal/service"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @title       CloudKey API
// @version     1.0
// @description 卡密管理系统 API 文档
// @host        localhost:8080
// @BasePath    /api
// @securityDefinitions.apikey ApiKeyAuth
// @in   header
// @name Authorization
// @description Bearer JWT token, 格式: Bearer <token>
// @securityDefinitions.apikey ServiceKeyAuth
// @in   header
// @name X-Service-Key
// @description 服务账号密钥
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

	// Redis
	rdb, err := database.ConnectRedis(cfg.Redis)
	if err != nil {
		log.Fatal("Redis 连接失败", zap.Error(err))
	}
	defer database.CloseRedis(rdb)
	log.Info("Redis 连接成功")

	// Services
	authSvc := service.NewAuthService(db, rdb, cfg.Auth.Secret, cfg.Auth.Expiration)
	keySvc := service.NewKeyService(db)
	usageLogSvc := service.NewUsageLogService(db)
	balanceLogSvc := service.NewBalanceLogService(db)
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
	superAdminUser := cfg.Auth.SuperAdminUsername
	if superAdminUser == "" {
		superAdminUser = "admin"
	}
	superAdminPass := cfg.Auth.SuperAdminPassword
	if superAdminPass == "" {
		log.Fatal("请在 config.yaml 的 auth.super_admin_password 中配置超级管理员密码")
	}
	if err := authSvc.SeedSuperAdmin(superAdminUser, superAdminPass); err != nil {
		log.Warn("创建超级管理员失败", zap.Error(err))
	}

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc, loginLogSvc)
	superHandler := handler.NewSuperHandler(tenantSvc, configSvc, loginLogSvc)
	tenantKeyHandler := handler.NewTenantKeyHandler(keySvc, usageLogSvc, balanceLogSvc, db, false)
	tenantSAHandler := handler.NewTenantServiceAccountHandler(keySvc, serviceAccountSvc, balanceLogSvc, db)
	tenantStatsHandler := handler.NewTenantStatsHandler(statsSvc)
	tenantUsageLogHandler := handler.NewTenantUsageLogHandler(usageLogSvc, loginLogSvc)
	tenantBalanceLogHandler := handler.NewTenantBalanceLogHandler(balanceLogSvc)
	serviceBalanceLogHandler := handler.NewServiceBalanceLogHandler(balanceLogSvc)

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
		tenantBalanceLogHandler, serviceBalanceLogHandler,
		cfg.Auth.Secret, db, serviceAccountSvc, rdb, cfg.App.Debug,
	)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Info("服务器启动", zap.String("address", addr))
	if err := r.Run(addr); err != nil {
		log.Fatal("服务器启动失败", zap.Error(err))
	}
}
