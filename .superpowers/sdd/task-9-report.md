## Task 9 Report: Wire Up Router + Main

**Status:** DONE

---

### What Was Done

1. **Deleted old files:**
   - `internal/model/admin.go` — old `Admin` model
   - `internal/service/admin_service.go` — old `AdminService`
   - `internal/handler/admin_handler.go` — old `AdminHandler`
   - `internal/service/admin_service_test.go` — orphaned test for deleted service

2. **Rewrote `internal/router/router.go`:**
   - New 3-tier route structure: public (`/api/auth`, `/api/key`), super admin (`/api/super`), tenant admin (`/api/tenant`)
   - Added `KeyHandler` param for public Status/Consume endpoints (separate from `TenantKeyHandler` which lacks those methods)
   - Added `strings` import for SPA fallback `NoRoute` handler
   - Added `service.ServiceAccountService` import for `ServiceAuthMiddleware`
   - Added `ExportKeysJSON` route missing from the task brief
   - `TenantBusinessGuard` applied only to write operations (create/update/delete), not reads

3. **Rewrote `main.go`:**
   - Switched from `AdminService` to `AuthService`
   - Seeds super admin via `authSvc.SeedSuperAdmin()` using `SUPER_ADMIN_USERNAME` / `SUPER_ADMIN_PASSWORD` env vars
   - All new handler constructors wired with correct param counts:
     - `NewTenantKeyHandler(keySvc, usageLogSvc, db, false)` — 4 params (brief had 3)
     - `NewSuperHandler(tenantSvc, configSvc, loginLogSvc)` — 3 params (brief had 4, incorrectly adding `statsSvc`)
     - `NewTenantUsageLogHandler(usageLogSvc, loginLogSvc)` — 2 params
   - `ServiceAuthMiddleware(saSvc, db)` — 2 params

4. **Updated `internal/config/config.go`:**
   - `AuthConfig.AdminUsername` → `AuthConfig.SuperAdminUsername`
   - `AuthConfig.AdminPassword` → `AuthConfig.SuperAdminPassword`

5. **Updated `config.yaml.example`:**
   - `admin_username` → `super_admin_username`
   - `admin_password` → `super_admin_password`

6. **Fixed `internal/model/migrate.go`:**
   - Removed `&Admin{}` from `AutoMigrate` list (replaced by `User`)

---

### Discrepancies With Task Brief

The brief's code had several constructor signature mismatches that were corrected:

| Constructor | Brief | Actual |
|---|---|---|
| `NewSuperHandler` | 4 params (tenantSvc, configSvc, statsSvc, loginLogSvc) | 3 params (tenantSvc, configSvc, loginLogSvc) — no statsSvc |
| `NewTenantKeyHandler` | 3 params (keySvc, usageLogSvc, recordParams) | 4 params (keySvc, usageLogSvc, db, recordParams) |
| Public Status/Consume | Used `tenantKeyHandler` | Required separate `KeyHandler` (TenantKeyHandler lacks those methods) |

The brief also omitted the `"strings"` and `"CloudKey/internal/service"` imports, and didn't include the `ExportKeysJSON` tenant route.

---

### Test Summary

- `go build ./...` — PASS (zero errors)
- No `go test` run needed (task is wiring only; deleted orphaned `admin_service_test.go` which referenced removed types)

---

### Files Modified

| File | Action |
|---|---|
| `internal/router/router.go` | Rewritten |
| `main.go` | Rewritten |
| `internal/config/config.go` | Edited (AuthConfig fields) |
| `config.yaml.example` | Edited (auth fields) |
| `internal/model/migrate.go` | Edited (removed Admin) |
| `internal/model/admin.go` | Deleted |
| `internal/service/admin_service.go` | Deleted |
| `internal/handler/admin_handler.go` | Deleted |
| `internal/service/admin_service_test.go` | Deleted |

### Commit

- `50ae6d9` — `feat: wire up new router and main.go for multi-tenant SaaS`

---

**Concerns:** None. Compilation clean, all signatures verified against actual code.
