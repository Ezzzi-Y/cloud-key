## Task 2: 错误码包 - Report

### What was done

Created `internal/errcode/errcode.go` with all error code constants specified in the task brief:

- `CodeSuccess` (0)
- Key-related codes: `CodeKeyNotFound` (1001), `CodeKeyDisabled` (1002), `CodeKeyExhausted` (1003), `CodeKeyInsufficient` (1004)
- Auth-related codes: `CodeInvalidCredentials` (2001), `CodeTOTPFailed` (2002), `CodeJWTInvalid` (2003), `CodeForbidden` (2004)
- Service account code: `CodeServiceKeyInvalid` (3001)
- System code: `CodeInternalError` (9999)

Also included:
- `codeMessages` map mapping each code to its human-readable Chinese message
- `GetMessage(code int) string` function with fallback to "未知错误" for unknown codes

### Verification

- `go build ./internal/errcode/` passed with no errors.
- `gofmt` formatting was applied and no changes were needed (code was already correctly formatted).

### Commit

- **SHA:** `0553df7`
- **Message:** `feat(errcode): add shared error code constants`

### Self-review notes

- The implementation matches the task spec exactly, character-for-character.
- The package is importable by both `handler` and `service` layers without circular dependencies.
- No external dependencies needed — this is a pure constants/utility package.
- Code ranges are clearly documented with comments (1001~1999, 2001~2999, 3001~3999, 9999).
