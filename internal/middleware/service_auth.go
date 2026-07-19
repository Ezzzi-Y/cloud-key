package middleware

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"CloudKey/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ServiceAuthMiddleware(svc *service.ServiceAccountService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceKey := c.GetHeader("X-Service-Key")
		if serviceKey == "" {
			c.JSON(http.StatusOK, gin.H{"code": errcode.CodeServiceKeyInvalid, "message": errcode.GetMessage(errcode.CodeServiceKeyInvalid), "data": nil})
			c.Abort()
			return
		}

		account, err := svc.ValidateServiceKey(serviceKey)
		if err != nil || account == nil {
			c.JSON(http.StatusOK, gin.H{"code": errcode.CodeServiceKeyInvalid, "message": errcode.GetMessage(errcode.CodeServiceKeyInvalid), "data": nil})
			c.Abort()
			return
		}

		// 检查租户状态
		var tenant model.Tenant
		if err := db.First(&tenant, account.TenantID).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": errcode.CodeTenantNotFound, "message": errcode.GetMessage(errcode.CodeTenantNotFound), "data": nil})
			c.Abort()
			return
		}
		if tenant.Status != model.TenantStatusActive {
			c.JSON(http.StatusOK, gin.H{"code": errcode.CodeTenantExpired, "message": errcode.GetMessage(errcode.CodeTenantExpired), "data": nil})
			c.Abort()
			return
		}

		c.Set("service_account", account)
		c.Set("tenant_id", account.TenantID)
		c.Next()
	}
}
