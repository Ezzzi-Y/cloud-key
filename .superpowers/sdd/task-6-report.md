# Task 6: 安装 Log 依赖 — Report

**Status:** DONE_WITH_CONCERNS

## Commits

- `304589a` — `feat(deps): add Zap and Lumberjack for logging`

## Test Summary

N/A — This task only installs dependencies; no tests to run.

## Steps Completed

1. Added `go.uber.org/zap@v1.27.0` via `go get`
2. Added `gopkg.in/natefinch/lumberjack.v2@v2.2.1` via `go get`
3. Ran `go mod tidy`
4. Verified `go.mod` contains both dependencies
5. Committed `go.mod` and `go.sum`

## Concerns

1. **Brief typo in lumberjack package name:** The brief specified `gopkg.in/natefinish/lumberjack.v2` (with an "i" in "finish"), but the correct package path is `gopkg.in/natefinch/lumberjack.v2` (with an "ch" in "finch"). The task was completed using the correct package name.

2. **Dependencies are marked `// indirect`:** Both `zap` and `lumberjack` are listed as `// indirect` in `go.mod` because no Go code in the project currently imports them. They will be promoted to direct dependencies once code is added that imports these packages. Running `go mod tidy` before that code exists would drop them, so they were re-added after tidy for this commit.
