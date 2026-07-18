package middleware

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ServiceAuthMiddleware(svc *service.ServiceAccountService) gin.HandlerFunc {
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

		c.Set("service_account", account)
		c.Next()
	}
}
