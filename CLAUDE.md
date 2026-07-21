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
├── sdk/java/
│   ├── api/openapi.yaml       # SDK 专用 OpenAPI 3.0 spec（手动维护）
│   ├── templates/             # 自定义 Mustache 模板（CloudKey 前缀）
│   ├── src/main/java/com/github/ezzzi_y/  # SDK 源码
│   ├── build.gradle           # Gradle 构建
│   └── pom.xml                # Maven 构建
├── scripts/post-gen-rename.sh # SDK 生成后重命名脚本
├── docs/                      # swaggo 生成的 Swagger 2.0 spec
└── web/                       # React 前端源码
```

## 两套 API Spec

项目有**两套独立的 OpenAPI 规范**，需手动保持同步：

1. **Go Server Spec** (`docs/swagger.yaml`) — 由 `swag init` 从 Go 注释自动生成，覆盖全部端点
2. **Java SDK Spec** (`sdk/java/api/openapi.yaml`) — **手动维护**，仅覆盖 `/service/*` 端点

修改 API 时必须同时更新两者。

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

## Java SDK 更新流程

当 `/service/*` 路由发生变化时，必须更新 SDK。完整流程：

### 步骤 1: 更新 OpenAPI Spec
编辑 `sdk/java/api/openapi.yaml`：
- 新增/修改 paths、schemas
- 确保 operationId 使用 camelCase
- 更新 `info.version`

### 步骤 2: 更新 SDK 版本号
同时修改：
- `sdk/java/build.gradle` 中的 `version`
- `sdk/java/pom.xml` 中的 `<version>`

### 步骤 3: 生成 SDK
```bash
cd sdk/java
openapi-generator-cli generate -i api/openapi.yaml -g java -o . --template-dir templates
```

### 步骤 4: 运行重命名脚本
```bash
bash scripts/post-gen-rename.sh
```
此脚本将 `ApiClient` → `CloudKeyClient` 等工具类重命名。

### 步骤 5: 复制新文件到正确包路径
生成器默认输出到 `org/openapitools/client/`，需要手动复制到 `com/github/ezzzi_y/`：

```bash
cd sdk/java/src/main/java

# 复制新增的 model 文件
for f in org/openapitools/client/model/<NewModel>.java; do
  fname=$(basename "$f")
  cp "$f" "com/github/ezzzi_y/model/$fname"
  sed -i 's/package org\.openapitools\.client\.model;/package com.github.ezzzi_y.model;/' "com/github/ezzzi_y/model/$fname"
  sed -i 's/import org\.openapitools\.client/import com.github.ezzzi_y/g' "com/github/ezzzi_y/model/$fname"
done

# 复制新增/更新的 api 文件
for f in org/openapitools/client/api/<UpdatedApi>.java; do
  fname=$(basename "$f")
  cp "$f" "com/github/ezzzi_y/api/$fname"
  sed -i 's/package org\.openapitools\.client\.api;/package com.github.ezzzi_y.api;/' "com/github/ezzzi_y/api/$fname"
  sed -i 's/import org\.openapitools\.client/import com.github.ezzzi_y/g' "com/github/ezzzi_y/api/$fname"
done
```

### 步骤 6: 再次运行重命名脚本
```bash
bash scripts/post-gen-rename.sh
```

### 步骤 7: 清理生成器临时目录
```bash
rm -rf sdk/java/src/main/java/org
rm -rf sdk/java/src/test/java/org
```

### 步骤 8: 验证编译
```bash
cd sdk/java && ./gradlew build -x test
```

## 构建验证

每次变更后必须验证：
```bash
# Go 后端
go build ./...

# Java SDK
cd sdk/java && ./gradlew build -x test
```

## 关键约定

- **额度语义分离**：消费（consume）记录在 `usage_logs`；管理调整（adjust-balance）记录在 `balance_logs`
- **额度只能增减**：不允许直接设置 remaining_amount，只能通过 delta 增加或减少
- **乐观锁**：所有余额变更使用 `version` 字段 + 重试（max 3）
- **租户隔离**：所有查询都带 `tenant_id` 条件
- **CloudKey 前缀**：SDK 工具类统一使用 CloudKey 前缀（CloudKeyClient、CloudKeyException 等）
- **错误码分段**：卡密 1001~、认证 2001~、服务账号 3001~、租户 4001~、权限 5001~、安全 6001~、系统 9999
