package middleware

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/handler"
	"CloudKey/internal/model"

	"github.com/gin-gonic/gin"
)

func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleI, exists := c.Get("role")
		if !exists {
			handler.Forbidden(c, errcode.CodeSuperAdminRequired, errcode.GetMessage(errcode.CodeSuperAdminRequired))
			c.Abort()
			return
		}
		role, ok := roleI.(model.UserRole)
		if !ok || role != model.RoleSuperAdmin {
			handler.Forbidden(c, errcode.CodeSuperAdminRequired, errcode.GetMessage(errcode.CodeSuperAdminRequired))
			c.Abort()
			return
		}
		c.Next()
	}
}
