# Task 3 Report: Config LoadConfig 函数

**Status:** DONE

**Commits:**
- `0578742` feat(config): implement LoadConfig with Viper

**Test Summary:**
- Code compiles successfully (`go build ./internal/config/`)
- No unit tests required for this task (per brief)

**Changes Made:**
- Modified `internal/config/config.go`:
  - Added `github.com/spf13/viper` import
  - Added `LoadConfig(path string) (*AppConfig, error)` function

**Concerns:** None