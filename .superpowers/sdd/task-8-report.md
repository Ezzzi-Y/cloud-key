# Task 8 Report: Refactor ServiceAccount/Stats/UsageLog Handlers

**Status:** DONE

## Commit

- `3f7389afe7b9b8bdd62c4028464460989aca5614` — `feat: refactor handlers for tenant-scoped access`

## Files Changed

- `internal/handler/service_handler.go` — `ServiceHandler` renamed to `TenantServiceAccountHandler`; constructor renamed to `NewTenantServiceAccountHandler`. All method receivers updated. `ServiceCreateKey` and `ServiceListKeys` preserved with `c.Get("service_account")` pattern, using hardcoded defaults `"sk-", 32, 4`.
- `internal/handler/stats_handler.go` — `StatsHandler` renamed to `TenantStatsHandler`; constructor renamed to `NewTenantStatsHandler`. All 5 methods (`Dashboard`, `Overview`, `Trends`, `TopKeys`, `TopIPs`) updated with new receiver type.
- `internal/handler/usage_log_handler.go` — `UsageLogHandler` renamed to `TenantUsageLogHandler`; constructor now takes `*service.LoginLogService` as second param. `ListLogs` and `ExportLogs` preserved. New `LoginLogs` method added, ported from `AuthHandler.LoginLogs` with identical `*uint64` tenantIDPtr pattern.
- `internal/service/key_service.go` — New `ListKeysByTenant(tenantID uint64, page, pageSize int)` method added after `ListKeysByCreatedBy`.

## Test Summary

`go build ./...` — expected build errors in `main.go` and `router.go` (still reference old handler names `ServiceHandler`, `StatsHandler`, `UsageLogHandler`). These will be resolved in Task 9 (router + main assembly). No logic errors in the handler or service packages themselves.

## Concerns

None. Build failures are confined to `main.go`/`router.go` which Task 9 owns.
