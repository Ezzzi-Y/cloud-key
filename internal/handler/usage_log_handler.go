package handler

import (
	"CloudKey/internal/service"

	"github.com/gin-gonic/gin"
)

type UsageLogHandler struct {
	usageLogSvc *service.UsageLogService
}

func NewUsageLogHandler(svc *service.UsageLogService) *UsageLogHandler {
	return &UsageLogHandler{usageLogSvc: svc}
}

func (h *UsageLogHandler) ListLogs(c *gin.Context) {
	page, pageSize := pageParams(c)

	logs, total, err := h.usageLogSvc.ListLogs(service.UsageLogQuery{
		Page: page, PageSize: pageSize,
		KeyAlias: c.Query("key_alias"), IP: c.Query("ip"),
		StartTime: c.Query("start_time"), EndTime: c.Query("end_time"),
	})
	if err != nil {
		InternalError(c)
		return
	}
	SuccessPaginated(c, logs, total, page, pageSize)
}

func (h *UsageLogHandler) ExportLogs(c *gin.Context) {
	logs, err := h.usageLogSvc.ExportLogs(service.UsageLogQuery{
		KeyAlias: c.Query("key_alias"), IP: c.Query("ip"),
		StartTime: c.Query("start_time"), EndTime: c.Query("end_time"),
	})
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, logs)
}
