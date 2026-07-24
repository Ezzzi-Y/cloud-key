package middleware

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/handler"
	"CloudKey/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RequireTenantAdmin 验证租户管理员权限
func RequireTenantAdmin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleI, exists := c.Get("role")
		if !exists {
			handler.Forbidden(c, errcode.CodeTenantAdminRequired, errcode.GetMessage(errcode.CodeTenantAdminRequired))
			c.Abort()
			return
		}
		role, ok := roleI.(model.UserRole)
		if !ok || role != model.RoleTenantAdmin {
			handler.Forbidden(c, errcode.CodeTenantAdminRequired, errcode.GetMessage(errcode.CodeTenantAdminRequired))
			c.Abort()
			return
		}

		// 从 JWT claims 中获取 tenant_id
		tenantIDRaw, exists := c.Get("tenant_id")
		if !exists {
			handler.Forbidden(c, errcode.CodeTenantAdminRequired, errcode.GetMessage(errcode.CodeTenantAdminRequired))
			c.Abort()
			return
		}
		tenantID, ok := tenantIDRaw.(uint64)
		if !ok || tenantID == 0 {
			handler.NotFound(c, errcode.CodeTenantNotFound, errcode.GetMessage(errcode.CodeTenantNotFound))
			c.Abort()
			return
		}

		// 查询租户，验证租户存在且未过期/禁用
		var tenant model.Tenant
		if err := db.First(&tenant, tenantID).Error; err != nil {
			handler.NotFound(c, errcode.CodeTenantNotFound, errcode.GetMessage(errcode.CodeTenantNotFound))
			c.Abort()
			return
		}

		if tenant.Status == model.TenantStatusExpired {
			handler.Forbidden(c, errcode.CodeTenantExpired, errcode.GetMessage(errcode.CodeTenantExpired))
			c.Abort()
			return
		}

		if tenant.Status == model.TenantStatusDisabled {
			handler.Forbidden(c, errcode.CodeTenantDisabled, errcode.GetMessage(errcode.CodeTenantDisabled))
			c.Abort()
			return
		}

		// 将完整的租户信息存入上下文，供后续使用
		c.Set("tenant", &tenant)
		c.Next()
	}
}
