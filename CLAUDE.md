# CloudKey 项目开发规范

## 项目概述

CloudKey 是一个卡密管理系统，提供 API Key 的创建、消费、额度管理等功能。后端使用 Go (Gin + GORM)，前端为内嵌 React SPA，SDK 为 Java (OkHttp)。

## 目录结构

```
├── main.go                    # 入口，依赖注入
├── internal/
│   ├── model/                 # 数据模型（GORM AutoMigrate）
│   ├── service/               # 业务逻辑
│   ├── handler/               # HTTP 处理器（Gin handlers）
│   ├── router/                # 路由注册
│   ├── errcode/               # 错误码定义
│   ├── middleware/             # 中间件（JWT、限流、BusinessGuard）
│   ├── config/                # 配置加载
│   ├── database/              # 数据库连接
│   ├── log/                   # 日志
│   └── web/                   # 嵌入的前端 SPA
├── sdk/
│   ├── java/
│   │   ├── api/openapi.yaml       # SDK 专用 OpenAPI 3.0 spec（手动维护）
│   │   ├── src/main/java/com/github/ezzzi_y/
│   │   │   ├── CloudKey.java      # SDK 入口 [手写]
│   │   │   ├── gen/               # OpenAPI 生成的通信层（不手动修改）
│   │   │   └── service/           # 手写 SDK 服务层 + 精简 model
│   │   ├── build.gradle           # Gradle 构建
│   │   └── pom.xml                # Maven 构建
│   └── python/
│       ├── pyproject.toml         # Python 包构建配置
│       ├── README.md              # 使用文档
│       └── cloudkey/              # SDK 源码
│           ├── __init__.py        # SDK 入口 + CloudKey 主类
│           ├── _client.py         # HTTP 客户端（httpx 封装）
│           ├── _exceptions.py     # CloudKeyException
│           ├── _models.py         # dataclass DTO
│           ├── key_service.py     # KeyService 便捷层
│           └── balance_log_service.py  # BalanceLogService 便捷层
├── scripts/regenerate-sdk.sh  # SDK gen/ 层重新生成脚本
├── docs/                      # swaggo 生成的 Swagger 2.0 spec
└── web/                       # React 前端源码
```

## 两套 API Spec

项目有**两套独立的 OpenAPI 规范**，需手动保持同步：

1. **Go Server Spec** (`docs/swagger.yaml`) — 由 `swag init` 从 Go 注释自动生成，覆盖全部端点
2. **Java SDK Spec** (`sdk/java/api/openapi.yaml`) — **手动维护**，仅覆盖 `/service/*` 端点

修改 API 时必须同时更新两者，**以及 Python SDK 的 `sdk/python/cloudkey/` 下对应的服务层代码**。

## Go 后端开发流程

新增或修改功能时，按以下层次顺序操作：

### 1. Model (`internal/model/`)
- 定义 GORM 结构体，实现 `TableName()` 方法
- 在 `internal/model/migrate.go` 的 `AutoMigrate()` 中注册新模型

### 2. Errcode (`internal/errcode/errcode.go`)
- 新增错误码常量（按类别分段：卡密 1001~、认证 2001~、服务账号 3001~、租户 4001~）
- 在 `codeMessages` map 中添加中文消息

### 3. Service (`internal/service/`)
- 纯业务逻辑，不依赖 Gin
- 使用 `*gorm.DB` 操作数据库
- 并发安全操作使用乐观锁（`version` 字段 + 重试循环，参考 `ConsumeKey`）

### 4. Handler (`internal/handler/`)
- 请求绑定、参数校验、调用 Service、返回响应
- 使用 `getTenantID(c)` 从 JWT 提取租户 ID
- 使用 `getServiceTenantID(c)` 从 X-Service-Key 提取租户 ID
- 响应辅助函数：`Success()`, `BadRequest()`, `InternalError()`, `NotFound()`, `Unauthorized()`, `SuccessPaginated()`
- swaggo 注释格式参考已有 handler

### 5. Router (`internal/router/router.go`)
- `SetupRouter()` 函数中注册新路由
- 租户路由需 `middleware.AuthMiddleware` + `middleware.RequireTenantAdmin`
- 写操作加 `middleware.TenantBusinessGuard(db)`
- 服务账号路由需 `middleware.ServiceAuthMiddleware`

### 6. Main (`main.go`)
- 实例化新 Service 和 Handler
- 注入到 `router.SetupRouter()`

## Java SDK 架构

SDK 采用分层架构：`gen/` 放 OpenAPI 生成的通信层，手写 `service/` 层提供面向用户的 API。

```
com.github.ezzzi_y/
  CloudKey.java                 # SDK 入口 [手写]
  CloudKeyOptions.java          # 配置项 [手写]
  CloudKeyException.java        # SDK 异常 [手写]
  gen/                          # OpenAPI 生成，不手动修改
    CloudKeyClient.java         # OkHttp 客户端
    api/                        # ServiceKeysApi, BalanceLogsApi
    model/                      # 全部生成的 model
    auth/                       # 认证
  service/                      # 手写 SDK 服务层
    KeyService.java
    BalanceLogService.java
    model/                      # 精简版 SDK model
```

### SDK 使用方式

```java
CloudKey ck = new CloudKey("sk_your_service_key");
ck.keys().create("my-key", 100L);
ck.keys().consume("ck_abc123", 10L);
ck.keys().adjustBalance(1, 50L, "充值");
ck.balanceLogs().list(q -> q.page(1).pageSize(20));
```

### Java SDK 更新流程

当 `/service/*` 路由变化时，更新 SDK：

#### 步骤 1: 更新 OpenAPI Spec
编辑 `sdk/java/api/openapi.yaml`

#### 步骤 2: 更新 SDK 版本号
修改 `build.gradle` 和 `pom.xml` 中的 version

#### 步骤 3: 重新生成 gen/ 通信层
```bash
bash scripts/regenerate-sdk.sh
```

#### 步骤 4: 如有新端点，更新手写 SDK 层
在 `service/KeyService.java` 或 `service/BalanceLogService.java` 中添加对应方法

#### 步骤 5: 验证编译
```bash
cd sdk/java && ./gradlew build -x test
```

#### 步骤 6: 更新前端文档
更新 `web/src/components/ServiceApiDocs.tsx` 中的 Java 示例代码

## Python SDK 架构

Python SDK 采用全手写方案（不用 OpenAPI Generator），HTTP 客户端基于 `httpx`，model 使用 `dataclasses`。

```
sdk/python/
  pyproject.toml               # 构建配置（hatchling）
  README.md                    # 使用文档
  cloudkey/
    __init__.py                # SDK 入口 + CloudKey 主类
    _client.py                 # HTTP 客户端（httpx 封装）
    _exceptions.py             # CloudKeyException
    _models.py                 # dataclass DTO
    key_service.py             # KeyService（卡密管理）
    balance_log_service.py     # BalanceLogService（余额日志）
```

### Python SDK 使用方式

```python
from cloudkey import CloudKey

ck = CloudKey("sk_your_service_key")
ck.keys.create("my-key", 100)
ck.keys.consume("ck_abc123", amount=10)
ck.keys.adjust_balance(1, 50, "充值")
ck.balance_logs.list(page=1, page_size=20)
ck.close()
```

### Python SDK 更新流程

当 `/service/*` 路由变化时，同步更新 Python SDK：

1. 更新 `sdk/python/cloudkey/key_service.py` 或 `balance_log_service.py` 中的方法
2. 如有新响应字段，更新 `_models.py` 中的 dataclass
3. 更新 `pyproject.toml` 中的版本号
4. 更新 README.md 中的 API 参考
5. 验证：`cd sdk/python && pip install -e . && python -c "from cloudkey import CloudKey"`

发布：推送 `python-sdk-v*` tag 触发 GitHub Actions 自动发布到 PyPI。

## 构建验证

每次变更后必须验证：
```bash
# Go 后端
go build ./...

# Java SDK
cd sdk/java && ./gradlew build -x test

# Python SDK
cd sdk/python && pip install -e . && python -c "from cloudkey import CloudKey"
```

## 关键约定

- **额度语义分离**：消费（consume）记录在 `usage_logs`；管理调整（adjust-balance）记录在 `balance_logs`
- **额度只能增减**：不允许直接设置 remaining_amount，只能通过 delta 增加或减少
- **乐观锁**：所有余额变更使用 `version` 字段 + 重试（max 3）
- **租户隔离**：所有查询都带 `tenant_id` 条件
- **CloudKey 前缀**：SDK 工具类统一使用 CloudKey 前缀（CloudKeyClient、CloudKeyException 等）
- **错误码分段**：卡密 1001~、认证 2001~、服务账号 3001~、租户 4001~、权限 5001~、安全 6001~、系统 9999
