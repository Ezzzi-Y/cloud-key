package handler

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SuperHandler struct {
	tenantSvc   *service.TenantService
	configSvc   *service.ConfigService
	loginLogSvc *service.LoginLogService
}

func NewSuperHandler(tenantSvc *service.TenantService, configSvc *service.ConfigService, loginLogSvc *service.LoginLogService) *SuperHandler {
	return &SuperHandler{tenantSvc: tenantSvc, configSvc: configSvc, loginLogSvc: loginLogSvc}
}

// GET /api/super/tenants
func (h *SuperHandler) ListTenants(c *gin.Context) {
	tenants, err := h.tenantSvc.ListTenants()
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, tenants)
}

// POST /api/super/tenants
func (h *SuperHandler) CreateTenant(c *gin.Context) {
	var req service.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeForbidden, "参数错误")
		return
	}
	result, err := h.tenantSvc.CreateTenant(req)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, gin.H{
		"tenant":         result.Tenant,
		"admin_username": result.AdminUsername,
		"admin_password": result.AdminPassword,
	})
}

// GET /api/super/tenants/:id
func (h *SuperHandler) GetTenant(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, errcode.CodeTenantNotFound, "无效的ID")
		return
	}
	tenant, err := h.tenantSvc.GetTenant(id)
	if err != nil {
		NotFound(c, errcode.CodeTenantNotFound, "租户不存在")
		return
	}
	Success(c, tenant)
}

// PATCH /api/super/tenants/:id
func (h *SuperHandler) UpdateTenant(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, errcode.CodeTenantNotFound, "无效的ID")
		return
	}
	var req service.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeForbidden, "参数错误")
		return
	}
	if err := h.tenantSvc.UpdateTenant(id, req); err != nil {
		InternalError(c)
		return
	}
	Success(c, nil)
}

// PATCH /api/super/tenants/:id/reset-password
func (h *SuperHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, errcode.CodeTenantNotFound, "无效的ID")
		return
	}
	newPass, err := h.tenantSvc.ResetPassword(id)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, gin.H{"new_password": newPass})
}

// GET /api/super/configs
func (h *SuperHandler) GetConfigs(c *gin.Context) {
	configs, err := h.configSvc.GetAllConfigs()
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, configs)
}

// PUT /api/super/configs
func (h *SuperHandler) UpdateConfigs(c *gin.Context) {
	var req []struct {
		Key         string `json:"key" binding:"required"`
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeForbidden, "参数错误")
		return
	}
	for _, item := range req {
		if err := h.configSvc.SetConfig(item.Key, item.Value, item.Description); err != nil {
			InternalError(c)
			return
		}
	}
	Success(c, nil)
}

// GET /api/super/login-logs
func (h *SuperHandler) LoginLogs(c *gin.Context) {
	page, pageSize := pageParams(c)
	logs, total, err := h.loginLogSvc.ListLoginLogs(page, pageSize, nil) // nil = all tenants
	if err != nil {
		InternalError(c)
		return
	}
	SuccessPaginated(c, logs, total, page, pageSize)
}
