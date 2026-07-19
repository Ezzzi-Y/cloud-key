# CloudKey React SPA 前端设计文档

> 日期：2026-07-19
> 状态：设计定稿
> 范围：为 SaaS 多租户后端构建 React SPA 前端，覆盖全部后端功能

## 1. 背景与目标

CloudKey 后端已完成 SaaS 多租户架构改造，拥有 40+ API 接口，覆盖认证、租户管理、Key 管理、服务账号、统计、使用日志等功能。当前前端是单文件 `web/admin.html`（538 行原生 JS），仅覆盖约 40% 的接口。

**目标：**
- 构建 React SPA，完整覆盖后端全部 API
- Go embed 嵌入构建产物，保持单一二进制部署
- 统一登录入口，前端根据角色（super_admin / tenant_admin）加载不同页面
- 全中文 UI

## 2. 技术栈

| 层 | 选型 | 版本 |
|---|---|---|
| 框架 | React + TypeScript | 18.x |
| 构建 | Vite | 5.x |
| UI 组件 | shadcn/ui + Tailwind CSS | v4 |
| 路由 | react-router-dom | v6 |
| 服务端状态 | TanStack Query (React Query) | v5 |
| 客户端状态 | React Context | - |
| 图表 | Recharts | 2.x |
| HTTP | axios | 1.x |
| 部署 | Go `//go:embed` | - |

## 3. 项目结构

```
CloudKey/
├── main.go
├── internal/
│   └── web/
│       └── embed.go              # //go:embed dist/*
├── web/                          # React 项目
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   ├── components.json           # shadcn/ui 配置
│   ├── index.html
│   ├── public/
│   └── src/
│       ├── main.tsx              # 入口：QueryClientProvider + AuthProvider + RouterProvider
│       ├── App.tsx               # 路由定义
│       ├── lib/
│       │   └── utils.ts          # cn() 等工具函数
│       ├── api/
│       │   ├── client.ts         # axios 实例 + JWT 拦截器 + 401 处理
│       │   ├── auth.ts           # POST /api/auth/* 接口
│       │   ├── keys.ts           # /api/tenant/keys/* 接口
│       │   ├── tenants.ts        # /api/super/tenants/* 接口
│       │   ├── stats.ts          # /api/tenant/stats/* 接口
│       │   ├── logs.ts           # /api/tenant/usage-logs/* 接口
│       │   ├── service-accounts.ts
│       │   └── config.ts         # /api/super/configs 接口
│       ├── hooks/
│       │   ├── useAuth.tsx        # AuthContext + AuthProvider + useAuth()
│       │   └── useApi.ts          # 各域的 React Query hooks
│       ├── components/
│       │   └── ui/               # shadcn/ui 生成的组件
│       │       ├── button.tsx
│       │       ├── dialog.tsx
│       │       ├── table.tsx
│       │       ├── input.tsx
│       │       ├── select.tsx
│       │       ├── toast.tsx
│       │       ├── tabs.tsx
│       │       ├── card.tsx
│       │       ├── badge.tsx
│       │       ├── dropdown-menu.tsx
│       │       ├── sheet.tsx
│       │       └── ...
│       ├── layouts/
│       │   ├── SuperAdminLayout.tsx    # 侧边栏：租户管理、平台配置、个人设置
│       │   └── TenantAdminLayout.tsx   # 侧边栏：仪表盘、Key管理、校验扣减、使用记录、服务账号、个人设置
│       └── pages/
│           ├── Login.tsx               # 统一登录 + TOTP 流程（三步状态机）
│           ├── super/
│           │   ├── Dashboard.tsx       # 租户概览统计
│           │   ├── Tenants.tsx         # 租户列表
│           │   ├── TenantDetail.tsx    # 租户详情/编辑
│           │   ├── PlatformConfig.tsx  # 平台配置管理
│           │   └── Profile.tsx         # 个人设置（Tabs：资料+密码 | TOTP | 登录日志）
│           └── tenant/
│               ├── Dashboard.tsx       # 统计仪表盘（卡片 + 趋势图 + Top10 + 最近记录）
│               ├── KeyManagement.tsx   # Key 列表/CRUD
│               ├── KeyVerify.tsx       # Key 校验与扣减
│               ├── UsageLogs.tsx       # 使用记录
│               ├── ServiceAccounts.tsx # 服务账号管理
│               └── Profile.tsx         # 个人设置
```

## 4. 路由结构

```
/login                          → Login.tsx（未认证可访问）

/super/                         → SuperAdminLayout（需 super_admin）
  /super/                       → Dashboard.tsx
  /super/tenants                → Tenants.tsx
  /super/tenants/:id            → TenantDetail.tsx
  /super/config                 → PlatformConfig.tsx
  /super/profile                → Profile.tsx

/tenant/                        → TenantAdminLayout（需 tenant_admin）
  /tenant/                      → Dashboard.tsx
  /tenant/keys                  → KeyManagement.tsx
  /tenant/keys/verify           → KeyVerify.tsx
  /tenant/logs                  → UsageLogs.tsx
  /tenant/service-accounts      → ServiceAccounts.tsx
  /tenant/profile               → Profile.tsx

*                               → SPA fallback → index.html
```

**路由守卫：** `<RequireAuth role="super_admin">` / `<RequireAuth role="tenant_admin">` 组件包裹 Layout。未登录 → redirect `/login`；角色不匹配 → redirect 到正确角色首页。

## 5. 认证流程

### 5.1 登录状态机

```
[输入账号密码]
    ↓ POST /api/auth/login
[判断返回]
    ├── require_totp=true → [输入 TOTP 验证码]
    │                           ↓ POST /api/auth/verify-2fa
    │                           [返回 JWT] → 根据 role 跳转
    └── need_setup=true → [TOTP 设置向导]
                              ↓ POST /api/auth/totp/setup-init
                              [显示 QR 码]
                              ↓ POST /api/auth/totp/confirm-init
                              [返回 JWT] → 根据 role 跳转
```

### 5.2 Auth 状态管理

```typescript
interface AuthState {
  token: string | null;
  role: 'super_admin' | 'tenant_admin' | null;
  tenantId: number | null;
  username: string | null;
  isAuthenticated: boolean;
  login: (token: string, role: string, tenantId: number | null, username: string) => void;
  logout: () => void;
}
```

- JWT 存 `localStorage`（键名 `ck_token`）
- 页面刷新时从 localStorage 恢复，通过调用 profile 接口验证有效性
- 登录成功后根据 role 跳转：`super_admin → /super/`，`tenant_admin → /tenant/`

### 5.3 axios 拦截器

```typescript
// 请求拦截：自动带 Bearer token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('ck_token');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

// 响应拦截：统一处理 401
api.interceptors.response.use(
  (res) => res.data,  // 直接返回 { code, message, data }
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('ck_token');
      window.location.href = '/login';
    }
    return Promise.reject(err.response?.data || err);
  }
);
```

## 6. 页面与功能映射

### 6.1 登录页 `/login`

| 步骤 | 状态 | API |
|------|------|-----|
| 输入账号密码 | form | `POST /api/auth/login` |
| TOTP 验证 | totp | `POST /api/auth/verify-2fa` |
| TOTP 首次设置 | setup | `POST /api/auth/totp/setup-init` + `POST /api/auth/totp/confirm-init` |

同一页面内通过状态切换显示不同表单。登录成功后根据返回的 role 跳转。

### 6.2 超级管理员 — Dashboard `/super/`

| 功能 | API |
|------|-----|
| 租户总数 + 各状态统计 | `GET /api/super/tenants`（前端聚合） |
| 最近创建的租户列表 | `GET /api/super/tenants` |

统计卡片：总租户数、活跃/过期/禁用数。

### 6.3 超级管理员 — 租户管理 `/super/tenants`

| 功能 | API |
|------|-----|
| 租户列表（表格） | `GET /api/super/tenants` |
| 创建租户 | `POST /api/super/tenants` |
| 点击进入详情 | 跳转 `/super/tenants/:id` |

表格列：名称、状态（Badge）、到期时间、Key 数量、用户数、创建时间。
创建后弹窗显示生成的管理员账号和密码（仅显示一次）。

### 6.4 超级管理员 — 租户详情 `/super/tenants/:id`

| 功能 | API |
|------|-----|
| 租户详情 | `GET /api/super/tenants/:id` |
| 编辑名称 | `PATCH /api/super/tenants/:id` |
| 切换状态 | `PATCH /api/super/tenants/:id` |
| 设置到期时间 | `PATCH /api/super/tenants/:id` |
| Key 配置（prefix/length/suffix） | `PATCH /api/super/tenants/:id` |
| 重置管理员密码 | `PATCH /api/super/tenants/:id/reset-password` |

状态切换注意事项：disabled 状态租户管理员无法登录。expired 状态仍可登录但无法执行业务操作。

### 6.5 超级管理员 — 平台配置 `/super/config`

| 功能 | API |
|------|-----|
| 配置列表 | `GET /api/super/configs` |
| 批量更新 | `PUT /api/super/configs` |

表格展示所有配置项（key、value、description），支持行内编辑 + 批量保存。

### 6.6 超级管理员 — 个人设置 `/super/profile`

三个 Tab：

| Tab | 功能 | API |
|-----|------|-----|
| 资料 & 密码 | 查看信息 + 修改密码 | `GET /api/super/profile` + `PUT /api/super/password` |
| TOTP 管理 | 重新绑定 TOTP | `POST /api/super/totp/setup` + `POST /api/super/totp/confirm` |
| 登录日志 | 分页展示登录历史 | `GET /api/super/login-logs` |

### 6.7 租户管理员 — Dashboard `/tenant/`

| 功能 | API |
|------|-----|
| Key 总数、各状态数 | `GET /api/tenant/stats/dashboard` |
| 今日/本周/本月调用量 | `GET /api/tenant/stats/dashboard` |
| 调用趋势图（today/week/month） | `GET /api/tenant/stats/trends` |
| Top 10 卡密 | `GET /api/tenant/stats/top-keys` |
| Top 10 IP | `GET /api/tenant/stats/top-ips` |
| 最近 20 条使用记录 | dashboard 接口内含 |

趋势图使用 Recharts `<AreaChart>` 或 `<BarChart>`。

### 6.8 租户管理员 — Key 管理 `/tenant/keys`

| 功能 | API |
|------|-----|
| Key 列表（分页、搜索、状态筛选） | `GET /api/tenant/keys` |
| 创建 Key | `POST /api/tenant/keys` |
| 查看详情 | `GET /api/tenant/keys/:id` |
| 编辑（别名、剩余额度） | `PATCH /api/tenant/keys/:id` |
| 禁用 / 启用 | `PATCH /api/tenant/keys/:id/disable` / `enable` |
| 删除（确认弹窗） | `DELETE /api/tenant/keys/:id` |
| 导出 CSV | `GET /api/tenant/keys/export` |
| 导出 JSON | `GET /api/tenant/keys/export/json` |

表格列：别名、前缀+后缀（`sk-****abcd`）、计费模式、初始额度、剩余额度、状态（Badge）、创建时间、最后使用时间。

创建弹窗字段：别名、计费模式（count/credit 单选）、初始额度、过期时间（可选）、最大使用次数（可选）。

### 6.9 租户管理员 — Key 校验与扣减 `/tenant/keys/verify`

| 功能 | API |
|------|-----|
| 校验 Key 状态 | `GET /api/key/status?sk=<raw_key>` |
| 扣减 Key | `POST /api/key/consume` |

**页面布局：**
1. 顶部：卡密输入框 +「校验」按钮
2. 中部：校验结果卡片（别名、计费模式、剩余额度、状态）
3. 底部：扣减数量输入 +「扣减」按钮 + 扣减结果实时反馈

**交互流程：**
1. 粘贴用户提供的 Key 明文
2. 点击「校验」→ 调用 status 接口 → 展示 Key 信息
3. 输入扣减数量 → 点击「扣减」→ 调用 consume 接口 → 显示扣减后余额
4. 错误（Key 不存在、已禁用、余额不足）有明确中文提示

### 6.10 租户管理员 — 使用记录 `/tenant/logs`

| 功能 | API |
|------|-----|
| 列表（分页） | `GET /api/tenant/usage-logs` |
| 时间范围筛选 | `start_time` / `end_time` 查询参数 |
| 按别名/IP 筛选 | `keyword` 查询参数 |
| 导出 | `GET /api/tenant/usage-logs/export` |

表格列：Key 别名、扣减数量、IP、User-Agent、请求路径、时间。

### 6.11 租户管理员 — 服务账号 `/tenant/service-accounts`

| 功能 | API |
|------|-----|
| 列表 | `GET /api/tenant/service-accounts` |
| 创建 | `POST /api/tenant/service-accounts` |
| 启用/禁用 | `PATCH /api/tenant/service-accounts/:id/toggle` |
| 删除（确认弹窗） | `DELETE /api/tenant/service-accounts/:id` |

创建后弹窗显示密钥明文（仅一次）。表格列：名称、状态（Badge）、创建时间。

### 6.12 租户管理员 — 个人设置 `/tenant/profile`

同超级管理员 Profile 结构（三个 Tab），API 路径不同（`/api/tenant/profile` 等）。

## 7. 错误处理

### 7.1 统一响应格式

后端返回 `{ code, message, data }`：
- `code === 0` → 成功，使用 `data`
- `code !== 0` → 业务错误，toast 显示 `message`

### 7.2 HTTP 错误

- 401 → 清除 token → 跳转 `/login`
- 403 → 页面级提示（权限不足 / 租户过期 / 租户禁用）
- 500 → toast 提示「服务器错误，请稍后重试」

### 7.3 业务错误码

| Code | 含义 | 前端处理 |
|------|------|----------|
| 4001 | 租户已过期 | 顶部 Banner 提示，禁用业务操作按钮 |
| 4002 | 租户已被禁用 | 全屏提示，自动跳转登录页 |
| 4003 | 租户不存在 | 404 页面 |
| 5001 | 需要系统管理员权限 | 跳转正确角色首页 |
| 5002 | 需要租户管理员权限 | 跳转正确角色首页 |

## 8. Go 端改造

### 8.1 embed.go

```go
package web

import "embed"

//go:embed dist/*
var FS embed.FS
```

创建 `internal/web/embed.go`，嵌入 `web/dist/` 目录。

### 8.2 router.go 改造

替换现有的静态文件服务：
- 移除 `r.StaticFile("/", "web/admin.html")`
- 使用 `http.FS(web.FS)` 提供 `web/dist/` 静态文件
- 非 `/api/*` 请求：先尝试提供静态文件，不存在则 fallback 到 `index.html`（SPA 路由支持）

### 8.3 开发模式

`config.yaml` 中 `app.debug = true` 时：
- 跳过 embed FS
- 前端独立运行 `cd web && npm run dev`（Vite dev server 端口 5173）
- Vite 配置代理 `/api` 到 `http://localhost:8080`

```typescript
// vite.config.ts
export default defineConfig({
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
});
```

### 8.4 Dockerfile

```dockerfile
# Stage 1: 前端构建
FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Go 构建
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist internal/web/dist
RUN CGO_ENABLED=0 go build -o cloudkey .

# Stage 3: 运行
FROM alpine:3.20
COPY --from=builder /app/cloudkey /usr/local/bin/
CMD ["cloudkey"]
```

## 9. 迁移策略

- 新 React 前端构建到 `web/dist/`
- 旧 `web/admin.html` 保留不删除，但 router.go 不再引用它
- 后续可选择性删除 `admin.html`

## 10. 不在范围内

- 批量创建 Key（后端暂无此接口）
- 多管理员支持（后端暂无此功能）
- 国际化（i18n）
- 深色模式
- 移动端适配（优先桌面端）
- 服务账号 API 的前端界面（服务账号通过 X-Service-Key 调用，无 Web 界面）
