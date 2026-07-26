package middleware

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/handler"
	"CloudKey/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TenantBusinessGuard 拦截 expired / disabled 租户的写操作
// 优先从 context 读取 RequireTenantAdmin 已加载的 tenant，避免重复查 DB
func TenantBusinessGuard(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tenant *model.Tenant

		// 优先从 context 读取（RequireTenantAdmin 或 ServiceAuthMiddleware 已加载）
		if tenantI, exists := c.Get("tenant"); exists {
			tenant = tenantI.(*model.Tenant)
		} else {
			// fallback：从 DB 查询
			tenantIDI, exists := c.Get("tenant_id")
			if !exists {
				handler.Forbidden(c, errcode.CodeTenantNotFound, errcode.GetMessage(errcode.CodeTenantNotFound))
				c.Abort()
				return
			}
			tenantID := tenantIDI.(uint64)

			var t model.Tenant
			if err := db.First(&t, tenantID).Error; err != nil {
				handler.NotFound(c, errcode.CodeTenantNotFound, errcode.GetMessage(errcode.CodeTenantNotFound))
				c.Abort()
				return
			}
			tenant = &t
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

		c.Next()
	}
}
