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

func (h *TenantServiceAccountHandler) ListServiceAccounts(c *gin.Context) {
	tenantID := getTenantID(c)
	accounts, err := h.serviceAccountSvc.ListServiceAccounts(tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, accounts)
}

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
