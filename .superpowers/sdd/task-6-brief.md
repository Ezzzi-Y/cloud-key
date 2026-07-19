### Task 6: 改造 service 层 — 加 tenant_id 过滤

**Files:**
- Modify: `internal/service/key_service.go`
- Modify: `internal/service/service_account_service.go`
- Modify: `internal/service/usage_log_service.go`
- Modify: `internal/service/login_log_service.go`
- Modify: `internal/service/stats_service.go`

**Interfaces:**
- 所有写操作从 `gin.Context` 或参数获取 `tenantID`
- Key CRUD/service/stats/usage log service 的所有 DB 查询加 `WHERE tenant_id = ?`

- [ ] **Step 1: 改造 `key_service.go` — 所有查询加 tenant_id 过滤**

修改 KeyService struct 的函数签名，全部加 `tenantID uint64` 参数：

**CreateKey** — 设置 `key.TenantID = tenantID`:
```go
func (s *KeyService) CreateKey(req CreateKeyRequest, tenantID uint64, keyPrefix string, keyLen, suffixLen int) (*CreateKeyResult, error) {
	// 使用传入的租户配置而非 service 默认值
	rawKey, err := s.generateRawKeyWithConfig(keyPrefix, keyLen)
	// ...
	key := model.Key{
		TenantID:        tenantID,
		Alias:           req.Alias,
		// ...
	}
	// ...
}
```

需要新增 `generateRawKeyWithConfig`:
```go
func (s *KeyService) generateRawKeyWithConfig(prefix string, keyLen int) (string, error) {
	bytes := make([]byte, keyLen)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random: %w", err)
	}
	return prefix + hex.EncodeToString(bytes), nil
}
```

**ListKeys** — 加 `tenantID`:
```go
func (s *KeyService) ListKeys(query KeyListQuery, tenantID uint64) ([]model.Key, int64, error) {
	// ...
	db := s.db.Model(&model.Key{}).Where("tenant_id = ?", tenantID)
	// ...
}
```

**GetKeyDetail, UpdateKey, DisableKey, EnableKey, DeleteKey** — 加 `tenantID`:
```go
func (s *KeyService) GetKeyDetail(id, tenantID uint64) (*model.Key, error) {
	var key model.Key
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}
```

类似更新所有 CRUD 方法。

**FindByRawKey — 公开接口不由 tenant admin 调用，暂不加 tenant 过滤（后续 ConsumeKey 关联时通过 Key 的 tenant_id 来隔离）**

**ConsumeKey** — 公开接口，消耗后记录 usage_log 时需写 tenant_id。通过 key 获取 tenant_id：
```go
// 在 record usage log 时:
key, _ := s.FindByRawKey(rawKey)
tenantID := uint64(0)
if key != nil { tenantID = key.TenantID }

// usageLogSvc.Record 传入 tenantID
```

改动 `RecordUsageParams`:
```go
type RecordUsageParams struct {
	TenantID       uint64
	KeyID          uint64
	KeyAlias       string
	Amount         int64
	IP             string
	UserAgent      string
	RequestPath    string
	RequestParams  string
	ResponseStatus int
}
```

**ExportKeys, ExportKeysJSON** — 加 `tenantID`:
```go
func (s *KeyService) ExportKeys(tenantID uint64) ([]model.Key, error) {
	var keys []model.Key
	if err := s.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}
```

- [ ] **Step 2: 改造 `service_account_service.go` — 加 tenant_id 过滤**

所有方法加 `tenantID` 参数:

**CreateServiceAccount**:
```go
func (s *ServiceAccountService) CreateServiceAccount(name string, tenantID uint64) (*model.ServiceAccount, string, error) {
	rawKey, err := s.GenerateServiceKey()
	// ...
	account := model.ServiceAccount{Name: name, KeyHash: hashServiceKey(rawKey), TenantID: tenantID}
	// ...
}
```

**ListServiceAccounts**:
```go
func (s *ServiceAccountService) ListServiceAccounts(tenantID uint64) ([]model.ServiceAccount, error) {
	var accounts []model.ServiceAccount
	if err := s.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}
```

**ToggleServiceAccount, DeleteServiceAccount** — 加 `tenantID`:
```go
func (s *ServiceAccountService) ToggleServiceAccount(id, tenantID uint64, isActive bool) error {
	return s.db.Model(&model.ServiceAccount{}).Where("id = ? AND tenant_id = ?", id, tenantID).Update("is_active", isActive).Error
}
```

- [ ] **Step 3: 改造 `usage_log_service.go` — Record 加 TenantID + 查询加 tenant_id**

**RecordUsageParams** 已有 TenantID 字段（Step 1 中已添加）:

**Record**:
```go
func (s *UsageLogService) Record(params RecordUsageParams) error {
	return s.db.Create(&model.UsageLog{
		TenantID:       params.TenantID,
		KeyID:          params.KeyID,
		// ...
	}).Error
}
```

**ListLogs, ExportLogs** — 加 `tenantID`:
```go
func (s *UsageLogService) ListLogs(query UsageLogQuery, tenantID uint64) ([]model.UsageLog, int64, error) {
	// ...
	db := s.db.Model(&model.UsageLog{}).Where("tenant_id = ?", tenantID)
	// ...
}
```

- [ ] **Step 4: 改造 `login_log_service.go` — 加 TenantID**

**RecordLogin**:
```go
func (s *LoginLogService) RecordLogin(userID uint64, tenantID *uint64, ip, userAgent string, success bool) error {
	status := model.LoginStatusFailed
	if success { status = model.LoginStatusSuccess }
	return s.db.Create(&model.LoginLog{
		UserID: userID, TenantID: tenantID, IP: ip, UserAgent: userAgent, Status: status,
	}).Error
}
```

**ListLoginLogs** — 加 `tenantID`（nullable for super_admin 查全部）:
```go
func (s *LoginLogService) ListLoginLogs(page, pageSize int, tenantID *uint64) ([]model.LoginLog, int64, error) {
	var logs []model.LoginLog
	var total int64
	db := s.db.Model(&model.LoginLog{})
	if tenantID != nil {
		db = db.Where("tenant_id = ?", *tenantID)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}
```

- [ ] **Step 5: 改造 `stats_service.go` — 加 tenant_id 过滤**

所有聚合查询加 `WHERE tenant_id = ?`:

**GetKeyOverview**:
```go
func (s *StatsService) GetKeyOverview(dateRange *DateRange, tenantID uint64) (*KeyOverview, error) {
	// ...
	keyDB := applyDateFilter(s.db.Model(&model.Key{}), dateRange).Where("tenant_id = ?", tenantID)
	// ...
}
```

类似更新 **GetTrends, GetTopKeys, GetTopIPs, GetDashboard** 全部加 `tenantID` 参数。

GetDashboard 和 GetTrends 中的 UsageLog 查询也需要加 `tenant_id` 过滤。

- [ ] ***Step 6: Commit earlier, then verify compilation. Update router.go to remove broken import paths and rerun: `go build ./...` (repeat until clean).***

Run: `cd D:/MyGoProject/CloudKey && go build ./...`
Expected: 编译通过（如果旧 handler 文件调用旧的 service 签名，会有编译错误 — 这是预期的，后续 task 会解决）

保留旧的 handler 文件暂时不删，它们会有编译错误。后续 task 创建新的 handler 后会删除旧文件。

- [ ] **Step 7: Commit**

```bash
git add internal/service/
git commit -m "feat: add tenant_id filtering to all service layer queries"
```

---

