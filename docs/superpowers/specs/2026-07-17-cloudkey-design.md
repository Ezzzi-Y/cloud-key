# CloudKey 卡密发放与验证平台设计文档

## 1. 项目概述

### 1.1 项目目标

构建一个卡密发放和验证平台，用于分发和管理自建 API 服务的访问密钥。

### 1.2 核心场景

- 管理员创建卡密并分发给用户
- 用户在客户端使用卡密调用 API
- 管理员监控使用情况，处理违规行为

### 1.3 技术栈

- **后端：** Go + Gin + GORM
- **数据库：** MySQL 8.0+
- **认证：** JWT + TOTP
- **API 文档：** Swagger (Swaggo)
- **容器化：** Docker

---

## 2. 核心功能

### 2.1 卡密管理

**卡密格式：**
- 自定义前缀 + 随机字符串（如 `sk-a1b2c3d4e5f6`）
- 前缀可配置
- 保存 3-5 位后缀用于脱敏显示（默认 4 位）

**计费模式：**
- **按次数（count）** — 每次调用扣减 1 次
- **按 Credit（credit）** — 每次调用扣减指定数量

**卡密状态：**
- `unused` — 未使用，有剩余额度
- `used` — 已使用，额度用尽
- `disabled` — 已禁用（管理员手动禁用）
- `expired` — 已过期（预留状态）

**过期机制：**
- 无过期时间
- 额度用尽即失效

**卡密可见性：**
- 卡密明文仅在创建时返回一次
- 管理员无法查看完整卡密
- 管理员通过别名和后缀识别卡密

### 2.2 验证服务

**查询接口（不扣减）：**
- 输入：卡密
- 输出：别名、计费模式、剩余额度、状态、创建时间、最后使用时间

**扣减接口（先扣减再返回）：**
- 输入：卡密、扣减数量（默认 1）
- 输出：扣减后的剩余额度、状态、是否用尽
- 并发安全：使用数据库事务 + 乐观锁

### 2.3 使用记录

**记录内容：**
- 卡密 ID、卡密别名
- 扣减数量
- 调用 IP、User-Agent
- 请求路径、请求参数（可配置是否记录）
- 响应状态
- 使用时间

**查询功能：**
- 按卡密别名查询
- 按 IP 查询
- 按时间范围查询
- 导出使用记录

### 2.4 数据统计

**卡密概览：**
- 总卡密数
- 各状态卡密数量
- 总额度（次数/credit）

**使用统计：**
- 今日/本周/本月调用量
- 调用趋势图（按天/周/月）
- 热门卡密 TOP 10
- 热门 IP TOP 10

### 2.5 管理后台

**仪表盘：**
- 总卡密数（按状态分类）
- 今日/本周/本月调用量
- 调用趋势图
- 最近使用记录

**卡密管理：**
- 卡密列表（分页、筛选、搜索）
- 创建卡密（单个/批量）
- 查看详情、修改别名、调整额度
- 禁用/启用、删除
- 导出卡密

**使用记录：**
- 分页显示
- 筛选：按时间范围、别名、IP

**数据统计：**
- 卡密概览（饼图）
- 调用趋势（折线图）
- 热门卡密/IP TOP 10

**系统管理：**
- 修改密码、重新生成 TOTP
- 系统配置
- 服务账号管理
- 登录日志

### 2.6 服务账号

**用途：** 自动化创建卡密

**权限：**
- 创建卡密
- 查询自己创建的卡密列表

**限制：**
- 无法查看完整卡密（创建时返回一次）
- 无法禁用/删除卡密
- 无法查看使用记录和统计数据

**认证方式：** 专用服务账号密钥（通过 Header 传递）

---

## 3. API 接口设计

### 3.1 公开接口（卡密持有者使用）

**查询卡密状态（不扣减）：**
```
GET /api/key/status?sk=卡密
```

**扣减并查询（先扣减再返回）：**
```
POST /api/key/consume
{
  "key": "卡密",
  "amount": 1
}
```

### 3.2 管理接口（管理员使用）

**卡密管理：**
- `POST /api/admin/keys` — 创建卡密
- `GET /api/admin/keys` — 卡密列表
- `GET /api/admin/keys/:id` — 卡密详情
- `PATCH /api/admin/keys/:id` — 修改别名、调整额度
- `PATCH /api/admin/keys/:id/disable` — 禁用卡密
- `PATCH /api/admin/keys/:id/enable` — 启用卡密
- `DELETE /api/admin/keys/:id` — 删除卡密
- `GET /api/admin/keys/export` — 导出卡密

**使用记录：**
- `GET /api/admin/usage-logs` — 使用记录列表

**数据统计：**
- `GET /api/admin/stats/overview` — 卡密概览
- `GET /api/admin/stats/trends` — 调用趋势

**系统管理：**
- `POST /api/admin/login` — 管理员登录
- `POST /api/admin/login/verify-2fa` — 验证 TOTP
- `GET /api/admin/profile` — 管理员信息
- `PUT /api/admin/password` — 修改密码
- `GET /api/admin/configs` — 系统配置
- `PUT /api/admin/configs` — 更新配置

### 3.3 服务接口（自动化使用）

**创建卡密：**
```
POST /api/service/keys
Headers:
  X-Service-Key: 服务账号密钥
Body:
  {
    "alias": "别名",
    "billing_mode": "count",
    "initial_amount": 100
  }
```

**查询自己创建的卡密：**
```
GET /api/service/keys
Headers:
  X-Service-Key: 服务账号密钥
```

---

## 4. 业务流程

### 4.1 卡密创建流程

1. 管理员填写信息（别名、计费模式、额度）
2. 系统生成卡密（前缀 + 随机字符串）
3. 系统计算卡密哈希，存储到数据库
4. 系统返回卡密明文（仅此一次）
5. 管理员保存卡密并分发给用户

### 4.2 卡密验证流程（查询接口）

1. 用户请求，携带卡密
2. 系统计算卡密哈希
3. 查询数据库匹配
4. 返回卡密状态（不扣减）

### 4.3 卡密扣减流程（扣减接口）

1. 用户请求，携带卡密和扣减数量
2. 系统计算卡密哈希
3. 查询数据库匹配
4. 检查卡密状态（必须为 `unused`）
5. 检查剩余额度 ≥ 扣减数量
6. 扣减额度（事务 + 乐观锁）
7. 记录使用日志
8. 返回扣减后状态

### 4.4 管理员禁用流程

1. 管理员查看使用日志
2. 通过别名或 IP 识别滥用行为
3. 禁用相关卡密

### 4.5 管理员登录流程

1. 输入用户名/密码
2. 验证成功后要求输入 TOTP
3. TOTP 验证成功后返回 JWT Token
4. 后续请求携带 JWT Token

### 4.6 普通用户使用流程

1. 用户获得卡密（管理员分发）
2. 在客户端配置卡密
3. 调用 API 查询/扣减
4. 查看剩余额度

---

## 5. 数据模型

### 5.1 卡密表 (keys)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键 |
| alias | VARCHAR(255) | 别名 |
| key_hash | VARCHAR(255) | 卡密哈希值 |
| key_prefix | VARCHAR(50) | 卡密前缀 |
| key_suffix | VARCHAR(10) | 卡密后缀（3-5位） |
| billing_mode | VARCHAR(20) | 计费模式（count/credit） |
| initial_amount | BIGINT | 初始额度 |
| remaining_amount | BIGINT | 剩余额度 |
| status | VARCHAR(20) | 状态（unused/used/disabled/expired） |
| created_by | VARCHAR(100) | 创建者（管理员ID或服务账号ID） |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |
| used_at | DATETIME | 最后使用时间 |

### 5.2 使用记录表 (usage_logs)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键 |
| key_id | BIGINT | 卡密ID |
| key_alias | VARCHAR(255) | 卡密别名 |
| amount | BIGINT | 扣减数量 |
| ip | VARCHAR(50) | 调用IP |
| user_agent | VARCHAR(500) | User-Agent |
| request_path | VARCHAR(500) | 请求路径 |
| request_params | TEXT | 请求参数（可选） |
| response_status | INT | 响应状态 |
| created_at | DATETIME | 使用时间 |

### 5.3 管理员表 (admins)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键 |
| username | VARCHAR(100) | 用户名 |
| password_hash | VARCHAR(255) | 密码哈希 |
| totp_secret | VARCHAR(255) | TOTP密钥 |
| is_active | BOOLEAN | 是否启用 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 5.4 服务账号表 (service_accounts)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键 |
| name | VARCHAR(100) | 账号名称 |
| key_hash | VARCHAR(255) | 密钥哈希 |
| is_active | BOOLEAN | 是否启用 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 5.5 登录日志表 (login_logs)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键 |
| admin_id | BIGINT | 管理员ID |
| ip | VARCHAR(50) | 登录IP |
| user_agent | VARCHAR(500) | User-Agent |
| status | VARCHAR(20) | 登录状态（success/failed） |
| created_at | DATETIME | 登录时间 |

### 5.6 系统配置表 (configs)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键 |
| key | VARCHAR(100) | 配置键 |
| value | VARCHAR(500) | 配置值 |
| description | VARCHAR(500) | 配置说明 |
| updated_at | DATETIME | 更新时间 |

---

## 6. API 响应格式

### 6.1 统一响应结构

**成功响应：**
```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

**错误响应：**
```json
{
  "code": 1001,
  "message": "卡密不存在",
  "data": null
}
```

### 6.2 错误码定义

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1001 | 卡密不存在 |
| 1002 | 卡密已禁用 |
| 1003 | 卡密额度已用尽 |
| 1004 | 扣减数量超过剩余额度 |
| 2001 | 管理员账号或密码错误 |
| 2002 | TOTP 验证失败 |
| 2003 | JWT Token 无效或已过期 |
| 2004 | 无权限执行此操作 |
| 3001 | 服务账号密钥无效 |
| 9999 | 系统内部错误 |

### 6.3 分页响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [ ... ],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

---

## 7. 部署方案

### 7.1 Docker 部署

**docker-compose.yml：**
```yaml
services:
  cloudkey:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=your-mysql-host
      - DB_PORT=3306
      - DB_USER=cloudkey
      - DB_PASSWORD=xxx
      - DB_NAME=cloudkey
      - JWT_SECRET=xxx
      - ADMIN_USERNAME=admin
      - ADMIN_PASSWORD=xxx
```

### 7.2 环境变量配置

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| PORT | 服务端口 | 8080 |
| DB_HOST | MySQL 地址 | localhost |
| DB_PORT | MySQL 端口 | 3306 |
| DB_USER | MySQL 用户名 | cloudkey |
| DB_PASSWORD | MySQL 密码 | - |
| DB_NAME | 数据库名 | cloudkey |
| JWT_SECRET | JWT 密钥 | - |
| ADMIN_USERNAME | 初始管理员用户名 | admin |
| ADMIN_PASSWORD | 初始管理员密码 | - |

### 7.3 首次启动

1. 自动创建数据库表
2. 自动创建初始管理员账号
3. 首次登录时引导设置 TOTP

---

## 8. 安全设计

### 8.1 卡密存储安全

- 不存储明文，只存储哈希值（bcrypt 或 argon2）
- 卡密明文仅在创建时返回一次
- 存储 3-5 位后缀用于脱敏显示

### 8.2 接口安全

- 生产环境强制 HTTPS
- CORS 可配置
- 请求频率限制（预留接口）

### 8.3 管理后台安全

- 管理员密码 bcrypt 哈希存储
- TOTP 两步验证
- JWT Token 有效期可配置（默认 24 小时）
- 登录日志记录

### 8.4 数据库安全

- 参数化查询防止 SQL 注入
- 最小权限原则

### 8.5 日志安全

- 日志中不记录完整卡密
- 只记录后缀和别名

---

## 9. 配置项

| 配置键 | 说明 | 默认值 |
|--------|------|--------|
| key_prefix | 卡密默认前缀 | sk- |
| key_length | 卡密随机部分长度 | 32 |
| key_suffix_length | 卡密后缀长度 | 4 |
| record_request_params | 是否记录请求参数 | false |
| jwt_expire_hours | JWT 过期时间（小时） | 24 |
