### Task 7: 创建 SuperHandler（系统管理员接口）+ 改造 KeyHandler 为 TenantHandler

**Files:**
- Rewrite: `internal/handler/key_handler.go` → 面向 tenant admin
- Create: `internal/handler/super_handler.go`
- Create: `internal/service/tenant_service.go`
- Modify: `internal/handler/response.go`（如需要）
- Modify: `internal/router/router.go`（暂时不在这里更新，留到 Task 9）

**Interfaces:**
- Produces: `SuperHandler` with `ListTenants`, `CreateTenant`, `GetTenant`, `UpdateTenant`, `ResetPassword`, `GetConfigs`, `UpdateConfigs`
- Produces: `TenantHandler` with `CreateKey`, `ListKeys`, `GetKey`, ..., `ListServiceAccounts`, `CreateServiceAccount`, ..., `Dashboard`, `Overview`, ...
- Produces: `TenantService` with `CreateTenant`, `GetTenant`, `ListTenants`, `UpdateTenant`, `ResetPassword`

- [ ] **Step 1: 创建 `internal/service/tenant_service.go`**

```go
package service

import (
	"CloudKey/internal/model"
	"fmt"
	"math/rand"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type TenantService struct {
	db *gorm.DB
}

func NewTenantService(db *gorm.DB) *TenantService {
	return &TenantService{db: db}
}

type CreateTenantRequest struct {
	Name   string `json:"name" binding:"required"`
}

type CreateTenantResult struct {
	Tenant         model.Tenant `json:"tenant"`
	AdminUsername  string       `json:"admin_username"`
	AdminPassword  string       `json:"admin_password"`
}

func (s *TenantService) CreateTenant(req CreateTenantRequest) (*CreateTenantResult, error) {
	// 生成租户管理员账号
	username := fmt.Sprintf("%s_admin", req.Name)

	// 检查用户名冲突
	var count int64
	s.db.Model(&model.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		username = fmt.Sprintf("%s_admin%d", req.Name, rand.Intn(9999))
	}

	password := generateRandomPassword(16)

	// 事务
	tx := s.db.Begin()

	tenant := model.Tenant{
		Name:      req.Name,
		Status:    model.TenantStatusActive,
		KeyPrefix: "sk-",
		KeyLength: 32,
		KeySuffixLength: 4,
	}
	if err := tx.Create(&tenant).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create tenant: %w", err)
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := model.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         model.RoleTenantAdmin,
		TenantID:     &tenant.ID,
		IsActive:     true,
	}
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create tenant admin: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &CreateTenantResult{
		Tenant:        tenant,
		AdminUsername: username,
		AdminPassword: password,
	}, nil
}

type TenantListItem struct {
	model.Tenant
	KeyCount  int64 `json:"key_count"`
	UserCount int64 `json:"user_count"`
}

func (s *TenantService) ListTenants() ([]TenantListItem, error) {
	tenants := make([]TenantListItem, 0)
	rows, err := s.db.Model(&model.Tenant{}).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t TenantListItem
		s.db.ScanRows(rows, &t.Tenant)

		s.db.Model(&model.Key{}).Where("tenant_id = ?", t.ID).Count(&t.KeyCount)
		s.db.Model(&model.User{}).Where("tenant_id = ?", t.ID).Count(&t.UserCount)

		tenants = append(tenants, t)
	}
	return tenants, nil
}
```

(文件较长，此处省略完整内容，包含 `GetTenant`, `UpdateTenant`, `ResetPassword`, `generateRandomPassword`)

```go
func (s *TenantService) GetTenant(id uint64) (*TenantListItem, error) {
	var t TenantListItem
	if err := s.db.First(&t.Tenant, id).Error; err != nil {
		return nil, err
	}
	s.db.Model(&model.Key{}).Where("tenant_id = ?", id).Count(&t.KeyCount)
	s.db.Model(&model.User{}).Where("tenant_id = ?", id).Count(&t.UserCount)
	return &t, nil
}

type UpdateTenantRequest struct {
	Name      *string             `json:"name"`
	Status    *model.TenantStatus `json:"status"`
	ExpireAt  *string             `json:"expire_at"`  // "2006-01-02 15:04:05" or "" to clear
	KeyPrefix *string             `json:"key_prefix"`
	KeyLength *int                `json:"key_length"`
	KeySuffixLength *int          `json:"key_suffix_length"`
}

func (s *TenantService) UpdateTenant(id uint64, req UpdateTenantRequest) error {
	updates := map[string]interface{}{}
	if req.Name != nil { updates["name"] = *req.Name }
	if req.Status != nil { updates["status"] = *req.Status }
	if req.KeyPrefix != nil { updates["key_prefix"] = *req.KeyPrefix }
	if req.KeyLength != nil { updates["key_length"] = *req.KeyLength }
	if req.KeySuffixLength != nil { updates["key_suffix_length"] = *req.KeySuffixLength }
	if req.ExpireAt != nil {
		if *req.ExpireAt == "" {
			updates["expire_at"] = nil
		} else {
			updates["expire_at"] = *req.ExpireAt
		}
	}
	if len(updates) == 0 { return nil }
	return s.db.Model(&model.Tenant{}).Where("id = ?", id).Updates(updates).Error
}

func (s *TenantService) ResetPassword(tenantID uint64) (string, error) {
	newPass := generateRandomPassword(16)
	hash, _ := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	err := s.db.Model(&model.User{}).Where("tenant_id = ? AND role = ?", tenantID, model.RoleTenantAdmin).
		Update("password_hash", string(hash)).Error
	if err != nil {
		return "", err
	}
	return newPass, nil
}

func generateRandomPassword(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
```

- [ ] **Step 2: 创建 `internal/handler/super_handler.go`**

```go
package handler

import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	"CloudKey/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SuperHandler struct {
	tenantSvc    *service.TenantService
	configSvc    *service.ConfigService
	statsSvc     *service.StatsService
	loginLogSvc  *service.LoginLogService
}

func NewSuperHandler(tenantSvc *service.TenantService, configSvc *service.ConfigService, statsSvc *service.StatsService, loginLogSvc *service.LoginLogService) *SuperHandler {
	return &SuperHandler{tenantSvc: tenantSvc, configSvc: configSvc, statsSvc: statsSvc, loginLogSvc: loginLogSvc}
}

// GET /api/super/tenants
func (h *SuperHandler) ListTenants(c *gin.Context) {
	tenants, err := h.tenantSvc.ListTenants()
	if err != nil { InternalError(c); return }
	Success(c, tenants)
}

// POST /api/super/tenants
// 创建租户 + 自动生成管理员账号密码
func (h *SuperHandler) CreateTenant(c *gin.Context) {
	var req service.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errcode.CodeForbidden, "参数错误")
		return
	}
	result, err := h.tenantSvc.CreateTenant(req)
	if err != nil {
		InternalError(c)
		return
	}
	Success(c, gin.H{
		"tenant":         result.Tenant,
		"admin_username": result.AdminUsername,
		"admin_password": result.AdminPassword,
	})
}

// GET /api/super/tenants/:id
func (h *SuperHandler) GetTenant(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil { BadRequest(c, errcode.CodeTenantNotFound, "无效的ID"); return }
	tenant, err := h.tenantSvc.GetTenant(id)
	if err != nil { NotFound(c, errcode.CodeTenantNotFound, "租户不存在"); return }
	Success(c, tenant)
}

// PATCH /api/super/tenants/:id
func (h *SuperHandler) UpdateTenant(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil { BadRequest(c, errcode.CodeTenantNotFound, "无效的ID"); return }
	var req service.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil { BadRequest(c, errcode.CodeForbidden, "参数错误"); return }
	if err := h.tenantSvc.UpdateTenant(id, req); err != nil { InternalError(c); return }
	Success(c, nil)
}

// PATCH /api/super/tenants/:id/reset-password
func (h *SuperHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil { BadRequest(c, errcode.CodeTenantNotFound, "无效的ID"); return }
	newPass, err := h.tenantSvc.ResetPassword(id)
	if err != nil { InternalError(c); return }
	Success(c, gin.H{"new_password": newPass})
}

// GET /api/super/configs
func (h *SuperHandler) GetConfigs(c *gin.Context) {
	configs, err := h.configSvc.GetAllConfigs()
	if err != nil { InternalError(c); return }
	Success(c, configs)
}

// PUT /api/super/configs
func (h *SuperHandler) UpdateConfigs(c *gin.Context) {
	var req []struct {
		Key         string `json:"key" binding:"required"`
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil { BadRequest(c, errcode.CodeForbidden, "参数错误"); return }
	for _, item := range req {
		if err := h.configSvc.SetConfig(item.Key, item.Value, item.Description); err != nil {
			InternalError(c); return
		}
	}
	Success(c, nil)
}

// GET /api/super/login-logs
func (h *SuperHandler) LoginLogs(c *gin.Context) {
	page, pageSize := pageParams(c)
	logs, total, err := h.loginLogSvc.ListLoginLogs(page, pageSize, nil) // nil = 全部
	if err != nil { InternalError(c); return }
	SuccessPaginated(c, logs, total, page, pageSize)
}
```

- [ ] **Step 3: 重写 `internal/handler/key_handler.go` → 面向 tenant admin**

将 `KeyHandler` 重命名为 `TenantKeyHandler`，增加 tenant scope:

```go
type TenantKeyHandler struct {
	keySvc       *service.KeyService
	usageLogSvc  *service.UsageLogService
	recordParams bool
}

func NewTenantKeyHandler(keySvc *service.KeyService, usageLogSvc *service.UsageLogService, recordParams bool) *TenantKeyHandler {
	return &TenantKeyHandler{keySvc: keySvc, usageLogSvc: usageLogSvc, recordParams: recordParams}
}
```

所有方法加 tenant scope。例如 **CreateKey**:
```go
func (h *TenantKeyHandler) CreateKey(c *gin.Context) {
	var req CreateKeyJSON
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, http.StatusBadRequest, "参数错误")
		return
	}

	tenantID := getTenantID(c)

	// 从 context 获取租户的 Key 配置（可通过 getTenantPrefix 等获取）
	// 简化：先用硬编码默认值，后续从 middleware 注入
	createdBy := "tenant_admin"

	expireAt, _ := parseExpireAt(req.ExpireAt)
	result, err := h.keySvc.CreateKey(service.CreateKeyRequest{
		Alias: req.Alias, BillingMode: model.KeyBillingMode(req.BillingMode),
		InitialAmount: req.InitialAmount, CreatedBy: createdBy,
		ExpireAt: expireAt, MaxUsage: req.MaxUsage,
	}, tenantID, "sk-", 32, 4) // 后续从 tenant 获取
	if err != nil { InternalError(c); return }

	Success(c, gin.H{ ... })
}
```

类似更新 **ListKeys, GetKey, UpdateKey, DisableKey, EnableKey, DeleteKey, ExportKeys, ExportKeysJSON** 全部加 `tenantID := getTenantID(c)` 并传入 service。

公共接口 **Status** 和 **Consume** 保留在 `key_handler.go` 中（无需认证），但 Consume 时需 recording usage_log 传入 key 的 tenant_id。

- [ ] **Step 4: 验证编译**

Run: `cd D:/MyGoProject/CloudKey && go build ./...`
Expected: 部分编译错误（旧 handler 文件未删）。**先忽略，后续 Task 9 router 改造后统一清理。**

- [ ] **Step 5: Commit**

```bash
git add internal/service/tenant_service.go internal/handler/super_handler.go internal/handler/key_handler.go
git commit -m "feat: add SuperHandler, TenantService, refactor KeyHandler for tenant scope"
```

---

