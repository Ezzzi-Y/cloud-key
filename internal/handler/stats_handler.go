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
	// Validate: if both provided, start must be <= end
	if start != "" && end != "" && start > end {
		BadRequest(c, errcode.CodeForbidden, "start_date 不能晚于 end_date")
		c.Abort()
		return nil
	}
	return &service.DateRange{StartDate: start, EndDate: end}
}

type StatsHandler struct {
	statsSvc *service.StatsService
}

func NewStatsHandler(svc *service.StatsService) *StatsHandler {
	return &StatsHandler{statsSvc: svc}
}

func (h *StatsHandler) Dashboard(c *gin.Context) {
	dash, err := h.statsSvc.GetDashboard()
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, dash)
}

func (h *StatsHandler) Overview(c *gin.Context) {
	dr := extractDateRange(c)
	if c.IsAborted() {
		return
	}
	overview, err := h.statsSvc.GetKeyOverview(dr)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, overview)
}

func (h *StatsHandler) Trends(c *gin.Context) {
	period := c.DefaultQuery("period", "today")
	dr := extractDateRange(c)
	if c.IsAborted() {
		return
	}
	points, err := h.statsSvc.GetTrends(period, dr)
	if err != nil {
		InternalError(c)
		return
	}
	if points == nil {
		points = make([]service.TrendPoint, 0)
	}
	Success(c, points)
}

func (h *StatsHandler) TopKeys(c *gin.Context) {
	dr := extractDateRange(c)
	if c.IsAborted() {
		return
	}
	items, err := h.statsSvc.GetTopKeys(dr)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, items)
}

func (h *StatsHandler) TopIPs(c *gin.Context) {
	dr := extractDateRange(c)
	if c.IsAborted() {
		return
	}
	items, err := h.statsSvc.GetTopIPs(dr)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, items)
}
