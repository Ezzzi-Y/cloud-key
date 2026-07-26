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

// ListTenants 租户列表
// @Summary     租户列表
// @Tags        系统管理
// @Produce     json
// @Security    ApiKeyAuth
// @Success     200 {object} Response "租户列表"
// @Router      /super/tenants [get]
func (h *SuperHandler) ListTenants(c *gin.Context) {
	tenants, err := h.tenantSvc.ListTenants()
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, tenants)
}

// CreateTenant 创建租户
// @Summary     创建租户
// @Tags        系统管理
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       body body object true "租户参数" Schema({"name":"string"})
// @Success     200 {object} Response{data=object{tenant=object,admin_username=string,admin_password=string}} "创建成功"
// @Failure     400 {object} Response "参数错误"
// @Router      /super/tenants [post]
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

// GetTenant 租户详情
// @Summary     租户详情
// @Tags        系统管理
// @Produce     json
// @Security    ApiKeyAuth
// @Param       id   path int true "租户ID"
// @Success     200 {object} Response "租户详情"
// @Failure     404 {object} Response "租户不存在"
// @Router      /super/tenants/{id} [get]
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

// UpdateTenant 更新租户
// @Summary     更新租户
// @Tags        系统管理
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       id   path int true "租户ID"
// @Param       body body object true "更新字段" Schema({"name":"string","expire_at":"string","status":"string"})
// @Success     200 {object} Response "更新成功"
// @Failure     400 {object} Response "参数错误"
// @Router      /super/tenants/{id} [patch]
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

// ResetPassword 重置租户管理员密码
// @Summary     重置租户管理员密码
// @Tags        系统管理
// @Produce     json
// @Security    ApiKeyAuth
// @Param       id   path int true "租户ID"
// @Success     200 {object} Response{data=object{new_password=string}} "重置成功"
// @Failure     400 {object} Response "无效的ID"
// @Router      /super/tenants/{id}/reset-password [patch]
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

// GetConfigs 获取系统配置
// @Summary     获取系统配置
// @Tags        系统管理
// @Produce     json
// @Security    ApiKeyAuth
// @Success     200 {object} Response "配置列表"
// @Router      /super/configs [get]
func (h *SuperHandler) GetConfigs(c *gin.Context) {
	configs, err := h.configSvc.GetAllConfigs()
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, configs)
}

// UpdateConfigs 更新系统配置
// @Summary     更新系统配置
// @Tags        系统管理
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       body body object true "配置数组" Schema([{"key":"string","value":"string","description":"string"}])
// @Success     200 {object} Response "更新成功"
// @Failure     400 {object} Response "参数错误"
// @Router      /super/configs [put]
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

// LoginLogs 登录日志（全租户）
// @Summary     登录日志
// @Tags        系统管理
// @Produce     json
// @Security    ApiKeyAuth
// @Param       page      query int false "页码"     default(1)
// @Param       page_size query int false "每页数量" default(20)
// @Success     200 {object} Response{data=PageData} "分页登录日志"
// @Router      /super/login-logs [get]
func (h *SuperHandler) LoginLogs(c *gin.Context) {
	page, pageSize := pageParams(c)
	logs, total, err := h.loginLogSvc.ListLoginLogs(page, pageSize, nil) // nil = all tenants
	if err != nil {
		InternalError(c)
		return
	}
	SuccessPaginated(c, logs, total, page, pageSize)
}
