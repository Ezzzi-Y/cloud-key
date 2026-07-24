package middleware

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/handler"
	"CloudKey/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TenantBusinessGuard(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantIDI, exists := c.Get("tenant_id")
		if !exists {
			handler.Forbidden(c, errcode.CodeTenantNotFound, errcode.GetMessage(errcode.CodeTenantNotFound))
			c.Abort()
			return
		}
		tenantID := tenantIDI.(uint64)

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

		c.Next()
	}
}
