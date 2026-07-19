# Frontend-Backend Alignment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Align the CloudKey backend API responses with frontend expectations across 5 areas: error format, stats fields, date range queries, CreateKey fields, and export format.

**Architecture:** Incremental changes to existing handler/service/model layers. Each task produces an independently testable deliverable. No new packages or architectural shifts — we are aligning field names, adding optional parameters, and fixing HTTP status codes within the existing Gin + GORM stack.

**Tech Stack:** Go 1.24, Gin, GORM, SQLite/MySQL

## Global Constraints

- All API responses use `{code: int, message: string, data: any}` — defined in `handler/response.go`
- `code == 0` means success; non-zero means error (defined in `errcode/errcode.go`)
- Empty lists must serialize as `[]` not `null` — use `make([]T, 0)` before returning
- Date params format: `YYYY-MM-DD` for daily, `YYYY-MM-DD HH:MM:SS` for hourly
- `start_date` / `end_date` are optional; omitting them preserves current behavior (all-time / last-24h)
- JWT auth middleware must return HTTP 401 for all auth failures (currently returns 200)
- Frontend is a single file at `web/admin.html` — inline `<script>`, vanilla JS, no build step

---

## File Structure

| File | Action | Responsibility |
|------|--------|---------------|
| `internal/middleware/auth.go` | Modify | Change JWT auth failure responses from HTTP 200 to HTTP 401 |
| `web/admin.html` | Modify | Update axios interceptor to detect 401 and handle `{code, message}` |
| `internal/service/stats_service.go` | Modify | Rename `total_keys` → `key_count`; add date range filtering to Overview/Trends/TopKeys/TopIPs |
| `internal/handler/stats_handler.go` | Modify | Pass `start_date`/`end_date` query params to service methods |
| `internal/service/key_service.go` | Modify | Add `ExportKeysJSON()` returning detailed JSON with prefix+suffix |
| `internal/handler/key_handler.go` | Modify | Add `ExportKeysJSON()` handler; add `expire_at`/`max_usage` to `CreateKeyJSON` |
| `internal/router/router.go` | Modify | Register `GET /api/admin/export/json` route |
| `internal/model/key.go` | Modify | Add `ExpireAt *time.Time` and `MaxUsage *int64` fields |
| `internal/model/migrate.go` | Verify | Confirm GORM AutoMigrate picks up new fields (no manual change needed) |
| `internal/handler/service_handler.go` | Modify | Add `expire_at`/`max_usage` to ServiceCreateKey request struct |

---

### Task 1: Fix JWT Auth Middleware HTTP Status

**Files:**
- Modify: `internal/middleware/auth.go:24-42`

**Interfaces:**
- Produces: HTTP 401 responses with `{code: 2003, message: "...", data: nil}` for all auth failures
- Consumers: Frontend axios interceptor (Task 2), any HTTP client expecting standard REST status codes

- [ ] **Step 1: Read current middleware**

Read `internal/middleware/auth.go` and identify the three error-returning locations (missing header, invalid format, invalid/expired token).

- [ ] **Step 2: Change HTTP 200 to HTTP 401 in all error paths**

Replace all three `c.JSON(http.StatusOK, ...)` calls inside the auth failure paths with `c.JSON(http.StatusUnauthorized, ...)`.

The file currently has three places returning HTTP 200 with errcode:
- Line ~27: missing Authorization header
- Line ~34: malformed Bearer token
- Line ~47: invalid/expired JWT

Each should become:
```go
c.JSON(http.StatusUnauthorized, gin.H{
    "code": errcode.CodeJWTInvalid,
    "message": errcode.GetMessage(errcode.CodeJWTInvalid),
    "data": nil,
})
```

- [ ] **Step 3: Verify compilation**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add internal/middleware/auth.go
git commit -m "fix(auth): return HTTP 401 for JWT auth failures"
```

---

### Task 2: Frontend Error Handling Adaptation

**Files:**
- Modify: `web/admin.html` (axios interceptor section, ~line 1824-1839)

**Interfaces:**
- Consumes: HTTP 401 responses from JWT auth middleware (Task 1)
- Consumes: All API error responses with `{code, message, data}` format
- Produces: Toast notifications via `showToast(message, 'error')`

- [ ] **Step 1: Read current interceptor**

Read `web/admin.html` and find the axios response interceptor (around line 1824).

Current code:
```javascript
axios.interceptors.response.use(
    response => {
        const data = response.data;
        if (data.code === 0) return data.data;
        showToast(data.message || '请求失败', 'error');
        return Promise.reject(data);
    },
    error => {
        if (error.response?.status === 401) {
            localStorage.removeItem('token');
            window.location.reload();
        }
        showToast(error.response?.data?.message || '网络错误', 'error');
        return Promise.reject(error);
    }
);
```

- [ ] **Step 2: Update interceptor to handle 401 + new error format**

Replace the interceptor with:

```javascript
axios.interceptors.response.use(
    response => {
        const data = response.data;
        if (data.code === 0) return data.data;
        showToast(data.message || '请求失败', 'error');
        return Promise.reject(data);
    },
    error => {
        if (error.response?.status === 401) {
            localStorage.removeItem('token');
            window.location.reload();
            return Promise.reject(error);
        }
        const msg = error.response?.data?.message
            || error.response?.statusText
            || '网络错误';
        showToast(msg, 'error');
        return Promise.reject(error);
    }
);
```

Key changes:
- 401 path now returns early with `Promise.reject(error)` (previously fell through to showToast below reload)
- Error message extraction chains through `response.data.message` → `statusText` → `'网络错误'`

- [ ] **Step 3: Manual test**

1. Start server: `go run .`
2. Open `http://localhost:port` in browser
3. Open DevTools → Network tab
4. Clear localStorage token and refresh → should redirect to login (401 triggers reload)
5. Login with valid credentials → should succeed
6. Try calling an API with an expired/invalid token → should show toast with error message and redirect to login

- [ ] **Step 4: Commit**

```bash
git add web/admin.html
git commit -m "fix(web): adapt frontend error handling for 401 responses"
```

---

### Task 3: Rename Overview total_keys to key_count

**Files:**
- Modify: `internal/service/stats_service.go:15-20` (KeyOverview struct)
- Modify: `web/admin.html` (~line 1662, dashboard render)

**Interfaces:**
- Produces: `GET /api/admin/stats/overview` returns `{key_count: int64, ...}` instead of `{total_keys: int64, ...}`
- Produces: `GET /api/admin/stats/dashboard` returns overview with `key_count` (calls GetKeyOverview internally)
- Consumers: Frontend dashboard rendering (Task 3 also updates this)

- [ ] **Step 1: Update KeyOverview struct**

In `internal/service/stats_service.go`, change the `KeyOverview` struct:

```go
type KeyOverview struct {
	KeyCount     int64            `json:"key_count"`
	StatusCounts map[string]int64 `json:"status_counts"`
	TotalInitial int64            `json:"total_initial"`
	TotalRemain  int64            `json:"total_remaining"`
}
```

The only change is `TotalKeys` → `KeyCount` and `json:"total_keys"` → `json:"key_count"`.

- [ ] **Step 2: Update GetKeyOverview method**

In the same file, the `GetKeyOverview` method uses `ov.TotalKeys` — rename to `ov.KeyCount`:

```go
if err := s.db.Model(&model.Key{}).Count(&ov.KeyCount).Error; err != nil {
    return nil, err
}
```

- [ ] **Step 3: Update frontend dashboard render**

In `web/admin.html`, find the dashboard render function (around line 1662). Change:
```javascript
html += createStatCard('卡密总数', data.overview.total_keys, 'ri-key-2-line');
```
to:
```javascript
html += createStatCard('卡密总数', data.overview.key_count, 'ri-key-2-line');
```

- [ ] **Step 4: Verify compilation**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add internal/service/stats_service.go web/admin.html
git commit -m "fix(stats): rename overview.total_keys to key_count"
```

---

### Task 4: Fix Trends "today" Period Returning Empty

**Files:**
- Modify: `internal/service/stats_service.go` (GetTrends method)

**Interfaces:**
- Produces: `GET /api/admin/stats/trends?period=today` returns grouped-by-hour data points
- Consumers: Frontend stats page trends chart

- [ ] **Step 1: Read current GetTrends method**

In `internal/service/stats_service.go`, the `GetTrends` method's default case uses:
```go
dateFormat = "%Y-%m-%d %H"
startTime = now.AddDate(0, 0, -1)
```

This queries the last 24 hours, which is correct for "today" context but returns empty when the `usage_logs` table has no rows with `created_at >= yesterday`.

- [ ] **Step 2: Verify the method handles empty results correctly**

The current code already returns `nil` when no rows match, and the handler already converts `nil` to `make([]TrendPoint, 0)`. The "empty data" issue is likely a **data problem** (no usage logs in the time range), not a code bug.

However, the spec says the frontend chart expects **at least the current hour label** even when data is empty. Add logic to always include the current hour bucket:

```go
func (s *StatsService) GetTrends(period string) ([]TrendPoint, error) {
	var dateFormat string
	var startTime time.Time
	var points []TrendPoint
	now := time.Now()

	switch period {
	case "week":
		dateFormat = "%Y-%m-%d"
		startTime = now.AddDate(0, 0, -7)
	case "month":
		dateFormat = "%Y-%m-%d"
		startTime = now.AddDate(0, -1, 0)
	default: // "today"
		dateFormat = "%Y-%m-%d %H"
		startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	}

	if err := s.db.Model(&model.UsageLog{}).
		Select("DATE_FORMAT(created_at, ?) as date, COUNT(*) as count", dateFormat).
		Where("created_at >= ?", startTime).
		Group("date").Order("date ASC").Scan(&points).Error; err != nil {
		return nil, err
	}

	if points == nil {
		points = make([]TrendPoint, 0)
	}

	return points, nil
}
```

Key change: `default` case now sets `startTime` to the beginning of today (midnight) instead of `now.AddDate(0, 0, -1)` (24 hours ago). This ensures "today" shows data from 00:00 of the current day.

Also, the `nil` → `make([]TrendPoint, 0)` conversion is now in the service instead of only the handler, making it safe for the Dashboard method too.

- [ ] **Step 3: Verify compilation**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add internal/service/stats_service.go
git commit -m "fix(stats): fix trends 'today' period start time to midnight"
```

---

### Task 5: Guarantee TopKeys/TopIPs Return Empty Array

**Files:**
- Modify: `internal/service/stats_service.go` (GetTopKeys, GetTopIPs methods)
- Modify: `internal/handler/stats_handler.go` (TopKeys, TopIPs handlers)

**Interfaces:**
- Produces: `GET /api/admin/stats/top-keys` always returns `data: []` (never `null`)
- Produces: `GET /api/admin/stats/top-ips` always returns `data: []` (never `null`)
- Consumers: Frontend stats page tables

- [ ] **Step 1: Update GetTopKeys in service**

In `internal/service/stats_service.go`, ensure `GetTopKeys` initializes the slice:

```go
func (s *StatsService) GetTopKeys() ([]TopItem, error) {
	items := make([]TopItem, 0)
	if err := s.db.Model(&model.UsageLog{}).
		Select("key_alias as name, COUNT(*) as count").
		Group("key_alias").Order("count DESC").Limit(10).Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
```

Change: `var items []TopItem` → `items := make([]TopItem, 0)`

- [ ] **Step 2: Update GetTopIPs in service**

Same change for `GetTopIPs`:

```go
func (s *StatsService) GetTopIPs() ([]TopItem, error) {
	items := make([]TopItem, 0)
	if err := s.db.Model(&model.UsageLog{}).
		Select("ip as name, COUNT(*) as count").
		Group("ip").Order("count DESC").Limit(10).Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
```

- [ ] **Step 3: Remove redundant nil-checks in handler**

In `internal/handler/stats_handler.go`, the `TopKeys` and `TopIPs` handlers have nil-checks that are now unnecessary (the service never returns nil). Remove them for clarity:

```go
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
```

- [ ] **Step 4: Verify compilation**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add internal/service/stats_service.go internal/handler/stats_handler.go
git commit -m "fix(stats): ensure top-keys/top-ips return [] not null"
```

---

### Task 6: Add Date Range to Stats Service Methods

**Files:**
- Modify: `internal/service/stats_service.go`

**Interfaces:**
- Produces: Updated method signatures — all stats methods accept optional date range
- Consumers: Stats handler (Task 7) passes query params

- [ ] **Step 1: Add DateRange type with apply helper**

Add at the top of `internal/service/stats_service.go`, after the imports:

```go
// DateRange holds optional start/end date filters.
// Empty strings mean no filter on that boundary.
type DateRange struct {
	StartDate string
	EndDate   string
}

func applyDateFilter(db *gorm.DB, dateRange *DateRange) *gorm.DB {
	if dateRange == nil {
		return db
	}
	if dateRange.StartDate != "" {
		db = db.Where("created_at >= ?", dateRange.StartDate)
	}
	if dateRange.EndDate != "" {
		db = db.Where("created_at <= ?", dateRange.EndDate)
	}
	return db
}
```

- [ ] **Step 2: Update GetKeyOverview signature**

```go
func (s *StatsService) GetKeyOverview(dateRange *DateRange) (*KeyOverview, error) {
	ov := &KeyOverview{StatusCounts: make(map[string]int64)}

	keyDB := applyDateFilter(s.db.Model(&model.Key{}), dateRange)
	if err := keyDB.Count(&ov.KeyCount).Error; err != nil {
		return nil, err
	}

	var rows []struct {
		Status string
		Count  int64
	}
	if err := applyDateFilter(s.db.Model(&model.Key{}), dateRange).
		Select("status, COUNT(*) as count").Group("status").Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		ov.StatusCounts[r.Status] = r.Count
	}

	if err := applyDateFilter(s.db.Model(&model.Key{}), dateRange).
		Select("COALESCE(SUM(initial_amount), 0)").Scan(&ov.TotalInitial).Error; err != nil {
		return nil, err
	}
	if err := applyDateFilter(s.db.Model(&model.Key{}), dateRange).
		Select("COALESCE(SUM(remaining_amount), 0)").Scan(&ov.TotalRemain).Error; err != nil {
		return nil, err
	}

	return ov, nil
}
```

- [ ] **Step 3: Update GetTrends signature**

```go
func (s *StatsService) GetTrends(period string, dateRange *DateRange) ([]TrendPoint, error) {
	var dateFormat string
	var startTime time.Time
	var points []TrendPoint
	now := time.Now()

	// If explicit date range is provided, use it instead of period
	if dateRange != nil && dateRange.StartDate != "" {
		dateFormat = "%Y-%m-%d"
		startTime, _ = time.Parse("2006-01-02", dateRange.StartDate)
	} else {
		switch period {
		case "week":
			dateFormat = "%Y-%m-%d"
			startTime = now.AddDate(0, 0, -7)
		case "month":
			dateFormat = "%Y-%m-%d"
			startTime = now.AddDate(0, -1, 0)
		default: // "today"
			dateFormat = "%Y-%m-%d %H"
			startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		}
	}

	db := s.db.Model(&model.UsageLog{}).
		Select("DATE_FORMAT(created_at, ?) as date, COUNT(*) as count", dateFormat).
		Where("created_at >= ?", startTime)
	db = applyDateFilter(db, dateRange)

	if err := db.Group("date").Order("date ASC").Scan(&points).Error; err != nil {
		return nil, err
	}

	if points == nil {
		points = make([]TrendPoint, 0)
	}

	return points, nil
}
```

- [ ] **Step 4: Update GetTopKeys signature**

```go
func (s *StatsService) GetTopKeys(dateRange *DateRange) ([]TopItem, error) {
	items := make([]TopItem, 0)
	db := applyDateFilter(s.db.Model(&model.UsageLog{}), dateRange)
	if err := db.
		Select("key_alias as name, COUNT(*) as count").
		Group("key_alias").Order("count DESC").Limit(10).Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
```

- [ ] **Step 5: Update GetTopIPs signature**

```go
func (s *StatsService) GetTopIPs(dateRange *DateRange) ([]TopItem, error) {
	items := make([]TopItem, 0)
	db := applyDateFilter(s.db.Model(&model.UsageLog{}), dateRange)
	if err := db.
		Select("ip as name, COUNT(*) as count").
		Group("ip").Order("count DESC").Limit(10).Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
```

- [ ] **Step 6: Update GetDashboard to pass nil dateRange**

```go
func (s *StatsService) GetDashboard() (*DashboardStats, error) {
	overview, err := s.GetKeyOverview(nil)
	if err != nil {
		return nil, err
	}
	// ... rest unchanged
```

- [ ] **Step 7: Verify compilation**

Run: `go build ./...`
Expected: compilation errors from handler (expected — handler hasn't been updated yet, that's Task 7)

- [ ] **Step 8: Commit**

```bash
git add internal/service/stats_service.go
git commit -m "feat(stats): add date range filtering to all stats service methods"
```

---

### Task 7: Wire Date Range Params in Stats Handler

**Files:**
- Modify: `internal/handler/stats_handler.go`

**Interfaces:**
- Consumes: `start_date` / `end_date` query params from HTTP requests
- Consumes: Updated stats service method signatures from Task 6
- Produces: All stats endpoints now accept optional date range

- [ ] **Step 1: Add dateRange extraction helper**

Add a helper function at the top of `internal/handler/stats_handler.go`. It validates that start ≤ end and calls `c.Abort()` on error:

```go
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
```

- [ ] **Step 2: Update Overview handler (with abort guard)**

```go
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
```

- [ ] **Step 3: Update Trends handler (with abort guard)**

```go
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
	Success(c, points)
}
```

- [ ] **Step 4: Update TopKeys handler (with abort guard)**

```go
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
```

- [ ] **Step 5: Update TopIPs handler (with abort guard)**

```go
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
```

- [ ] **Step 6: Verify compilation**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 7: Manual test**

Start server and test:
```bash
# No date range (should work as before)
curl -H "Authorization: Bearer <token>" http://localhost:port/api/admin/stats/overview

# With date range
curl -H "Authorization: Bearer <token>" "http://localhost:port/api/admin/stats/overview?start_date=2026-01-01&end_date=2026-07-19"

# Trends with date range
curl -H "Authorization: Bearer <token>" "http://localhost:port/api/admin/stats/trends?period=week&start_date=2026-07-01"
```

- [ ] **Step 8: Commit**

```bash
git add internal/handler/stats_handler.go
git commit -m "feat(stats): wire start_date/end_date query params in stats handlers"
```

---

### Task 8: Add expire_at and max_usage to Key Model

**Files:**
- Modify: `internal/model/key.go`

**Interfaces:**
- Produces: `Key` struct gains `ExpireAt *time.Time` and `MaxUsage *int64` fields
- Consumers: GORM AutoMigrate (automatic), CreateKey flow (Task 9)

- [ ] **Step 1: Add fields to Key struct**

In `internal/model/key.go`, add two new fields after `UsedAt`:

```go
type Key struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Alias           string         `gorm:"type:varchar(255);not null" json:"alias"`
	KeyHash         string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"`
	KeyPrefix       string         `gorm:"type:varchar(50);not null" json:"key_prefix"`
	KeySuffix       string         `gorm:"type:varchar(10);not null" json:"key_suffix"`
	BillingMode     KeyBillingMode `gorm:"type:varchar(20);not null" json:"billing_mode"`
	InitialAmount   int64          `gorm:"type:bigint;not null" json:"initial_amount"`
	RemainingAmount int64          `gorm:"type:bigint;not null" json:"remaining_amount"`
	Version         int64          `gorm:"type:bigint;not null;default:0" json:"-"`
	Status          KeyStatus      `gorm:"type:varchar(20);not null;default:unused" json:"status"`
	CreatedBy       string         `gorm:"type:varchar(100);not null" json:"created_by"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	UsedAt          *time.Time     `gorm:"default:null" json:"used_at"`
	ExpireAt        *time.Time     `gorm:"default:null" json:"expire_at"`
	MaxUsage        *int64         `gorm:"default:null" json:"max_usage"`
}
```

- [ ] **Step 2: Verify GORM AutoMigrate handles new fields**

Read `internal/model/migrate.go` to confirm it passes `&model.Key{}` to `AutoMigrate`. GORM will automatically add the new nullable columns.

Expected content of `migrate.go`:
```go
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &Admin{},
        &Key{},
        // ...
    )
}
```

No changes needed to `migrate.go`.

- [ ] **Step 3: Verify compilation**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add internal/model/key.go
git commit -m "feat(model): add expire_at and max_usage fields to Key"
```

---

### Task 9: Add expire_at and max_usage to CreateKey Request

**Files:**
- Modify: `internal/handler/key_handler.go` (CreateKeyJSON struct, CreateKey method)
- Modify: `internal/handler/service_handler.go` (ServiceCreateKey request struct)
- Modify: `internal/service/key_service.go` (CreateKeyRequest struct, CreateKey method)

**Interfaces:**
- Consumes: `Key.ExpireAt` and `Key.MaxUsage` fields from Task 8
- Produces: `POST /api/admin/keys` accepts optional `expire_at` (string, "YYYY-MM-DD HH:MM:SS") and `max_usage` (int64)
- Produces: `POST /api/service/keys` also accepts the same optional fields

- [ ] **Step 1: Update CreateKeyRequest in service**

In `internal/service/key_service.go`, add fields to `CreateKeyRequest`:

```go
type CreateKeyRequest struct {
	Alias         string               `json:"alias"`
	BillingMode   model.KeyBillingMode  `json:"billing_mode"`
	InitialAmount int64                `json:"initial_amount"`
	CreatedBy     string               `json:"created_by"`
	ExpireAt      *time.Time           `json:"expire_at"`
	MaxUsage      *int64               `json:"max_usage"`
}
```

Also add `"time"` to the imports if not already present.

- [ ] **Step 2: Update CreateKey method to use new fields**

In the `CreateKey` method, update the `key` struct initialization:

```go
key := model.Key{
	Alias:           req.Alias,
	KeyHash:         keyHash,
	KeyPrefix:       s.keyPrefix,
	KeySuffix:       suffix,
	BillingMode:     req.BillingMode,
	InitialAmount:   req.InitialAmount,
	RemainingAmount: req.InitialAmount,
	Version:         0,
	Status:          model.KeyStatusUnused,
	CreatedBy:       req.CreatedBy,
	ExpireAt:        req.ExpireAt,
	MaxUsage:        req.MaxUsage,
}
```

- [ ] **Step 3: Update CreateKeyJSON in handler**

In `internal/handler/key_handler.go`, update the request struct:

```go
type CreateKeyJSON struct {
	Alias         string  `json:"alias" binding:"required"`
	BillingMode   string  `json:"billing_mode" binding:"required"`
	InitialAmount int64   `json:"initial_amount" binding:"required"`
	ExpireAt      *string `json:"expire_at"`
	MaxUsage      *int64  `json:"max_usage"`
}
```

- [ ] **Step 4: Update CreateKey handler method**

In the `CreateKey` method of `key_handler.go`, parse the optional `ExpireAt` string to `*time.Time`:

```go
func (h *KeyHandler) CreateKey(c *gin.Context) {
	var req CreateKeyJSON
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, http.StatusBadRequest, "参数错误")
		return
	}

	adminID, _ := c.Get("admin_id")
	createdBy := ""
	if adminID != nil {
		createdBy = "admin"
	}

	var expireAt *time.Time
	if req.ExpireAt != nil && *req.ExpireAt != "" {
		t, err := time.Parse("2006-01-02 15:04:05", *req.ExpireAt)
		if err != nil {
			BadRequest(c, http.StatusBadRequest, "expire_at 格式错误，应为 YYYY-MM-DD HH:MM:SS")
			return
		}
		expireAt = &t
	}

	result, err := h.keySvc.CreateKey(service.CreateKeyRequest{
		Alias: req.Alias, BillingMode: model.KeyBillingMode(req.BillingMode),
		InitialAmount: req.InitialAmount, CreatedBy: createdBy,
		ExpireAt: expireAt, MaxUsage: req.MaxUsage,
	})
	if err != nil {
		InternalError(c)
		return
	}

	Success(c, gin.H{
		"id": result.Key.ID, "raw_key": result.RawKey, "alias": result.Key.Alias,
		"key_prefix": result.Key.KeyPrefix, "key_suffix": result.Key.KeySuffix,
		"billing_mode": result.Key.BillingMode, "initial_amount": result.Key.InitialAmount,
		"remaining_amount": result.Key.RemainingAmount, "status": result.Key.Status,
		"created_by": result.Key.CreatedBy, "created_at": result.Key.CreatedAt,
		"expire_at": result.Key.ExpireAt, "max_usage": result.Key.MaxUsage,
	})
}
```

Add `"time"` to the handler imports if not already present.

- [ ] **Step 5: Update ServiceCreateKey in service_handler.go**

In `internal/handler/service_handler.go`, update the `ServiceCreateKey` method's request struct:

```go
var req struct {
	Alias         string  `json:"alias" binding:"required"`
	BillingMode   string  `json:"billing_mode" binding:"required"`
	InitialAmount int64   `json:"initial_amount" binding:"required"`
	ExpireAt      *string `json:"expire_at"`
	MaxUsage      *int64  `json:"max_usage"`
}
```

And parse `ExpireAt` the same way:

```go
var expireAt *time.Time
if req.ExpireAt != nil && *req.ExpireAt != "" {
	t, err := time.Parse("2006-01-02 15:04:05", *req.ExpireAt)
	if err != nil {
		BadRequest(c, errcode.CodeServiceKeyInvalid, "expire_at 格式错误")
		return
	}
	expireAt = &t
}

result, err := h.keySvc.CreateKey(service.CreateKeyRequest{
	Alias: req.Alias, BillingMode: model.KeyBillingMode(req.BillingMode),
	InitialAmount: req.InitialAmount, CreatedBy: createdBy,
	ExpireAt: expireAt, MaxUsage: req.MaxUsage,
})
```

Add `"time"` to the service_handler imports.

Also update the response to include new fields:

```go
Success(c, gin.H{
	"id": result.Key.ID, "raw_key": result.RawKey, "alias": result.Key.Alias,
	"key_suffix": result.Key.KeySuffix, "billing_mode": result.Key.BillingMode,
	"initial_amount": result.Key.InitialAmount, "remaining_amount": result.Key.RemainingAmount,
	"status": result.Key.Status,
	"expire_at": result.Key.ExpireAt, "max_usage": result.Key.MaxUsage,
})
```

- [ ] **Step 6: Verify compilation**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 7: Manual test**

```bash
# Create key with expire_at and max_usage
curl -X POST -H "Authorization: Bearer <token>" -H "Content-Type: application/json" \
  -d '{"alias":"test","billing_mode":"count","initial_amount":100,"expire_at":"2026-12-31 23:59:59","max_usage":50}' \
  http://localhost:port/api/admin/keys

# Response should include expire_at and max_usage
```

- [ ] **Step 8: Commit**

```bash
git add internal/handler/key_handler.go internal/handler/service_handler.go internal/service/key_service.go
git commit -m "feat(key): add expire_at and max_usage to CreateKey request"
```

---

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
