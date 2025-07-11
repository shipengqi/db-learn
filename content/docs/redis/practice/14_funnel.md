---
title: 漏斗限流
weight: 14
---

Redis 漏斗限流（Rate Limiter）是一种常用的限流技术，用于控制对某个资源或服务的访问频率，以防止服务被过度使用或遭受滥用。漏斗限流算法通过模拟水流从一个漏斗中流出，来限制数据的传输速率。在 Redis 中，可以通过使用 Redis 的原子操作和一些数据结构来实现漏斗限流，例如使用 Redis 的 `INCR`、`INCRBY`、`EXPIRE`、`SETEX` 等命令，或者使用 Redis 的 Lua 脚本来实现更复杂的逻辑。

漏斗限流的基本原理
- 固定容量和速率：
  - 漏斗有一个固定的容量（capacity），表示在单位时间内可以处理的最大请求数。
  - 漏斗还有一个固定的泄漏速率（rate），表示每秒可以从漏斗中泄漏（即允许处理的）请求数。
- 时间窗口：
  - 通常，漏斗限流是基于一个固定时间窗口（例如1秒）来计算和执行限制的。

## 实现方式

### 使用 Redis 命令实现

**存储请求数**：

使用一个 Redis 键来存储当前时间窗口内的请求数。例如，使用 `INCR` 命令增加请求数。

```bash
INCR key:requests:user_id
```

**设置时间窗口**：

使用 `EXPIRE` 命令设置时间窗口的过期时间，例如每秒重置一次。

```bash
EXPIRE key:requests:user_id 1
```

**检查请求数**：

在执行请求前，先检查当前时间窗口内的请求数是否超过了容量限制。

```bash
GET key:requests:user_id
```

如果请求数大于等于容量，则拒绝请求；否则，继续处理。

### 使用 Lua 脚本实现更复杂的逻辑

对于更复杂的漏斗限流逻辑（如动态调整速率），可以使用 Redis 的 Lua 脚本来实现。Lua 脚本可以在服务器端原子地执行多个命令，避免了多命令执行中的竞态条件。

```lua
--[[
漏斗限流算法核心逻辑：
1. KEYS[1] 限流器唯一标识
2. ARGV[1] rate 每秒允许的请求数
3. ARGV[2] capacity 漏斗总容量
--]]
local key = KEYS[1]  -- 限流器存储 key，例如: user_123_api_limit
local rate = tonumber(ARGV[1])  -- 漏嘴流速（每秒允许请求数）
local capacity = tonumber(ARGV[2])  -- 漏斗总容量（最大突发请求数）

-- 获取当前漏斗状态
local current = tonumber(redis.call('get', key) or "0")  -- 当前剩余容量
local timestamp = tonumber(redis.call('get', key .. ':timestamp') or "0")  -- 上次更新时间戳
local now = tonumber(redis.call('time')[1])  -- 当前 Redis 服务器时间（秒）

-- 时间窗口检测（每秒重置一次漏斗）
if now > timestamp + 1 then
    -- 距离上次请求超过 1 秒，重置漏斗
    current = capacity  -- 恢复漏斗最大容量
    timestamp = now  -- 记录新时间窗口起点
    
    -- 打印调试信息（生产环境需移除）
    -- redis.log(redis.LOG_NOTICE, "漏斗重置 current:"..current.." timestamp:"..timestamp)
end

-- 计算剩余容量（核心算法）
if current > 0 then
    -- 漏斗有剩余容量时：
    local allowed = 1  -- 允许本次请求
    current = current - 1  -- 消耗 1 个容量单位
    
    -- 更新存储（带1秒过期时间）
    redis.call('setex', key, 1, current)  -- 存储剩余容量
    redis.call('setex', key .. ':timestamp', 1, timestamp)  -- 存储时间窗口起点
    
    -- 返回允许请求，并返回剩余容量
    return {1, current}  -- 第一个值1表示允许，第二个值返回剩余容量
else
    -- 漏斗容量已耗尽
    -- 返回拒绝请求，并返回剩余容量
    return {0, 0}
end
```

## 总结

Redis 的漏斗限流通过结合原子操作和过期策略，可以有效地限制对资源的访问速率。无论是使用简单的 Redis 命令还是通过 Lua 脚本实现更复杂的逻辑，都可以根据实际需求选择合适的方法来达到限流的目的。