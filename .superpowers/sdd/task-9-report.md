### Task 9: 卡密服务 — 扣减（并发安全）

**Status:** DONE

**Commits:**
- `ff7ecbd` feat(service): add ConsumeKey with transaction + optimistic locking

**Build summary:**
- `gofmt -w internal/service/key_service.go` — no changes needed
- `go build ./internal/service/` — success, clean build

**Changes:**
- Added `CloudKey/internal/errcode` to the import block in `internal/service/key_service.go`
- Appended `ConsumeResult` struct (RemainingAmount, Status, Exhausted) at line 142
- Appended `ConsumeKey(rawKey, amount)` method at line 148 implementing:
  - Pre-transaction validation: key existence, status checks (disabled/used), `CanDeduct` guard
  - Database transaction with optimistic locking via `WHERE id = ? AND version = ?`
  - Atomic update of remaining_amount, version increment, conditional status flip to 'used'
  - `NOW()` timestamp for used_at
  - Single retry on optimistic lock conflict (RowsAffected == 0)
  - Panic-recovery defer with rollback

**Concerns:** None

---

### Task 9 Fix Report: ConsumeKey Code Review Issues

**Status:** DONE

**Commits:**
- `45bb308` fix(service): validate amount, cap retry, conditional used_at in ConsumeKey

**Build summary:**
- `gofmt -w internal/service/key_service.go` — clean, no reformatting needed
- `go build ./internal/service/` — success, no errors

**Fixes applied:**

| Issue | Severity | Description | Fix |
|-------|----------|-------------|-----|
| C1 | Critical | No input validation on `amount`; zero/negative values reached SQL, negative amounts would increase balance | Added `if amount <= 0` guard returning `fmt.Errorf("invalid amount: %d", amount)` before `CanDeduct` check (line 162) |
| I1 | Important | Unbounded recursive retry on optimistic lock conflict (`RowsAffected == 0`) could stack-overflow under contention | Replaced recursive `s.ConsumeKey(rawKey, amount)` call with `for retry := 0; retry < 3; retry++` iterative loop; returns `fmt.Errorf("concurrency conflict after %d retries", maxRetries)` on exhaustion |
| I2 | Important | `used_at = NOW()` set unconditionally on every deduction, even when key is not exhausted | Changed to `gorm.Expr("CASE WHEN remaining_amount - ? <= 0 THEN NOW() ELSE used_at END", amount)` so timestamp is only set when remaining balance reaches zero |

**Additional improvement:** Removed bare `defer` panic-recovery in favour of explicit `tx.Error` check on `Begin()` inside each iteration, making transaction lifecycle clearer.

---

### Task 9 Fix Report: ConsumeKey Retry Loop Dead Code

**Status:** DONE

**Commit:** `36525ae` fix(service): re-fetch key on retry in ConsumeKey

**Build summary:**
- `gofmt -w internal/service/key_service.go` — clean, no reformatting needed
- `go build ./internal/service/` — success, no errors

**Bug fixed:**

| Issue | Severity | Description | Fix |
|-------|----------|-------------|-----|
| R1 | Critical | Retry loop never re-fetched the key; stale `key.Version` caused every retry to also fail `RowsAffected == 0`, making the retry mechanism dead code | Moved key fetch (`FindByRawKey`), status checks (`KeyStatusDisabled`, `KeyStatusUsed`), and `CanDeduct` guard inside the `for retry` loop so each iteration operates on a fresh key version |

**Structural change:** The `ConsumeKey` method was restructured so that:
- `amount <= 0` validation remains outside the loop (constant, no re-check needed)
- Key fetch via `FindByRawKey(rawKey)` + all status/amount validation is performed at the top of each retry iteration
- On `RowsAffected == 0`, `tx.Rollback()` + `continue` now correctly triggers a fresh key fetch on the next iteration
- After exhausting all retries, returns `fmt.Errorf("concurrency conflict after %d retries", maxRetries)`
