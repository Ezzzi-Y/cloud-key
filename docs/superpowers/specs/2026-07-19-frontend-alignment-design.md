# CloudKey 前端对齐设计文档：React SPA 全功能迁移

## 1. 项目概述

### 1.1 背景

当前管理后台是单文件 `web/admin.html`（原生 JS，约 500 行），仅覆盖了后端约 40% 的接口。本项目目标是将前端迁移为 React SPA，**完整对齐后端所有 API**，为后续功能扩展奠定基础。

### 1.2 技术栈

| 层 | 选型 |
|---|---|
| 框架 | React 18 + TypeScript |
| 构建 | Vite 5 |
| UI 组件库 | MUI (Material UI) v5 |
| 路由 | react-router-dom v6 |
| 状态管理 | React Context + useAuth hook（无需 Redux） |
| 部署 | Go `embed` 嵌入单一二进制 |

### 1.3 核心目标

- 完整覆盖后端 **全部 25+ API 接口**
- 管理员可执行**卡密校验与扣减**业务操作
- 保持单一二进制部署的简洁性
- 为后续复杂功能提供可扩展的前端架构

---

## 2. 项目结构

```
CloudKey/
├── main.go
├── internal/
│   ├── web/
│   │   └── embed.go                 # //go:embed web/dist/*
│   └── router/router.go            # 静态文件服务 + SPA fallback
├── web/                             # React 项目
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   ├── index.html
│   ├── public/
│   └── src/
│       ├── main.tsx
│       ├── App.tsx
│       ├── api/
│       │   ├── client.ts            # fetch 封装 + JWT 拦截
│       │   ├── keys.ts
│       │   ├── logs.ts
│       │   ├── stats.ts
│       │   ├── admin.ts
│       │   ├── config.ts
│       │   └── accounts.ts
│       ├── pages/
│       │   ├── Login.tsx
│       │   ├── Dashboard.tsx
│       │   ├── KeyManagement.tsx
│       │   ├── KeyVerify.tsx        # 卡密校验与扣减
│       │   ├── UsageLogs.tsx
│       │   ├── ServiceAccounts.tsx
│       │   ├── SystemConfig.tsx
│       │   └── AdminProfile.tsx
│       ├── components/
│       │   ├── Layout.tsx
│       │   ├── StatsChart.tsx
│       │   └── ConfirmDialog.tsx
│       ├── hooks/
│       │   └── useAuth.ts
│       └── types/
│           └── index.ts
├── Dockerfile
└── docker-compose.yml
```

---

## 3. 页面与功能映射

### 3.1 登录页 `/login`

| 功能 | API |
|---|---|
| 用户名密码登录 | `POST /api/admin/login` |
| TOTP 验证 | `POST /api/admin/login/verify-2fa` |
| 首次 TOTP 绑定 | `POST /api/admin/totp/setup-init` + `POST /api/admin/totp/confirm-init` |

### 3.2 仪表盘 `/`

| 功能 | API |
|---|---|
| 统计卡片（总数、各状态、调用量） | `GET /api/admin/stats/dashboard` |
| 使用趋势图（today/week/month） | `GET /api/admin/stats/trends` |
| Top 10 卡密 | `GET /api/admin/stats/top-keys` |
| Top 10 IP | `GET /api/admin/stats/top-ips` |
| 最近 20 条使用记录 | dashboard 接口内含 |

趋势图使用纯 CSS 条形图实现，零外部依赖。

### 3.3 卡密管理 `/keys`

| 功能 | API |
|---|---|
| 列表（分页、搜索、状态筛选） | `GET /api/admin/keys` |
| 创建卡密 | `POST /api/admin/keys` |
| 查看详情 | `GET /api/admin/keys/:id` |
| 编辑（别名、剩余额度） | `PATCH /api/admin/keys/:id` |
| 禁用 / 启用 | `PATCH /api/admin/keys/:id/disable` / `enable` |
| 删除 | `DELETE /api/admin/keys/:id` |
| 导出 | `GET /api/admin/keys/export` |

**新增功能**：详情弹窗、编辑弹窗、导出按钮、已用尽状态也可启用。

### 3.4 卡密校验与扣减 `/verify`（核心新增）

管理员帮助用户查询卡密状态并代为扣减的业务操作页面。

| 功能 | API |
|---|---|
| 输入卡密查询状态 | `GET /api/key/status?sk=<raw_key>` |
| 扣减指定数量 | `POST /api/key/consume` |

**页面布局**：
- 顶部：卡密输入框 + 「校验」按钮
- 中部：校验结果卡片（别名、计费模式、剩余额度、状态）
- 底部：扣减数量输入 + 「扣减」按钮 + 扣减结果反馈

**交互流程**：
1. 管理员粘贴用户提供的卡密明文
2. 点击「校验」→ 调用 status 接口 → 展示卡密信息
3. 输入扣减数量 → 点击「扣减」→ 调用 consume 接口 → 实时显示扣减后余额
4. 错误情况（卡密不存在、已禁用、余额不足）有明确提示

### 3.5 使用记录 `/logs`

| 功能 | API |
|---|---|
| 列表（分页、按别名/IP 筛选） | `GET /api/admin/usage-logs` |
| 时间范围筛选 | `start_time` / `end_time` 查询参数 |
| 导出 | `GET /api/admin/usage-logs/export` |

**新增功能**：时间范围 datepicker、导出按钮。

### 3.6 服务账号 `/accounts`

| 功能 | API |
|---|---|
| 列表 | `GET /api/admin/service-accounts` |
| 创建 | `POST /api/admin/service-accounts` |
| 启用/禁用 | `PATCH /api/admin/service-accounts/:id/toggle` |
| 删除 | `DELETE /api/admin/service-accounts/:id` |

保持现有功能，增加分页。

### 3.7 系统配置 `/settings`（全新）

| 功能 | API |
|---|---|
| 查看所有配置 | `GET /api/admin/configs` |
| 修改配置 | `PUT /api/admin/configs` |

**页面布局**：表格展示所有配置项（key、value、description），支持行内编辑 + 批量保存。

### 3.8 管理员中心 `/profile`（全新）

三个 Tab：

| Tab | 功能 | API |
|---|---|---|
| 个人资料 | 查看信息 + 修改密码 | `GET /api/admin/profile` + `PUT /api/admin/password` |
| TOTP 管理 | 重新绑定 TOTP | `POST /api/admin/totp/setup` + `POST /api/admin/totp/confirm` |
| 登录日志 | 分页展示登录历史 | `GET /api/admin/login-logs` |

---

## 4. Go 端改造

### 4.1 embed.go

```go
package web

import "embed"

//go:embed dist/*
var FS embed.FS
```

### 4.2 router.go 改造

- 用 `http.FS(web.FS)` 提供 `web/dist/` 静态文件
- 非 `/api/*` 请求 fallback 到 `index.html`（SPA 路由支持）
- `app.debug=true` 时跳过 embed，由 Vite dev server 独立运行

### 4.3 Dockerfile 改造

```dockerfile
# Stage 1: Node 构建前端
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
COPY --from=frontend /app/web/dist web/dist
RUN CGO_ENABLED=0 go build -o cloudkey .

# Stage 3: 运行时
FROM alpine:3.20
COPY --from=builder /app/cloudkey /usr/local/bin/
CMD ["cloudkey"]
```

---

## 5. 开发工作流

### 5.1 开发模式

```bash
# 终端 1: Go 后端
go run . 

# 终端 2: Vite 前端（代理 /api 到 localhost:8080）
cd web && npm run dev
```

Vite 配置：
```ts
export default defineConfig({
  server: {
    proxy: {
      '/api': 'http://localhost:8080'
    }
  }
})
```

### 5.2 生产构建

```bash
cd web && npm run build   # 生成 web/dist/
go build -o cloudkey .    # embed 嵌入
```

---

## 6. 认证与路由

- JWT 存 `localStorage`（键名 `ck_token`，与现有方案兼容）
- `api/client.ts` 封装 fetch，自动带 `Authorization: Bearer` 头
- 401 / code 2003 → 自动跳转 `/login`
- 未登录访问受保护路由 → redirect 到 `/login`
- 登录成功后 redirect 到 `/`（仪表盘）

---

## 7. 迁移策略

- 新 React 前端构建到 `web/dist/`，Go embed 嵌入
- 旧 `web/admin.html` **保留不删除**，但 router.go 不再引用它（`StaticFile("/", ...)` 替换为 embed FS 服务）
- 后续可选择性删除 `admin.html`

---

## 8. 不在本次范围内

- 批量创建卡密（后端暂无此接口）
- 多管理员支持（后端暂无此功能）
- 国际化（i18n）
- 深色模式
- 移动端适配（优先桌面端）
