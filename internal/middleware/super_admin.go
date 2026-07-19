package middleware

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleI, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeSuperAdminRequired, "message": errcode.GetMessage(errcode.CodeSuperAdminRequired), "data": nil})
			c.Abort()
			return
		}
		role, ok := roleI.(model.UserRole)
		if !ok || role != model.RoleSuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{"code": errcode.CodeSuperAdminRequired, "message": errcode.GetMessage(errcode.CodeSuperAdminRequired), "data": nil})
			c.Abort()
			return
		}
		c.Next()
	}
}
