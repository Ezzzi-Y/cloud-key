package middleware

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RequireTenantAdmin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleI, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeTenantAdminRequired, "message": errcode.GetMessage(errcode.CodeTenantAdminRequired), "data": nil})
			c.Abort()
			return
		}
		role, ok := roleI.(model.UserRole)
		if !ok || role != model.RoleTenantAdmin {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeTenantAdminRequired, "message": errcode.GetMessage(errcode.CodeTenantAdminRequired), "data": nil})
			c.Abort()
			return
		}

		tenantIDI, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeTenantAdminRequired, "message": "租户信息缺失", "data": nil})
			c.Abort()
			return
		}
		tenantID := tenantIDI.(uint64)

		// 检查租户是否 disabled
		var tenant model.Tenant
		if err := db.First(&tenant, tenantID).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeTenantNotFound, "message": errcode.GetMessage(errcode.CodeTenantNotFound), "data": nil})
			c.Abort()
			return
		}
		if tenant.Status == model.TenantStatusDisabled {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeTenantDisabled, "message": errcode.GetMessage(errcode.CodeTenantDisabled), "data": nil})
			c.Abort()
			return
		}
		// expired 仍放行，由 TenantBusinessGuard 控制业务操作

		c.Next()
	}
}
