# Task 8: Log 测试

**Status:** DONE_WITH_CONCERNS

## Commits

- `6d20ff3` - `test(log): add unit tests for InitLogger`

## Test Summary

| Test | Result |
|------|--------|
| TestInitLogger_Console | PASS |
| TestInitLogger_JSON | PASS |
| TestInitLogger_File | PASS |
| TestInitLogger_InvalidLevel | PASS |

**Total: 4 passed, 0 failed**

## Changes

Created `internal/log/logger_test.go` with 4 test cases covering:
- Console output with info level
- JSON output with debug level (all log levels exercised)
- File output with lumberjack rotation config
- Invalid level fallback to default (info)

Modified `internal/log/logger.go`:
- Added `lumberjackIO` package variable to track the lumberjack logger instance
- Added `Close()` function to close the underlying file handle

## Concerns

The brief's original test code for `TestInitLogger_File` caused a `TempDir` cleanup failure on Windows because the lumberjack logger keeps the file handle open. `t.TempDir()` attempts to `RemoveAll` the temp directory on test exit, which fails when a process still holds the file. I resolved this by:

1. Adding a `Close()` function to the logger module that closes the lumberjack writer
2. Calling `defer Close()` in the file test before `TempDir` cleanup runs

This is a minor deviation from the brief. On Linux the same issue may manifest as an unlink error. The fix is non-breaking and improves the logger API by allowing callers to cleanly shut down file-backed loggers.
