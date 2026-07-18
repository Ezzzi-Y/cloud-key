### Task 10: 卡密服务 — 管理操作

**Status:** DONE

**Commit:**
- `6a6b7d1` feat(service): add key management operations (list, update, disable, enable, delete, export)

**Build:** `go build ./internal/service/` — clean, no errors

**Changes:**
- Added `KeyListQuery` struct for paginated list queries with status filter and search
- Added `GetKeyDetail(id)` — fetch single key by ID
- Added `ListKeys(query)` — paginated list with optional status filter and alias/suffix search
- Added `ListKeysByCreatedBy(createdBy, page, pageSize)` — paginated list filtered by creator
- Added `UpdateKey(id, req)` — partial update of alias and/or remaining_amount (auto-resets status from `used` to `unused` if amount > 0)
- Added `DisableKey(id)` — set status to `disabled`
- Added `EnableKey(id)` — set status to `unused`
- Added `DeleteKey(id)` — hard delete
- Added `ExportKeys()` — return all keys ordered by created_at DESC
