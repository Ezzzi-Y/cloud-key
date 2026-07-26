package middleware

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/handler"
	"CloudKey/internal/model"

	"github.com/gin-gonic/gin"
)

// TenantDisabledGuard 拦截已禁用的租户（expired 放行）
// 必须放在 RequireTenantAdmin 之后使用，依赖其写入 context 的 tenant
func TenantDisabledGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantI, exists := c.Get("tenant")
		if !exists {
			handler.Forbidden(c, errcode.CodeTenantNotFound, errcode.GetMessage(errcode.CodeTenantNotFound))
			c.Abort()
			return
		}

		tenant := tenantI.(*model.Tenant)
		if tenant.Status == model.TenantStatusDisabled {
			handler.Forbidden(c, errcode.CodeTenantDisabled, errcode.GetMessage(errcode.CodeTenantDisabled))
			c.Abort()
			return
		}

		c.Next()
	}
}
