#!/usr/bin/env bash
# regenerate-sdk.sh — 重新生成 SDK 的 gen/ 通信层代码
# 用法: bash scripts/regenerate-sdk.sh
# 前提: 已安装 openapi-generator-cli
#
# 工作流程:
#   1. 清理旧的 gen/ 目录
#   2. 用 OpenAPI Generator 生成到临时目录
#   3. 移动生成代码到 gen/ 包
#   4. 更新 package 声明（com.github.ezzzi_y → com.github.ezzzi_y.gen）

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SPEC="$PROJECT_ROOT/sdk/java/api/openapi.yaml"
GEN_SRC="$PROJECT_ROOT/sdk/java/src/main/java/com/github/ezzzi_y/gen"
GEN_TEST="$PROJECT_ROOT/sdk/java/src/test/java/com/github/ezzzi_y/gen"
TEMP_OUT="$PROJECT_ROOT/sdk/java/.gen-temp"

echo "=== 1. 清理旧的 gen/ 目录 ==="
rm -rf "$GEN_SRC" "$GEN_TEST"
rm -rf "$TEMP_OUT"
mkdir -p "$TEMP_OUT"

echo "=== 2. OpenAPI Generator 生成 ==="
openapi-generator-cli generate \
  -i "$SPEC" \
  -g java \
  -o "$TEMP_OUT" \
  --additional-properties=library=okhttp-gson \
  --additional-properties=java21=true \
  --additional-properties=dateLibrary=java8 \
  --additional-properties=booleanGetterPrefix=is \
  --model-package com.github.ezzzi_y.gen.model \
  --api-package com.github.ezzzi_y.gen.api \
  --invoker-package com.github.ezzzi_y.gen \
  --global-property=models,apis,supportingFiles \
  -p hideGenerationTimestamp=true

echo "=== 3. 移动生成代码到 gen/ ==="
# 移动模型和 API 类
mkdir -p "$GEN_SRC/model" "$GEN_SRC/api" "$GEN_SRC/auth"
cp -r "$TEMP_OUT/src/main/java/com/github/ezzzi_y/gen/model/"* "$GEN_SRC/model/"
cp -r "$TEMP_OUT/src/main/java/com/github/ezzzi_y/gen/api/"* "$GEN_SRC/api/"

# 移动 supporting files（ApiClient, JSON, auth 等）
SUPPORT_SRC="$TEMP_OUT/src/main/java/com/github/ezzzi_y/gen"
for f in "$SUPPORT_SRC"/*.java; do
  [ -f "$f" ] && cp "$f" "$GEN_SRC/"
done
[ -d "$SUPPORT_SRC/auth" ] && cp -r "$SUPPORT_SRC/auth/"* "$GEN_SRC/auth/"

# 移动测试文件（如果存在）
TEMP_TEST="$TEMP_OUT/src/test/java/com/github/ezzzi_y/gen"
if [ -d "$TEMP_TEST" ]; then
  mkdir -p "$GEN_TEST/model" "$GEN_TEST/api"
  [ -d "$TEMP_TEST/model" ] && cp -r "$TEMP_TEST/model/"* "$GEN_TEST/model/"
  [ -d "$TEMP_TEST/api" ] && cp -r "$TEMP_TEST/api/"* "$GEN_TEST/api/"
fi

echo "=== 4. 清理临时目录 ==="
rm -rf "$TEMP_OUT"

echo "=== Done ==="
echo ""
echo "生成文件位于: $GEN_SRC"
echo "请运行 './gradlew compileJava' 验证编译"
