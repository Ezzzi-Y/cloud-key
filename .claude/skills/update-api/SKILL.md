---
name: update-api
description: 当用户要求修改项目的API，或者是修改代码时Go项目的API发生变化时调用，涉及Go项目、OpenAPI、SDK生成流程。
---
# 更新 API

CloudKey 的 API 变更涉及两条独立的流水线，按需执行：

1. **Go 后端** — 新增或修改端点（model → service → handler → router → main.go）
2. **Java SDK** — 当 `/service/*` 路由变更时，同步更新 SDK

两套 API Spec 需手动保持同步：
- `docs/swagger.yaml` — 由 `swag init` 自动生成，覆盖全部端点
- `sdk/java/api/openapi.yaml` — **手动维护**，仅覆盖 `/service/*` 端点

---

## 执行计划

根据用户需求判断需要执行哪些步骤：

| 场景                        | 执行                                |
|---------------------------|-----------------------------------|
| 新增/修改仅租户端点（`/tenant/*`）   | Go 后端 → swag init                 |
| 新增/修改服务账号端点（`/service/*`） | Go 后端 → swag init → **SDK 更新**    |
| 仅调整响应结构                   | 更新对应 handler + SDK（如适用）           |
| 该能力属于租户和服务账号              | 租户和服务账号的Handler分开，但是Service代码共用一套 |
---

## Go 后端开发

按 model → errcode → service → handler → router → main.go 的层次顺序操作。详见 `go-backend.md`。

每步完成后验证编译：`go build ./...`

---

## Java SDK 更新

当 `/service/*` 路由或其请求/响应结构发生变化时执行。详见 `references/sdk-update.md`。

关键步骤摘要：
1. 编辑 `sdk/java/api/openapi.yaml`（路径 + schema）
2. 更新 `build.gradle` 和 `pom.xml` 版本号
3. `openapi-generator-cli generate`
4. `bash scripts/post-gen-rename.sh`
5. 将新文件从 `org/openapitools/` 复制到 `com/github/ezzzi_y/` 并修正包名
6. 再次运行重命名脚本
7. 清理 `org/` 目录
8. `./gradlew build -x test`
9. 打`tag`，发布到Github Packages
---
