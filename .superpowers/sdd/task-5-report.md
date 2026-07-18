### Task 5: 管理员 + 服务账号 + 登录日志 + 系统配置数据模型

**Status:** DONE

**Commits:**
- `038b316` feat(model): add Admin, ServiceAccount, LoginLog, SysConfig GORM models

**Build summary:**
- `gofmt -w` -- no formatting changes needed (files were already properly formatted)
- `go build ./internal/model/` -- succeeded with zero errors

**Files created:**
- `internal/model/admin.go` -- Admin struct (username, password_hash, totp_secret, totp_setup, is_active)
- `internal/model/service_account.go` -- ServiceAccount struct (name, key_hash, is_active)
- `internal/model/login_log.go` -- LoginLog struct + LoginStatus enum (success/failed), indexed on admin_id
- `internal/model/config.go` -- SysConfig struct (key, value, description), uniqueIndex on key

**Report file:** D:\MyGoProject\CloudKey\.superpowers\sdd\task-5-report.md
