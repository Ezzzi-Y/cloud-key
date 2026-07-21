package handler

import (
	"CloudKey/internal/model"
	"CloudKey/internal/service"
	"fmt"

	"github.com/gin-gonic/gin"
)

// ========== Tenant Balance Log ==========

type TenantBalanceLogHandler struct {
	balanceLogSvc *service.BalanceLogService
}

func NewTenantBalanceLogHandler(svc *service.BalanceLogService) *TenantBalanceLogHandler {
	return &TenantBalanceLogHandler{balanceLogSvc: svc}
}

// ListLogs 额度流转日志
// @Summary     额度流转日志
// @Tags        租户-额度流转日志
// @Produce     json
// @Security    ApiKeyAuth
// @Param       page       query int    false "页码"     default(1)
// @Param       page_size  query int    false "每页数量" default(20)
// @Param       key_id     query int    false "卡密ID过滤"
// @Param       operator   query string false "操作人过滤"
// @Param       start_time query string false "开始时间"
// @Param       end_time   query string false "结束时间"
// @Success     200 {object} Response{data=PageData} "分页流转日志"
// @Router      /tenant/balance-logs [get]
func (h *TenantBalanceLogHandler) ListLogs(c *gin.Context) {
	tenantID := getTenantID(c)
	page, pageSize := pageParams(c)

	var query service.BalanceLogQuery
	query.Page = page
	query.PageSize = pageSize
	query.KeyID, _ = parseUint64(c.Query("key_id"))
	query.Operator = c.Query("operator")
	query.StartTime = c.Query("start_time")
	query.EndTime = c.Query("end_time")

	logs, total, err := h.balanceLogSvc.ListLogs(query, tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	SuccessPaginated(c, logs, total, page, pageSize)
}

// ExportLogs 导出额度流转日志
// @Summary     导出额度流转日志
// @Tags        租户-额度流转日志
// @Produce     json
// @Security    ApiKeyAuth
// @Param       key_id     query int    false "卡密ID过滤"
// @Param       operator   query string false "操作人过滤"
// @Param       start_time query string false "开始时间"
// @Param       end_time   query string false "结束时间"
// @Success     200 {object} Response "导出数据"
// @Router      /tenant/balance-logs/export [get]
func (h *TenantBalanceLogHandler) ExportLogs(c *gin.Context) {
	tenantID := getTenantID(c)

	var query service.BalanceLogQuery
	query.KeyID, _ = parseUint64(c.Query("key_id"))
	query.Operator = c.Query("operator")
	query.StartTime = c.Query("start_time")
	query.EndTime = c.Query("end_time")

	logs, err := h.balanceLogSvc.ExportLogs(query, tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, logs)
}

// ========== Service Account Balance Log ==========

type ServiceBalanceLogHandler struct {
	balanceLogSvc *service.BalanceLogService
}

func NewServiceBalanceLogHandler(svc *service.BalanceLogService) *ServiceBalanceLogHandler {
	return &ServiceBalanceLogHandler{balanceLogSvc: svc}
}

// ServiceListLogs 服务账号查询额度流转日志
// @Summary     服务账号查询额度流转日志
// @Tags        服务账号API
// @Produce     json
// @Security    ServiceKeyAuth
// @Param       page       query int    false "页码"     default(1)
// @Param       page_size  query int    false "每页数量" default(20)
// @Param       key_id     query int    false "卡密ID过滤"
// @Param       operator   query string false "操作人过滤"
// @Param       start_time query string false "开始时间"
// @Param       end_time   query string false "结束时间"
// @Success     200 {object} Response{data=PageData} "分页流转日志"
// @Failure     401 {object} Response "服务账号密钥无效"
// @Router      /service/balance-logs [get]
func (h *ServiceBalanceLogHandler) ServiceListLogs(c *gin.Context) {
	tenantID, ok := getServiceTenantID(c)
	if !ok {
		return
	}
	page, pageSize := pageParams(c)

	var query service.BalanceLogQuery
	query.Page = page
	query.PageSize = pageSize
	query.KeyID, _ = parseUint64(c.Query("key_id"))
	query.Operator = c.Query("operator")
	query.StartTime = c.Query("start_time")
	query.EndTime = c.Query("end_time")

	logs, total, err := h.balanceLogSvc.ListLogs(query, tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	SuccessPaginated(c, logs, total, page, pageSize)
}

// ServiceExportLogs 服务账号导出额度流转日志
// @Summary     服务账号导出额度流转日志
// @Tags        服务账号API
// @Produce     json
// @Security    ServiceKeyAuth
// @Param       key_id     query int    false "卡密ID过滤"
// @Param       operator   query string false "操作人过滤"
// @Param       start_time query string false "开始时间"
// @Param       end_time   query string false "结束时间"
// @Success     200 {object} Response "导出数据"
// @Failure     401 {object} Response "服务账号密钥无效"
// @Router      /service/balance-logs/export [get]
func (h *ServiceBalanceLogHandler) ServiceExportLogs(c *gin.Context) {
	tenantID, ok := getServiceTenantID(c)
	if !ok {
		return
	}

	var query service.BalanceLogQuery
	query.KeyID, _ = parseUint64(c.Query("key_id"))
	query.Operator = c.Query("operator")
	query.StartTime = c.Query("start_time")
	query.EndTime = c.Query("end_time")

	logs, err := h.balanceLogSvc.ExportLogs(query, tenantID)
	if err != nil {
		InternalError(c)
		return
	}
	if logs == nil {
		logs = make([]model.BalanceLog, 0)
	}
	Success(c, logs)
}

// helper
func parseUint64(s string) (uint64, error) {
	if s == "" {
		return 0, nil
	}
	var v uint64
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}
