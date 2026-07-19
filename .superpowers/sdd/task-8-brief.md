### Task 8: 改造 ServiceHandler + StatsHandler + UsageLogHandler（面向 tenant）

**Files:**
- Rewrite: `internal/handler/service_handler.go`
- Rewrite: `internal/handler/stats_handler.go`
- Rewrite: `internal/handler/usage_log_handler.go`
- Rewrite: `internal/handler/config_handler.go`

- [ ] **Step 1: 重写 `service_handler.go` — 面向 tenant admin**

改造所有方法加 tenant scope:

```go
type TenantServiceAccountHandler struct {
	keySvc            *service.KeyService
	serviceAccountSvc *service.ServiceAccountService
}

func NewTenantServiceAccountHandler(keySvc *service.KeyService, saSvc *service.ServiceAccountService) *TenantServiceAccountHandler {
	return &TenantServiceAccountHandler{keySvc: keySvc, serviceAccountSvc: saSvc}
}

// ListServiceAccounts
func (h *TenantServiceAccountHandler) ListServiceAccounts(c *gin.Context) {
	tenantID := getTenantID(c)
	accounts, err := h.serviceAccountSvc.ListServiceAccounts(tenantID)
	// ...
}

// CreateServiceAccount
func (h *TenantServiceAccountHandler) CreateServiceAccount(c *gin.Context) {
	var req struct { Name string `json:"name" binding:"required"` }
	// ...
	tenantID := getTenantID(c)
	account, rawKey, err := h.serviceAccountSvc.CreateServiceAccount(req.Name, tenantID)
	// ...
}

// ToggleServiceAccount, DeleteServiceAccount — 类似加 tenantID
```

Service account 的 key create/list 方法 **ServiceCreateKey, ServiceListKeys** — 保留，使用 `c.Get("service_account")` 获取 sa，从中取 `sa.TenantID`:

```go
func (h *TenantServiceAccountHandler) ServiceCreateKey(c *gin.Context) {
	saI, _ := c.Get("service_account")
	sa := saI.(*model.ServiceAccount)
	tenantID := sa.TenantID
	// ...
	result, err := h.keySvc.CreateKey(req, tenantID, "sk-", 32, 4)
	// ...
}

func (h *TenantServiceAccountHandler) ServiceListKeys(c *gin.Context) {
	saI, _ := c.Get("service_account")
	sa := saI.(*model.ServiceAccount)
	tenantID := sa.TenantID
	// 使用 tenantID 过滤 keys
	keys, total, err := h.keySvc.ListKeysByTenant(tenantID, page, pageSize)
	// ...
}
```

需要在 KeyService 新增 `ListKeysByTenant`:
```go
func (s *KeyService) ListKeysByTenant(tenantID uint64, page, pageSize int) ([]model.Key, int64, error) {
	var keys []model.Key
	var total int64
	db := s.db.Model(&model.Key{}).Where("tenant_id = ?", tenantID)
	db.Count(&total)
	offset := (page - 1) * pageSize
	db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&keys)
	return keys, total, nil
}
```

- [ ] **Step 2: 重写 `stats_handler.go` — 面向 tenant admin**

```go
type TenantStatsHandler struct {
	statsSvc *service.StatsService
}

func NewTenantStatsHandler(svc *service.StatsService) *TenantStatsHandler {
	return &TenantStatsHandler{statsSvc: svc}
}

func (h *TenantStatsHandler) Dashboard(c *gin.Context) {
	tenantID := getTenantID(c)
	dash, err := h.statsSvc.GetDashboard(tenantID)
	// ...
}

// Overview, Trends, TopKeys, TopIPs — 全部加 tenantID
```

- [ ] **Step 3: 重写 `usage_log_handler.go` — 面向 tenant admin**

```go
type TenantUsageLogHandler struct {
	usageLogSvc *service.UsageLogService
}

func (h *TenantUsageLogHandler) ListLogs(c *gin.Context) {
	tenantID := getTenantID(c)
	page, pageSize := pageParams(c)
	logs, total, err := h.usageLogSvc.ListLogs(service.UsageLogQuery{...}, tenantID)
	// ...
}

func (h *TenantUsageLogHandler) ExportLogs(c *gin.Context) {
	tenantID := getTenantID(c)
	logs, err := h.usageLogSvc.ExportLogs(service.UsageLogQuery{...}, tenantID)
	// ...
}
```

- [ ] **Step 4: 保留 config_handler.go（仅 super_admin 使用，已在 SuperHandler 中重新实现）**

旧 config_handler.go 可保留或等 Task 9 删除。

- [ ] **Step 5: 验证编译**

Run: `cd D:/MyGoProject/CloudKey && go build ./...`
Expected: 如果有旧文件引用错误，忽略（Task 9 统一处理）

- [ ] **Step 6: Commit**

```bash
git add internal/handler/
git commit -m "feat: refactor all handlers for tenant-scoped access"
```

---

