---
title: 开发规范
wight: 4
---


## 键值设计

### key 名设计

1. 可读性和可管理性

以业务名(或数据库名)为前缀(防止 key 冲突)，用冒号分隔，比如 `业务名:表名:id`：`trade:order:1`。

2. 简洁性

保证语义的前提下，控制 key 的长度，当 key 较多时，内存占用也不容忽视，例如：

```bash
user:{uid}:friends:messages:{mid} 简化为 u:{uid}:fr:m:{mid}。
```

3. 不要包含特殊字符

反例：包含空格、换行、单双引号以及其他转义字符。

### value 设计

#### 拒绝 bigkey

在 Redis 中，一个字符串最大 512MB，一个二级数据结构（例如 hash、list、set、zset）可以存储大约 40 亿个 (2^32-1) 个元素，但实际中如果下面两种情况，我就会认为它是 bigkey。

1. 字符串类型：它的 big 体现在单个 value 值很大，一般认为超过 10KB 就是 bigkey。
2. 非字符串类型：哈希、列表、集合、有序集合，它们的big体现在元素个数太多。

一般来说，string 类型控制在 10KB 以内，hash、list、set、zset 元素个数不要超过5000。

{{< callout type="info" >}}
非字符串的 bigkey，不要使用 `del` 删除，使用 `hscan`、`sscan`、`zscan` 方式渐进式删除，同时要注意**防止 bigkey 过期时间自动删除问题**（例如一个 200 万的 zset 设置 1 小时过期，会触发 `del` 操作，造成阻塞）。
{{< /callout >}}

##### bigkey 的危害

1. **导致 Redis 阻塞**。
2. **网络拥塞**。bigkey 也就意味着每次获取要产生的网络流量较大，假设一个 bigkey 为 1MB，客户端每秒访问量为 1000，那么每秒产生 1000MB 的流量，对于普通的千兆网卡(按照字节算是 `128MB/s`)的服务器来说简直是灭顶之灾，而且一般服务器会采用单机多实例的方式来部署，也就是说一个 bigkey 可能会对其他实例也造成影响，其后果不堪设想。
3. **过期删除**。有个 bigkey，它安分守己（只执行简单的命令，例如 `hget`、`lpop`、`zscore` 等），但它设置了过期时间，当它过期后，会被删除，如果没有使用 Redis 4.0 的过期异步删除(`lazyfree-lazy-expire yes`)，就会存在阻塞 Redis 的可能性。

##### bigkey 的产生

一般来说，bigkey 的产生都是由于程序设计不当，或者对于数据规模预料不清楚造成的，来看几个例子：
1. 社交类：粉丝列表，如果某些明星或者大 v 不精心设计下，必是 bigkey。
2. 统计类：例如按天存储某项功能或者网站的用户集合，除非没几个人用，否则必是 bigkey。
3. 缓存类：将数据从数据库 `load` 出来序列化放到 Redis 里，这个方式非常常用，但有两个地方需要注意，第一，是不是有必要把所有字段都缓存；第二，有没有相关关联的数据，有的同学为了图方便把相关数据都存一个 key 下，产生 bigkey。

##### 如何优化 bigkey

1. 拆

big list： list1、list2、...listN
big hash：可以将数据分段存储，比如一个大的 key，假设存了 1 百万的用户数据，可以拆分成 200 个 key，每个 key 下面存放 5000 个用户数据

2. 如果 bigkey 不可避免，也要思考一下**要不要每次把所有元素都取出来** (例如有时候仅仅需要 `hmget`，而不是 `hgetall`)，**删除也是一样**，尽量使用优雅的方式来处理。

#### 选择适合的数据类型

例如：实体类型(要合理控制和使用数据结构内存编码优化配置,例如 ziplist，但也要注意节省内存和性能之间的平衡)，设置合理的过期时间。

反例：

```bash
set user:1:name tom
set user:1:age 19
set user:1:favor football
```

正例:

```bash
hmset user:1 name tom age 19 favor football
```

#### 控制 key 的生命周期

使用 `expire` 设置过期时间(条件允许可以打散过期时间，防止集中过期)。

## 命令使用

1. `O(N)` 命令关注 N 的数量

例如 `hgetall`、`lrange`、`smembers`、`zrange`、`sinter` 等并非不能使用，但是需要明确 N 的值。有遍历的需求可以使用 `hscan`、`sscan`、`zscan` 代替。

2. 禁用命令

禁止线上使用 `keys`、`flushall`、`flushdb` 等，通过 Redis 的 **`rename` 机制禁掉命令**，或者使用 `scan` 的方式渐进式处理。

3. 合理使用 `select`

Redis 的多数据库较弱，使用数字进行区分，很多客户端支持较差，同时**多业务用多数据库实际还是单线程处理**，会有干扰。

4. 使用批量操作提高效率

原生命令：例如 `mget`、`mset`。
非原生命令：可以使用 `pipeline` 提高效率。

但要注意控制一次批量操作的元素个数(例如 500 以内，实际也和元素字节数有关)。
注意两者不同：


- 原生命令是原子操作，pipeline 是非原子操作。
- pipeline 可以打包不同的命令，原生命令做不到
- pipeline 需要客户端和服务端同时支持。

5. Redis 事务功能较弱，不建议过多使用，可以用 lua 替代。

## 客户端使用

1. 避免多个应用使用一个 Redis 实例。不相干的业务拆分，公共数据做服务化。
2. 使用带有连接池的数据库，可以有效控制连接，同时提高效率。
3. 高并发下建议客户端添加熔断功能(例如 sentinel、hystrix)。
4. 设置合理的密码，如有必要可以使用 SSL 加密访问。

### 连接池

使用带有连接池，可以有效控制连接，同时提高效率，标准使用方式：

```java
JedisPoolConfig jedisPoolConfig = new JedisPoolConfig();
jedisPoolConfig.setMaxTotal(5);
jedisPoolConfig.setMaxIdle(2);
jedisPoolConfig.setTestOnBorrow(true);

JedisPool jedisPool = new JedisPool(jedisPoolConfig, "192.168.0.60", 6379, 3000, null);

Jedis jedis = null;
try {
    jedis = jedisPool.getResource();
    //具体的命令
    jedis.executeCommand()
} catch (Exception e) {
    logger.error("op key {} error: " + e.getMessage(), key, e);
} finally {
    //注意这里不是关闭连接，在JedisPool模式下，Jedis会被归还给资源池。
    if (jedis != null) 
        jedis.close();
}
```

连接池参数含义：

- `maxTotal`：最大连接数，早期的版本叫 maxActive。设置该值，需要考虑的因素
  - 业务期望的 QPS
  - 客户端执行命令时间
  - Redis 资源：例如 `nodes(例如应用个数) * maxTotal` 是不能超过 Redis 的最大连接数 `maxclients`。
  - 资源开销：例如虽然希望控制空闲连接(连接池此刻可马上使用的连接)，但是不希望因为连接池的频繁释放创建连接造成不必靠开销。
  - 假设: 一次命令时间（borrow|return resource + Jedis 执行命令(含网络) ）的平均耗时约为 1ms，一个连接的 QPS 大约是 1000。业务期望的 QPS 是 50000。那么理论上需要的资源池大小是 `50000 / 1000 = 50` 个。但事实上这是个理论值，还要考虑到要比理论值预留一些资源，通常来讲 `maxTotal` 可以比理论值大一些。但这个值不是越大越好，一方面连接太多占用客户端和服务端资源，另一方面对于 Redis 这种高 QPS 的服务器，一个大命令的阻塞即使设置再大资源池仍然会无济于事。
- `maxIdle` 和 `minIdle`：`maxIdle` 实际上才是业务需要的最大连接数，`maxTotal` 是为了给出**余量**，所以 `maxIdle` 不要设置过小，否则会有 `new Jedis` (新连接)开销。**连接池的最佳性能是 `maxTotal = maxIdle`**。这样就避免连接池伸缩带来的性能干扰。但是如果并发量不大或者 `maxTotal` 设置过高，会导致不必要的连接资源浪费。一般推荐 `maxIdle` 可以设置为按业务期望 QPS 计算出来的理论连接数，`maxTotal` 可以再放大一倍。
- `minIdle`：`minIdle`（最小空闲连接数），与其说是最小空闲连接数，不如说是"至少需要保持的空闲连接数"，在使用连接的过程中，如果连接数**超过了 `minIdle`，那么继续建立连接**，如果超过了 `maxIdle`，当**超过的连接执行完业务后会慢慢被移出连接池释放掉**。
- `testOnBorrow`：在borrow一个 jedis 实例时，是否提前进行 validate 操作；如果为 true，则得到的 jedis 实例均是可用的；   
- `testOnReturn`：在 return 一个 jedis 实例时，是否提前进行 validate 操作；如果为 true，则返回的 jedis 实例均是可用的。

#### 连接池预热

Redis 初始化后是没有连接的，当需要使用连接时，才会创建连接。

连接池预热在应用启动时，就创建好一定数量的连接，避免在使用时创建连接。

```java
List<Jedis> minIdleJedisList = new ArrayList<Jedis>(jedisPoolConfig.getMinIdle());

for (int i = 0; i < jedisPoolConfig.getMinIdle(); i++) {
    Jedis jedis = null;
    try {
        jedis = pool.getResource();
        minIdleJedisList.add(jedis);
        jedis.ping();
    } catch (Exception e) {
        logger.error(e.getMessage(), e);
    } finally {
        // 注意，这里不能马上close将连接还回连接池，否则最后连接池里只会建立 1 个连接。。
        // jedis.close();
    }
}
// 统一将预热的连接还回连接池
for (int i = 0; i < jedisPoolConfig.getMinIdle(); i++) {
    Jedis jedis = null;
    try {
        jedis = minIdleJedisList.get(i);
        //将连接归还回连接池
        jedis.close();
    } catch (Exception e) {
        logger.error(e.getMessage(), e);
    } finally {
    }
}
```

## Redis 对于过期键有三种清除策略

1. 被动删除：当读/写一个已经过期的 key 时，会触发惰性删除策略，直接删除掉这个过期 key。
2. 主动删除：由于惰性删除策略无法保证冷数据被及时删掉，所以 Redis 会定期(默认每 100ms)主动淘汰一批已过期的 key，这里的一批只是部分过期 key，所以可能会出现部分 key 已经过期但还没有被清理掉的情况，导致内存并没有被释放。
3. 当前已用内存超过 maxmemory 限定时，触发主动清理策略

### 主动清理策略

主动清理策略在 Redis 4.0 之前一共实现了 6 种内存淘汰策略，在 4.0 之后，又增加了 2 种策略，总共 8 种：

#### 针对设置了过期时间的 key 做处理

- `volatile-ttl`：在筛选时，会针对设置了过期时间的键值对，根据过期时间的先后进行删除，越早过期的越先被删除。
- `volatile-random`：就像它的名称一样，在设置了过期时间的键值对中，进行随机删除。
- `volatile-lru`：会使用 LRU 算法筛选设置了过期时间的键值对删除。
- `volatile-lfu`：会使用 LFU 算法筛选设置了过期时间的键值对删除。

#### 针对所有的 key 做处理

- `allkeys-random`：从所有键值对中随机选择并删除数据。
- `allkeys-lru`：使用 LRU 算法在所有数据中进行筛选删除。
- `allkeys-lfu`：使用 LFU 算法在所有数据中进行筛选删除。

#### 不处理

- `noeviction`：不会剔除任何数据，拒绝所有写入操作并返回客户端错误信息"(error) OOM command not allowed when used memory"，此时 Redis 只响应读操作。

#### LRU 算法 和 LFU 算法

LRU（Least Recently Used），最近最少使用，淘汰很久没被访问过的数据，**以最近一次访问时间作为参考**。例如有两个 key，一个是 `key1`，最近一次访问时间是 10 分钟前，一个是 `key2`，最近一次访问时间是 5 分钟前，那么 `key2` 会被淘汰。

LFU（Least Frequently Used），最不经常使用，淘汰最近一段时间被访问次数最少的数据，**以次数作为参考**。例如有两个 key，一个是 `key1`，最近 10 分钟内被访问了 10 次，一个是 `key2`，最近 10 分钟内被访问了 5 次，那么 `key2` 会被淘汰。

当存在热点数据时，使用 LFU 可能更好。因为 LRU 时，如果有偶发性的、周期性的批量操作会使一些冷数据被访问。

大部分情况下，使用 LRU 算法就可以了。

#### 如何选择

根据自身业务类型，配置好 `maxmemory-policy`(默认是 `noeviction`)，推荐使用`volatile-lru`。如果不设置最大内存，当 Redis 内存超出物理内存限制时，内存的数据会开始和磁盘产生频繁的交换 (swap)，会让 Redis 的性能急剧下降。

当 Redis 运行在主从模式时，只有主结点才会执行过期删除策略，然后把删除操作 `del key` 同步到从结点删除数据。