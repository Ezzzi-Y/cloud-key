# CloudKey Python SDK

卡密管理系统（CloudKey）的 Python 服务账号客户端。

## 安装

```bash
pip install cloudkey-python
```

## 快速开始

```python
from cloudkey import CloudKey, CloudKeyException

ck = CloudKey("sk_your_service_key")

# 创建卡密
key = ck.keys.create("my-key", 100)
print(key.raw_key)  # 仅创建时可见，请妥善保存

# 消费额度
result = ck.keys.consume(key.raw_key, amount=10)
print(result.remaining_amount, result.exhausted)

# 查询状态
status = ck.keys.get_status(key.raw_key)
print(status.alias, status.remaining_amount)

# 调整额度（管理员操作）
adj = ck.keys.adjust_balance(key.id, delta=50, remark="充值")
print(adj.before_amount, adj.after_amount)

# 查询余额日志
logs = ck.balance_logs.list(page=1, page_size=20)
for log in logs.items:
    print(f"{log.key_alias}: {log.delta:+d} ({log.remark})")

# 关闭客户端
ck.close()
```

也可以使用 context manager：

```python
with CloudKey("sk_your_service_key") as ck:
    key = ck.keys.create("my-key", 100)
```

## 自定义配置

```python
ck = CloudKey(
    "sk_your_service_key",
    base_url="https://your-server.com/api",
    connect_timeout=5.0,
    read_timeout=60.0,
)
```

## API 参考

### `ck.keys` — 卡密管理

#### 创建卡密

```python
result = ck.keys.create("my-key", 100)
# result: CreateKeyResult(id, raw_key, alias, key_suffix, remaining_amount, status)
```

可选参数 `rate_limit`、`rate_limit_window` 控制单卡限流。

#### 消费额度

```python
result = ck.keys.consume("ck_xxx")           # 默认消费 1
result = ck.keys.consume("ck_xxx", amount=10)  # 指定数量
result = ck.keys.consume("ck_xxx", amount=10, request_id="my-id")  # 指定幂等 ID
# result: ConsumeResult(request_id, remaining_amount, status, exhausted)
```

#### 调整额度

```python
adj = ck.keys.adjust_balance(key_id=1, delta=50, remark="充值")
adj = ck.keys.adjust_balance(key_id=1, delta=-20, remark="扣减", request_id="my-id")
# adj: AdjustResult(request_id, before_amount, after_amount)
```

#### 查询卡密详情

```python
info = ck.keys.get(key_id=1)
# info: KeyInfo(id, alias, status, remaining_amount, created_at, expire_at)
```

#### 查询卡密状态（按密钥值）

```python
status = ck.keys.get_status("ck_xxx")
# status: KeyStatus(alias, status, remaining_amount)
# status.is_exhausted -> bool
```

#### 分页列表

```python
page = ck.keys.list(page=1, page_size=20, status="active", alias="my-key")
# page: PageResult[KeyInfo]
# page.items, page.total, page.page, page.page_size, page.total_pages
```

#### 更新 / 启用 / 禁用 / 删除

```python
ck.keys.update(key_id=1, alias="new-name")
ck.keys.enable(key_id=1)
ck.keys.disable(key_id=1)
ck.keys.delete(key_id=1)
```

#### 查询操作结果

```python
result = ck.keys.get_consume_result("request-id")
# result: ConsumeResultQuery(source, request_id, key_id, key_alias, ...)
```

#### 导出

```python
text = ck.keys.export()           # 文本格式
data = ck.keys.export_json()      # JSON 列表
```

### `ck.balance_logs` — 余额日志

#### 分页查询

```python
page = ck.balance_logs.list(page=1, page_size=20, alias="my-key")
# page: PageResult[BalanceLog]
# BalanceLog: id, key_id, key_alias, key_suffix, delta, before_amount, after_amount, operator, remark, created_at
```

可选参数 `key_suffix`、`start_time`、`end_time`（格式 `"2006-01-02 15:04:05"`）。

#### 导出

```python
data = ck.balance_logs.export(alias="my-key")
```

## 错误处理

```python
from cloudkey import CloudKeyException

try:
    ck.keys.get(999)
except CloudKeyException as e:
    print(e.http_status)       # HTTP 状态码（如 404）
    print(e.code)              # 业务错误码（如 1001）
    print(e.message)           # 错误消息
    print(e.raw_body)          # 原始 JSON 响应
    print(e.is_transport_error)  # 是否为 HTTP 层错误
```
