# CloudKey SaaS 多租户架构设计

> 日期：2026-07-19
> 状态：设计定稿
> 范围：将 CloudKey 从单租户服务账号系统改造为多租户 SaaS 平台

## 1. 背景与目标

CloudKey 当前是一个单租户的 API Key / 卡密管理平台，只有一个管理员体系和一套服务账号。目标是改造为 SaaS 多租户系统：

- 每个**租户**代表一个应用，拥有独立的 Key 空间和服务账号
- **系统管理员**管理平台级事务（租户 CRUD、平台配置、聚合统计）
- **租户管理员**管理自己的全部业务（Key、服务账号、统计、导出）
- 统一登录入口，后端判断角色，前端加载对应页面
- 前端后续迁移到 React SPA，由 Go 静态服务

## 2. 数据模型

### 2.1 tenants 表（新增）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK AUTO_INCREMENT | |
| name | VARCHAR(100) NOT NULL UNIQUE | 应用名称 |
| status | VARCHAR(20) DEFAULT 'active' | active / expired / disabled |
| expire_at | DATETIME NULLABLE | 到期时间，NULL 表示永不过期 |
| key_prefix | VARCHAR(20) DEFAULT 'sk-' | 租户自定义 Key 前缀 |
| key_length | INT DEFAULT 32 | Key 长度 |
| key_suffix_length | INT DEFAULT 4 | 后缀显示长度 |
| created_at | DATETIME | |
| updated_at | DATETIME | |

### 2.2 users 表（替代 admins）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT PK AUTO_INCREMENT | |
| username | VARCHAR(100) NOT NULL UNIQUE | 登录名 |
| password_hash | VARCHAR(255) NOT NULL | bcrypt 哈希 |
| totp_secret | VARCHAR(255) | TOTP 密钥 |
| totp_setup | BOOLEAN DEFAULT FALSE | 是否已完成 TOTP 设置 |
| role | VARCHAR(20) NOT NULL | 'super_admin' / 'tenant_admin' |
| tenant_id | BIGINT NULLABLE | super_admin 为 NULL |
| is_active | BOOLEAN DEFAULT TRUE | |
| created_at | DATETIME | |
| updated_at | DATETIME | |

### 2.3 keys 表（改造）

新增字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| tenant_id | BIGINT NOT NULL | 归属租户，加 INDEX |

其余字段不变（id, alias, key_hash, key_prefix, key_suffix, billing_mode, initial_amount, remaining_amount, version, status, created_by, created_at, updated_at, used_at, expire_at, max_usage）。

### 2.4 service_accounts 表（改造）

新增字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| tenant_id | BIGINT NOT NULL | 归属租户，加 INDEX |

其余字段不变。

### 2.5 usage_logs 表（改造）

新增字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| tenant_id | BIGINT NOT NULL | 归属租户，加 INDEX |

其余字段不变。

### 2.6 login_logs 表（改造）

新增字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| tenant_id | BIGINT NOT NULL | 归属租户，加 INDEX |

其余字段不变。

### 2.7 sys_configs 表

不变，作为平台级默认配置。新建租户时用这些默认值初始化租户的 key_prefix / key_length / key_suffix_length。

## 3. 认证与权限

### 3.1 JWT Claims

```json
{
  "user_id": 1,
  "username": "admin",
  "role": "super_admin",
  "tenant_id": null,
  "exp": 1234567890,
  "iat": 1234567890
}
```

`role` 取值：`super_admin` / `tenant_admin`。`tenant_id` 对 super_admin 为 null。

### 3.2 中间件链

**AuthMiddleware**（JWT 验证）
- 解析 JWT，提取 user_id、role、tenant_id
- 存入 Gin context
- 失败返回 HTTP 401

**RequireSuperAdmin**
- 检查 `role == "super_admin"`
- 不满足返回 HTTP 403

**RequireTenantAdmin**
- 检查 `role == "tenant_admin"` 且 `tenant_id` 有效
- 检查租户是否为 `disabled` 状态 → 403 "账号已被禁用"
- expired 状态仍放行（后续由 TenantBusinessGuard 控制业务能力）

**TenantBusinessGuard**（业务操作守卫）
- 检查租户 `status == "active"`
- `expired` → 403 "租户已到期，仅可查看统计数据"
- `disabled` → 403 "租户已被禁用"
- 用于：Key CRUD、Key 验证/扣减、服务账号 CRUD、服务账号 API

**ServiceAuthMiddleware**（改造）
- 验证 X-Service-Key 后，查找关联的租户
- 租户 expired/disabled 时拒绝请求

### 3.3 租户状态与能力矩阵

| 状态 | 登录 | 业务操作 | 统计 & 导出 |
|------|------|----------|-------------|
| active | ✅ | ✅ | ✅ |
| expired | ✅ | 🚫 | ✅ |
| disabled | 🚫 | 🚫 | 🚫 |

业务操作：Key CRUD、Key 验证/扣减、服务账号 CRUD、服务账号 API。
统计 & 导出：仪表盘、趋势、Top Key/IP、使用日志查看与导出。

### 3.4 登录流程

```
POST /api/auth/login { username, password }
  → 验证用户名密码
  → 检查 is_active（disabled 租户的管理员直接拒绝）
  → 检查 TOTP 是否已设置
    → 未设置: 返回 { need_setup: true, user_id, temp_token }
    → 已设置: 返回 { require_totp: true, user_id, temp_token }

POST /api/auth/verify-2fa { user_id, code, temp_token }
  → 验证 TOTP code
  → 返回 { token: "JWT", role, tenant_id, username }
  → 前端根据 role 加载对应页面

POST /api/auth/totp/setup-init { user_id, temp_token }
  → 首次 TOTP 设置，返回二维码

POST /api/auth/totp/confirm-init { user_id, code, temp_token }
  → 确认 TOTP → 返回 JWT
```

temp_token 有效期 5 分钟，仅用于首次登录到完成 2FA 之间的临时凭证。

## 4. API 路由

### 4.1 认证（公开）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/auth/login | 统一登录 |
| POST | /api/auth/verify-2fa | TOTP 验证 |
| POST | /api/auth/totp/setup-init | 首次 TOTP 设置 |
| POST | /api/auth/totp/confirm-init | 确认 TOTP |

### 4.2 系统管理员（JWT + RequireSuperAdmin）

**租户管理：**

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/super/tenants | 租户列表（含聚合统计） |
| POST | /api/super/tenants | 创建租户（自动生成管理员账号+密码） |
| GET | /api/super/tenants/:id | 租户详情 + 统计 |
| PATCH | /api/super/tenants/:id | 编辑（名称、状态、到期时间、Key 配置） |
| PATCH | /api/super/tenants/:id/reset-password | 重置租户管理员密码 |

**平台配置：**

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/super/configs | 平台配置列表 |
| PUT | /api/super/configs | 更新平台配置 |

**个人设置：**

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/super/profile | 个人信息 |
| PUT | /api/super/password | 修改密码 |
| POST | /api/super/totp/setup | TOTP 设置 |
| POST | /api/super/totp/confirm | TOTP 确认 |
| GET | /api/super/login-logs | 登录日志 |

系统管理员**没有** Key 管理、服务账号管理接口。

### 4.3 租户管理员（JWT + RequireTenantAdmin）

**Key 管理（业务操作需 TenantBusinessGuard，统计/导出不需）：**

| 方法 | 路径 | BusinessGuard | 说明 |
|------|------|---------------|------|
| POST | /api/tenant/keys | ✅ | 创建 Key |
| GET | /api/tenant/keys | - | 列表（分页、筛选） |
| GET | /api/tenant/keys/:id | - | 详情 |
| PATCH | /api/tenant/keys/:id | ✅ | 编辑 |
| PATCH | /api/tenant/keys/:id/disable | ✅ | 禁用 |
| PATCH | /api/tenant/keys/:id/enable | ✅ | 启用 |
| DELETE | /api/tenant/keys/:id | ✅ | 删除 |
| GET | /api/tenant/keys/export | - | 导出 |

**服务账号管理（需 TenantBusinessGuard）：**

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/tenant/service-accounts | 列表 |
| POST | /api/tenant/service-accounts | 创建 |
| PATCH | /api/tenant/service-accounts/:id/toggle | 启用/禁用 |
| DELETE | /api/tenant/service-accounts/:id | 删除 |

**统计 & 导出（不需要 TenantBusinessGuard）：**

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/tenant/stats/dashboard | 仪表盘 |
| GET | /api/tenant/stats/overview | 概览 |
| GET | /api/tenant/stats/trends | 趋势 |
| GET | /api/tenant/stats/top-keys | Top Key |
| GET | /api/tenant/stats/top-ips | Top IP |
| GET | /api/tenant/usage-logs | 使用日志 |
| GET | /api/tenant/usage-logs/export | 导出使用日志 |

**个人设置：**

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/tenant/profile | 个人信息 |
| PUT | /api/tenant/password | 修改密码 |
| POST | /api/tenant/totp/setup | TOTP 设置 |
| POST | /api/tenant/totp/confirm | TOTP 确认 |
| GET | /api/tenant/login-logs | 登录日志 |

### 4.4 服务账号 API（X-Service-Key）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/service/keys | 创建 Key |
| GET | /api/service/keys | 查看自己的 Key |

受租户状态限制：expired/disabled 时拒绝。

## 5. 代码结构

### 5.1 目录变更

```
internal/
  model/
    tenant.go                 -- 新增
    user.go                   -- 重命名自 admin.go（User 模型替代 Admin）
    key.go                    -- 改动：加 tenant_id
    service_account.go        -- 改动：加 tenant_id
    usage_log.go              -- 改动：加 tenant_id
    login_log.go              -- 改动：加 tenant_id
    config.go                 -- 不变
    migrate.go                -- 更新：新增 tenants、users 替代 admins

  middleware/
    auth.go                   -- 改造：JWT 解析 role + tenant_id
    super_admin.go            -- 新增：RequireSuperAdmin
    tenant_admin.go           -- 新增：RequireTenantAdmin
    tenant_business.go        -- 新增：TenantBusinessGuard
    service_auth.go           -- 改造：验证后检查租户状态
    cors.go                   -- 不变

  handler/
    auth_handler.go           -- 新增：统一登录/2FA
    super_handler.go          -- 新增：系统管理员接口
    tenant_handler.go         -- 新增：租户管理员接口
    service_handler.go        -- 改造：增加租户状态检查
    response.go               -- 不变

  service/
    auth_service.go           -- 新增：登录、2FA、JWT、temp_token
    tenant_service.go         -- 新增：租户 CRUD + 统计聚合
    user_service.go           -- 重命名自 admin_service.go
    key_service.go            -- 改造：所有查询加 tenant_id
    service_account_service.go -- 改造：加 tenant_id
    stats_service.go          -- 改造：支持租户级和平台级统计
    usage_log_service.go      -- 改造：加 tenant_id
    login_log_service.go      -- 改造：加 tenant_id
    config_service.go         -- 不变
```

### 5.2 设计原则

1. **TenantScope 注入**：middleware 将 tenant_id 注入 Gin context，service 层通过 `c.GetInt64("tenant_id")` 获取并追加到所有查询，不在 service 层自行判断权限。
2. **Handler 层按角色分文件**：super_handler.go 和 tenant_handler.go 职责清晰不混杂。
3. **服务账号 API 路径不变**（`/api/service/keys`），仅内部逻辑增加租户隔离。

## 6. 部署与配置

### 6.1 config.yaml 变化

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  # 不变

auth:
  secret: "jwt-signing-key"
  expiration: 24
  super_admin_username: "admin"    # 原 admin_username
  super_admin_password: "xxx"      # 原 admin_password

app:
  debug: false
```

环境变量：`SUPER_ADMIN_USERNAME` / `SUPER_ADMIN_PASSWORD`（优先级高于配置文件）。

### 6.2 初始化流程

1. 读取配置 → 连接数据库 → AutoMigrate
2. 检查是否存在 `role = super_admin` 的用户，不存在则用 config 凭据自动创建
3. 启动 Gin 服务

### 6.3 静态文件服务

```go
router.StaticFile("/", "web/index.html")
router.NoRoute(func(c *gin.Context) {
    if !strings.HasPrefix(c.Request.URL.Path, "/api") {
        c.File("web/index.html")
    }
})
```

当前 admin.html 可先保留，后续 React 构建产物放到 `web/` 目录替换。

## 7. 错误码

在 `errcode/errcode.go` 中新增：

| Code | 说明 |
|------|------|
| 4001 | 租户已过期 |
| 4002 | 租户已被禁用 |
| 4003 | 租户不存在 |
| 5001 | 系统管理员权限不足 |
| 5002 | 租户管理员权限不足 |

统一响应格式不变：
```json
{ "code": 0, "message": "success", "data": {...} }
{ "code": 4001, "message": "租户已过期", "data": null }
```

## 8. 租户生命周期

### 创建租户
1. 系统管理员填写租户名称
2. 系统自动生成管理员账号（格式：`{tenant_name}_xxxx`，如 `myapp_asdasd`）和，账号字符随机6位。随机初始密码（16 位，含大小写字母、数字、特殊字符）
3. 用平台默认配置初始化租户的 key_prefix / key_length / key_suffix_length
4. 返回账号密码给系统管理员，由其线下分发

### 租户到期
1. 系统定期检查或在请求时检查 expire_at
2. 到期后 status 自动标记为 expired
3. 租户管理员仍可登录，查看/导出统计，但无法执行业务操作
4. 服务账号 API 请求被拒绝
5. 系统管理员可续期（修改 expire_at），恢复为 active

### 禁用租户
1. 系统管理员手动将租户 status 设为 disabled
2. 租户管理员无法登录
3. 所有请求被拒绝

## 9. 不在范围内

- 租户间数据共享或跨租户 Key 分发
- 租户自助注册（仅系统管理员创建）
- 计费系统 / 套餐管理
- 租户级 API 限流
- 前端 React 改造（本次仅设计后端架构，前端后续独立推进）
