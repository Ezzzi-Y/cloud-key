### Task 3: 更新错误码 + 改造 JWT Claims（中间件层）

**Files:**
- Modify: `internal/errcode/errcode.go`
- Rewrite: `internal/middleware/auth.go`
- Create: `internal/middleware/super_admin.go`
- Create: `internal/middleware/tenant_admin.go`
- Create: `internal/middleware/tenant_business.go`
- Rewrite: `internal/middleware/service_auth.go`

**Interfaces:**
- Consumes: `model.User.Role`, `model.Tenant.Status`
- Produces: `AuthMiddleware(jwtSecret) -> gin.HandlerFunc`; `RequireSuperAdmin() -> gin.HandlerFunc`; `RequireTenantAdmin() -> gin.HandlerFunc`; `TenantBusinessGuard() -> gin.HandlerFunc`; `ServiceAuthMiddleware(svc) -> gin.HandlerFunc`
- Context keys: `"user_id"` (uint64), `"username"` (string), `"role"` (model.UserRole), `"tenant_id"` (uint64)

- [ ] **Step 1: 更新 `internal/errcode/errcode.go` — 新增租户相关错误码**

在 const block 末尾新增:
```go
// 租户相关 4001~4999
CodeTenantExpired  = 4001
CodeTenantDisabled = 4002
CodeTenantNotFound = 4003

// 权限相关 5001~5999
CodeSuperAdminRequired = 5001
CodeTenantAdminRequired = 5002
```

在 codeMessages map 中新增:
```go
CodeTenantExpired:         "租户已到期，仅可查看统计数据",
CodeTenantDisabled:        "租户已被禁用",
CodeTenantNotFound:        "租户不存在",
CodeSuperAdminRequired:    "需要系统管理员权限",
CodeTenantAdminRequired:   "需要租户管理员权限",
```

- [ ] **Step 2: 重写 `internal/middleware/auth.go` — 新的 JWT Claims + AuthMiddleware**

```go
package middleware

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint64         `json:"user_id"`
	Username string         `json:"username"`
	Role     model.UserRole `json:"role"`
	TenantID *uint64        `json:"tenant_id"`
	jwt.RegisteredClaims
}

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": errcode.CodeJWTInvalid, "message": errcode.GetMessage(errcode.CodeJWTInvalid), "data": nil})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": errcode.CodeJWTInvalid, "message": errcode.GetMessage(errcode.CodeJWTInvalid), "data": nil})
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"code": errcode.CodeJWTInvalid, "message": errcode.GetMessage(errcode.CodeJWTInvalid), "data": nil})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		if claims.TenantID != nil {
			c.Set("tenant_id", *claims.TenantID)
		}
		c.Next()
	}
}
```

- [ ] **Step 3: 创建 `internal/middleware/super_admin.go`**

```go
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
```

- [ ] **Step 4: 创建 `internal/middleware/tenant_admin.go`**

```go
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
```

- [ ] **Step 5: 创建 `internal/middleware/tenant_business.go` — 业务操作守卫**

```go
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
			// 非租户管理员，跳过（可能是 super_admin，但它不应调用租户业务接口）
			c.Next()
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
```

- [ ] **Step 6: 重写 `internal/middleware/service_auth.go` — 加租户状态检查**

```go
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
```

- [ ] **Step 7: 验证编译通过**

Run: `cd D:/MyGoProject/CloudKey && go build ./...`
Expected: 编译通过

- [ ] **Step 8: Commit**

```bash
git add internal/errcode/errcode.go internal/middleware/
git commit -m "feat: add new JWT claims with role+tenant_id, super/tenant/business middlewares"
```

---

