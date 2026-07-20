package handler

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"CloudKey/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc     *service.AuthService
	loginLogSvc *service.LoginLogService
}

func NewAuthHandler(authSvc *service.AuthService, loginLogSvc *service.LoginLogService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, loginLogSvc: loginLogSvc}
}

// ========== Login flow ==========

type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"123456"`
}

// Login 用户名密码登录
// @Summary     用户名密码登录
// @Description 用户名密码验证，成功后返回 pre_auth_token 用于后续 TOTP 验证
// @Tags        认证
// @Accept      json
// @Produce     json
// @Param       body body LoginRequest true "登录参数"
// @Success     200 {object} Response "成功: data 含 require_totp/user_id/pre_auth_token"
// @Failure     401 {object} Response "用户名或密码错误"
// @Failure     403 {object} Response "账号已被锁定"
// @Failure     429 {object} Response "请求过于频繁"
// @Router      /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Unauthorized(c, errcode.CodeInvalidCredentials, errcode.GetMessage(errcode.CodeInvalidCredentials))
		return
	}

	result, err := h.authSvc.Login(req.Username, req.Password)
	if err != nil {
		// Account locked
		if strings.Contains(err.Error(), "锁定") {
			Forbidden(c, errcode.CodeAccountLocked, errcode.GetMessage(errcode.CodeAccountLocked))
			return
		}
		InternalError(c)
		return
	}
	if result == nil {
		h.loginLogSvc.RecordLogin(0, nil, c.ClientIP(), c.GetHeader("User-Agent"), false)
		Unauthorized(c, errcode.CodeInvalidCredentials, errcode.GetMessage(errcode.CodeInvalidCredentials))
		return
	}

	// Record login with tenant ID from the user record.
	// For TOTP-required users, login is not fully successful until TOTP is verified.
	h.loginLogSvc.RecordLogin(result.UserID, result.TenantID, c.ClientIP(), c.GetHeader("User-Agent"), !result.RequireTOTP)

	if result.RequireTOTP {
		Success(c, gin.H{
			"require_totp":  true,
			"user_id":       result.UserID,
			"pre_auth_token": result.PreAuthToken,
		})
		return
	}

	Success(c, gin.H{
		"require_totp":  false,
		"need_setup":    true,
		"user_id":       result.UserID,
		"pre_auth_token": result.PreAuthToken,
		"message":       "请设置 TOTP 两步验证",
	})
}

// ========== 2FA ==========

type Verify2FARequest struct {
	UserID       uint64 `json:"user_id" binding:"required" example:"1"`
	Code         string `json:"code" binding:"required" example:"123456"`
	PreAuthToken string `json:"pre_auth_token" binding:"required"`
}

// Verify2FA TOTP 二次验证
// @Summary     TOTP 二次验证
// @Description 登录成功后使用 pre_auth_token + TOTP 验证码完成认证，返回 JWT
// @Tags        认证
// @Accept      json
// @Produce     json
// @Param       body body Verify2FARequest true "验证参数"
// @Success     200 {object} Response{data=object{token=string,token_type=string,role=string,tenant_id=int,username=string}} "验证成功，返回 JWT"
// @Failure     401 {object} Response "验证码错误或 pre_auth_token 无效"
// @Failure     429 {object} Response "请求过于频繁"
// @Router      /auth/verify-2fa [post]
func (h *AuthHandler) Verify2FA(c *gin.Context) {
	var req Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Unauthorized(c, errcode.CodeTOTPFailed, errcode.GetMessage(errcode.CodeTOTPFailed))
		return
	}

	_, resp, err := h.authSvc.VerifyTOTP(req.UserID, req.Code, req.PreAuthToken)
	if err != nil {
		InternalError(c)
		return
	}
	if resp == nil {
		h.loginLogSvc.RecordLogin(req.UserID, nil, c.ClientIP(), c.GetHeader("User-Agent"), false)
		Unauthorized(c, errcode.CodeTOTPFailed, errcode.GetMessage(errcode.CodeTOTPFailed))
		return
	}

	// Record successful login after TOTP verification
	h.loginLogSvc.RecordLogin(req.UserID, resp.TenantID, c.ClientIP(), c.GetHeader("User-Agent"), true)

	Success(c, gin.H{
		"token":      resp.Token,
		"token_type": "Bearer",
		"role":       resp.Role,
		"tenant_id":  resp.TenantID,
		"username":   resp.Username,
	})
}

// ========== TOTP setup (authenticated user) ==========

// SetupTOTP 生成 TOTP 密钥（已认证）
// @Summary     生成 TOTP 密钥
// @Description 已登录用户生成新的 TOTP 密钥，用于更换验证器
// @Tags        认证
// @Produce     json
// @Security    ApiKeyAuth
// @Success     200 {object} Response{data=object{secret=string,url=string}} "生成成功"
// @Failure     401 {object} Response "未认证"
// @Router      /super/totp/setup [post]
// @Router      /tenant/totp/setup [post]
func (h *AuthHandler) SetupTOTP(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		Unauthorized(c, errcode.CodeJWTInvalid, errcode.GetMessage(errcode.CodeJWTInvalid))
		return
	}

	secret, url, err := h.authSvc.GenerateTOTPSecret(userID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, gin.H{"secret": secret, "url": url})
}

// ConfirmTOTP 确认 TOTP 绑定（已认证）
// @Summary     确认 TOTP 绑定
// @Description 已登录用户输入验证码确认 TOTP 绑定
// @Tags        认证
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       body body object true "验证码" Schema({"code":"string"})
// @Success     200 {object} Response "绑定成功"
// @Failure     400 {object} Response "验证码错误"
// @Failure     401 {object} Response "未认证"
// @Router      /super/totp/confirm [post]
// @Router      /tenant/totp/confirm [post]
func (h *AuthHandler) ConfirmTOTP(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeTOTPFailed, "参数错误")
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		Unauthorized(c, errcode.CodeJWTInvalid, errcode.GetMessage(errcode.CodeJWTInvalid))
		return
	}

	if err := h.authSvc.ConfirmTOTPSetup(userID, req.Code); err != nil {
		BadRequest(c, errcode.CodeTOTPFailed, err.Error())
		return
	}
	Success(c, nil)
}

// ========== TOTP setup (public endpoint: first-time setup, no JWT) ==========

// SetupTOTPPublic 初始化 TOTP 绑定（首次登录）
// @Summary     初始化 TOTP 绑定
// @Description 首次登录用户生成 TOTP 密钥，需先通过密码验证获取 user_id
// @Tags        认证
// @Accept      json
// @Produce     json
// @Param       body body object true "用户ID" Schema({"user_id":1})
// @Success     200 {object} Response{data=object{secret=string,url=string}} "生成成功"
// @Failure     401 {object} Response "用户不存在或 TOTP 已设置"
// @Failure     429 {object} Response "请求过于频繁"
// @Router      /auth/totp/setup-init [post]
func (h *AuthHandler) SetupTOTPPublic(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Unauthorized(c, errcode.CodeInvalidCredentials, errcode.GetMessage(errcode.CodeInvalidCredentials))
		return
	}

	profile, err := h.authSvc.GetProfile(req.UserID)
	if err != nil || profile == nil {
		Unauthorized(c, errcode.CodeInvalidCredentials, errcode.GetMessage(errcode.CodeInvalidCredentials))
		return
	}
	if profile.TotpSetup {
		Unauthorized(c, errcode.CodeInvalidCredentials, errcode.GetMessage(errcode.CodeInvalidCredentials))
		return
	}

	secret, url, err := h.authSvc.GenerateTOTPSecret(req.UserID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, gin.H{"secret": secret, "url": url})
}

// ConfirmTOTPPublic 确认 TOTP 绑定并登录（首次登录）
// @Summary     确认 TOTP 绑定并登录
// @Description 首次登录用户确认 TOTP 绑定后自动登录，返回 JWT
// @Tags        认证
// @Accept      json
// @Produce     json
// @Param       body body object true "验证参数" Schema({"user_id":1,"code":"string","pre_auth_token":"string"})
// @Success     200 {object} Response{data=object{token=string,token_type=string,role=string,tenant_id=int,username=string}} "绑定成功并返回 JWT"
// @Failure     400 {object} Response "验证码错误"
// @Failure     401 {object} Response "pre_auth_token 无效"
// @Failure     429 {object} Response "请求过于频繁"
// @Router      /auth/totp/confirm-init [post]
func (h *AuthHandler) ConfirmTOTPPublic(c *gin.Context) {
	var req struct {
		UserID       uint64 `json:"user_id" binding:"required"`
		Code         string `json:"code" binding:"required"`
		PreAuthToken string `json:"pre_auth_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Unauthorized(c, errcode.CodeTOTPFailed, errcode.GetMessage(errcode.CodeTOTPFailed))
		return
	}

	// Validate pre-auth token first
	if _, err := h.authSvc.ValidatePreAuthToken(req.UserID, req.PreAuthToken); err != nil {
		Unauthorized(c, errcode.CodePreAuthInvalid, errcode.GetMessage(errcode.CodePreAuthInvalid))
		return
	}

	profile, err := h.authSvc.GetProfile(req.UserID)
	if err != nil || profile == nil {
		Unauthorized(c, errcode.CodeTOTPFailed, errcode.GetMessage(errcode.CodeTOTPFailed))
		return
	}
	if profile.TotpSetup {
		Unauthorized(c, errcode.CodeTOTPFailed, errcode.GetMessage(errcode.CodeTOTPFailed))
		return
	}

	if err := h.authSvc.ConfirmTOTPSetup(req.UserID, req.Code); err != nil {
		BadRequest(c, errcode.CodeTOTPFailed, errcode.GetMessage(errcode.CodeTOTPFailed))
		return
	}

	// TOTP setup confirmed — issue JWT directly (pre-auth token already validated above)
	resp, err := h.authSvc.GenerateLoginJWT(req.UserID)
	if err != nil {
		// TOTP saved but JWT failed — user can re-login
		Success(c, gin.H{"message": "TOTP 设置成功，请重新登录"})
		return
	}

	// Record successful login
	h.loginLogSvc.RecordLogin(req.UserID, resp.TenantID, c.ClientIP(), c.GetHeader("User-Agent"), true)

	Success(c, gin.H{
		"token":      resp.Token,
		"token_type": "Bearer",
		"role":       resp.Role,
		"tenant_id":  resp.TenantID,
		"username":   resp.Username,
	})
}

// ========== Profile ==========

// Profile 获取当前用户信息
// @Summary     获取当前用户信息
// @Tags        认证
// @Produce     json
// @Security    ApiKeyAuth
// @Success     200 {object} Response "用户信息"
// @Failure     401 {object} Response "未认证"
// @Router      /super/profile [get]
// @Router      /tenant/profile [get]
func (h *AuthHandler) Profile(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		Unauthorized(c, errcode.CodeJWTInvalid, errcode.GetMessage(errcode.CodeJWTInvalid))
		return
	}
	user, err := h.authSvc.GetProfile(userID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, user)
}

// ChangePassword 修改密码
// @Summary     修改密码
// @Tags        认证
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       body body object true "密码参数" Schema({"old_password":"string","new_password":"string"})
// @Success     200 {object} Response "修改成功"
// @Failure     400 {object} Response "旧密码错误"
// @Failure     401 {object} Response "未认证"
// @Router      /super/password [put]
// @Router      /tenant/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeForbidden, "参数错误")
		return
	}

	userID := getUserID(c)
	if err := h.authSvc.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		BadRequest(c, errcode.CodeForbidden, err.Error())
		return
	}
	Success(c, nil)
}

// ========== Login logs ==========

func (h *AuthHandler) LoginLogs(c *gin.Context) {
	page, pageSize := pageParams(c)
	tenantID := getTenantID(c)
	var tenantIDPtr *uint64
	if tenantID > 0 {
		tenantIDPtr = &tenantID
	}
	logs, total, err := h.loginLogSvc.ListLoginLogs(page, pageSize, tenantIDPtr)
	if err != nil {
		InternalError(c)
		return
	}
	SuccessPaginated(c, logs, total, page, pageSize)
}

// ========== helpers ==========

func getUserID(c *gin.Context) uint64 {
	idI, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	id, ok := idI.(uint64)
	if !ok {
		return 0
	}
	return id
}

func getTenantID(c *gin.Context) uint64 {
	idI, exists := c.Get("tenant_id")
	if !exists {
		return 0
	}
	id, ok := idI.(uint64)
	if !ok {
		return 0
	}
	return id
}

func getRole(c *gin.Context) model.UserRole {
	r, _ := c.Get("role")
	role, _ := r.(model.UserRole)
	return role
}
