package handler

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"CloudKey/internal/service"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TenantServiceAccountHandler struct {
	keySvc            *service.KeyService
	serviceAccountSvc *service.ServiceAccountService
	balanceLogSvc     *service.BalanceLogService
	db                *gorm.DB
}

func NewTenantServiceAccountHandler(keySvc *service.KeyService, saSvc *service.ServiceAccountService, balanceLogSvc *service.BalanceLogService, db *gorm.DB) *TenantServiceAccountHandler {
	return &TenantServiceAccountHandler{keySvc: keySvc, serviceAccountSvc: saSvc, balanceLogSvc: balanceLogSvc, db: db}
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
// @Param       body body object true "卡密参数" Schema({"alias":"string","remaining_amount":100,"expire_at":"string","max_usage":10})
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
		Alias           string  `json:"alias" binding:"required"`
		RemainingAmount int64   `json:"remaining_amount" binding:"required"`
		ExpireAt        *string `json:"expire_at"`
		MaxUsage        *int64  `json:"max_usage"`
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
		Alias: req.Alias, RemainingAmount: req.RemainingAmount, CreatedBy: createdBy,
		ExpireAt: expireAt, MaxUsage: req.MaxUsage,
	}, sa.TenantID, prefix, keyLen, suffixLen)
	if err != nil {
		InternalError(c)
		return
	}

	// 创建时有初始额度，记录到流转日志
	if result.Key.RemainingAmount > 0 {
		_ = h.balanceLogSvc.Record(service.RecordBalanceParams{
			TenantID:     sa.TenantID,
			KeyID:        result.Key.ID,
			KeyAlias:     result.Key.Alias,
			Delta:        result.Key.RemainingAmount,
			BeforeAmount: 0,
			AfterAmount:  result.Key.RemainingAmount,
			Operator:     createdBy,
			Remark:       "创建卡密初始额度",
		})
	}

	Success(c, gin.H{
		"id": result.Key.ID, "raw_key": result.RawKey, "alias": result.Key.Alias,
		"key_suffix": result.Key.KeySuffix,
		"remaining_amount": result.Key.RemainingAmount, "status": result.Key.Status,
		"expire_at": result.Key.ExpireAt, "max_usage": result.Key.MaxUsage,
	})
}

// ServiceListKeys 服务账号查询卡密列表
// @Summary     服务账号查询卡密列表
// @Tags        服务账号API
// @Produce     json
// @Security    ServiceKeyAuth
// @Param       page      query int    false "页码"     default(1)
// @Param       page_size query int    false "每页数量" default(20)
// @Param       status    query string false "状态过滤: unused/used/disabled/expired"
// @Param       search    query string false "关键字搜索"
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

	keys, total, err := h.keySvc.ListKeys(service.KeyListQuery{
		Page: page, PageSize: pageSize,
		Status: c.Query("status"), Search: c.Query("search"),
	}, sa.TenantID)
	if err != nil {
		InternalError(c)
		return
	}
	SuccessPaginated(c, keys, total, page, pageSize)
}

// getServiceTenantID extracts tenantID from the service account in context.
// Returns (tenantID, true) on success; responds with error and returns (0, false) on failure.
func getServiceTenantID(c *gin.Context) (uint64, bool) {
	saI, exists := c.Get("service_account")
	if !exists {
		Unauthorized(c, errcode.CodeServiceKeyInvalid, "未认证的服务账号")
		return 0, false
	}
	sa, ok := saI.(*model.ServiceAccount)
	if !ok {
		Unauthorized(c, errcode.CodeServiceKeyInvalid, "服务账号信息异常")
		return 0, false
	}
	return sa.TenantID, true
}

// --- 以下为服务账号的卡密操作包装，业务逻辑复用 KeyService，仅路由和认证不同 ---

// ServiceGetKeyStatus 服务账号查询卡密状态
// @Summary     服务账号查询卡密状态
// @Description 通过 X-Service-Key 认证，根据卡密值查询状态
// @Tags        服务账号API
// @Produce     json
// @Security    ServiceKeyAuth
// @Param       sk query string true "卡密值"
// @Success     200 {object} Response "卡密状态信息"
// @Failure     400 {object} Response "缺少卡密参数"
// @Failure     401 {object} Response "服务账号密钥无效"
// @Failure     404 {object} Response "卡密不存在"
// @Router      /service/keys/status [get]
func (h *TenantServiceAccountHandler) ServiceGetKeyStatus(c *gin.Context) {
	tenantID, ok := getServiceTenantID(c)
	if !ok {
		return
	}
	rawKey := c.Query("sk")
	if rawKey == "" {
		BadRequest(c, http.StatusBadRequest, "缺少卡密参数")
		return
	}

	result, err := h.keySvc.GetKeyStatusByTenant(rawKey, tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	if result == nil {
		NotFound(c, errcode.CodeKeyNotFound, errcode.GetMessage(errcode.CodeKeyNotFound))
		return
	}
	Success(c, result)
}

type serviceConsumeReq struct {
	Key    string `json:"key" binding:"required"`
	Amount int64  `json:"amount"`
}

// ServiceConsumeKey 服务账号扣减卡密额度
// @Summary     服务账号扣减卡密额度
// @Description 通过 X-Service-Key 认证，扣减指定卡密的剩余额度
// @Tags        服务账号API
// @Accept      json
// @Produce     json
// @Security    ServiceKeyAuth
// @Param       body body serviceConsumeReq true "扣减参数"
// @Success     200 {object} Response "扣减结果"
// @Failure     400 {object} Response "参数错误或卡密无效"
// @Failure     401 {object} Response "服务账号密钥无效"
// @Router      /service/keys/consume [post]
func (h *TenantServiceAccountHandler) ServiceConsumeKey(c *gin.Context) {
	tenantID, ok := getServiceTenantID(c)
	if !ok {
		return
	}
	var req serviceConsumeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.Amount <= 0 {
		BadRequest(c, errcode.CodeInvalidConsumeAmount, errcode.GetMessage(errcode.CodeInvalidConsumeAmount))
		return
	}

	result, code, err := h.keySvc.ConsumeKeyByTenant(req.Key, req.Amount, tenantID, &service.ConsumeMeta{
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		InternalError(c)
		return
	}
	if code != 0 {
		BadRequest(c, code, errcode.GetMessage(code))
		return
	}
	Success(c, result)
}

// ServiceGetKey 服务账号查询卡密详情
// @Summary     服务账号查询卡密详情
// @Tags        服务账号API
// @Produce     json
// @Security    ServiceKeyAuth
// @Param       id   path int true "卡密ID"
// @Success     200 {object} Response "卡密详情"
// @Failure     400 {object} Response "无效的卡密 ID"
// @Failure     401 {object} Response "服务账号密钥无效"
// @Failure     404 {object} Response "卡密不存在"
// @Router      /service/keys/{id} [get]
func (h *TenantServiceAccountHandler) ServiceGetKey(c *gin.Context) {
	tenantID, ok := getServiceTenantID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, http.StatusBadRequest, "无效的卡密 ID")
		return
	}

	key, err := h.keySvc.GetKeyDetail(id, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, errcode.CodeKeyNotFound, errcode.GetMessage(errcode.CodeKeyNotFound))
			return
		}
		InternalError(c)
		return
	}
	Success(c, key)
}

// ServiceUpdateKey 服务账号更新卡密
// @Summary     服务账号更新卡密
// @Tags        服务账号API
// @Accept      json
// @Produce     json
// @Security    ServiceKeyAuth
// @Param       id   path int true "卡密ID"
// @Param       body body object true "更新字段" Schema({"alias":"string","remaining_amount":0})
// @Success     200 {object} Response "更新成功"
// @Failure     400 {object} Response "参数错误"
// @Failure     401 {object} Response "服务账号密钥无效"
// @Router      /service/keys/{id} [patch]
func (h *TenantServiceAccountHandler) ServiceUpdateKey(c *gin.Context) {
	tenantID, ok := getServiceTenantID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, http.StatusBadRequest, "无效的卡密 ID")
		return
	}

	var req service.UpdateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, http.StatusBadRequest, "参数错误")
		return
	}

	if err := h.keySvc.UpdateKey(id, tenantID, req); err != nil {
		InternalError(c)
		return
	}
	Success(c, nil)
}

type serviceAdjustBalanceReq struct {
	Delta  int64  `json:"delta" binding:"required"`
	Remark string `json:"remark"`
}

// ServiceAdjustBalance 服务账号调整卡密额度
// @Summary     服务账号调整卡密额度
// @Description 通过 X-Service-Key 认证，增加或减少卡密额度，所有变动记录在流转日志中
// @Tags        服务账号API
// @Accept      json
// @Produce     json
// @Security    ServiceKeyAuth
// @Param       id   path int true "卡密ID"
// @Param       body body serviceAdjustBalanceReq true "调整参数"
// @Success     200 {object} Response "调整结果"
// @Failure     400 {object} Response "参数错误"
// @Failure     401 {object} Response "服务账号密钥无效"
// @Router      /service/keys/{id}/adjust-balance [post]
func (h *TenantServiceAccountHandler) ServiceAdjustBalance(c *gin.Context) {
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

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, http.StatusBadRequest, "无效的卡密 ID")
		return
	}

	var req serviceAdjustBalanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, http.StatusBadRequest, "参数错误")
		return
	}

	if req.Delta == 0 {
		BadRequest(c, errcode.CodeInvalidAdjustment, "调整量不能为 0")
		return
	}

	operator := "sa:" + sa.Name

	result, err := h.keySvc.AdjustBalance(id, sa.TenantID, service.AdjustBalanceRequest{
		Delta:    req.Delta,
		Operator: operator,
		Remark:   req.Remark,
	})
	if err != nil {
		BadRequest(c, errcode.CodeInvalidAdjustment, err.Error())
		return
	}

	// 记录流转日志
	key, _ := h.keySvc.GetKeyDetail(id, sa.TenantID)
	keyAlias := ""
	if key != nil {
		keyAlias = key.Alias
	}
	_ = h.balanceLogSvc.Record(service.RecordBalanceParams{
		TenantID:     sa.TenantID,
		KeyID:        id,
		KeyAlias:     keyAlias,
		Delta:        req.Delta,
		BeforeAmount: result.BeforeAmount,
		AfterAmount:  result.AfterAmount,
		Operator:     operator,
		Remark:       req.Remark,
	})

	Success(c, result)
}

// ServiceDisableKey 服务账号禁用卡密
// @Summary     服务账号禁用卡密
// @Tags        服务账号API
// @Produce     json
// @Security    ServiceKeyAuth
// @Param       id   path int true "卡密ID"
// @Success     200 {object} Response "禁用成功"
// @Failure     400 {object} Response "无效的卡密 ID"
// @Failure     401 {object} Response "服务账号密钥无效"
// @Router      /service/keys/{id}/disable [patch]
func (h *TenantServiceAccountHandler) ServiceDisableKey(c *gin.Context) {
	tenantID, ok := getServiceTenantID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, http.StatusBadRequest, "无效的卡密 ID")
		return
	}
	if err := h.keySvc.DisableKey(id, tenantID); err != nil {
		InternalError(c)
		return
	}
	Success(c, nil)
}

// ServiceEnableKey 服务账号启用卡密
// @Summary     服务账号启用卡密
// @Tags        服务账号API
// @Produce     json
// @Security    ServiceKeyAuth
// @Param       id   path int true "卡密ID"
// @Success     200 {object} Response "启用成功"
// @Failure     400 {object} Response "无效的卡密 ID"
// @Failure     401 {object} Response "服务账号密钥无效"
// @Router      /service/keys/{id}/enable [patch]
func (h *TenantServiceAccountHandler) ServiceEnableKey(c *gin.Context) {
	tenantID, ok := getServiceTenantID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, http.StatusBadRequest, "无效的卡密 ID")
		return
	}
	if err := h.keySvc.EnableKey(id, tenantID); err != nil {
		InternalError(c)
		return
	}
	Success(c, nil)
}

// ServiceDeleteKey 服务账号删除卡密
// @Summary     服务账号删除卡密
// @Tags        服务账号API
// @Produce     json
// @Security    ServiceKeyAuth
// @Param       id   path int true "卡密ID"
// @Success     200 {object} Response "删除成功"
// @Failure     400 {object} Response "无效的卡密 ID"
// @Failure     401 {object} Response "服务账号密钥无效"
// @Router      /service/keys/{id} [delete]
func (h *TenantServiceAccountHandler) ServiceDeleteKey(c *gin.Context) {
	tenantID, ok := getServiceTenantID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, http.StatusBadRequest, "无效的卡密 ID")
		return
	}
	if err := h.keySvc.DeleteKey(id, tenantID); err != nil {
		InternalError(c)
		return
	}
	Success(c, nil)
}

// ServiceExportKeys 服务账号导出卡密（文本格式）
// @Summary     服务账号导出卡密（文本格式）
// @Tags        服务账号API
// @Produce     json
// @Security    ServiceKeyAuth
// @Success     200 {object} Response "导出数据"
// @Failure     401 {object} Response "服务账号密钥无效"
// @Router      /service/keys/export [get]
func (h *TenantServiceAccountHandler) ServiceExportKeys(c *gin.Context) {
	tenantID, ok := getServiceTenantID(c)
	if !ok {
		return
	}

	keys, err := h.keySvc.ExportKeys(tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, keys)
}

// ServiceExportKeysJSON 服务账号导出卡密（JSON 格式）
// @Summary     服务账号导出卡密（JSON 格式）
// @Tags        服务账号API
// @Produce     json
// @Security    ServiceKeyAuth
// @Success     200 {object} Response "导出数据 JSON 数组"
// @Failure     401 {object} Response "服务账号密钥无效"
// @Router      /service/keys/export/json [get]
func (h *TenantServiceAccountHandler) ServiceExportKeysJSON(c *gin.Context) {
	tenantID, ok := getServiceTenantID(c)
	if !ok {
		return
	}

	items, err := h.keySvc.ExportKeysJSON(tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	if items == nil {
		items = make([]service.ExportKeyItem, 0)
	}
	Success(c, items)
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
