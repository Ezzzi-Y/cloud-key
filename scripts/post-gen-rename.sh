#!/usr/bin/env bash
# post-gen-rename.sh — 重命名 OpenAPI Generator 生成的 SDK 工具类
# 用法: bash scripts/post-gen-rename.sh
# 在项目根目录运行，生成后自动执行一次即可

set -euo pipefail

BASE="sdk/java/src/main/java/com/github/ezzzi_y"
TEST="sdk/java/src/test/java/com/github/ezzzi_y"
DOCS="sdk/java/docs"
README="sdk/java/README.md"

echo "=== 1. 替换文件内容 ==="

# 构建搜索目录列表（忽略不存在的目录）
DIRS=("$BASE")
[ -d "$TEST" ] && DIRS+=("$TEST")

# 替换所有 .java 文件中的类引用
find "${DIRS[@]}" -name "*.java" -exec sed -i \
  -e 's/\bApiClient\b/CloudKeyClient/g' \
  -e 's/\bApiException\b/CloudKeyException/g' \
  -e 's/\bApiResponse\b/CloudKeyResponse/g' \
  -e 's/\bApiCallback\b/CloudKeyCallback/g' \
  {} +

# Configuration: 只替换 import 和静态调用，不影响 ServerConfiguration
find "${DIRS[@]}" -name "*.java" -exec sed -i \
  -e 's/Configuration\.getDefaultCloudKeyClient/CloudKeyConfiguration.getDefaultCloudKeyClient/g' \
  -e 's/Configuration\.getDefaultApiClient/CloudKeyConfiguration.getDefaultApiClient/g' \
  -e 's/public class Configuration /public class CloudKeyConfiguration /g' \
  -e 's/private Configuration()/private CloudKeyConfiguration()/g' \
  {} +

# Configuration: 更精确的 import 替换
find "${DIRS[@]}" -name "*.java" -exec sed -i \
  's/import com\.github\.ezzzi_y\.Configuration;/import com.github.ezzzi_y.CloudKeyConfiguration;/g' \
  {} +

# JSON 类名: 只替换作为类引用的 JSON（大写，word boundary）
find "${DIRS[@]}" -name "*.java" -exec sed -i \
  -e 's/public class JSON {/public class CloudKeyJSON {/g' \
  -e 's/import com\.github\.ezzzi_y\.JSON;/import com.github.ezzzi_y.CloudKeyJSON;/g' \
  -e 's/return JSON\.getGson/return CloudKeyJSON.getGson/g' \
  -e 's/JSON\.getGson/CloudKeyJSON.getGson/g' \
  -e 's/protected JSON json;/protected CloudKeyJSON json;/g' \
  -e 's/public JSON getJSON/public CloudKeyJSON getJSON/g' \
  -e 's/setJSON(JSON json)/setJSON(CloudKeyJSON json)/g' \
  {} +

# CloudKeyClient 内部 json 字段引用（避免 CloudKeyJSON.xxx 残留）
find "$BASE" -name "ApiClient.java" -exec sed -i \
  -e 's/CloudKeyJSON\.setDateFormat/json.setDateFormat/g' \
  -e 's/CloudKeyJSON\.setSqlDateFormat/json.setSqlDateFormat/g' \
  -e 's/CloudKeyJSON\.setOffsetDateTimeFormat/json.setOffsetDateTimeFormat/g' \
  -e 's/CloudKeyJSON\.setLocalDateFormat/json.setLocalDateFormat/g' \
  -e 's/CloudKeyJSON\.setLenientOnJson/json.setLenientOnJson/g' \
  -e 's/CloudKeyJSON\.serialize/json.serialize/g' \
  -e 's/CloudKeyJSON\.deserialize/json.deserialize/g' \
  -e 's/application\/CloudKeyJSON/application\/json/g' \
  {} +

# 修复 docs 和 README
find "$DOCS" -name "*.md" -exec sed -i \
  -e 's/\bApiClient\b/CloudKeyClient/g' \
  -e 's/\bApiException\b/CloudKeyException/g' \
  -e 's/Configuration\.getDefaultCloudKeyClient/CloudKeyConfiguration.getDefaultCloudKeyClient/g' \
  -e 's/Configuration\.getDefaultApiClient/CloudKeyConfiguration.getDefaultApiClient/g' \
  {} +

if [ -f "$README" ]; then
  sed -i \
    -e 's/\bApiClient\b/CloudKeyClient/g' \
    -e 's/\bApiException\b/CloudKeyException/g' \
    -e 's/Configuration\.getDefaultCloudKeyClient/CloudKeyConfiguration.getDefaultCloudKeyClient/g' \
    -e 's/Configuration\.getDefaultApiClient/CloudKeyConfiguration.getDefaultApiClient/g' \
    "$README"
fi

echo "=== 2. 重命名文件 ==="

for pair in \
  "ApiClient.java:CloudKeyClient.java" \
  "ApiException.java:CloudKeyException.java" \
  "ApiResponse.java:CloudKeyResponse.java" \
  "ApiCallback.java:CloudKeyCallback.java" \
  "Configuration.java:CloudKeyConfiguration.java" \
  "JSON.java:CloudKeyJSON.java"
do
  old="${pair%%:*}"
  new="${pair##*:}"
  [ -f "$BASE/$old" ] && mv "$BASE/$old" "$BASE/$new" && echo "  $old -> $new"
done

echo "=== 3. 清理残留 ==="

# 修复双重替换（只在目录存在时搜索）
DIRS_TO_FIX="$BASE"
[ -d "$TEST" ] && DIRS_TO_FIX="$DIRS_TO_FIX $TEST"

find $DIRS_TO_FIX -name "*.java" -exec sed -i \
  -e 's/CloudKeyCloudKeyJSON/CloudKeyJSON/g' \
  -e 's/CloudKeyCloudKeyClient/CloudKeyClient/g' \
  -e 's/CloudKeyCloudKeyException/CloudKeyException/g' \
  -e 's/CloudKeyCloudKeyResponse/CloudKeyResponse/g' \
  -e 's/CloudKeyCloudKeyCallback/CloudKeyCallback/g' \
  {} +

echo "=== Done ==="
