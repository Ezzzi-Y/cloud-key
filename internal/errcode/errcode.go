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

	// 租户相关 4001~4999
	CodeTenantExpired  = 4001
	CodeTenantDisabled = 4002
	CodeTenantNotFound = 4003

	// 权限相关 5001~5999
	CodeSuperAdminRequired  = 5001
	CodeTenantAdminRequired = 5002

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
	CodeServiceKeyInvalid:     "服务账号密钥无效",
	CodeTenantExpired:         "租户已到期，仅可查看统计数据",
	CodeTenantDisabled:        "租户已被禁用",
	CodeTenantNotFound:        "租户不存在",
	CodeSuperAdminRequired:    "需要系统管理员权限",
	CodeTenantAdminRequired:   "需要租户管理员权限",
	CodeInternalError:         "系统内部错误",
}

func GetMessage(code int) string {
	if msg, ok := codeMessages[code]; ok {
		return msg
	}
	return "未知错误"
}
