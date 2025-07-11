---
title: 开发规范
weight: 11
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
2. 非字符串类型：哈希、列表、集合、有序集合，它们的 big 体现在元素个数太多。

一般来说，string 类型控制在 10KB 以内，hash、list、set、zset 元素个数不要超过5000。

{{< callout type="info" >}}
非字符串的 bigkey，不要使用 `del` 删除，使用 `hscan`、`sscan`、`zscan` 方式渐进式删除，同时要注意**防止 bigkey 过期时间自动删除问题**（例如一个 200 万的 zset 设置 1 小时过期，会触发 `del` 操作，造成阻塞）。
{{< /callout >}}

##### bigkey 的危害

1. **导致 Redis 阻塞**。
2. **网络拥塞**。bigkey 也就意味着每次获取要产生的网络流量较大，假设一个 bigkey 为 1MB，客户端每秒访问量为 1000，那么每秒产生 1000MB 的流量，对于普通的千兆网卡(按照字节算是 `128MB/s`)的服务器来说简直是灭顶之灾，而且一般服务器会采用单机多实例的方式来部署，也就是说一个 bigkey 可能会对其他实例也造成影响，其后果不堪设想。
3. **过期删除**。有个 bigkey，它安分守己（只执行简单的命令，例如 `hget`、`lpop`、`zscore` 等），但它设置了过期时间，当它过期后，会被删除，如果没有使用 Redis 4.0 的**过期异步删除(`lazyfree-lazy-expire yes`)**，就会存在阻塞 Redis 的可能性。

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

