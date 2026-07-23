package middleware

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TenantBusinessGuard(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantIDI, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeTenantNotFound, "message": "租户信息缺失，拒绝访问", "data": nil})
			c.Abort()
			return
		}
		tenantID := tenantIDI.(uint64)

		var tenant model.Tenant
		if err := db.First(&tenant, tenantID).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeTenantNotFound, "message": errcode.GetMessage(errcode.CodeTenantNotFound), "data": nil})
			c.Abort()
			return
		}

		if tenant.Status == model.TenantStatusExpired {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeTenantExpired, "message": errcode.GetMessage(errcode.CodeTenantExpired), "data": nil})
			c.Abort()
			return
		}
		if tenant.Status == model.TenantStatusDisabled {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeTenantDisabled, "message": errcode.GetMessage(errcode.CodeTenantDisabled), "data": nil})
			c.Abort()
			return
		}

		c.Next()
	}
}
