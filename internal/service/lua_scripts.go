package service

import "github.com/redis/go-redis/v9"

// consumeLuaScript 消费扣减 Lua 脚本
// KEYS[1] = ck:<keyHash>
// ARGV[1] = amount
// ARGV[2] = timestamp (unix ms)
//
// 返回 JSON:
//
//	code: 0=成功, -1=cache miss, 1001=不存在, 1002=禁用, 1003=耗尽, 1004=额度不足, 1006=过期
//	remaining, status, key_id, tenant_id, alias (成功时)
var consumeLuaScript = redis.NewScript(`
local key = KEYS[1]
local amount = tonumber(ARGV[1])
local now = tonumber(ARGV[2])

if redis.call('EXISTS', key) == 0 then
    return cjson.encode({code = -1})
end

local status = redis.call('HGET', key, 'status')

if status == 'disabled' then
    return cjson.encode({code = 1002})
end

if status == 'exhausted' then
    return cjson.encode({code = 1003})
end

if status == 'expired' then
    return cjson.encode({code = 1006})
end

-- 检查过期时间
local expireAt = tonumber(redis.call('HGET', key, 'expire_at') or '0')
if expireAt > 0 and now > expireAt then
    redis.call('HSET', key, 'status', 'expired')
    return cjson.encode({code = 1006})
end

local remaining = tonumber(redis.call('HGET', key, 'remaining'))

if remaining < amount then
    if remaining <= 0 then
        redis.call('HSET', key, 'status', 'exhausted')
        return cjson.encode({code = 1003})
    end
    return cjson.encode({code = 1004})
end

-- 扣减
local newRemaining = remaining - amount
redis.call('HSET', key, 'remaining', tostring(newRemaining), 'used_at', tostring(now))

local newStatus = 'active'
if newRemaining <= 0 then
    redis.call('HSET', key, 'status', 'exhausted')
    newStatus = 'exhausted'
end

return cjson.encode({
    code = 0,
    remaining = newRemaining,
    status = newStatus,
    key_id = tonumber(redis.call('HGET', key, 'id')),
    tenant_id = tonumber(redis.call('HGET', key, 'tenant_id')),
    alias = redis.call('HGET', key, 'alias')
})
`)

// rateLimitCheckScript Key 限流滑动窗口 Lua 脚本（含租户默认配置查找）
//
// KEYS[1] = rl:key:<keyID>     (ZSET，滑动窗口)
// KEYS[2] = ck:trl:<tenantID>  (Hash，租户默认限流配置)
// ARGV[1] = key 的 rate_limit 值（-1 表示使用租户默认）
// ARGV[2] = key 的 rate_limit_window 值
// ARGV[3] = 当前时间（纳秒时间戳）
// ARGV[4] = 唯一请求标识
//
// 返回 {allowed(1|0), retryAfter(秒)} 或 {-2, 0} 表示租户配置未缓存（需回源）
var rateLimitCheckScript = redis.NewScript(`
local rlKey = KEYS[1]
local tenantKey = KEYS[2]

local rateLimit = tonumber(ARGV[1])
local rateLimitWindow = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local uid = ARGV[4]

-- Key 未配置限流（-1=nil），从租户默认配置读取
if rateLimit == -1 or rateLimitWindow == -1 then
    if redis.call('EXISTS', tenantKey) == 0 then
        return {-2, 0}
    end
    rateLimit = tonumber(redis.call('HGET', tenantKey, 'rate_limit') or '-1')
    rateLimitWindow = tonumber(redis.call('HGET', tenantKey, 'rate_limit_window') or '-1')
end

-- 不限速
if rateLimit <= 0 or rateLimitWindow <= 0 then
    return {1, 0}
end

-- 滑动窗口限流检查
local windowNs = rateLimitWindow * 1000000000
local start = now - windowNs

redis.call('ZREMRANGEBYSCORE', rlKey, 0, start)
local count = redis.call('ZCARD', rlKey)

if count >= rateLimit then
    local oldest = redis.call('ZRANGEBYSCORE', rlKey, start, '+inf', 'WITHSCORES', 'LIMIT', 0, 1)
    if #oldest > 0 then
        local oldestScore = tonumber(oldest[2])
        if oldestScore then
            local retryAfter = math.ceil((oldestScore + windowNs - now) / 1000000000)
            if retryAfter < 1 then retryAfter = 1 end
            return {0, retryAfter}
        end
    end
    return {0, rateLimitWindow}
end

redis.call('ZADD', rlKey, now, uid)
redis.call('EXPIRE', rlKey, rateLimitWindow)
return {1, 0}
`)

// adjustLuaScript 管理员调额 Lua 脚本
// KEYS[1] = ck:<keyHash>
// ARGV[1] = delta (正=增加, 负=减少)
//
// 返回 JSON:
//
//	code: 0=成功, -1=cache miss, 1=余额不足(扣减后为负)
//	before, after, status, key_id, tenant_id, alias (成功时)
var adjustLuaScript = redis.NewScript(`
local key = KEYS[1]
local delta = tonumber(ARGV[1])

if redis.call('EXISTS', key) == 0 then
    return cjson.encode({code = -1})
end

local remaining = tonumber(redis.call('HGET', key, 'remaining'))
local newRemaining = remaining + delta

if newRemaining < 0 then
    return cjson.encode({code = 1, before = remaining, error = "余额不足"})
end

redis.call('HSET', key, 'remaining', tostring(newRemaining))

local curStatus = redis.call('HGET', key, 'status')

-- 增加额度且之前已耗尽/过期，恢复为 active
if delta > 0 and remaining <= 0 and newRemaining > 0 then
    if curStatus == 'exhausted' then
        redis.call('HSET', key, 'status', 'active')
        curStatus = 'active'
    end
end

return cjson.encode({
    code = 0,
    before = remaining,
    after = newRemaining,
    status = curStatus,
    key_id = tonumber(redis.call('HGET', key, 'id')),
    tenant_id = tonumber(redis.call('HGET', key, 'tenant_id')),
    alias = redis.call('HGET', key, 'alias')
})
`)
