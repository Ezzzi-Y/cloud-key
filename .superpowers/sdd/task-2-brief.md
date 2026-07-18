### Task 2: 错误码包

**Files:**
- Create: `internal/errcode/errcode.go`

**Interfaces:**
- Produces: 错误码常量 `CodeSuccess`, `CodeKeyNotFound` 等，供 handler 和 service 共用

- [ ] **Step 1: 创建 errcode 目录**

```bash
mkdir -p internal/errcode
```

- [ ] **Step 2: 编写 errcode.go**

```go
package errcode

const (
	CodeSuccess = 0

	// 卡密相关 1001~1999
	CodeKeyNotFound     = 1001
	CodeKeyDisabled     = 1002
	CodeKeyExhausted    = 1003
	CodeKeyInsufficient = 1004

	// 管理员认证相关 2001~2999
	CodeInvalidCredentials = 2001
	CodeTOTPFailed         = 2002
	CodeJWTInvalid         = 2003
	CodeForbidden          = 2004

	// 服务账号相关 3001~3999
	CodeServiceKeyInvalid = 3001

	// 系统 9999
	CodeInternalError = 9999
)

var codeMessages = map[int]string{
	CodeSuccess:            "success",
	CodeKeyNotFound:        "卡密不存在",
	CodeKeyDisabled:        "卡密已禁用",
	CodeKeyExhausted:       "卡密额度已用尽",
	CodeKeyInsufficient:    "扣减数量超过剩余额度",
	CodeInvalidCredentials: "管理员账号或密码错误",
	CodeTOTPFailed:         "TOTP 验证失败",
	CodeJWTInvalid:         "JWT Token 无效或已过期",
	CodeForbidden:          "无权限执行此操作",
	CodeServiceKeyInvalid:  "服务账号密钥无效",
	CodeInternalError:      "系统内部错误",
}

func GetMessage(code int) string {
	if msg, ok := codeMessages[code]; ok {
		return msg
	}
	return "未知错误"
}
```

- [ ] **Step 3: 格式化并编译**

```bash
gofmt -w internal/errcode/errcode.go
go build ./internal/errcode/
```

- [ ] **Step 4: 提交**

```bash
git add internal/errcode/
git commit -m "feat(errcode): add shared error code constants"
```

---

