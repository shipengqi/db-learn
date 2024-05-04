---
title: Redis 入门
---

[Redis 中文官网的介绍](http://www.redis.cn/)：

Redis（Remote Dictionary Service）是目前互联网技术领域使用最为广泛的存储中间件，它是一个开源（BSD 许可）的，内存中的数据结构存储系
统，它可以用作数据库、缓存和消息中间件。它支持多种类型的数据结构，如 字符串（strings）， 散列（hashes）， 列表（lists），
集合（sets）， 有序集合（sorted sets） 与范围查询，bitmaps， hyperloglogs 和 地理空间（geospatial） 索引半径查询。 Redis 内置了
 复制（replication），LUA 脚本（Lua scripting），LRU 驱动事件（LRU eviction），事务（transactions） 和不同级别的 磁
 盘持久化（persistence），并通过 Redis 哨兵（Sentinel）和自动 分区（Cluster）提供高可用性（high availability）。

## 数据类型

Redis 一共支持 5 种数据类型：

- [字符串(Strings)](03_redis-string.md)
- [哈希(Hashs)](04_redis-hash.md)
- [列表(Lists)](07_redis-list.md)
- [集合(Sets)](05_redis-set.md)
- [有序集合(SortedSets)](06_redis-sortedset.md)

### String（字符串）

`String` 类型是最常用，也是最简单的的一种类型，`string` 类型是二进制安全的。也就是说 `string` 可以包含任何数据。比如 `jpg 图片`
或者 `序列化的对象` 。一个键**最大能存储 `512MB`**。

``` bash
redis> set testkey hello
OK
redis> get testkey
"hello"
```

### Hash（哈希）

Redis 对 `JSON` 数据的支持不是很友好。通常把 `JSON` 转成 `String` 存储到 Redis 中，但现在的 `JSON` 数据都是连环嵌套的，每次更新
时都要先获取整个 `JSON`，然后更改其中一个字段再放上去。这种使用方式，如果在海量的请求下，`JSON` 字符串比较复杂，会导致在频繁更新数
据使网络 I/O 跑满，甚至导致系统超时、崩溃。所以 Redis 官方推荐采用哈希保存对象。

- `HSET` 设置哈希类型的值
- `HGET` 获取单个哈希属性值
- `HGETALL` 获取所有哈希属性和值

``` bash
redis> HSET  xiaoming age 18
(integer) 1
redis> HSET xiaoming phone 15676666666
(integer) 1
redis> HMSET xiaoqiang age 18 phone 13816666666
OK
redis> HGET xiaoming age
"18"
redis> HGETALL xiaoming
1)"age"
2)"18"
3)"phone"
4)"15676666666"
redis> HGETALL xiaoqiang
1)"age"
2)"18"
3)"phone"
4)"13816666666"
```

### List（列表）

Redis 列表是简单的字符串列表，并根据插入顺序进行排序。一个 Redis 列表中最多可存储 `232-1` (40 亿)个元素。

- `LPUSH` 向列表的开头插入新元素，
- `RPUSH` 向列表的结尾插入新元素。
- `LPOP` 返回并移除列表头部的元素
- `RPOP` 返回并移除列表尾部的元素。
- `LSET` 对列表中指定索引位的元素进行操作。**`LSET` 不允许对不存在的列表进行操作**。
- `LINDEX` 获取列表指定索引位的元素

``` bash
redis> LPUSH testlist one
(integer) 1
redis> RPUSH testlist two
(integer) 2
redis> LSET testlist 0 1
OK
redis> LINDEX testlist 0
"1"
redis> LPOP testlist
"1"
redis> RPOP testlist
"two"
redis> lindex testlist 0
(nil)
```

### Set（集合）

Redis 的 `Set` 是 `string` 类型的无序集合。**集合中不允许重复成员的存在**。一个 Redis 集合中最多可包含 `232-1`(40 亿)个元素。

- `SADD` 向集合中插入值
- `SMEMBERS` 获取集合中的元素
- `SPOP` 随机获取并删除一个值

``` bash
redis> sadd class xiaoming xiaoqiang xiaogang
(integer) 3
redis> smembers class
1) "xiaoming"
2) "xiaoqiang"
3) "xiaogang"
redis> spop class
"xiaoqiang"
redis> smembers class
1) "xiaoming"
2) "xiaogang"
```

集合间的操作:

- `SINTER` 查询集合间的交集
- `SUNION`获取集间的并集
- `SDIFF` 获取集合间的差集
- `SMOVE` 把集合中的元素从一个集合移到另一个集中

``` bash
redis> sadd class1 xiaoming xiaoqiang xiaogang
(integer) 1
redis> sadd class2 xiaoming xiaoli
(integer) 1
redis> sinter class1 class2
1) "xiaoming"
redis> sunion class1 class2
1) "xiaoming"
2) "xiaoqiang"
3) "xiaogang"
4) "xiaoming"
5) "xiaoli"
redis> sdiff class1 class2
1) "xiaoqiang"
2) "xiaogang"
redis> smove class1 class2 xiaoqiang
(integer) 0
redis> smembers class2
1) "xiaoming"
2) "xiaoqiang"
3) "xiaoli"
```

### zset (sorted set, 有序集合)

Redis 的 `zset` 和 `set` 一样也是 `string` 类型元素的集合，且**不允许重复的成员**。不同的是每个元素都会关联一个 `double` 类型的
分数。Redis 正是通过分数来为集合中的成员进行从小到大的排序。`zset` 的成员是唯一的,但分数(`score`)却可以重复。

- `ZADD` 添加元素到集合，元素在集合中存在则更新对应 `score`
- `ZRANGE` 集合指定范围内的元素
- `ZRANK` 获取指定成员的排名
- `ZSCORE` 返回元素的权重(`score`)值

``` bash
ZADD key score member
```

``` bash
redis> ZADD class 1 xiaoming
(integer) 1
redis> ZADD class 2 xiaoqiang 3 xiaogang
(integer) 2
redis> ZRANGE class 1 2
1) "xiaoming"
2) "xiaoqiang"
redis> ZRANK class xiaogang
(integer) 2
redis> ZSCORE class xiaogang
"3"
```

### 容器型数据结构

`list/set/hash/zset` 这四种都属于容器型数据结构，他们有两条通用规则：

- 如果容器不存在，那就创建一个，再进行操作。比如 `RPUSH`，如果列表不存在，Redis 就会自动创建一个，然后再执行 `RPUSH`。
- 如果容器里元素没有了，那么立即删除 key，释放内存。比如 `LPOP` 操作到最后一个元素，列表 key 就会自动删除。
