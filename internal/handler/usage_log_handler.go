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

// ListLogs 使用日志列表
// @Summary     使用日志列表
// @Tags        租户-使用日志
// @Produce     json
// @Security    ApiKeyAuth
// @Param       page       query int    false "页码"       default(1)
// @Param       page_size  query int    false "每页数量"   default(20)
// @Param       alias      query string false "别名前缀搜索"
// @Param       key_suffix query string false "后缀精准搜索"
// @Param       start_time query string false "开始时间"
// @Param       end_time   query string false "结束时间"
// @Success     200 {object} Response{data=PageData} "分页使用日志"
// @Router      /tenant/usage-logs [get]
func (h *TenantUsageLogHandler) ListLogs(c *gin.Context) {
	tenantID := getTenantID(c)
	page, pageSize := pageParams(c)

	logs, total, err := h.usageLogSvc.ListLogs(service.UsageLogQuery{
		Page: page, PageSize: pageSize,
		Alias: c.Query("alias"), KeySuffix: c.Query("key_suffix"),
		StartTime: c.Query("start_time"), EndTime: c.Query("end_time"),
	}, tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	SuccessPaginated(c, logs, total, page, pageSize)
}

// ExportLogs 导出使用日志
// @Summary     导出使用日志
// @Tags        租户-使用日志
// @Produce     json
// @Security    ApiKeyAuth
// @Param       alias      query string false "别名前缀搜索"
// @Param       key_suffix query string false "后缀精准搜索"
// @Param       start_time query string false "开始时间"
// @Param       end_time   query string false "结束时间"
// @Success     200 {object} Response "使用日志列表"
// @Router      /tenant/usage-logs/export [get]
func (h *TenantUsageLogHandler) ExportLogs(c *gin.Context) {
	tenantID := getTenantID(c)
	logs, err := h.usageLogSvc.ExportLogs(service.UsageLogQuery{
		Alias: c.Query("alias"), KeySuffix: c.Query("key_suffix"),
		StartTime: c.Query("start_time"), EndTime: c.Query("end_time"),
	}, tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, logs)
}

// LoginLogs 登录日志（当前租户）
// @Summary     登录日志
// @Tags        租户-使用日志
// @Produce     json
// @Security    ApiKeyAuth
// @Param       page      query int false "页码"     default(1)
// @Param       page_size query int false "每页数量" default(20)
// @Success     200 {object} Response{data=PageData} "分页登录日志"
// @Router      /tenant/login-logs [get]
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
