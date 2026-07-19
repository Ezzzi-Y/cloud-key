package handler

import (
	"CloudKey/internal/service"

	"github.com/gin-gonic/gin"
)

type TenantUsageLogHandler struct {
	usageLogSvc *service.UsageLogService
	loginLogSvc *service.LoginLogService
}

func NewTenantUsageLogHandler(usageLogSvc *service.UsageLogService, loginLogSvc *service.LoginLogService) *TenantUsageLogHandler {
	return &TenantUsageLogHandler{usageLogSvc: usageLogSvc, loginLogSvc: loginLogSvc}
}

func (h *TenantUsageLogHandler) ListLogs(c *gin.Context) {
	tenantID := getTenantID(c)
	page, pageSize := pageParams(c)

	logs, total, err := h.usageLogSvc.ListLogs(service.UsageLogQuery{
		Page: page, PageSize: pageSize,
		KeyAlias: c.Query("key_alias"), IP: c.Query("ip"),
		StartTime: c.Query("start_time"), EndTime: c.Query("end_time"),
	}, tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	SuccessPaginated(c, logs, total, page, pageSize)
}

func (h *TenantUsageLogHandler) ExportLogs(c *gin.Context) {
	tenantID := getTenantID(c)
	logs, err := h.usageLogSvc.ExportLogs(service.UsageLogQuery{
		KeyAlias: c.Query("key_alias"), IP: c.Query("ip"),
		StartTime: c.Query("start_time"), EndTime: c.Query("end_time"),
	}, tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, logs)
}

func (h *TenantUsageLogHandler) LoginLogs(c *gin.Context) {
	page, pageSize := pageParams(c)
	tenantID := getTenantID(c)
	var tenantIDPtr *uint64
	if tenantID > 0 {
		tenantIDPtr = &tenantID
	}
	logs, total, err := h.loginLogSvc.ListLoginLogs(page, pageSize, tenantIDPtr)
	if err != nil {
		InternalError(c)
		return
	}
	SuccessPaginated(c, logs, total, page, pageSize)
}
