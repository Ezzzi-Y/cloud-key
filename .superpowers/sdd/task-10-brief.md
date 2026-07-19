### Task 10: Add JSON Export Endpoint

**Files:**
- Modify: `internal/service/key_service.go` (add ExportKeysJSON method)
- Modify: `internal/handler/key_handler.go` (add ExportKeysJSON handler)
- Modify: `internal/router/router.go` (register new route)

**Interfaces:**
- Produces: `GET /api/admin/export/json` returns `[{id, key_prefix+key_suffix, alias, billing_mode, initial_amount, remaining_amount, status, created_at, expire_at, max_usage}, ...]`
- Note: The export intentionally does NOT return `raw_key` (it was never stored). It returns `key_prefix + key_suffix` as a visual identifier.

- [ ] **Step 1: Add ExportKeysJSON method to service**

In `internal/service/key_service.go`, add:

```go
type ExportKeyItem struct {
	ID              uint64              `json:"id"`
	KeyPrefix       string              `json:"key_prefix"`
	KeySuffix       string              `json:"key_suffix"`
	Alias           string              `json:"alias"`
	BillingMode     model.KeyBillingMode `json:"billing_mode"`
	InitialAmount   int64               `json:"initial_amount"`
	RemainingAmount int64               `json:"remaining_amount"`
	Status          model.KeyStatus     `json:"status"`
	CreatedAt       time.Time           `json:"created_at"`
	ExpireAt        *time.Time          `json:"expire_at"`
	MaxUsage        *int64              `json:"max_usage"`
}

func (s *KeyService) ExportKeysJSON() ([]ExportKeyItem, error) {
	var keys []model.Key
	if err := s.db.Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, err
	}

	items := make([]ExportKeyItem, len(keys))
	for i, k := range keys {
		items[i] = ExportKeyItem{
			ID:              k.ID,
			KeyPrefix:       k.KeyPrefix,
			KeySuffix:       k.KeySuffix,
			Alias:           k.Alias,
			BillingMode:     k.BillingMode,
			InitialAmount:   k.InitialAmount,
			RemainingAmount: k.RemainingAmount,
			Status:          k.Status,
			CreatedAt:       k.CreatedAt,
			ExpireAt:        k.ExpireAt,
			MaxUsage:        k.MaxUsage,
		}
	}
	return items, nil
}
```

- [ ] **Step 2: Add ExportKeysJSON handler**

In `internal/handler/key_handler.go`, add:

```go
func (h *KeyHandler) ExportKeysJSON(c *gin.Context) {
	items, err := h.keySvc.ExportKeysJSON()
	if err != nil {
		InternalError(c)
		return
	}
	if items == nil {
		items = make([]service.ExportKeyItem, 0)
	}
	Success(c, items)
}
```

- [ ] **Step 3: Register the new route**

In `internal/router/router.go`, add the new endpoint inside the `adminAuth` group, near the existing export route:

```go
adminAuth.GET("/keys/export", keyHandler.ExportKeys)
adminAuth.GET("/export/json", keyHandler.ExportKeysJSON)  // new
```

- [ ] **Step 4: Verify compilation**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 5: Manual test**

```bash
# JSON export
curl -H "Authorization: Bearer <token>" http://localhost:port/api/admin/export/json

# Expected: array of objects with id, key_prefix, key_suffix, alias, etc.
```

- [ ] **Step 6: Commit**

```bash
git add internal/service/key_service.go internal/handler/key_handler.go internal/router/router.go
git commit -m "feat(export): add GET /api/admin/export/json endpoint"
```

---

## Self-Review

**1. Spec coverage check:**

| Spec Requirement | Task(s) |
|---|---|
| Standardize error format `{code, message, data}` | Already done (existing code). Task 1, 2 handle HTTP status alignment. |
| JWT auth returns HTTP 401 | Task 1 |
| Frontend handles 401 + error format | Task 2 |
| Overview: `total_keys` → `key_count` | Task 3 |
| Trends: `today` period returns data | Task 4 |
| TopKeys/TopIPs: return `[]` not `null` | Task 5 |
| Date range: Overview accepts `start_date`/`end_date` | Task 6, 7 |
| Date range: Trends accepts `start_date`/`end_date` | Task 6, 7 |
| Date range: TopKeys accepts `start_date`/`end_date` | Task 6, 7 |
| Date range: TopIPs accepts `start_date`/`end_date` | Task 6, 7 |
| Date range validation: start > end returns 400 | Task 7 (extractDateRange validates and calls c.Abort()) |
| CreateKey: add `expire_at` optional param | Task 9 |
| CreateKey: add `max_usage` optional param | Task 9 |
| CreateKey: auto-set status to unused | Already done (existing code sets `Status: model.KeyStatusUnused`) |
| Export: add `GET /api/admin/export/json` | Task 10 |
| Export: return prefix+suffix, alias, billing_mode, amounts, status, created_at | Task 10 |
| Error codes: 1001~1999 key, 2001~2999 auth, 3001~3999 service, 9999 system | Already done |

All spec requirements are covered. No gaps remain.

**2. Placeholder scan:** No TBD/TODO/placeholders found.

**3. Type consistency check:**
- `KeyOverview.KeyCount` (json `key_count`) — used consistently in service + frontend
- `DateRange.StartDate/EndDate` — used consistently in service + handler
- `ExportKeyItem` — fields match spec exactly
- `CreateKeyRequest.ExpireAt *time.Time` / `MaxUsage *int64` — consistent across handler + service
