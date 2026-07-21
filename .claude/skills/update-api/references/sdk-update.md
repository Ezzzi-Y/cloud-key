# Java SDK 更新流程

当 `/service/*` 路由或其请求/响应结构发生变化时，按此流程同步更新 Java SDK。

## 1. 更新 OpenAPI Spec

编辑 `sdk/java/api/openapi.yaml`：

- 新增/修改/删除对应的 paths
- 新增/修改对应的 `components/schemas`
- `operationId` 使用 camelCase
- 响应 schema 使用 `allOf` + `Result` 组合模式（参考已有 ConsumeResponse）
- 更新 `info.version`

## 2. 更新版本号

同时修改两处：
- `sdk/java/build.gradle` → `version = 'x.y.z'`
- `sdk/java/pom.xml` → `<version>x.y.z</version><!-- sdk-version -->`

## 3. 生成 SDK

```bash
cd sdk/java
openapi-generator-cli generate -i api/openapi.yaml -g java -o . --template-dir templates
```

## 4. 运行重命名脚本

```bash
bash scripts/post-gen-rename.sh
```

将工具类重命名为 CloudKey 前缀（ApiClient→CloudKeyClient 等）。

## 5. 复制新文件到正确包路径

生成器默认输出到 `org/openapitools/client/`，项目使用 `com.github.ezzzi_y` 包。

**仅复制新增或有变更的文件**：

```bash
cd sdk/java/src/main/java

# model 文件
for f in org/openapitools/client/model/<ModelName>.java; do
  fname=$(basename "$f")
  cp "$f" "com/github/ezzzi_y/model/$fname"
  sed -i 's/package org\.openapitools\.client\.model;/package com.github.ezzzi_y.model;/' "com/github/ezzzi_y/model/$fname"
  sed -i 's/import org\.openapitools\.client/import com.github.ezzzi_y/g' "com/github/ezzzi_y/model/$fname"
done

# api 文件
for f in org/openapitools/client/api/<ApiName>.java; do
  fname=$(basename "$f")
  cp "$f" "com/github/ezzzi_y/api/$fname"
  sed -i 's/package org\.openapitools\.client\.api;/package com.github.ezzzi_y.api;/' "com/github/ezzzi_y/api/$fname"
  sed -i 's/import org\.openapitools\.client/import com.github.ezzzi_y/g' "com/github/ezzzi_y/api/$fname"
done
```

## 6. 再次运行重命名脚本

```bash
bash scripts/post-gen-rename.sh
```

确保新文件中的 CloudKey 前缀也被正确替换。

## 7. 清理生成器临时目录

```bash
rm -rf sdk/java/src/main/java/org
rm -rf sdk/java/src/test/java/org
```

## 8. 验证编译

```bash
cd sdk/java && ./gradlew build -x test
```

## 注意事项

- `sdk/java/.openapi-generator-ignore` 中的文件不会被生成器覆盖
- `sdk/java/templates/` 中的自定义 Mustache 模板会覆盖默认模板，不要修改
- 新增全新 API tag 会生成新的 `XxxApi.java`，需同步复制
- 每次只复制有变更的文件，避免丢失手动修改
