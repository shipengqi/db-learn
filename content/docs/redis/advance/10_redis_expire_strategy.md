---
title: Redis 的过期策略和内存淘汰机制
weight: 10
---

Redis 是怎么删除过期的 key 的？而且 Redis 是单线程的，删除 key 过于频繁会不会造成阻塞？

Redis 有三种清除策略

- **懒惰删除**就是在客户端访问这个 key 的时候，redis 对 key 的过期时间进行检查，如果过期了就立即删除。
- **定时删除**是集中处理，惰性删除是零散处理。
- 当前已用内存超过 maxmemory 限定时，触发主动清理策略

## 定期删除策略

Redis 会将每个**设置了过期时间的 key 放入到一个独立的字典**中，默认每 `100ms` 进行一次过期扫描：

1. 随机抽取 `20` 个 key。
2. 删除这 `20` 个 key 中过期的 key。
3. 如果过期的 key 比例超过 `1/4`，就重复步骤 `1`，继续删除。

之所以**不扫描所有的 key，是因为 Redis 是单线程，全部扫描会导致线程卡死**。

而且为了防止每次扫描过期的 key 比例都超过 `1/4`，导致不停循环卡死线程，Redis 为每次扫描添加了**上限时间**，默认是 `25ms`。

### 如果一个大型的 Redis 实例中所有的 key 在同一时间过期了，会出现怎样的结果

大量的 key 在同一时间过期，那么 Redis 会**持续扫描过期字典 (循环多次)，直到过期字典中过期的 key 变得稀疏**，才会停止 (循环次数明显下降)。这会导致线上读写请求出现**明显的卡顿**现象。导致这种卡顿的另外一种原因是内存管理器需要频繁回收内存页，这也会产生一定的 CPU 消耗。

而且，如果客户端将请求超时时间设置的比较短，比如 10ms，但是请求以为过期扫描导致至少等待 25ms 后才会进行处理，那么就会出现大量的请求因为超时而关闭，业务端就会出现很多异常。这时你还**无法从 Redis 的 `slowlog` 中看到慢查询记录，因为慢查询指的是逻辑处理过程慢，不包含等待时间**。

所以要避免大批量的 key 同时过期，可以给过期时间设置一个随机范围，分散过期处理的压力。

## 内存淘汰机制

当 Redis 内存超出物理内存限制时，内存的数据会开始和磁盘产生频繁的交换 (swap)。交换会让 Redis 的性能急剧下降，对于 Redis 来说，这样龟速的存取效率基本上等于不可用。

Redis 为了限制最大使用内存，提供了配置参数 `maxmemory`，可以在 `redis.conf` 中配置。当内存超出 `maxmemory`，Redis 提供了几种策略（maxmemory-policy）让用户选择：

- `noeviction`：当内存超出 `maxmemory`，写入请求会报错，但是删除和读请求可以继续。（这个可是默认的策略）。
- `allkeys-lru`：当内存超出 `maxmemory`，在所有的 key 中，移除最少使用的 key。
- `allkeys-random`：当内存超出 `maxmemory`，在所有的 key 中，随机移除某个 key。（应该没人用吧）
- `volatile-lru`：当内存超出 `maxmemory`，在设置了过期时间 key 的字典中，移除最少使用的 key。
- `volatile-random`：当内存超出 `maxmemory`，在设置了过期时间 key 的字典中，随机移除某个key。
- `volatile-ttl`：当内存超出 `maxmemory`，在设置了过期时间 key 的字典中，优先移除 `ttl` 小的。

### volatile 和 allkeys 的区别

- `volatile-xxx` 策略只会针对带过期时间的 key 进行淘汰。
- `allkeys-xxx` 策略会对所有的 key 进行淘汰。

如果只是拿 Redis 做缓存，那应该使用 `allkeys-xxx`，客户端写缓存时不必携带过期时间。如果还想同时使用 Redis 的持久化功能，那就使用 `volatile-xxx` 策略，这样可以保留没有设置过期时间的 key，它们是永久的 key 不会被 LRU 算法淘汰。


### LFU

Redis 4.0 里引入了一个新的淘汰策略 —— LFU（Least Frequently Used） 模式。

LFU 表示按最近的访问频率进行淘汰，它比 LRU 更加精准地表示了一个 key 被访问的热度。

如果一个 key 长时间不被访问，只是刚刚偶然被用户访问了一下，那么在使用 LRU 算法下它是不容易被淘汰的，因为 LRU 算法认为当前这个 key 是很热的。而 LFU 是需要追踪最近一段时间的访问频率，如果某个 key 只是偶然被访问一次是不足以变得很热的，它需要在近期一段时间内被访问很多次才有机会被认为很热。

#### 启用 LFU

Redis 4.0 给淘汰策略配置参数 `maxmemory-policy` 增加了 2 个选项，

- `volatile-lfu`：对带过期时间的 key 执行 lfu 淘汰算法
- `allkeys-lfu`：对所有的 key 执行 lfu 淘汰算法

使用 `object freq` 指令获取对象的 lfu 计数值：

```sh
> config set maxmemory-policy allkeys-lfu
OK
> set codehole yeahyeahyeah
OK
// 获取计数值，初始化为 LFU_INIT_VAL=5
> object freq codehole
(integer) 5
// 访问一次
> get codehole
"yeahyeahyeah"
// 计数值增加了
> object freq codehole
(integer) 6
```

## 懒惰删除

### Redis 为什么要懒惰删除(lazy free)

删除指令 `del` 会直接释放对象的内存，大部分情况下，这个指令非常快，没有明显延迟。不过如果删除的 key 是一个非常大的对象，比如一个包含了千万元素的 `hash`，又或者在使用 `FLUSHDB` 和 `FLUSHALL` 删除包含大量键的数据库时，那么删除操作就会导致线程卡顿。

redis 4.0 引入了 `lazyfree` 的机制，它可以将删除键或数据库的操作放在后台线程里执行， 从而尽可能地避免服务器阻塞。

### unlink

`unlink` 指令，它能对删除操作进行懒处理，丢给后台线程来异步回收内存。

```sh
> unlink key
OK
```

### flush

`flushdb` 和 `flushall` 指令，用来清空数据库，这也是极其缓慢的操作。Redis 4.0 同样给这两个指令也带来了异步化，在指令后面增加 `async` 参数就可以扔给后台线程慢慢处理。

```sh
> flushall async
OK
```
