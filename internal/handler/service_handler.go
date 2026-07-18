package handler

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"CloudKey/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ServiceHandler struct {
	keySvc            *service.KeyService
	serviceAccountSvc *service.ServiceAccountService
}

func NewServiceHandler(keySvc *service.KeyService, saSvc *service.ServiceAccountService) *ServiceHandler {
	return &ServiceHandler{keySvc: keySvc, serviceAccountSvc: saSvc}
}

func (h *ServiceHandler) ServiceCreateKey(c *gin.Context) {
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
		Alias         string `json:"alias" binding:"required"`
		BillingMode   string `json:"billing_mode" binding:"required"`
		InitialAmount int64  `json:"initial_amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeServiceKeyInvalid, "参数错误")
		return
	}

	createdBy := "sa:" + sa.Name

	result, err := h.keySvc.CreateKey(service.CreateKeyRequest{
		Alias: req.Alias, BillingMode: model.KeyBillingMode(req.BillingMode),
		InitialAmount: req.InitialAmount, CreatedBy: createdBy,
	})
	if err != nil {
		InternalError(c)
		return
	}

	Success(c, gin.H{
		"id": result.Key.ID, "raw_key": result.RawKey, "alias": result.Key.Alias,
		"key_suffix": result.Key.KeySuffix, "billing_mode": result.Key.BillingMode,
		"initial_amount": result.Key.InitialAmount, "remaining_amount": result.Key.RemainingAmount,
		"status": result.Key.Status,
	})
}

func (h *ServiceHandler) ServiceListKeys(c *gin.Context) {
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
	createdBy := "sa:" + sa.Name
	page, pageSize := pageParams(c)

	keys, total, err := h.keySvc.ListKeysByCreatedBy(createdBy, page, pageSize)
	if err != nil {
		InternalError(c)
		return
	}
	SuccessPaginated(c, keys, total, page, pageSize)
}

func (h *ServiceHandler) ListServiceAccounts(c *gin.Context) {
	accounts, err := h.serviceAccountSvc.ListServiceAccounts()
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, accounts)
}

func (h *ServiceHandler) CreateServiceAccount(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeServiceKeyInvalid, "参数错误")
		return
	}

	account, rawKey, err := h.serviceAccountSvc.CreateServiceAccount(req.Name)
	if err != nil {
		InternalError(c)
		return
	}

	Success(c, gin.H{
		"id": account.ID, "name": account.Name, "raw_key": rawKey,
		"is_active": account.IsActive, "created_at": account.CreatedAt,
	})
}

func (h *ServiceHandler) ToggleServiceAccount(c *gin.Context) {
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

	if err := h.serviceAccountSvc.ToggleServiceAccount(id, req.IsActive); err != nil {
		InternalError(c)
		return
	}
	Success(c, nil)
}

func (h *ServiceHandler) DeleteServiceAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, errcode.CodeServiceKeyInvalid, "无效的服务账号 ID")
		return
	}

	if err := h.serviceAccountSvc.DeleteServiceAccount(id); err != nil {
		InternalError(c)
		return
	}
	Success(c, nil)
}
