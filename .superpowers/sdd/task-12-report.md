## Task 12: 服务层单元测试 — 完成报告

### 状态: DONE_WITH_CONCERNS

### 提交
- `0f338b4` test(service): add unit tests for KeyService and AdminService

### 测试结果
全部 PASS（10/10）：

**admin_service_test.go (4 tests)**
- TestSeedAdmin — PASS
- TestLogin_Success — PASS
- TestLogin_WrongPassword — PASS
- TestChangePassword — PASS

**key_service_test.go (6 tests)**
- TestCreateKey — PASS
- TestFindByRawKey — PASS
- TestFindByRawKey_NotFound — PASS（含预期 GORM 日志输出）
- TestGetKeyStatus — PASS
- TestListKeys — PASS
- TestDisableEnableKey — PASS

### Concerns

1. **CGO 编译问题**：系统仅有 32 位 MinGW GCC（`MinGW.org GCC-6.3.0-1`），无法编译 `mattn/go-sqlite3`（需 64 位）。改用纯 Go 替代方案 `github.com/glebarez/sqlite`（包装 `modernc.org/sqlite`），API 完全兼容，无需修改服务代码。

2. **`glebarez/sqlite` 已加入 go.mod**：`go.mod` 和 `go.sum` 增加了 `github.com/glebarez/sqlite`、`github.com/glebarez/go-sqlite`、`modernc.org/sqlite` 等间接依赖。若不想保留可考虑后续清理（仅测试依赖）。

3. **未测试 ConsumeKey**：`ConsumeKey` 使用 `gorm.Expr("CASE WHEN ... THEN ... END")` 和 `gorm.Expr("NOW()")`，其中 `NOW()` 在 SQLite 中不可用（需 `datetime('now')`）。按 brief 指示未覆盖该方法，但后续需注意：若想在 SQLite 下测试 ConsumeKey，服务代码本身可能需要适配。

4. **GORM 日志噪声**：`TestFindByRawKey_NotFound` 会触发 GORM 的 `record not found` 日志输出（红色），属正常行为，不影响测试结果。可考虑在测试中设置 `logger.Silent` 模式消除。
