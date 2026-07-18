### Task 3 Report: 卡密数据模型

**Status:** DONE

**Commit:** b4354e0 — feat(model): add Key GORM model with optimistic lock

**What was done:**
- Created `internal/model/` directory (new)
- Created `internal/model/key.go` containing:
  - `KeyBillingMode` type with `BillingModeCount` and `BillingModeCredit` constants
  - `KeyStatus` type with `KeyStatusUnused`, `KeyStatusUsed`, `KeyStatusDisabled`, `KeyStatusExpired` constants
  - `Key` struct — full GORM model with table name `keys`, 14 fields including `Version` for optimistic locking
  - `TableName()` method returning `"keys"`
  - `IsUsable()` helper: checks status is unused and remaining amount > 0
  - `CanDeduct(amount)` helper: checks usability and sufficient balance
- `gofmt -w` applied — no formatting changes needed
- `go build ./internal/model/` — compiled successfully, no errors

**Build/test:** `go build ./internal/model/` passed cleanly.

**Concerns:** None.
