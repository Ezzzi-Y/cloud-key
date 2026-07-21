package handler

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"CloudKey/internal/service"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ========== Tenant Admin API (tenant-scoped) ==========

// TenantKeyHandler handles tenant-scoped key management endpoints.
type TenantKeyHandler struct {
	keySvc        *service.KeyService
	usageLogSvc   *service.UsageLogService
	balanceLogSvc *service.BalanceLogService
	db            *gorm.DB
	recordParams  bool
}

func NewTenantKeyHandler(keySvc *service.KeyService, usageLogSvc *service.UsageLogService, balanceLogSvc *service.BalanceLogService, db *gorm.DB, recordParams bool) *TenantKeyHandler {
	return &TenantKeyHandler{keySvc: keySvc, usageLogSvc: usageLogSvc, balanceLogSvc: balanceLogSvc, db: db, recordParams: recordParams}
}

// parseExpireAt converts an optional expiry string pointer into *time.Time.
// Returns nil if raw is nil or empty.
func parseExpireAt(raw *string) (*time.Time, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02 15:04:05", *raw)
	if err != nil {
		return nil, fmt.Errorf("expire_at 格式错误，应为 YYYY-MM-DD HH:MM:SS")
	}
	return &t, nil
}

// Status 查询卡密状态
// @Summary     查询卡密状态
// @Description 根据卡密值查询卡密状态，不扣减额度
// @Tags        租户-卡密管理
// @Produce     json
// @Security    ApiKeyAuth
// @Param       sk query string true "卡密值"
// @Success     200 {object} Response "卡密状态信息"
// @Failure     400 {object} Response "缺少卡密参数"
// @Failure     404 {object} Response "卡密不存在"
// @Router      /tenant/keys/status [get]
func (h *TenantKeyHandler) Status(c *gin.Context) {
	tenantID := getTenantID(c)
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

type ConsumeRequest struct {
	Key    string `json:"key" binding:"required" example:"CK-xxxx-xxxx"`
	Amount int64  `json:"amount" example:"1"`
}

// Consume 扣减卡密额度
// @Summary     扣减卡密额度
// @Description 扣减指定卡密的剩余额度
// @Tags        租户-卡密管理
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       body body ConsumeRequest true "扣减参数"
// @Success     200 {object} Response "扣减结果"
// @Failure     400 {object} Response "参数错误或卡密无效"
// @Router      /tenant/keys/consume [post]
func (h *TenantKeyHandler) Consume(c *gin.Context) {
	tenantID := getTenantID(c)

	var req ConsumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.Amount <= 0 {
		req.Amount = 1
	}

	requestParams := ""
	if h.recordParams {
		requestParams = c.Request.URL.RawQuery
	}

	result, code, err := h.keySvc.ConsumeKeyByTenant(req.Key, req.Amount, tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	if code != 0 {
		key, _ := h.keySvc.FindByRawKeyTenant(req.Key, tenantID)
		keyID, keyAlias := uint64(0), ""
		if key != nil {
			keyID, keyAlias = key.ID, key.Alias
		}
		h.usageLogSvc.Record(service.RecordUsageParams{
			TenantID: tenantID, KeyID: keyID, KeyAlias: keyAlias, Amount: req.Amount,
			IP: c.ClientIP(), UserAgent: c.GetHeader("User-Agent"),
			RequestPath: c.Request.URL.Path, RequestParams: requestParams, ResponseStatus: code,
		})
		BadRequest(c, code, errcode.GetMessage(code))
		return
	}

	key, _ := h.keySvc.FindByRawKeyTenant(req.Key, tenantID)
	keyID, keyAlias := uint64(0), ""
	if key != nil {
		keyID, keyAlias = key.ID, key.Alias
	}
	h.usageLogSvc.Record(service.RecordUsageParams{
		TenantID: tenantID, KeyID: keyID, KeyAlias: keyAlias, Amount: req.Amount,
		IP: c.ClientIP(), UserAgent: c.GetHeader("User-Agent"),
		RequestPath: c.Request.URL.Path, RequestParams: requestParams, ResponseStatus: http.StatusOK,
	})

	Success(c, result)
}

type CreateKeyJSON struct {
	Alias         string  `json:"alias" binding:"required" example:"测试卡密"`
	BillingMode   string  `json:"billing_mode" binding:"required" example:"count"`
	InitialAmount int64   `json:"initial_amount" binding:"required" example:"100"`
	ExpireAt      *string `json:"expire_at" example:"2025-12-31 23:59:59"`
	MaxUsage      *int64  `json:"max_usage" example:"10"`
}

func (h *TenantKeyHandler) getTenantKeyConfig(tenantID uint64) (string, int, int, error) {
	var tenant model.Tenant
	if err := h.db.First(&tenant, tenantID).Error; err != nil {
		return "", 0, 0, fmt.Errorf("failed to get tenant key config: %w", err)
	}
	return tenant.KeyPrefix, tenant.KeyLength, tenant.KeySuffixLength, nil
}

// CreateKey 创建卡密
// @Summary     创建卡密
// @Description 租户管理员创建新卡密，需业务状态正常
// @Tags        租户-卡密管理
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       body body CreateKeyJSON true "卡密参数"
// @Success     200 {object} Response "创建成功，含 raw_key"
// @Failure     400 {object} Response "参数错误"
// @Router      /tenant/keys [post]
func (h *TenantKeyHandler) CreateKey(c *gin.Context) {
	var req CreateKeyJSON
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, http.StatusBadRequest, "参数错误")
		return
	}

	tenantID := getTenantID(c)

	keyPrefix, keyLen, suffixLen, err := h.getTenantKeyConfig(tenantID)
	if err != nil {
		InternalError(c)
		return
	}

	createdBy := "tenant_admin"

	expireAt, err := parseExpireAt(req.ExpireAt)
	if err != nil {
		BadRequest(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.keySvc.CreateKey(service.CreateKeyRequest{
		Alias: req.Alias, BillingMode: model.KeyBillingMode(req.BillingMode),
		InitialAmount: req.InitialAmount, CreatedBy: createdBy,
		ExpireAt: expireAt, MaxUsage: req.MaxUsage,
	}, tenantID, keyPrefix, keyLen, suffixLen)
	if err != nil {
		InternalError(c)
		return
	}

	Success(c, gin.H{
		"id": result.Key.ID, "raw_key": result.RawKey, "alias": result.Key.Alias,
		"key_prefix": result.Key.KeyPrefix, "key_suffix": result.Key.KeySuffix,
		"billing_mode": result.Key.BillingMode, "initial_amount": result.Key.InitialAmount,
		"remaining_amount": result.Key.RemainingAmount, "status": result.Key.Status,
		"created_by": result.Key.CreatedBy, "created_at": result.Key.CreatedAt,
		"expire_at": result.Key.ExpireAt, "max_usage": result.Key.MaxUsage,
	})
}

// ListKeys 卡密列表
// @Summary     卡密列表
// @Tags        租户-卡密管理
// @Produce     json
// @Security    ApiKeyAuth
// @Param       page     query int    false "页码"     default(1)
// @Param       page_size query int   false "每页数量" default(20)
// @Param       status   query string false "状态过滤: unused/used/disabled/expired"
// @Param       search   query string false "关键字搜索"
// @Success     200 {object} Response{data=PageData} "分页卡密列表"
// @Router      /tenant/keys [get]
func (h *TenantKeyHandler) ListKeys(c *gin.Context) {
	tenantID := getTenantID(c)
	page, pageSize := pageParams(c)

	keys, total, err := h.keySvc.ListKeys(service.KeyListQuery{
		Page: page, PageSize: pageSize,
		Status: c.Query("status"), Search: c.Query("search"),
	}, tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	SuccessPaginated(c, keys, total, page, pageSize)
}

// GetKey 卡密详情
// @Summary     卡密详情
// @Tags        租户-卡密管理
// @Produce     json
// @Security    ApiKeyAuth
// @Param       id   path int true "卡密ID"
// @Success     200 {object} Response "卡密详情"
// @Failure     404 {object} Response "卡密不存在"
// @Router      /tenant/keys/{id} [get]
func (h *TenantKeyHandler) GetKey(c *gin.Context) {
	tenantID := getTenantID(c)

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

// UpdateKey 更新卡密
// @Summary     更新卡密
// @Tags        租户-卡密管理
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       id   path int true "卡密ID"
// @Param       body body object true "更新字段" Schema({"alias":"string","remaining_amount":0})
// @Success     200 {object} Response "更新成功"
// @Failure     400 {object} Response "参数错误"
// @Router      /tenant/keys/{id} [patch]
func (h *TenantKeyHandler) UpdateKey(c *gin.Context) {
	tenantID := getTenantID(c)

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

type AdjustBalanceRequest struct {
	Delta  int64  `json:"delta" binding:"required" example:"100"`
	Remark string `json:"remark" example:"管理员充值"`
}

// AdjustBalance 调整卡密额度（管理员行为）
// @Summary     调整卡密额度
// @Description 管理员增加或减少卡密额度，仅支持增量操作，所有变动记录在流转日志中
// @Tags        租户-卡密管理
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       id   path int true "卡密ID"
// @Param       body body AdjustBalanceRequest true "调整参数"
// @Success     200 {object} Response "调整结果"
// @Failure     400 {object} Response "参数错误"
// @Router      /tenant/keys/{id}/adjust-balance [post]
func (h *TenantKeyHandler) AdjustBalance(c *gin.Context) {
	tenantID := getTenantID(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, http.StatusBadRequest, "无效的卡密 ID")
		return
	}

	var req AdjustBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, http.StatusBadRequest, "参数错误")
		return
	}

	if req.Delta == 0 {
		BadRequest(c, errcode.CodeInvalidAdjustment, "调整量不能为 0")
		return
	}

	result, err := h.keySvc.AdjustBalance(id, tenantID, service.AdjustBalanceRequest{
		Delta:    req.Delta,
		Operator: "tenant_admin",
		Remark:   req.Remark,
	})
	if err != nil {
		BadRequest(c, errcode.CodeInvalidAdjustment, err.Error())
		return
	}

	// 记录流转日志
	key, _ := h.keySvc.GetKeyDetail(id, tenantID)
	keyAlias := ""
	if key != nil {
		keyAlias = key.Alias
	}
	_ = h.balanceLogSvc.Record(service.RecordBalanceParams{
		TenantID:     tenantID,
		KeyID:        id,
		KeyAlias:     keyAlias,
		Delta:        req.Delta,
		BeforeAmount: result.BeforeAmount,
		AfterAmount:  result.AfterAmount,
		Operator:     "tenant_admin",
		Remark:       req.Remark,
	})

	Success(c, result)
}

// DisableKey 禁用卡密
// @Summary     禁用卡密
// @Tags        租户-卡密管理
// @Produce     json
// @Security    ApiKeyAuth
// @Param       id   path int true "卡密ID"
// @Success     200 {object} Response "禁用成功"
// @Failure     400 {object} Response "无效的卡密 ID"
// @Router      /tenant/keys/{id}/disable [patch]
func (h *TenantKeyHandler) DisableKey(c *gin.Context) {
	tenantID := getTenantID(c)

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

// EnableKey 启用卡密
// @Summary     启用卡密
// @Tags        租户-卡密管理
// @Produce     json
// @Security    ApiKeyAuth
// @Param       id   path int true "卡密ID"
// @Success     200 {object} Response "启用成功"
// @Failure     400 {object} Response "无效的卡密 ID"
// @Router      /tenant/keys/{id}/enable [patch]
func (h *TenantKeyHandler) EnableKey(c *gin.Context) {
	tenantID := getTenantID(c)

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

// DeleteKey 删除卡密
// @Summary     删除卡密
// @Tags        租户-卡密管理
// @Produce     json
// @Security    ApiKeyAuth
// @Param       id   path int true "卡密ID"
// @Success     200 {object} Response "删除成功"
// @Failure     400 {object} Response "无效的卡密 ID"
// @Router      /tenant/keys/{id} [delete]
func (h *TenantKeyHandler) DeleteKey(c *gin.Context) {
	tenantID := getTenantID(c)

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

// ExportKeys 导出卡密（文本格式）
// @Summary     导出卡密（文本格式）
// @Tags        租户-卡密管理
// @Produce     json
// @Security    ApiKeyAuth
// @Success     200 {object} Response "导出数据"
// @Router      /tenant/keys/export [get]
func (h *TenantKeyHandler) ExportKeys(c *gin.Context) {
	tenantID := getTenantID(c)

	keys, err := h.keySvc.ExportKeys(tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, keys)
}

// ExportKeysJSON 导出卡密（JSON 格式）
// @Summary     导出卡密（JSON 格式）
// @Tags        租户-卡密管理
// @Produce     json
// @Security    ApiKeyAuth
// @Success     200 {object} Response "导出数据 JSON 数组"
// @Router      /tenant/keys/export/json [get]
func (h *TenantKeyHandler) ExportKeysJSON(c *gin.Context) {
	tenantID := getTenantID(c)

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

// ========== Shared helpers ==========

func pageParams(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}
