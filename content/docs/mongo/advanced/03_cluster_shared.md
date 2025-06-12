---
title: 集群架构 - 分片集群
weight: 3
---

## 分片简介

**分片（shard）是指在将数据进行水平切分之后，将其存储到多个不同的服务器节点上的一种扩展方式**。分片在概念上非常类似于应用开发中的“水平分表”。不同的点在于，MongoDB 本身就自带了分片管理的能力，对于开发者来说可以做到开箱即用。

### 为什么要使用分片？

**MongoDB复制集实现了数据的多副本复制及高可用，但是一个复制集能承载的容量和负载是有限的**。在你遇到下面的场景时，就需要考虑使用分片了：

- 存储容量需求超出单机的磁盘容量。
- 活跃的数据集**超出单机内存容量，导致很多请求都要从磁盘读取数据**，影响性能。
- 写 IOPS 超出单个 MongoDB 节点的写服务能力。

垂直扩容（Scale Up） VS 水平扩容（Scale Out）： 

- 垂直扩容 ： 用更好的服务器，提高 CPU 处理核数、内存数、带宽等 
- 水平扩容 ： 将任务分配到多台计算机上

### 分片集群架构

MongoDB 分片集群（Sharded Cluster）是对数据进行水平扩展的一种方式。**MongoDB 使用分片集群来支持大数据集和高吞吐量的业务场景**。在分片模式下，存储不同的切片数据的节点被称为**分片节点**，一个分片集群内包含了多个分片节点。当然，除了分片节点，集群中还需要一些配置节点、路由节点，以保证分片机制的正常运作。

![mongodb-shards-arch](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-shards-arch.png)


比 Redis 的集群架构更加复杂，多了路由，配置节点。

一般生产环境如果有三个分片，每个分片包含三个节点，那么整个集群就有 9 个节点。每个节点的内存需要大一点（32G/64G），因为复制集中每个节点都可能成为主节点。

配置节点的数量一般为 3 个，组成一个复制集，内存不需要太大（4G/8G），不需要存储数据。

Mongos 至少需要 2 个节点，避免单点故障，内存不需要太大（4G/8G），不需要存储数据。

### 核心概念

- **数据分片**：分片用于存储真正的数据，并提供最终的数据读写访问。分片仅仅是一个逻辑的概念，它**可以是一个单独的 mongod 实例，也可以是一个复制集**。图中的 Shard1、Shard2 都是一个复制集分片。在生产环境中也**一般会使用复制集的方式，这是为了防止数据节点出现单点故障**。
- **配置服务器**（Config Server）：配置服务器包含多个节点，并组成一个复制集结构，对应于图中的 ConfigReplSet。**配置复制集中保存了整个分片集群中的元数据，其中包含各个集合的分片策略，以及分片的路由表等**。
- **查询路由**（mongos）：**mongos 是分片集群的访问入口，其本身并不持久化数据**。mongos 启动后，会从配置服务器中加载元数据。之后 mongos 开始提供访问服务，并将用户的请求正确路由到对应的分片。在分片集群中可以部署多个mongos 以分担客户端请求的压力。

## 搭建分片集群

- 使用 [mtools](https://github.com/rueckstiess/mtools) 搭建分片集群，mtools 可以快速搭建一个简单的分片集群，可用于测试。
- [标准方式搭建分片集群](https://note.youdao.com/ynoteshare/index.html?id=26c9b7e8007efd46a2eed3d28dc06ea2&type=note&_time=1749611678953)，需要手动搭建分片集群，比较复杂。

### 使用分片集群

为了使集合支持分片，需要先开启 database 的分片功能：

```bash
use shop
sh.enableSharding("shop")
```

执行 `shardCollection` 命令，对集合执行分片初始化：

```javascript
sh.shardCollection("shop.product",{productId:"hashed"},false,{numInitialChunks:4})
```

**`shop.product` 集合将 `productId` 作为分片键，并采用了哈希分片策略**，除此以外，`numInitialChunks：4` 表示将初始化 4 个 chunk。**`numInitialChunks` 必须和哈希分片策略配合使用**。而且，这个选项只能用于空的集合，如果已经存在数据则会返回错误。

#### 向分片集合写入数据

向 `shop.product` 集合写入一批数据：

```javascript
db=db.getSiblingDB("shop");
var count=0;
for(var i=0;i<1000;i++){
    var p=[];
    for(var j=0;j<100;j++){
        p.push({
            "productId":"P-"+i+"-"+j,
            name:"羊毛衫",
            tags:[
                {tagKey:"size",tagValue:["L","XL","XXL"]},
                {tagKey:"color",tagValue:["蓝色","杏色"]},
                {tagKey:"style",tagValue:"韩风"}
            ]
        });
    }
    count+=p.length;
    db.product.insertMany(p);
    print("insert ",count)
}
```

#### 查询数据的分布

```javascript
db.product.getShardDistribution()
```

输出结果示例：

```bash
Shard shard01 at shard01/localhost:27053,localhost:27054,localhost:27055
  data : 12.54MiB docs : 50065 chunks: 2 # 该分片有 2 个 chunk
  estimated data per chunk : 6.27Mib
  estimated docs per chunk : 25032
Shard shard02 at shard02/localhost:27056,localhost:27057,localhost:27058
  data : 12.51MiB docs : 49935 chunks: 2 # 该分片有 2 个 chunk
  estimated data per chunk : 6.25Mib
  estimated docs per chunk : 24967
Totals
  data : 25.06MiB docs : 100000 chunks : 4
  Shard shard01 contains 50.06% of data, 50.06% of docs in cluster, avg obj sieze on shard : 262B
  Shard shard02 contains 49.93% of data, 49.93% of docs in cluster, avg obj sieze on shard : 262B
```

## 分片策略

通过分片功能，可以将一个非常大的集合分散存储到不同的分片上，如图：

![mongodb-shards-strategy](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-shards-strategy.png)


假设这个集合大小是 1TB，那么拆分到 4 个分片上之后，每个分片存储 256GB 的数据。这个当然是最理想化的场景，**实际上很难做到如此绝对的平衡**。一个集合在拆分后如何存储、读写，与该集合的分片策略设定是息息相关的。每个分片不可能直接存储一个 256GB 的数据，这些数据会分成多个 chunk，每个 chunk 存储一部分数据，保证每一个分片一共是 256GB 的数据。

也就是说**一个集合会分成多个 chunk，分布在多个分片上，每个 chunk 存储一部分数据**。

### 什么是 chunk

chunk 的意思是数据块，**一个 chunk 代表了集合中的“一段数据”**，例如，用户集合（`db.users`）在切分成多个 chunk 之后如图所示：

![mongodb-shards-chunk](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-shards-chunk.png)


**chunk 所描述的是范围区间**，例如，`db.users` 使用了 `userId` 作为分片键，那么 chunk 就是 `userId` 的各个值（或哈希值）的连续区间。**集群在操作分片集合时，会根据分片键找到对应的 chunk，并向该 chunk 所在的分片发起操作请求**，而 chunk 的分布在一定程度上会影响数据的读写路径，这由以下两点决定：

- chunk 的切分方式，决定如何找到数据所在的 chunk
- chunk 的分布状态，决定如何找到 chunk 所在的分片

### 分片算法

**chunk 切分是根据分片策略进行实施的，分片策略的内容包括分片键和分片算法**。当前，MongoDB支持两种分片算法。

#### 范围分片

假设集合根据 x 字段来分片，x 的完整取值范围为 `[minKey, maxKey]`（x 为整数，这里的 minKey、maxKey 为整型的最小值和最大值），其将整个取值范围划分为多个 chunk，例如：

- chunk1 包含 x 的取值在 `[minKey，-75)` 的所有文档。
- chunk2 包含 x 取值在 `[-75，25)` 之间的所有文档，依此类推。

![mongodb-shards-range](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-shards-range.png)

**范围分片能很好的满足范围查询的需求**，比如要查询 x 在 `[-30, 10]` 之间的所有文档，这是 mongos 直接将请求定位到 chunk2 所在的分片服务器上，就能查询出所有符合条件的文档。**范围分片的缺点在于，如果 Shard Key 有明显的递增或递减的趋势，则新插入的文档会分布到同一个 chunk，此时写压力会集中到一个节点，从而导致单点的性能瓶颈**。

一些常见的导致递增 Key 的场景：

- 时间值
- `ObjectId`
- UUID，包含系统时间、时钟序列。
- 自增整数

#### 哈希分片

**哈希分片会实现根据分片键 Shard Key 计算出一个新的哈希值（64 位整数），再根据哈希值按照范围分片的策略进行 chunk 切分**。适用于日志，物联网等高并发场景。

![mongodb-shards-hash](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-shards-hash.png)


哈希分片与范围分片是互补的，**由于哈希算法保证了随机性，所以文档可以更加离散的分布到不同的 chunk 上，这避免了范围分片的集中写问题**。然而，**在执行一些范围查询时，哈希分片的效率不如范围分片**。因为所有的范围查询都必然导致对多有的 chunk 进行检索，如果集群有 10 个分片，那么 mongos 将需要对 10 个分片进行查询。哈希分片与范围分片的另一个区别是，**哈希分片只能选择单个字段，而范围分片允许采用组合式的多个字段作为分片键**。

哈希分片仅支持单个字段的哈希分片：

```javascript
{x : "hashed" } 
{x : 1 , y : "hashed"} // 4.4 new
```

4.4 以后的版本，可以将单个字段的哈希分片和一个到多个的范围分片键字段来进行组合，比如指定 `x:1,y: "hashed"` 中 `y` 是哈希的方式。

### 分片标签

**MongoDB 允许通过为分片添加标签（tag）的方式来控制数据分发**。一个标签可以关联到多个分片区间（TagRange）。均衡器会优先考虑 chunk 是否正处于某个分片区间上（被完全包含），如果是则会将 chunk 迁移到分片区间所关联的分片，否则按一般情况处理。

分片标签适用于一些特定的场景。例如，集群中可能同时存在 OLTP 和 OLAP 处理，一些系统日志的重要性相对较低，而且主要以少量的统计分析为主。为了便于单独扩展，我们可能希望将日志与实时类的业务数据分开，此时就可以使用标签。

为了让分片拥有指定的标签，需执行 `addShardTag` 命令：

```javascript
sh.addShardTag("shard01","oltp")
sh.addShardTag("shard02","oltp")
sh.addShardTag("shard03","olap")
```

实时计算的集合应该属于 `oltp` 标签，声明 `TagRange`：

```javascript
sh.addTagRange("main.devices",{shardKey:MinKey},{shardKey:MaxKey},"oltp")
```

`shardKey` 要换成你的分片键。

而离线计算的集合，则属于 `olap` 标签：

```javascript
sh.addTagRange("other.systemLogs",{shardKey:MinKey},{shardKey:MaxKey},"olap")
```

`main.devices` 集合将被均衡地分发到 shard01、shard02 分片上，而 `other.systemLogs` 集合将被单独分发到 shard03 分片上。

### 分片键选择

在选择分片键时，需要根据业务的需求及范围分片、哈希分片的不同特点进行权衡。一般来说，在设计分片键时需要考虑的因素包括：
- 分片键的基数（cardinality），取值基数越大越有利于扩展。 
  - 以性别作为分片键：数据最多被拆分为 2 份 
  - 以月份作为分片键：数据最多被拆分为 12 份
- 分片键的取值分布应该尽可能均匀。
- **业务读写模式，尽可能分散写压力，而读操作尽可能来自一个或少量的分片**。
- 分片键应该能适应大部分的业务操作。

### 分片键约束

**ShardKey 必须是一个索引**。非空集合须在 ShardCollection 前创建索引；空集合 ShardCollection 自动创建索引 

4.4 版本之前：

- ShardKey 大小不能超过 512 Bytes； 
- 仅支持单字段的哈希分片键； 
- Document 中必须包含 ShardKey； 
- ShardKey 包含的 Field 不可以修改。 

4.4 版本之后: 

- ShardKey 大小无限制； 
- 支持复合哈希分片键； 
- Document 中可以不包含 ShardKey，插入时被当 做 `Null` 处理； 
- 为 ShardKey 添加后缀 `refineCollectionShardKey` 命令，可以修改 ShardKey 包含的 Field；     

而在 4.2 版本之前，ShardKey 对应的值不可以修改；4.2 版本之后，如果 ShardKey 为非 `_id` 字段， 那么可以修改 ShardKey 对应的值。

## 数据均衡

### 均衡的方式

一种理想的情况是，所有加入的分片都发挥了相当的作用，包括提供更大的存储容量，以及读写访问性能。因此，为了保证分片集群的水平扩展能力，业务数据应当尽可能地保持均匀分布。这里的均匀性包含以下两个方面：

1. 所有的数据应均匀地分布于不同的 chunk 上。
2. 每个分片上的c hunk 数量尽可能是相近的。

其中，第 1 点由业务场景和分片策略来决定，而关于第 2 点，有两种选择。

#### 手动均衡

- 一种做法是，在**初始化集合时预分配一定数量的 `chunk`（仅适用于哈希分片）**，比如给 10 个分片分配 1000 个 chunk，那么每个分片拥有 100 个 chunk。
- 另一种做法则是，可以通过 `splitAt`、`moveChunk` 命令进行手动切分、迁移。

#### 自动均衡

开启 MongoDB 集群的自动均衡功能。**均衡器会在后台对各分片的 chunk 进行监控，一旦发现了不均衡状态就会自动进行 chunk 的搬迁以达到均衡**。其中，chunk 不均衡通常来自于两方面的因素：

- 一方面，在没有人工干预的情况下，chunk 会持续增长并产生分裂（split），而不断分裂的结果就会出现数量上的不均衡；
- 另一方面，在动态增加分片服务器时，也会出现不均衡的情况。自动均衡是开箱即用的，可以极大简化集群的管理工作。

### chunk 分裂

**在默认情况下，一个 chunk 的大小为 64 MB（MongoDB 6.0默认是 128M）**，该参数由配置的 `chunksize` 参数指定。如果持续地向该 chunk 写入数据，并导致数据量超过了 chunk 大小，则 MongoDB 会自动进行分裂，将该 chunk 切分为两个相同大小的 chunk。

![mongodb-shards-split](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-shards-split.png)


**chunk 分裂是基于分片键进行的，如果分片键的基数太小，则可能因为无法分裂而会出现 jumbo chunk（超大块）的问题**。例如，对 `db.users` 使用 gender（性别）作为分片键，由于同一种性别的用户数可能达到数千万，分裂程序并不知道如何对分片键（gender）的一个单值进行切分，因此最终导致在一个 chunk 上集中存储了大量的 user 记录（总大小超过 64MB）。

**jumbo chunk 对水平扩展有负面作用，该情况不利于数据的均衡，业务上应尽可能避免**。一些写入压力过大的情况可能会导致 chunk 多次失败（split），最终当 chunk 中的文档数大于 `1.3×avgObjectSize` 时会导致无法迁移。此外在一些老版本中，如果 chunk 中的文档数超过 250000 个，也会导致无法迁移。

### 自动均衡原理

**均衡器运行于 Primary Config Server（配置服务器的主节点）上**，而该节点也同时会控制 chunk 数据的搬迁流程。

![mongodb-shards-balance](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-shards-balance.png)


流程说明：

1. 分片 shard0 在持续的写入压力下，产生了 chunk 分裂。
2. 分片服务器通知 Config Server 更新元数据。
3. Config Server 上的自动均衡器对 chunk 分布进行检查，发现 shard0 和 shard1 上的 chunk 数量差异达到的阈值，于是向 shard0 下发 `moveChunk` 命令，将 chunk 迁移到 shard1 上。
4. shard0 收到 `moveChunk` 命令后，将 chunk 复制到 shard1 上。该阶段hi完成索引、chunk 数据的复制，而且在整个过程中业务侧对数据的操作仍然指向 shard0。所以在第一轮复制完毕之后，目标 shard1 会向 shard0 确认是否还存在增量更新的数据，如果存在则继续复制。
5. 迁移完成后，shard0 通知 Config Server 更新元数据，将 chunk 的位置更新为 shard1。在更新元数据之后确保没有关联 cursor 的情况下，shard0 会删除被迁移的 chunk 副本。
6. Config Server 通知 mongos 服务器更新路由表。新的请求会被路由到 shard1。

#### 迁移阈值

**MongoDB 4.4 迁移条件**：

均衡器对数据的不均衡判断是根据分片上的 chunk 个数差异来进行的：

| chunk 个数 | 迁移阈值 |
| ---------- | -------- |
| 少于 20 个  | 2        |
| `20 ~ 79`  | 4        |
| 80 及以上   | 8        |

[官方文档](https://www.mongodb.com/docs/v4.4/core/sharding-balancer-administration/)

**MongoDB 6.0 迁移条件**：

如果碎片之间的数据差异 (对于该集合) 小于该集合配置范围大小的三倍，则认为该集合是平衡的。对于 128MB 的默认范围大小，对于给定的集合，两个分片必须具有至少 384MB 的数据大小差异，才能进行迁移。

[官方文档](https://www.mongodb.com/docs/v6.0/core/sharding-balancer-administration/)

#### 迁移的速度

数据均衡的整个过程并不是很快，影响 MongoDB 均衡速度的几个选项如下：

- `_secondaryThrottle`：**用于调整迁移数据写到目标分片的安全级别**。如果没有设定，则会使用 `w：2` 选项，即至少一个备节点确认写入迁移数据后才算成功。从 MongoDB 3.4 版本开始，`_secondaryThrottle` 被默认设定为 `false`, chunk 迁移不再等待备节点写入确认。
- `_waitForDelete`：在 chunk 迁移完成后，源分片会将不再使用的 chunk 删除。如果 `_waitForDelete` 是 `true`，那么均衡器需要等待chunk 同步删除后才进行下一次迁移。该选项默认为 **`false`，这意味着对于旧 chunk 的清理是异步进行的**。
- 并行迁移数量：在早期版本的实现中，均衡器在同一时刻只能有一个 chunk 迁移任务。从 MongoDB 3.4 版本开始，**允许 n 个分片的集群同时执行 `n/2` 个并发任务**。

随着版本的迭代，MongoDB 迁移的能力也在逐步提升。从 MongoDB 4.0 版本开始，支持在迁移数据的过程中并发地读取源端和写入目标端，迁移的整体性能提升了约 40%。这样也使得新加入的分片能更快地分担集群的访问读写压力。


### 数据均衡带来的问题

**数据均衡会影响性能**，在分片间进行数据块的迁移是一个“繁重”的工作，很容易带来磁盘 I/O 使用率飙升，或业务时延陡增等一些问题。因此，建议尽可能提升磁盘能力，如使用 SSD。除此之外，我们还可以将数据均衡的窗口对齐到业务的低峰期以降低影响。

登录 mongos，在 `config` 数据库上更新配置，代码如下：

```bash
use config
sh.setBalancerState(true)
db.settings.update(
    {_id:"balancer"},
    {$set:{activeWindow:{start:"02:00",stop:"04:00"}}},
    {upsert:true}
)
```

在上述操作中启用了自动均衡器，同时在每天的凌晨 2 点到 4 点运行数据均衡操作。

对**分片集合中执行 `count` 命令可能会产生不准确的结果**，mongos 在处理 `count` 命令时会分别向各个分片发送请求，并累加最终的结果。**如果分片上正在执行数据迁移，则可能导致重复的计算**。替代办法是使用 `db.collection.countDocuments({})` 方法，该方法会执行聚合操作进行实时扫描，可以避免元数据读取的问题，但需要更长时间。

**在执行数据库备份的期间，不能进行数据均衡操作**，否则会产生不一致的备份数据。在备份操作之前，可以通过如下命令确认均衡器的状态:

1. `sh.getBalancerState()`：查看均衡器是否开启。
2. `sh.isBalancerRunning()`：查看均衡器是否正在运行。
3. `sh.getBalancerWindow()`：查看当前均衡的窗口设定。

## MongoDB 高级集群架构设计

### 两地三中心集群架构设计

#### 容灾级别

![mongodb-shards-dr](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-shards-dr.png)


#### RPO&RTO

- RPO（Recovery Point Objective）：即**数据恢复点目标，主要指的是业务系统所能容忍的数据丢失量**。
- RTO（Recovery Time Objective）：即**恢复时间目标，主要指的是所能容忍的业务停止服务的最长时间**，也就是从灾难发生到业务系统恢复服务功能所需要的最短时间周期。

![mongodb-shards-dr2](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-shards-dr2.png)

#### MongoDB 两地三中心方案：复制集跨中心部署

[两地三中心方案](https://www.processon.com/view/link/6239de401e085306f8cc23ef)

**双中心双活＋异地热备=两地三中心**：

![mongodb-shards-dr3](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-shards-dr3.png)

MongoDB 集群两地三中心部署的考量点

- 节点数量建议要 5 个，`2+2+1` 模式
- **主数据中心的两个节点要设置高一点的优先级**，减少跨中心换主节点
- 同城双中心之间的网络要保证低延迟和频宽，满足 `writeConcern: Majority` 的双中心写需求
- 使用 Retryable Writes and Retryable Reads 来保证零下线时间
- 用户需要自行处理好业务层的双中心切换


### 两地三中心复制集搭建

**环境准备**

- 3 台 Linux 虚拟机，准备 MongoDB 环境，配置环境变量。
- 一定要版本一致（重点）

![mongodb-shards-dr](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-shards-dr.png)

配置域名解析

在 3 台虚拟机上分别执行以下 3 条命令，注意替换实际 IP 地址：

```bash
echo "192.168.65.97  mongo1 mongo01.com mongo02.com" >> /etc/hosts
echo "192.168.65.190 mongo2 mongo03.com mongo04.com" >> /etc/hosts
echo "192.168.65.200 mongo3 mongo05.com " >> /etc/hosts
```

启动 5 个 MongoDB 实例：

```bash
# mongo1
mkdir -p /data/member1/db  /data/member1/log /data/member2/db  /data/member2/log
mongod --dbpath /data/member1/db --replSet demo --bind_ip 0.0.0.0 --port 10001 --fork --logpath /data/member1/log/member1.log
mongod --dbpath /data/member2/db --replSet demo --bind_ip 0.0.0.0 --port 10002 --fork --logpath /data/member2/log/member2.log


# mongo2
mkdir -p /data/member3/db  /data/member3/log /data/member4/db  /data/member4/log
mongod --dbpath /data/member3/db --replSet demo --bind_ip 0.0.0.0 --port 10001 --fork --logpath /data/member3/log/member3.log
mongod --dbpath /data/member4/db --replSet demo --bind_ip 0.0.0.0 --port 10002 --fork --logpath /data/member4/log/member4.log


# mongo3
mkdir -p /data/member5/db  /data/member5/log
mongod --dbpath /data/member5/db --replSet demo --bind_ip 0.0.0.0 --port 10001 --fork --logpath /data/member5/log/member5.log
```

初始化复制集：

```bash
mongo mongo01.com:10001
# 初始化复制集
rs.initiate({
    "_id" : "demo",
    "version" : 1,
    "members" : [
        { "_id" : 0, "host" : "mongo01.com:10001" },
        { "_id" : 1, "host" : "mongo02.com:10002" },
        { "_id" : 2, "host" : "mongo03.com:10001" },
       { "_id" : 3, "host" : "mongo04.com:10002" },
       { "_id" : 4, "host" : "mongo05.com:10001" }
    ]
})
# 查看复制集状态
rs.status()
```

配置选举优先级

把 mongo1 上的 2 个实例的选举优先级调高为 5 和 10（默认为1），给**主数据中心更高的优先级**：

```bash
mongosh mongo01.com:10001
conf = rs.conf()
conf.members[0].priority = 5
conf.members[1].priority = 10
rs.reconfig(conf)
```

启动持续写脚本（每 2 秒写一条记录）:

```bash
# mongo3
mongosh --retryWrites mongodb://mongo01.com:10001,mongo02.com:10002,mongo03.com:10001,mongo04.com:10002,mongo05.com:10001/test?replicaSet=demo ingest-script

# vim ingest-script
db.test.drop()
for(var i=1;i<1000;i++){
    db.test.insert({item: i});
    inserted = db.test.findOne({item: i});
    if(inserted)
        print(" Item "+ i +" was inserted " + new Date().getTime()/1000);
    else
        print("Unexpected "+ inserted)
    sleep(2000);
}
```

**总结**

- 搭建简单，使用复制集机制，无需第三方软件 
- 使用 Retryable Writes 以后，即使出现数据中心故障，对前端业务没有任何中断 （Retryable Writes 在 4.2 以后就是默认设置）

### 全球多写集群架构设计

[全球多写集群方案](https://www.processon.com/view/link/6239de277d9c08070e59dc0d)，必须用到分片集群。

![mongodb-shards-global-multi](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-shards-global-multi.png)