# CloudKey

多租户 SaaS 平台，专注于 API Key / 卡密的全生命周期管理。适用于 API 按量计费、虚拟产品兑换码等场景。

## 📖 About

CloudKey 提供 API Key 的创建、分发、验证、扣减和审计全链路能力。管理员通过 Web UI 管理卡密与租户；终端用户或自动化系统通过公开 API 进行卡密验证和消费。

## ✨ 核心特性

- **多租户 SaaS 架构** — 每个租户独立管理卡密、用户、服务账号，支持自定义卡密前缀、长度和后缀长度
- **高性能消费路径** — Redis Lua 脚本原子扣减 + RabbitMQ 异步日志持久化，乐观锁保障数据库一致性
- **卡密安全管理** — 卡密仅存储哈希值，明文仅在创建时返回一次；支持 AES-256-GCM 可选加密
- **余额调整审计** — 消费用 `usage_logs`，管理调整用 `balance_logs`，语义分离；余额仅允许 delta 增减
- **服务账号（Service Account）** — 机器对机器的 API 访问，通过 `X-Service-Key` 头认证
- **双因素认证（2FA）** — 基于 TOTP 的两步验证，支持登录和设置流程
- **租户生命周期管理** — 三层状态（active / expired / disabled），过期租户仅可查看统计，禁用租户完全封锁
- **数据统计仪表盘** — 卡密数量分布、调用趋势（日/周/月）、热门卡密 Top 10、额度消耗 Top 10
- **使用日志与导出** — 完整的消费和余额变更审计日志，支持筛选，CSV / JSON 双格式导出
- **前端嵌入式部署** — React SPA 编译后嵌入 Go 二进制文件，单文件部署
- **Java SDK** — 通过 OpenAPI Generator 自动生成，发布到 GitHub Packages
- **Python SDK** — 基于 httpx，发布到 PyPI

## 🛠 技术栈

### 后端

| 组件 | 技术 |
|---|---|
| 语言 | Go 1.25 |
| HTTP 框架 | Gin |
| ORM | GORM (MySQL) |
| 缓存 / 限流 | Redis (go-redis v9) + Lua 脚本 |
| 消息队列 | RabbitMQ (amqp091-go) |
| 认证 | JWT (golang-jwt v5) + TOTP (pquerna/otp) |
| 日志 | Zap + Lumberjack |
| 配置 | Viper (YAML) |
| API 文档 | Swaggo (Swagger 2.0) |
| 加密 | AES-256-GCM（可选） |

### 前端

| 组件 | 技术 |
|---|---|
| 框架 | React 18 + TypeScript |
| 构建工具 | Vite 5 |
| 样式 | Tailwind CSS 3 |
| 组件库 | Radix UI + shadcn/ui |
| 数据请求 | TanStack React Query 5 + Axios |
| 图表 | Recharts |
| 路由 | React Router v6 |
| Toast | Sonner |

### 基础设施

| 组件 | 技术 |
|---|---|
| 数据库 | MySQL 8.0+ |
| 缓存 / 消息 | Redis 6.0+、RabbitMQ |
| 容器化 | Docker (多阶段构建) + Docker Compose |
| CI/CD | GitHub Actions (Java SDK → GitHub Packages, Python SDK → PyPI) |

## 📁 项目结构

```
CloudKey/
├── main.go                  # 应用入口，依赖注入
├── config.yaml.example      # 配置示例
├── Dockerfile               # 多阶段构建（Node → Go → Alpine）
├── docker-compose.yml       # Docker Compose 编排
├── internal/                # 后端 Go 源码
│   ├── config/              # Viper 配置加载
│   ├── database/            # MySQL & Redis 连接池
│   ├── model/               # 数据模型 (GORM, 8 个表)
│   ├── service/             # 业务逻辑层（含 MQ 生产/消费）
│   ├── handler/             # HTTP 处理器
│   ├── router/              # 路由定义
│   ├── middleware/           # 中间件（JWT、限流、角色、租户守卫）
│   ├── errcode/             # 统一错误码（分段：1001~ 卡密、2001~ 认证...）
│   ├── log/                 # 日志初始化
│   └── web/                 # 前端 embed.FS 嵌入
├── web/                     # React SPA 前端源码
├── sdk/
│   ├── java/                # Java SDK
│   │   ├── api/openapi.yaml # SDK 专用 OpenAPI 3.0 spec
│   │   └── src/...          # SDK 源码
│   └── python/              # Python SDK
│       ├── pyproject.toml   # 构建配置
│       └── cloudkey/        # SDK 源码
├── scripts/                 # SDK 生成后重命名脚本
├── docs/                    # Swaggo 生成的 Swagger 2.0 文档
└── .github/workflows/       # CI（SDK 发布到 GitHub Packages）
```



## 🚀 快速开始

### 环境要求

- Go 1.25+
- MySQL 8.0+
- Redis 6.0+
- RabbitMQ 3.x+
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
app:
  debug: false

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
  password: ""
  db: 0

rabbitmq:
  host: "127.0.0.1"
  port: 5672
  username: "guest"
  password: "guest"
  vhost: "/"

auth:
  secret: "your-jwt-secret"
  expire_hours: 24
  super_admin:
    username: "admin"
    password: "your-password"

log:
  level: "info"
  format: "json"
  output: "both"
  file:
    filename: "logs/cloudkey.log"
    max_size: 100
    max_backups: 30

security:
  encryption:
    enabled: false
    key: "your-32-byte-encryption-key-here"
```

## 📡 API 概览

| 分组 | 前缀 | 认证方式 | 说明 |
|---|---|---|---|
| 公开 / 认证 | `/api/auth` | 限流（5 次/分/IP），无 JWT | 登录、2FA 验证、TOTP 初始化 |
| 超级管理员 | `/api/super` | JWT + super_admin 角色 | 租户 CRUD、平台配置、登录日志 |
| 租户管理员 | `/api/tenant` | JWT + tenant_admin 角色 | 卡密管理、服务账号、统计、日志导出 |
| 服务账号 | `/api/service` | X-Service-Key Header | 程序化卡密管理、余额日志 |

### 核心端点

**卡密管理（租户 / 服务账号共用能力）：**
- 创建卡密、批量查询、单条查询、更新
- 卡密消费（Redis Lua 原子扣减 → RabbitMQ 异步记账）
- 余额调整（仅允许 delta 增减，记录 balance_logs）
- 卡密状态查询、启用/禁用、删除
- 卡密导出（CSV / JSON）

**统计与日志：**
- 仪表盘（总览、趋势、分布）
- Top 10 热门卡密 & 额度消耗排行（Redis 缓存）
- 消费日志 & 余额变更日志（筛选 + 导出）

详细 API 文档请访问 `/swagger/index.html`（需开启 debug 模式）。

## ☕ SDK

支持 **Java** 和 **Python** 两种 SDK，覆盖 `/service/*` 端点。Java SDK 发布到 GitHub Packages，Python SDK 发布到 PyPI。

```bash
# Python
pip install cloudkey-sdk
```

```xml
<!-- Java (Maven) -->
<dependency>
  <groupId>com.github.ezzzi-y</groupId>
  <artifactId>cloudkey-client</artifactId>
  <version>LATEST</version>
</dependency>
```

详见各 SDK 目录下的 README。

## 📄 许可证

本项目基于 [MIT 许可证](LICENSE) 开源。
