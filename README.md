# CloudKey

> 基于 Go 语言编写的卡密发放、扣减、验证平台，适用于 API 计费和虚拟产品兑换码。

## 📖 About

CloudKey 是一个多租户 SaaS 平台，专注于 API Key / 卡密的全生命周期管理。管理员可以创建、分发、监控卡密的使用情况；终端用户或自动化系统通过公开 API 进行卡密验证和扣减，实现对外服务的按量计费与虚拟产品兑换。

## ✨ 核心特性

- **多租户 SaaS 架构** — 每个租户独立管理卡密、用户、服务账号，支持自定义卡密前缀和长度
- **两种计费模式** — 按次计费（count）和额度计费（credit）
- **卡密安全管理** — 卡密仅存储哈希值，明文仅在创建时返回一次
- **卡密验证与扣减** — 提供状态查询和扣减消费的公开接口，基于数据库事务 + 乐观锁实现原子扣减
- **服务账号（Service Account）** — 机器对机器的 API 访问，支持自动化卡密管理
- **双因素认证（2FA）** — 基于 TOTP 的两步验证
- **JWT 认证** — 可配置过期时间，支持 2FA 流程中的预认证 Token
- **Redis 限流** — 认证端点接入 Redis 速率限制
- **租户生命周期管理** — 支持活跃 / 过期 / 禁用状态，过期租户仅可读
- **数据统计仪表盘** — 卡密数量分布、调用趋势（日/周/月）、热门卡密、最近日志
- **使用日志与导出** — 完整的卡密消费审计日志，支持筛选和导出
- **前端嵌入式部署** — React SPA 编译后嵌入 Go 二进制文件，单文件部署
- **自动生成 SDK** — 通过 OpenAPI Generator 自动生成 Java 客户端 SDK

## 🛠 技术栈

### 后端

| 组件 | 技术 |
|---|---|
| 语言 | Go 1.25 |
| HTTP 框架 | Gin |
| ORM | GORM (MySQL) |
| 缓存 / 限流 | Redis (go-redis v9) |
| 认证 | JWT (golang-jwt v5) + TOTP |
| 日志 | Zap + Lumberjack |
| 配置 | Viper (YAML) |
| API 文档 | Swaggo (OpenAPI) |

### 前端

| 组件 | 技术 |
|---|---|
| 框架 | React 18 + TypeScript |
| 构建工具 | Vite 5 |
| 样式 | Tailwind CSS 3 |
| 组件库 | Radix UI + shadcn/ui |
| 数据请求 | TanStack React Query + Axios |
| 图表 | Recharts |
| 路由 | React Router v6 |

### 基础设施

| 组件 | 技术 |
|---|---|
| 数据库 | MySQL 8.0+ |
| 容器化 | Docker + Docker Compose |
| CI/CD | GitHub Actions |

## 📁 项目结构

```
CloudKey/
├── main.go                  # 应用入口
├── config.yaml.example      # 配置示例
├── Dockerfile               # Docker 构建文件
├── docker-compose.yml       # Docker Compose 编排
├── internal/                # 后端 Go 源码
│   ├── config/              # 配置加载
│   ├── database/            # MySQL & Redis 连接
│   ├── model/               # 数据模型 (GORM)
│   ├── service/             # 业务逻辑层
│   ├── handler/             # HTTP 处理器
│   ├── router/              # 路由定义
│   ├── middleware/           # 中间件 (JWT, CORS, 限流等)
│   ├── errcode/             # 统一错误码
│   ├── log/                 # 日志初始化
│   └── web/                 # 前端嵌入
├── web/                     # React SPA 前端
├── sdk/                     # 客户端 SDK (Java)
└── docs/                    # Swagger 文档 & 设计文档
```

## 🚀 快速开始

### 环境要求

- Go 1.25+
- MySQL 8.0+
- Redis 6.0+
- Node.js 18+（构建前端）

### Docker Compose 部署（推荐）

```bash
# 克隆项目
git clone https://github.com/Ezzzi-Y/cloud-key.git
cd CloudKey

# 复制并修改配置
cp config.yaml.example config.yaml

# 启动服务
docker-compose up -d
```

### 手动部署

```bash
# 1. 构建前端并嵌入 Go 项目
cd web
npm install
npm run build
cp -r dist ../internal/web/dist
cd ..

# 2. 复制并修改配置
cp config.yaml.example config.yaml

# 3. 编译运行
go build -o cloudkey .
./cloudkey
```

> **说明：** `//go:embed dist/*` 在 Go 编译时将前端产物嵌入二进制文件，因此必须先构建前端到 `internal/web/dist/`，再执行 `go build`。开发时使用 `npm run dev` 启动前端，通过 Vite 代理 `/api` 请求到后端 `localhost:8080`。

服务默认监听 `0.0.0.0:8080`。

### 配置说明

编辑 `config.yaml`：

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  type: "mysql"
  host: "127.0.0.1"
  port: 3306
  username: "root"
  password: "your_password"
  dbname: "cloudkey"

redis:
  host: "127.0.0.1"
  port: 6379

auth:
  secret: "your-jwt-secret"
  expire_hours: 24
  super_admin:
    username: "admin"
    password: "your-password"
```

## 📡 API 概览

| 分组 | 前缀 | 认证方式 | 说明 |
|---|---|---|---|
| 公开 / 认证 | `/api/auth` | 限流，无 JWT | 登录、2FA 验证 |
| 超级管理员 | `/api/super` | JWT + super_admin | 租户管理、平台配置 |
| 租户管理员 | `/api/tenant` | JWT + tenant_admin | 卡密管理、统计、日志 |
| 服务账号 | `/api/service` | X-Service-Key Header | 程序化卡密管理 |

详细 API 文档请访问 `/swagger/index.html`（需开启 debug 模式）。

## 📄 许可证

本项目基于 [MIT 许可证](LICENSE) 开源。
