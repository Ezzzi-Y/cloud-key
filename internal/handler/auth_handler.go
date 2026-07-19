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
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

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
	UserID       uint64 `json:"user_id" binding:"required"`
	Code         string `json:"code" binding:"required"`
	PreAuthToken string `json:"pre_auth_token" binding:"required"`
}

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
