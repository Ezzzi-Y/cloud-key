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

// ========== Public API (no auth, no tenant scope) ==========

// KeyHandler handles public Status and Consume endpoints.
type KeyHandler struct {
	keySvc       *service.KeyService
	usageLogSvc  *service.UsageLogService
	recordParams bool
}

func NewKeyHandler(keySvc *service.KeyService, usageLogSvc *service.UsageLogService, recordParams bool) *KeyHandler {
	return &KeyHandler{keySvc: keySvc, usageLogSvc: usageLogSvc, recordParams: recordParams}
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

// Status queries key status (no deduction).
func (h *KeyHandler) Status(c *gin.Context) {
	rawKey := c.Query("sk")
	if rawKey == "" {
		BadRequest(c, http.StatusBadRequest, "缺少卡密参数")
		return
	}

	result, err := h.keySvc.GetKeyStatus(rawKey)
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
	Key    string `json:"key" binding:"required"`
	Amount int64  `json:"amount"`
}

// Consume deducts key amount (public API, no tenant scope).
func (h *KeyHandler) Consume(c *gin.Context) {
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

	result, code, err := h.keySvc.ConsumeKey(req.Key, req.Amount)
	if err != nil {
		InternalError(c)
		return
	}
	if code != 0 {
		key, _ := h.keySvc.FindByRawKey(req.Key)
		keyID, keyAlias, keyTenantID := uint64(0), "", uint64(0)
		if key != nil {
			keyID, keyAlias, keyTenantID = key.ID, key.Alias, key.TenantID
		}
		h.usageLogSvc.Record(service.RecordUsageParams{
			TenantID: keyTenantID, KeyID: keyID, KeyAlias: keyAlias, Amount: req.Amount,
			IP: c.ClientIP(), UserAgent: c.GetHeader("User-Agent"),
			RequestPath: c.Request.URL.Path, RequestParams: requestParams, ResponseStatus: code,
		})
		BadRequest(c, code, errcode.GetMessage(code))
		return
	}

	key, _ := h.keySvc.FindByRawKey(req.Key)
	keyID, keyAlias, keyTenantID := uint64(0), "", uint64(0)
	if key != nil {
		keyID, keyAlias, keyTenantID = key.ID, key.Alias, key.TenantID
	}
	h.usageLogSvc.Record(service.RecordUsageParams{
		TenantID: keyTenantID, KeyID: keyID, KeyAlias: keyAlias, Amount: req.Amount,
		IP: c.ClientIP(), UserAgent: c.GetHeader("User-Agent"),
		RequestPath: c.Request.URL.Path, RequestParams: requestParams, ResponseStatus: http.StatusOK,
	})

	Success(c, result)
}

// ========== Tenant Admin API (tenant-scoped) ==========

// TenantKeyHandler handles tenant-scoped key management endpoints.
type TenantKeyHandler struct {
	keySvc       *service.KeyService
	usageLogSvc  *service.UsageLogService
	db           *gorm.DB
	recordParams bool
}

func NewTenantKeyHandler(keySvc *service.KeyService, usageLogSvc *service.UsageLogService, db *gorm.DB, recordParams bool) *TenantKeyHandler {
	return &TenantKeyHandler{keySvc: keySvc, usageLogSvc: usageLogSvc, db: db, recordParams: recordParams}
}

type CreateKeyJSON struct {
	Alias         string  `json:"alias" binding:"required"`
	BillingMode   string  `json:"billing_mode" binding:"required"`
	InitialAmount int64   `json:"initial_amount" binding:"required"`
	ExpireAt      *string `json:"expire_at"`
	MaxUsage      *int64  `json:"max_usage"`
}

func (h *TenantKeyHandler) getTenantKeyConfig(tenantID uint64) (string, int, int, error) {
	var tenant model.Tenant
	if err := h.db.First(&tenant, tenantID).Error; err != nil {
		return "", 0, 0, fmt.Errorf("failed to get tenant key config: %w", err)
	}
	return tenant.KeyPrefix, tenant.KeyLength, tenant.KeySuffixLength, nil
}

// CreateKey creates a key scoped to the authenticated tenant.
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

// ListKeys lists keys scoped to the authenticated tenant.
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

// GetKey retrieves a key detail scoped to the authenticated tenant.
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

// UpdateKey updates a key scoped to the authenticated tenant.
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

// DisableKey disables a key scoped to the authenticated tenant.
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

// EnableKey enables a key scoped to the authenticated tenant.
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

// DeleteKey deletes a key scoped to the authenticated tenant.
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

// ExportKeys exports keys scoped to the authenticated tenant.
func (h *TenantKeyHandler) ExportKeys(c *gin.Context) {
	tenantID := getTenantID(c)

	keys, err := h.keySvc.ExportKeys(tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, keys)
}

// ExportKeysJSON exports keys as JSON scoped to the authenticated tenant.
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
