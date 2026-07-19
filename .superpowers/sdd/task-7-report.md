# Task 7 Report: SuperHandler + TenantHandler + TenantService

## Status: DONE

## Summary

Created `TenantService`, `SuperHandler`, and refactored `KeyHandler` into `TenantKeyHandler` (tenant-scoped) + `KeyHandler` (public API). Fixed all handler compilation errors caused by the service-layer tenantID signature changes from Task 6.

## Files Changed

### New files
- `internal/service/tenant_service.go` — TenantService with CreateTenant, ListTenants, GetTenant, UpdateTenant, ResetPassword, generateRandomPassword
- `internal/handler/super_handler.go` — SuperHandler with ListTenants, CreateTenant, GetTenant, UpdateTenant, ResetPassword, GetConfigs, UpdateConfigs, LoginLogs

### Modified files
- `internal/handler/key_handler.go` — Refactored: `KeyHandler` now only holds public API (Status, Consume). `TenantKeyHandler` added with tenant-scoped admin methods (CreateKey, ListKeys, GetKey, UpdateKey, DisableKey, EnableKey, DeleteKey, ExportKeys, ExportKeysJSON). `TenantKeyHandler.CreateKey` fetches tenant KeyPrefix/KeyLength/KeySuffixLength from DB via `getTenantKeyConfig()` helper. `pageParams(c)` and `parseExpireAt()` preserved as-is.
- `internal/handler/auth_handler.go` — Fixed `RecordLogin` calls to pass `*uint64` tenantID (nil at login time, only available after TOTP verification). Fixed `ListLoginLogs` call to pass tenantID pointer.
- `internal/handler/admin_handler.go` — Fixed `RecordLogin` calls to pass nil tenantID (super admin has no tenant). Fixed `ListLoginLogs` call to pass nil.
- `internal/handler/service_handler.go` — Fixed all service calls to pass `sa.TenantID` for ServiceCreateKey/ServiceListKeys, `getTenantID(c)` for ListServiceAccounts/CreateServiceAccount/ToggleServiceAccount/DeleteServiceAccount.
- `internal/handler/stats_handler.go` — Fixed all StatsService calls to pass `getTenantID(c)`.
- `internal/handler/usage_log_handler.go` — Fixed ListLogs/ExportLogs calls to pass `getTenantID(c)`.
- `internal/service/key_service_test.go` — Updated test table schema (added `tenant_id`, `expire_at`, `max_usage` columns) and all test function calls to use new method signatures.

## Compilation Status

- `./internal/handler/` — PASS
- `./internal/service/` — PASS
- `./internal/model/` — PASS
- `./internal/middleware/` — PASS
- `./...` — Expected errors in `internal/router/router.go` (old KeyHandler method references). Will be fixed in Task 9.

## Test Results

```
go test ./internal/service/ -v
=== RUN   TestSeedAdmin          --- PASS
=== RUN   TestLogin_Success      --- PASS
=== RUN   TestLogin_WrongPassword --- PASS
=== RUN   TestChangePassword     --- PASS
=== RUN   TestCreateKey          --- PASS
=== RUN   TestFindByRawKey       --- PASS
=== RUN   TestFindByRawKey_NotFound --- PASS
=== RUN   TestGetKeyStatus       --- PASS
=== RUN   TestListKeys           --- PASS
=== RUN   TestDisableEnableKey   --- PASS
PASS (10/10)
```

## Concerns

- `router.go` and `main.go` will not compile until Task 9 (expected, as noted in the task brief).
- `AdminHandler.LoginLogs` and `AuthHandler.Login` pass nil for tenantID to `RecordLogin` since tenantID isn't available at those call sites. Full login-log tenant attribution will require AuthService changes (out of scope for this task).

## Commits

- `bff4431` — feat: add SuperHandler, TenantService, refactor KeyHandler for tenant scope
