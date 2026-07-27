package handler

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"CloudKey/internal/service"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ========== Tenant Admin API (tenant-scoped) ==========

// TenantKeyHandler handles tenant-scoped key management endpoints.
type TenantKeyHandler struct {
	keySvc        *service.KeyService
	balanceLogSvc *service.BalanceLogService
	db            *gorm.DB
	rdb           *redis.Client
}

func NewTenantKeyHandler(keySvc *service.KeyService, balanceLogSvc *service.BalanceLogService, db *gorm.DB, rdb *redis.Client) *TenantKeyHandler {
	return &TenantKeyHandler{keySvc: keySvc, balanceLogSvc: balanceLogSvc, db: db, rdb: rdb}
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
		BadRequest(c, errcode.CodeInvalidConsumeAmount, errcode.GetMessage(errcode.CodeInvalidConsumeAmount))
		return
	}

	// 幂等检查
	requestID := c.GetString("request_id")
	if claimed, _ := IdempotentCheck(c, h.rdb, requestID); !claimed {
		return // 重复请求，已返回缓存
	}

	result, code, err := h.keySvc.ConsumeKeyByTenant(req.Key, req.Amount, tenantID, &service.ConsumeMeta{
		RequestID: requestID,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		// 系统错误不缓存（可重试）
		CacheIdempotentError(h.rdb, requestID, http.StatusInternalServerError, errcode.CodeInternalError)
		InternalError(c)
		return
	}
	if code != 0 {
		// 业务错误也缓存（防止重试时重复执行）
		CacheIdempotentError(h.rdb, requestID, http.StatusBadRequest, code)
		BadRequest(c, code, errcode.GetMessage(code))
		return
	}

	// 缓存成功结果
	CacheIdempotentResult(h.rdb, requestID, http.StatusOK, errcode.CodeSuccess, result)
	Success(c, result)
}

type CreateKeyJSON struct {
	Alias           string `json:"alias" binding:"required" example:"测试卡密"`
	RemainingAmount int64  `json:"remaining_amount" binding:"required" example:"100"`
	RateLimit       *int   `json:"rate_limit"`        // nil=使用租户默认，0=不限速
	RateLimitWindow *int   `json:"rate_limit_window"` // 窗口大小（秒）
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

	// 限流配置校验
	if req.RateLimit != nil || req.RateLimitWindow != nil {
		if req.RateLimit == nil || req.RateLimitWindow == nil {
			BadRequest(c, errcode.CodeKeyConfigInvalid, "限流次数和窗口大小必须同时设置")
			return
		}
		if *req.RateLimit < 0 {
			BadRequest(c, errcode.CodeKeyConfigInvalid, "限流次数不能为负数")
			return
		}
		if *req.RateLimitWindow < 1 {
			BadRequest(c, errcode.CodeKeyConfigInvalid, "限流窗口大小必须大于 0")
			return
		}
	}

	tenantID := getTenantID(c)

	keyPrefix, keyLen, suffixLen, err := h.getTenantKeyConfig(tenantID)
	if err != nil {
		InternalError(c)
		return
	}

	createdBy := "tenant_admin"

	result, err := h.keySvc.CreateKey(service.CreateKeyRequest{
		Alias: req.Alias, RemainingAmount: req.RemainingAmount, CreatedBy: createdBy,
		RateLimit: req.RateLimit, RateLimitWindow: req.RateLimitWindow,
	}, tenantID, keyPrefix, keyLen, suffixLen)
	if err != nil {
		InternalError(c)
		return
	}

	// 创建时有初始额度，记录到流转日志
	if result.Key.RemainingAmount > 0 {
		_ = h.balanceLogSvc.Record(service.RecordBalanceParams{
			TenantID:     tenantID,
			KeyID:        result.Key.ID,
			KeyAlias:     result.Key.Alias,
			KeySuffix:    result.Key.KeySuffix,
			Delta:        result.Key.RemainingAmount,
			BeforeAmount: 0,
			AfterAmount:  result.Key.RemainingAmount,
			Operator:     createdBy,
			Remark:       "创建卡密初始额度",
		})
	}

	Success(c, gin.H{
		"id": result.Key.ID, "raw_key": result.RawKey, "alias": result.Key.Alias,
		"key_prefix": result.Key.KeyPrefix, "key_suffix": result.Key.KeySuffix,
		"remaining_amount": result.Key.RemainingAmount, "status": result.Key.Status,
		"created_by": result.Key.CreatedBy, "created_at": result.Key.CreatedAt,
	})
}

// ListKeys 卡密列表
// @Summary     卡密列表
// @Tags        租户-卡密管理
// @Produce     json
// @Security    ApiKeyAuth
// @Param       page       query int    false "页码"     default(1)
// @Param       page_size  query int    false "每页数量" default(20)
// @Param       status     query string false "状态过滤: active/exhausted/disabled/expired"
// @Param       alias      query string false "别名前缀搜索"
// @Param       key_suffix query string false "后缀精准搜索"
// @Success     200 {object} Response{data=PageData} "分页卡密列表"
// @Router      /tenant/keys [get]
func (h *TenantKeyHandler) ListKeys(c *gin.Context) {
	tenantID := getTenantID(c)
	page, pageSize := pageParams(c)

	keys, total, err := h.keySvc.ListKeys(service.KeyListQuery{
		Page: page, PageSize: pageSize,
		Status: c.Query("status"), Alias: c.Query("alias"), KeySuffix: c.Query("key_suffix"),
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

	// 限流配置校验：两个字段必须同时设置或同时为 nil
	if req.RateLimit != nil || req.RateLimitWindow != nil {
		if req.RateLimit == nil || req.RateLimitWindow == nil {
			BadRequest(c, errcode.CodeKeyConfigInvalid, "限流次数和窗口大小必须同时设置")
			return
		}
		if *req.RateLimit < 0 {
			BadRequest(c, errcode.CodeKeyConfigInvalid, "限流次数不能为负数")
			return
		}
		if *req.RateLimitWindow < 1 {
			BadRequest(c, errcode.CodeKeyConfigInvalid, "限流窗口大小必须大于 0")
			return
		}
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

	// 幂等检查
	requestID := c.GetString("request_id")
	if claimed, _ := IdempotentCheck(c, h.rdb, requestID); !claimed {
		return
	}

	result, err := h.keySvc.AdjustBalance(id, tenantID, service.AdjustBalanceRequest{
		Delta:     req.Delta,
		Operator:  "tenant_admin",
		Remark:    req.Remark,
		RequestID: requestID,
	})
	if err != nil {
		CacheIdempotentError(h.rdb, requestID, http.StatusBadRequest, errcode.CodeInvalidAdjustment)
		BadRequest(c, errcode.CodeInvalidAdjustment, err.Error())
		return
	}

	CacheIdempotentResult(h.rdb, requestID, http.StatusOK, errcode.CodeSuccess, result)
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

// ========== Key Config API ==========

// GetKeyConfig 获取当前租户的 Key 配置
// @Summary     获取 Key 配置
// @Description 租户管理员获取当前租户的卡密生成配置
// @Tags        租户-配置
// @Produce     json
// @Security    ApiKeyAuth
// @Success     200 {object} Response "Key 配置信息"
// @Failure     404 {object} Response "租户不存在"
// @Router      /tenant/key-config [get]
func (h *TenantKeyHandler) GetKeyConfig(c *gin.Context) {
	tenantID := getTenantID(c)

	var tenant model.Tenant
	if err := h.db.First(&tenant, tenantID).Error; err != nil {
		NotFound(c, errcode.CodeTenantNotFound, errcode.GetMessage(errcode.CodeTenantNotFound))
		return
	}

	Success(c, gin.H{
		"key_prefix":                 tenant.KeyPrefix,
		"key_length":                 tenant.KeyLength,
		"key_suffix_length":          tenant.KeySuffixLength,
		"default_rate_limit":         tenant.DefaultRateLimit,
		"default_rate_limit_window":  tenant.DefaultRateLimitWindow,
	})
}

var keyPrefixRegexp = regexp.MustCompile(`^[a-zA-Z0-9_]+-$`)

type UpdateKeyConfigRequest struct {
	KeyPrefix               *string `json:"key_prefix"`
	KeyLength               *int    `json:"key_length"`
	KeySuffixLength         *int    `json:"key_suffix_length"`
	DefaultRateLimit        *int    `json:"default_rate_limit"`        // nil=不更新，0=不限速
	DefaultRateLimitWindow  *int    `json:"default_rate_limit_window"` // nil=不更新
}

// UpdateKeyConfig 更新当前租户的 Key 配置
// @Summary     更新 Key 配置
// @Description 租户管理员更新当前租户的卡密生成配置
// @Tags        租户-配置
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       body body UpdateKeyConfigRequest true "配置参数"
// @Success     200 {object} Response "更新成功"
// @Failure     400 {object} Response "参数校验失败"
// @Router      /tenant/key-config [patch]
func (h *TenantKeyHandler) UpdateKeyConfig(c *gin.Context) {
	tenantID := getTenantID(c)

	var req UpdateKeyConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeKeyConfigInvalid, "参数错误")
		return
	}

	// 读取当前配置作为基础
	var tenant model.Tenant
	if err := h.db.First(&tenant, tenantID).Error; err != nil {
		NotFound(c, errcode.CodeTenantNotFound, errcode.GetMessage(errcode.CodeTenantNotFound))
		return
	}

	keyLen := tenant.KeyLength
	suffixLen := tenant.KeySuffixLength
	prefix := tenant.KeyPrefix

	if req.KeyPrefix != nil {
		p := *req.KeyPrefix
		if len(p) < 1 || len(p) > 10 {
			BadRequest(c, errcode.CodeKeyConfigInvalid, "Key 前缀长度需在 1~10 之间")
			return
		}
		if !keyPrefixRegexp.MatchString(p) {
			BadRequest(c, errcode.CodeKeyConfigInvalid, "Key 前缀必须以 - 结尾，仅允许字母、数字、下划线")
			return
		}
		prefix = p
	}
	if req.KeyLength != nil {
		v := *req.KeyLength
		if v < 8 || v > 32 {
			BadRequest(c, errcode.CodeKeyConfigInvalid, "Key 长度需在 8~32 之间")
			return
		}
		keyLen = v
	}
	if req.KeySuffixLength != nil {
		v := *req.KeySuffixLength
		if v < 4 {
			BadRequest(c, errcode.CodeKeyConfigInvalid, "后缀显示长度不能小于 4")
			return
		}
		suffixLen = v
	}
	if suffixLen > keyLen {
		BadRequest(c, errcode.CodeKeyConfigInvalid, "后缀显示长度不能超过 Key 长度")
		return
	}

	// 限流配置校验：两个字段必须同时设置或同时为 nil
	var defaultRateLimit, defaultRateLimitWindow *int
	if req.DefaultRateLimit != nil || req.DefaultRateLimitWindow != nil {
		if req.DefaultRateLimit == nil || req.DefaultRateLimitWindow == nil {
			BadRequest(c, errcode.CodeKeyConfigInvalid, "限流次数和窗口大小必须同时设置")
			return
		}
		if *req.DefaultRateLimit < 0 {
			BadRequest(c, errcode.CodeKeyConfigInvalid, "限流次数不能为负数")
			return
		}
		if *req.DefaultRateLimitWindow < 1 {
			BadRequest(c, errcode.CodeKeyConfigInvalid, "限流窗口大小必须大于 0")
			return
		}
		defaultRateLimit = req.DefaultRateLimit
		defaultRateLimitWindow = req.DefaultRateLimitWindow
	}

	updates := map[string]interface{}{
		"key_prefix":        prefix,
		"key_length":        keyLen,
		"key_suffix_length": suffixLen,
	}
	if defaultRateLimit != nil {
		updates["default_rate_limit"] = *defaultRateLimit
		updates["default_rate_limit_window"] = *defaultRateLimitWindow
	}
	if err := h.db.Model(&model.Tenant{}).Where("id = ?", tenantID).Updates(updates).Error; err != nil {
		InternalError(c)
		return
	}

	// 租户默认限流变更后，清除该租户下所有 key 的缓存，使新配置生效
	if defaultRateLimit != nil {
		h.keySvc.InvalidateTenantKeyCaches(tenantID)
	}

	Success(c, gin.H{
		"key_prefix":                 prefix,
		"key_length":                 keyLen,
		"key_suffix_length":          suffixLen,
		"default_rate_limit":         tenant.DefaultRateLimit,
		"default_rate_limit_window":  tenant.DefaultRateLimitWindow,
	})
}

// GetConsumeResult 根据 request_id 查询消费/调额结果
// @Summary     根据请求ID查询操作结果
// @Description 租户管理员根据 request_id 查询消费或调额的结果
// @Tags        租户-卡密管理
// @Produce     json
// @Security    ApiKeyAuth
// @Param       request_id query string true "请求ID"
// @Success     200 {object} Response "操作结果"
// @Failure     400 {object} Response "缺少 request_id 参数"
// @Failure     404 {object} Response "未找到对应的操作记录"
// @Router      /tenant/keys/consume-result [get]
func (h *TenantKeyHandler) GetConsumeResult(c *gin.Context) {
	tenantID := getTenantID(c)

	requestID := c.Query("request_id")
	if requestID == "" {
		BadRequest(c, http.StatusBadRequest, "缺少 request_id 参数")
		return
	}

	// 先查 Redis 缓存
	cached, err := LookupIdempotentResult(h.rdb, requestID)
	if err == nil && cached != nil {
		Success(c, gin.H{
			"source":      "cache",
			"request_id":  requestID,
			"http_status": cached.HTTPStatus,
			"code":        cached.Code,
			"data":        cached.Data,
		})
		return
	}

	// fallback 查询 usage_logs
	var usageLog model.UsageLog
	if err := h.db.Where("request_id = ? AND tenant_id = ?", requestID, tenantID).
		Order("created_at DESC").First(&usageLog).Error; err == nil {
		Success(c, gin.H{
			"source":     "usage_log",
			"request_id": requestID,
			"key_id":     usageLog.KeyID,
			"key_alias":  usageLog.KeyAlias,
			"key_suffix": usageLog.KeySuffix,
			"amount":     usageLog.Amount,
			"ip":         usageLog.IP,
			"created_at": usageLog.CreatedAt,
		})
		return
	}

	// fallback 查询 balance_logs
	var balanceLog model.BalanceLog
	if err := h.db.Where("request_id = ? AND tenant_id = ?", requestID, tenantID).
		Order("created_at DESC").First(&balanceLog).Error; err == nil {
		Success(c, gin.H{
			"source":        "balance_log",
			"request_id":    requestID,
			"key_id":        balanceLog.KeyID,
			"key_alias":     balanceLog.KeyAlias,
			"key_suffix":    balanceLog.KeySuffix,
			"delta":         balanceLog.Delta,
			"before_amount": balanceLog.BeforeAmount,
			"after_amount":  balanceLog.AfterAmount,
			"operator":      balanceLog.Operator,
			"remark":        balanceLog.Remark,
			"created_at":    balanceLog.CreatedAt,
		})
		return
	}

	NotFound(c, http.StatusNotFound, "未找到该 request_id 对应的操作记录")
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
