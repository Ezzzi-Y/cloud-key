package handler

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/service"

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
}

func NewTenantStatsHandler(svc *service.StatsService) *TenantStatsHandler {
	return &TenantStatsHandler{statsSvc: svc}
}

func (h *TenantStatsHandler) Dashboard(c *gin.Context) {
	tenantID := getTenantID(c)
	dash, err := h.statsSvc.GetDashboard(tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, dash)
}

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

func (h *TenantStatsHandler) TopIPs(c *gin.Context) {
	tenantID := getTenantID(c)
	dr := extractDateRange(c)
	if c.IsAborted() {
		return
	}
	items, err := h.statsSvc.GetTopIPs(dr, tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, items)
}
