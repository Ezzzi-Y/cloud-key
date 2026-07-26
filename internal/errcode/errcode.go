package errcode

const (
	CodeSuccess = 0

	// 卡密相关 1001~1999
	CodeKeyNotFound       = 1001
	CodeKeyDisabled       = 1002
	CodeKeyExhausted      = 1003
	CodeKeyInsufficient   = 1004
	CodeInvalidAdjustment = 1005 // 额度调整参数无效
	CodeKeyExpired            = 1006
	CodeInvalidConsumeAmount  = 1007 // 消费数量参数无效

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
	CodeKeyConfigInvalid = 4004

	// 权限相关 5001~5999
	CodeSuperAdminRequired  = 5001
	CodeTenantAdminRequired = 5002

	// 登录安全相关 6001~6099
	CodeRateLimit      = 6001
	CodeAccountLocked  = 6002
	CodePreAuthInvalid = 6003

	// 系统 9001~9999
	CodeRouteNotFound = 9001
	CodeInternalError = 9999
)

var codeMessages = map[int]string{
	CodeSuccess:            "success",
	CodeKeyNotFound:        "卡密不存在",
	CodeKeyDisabled:        "卡密已禁用",
	CodeKeyExhausted:       "卡密额度已用尽",
	CodeKeyExpired:         "卡密已过期",
	CodeKeyInsufficient:       "扣减数量超过剩余额度",
	CodeInvalidConsumeAmount:  "消费数量必须大于 0",
	CodeInvalidAdjustment:  "额度调整参数无效",
	CodeInvalidCredentials: "管理员账号或密码错误",
	CodeTOTPFailed:         "TOTP 验证失败",
	CodeJWTInvalid:         "JWT Token 无效或已过期",
	CodeForbidden:          "无权限执行此操作",
	CodeServiceKeyInvalid:     "服务账号密钥无效",
	CodeTenantExpired:         "租户已到期，仅可查看统计数据",
	CodeTenantDisabled:        "租户已被禁用",
	CodeTenantNotFound:        "租户不存在",
	CodeKeyConfigInvalid:        "Key 配置参数无效",
	CodeSuperAdminRequired:    "需要系统管理员权限",
	CodeTenantAdminRequired:   "需要租户管理员权限",
	CodeRateLimit:             "请求过于频繁，请稍后再试",
	CodeAccountLocked:         "账号已被锁定，请稍后再试",
	CodePreAuthInvalid:        "认证凭证无效或已过期，请重新登录",
	CodeInternalError:         "系统内部错误",
}

func GetMessage(code int) string {
	if msg, ok := codeMessages[code]; ok {
		return msg
	}
	return "未知错误"
}
