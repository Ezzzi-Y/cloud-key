### Task 11: 使用记录 + 统计 + 管理员 + 服务账号 + 登录日志 + 配置服务

**Status:** DONE

**Commit:** e8e2c20 feat(service): add all service layer (usage_log, stats, admin, service_account, login_log, config)

**Build summary:**
- `gofmt -w internal/service/` -- clean, no formatting issues
- `go build ./internal/service/` -- clean, 0 errors

**Files created (6):**
- `internal/service/usage_log_service.go` -- UsageLogService: Record, ListLogs, ExportLogs
- `internal/service/stats_service.go` -- StatsService: GetKeyOverview, GetTrends, GetTopKeys, GetTopIPs, GetDashboard
- `internal/service/admin_service.go` -- AdminService: Login, VerifyTOTP, GenerateTOTPSecret, ConfirmTOTPSetup, ChangePassword, GetAdminProfile, SeedAdmin
- `internal/service/service_account_service.go` -- ServiceAccountService: CreateServiceAccount, ValidateServiceKey, ListServiceAccounts, ToggleServiceAccount, DeleteServiceAccount
- `internal/service/login_log_service.go` -- LoginLogService: RecordLogin, ListLoginLogs
- `internal/service/config_service.go` -- ConfigService: GetConfig, GetAllConfigs, SetConfig, InitDefaultConfigs

**Concerns:** None. All 6 service files compile cleanly against existing models and dependencies.

---

### Task 11 Critical Fixes (post-review)

**Status:** DONE

**Commit:** 520fcb8 fix(service): fix HMAC key, check DB errors in stats and login_log services

**Build summary:**
- `gofmt -w internal/service/` -- clean
- `go build ./internal/service/` -- clean, 0 errors

**Fixes applied:**

| ID  | File                                 | Issue                                      | Fix                                                                 |
|-----|--------------------------------------|--------------------------------------------|----------------------------------------------------------------------|
| C1  | `service_account_service.go`         | `hashServiceKey` used key as HMAC secret (equivalent to plain SHA256) | Added `hmacServiceKeySecret` constant; pass it as HMAC key, use input as message |
| C2  | `stats_service.go`                   | 7 GORM queries silently discarded `.Error` | All queries now check and return errors                              |
| C3  | `login_log_service.go`               | Count and Find in `ListLoginLogs` discarded errors | Both queries now check and return errors                            |
