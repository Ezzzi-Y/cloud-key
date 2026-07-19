package handler

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminSvc    *service.AdminService
	loginLogSvc *service.LoginLogService
}

func NewAdminHandler(adminSvc *service.AdminService, loginLogSvc *service.LoginLogService) *AdminHandler {
	return &AdminHandler{adminSvc: adminSvc, loginLogSvc: loginLogSvc}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AdminHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeInvalidCredentials, "参数错误")
		return
	}

	result, err := h.adminSvc.Login(req.Username, req.Password)
	if err != nil {
		InternalError(c)
		return
	}
	if result == nil {
		h.loginLogSvc.RecordLogin(0, c.ClientIP(), c.GetHeader("User-Agent"), false)
		Unauthorized(c, errcode.CodeInvalidCredentials, errcode.GetMessage(errcode.CodeInvalidCredentials))
		return
	}

	h.loginLogSvc.RecordLogin(result.AdminID, c.ClientIP(), c.GetHeader("User-Agent"), !result.RequireTOTP)

	if result.RequireTOTP {
		Success(c, gin.H{"require_totp": true, "admin_id": result.AdminID})
		return
	}

	Success(c, gin.H{
		"require_totp": false, "need_setup": true,
		"admin_id": result.AdminID, "message": "请设置 TOTP 两步验证",
	})
}

type Verify2FARequest struct {
	AdminID uint64 `json:"admin_id" binding:"required"`
	Code    string `json:"code" binding:"required"`
}

func (h *AdminHandler) Verify2FA(c *gin.Context) {
	var req Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeTOTPFailed, "参数错误")
		return
	}

	token, err := h.adminSvc.VerifyTOTP(req.AdminID, req.Code)
	if err != nil {
		InternalError(c)
		return
	}
	if token == "" {
		Unauthorized(c, errcode.CodeTOTPFailed, errcode.GetMessage(errcode.CodeTOTPFailed))
		return
	}

	Success(c, gin.H{"token": token, "token_type": "Bearer"})
}

func (h *AdminHandler) Profile(c *gin.Context) {
	adminID, _ := c.Get("admin_id")
	id, ok := adminID.(uint64)
	if !ok {
		Unauthorized(c, errcode.CodeJWTInvalid, errcode.GetMessage(errcode.CodeJWTInvalid))
		return
	}

	admin, err := h.adminSvc.GetAdminProfile(id)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, admin)
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (h *AdminHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeForbidden, "参数错误")
		return
	}

	adminIDI, _ := c.Get("admin_id")
	id, ok := adminIDI.(uint64)
	if !ok {
		Unauthorized(c, errcode.CodeJWTInvalid, errcode.GetMessage(errcode.CodeJWTInvalid))
		return
	}

	if err := h.adminSvc.ChangePassword(id, req.OldPassword, req.NewPassword); err != nil {
		BadRequest(c, errcode.CodeForbidden, err.Error())
		return
	}
	Success(c, nil)
}

func (h *AdminHandler) SetupTOTP(c *gin.Context) {
	adminIDI, _ := c.Get("admin_id")
	id, ok := adminIDI.(uint64)
	if !ok {
		Unauthorized(c, errcode.CodeJWTInvalid, errcode.GetMessage(errcode.CodeJWTInvalid))
		return
	}

	secret, url, err := h.adminSvc.GenerateTOTPSecret(id)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, gin.H{"secret": secret, "url": url})
}

func (h *AdminHandler) ConfirmTOTP(c *gin.Context) {
	var req Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeTOTPFailed, "参数错误")
		return
	}

	adminIDI, _ := c.Get("admin_id")
	id, ok := adminIDI.(uint64)
	if !ok {
		Unauthorized(c, errcode.CodeJWTInvalid, errcode.GetMessage(errcode.CodeJWTInvalid))
		return
	}

	if err := h.adminSvc.ConfirmTOTPSetup(id, req.Code); err != nil {
		BadRequest(c, errcode.CodeTOTPFailed, err.Error())
		return
	}
	Success(c, nil)
}

func (h *AdminHandler) LoginLogs(c *gin.Context) {
	page, pageSize := pageParams(c)
	logs, total, err := h.loginLogSvc.ListLoginLogs(page, pageSize)
	if err != nil {
		InternalError(c)
		return
	}
	SuccessPaginated(c, logs, total, page, pageSize)
}
