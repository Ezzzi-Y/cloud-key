package handler

import (
	"CloudKey/internal/service"

	"github.com/gin-gonic/gin"
)

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
	overview, err := h.statsSvc.GetKeyOverview()
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, overview)
}

func (h *StatsHandler) Trends(c *gin.Context) {
	period := c.DefaultQuery("period", "today")
	points, err := h.statsSvc.GetTrends(period)
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
	items, err := h.statsSvc.GetTopKeys()
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, items)
}

func (h *StatsHandler) TopIPs(c *gin.Context) {
	items, err := h.statsSvc.GetTopIPs()
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, items)
}
