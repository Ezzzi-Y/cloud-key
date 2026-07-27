package handler

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

func extractDateRange(c *gin.Context) *service.DateRange {
	start := c.Query("start_date")
	end := c.Query("end_date")
	if start == "" && end == "" {
		return nil
	}
	if start != "" && end != "" && start > end {
		BadRequest(c, errcode.CodeForbidden, "start_date 不能晚于 end_date")
		c.Abort()
		return nil
	}
	return &service.DateRange{StartDate: start, EndDate: end}
}

type TenantStatsHandler struct {
	statsSvc *service.StatsService
	keySvc   *service.KeyService
}

func NewTenantStatsHandler(svc *service.StatsService, keySvc *service.KeyService) *TenantStatsHandler {
	return &TenantStatsHandler{statsSvc: svc, keySvc: keySvc}
}

// Dashboard 租户仪表盘数据
// @Summary     仪表盘数据
// @Tags        租户-统计
// @Produce     json
// @Security    ApiKeyAuth
// @Success     200 {object} Response "仪表盘数据"
// @Router      /tenant/stats/dashboard [get]
func (h *TenantStatsHandler) Dashboard(c *gin.Context) {
	tenantID := getTenantID(c)
	dash, err := h.statsSvc.GetDashboard(tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, dash)
}

// Overview 卡密概览统计
// @Summary     卡密概览统计
// @Tags        租户-统计
// @Produce     json
// @Security    ApiKeyAuth
// @Param       start_date query string false "开始日期 YYYY-MM-DD"
// @Param       end_date   query string false "结束日期 YYYY-MM-DD"
// @Success     200 {object} Response "概览数据"
// @Failure     400 {object} Response "日期范围错误"
// @Router      /tenant/stats/overview [get]
func (h *TenantStatsHandler) Overview(c *gin.Context) {
	tenantID := getTenantID(c)
	dr := extractDateRange(c)
	if c.IsAborted() {
		return
	}
	overview, err := h.statsSvc.GetKeyOverview(dr, tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, overview)
}

// Trends 调用趋势
// @Summary     调用趋势
// @Tags        租户-统计
// @Produce     json
// @Security    ApiKeyAuth
// @Param       period     query string false "周期: today/week/month" default(today)
// @Param       start_date query string false "开始日期 YYYY-MM-DD"
// @Param       end_date   query string false "结束日期 YYYY-MM-DD"
// @Success     200 {object} Response "趋势数据点"
// @Failure     400 {object} Response "日期范围错误"
// @Router      /tenant/stats/trends [get]
func (h *TenantStatsHandler) Trends(c *gin.Context) {
	tenantID := getTenantID(c)
	period := c.DefaultQuery("period", "today")
	dr := extractDateRange(c)
	if c.IsAborted() {
		return
	}
	points, err := h.statsSvc.GetTrends(period, dr, tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, points)
}

// TopKeys 热门卡密
// @Summary     热门卡密
// @Tags        租户-统计
// @Produce     json
// @Security    ApiKeyAuth
// @Param       start_date query string false "开始日期 YYYY-MM-DD"
// @Param       end_date   query string false "结束日期 YYYY-MM-DD"
// @Success     200 {object} Response "热门卡密列表"
// @Failure     400 {object} Response "日期范围错误"
// @Router      /tenant/stats/top-keys [get]
func (h *TenantStatsHandler) TopKeys(c *gin.Context) {
	tenantID := getTenantID(c)
	dr := extractDateRange(c)
	if c.IsAborted() {
		return
	}
	items, err := h.statsSvc.GetTopKeys(dr, tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, items)
}

// TopAmount 额度消耗 Top10
// @Summary     额度消耗 Top10
// @Tags        租户-统计
// @Produce     json
// @Security    ApiKeyAuth
// @Param       start_date query string false "开始日期 YYYY-MM-DD"
// @Param       end_date   query string false "结束日期 YYYY-MM-DD"
// @Success     200 {object} Response "额度消耗 Top10 列表"
// @Failure     400 {object} Response "日期范围错误"
// @Router      /tenant/stats/top-amount [get]
func (h *TenantStatsHandler) TopAmount(c *gin.Context) {
	tenantID := getTenantID(c)
	dr := extractDateRange(c)
	if c.IsAborted() {
		return
	}
	items, err := h.statsSvc.GetTopAmount(dr, tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, items)
}

// RefreshTopStats 手动刷新 Top 统计
// @Summary     手动刷新 Top 统计
// @Tags        租户-统计
// @Produce     json
// @Security    ApiKeyAuth
// @Success     200 {object} Response "刷新成功"
// @Failure     429 {object} Response "今天已刷新过"
// @Router      /tenant/stats/refresh-top [post]
func (h *TenantStatsHandler) RefreshTopStats(c *gin.Context) {
	tenantID := getTenantID(c)

	canRefresh, err := h.keySvc.CanRefreshTop(tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	if !canRefresh {
		BadRequest(c, errcode.CodeForbidden, "24小时内已刷新过，请稍后再试")
		return
	}

	if err := h.keySvc.RefreshTopStats(tenantID); err != nil {
		if err.Error() == "already refreshed" {
			BadRequest(c, errcode.CodeForbidden, "24小时内已刷新过，请稍后再试")
			return
		}
		InternalError(c)
		return
	}
	Success(c, nil)
}

// GetKeyUsage 单个 Key 使用情况统计
// @Summary     单个 Key 使用情况
// @Tags        租户-统计
// @Produce     json
// @Security    ApiKeyAuth
// @Param       id path int true "Key ID"
// @Success     200 {object} Response "使用情况数据"
// @Failure     400 {object} Response "参数错误"
// @Router      /tenant/keys/{id}/usage [get]
func (h *TenantStatsHandler) GetKeyUsage(c *gin.Context) {
	tenantID := getTenantID(c)
	keyID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, errcode.CodeForbidden, "无效的 Key ID")
		return
	}

	stats, err := h.statsSvc.GetKeyUsage(tenantID, keyID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, stats)
}

// RefreshKeyUsage 手动刷新单个 Key 使用统计
// @Summary     刷新单个 Key 使用统计
// @Tags        租户-统计
// @Produce     json
// @Security    ApiKeyAuth
// @Param       id path int true "Key ID"
// @Success     200 {object} Response "刷新成功"
// @Failure     400 {object} Response "参数错误或今天已刷新过"
// @Router      /tenant/keys/{id}/usage/refresh [post]
func (h *TenantStatsHandler) RefreshKeyUsage(c *gin.Context) {
	tenantID := getTenantID(c)
	keyID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, errcode.CodeForbidden, "无效的 Key ID")
		return
	}

	if err := h.statsSvc.RefreshKeyUsage(tenantID, keyID); err != nil {
		if err.Error() == "already refreshed" {
			BadRequest(c, errcode.CodeForbidden, "24小时内已刷新过，请稍后再试")
			return
		}
		InternalError(c)
		return
	}
	Success(c, nil)
}
