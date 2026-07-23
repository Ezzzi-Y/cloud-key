# 消费流程优化 — 执行计划

## 任务总览（串行执行）

| # | 任务 | 涉及文件 | 说明 |
|---|------|---------|------|
| 1 | 配置与依赖 | `go.mod`, `config.yaml`, `internal/config/config.go` | 添加 RabbitMQ 依赖和配置 |
| 2 | MQ 基础设施 | `internal/service/mq_service.go` (新), `internal/service/mq_worker.go` (新) | 连接管理、发布、消费者、事件结构体、DLQ |
| 3 | Redis Lua 脚本 | `internal/service/lua_scripts.go` (新) | 消费扣减 Lua、调额 Lua，嵌入为 Go 常量 |
| 4 | KeyService 重构 | `internal/service/key_service.go` | Redis-First 消费/调额、singleflight 回源、Admin 操作 Redis 同步、移除乐观锁 |
| 5 | Handler 改造 | `internal/handler/key_handler.go`, `internal/handler/service_handler.go` | 消费改用 Lua 返回值、移除同步 UsageLog 写入 |
| 6 | MQ 消费者逻辑 | `internal/service/mq_worker.go` | consume/adjust 事件处理、MySQL 写入、日志记录、重试+DLQ |
| 7 | 主入口接线 | `main.go` | 初始化 MQ、启动 Worker、移除旧 BackfillTopKeys |
| 8 | 构建验证 | — | `go build ./...` |

---

## Task 1：配置与依赖

**目标**：项目能识别 RabbitMQ 配置，依赖已下载。

**变更**：
- `go.mod` — 添加 `github.com/rabbitmq/amqp091-go`
- `config.yaml` — 新增 `rabbitmq` 配置段：
  ```yaml
  rabbitmq:
    url: "amqp://guest:guest@localhost:5672/"
    consume_queue: "cloudkey.consume"
    adjust_queue: "cloudkey.adjust"
    dlq_queue: "cloudkey.dlq"
    exchange: "cloudkey.events"
    prefetch: 10
    max_retries: 3
  ```
- `internal/config/config.go` — 新增 `MQConfig` 结构体，在 `Config` 中增加 `MQ MQConfig` 字段

---

## Task 2：MQ 基础设施

**目标**：完成 RabbitMQ 连接管理、消息发布、Worker 框架、事件结构体定义。

**新增文件**：
- `internal/service/mq_service.go`
- `internal/service/mq_worker.go`

**mq_service.go 职责**：
- `MQService` 结构体，持有 `*amqp091.Connection` 和两个 `*amqp091.Channel`（consume/adjust）
- `NewMQService(url string) (*MQService, error)` — 建立连接，声明 exchange、两个业务队列、DLQ 队列，绑定 routing key
- `PublishConsumeEvent(event ConsumeEvent) error` — 发送到 consume 队列
- `PublishAdjustEvent(event AdjustEvent) error` — 发送到 adjust 队列
- `Close()` — 关闭连接和 channel

**事件结构体**（定义在 mq_service.go）：

```go
type ConsumeEvent struct {
    EventID       string `json:"event_id"`
    KeyID         uint64 `json:"key_id"`
    KeyAlias      string `json:"key_alias"`
    TenantID      uint64 `json:"tenant_id"`
    Amount        int64  `json:"amount"`
    RemainingAfter int64 `json:"remaining_after"`
    StatusAfter   string `json:"status_after"`
    IP            string `json:"ip"`
    UserAgent     string `json:"user_agent"`
    Path          string `json:"path"`
    Timestamp     int64  `json:"timestamp"`
}

type AdjustEvent struct {
    EventID        string `json:"event_id"`
    KeyID          uint64 `json:"key_id"`
    KeyAlias       string `json:"key_alias"`
    TenantID       uint64 `json:"tenant_id"`
    Delta          int64  `json:"delta"`
    RemainingAfter int64  `json:"remaining_after"`
    StatusAfter    string `json:"status_after"`
    Operator       string `json:"operator"`
    Remark         string `json:"remark"`
    Timestamp      int64  `json:"timestamp"`
}
```

**mq_worker.go 框架**：
- `MQWorker` 结构体，持有 `*MQService`、`*gorm.DB`、重试配置
- `Start()` — 启动两个 goroutine，分别消费 consume 队列和 adjust 队列
- `Stop()` — 优雅停止
- 处理逻辑在 Task 6 填充

**队列拓扑**：
- Exchange: `cloudkey.events` (topic)
- Queue `cloudkey.consume` 绑定 `consume.*`
- Queue `cloudkey.adjust` 绑定 `adjust.*`
- Queue `cloudkey.dlq`（死信队列，业务队列 max-retries 后路由到这里）

---

## Task 3：Redis Lua 脚本

**目标**：定义消费扣减和管理员调额的 Lua 脚本，作为 Go 常量嵌入。

**新增文件**：`internal/service/lua_scripts.go`

**脚本一：`consumeLuaScript`**
- 输入：KEYS[1]=`ck:<keyHash>`, ARGV[1]=amount, ARGV[2]=timestamp(ms)
- 逻辑：校验存在 → 校验状态(disabled/expired/exhausted) → 校验过期时间 → 校验余额 → HINCRBY 扣减 → 标记 exhausted
- 输出：JSON `{code, remaining, status, key_id, tenant_id, alias}`
  - code: -1=cache miss, 0=成功, 1004=禁用, 1005=过期, 1006=耗尽, 1007=额度不足

**脚本二：`adjustLuaScript`**
- 输入：KEYS[1]=`ck:<keyHash>`, ARGV[1]=delta
- 逻辑：读 remaining → 校验 newRemaining ≥ 0 → HINCRBY → 如果从非正恢复到正，HSET status=active
- 输出：JSON `{code, before, after, status, key_id, tenant_id, alias}`
  - code: -1=cache miss, 0=成功, 1=余额不足

---

## Task 4：KeyService 重构

**目标**：消费和调额改为 Redis-First，Admin 操作同步 Redis，移除乐观锁。

**变更文件**：`internal/service/key_service.go`

**新增/改造函数**：

| 函数 | 说明 |
|------|------|
| `cacheKey(hash string) string` | 返回 `ck:<hash>` |
| `loadKeyToCache(rawKey string, tenantID uint64) (*model.Key, error)` | singleflight 回源：MySQL 查询 → HSET 到 Redis |
| `consumeViaRedis(rawKey, amount, tenantID) (*ConsumeResult, int, error)` | 执行 Lua → 根据 code 判断 → cache miss 触发回源重试 → 发 MQ ConsumeEvent |
| `consumeViaMySQL(rawKey, amount, tenantID) (*ConsumeResult, int, error)` | 现有 `ConsumeKeyByTenant` 逻辑原样保留，作为 Redis 故障降级路径 |
| `ConsumeKeyByTenant(...)` | 入口：Redis 可用走 `consumeViaRedis`，否则降级走 `consumeViaMySQL` |
| `adjustViaRedis(id, tenantID, req) (*AdjustBalanceResult, error)` | 执行调额 Lua → 发 MQ AdjustEvent → 返回 before/after |
| `AdjustBalance(...)` | 改为调用 `adjustViaRedis`，失败时降级走 MySQL |
| `syncCacheOnCreate(key *model.Key)` | CreateKey 后 HSET 新缓存 |
| `syncCacheOnStatusChange(id, tenantID, status)` | Disable/Enable/Expire 后 HSET status |
| `syncCacheOnDelete(keyHash string)` | DeleteKey 后 DEL 缓存 |

**移除**：
- `ConsumeKeyByTenant` 中的乐观锁重试循环（`maxRetries` + `version` 检查）
- `AdjustBalance` 中的乐观锁重试循环
- 两处 `WHERE version = ?` 条件

**CreateKey / DisableKey / EnableKey / DeleteKey / ExpireKeys** 尾部各加一行 Redis 同步调用。

---

## Task 5：Handler 改造

**目标**：Consume handler 不再同步写 UsageLog，改为通过 MQ 异步。

**变更文件**：
- `internal/handler/key_handler.go`
- `internal/handler/service_handler.go`

**key_handler.go Consume() 改造**：
```
旧流程：ConsumeKeyByTenant → 查 MySQL 拿 keyID/keyAlias → 同步 INSERT usage_logs
新流程：ConsumeKeyByTenant（内部已发 MQ）→ 直接返回 result
```
- 删除 handler 中的 `FindByRawKeyTenant` 调用
- 删除 `usageLogSvc.Record(...)` 调用
- 删除 `h.recordParams` 相关逻辑（MQ 事件已包含 IP/UA/path）
- `TenantKeyHandler` 结构体移除 `usageLogSvc`、`recordParams` 字段（如果其他地方不再使用）

**service_handler.go ServiceConsumeKey()**：同理，ConsumeKeyByTenant 内部已处理 MQ，handler 不需要额外操作。

**key_handler.go AdjustBalance() 改造**：
```
旧流程：AdjustBalance → 查 MySQL 拿 keyAlias → 同步 INSERT balance_logs
新流程：AdjustBalance（内部已发 MQ）→ 直接返回 result
```
- 删除 `balanceLogSvc.Record(...)` 调用
- 删除 `GetKeyDetail` 调用

---

## Task 6：MQ 消费者逻辑

**目标**：填充 Task 2 中 MQWorker 的实际消费处理逻辑。

**变更文件**：`internal/service/mq_worker.go`

**consume 队列消费者**：
```
收到 ConsumeEvent →
  开事务 →
    UPDATE keys SET remaining_amount=?, status=? WHERE id=? →
    INSERT INTO usage_logs (tenant_id, key_id, key_alias, amount, ip, user_agent, request_path, response_status, created_at) →
  Commit → ACK
失败 → Reject+Requeue (retry < max_retries)
超过重试 → 发送到 DLQ + ACK
```

**adjust 队列消费者**：
```
收到 AdjustEvent →
  开事务 →
    UPDATE keys SET remaining_amount=?, status=? WHERE id=? →
    INSERT INTO balance_logs (tenant_id, key_id, key_alias, delta, before_amount, after_amount, operator, remark, created_at) →
  Commit → ACK
失败 → 同上重试+DLQ 逻辑
```

**关键**：`UPDATE` 不带 `WHERE version = ?`，直接用事件中的 `remaining_after` 和 `status_after` 覆盖。MQ per-key 顺序性保证正确性。

---

## Task 7：主入口接线

**目标**：main.go 初始化 MQ 服务，启动 Worker，更新 Handler 构造函数。

**变更文件**：`main.go`

**新增**：
- 初始化 `MQService`
- 初始化 `MQWorker`（传入 MQService + DB + 配置）
- 启动 Worker：`mqWorker.Start()`
- defer 关闭：`mqWorker.Stop()` + `mqService.Close()`

**修改**：
- `NewTenantKeyHandler` 参数变更（移除 `usageLogSvc`、`recordParams`）
- 移除旧的 `BackfillTopKeys()` 调用（Redis 缓存改为 lazy-load）
- 新增 `keySvc.WarmupCache()` 或不预热（依赖 lazy-load，看 Task 4 的决策）

---

## Task 8：构建验证

**目标**：确保编译通过，无语法错误。

```bash
go build ./...
```

如有编译错误，修复后重新构建直到通过。

---

## 关键设计决策备忘

1. **MySQL 不再有乐观锁** — per-key MQ 顺序消费替代了 version 控制
2. **Redis Key 不过期** — 配置中不设 TTL
3. **Cache Miss 回源** — singleflight + 200ms 超时降级
4. **两个独立队列** — consume 和 adjust 分开，跨队列靠 MySQL 顺序写保证最终一致
5. **降级路径** — Redis 不可用时走 consumeViaMySQL（保留完整旧逻辑）
6. **Consumer 写 MySQL** — 直接 SET remaining_amount + status，不检查 version
