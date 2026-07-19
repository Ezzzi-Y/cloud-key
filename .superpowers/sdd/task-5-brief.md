### Task 5: 创建 AuthHandler（统一登录入口）

**Files:**
- Create: `internal/handler/auth_handler.go`

**Interfaces:**
- Consumes: `AuthService.Login`, `AuthService.VerifyTOTP`, `AuthService.GenerateTOTPSecret`, `AuthService.ConfirmTOTPSetup`, `AuthService.ChangePassword`, `AuthService.GetProfile`
- Produces: `AuthHandler` with `Login`, `Verify2FA`, `SetupTOTP`, `ConfirmTOTP`, `SetupTOTPPublic`, `ConfirmTOTPPublic`, `Profile`, `ChangePassword`

- [ ] **Step 1: 创建 `internal/handler/auth_handler.go`**

```go
package handler

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"CloudKey/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc     *service.AuthService
	loginLogSvc *service.LoginLogService
}

func NewAuthHandler(authSvc *service.AuthService, loginLogSvc *service.LoginLogService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, loginLogSvc: loginLogSvc}
}

// ========== 登录流程 ==========

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeInvalidCredentials, "参数错误")
		return
	}

	result, err := h.authSvc.Login(req.Username, req.Password)
	if err != nil {
		InternalError(c)
		return
	}
	if result == nil {
		h.loginLogSvc.RecordLogin(0, nil, c.ClientIP(), c.GetHeader("User-Agent"), false)
		Unauthorized(c, errcode.CodeInvalidCredentials, errcode.GetMessage(errcode.CodeInvalidCredentials))
		return
	}

	h.loginLogSvc.RecordLogin(result.UserID, nil, c.ClientIP(), c.GetHeader("User-Agent"), !result.RequireTOTP)

	if result.RequireTOTP {
		Success(c, gin.H{"require_totp": true, "user_id": result.UserID})
		return
	}

	Success(c, gin.H{
		"require_totp": false, "need_setup": true,
		"user_id": result.UserID, "message": "请设置 TOTP 两步验证",
	})
}

// ========== 2FA ==========

type Verify2FARequest struct {
	UserID uint64 `json:"user_id" binding:"required"`
	Code   string `json:"code" binding:"required"`
}

func (h *AuthHandler) Verify2FA(c *gin.Context) {
	var req Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeTOTPFailed, "参数错误")
		return
	}

	_, resp, err := h.authSvc.VerifyTOTP(req.UserID, req.Code)
	if err != nil {
		InternalError(c)
		return
	}
	if resp == nil {
		Unauthorized(c, errcode.CodeTOTPFailed, errcode.GetMessage(errcode.CodeTOTPFailed))
		return
	}

	Success(c, gin.H{
		"token":     resp.Token,
		"token_type": "Bearer",
		"role":      resp.Role,
		"tenant_id": resp.TenantID,
		"username":  resp.Username,
	})
}

// ========== TOTP 设置（已认证用户） ==========

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
	var req Verify2FARequest
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

// ========== TOTP 设置（公开接口：首次设置，无需 JWT） ==========

func (h *AuthHandler) SetupTOTPPublic(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeTOTPFailed, "参数错误")
		return
	}

	profile, err := h.authSvc.GetProfile(req.UserID)
	if err != nil || profile == nil {
		BadRequest(c, errcode.CodeInvalidCredentials, "用户不存在")
		return
	}
	if profile.TotpSetup {
		BadRequest(c, errcode.CodeTOTPFailed, "TOTP 已设置，请直接登录")
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
		UserID uint64 `json:"user_id" binding:"required"`
		Code   string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeTOTPFailed, "参数错误")
		return
	}

	profile, err := h.authSvc.GetProfile(req.UserID)
	if err != nil || profile == nil {
		BadRequest(c, errcode.CodeInvalidCredentials, "用户不存在")
		return
	}
	if profile.TotpSetup {
		BadRequest(c, errcode.CodeTOTPFailed, "TOTP 已设置，请直接登录")
		return
	}

	if err := h.authSvc.ConfirmTOTPSetup(req.UserID, req.Code); err != nil {
		BadRequest(c, errcode.CodeTOTPFailed, "验证码错误")
		return
	}

	_, resp, _ := h.authSvc.VerifyTOTP(req.UserID, req.Code)
	if resp == nil {
		Success(c, gin.H{"message": "TOTP 设置成功，请重新登录"})
		return
	}

	Success(c, gin.H{
		"token":      resp.Token,
		"token_type": "Bearer",
		"role":       resp.Role,
		"tenant_id":  resp.TenantID,
		"username":   resp.Username,
	})
}

// ========== 个人设置 ==========

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
```

- [ ] **Step 2: 验证编译通过**

Run: `cd D:/MyGoProject/CloudKey && go build ./...`
Expected: 编译通过

- [ ] **Step 3: Commit**

```bash
git add internal/handler/auth_handler.go
git commit -m "feat: add AuthHandler with unified login, 2FA, and profile endpoints"
```

---

