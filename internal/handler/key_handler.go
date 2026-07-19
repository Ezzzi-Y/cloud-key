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

// Status 查询卡密状态（不扣减）
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

// Consume 扣减卡密额度
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
		keyID, keyAlias := uint64(0), ""
		if key != nil {
			keyID, keyAlias = key.ID, key.Alias
		}
		h.usageLogSvc.Record(service.RecordUsageParams{
			KeyID: keyID, KeyAlias: keyAlias, Amount: req.Amount,
			IP: c.ClientIP(), UserAgent: c.GetHeader("User-Agent"),
			RequestPath: c.Request.URL.Path, RequestParams: requestParams, ResponseStatus: code,
		})
		BadRequest(c, code, errcode.GetMessage(code))
		return
	}

	key, _ := h.keySvc.FindByRawKey(req.Key)
	keyID, keyAlias := uint64(0), ""
	if key != nil {
		keyID, keyAlias = key.ID, key.Alias
	}
	h.usageLogSvc.Record(service.RecordUsageParams{
		KeyID: keyID, KeyAlias: keyAlias, Amount: req.Amount,
		IP: c.ClientIP(), UserAgent: c.GetHeader("User-Agent"),
		RequestPath: c.Request.URL.Path, RequestParams: requestParams, ResponseStatus: http.StatusOK,
	})

	Success(c, result)
}

// ========== 管理员接口 ==========

type CreateKeyJSON struct {
	Alias         string  `json:"alias" binding:"required"`
	BillingMode   string  `json:"billing_mode" binding:"required"`
	InitialAmount int64   `json:"initial_amount" binding:"required"`
	ExpireAt      *string `json:"expire_at"`
	MaxUsage      *int64  `json:"max_usage"`
}

// CreateKey 管理员创建卡密
func (h *KeyHandler) CreateKey(c *gin.Context) {
	var req CreateKeyJSON
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, http.StatusBadRequest, "参数错误")
		return
	}

	adminID, _ := c.Get("admin_id")
	createdBy := ""
	if adminID != nil {
		createdBy = "admin"
	}

	expireAt, err := parseExpireAt(req.ExpireAt)
	if err != nil {
		BadRequest(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.keySvc.CreateKey(service.CreateKeyRequest{
		Alias: req.Alias, BillingMode: model.KeyBillingMode(req.BillingMode),
		InitialAmount: req.InitialAmount, CreatedBy: createdBy,
		ExpireAt: expireAt, MaxUsage: req.MaxUsage,
	})
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

// ListKeys 管理员查询卡密列表
func (h *KeyHandler) ListKeys(c *gin.Context) {
	page, pageSize := pageParams(c)
	keys, total, err := h.keySvc.ListKeys(service.KeyListQuery{
		Page: page, PageSize: pageSize,
		Status: c.Query("status"), Search: c.Query("search"),
	})
	if err != nil {
		InternalError(c)
		return
	}
	SuccessPaginated(c, keys, total, page, pageSize)
}

// GetKey 管理员查看卡密详情
func (h *KeyHandler) GetKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, http.StatusBadRequest, "无效的卡密 ID")
		return
	}

	key, err := h.keySvc.GetKeyDetail(id)
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

// UpdateKey 管理员修改卡密
func (h *KeyHandler) UpdateKey(c *gin.Context) {
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

	if err := h.keySvc.UpdateKey(id, req); err != nil {
		InternalError(c)
		return
	}
	Success(c, nil)
}

// DisableKey 管理员禁用卡密
func (h *KeyHandler) DisableKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, http.StatusBadRequest, "无效的卡密 ID")
		return
	}
	if err := h.keySvc.DisableKey(id); err != nil {
		InternalError(c)
		return
	}
	Success(c, nil)
}

// EnableKey 管理员启用卡密
func (h *KeyHandler) EnableKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, http.StatusBadRequest, "无效的卡密 ID")
		return
	}
	if err := h.keySvc.EnableKey(id); err != nil {
		InternalError(c)
		return
	}
	Success(c, nil)
}

// DeleteKey 管理员删除卡密
func (h *KeyHandler) DeleteKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, http.StatusBadRequest, "无效的卡密 ID")
		return
	}
	if err := h.keySvc.DeleteKey(id); err != nil {
		InternalError(c)
		return
	}
	Success(c, nil)
}

// ExportKeys 管理员导出卡密
func (h *KeyHandler) ExportKeys(c *gin.Context) {
	keys, err := h.keySvc.ExportKeys()
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, keys)
}

// ExportKeysJSON 管理员导出卡密（JSON 格式）
func (h *KeyHandler) ExportKeysJSON(c *gin.Context) {
	items, err := h.keySvc.ExportKeysJSON()
	if err != nil {
		InternalError(c)
		return
	}
	if items == nil {
		items = make([]service.ExportKeyItem, 0)
	}
	Success(c, items)
}

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
