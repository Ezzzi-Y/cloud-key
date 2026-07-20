package handler

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"CloudKey/internal/service"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TenantServiceAccountHandler struct {
	keySvc            *service.KeyService
	serviceAccountSvc *service.ServiceAccountService
	db                *gorm.DB
}

func NewTenantServiceAccountHandler(keySvc *service.KeyService, saSvc *service.ServiceAccountService, db *gorm.DB) *TenantServiceAccountHandler {
	return &TenantServiceAccountHandler{keySvc: keySvc, serviceAccountSvc: saSvc, db: db}
}

func (h *TenantServiceAccountHandler) getTenantKeyConfig(tenantID uint64) (string, int, int, error) {
	var tenant model.Tenant
	if err := h.db.First(&tenant, tenantID).Error; err != nil {
		return "", 0, 0, fmt.Errorf("failed to get tenant key config: %w", err)
	}
	return tenant.KeyPrefix, tenant.KeyLength, tenant.KeySuffixLength, nil
}

// ServiceCreateKey 服务账号创建卡密
// @Summary     服务账号创建卡密
// @Description 通过 X-Service-Key 认证，服务账号创建卡密
// @Tags        服务账号API
// @Accept      json
// @Produce     json
// @Security    ServiceKeyAuth
// @Param       body body object true "卡密参数" Schema({"alias":"string","billing_mode":"count","initial_amount":100,"expire_at":"string","max_usage":10})
// @Success     200 {object} Response "创建成功"
// @Failure     401 {object} Response "服务账号密钥无效"
// @Router      /service/keys [post]
func (h *TenantServiceAccountHandler) ServiceCreateKey(c *gin.Context) {
	saI, exists := c.Get("service_account")
	if !exists {
		Unauthorized(c, errcode.CodeServiceKeyInvalid, "未认证的服务账号")
		return
	}
	sa, ok := saI.(*model.ServiceAccount)
	if !ok {
		Unauthorized(c, errcode.CodeServiceKeyInvalid, "服务账号信息异常")
		return
	}

	var req struct {
		Alias         string  `json:"alias" binding:"required"`
		BillingMode   string  `json:"billing_mode" binding:"required"`
		InitialAmount int64   `json:"initial_amount" binding:"required"`
		ExpireAt      *string `json:"expire_at"`
		MaxUsage      *int64  `json:"max_usage"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeServiceKeyInvalid, "参数错误")
		return
	}

	createdBy := "sa:" + sa.Name

	expireAt, err := parseExpireAt(req.ExpireAt)
	if err != nil {
		BadRequest(c, errcode.CodeServiceKeyInvalid, err.Error())
		return
	}

	prefix, keyLen, suffixLen, err := h.getTenantKeyConfig(sa.TenantID)
	if err != nil {
		InternalError(c)
		return
	}

	result, err := h.keySvc.CreateKey(service.CreateKeyRequest{
		Alias: req.Alias, BillingMode: model.KeyBillingMode(req.BillingMode),
		InitialAmount: req.InitialAmount, CreatedBy: createdBy,
		ExpireAt: expireAt, MaxUsage: req.MaxUsage,
	}, sa.TenantID, prefix, keyLen, suffixLen)
	if err != nil {
		InternalError(c)
		return
	}

	Success(c, gin.H{
		"id": result.Key.ID, "raw_key": result.RawKey, "alias": result.Key.Alias,
		"key_suffix": result.Key.KeySuffix, "billing_mode": result.Key.BillingMode,
		"initial_amount": result.Key.InitialAmount, "remaining_amount": result.Key.RemainingAmount,
		"status": result.Key.Status,
		"expire_at": result.Key.ExpireAt, "max_usage": result.Key.MaxUsage,
	})
}

// ServiceListKeys 服务账号查询卡密列表
// @Summary     服务账号查询卡密列表
// @Tags        服务账号API
// @Produce     json
// @Security    ServiceKeyAuth
// @Param       page      query int false "页码"     default(1)
// @Param       page_size query int false "每页数量" default(20)
// @Success     200 {object} Response{data=PageData} "分页卡密列表"
// @Failure     401 {object} Response "服务账号密钥无效"
// @Router      /service/keys [get]
func (h *TenantServiceAccountHandler) ServiceListKeys(c *gin.Context) {
	saI, exists := c.Get("service_account")
	if !exists {
		Unauthorized(c, errcode.CodeServiceKeyInvalid, "未认证的服务账号")
		return
	}
	sa, ok := saI.(*model.ServiceAccount)
	if !ok {
		Unauthorized(c, errcode.CodeServiceKeyInvalid, "服务账号信息异常")
		return
	}
	page, pageSize := pageParams(c)

	keys, total, err := h.keySvc.ListKeysByTenant(sa.TenantID, page, pageSize)
	if err != nil {
		InternalError(c)
		return
	}
	SuccessPaginated(c, keys, total, page, pageSize)
}

// ListServiceAccounts 服务账号列表
// @Summary     服务账号列表
// @Tags        租户-服务账号
// @Produce     json
// @Security    ApiKeyAuth
// @Success     200 {object} Response "服务账号列表"
// @Router      /tenant/service-accounts [get]
func (h *TenantServiceAccountHandler) ListServiceAccounts(c *gin.Context) {
	tenantID := getTenantID(c)
	accounts, err := h.serviceAccountSvc.ListServiceAccounts(tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, accounts)
}

// CreateServiceAccount 创建服务账号
// @Summary     创建服务账号
// @Tags        租户-服务账号
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       body body object true "服务账号参数" Schema({"name":"string"})
// @Success     200 {object} Response{data=object{id=int,name=string,raw_key=string,is_active=bool,created_at=string}} "创建成功"
// @Failure     400 {object} Response "参数错误"
// @Router      /tenant/service-accounts [post]
func (h *TenantServiceAccountHandler) CreateServiceAccount(c *gin.Context) {
	tenantID := getTenantID(c)
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeServiceKeyInvalid, "参数错误")
		return
	}

	account, rawKey, err := h.serviceAccountSvc.CreateServiceAccount(req.Name, tenantID)
	if err != nil {
		InternalError(c)
		return
	}

	Success(c, gin.H{
		"id": account.ID, "name": account.Name, "raw_key": rawKey,
		"is_active": account.IsActive, "created_at": account.CreatedAt,
	})
}

// ToggleServiceAccount 启用/禁用服务账号
// @Summary     启用/禁用服务账号
// @Tags        租户-服务账号
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       id   path int true "服务账号ID"
// @Param       body body object true "状态" Schema({"is_active":true})
// @Success     200 {object} Response "操作成功"
// @Failure     400 {object} Response "参数错误"
// @Router      /tenant/service-accounts/{id}/toggle [patch]
func (h *TenantServiceAccountHandler) ToggleServiceAccount(c *gin.Context) {
	tenantID := getTenantID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, errcode.CodeServiceKeyInvalid, "无效的服务账号 ID")
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeServiceKeyInvalid, "参数错误")
		return
	}

	if err := h.serviceAccountSvc.ToggleServiceAccount(id, tenantID, req.IsActive); err != nil {
		InternalError(c)
		return
	}
	Success(c, nil)
}

// DeleteServiceAccount 删除服务账号
// @Summary     删除服务账号
// @Tags        租户-服务账号
// @Produce     json
// @Security    ApiKeyAuth
// @Param       id   path int true "服务账号ID"
// @Success     200 {object} Response "删除成功"
// @Failure     400 {object} Response "无效的服务账号 ID"
// @Router      /tenant/service-accounts/{id} [delete]
func (h *TenantServiceAccountHandler) DeleteServiceAccount(c *gin.Context) {
	tenantID := getTenantID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, errcode.CodeServiceKeyInvalid, "无效的服务账号 ID")
		return
	}

	if err := h.serviceAccountSvc.DeleteServiceAccount(id, tenantID); err != nil {
		InternalError(c)
		return
	}
	Success(c, nil)
}
