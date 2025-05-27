---
title: 限流
weight: 9
---

限流算法在分布式领域是一个经常被提起的话题，当系统的处理能力有限时，如何阻止计划外的请求继续对系统施压，这是一个需要重视的问题。

除了控制流量，限流还有一个应用目的是用于控制用户行为，避免垃圾请求。比如在 UGC 社区，用户的发帖、回复、点赞等行为都要严格受控，一般要
严格限定某行为在规定时间内允许的次数，超过了次数那就是非法行为。

## Redis 实现简单限流

系统要限定用户的某个行为在指定的时间里只能允许发生 N 次，如何使用 Redis 的数据结构来实现这个限流的功能？

```py
# 指定用户 user_id 的某个行为 action_key 在特定的时间内 period 只允许发生一定的次数 max_count
def is_action_allowed(user_id, action_key, period, max_count):
    return True
# 调用这个接口 , 一分钟内只允许最多回复 5 个帖子
can_reply = is_action_allowed("laoqian", "reply", 60, 5)
if can_reply:
    do_reply()
else:
    raise ActionThresholdOverflow()
```

### 解决方案

这个限流需求中存在一个滑动时间窗口，想想 `zset` 数据结构的 `score` 值，是不是可以通过 `score` 来圈出这个时间窗口来。
而且我们只需要保留这个时间窗口，窗口之外的数据都可以砍掉。那这个 `zset` 的 `value` 填什么比较合适？它只需要保证唯一性即可，
用 `uuid` 会比较浪费空间，那就改用毫秒时间戳吧。

一个 `zset` 结构记录用户的行为历史，每一个行为都会作为 `zset` 中的一个 `key` 保存下来。同一个用户同一种行为用一个 `zset` 记录。

为节省内存，我们只需要保留时间窗口内的行为记录，同时如果用户是冷用户，滑动时间窗口内的行为是空记录，那么这个 `zset` 就可以从内存
中移除，不再占用空间。

通过统计滑动窗口内的行为数量与阈值 `max_count` 进行比较就可以得出当前的行为是否允许。

```py
import time
import redis

client = redis.StrictRedis()

def is_action_allowed(user_id, action_key, period, max_count):
    key = 'hist:%s:%s' % (user_id, action_key)
    now_ts = int(time.time() * 1000)  # 毫秒时间戳
    with client.pipeline() as pipe:  # client 是 StrictRedis 实例
        # 记录行为
        pipe.zadd(key, now_ts, now_ts)  # value 和 score 都使用毫秒时间戳
        # 移除时间窗口之前的行为记录，剩下的都是时间窗口内的
        pipe.zremrangebyscore(key, 0, now_ts - period * 1000)
        # 获取窗口内的行为数量
        pipe.zcard(key)
        # 设置 zset 过期时间，避免冷用户持续占用内存
        # 过期时间应该等于时间窗口的长度，再多宽限 1s
        pipe.expire(key, period + 1)
        # 批量执行
        _, _, current_count, _ = pipe.execute()
    # 比较数量是否超标
    return current_count <= max_count


for i in range(20):
    print is_action_allowed("laoqian", "reply", 60, 5)
```

### 缺点

因为它要记录时间窗口内所有的行为记录，如果这个量很大，比如限定 60s 内操作不得超过 100w 次这样的参数，它是不适合做这样的限流的，因为
会消耗大量的存储空间。

## 漏斗限流

漏斗的容量是有限的，如果将漏嘴堵住，然后一直往里面灌水，它就会变满，直至再也装不进去。如果将漏嘴放开，水就会往下流，流走一部分之后，就又
可以继续往里面灌水。如果漏嘴流水的速率大于灌水的速率，那么漏斗永远都装不满。如果漏嘴流水速率小于灌水的速率，那么一旦漏斗满了，灌水就需
要暂停并等待漏斗腾空。

所以，**漏斗的剩余空间就代表着当前行为可以持续进行的数量，漏嘴的流水速率代表着系统允许该行为的最大频率**。

### Redis-Cell

Redis 4.0 提供了一个限流 Redis 模块，它叫 `redis-cell`。该模块也使用了漏斗算法，并提供了原子的限流指令。有了这个模块，限流问题
就非常简单了。

该模块只有 1 条指令 `cl.throttle`，它的参数和返回值都略显复杂：

```sh
> cl.throttle laoqian:reply 15 30 60 1
                      ▲     ▲  ▲  ▲  ▲
                      |     |  |  |  └───── need 1 quota (可选参数，表示本次要申请的数量，默认值也是 1)
                      |     |  └──┴─────── 30 operations / 60 seconds 这是漏水速率
                      |     └───────────── 15 capacity 这是漏斗容量
                      └─────────────────── key laoqian
```

上面这个指令的意思是允许「用户老钱回复行为」的频率为每 60s 最多 30 次(漏水速率)，漏斗的初始容量为 15，也就是说一开始可以连续回
复 15 个帖子，然后才开始受漏水速率的影响。

```sh
> cl.throttle laoqian:reply 15 30 60
1) (integer) 0   # 0 表示允许，1 表示拒绝
2) (integer) 15  # 漏斗容量 capacity
3) (integer) 14  # 漏斗剩余空间 left_quota
4) (integer) -1  # 如果拒绝了，需要多长时间后再试(漏斗有空间了，单位秒)
5) (integer) 2   # 多长时间后，漏斗完全空出来(left_quota==capacity，单位秒)
```
