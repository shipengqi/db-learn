---
title: 存储原理
weight: 4
---

存储引擎是数据库的组件，负责管理数据如何存储在内存和磁盘上。MongoDB 支持多个存储引擎，因为不同的引擎对于特定的工作负载表现更好。选择合适的存储引擎可以显著影响应用程序的性能。

## WiredTiger
MongoDB 从 3.0 开始引入可插拔存储引擎的概念，主要有 MMAPV1、WiredTiger 存储引擎可供选择。**从 MongoDB 3.2 开始，WiredTiger 存储引擎是默认的存储引擎**。从 4.2 版开始，MongoDB 删除了废弃的 MMAPv1 存储引擎。

### WiredTiger 读写模型

#### 读缓存

**理想情况下，MongoDB 可以提供近似内存式的读写性能**。WiredTiger 引擎实现了数据的二级缓存，第一层是操作系统的页面缓存，第二层则是引擎提供的内部缓存。

**MongoDB 为了尽可能保证业务查询的“热数据”能快速被访问，其内部缓存的默认大小达到了内存的一半**，该值由`wiredTigerCacheSize` 参数指定，其默认的计算公式如下：

```javascript
wiredTigerCacheSize=Math.max(0.5*(RAM-1GB),256MB)
```

#### 写缓冲

**当数据发生写入时，MongoDB 并不会立即持久化到磁盘上，而是先在内存中记录这些变更，之后通过C heckPoint 机制将变化的数据写入磁盘**。这么处理主要有以下两个原因：

- 如果每次写入都触发一次磁盘 I/O，那么开销太大，而且响应时延会比较大。
- 多个变更的写入可以尽可能进行 I/O 合并，降低资源负荷。

#### MongoDB 会丢数据吗？

MongoDB 单机下保证数据可靠性的机制包括以下两个部分。

##### CheckPoint（检查点）机制

**快照（snapshot）描述了某一时刻（point-in-time）数据在内存中的一致性视图，而这种数据的一致性是 WiredTiger 通过 MVCC（多版本并发控制）实现的**。当建立 CheckPoint 时，WiredTiger 会在内存中建立所有数据的一致性快照，并将该快照覆盖的所有数据变化一并进行持久化（fsync）。成功之后，内存中数据的修改才得以真正保存。**默认情况下，MongoDB 每 60s 建立一次 CheckPoint**，在检查点写入过程中，上一个检查点仍然是可用的。这样可以保证一旦出错，MongoDB 仍然能恢复到上一个检查点。

{{< callout type="info" >}}
**如果只有 CheckPoint 机制，那么在发生宕机时，还没有刷盘，那么这 60s 内的数据就会丢失**。为了保证数据的完整性，MongoDB 还引入了 Journal 日志机制。
{{< /callout >}}

##### Journal 日志

**Journal 是一种预写式日志（write ahead log）机制**，主要用来弥补 CheckPoint 机制的不足。如果开启了 Journal 日志，那么 WiredTiger 会将每个写操作的 redo 日志写入 Journal 缓冲区，该缓冲区会频繁地将日志持久化到磁盘上。**默认情况下，Journal 缓冲区每 100ms 执行一次持久化**。此外，**Journal 日志达到 100MB，或是应用程序指定 `journal：true`，写操作都会触发日志的持久化**。一旦 MongoDB 发生宕机，**重启程序时会先恢复到上一个检查点，然后根据 Journal 日志恢复增量的变化**。由于 Journal 日志持久化的间隔非常短，数据能得到更高的保障，如果按照当前版本的默认配置，则其在断电情况下最多会丢失 100ms 的写入数据。

{{< callout type="info" >}}
对于类似于订单系统这样的业务场景，由于数据的重要性，插入数据时直接指定 `journal：true`，这样可以保证数据的可靠性。
{{< /callout >}}

![wiredtiger-journal]()

WiredTiger 写入数据的流程：

1. 应用向 MongoDB 写入数据（插入、修改或删除）。
2. 数据库从内部缓存中获取当前记录所在的页块，如果不存在则会从磁盘中加载（Buffer I/O） 
3. WiredTiger 开始执行写事务，修改的数据写入页块的一个更新记录表，此时原来的记录仍然保持不变。
4. 如果开启了 Journal 日志，则在写数据的同时会写入一条 Journal 日志（Redo Log）。该日志在最长不超过 100ms 之后写入磁盘。
5. 数据库每隔 60s 执行一次 CheckPoint 操作，此时内存中的修改会真正刷入磁盘。

**Journal 日志的刷新周期可以通过参数 `storage.journal.commitIntervalMs` 指定**，MongoDB 3.4 及以下版本的默认值是 50ms，而 3.6 版本之后调整到了 100ms。由于 Journal 日志采用的是顺序 I/O 写操作，频繁地写入对磁盘的影响并不是很大。

**CheckPoint 的刷新周期可以调整 `storage.syncPeriodSecs` 参数（默认值 60s），在 MongoDB 3.4 及以下版本中，当 Journal 日志达到 2GB 时同样会触发 CheckPoint 行为。如果应用存在大量随机写入，则 CheckPoint 可能会造成磁盘 I/O 的抖动。在磁盘性能不足的情况下，问题会更加显著，此时适当缩短 CheckPoint 周期可以让写入平滑一些。

## 多文档事务

事务（transaction）是传统数据库所具备的一项基本能力，其根本目的是为数据的可靠性与一致性提供保障。而在通常的实现中，**事务包含了一个系列的数据库读写操作，这些操作要么全部完成，要么全部撤销**。例如，在电子商城场景中，当顾客下单购买某件商品时，除了生成订单，还应该同时扣减商品的库存，这些操作应该被作为一个整体的执行单元进行处理，否则就会产生不一致的情况。

**在 MongoDB 中，对单个文档的操作是原子的**。由于可以在单个文档结构中使用内嵌文档和数组来获得数据之间的关系，而不必跨多个文档和集合进行范式化，所以这种单文档原子性避免了许多实际场景中对多文档事务的需求。

对于那些需要对多个文档（在单个或多个集合中）进行原子性读写的场景，MongoDB 支持多文档事务。而使用分布式事务，事务可以跨多个操作、集合、数据库、文档和分片使用。

**MongoDB 虽然已经在 4.2 开始全面支持了多文档事务**，但并不代表大家应该毫无节制地使用它。相反，**对事务的使用原则应该是：能不用尽量不用**。通过合理地设计文档模型，可以规避绝大部分使用事务的必要性。

**使用事务的原则**：

- 无论何时，事务的使用总是**能避免则避免**。
- **模型设计先于事务，尽可能用模型设计规避事务**，使用内嵌文档和数组来避免跨表的操作。
- 不要使用过大的事务（尽量控制在 1000 个文档更新以内）。
- 当必须使用事务时，**尽可能让涉及事务的文档分布在同一个分片上**，这将有效地提高效率。

### MongoDB 对事务支持

| 事务属性 | 支持程度 |
| --- | --- |
| A 原子性 | 单文档支持：1.x 就已经支持。<br /> 复制集多表多行：4.0。<br /> 分片集群多表多行：4.2 |
| C 一致性 | `writeConcern`、`readConcern`：3.2 |
| I 隔离性 | `readConcern`：3.2 |
| D 持久性 | Journal and Replication |

#### 使用

MongoDB 多文档事务的使用方式与关系数据库非常相似：

```javascript
try (ClientSession clientSession = client.startSession()) {
   clientSession.startTransaction(); 
   collection.insertOne(clientSession, docOne); 
   collection.insertOne(clientSession, docTwo); 
   clientSession.commitTransaction(); 
}
```

#### writeConcern

**`writeConcern` 决定一个写操作落到多少个节点上才算成功**。MongoDB支持客户端灵活配置写入策略（`writeConcern`），以满足不同场景的需求。

语法格式：

```javascript
{ w: <value>, j: <boolean>, wtimeout: <number> }
```

- `w`：数据写入到 number 个节点才向客户端发送确认
  - `{w: 0}` 对客户端的写入不需要发送任何确认，适用于性能要求高，但不关注正确性的场景。
  - `{w: 1}` 默认的 `writeConcern`，数据写入到 Primary 就向客户端发送确认。
  - `{w: "majority"}` 数据写入到副本集大多数成员后向客户端发送确认，适用于对数据安全性要求比较高的场景，该选项会降低写入性能。
- `j`：写入到 journal 持久化之后才向客户端确认
  - `{j: false}`，默认值，如果要求 Primary 写入持久化了才向客户端确认，则指定该选项为 `true`。
- `wtimeout`: 写入超时时间，仅 `w` 的值大于 1 时有效。
  - 当指定 `w` 时，数据需要成功写入 number 个节点才算成功，**如果写入过程中有节点故障，可能导致这个条件一直不能满足，从而一直不能向客户端发送确认结果**，针对这种情况，客户端可**设置 `wtimeout` 选项来指定超时时间，当写入过程持续超过该时间仍未结束，则认为写入失败**。

**对于 5 个节点的复制集来说，写操作落到多少个节点上才算是安全的**?

3 个。最好使用 `{w: "majority"}`，这种方式更灵活一点，节点增加，减少不用再修改代码。

测试，包含延迟节点的 3 个节点 pss 复制集：

```javascript
db.user.insertOne({name:"李四"},{writeConcern:{w:"majority"}})

// 配置延迟节点
cfg = rs.conf()
cfg.members[2].priority = 0
cfg.members[2].hidden = true
cfg.members[2].secondaryDelaySecs = 60
rs.reconfig(cfg)

// 等待延迟节点写入数据后才会响应
db.user.insertOne({name:"王五"},{writeConcern:{w:3}})
// 超时写入失败
db.user.insertOne({name:"小明"},{writeConcern:{w:3,wtimeout:3000}})
```

{{< callout type="info" >}}
- 虽然多于半数的 `writeConcern` 都是安全的，但通常只会**设置 `majority`，因为这是等待写入延迟时间最短的选择**； 
- **不要设置 `writeConcern` 等于总节点数**，因为一旦有一个节点故障，所有写操作都将失败；
- `writeConcern` 虽然会增加写操作延迟时间，但并不会显著增加集群压力，因此无论是否等待，写操作最终都会复制到所有节点上。设置 `writeConcern` 只是让写操作等待复制后再返回而已；
- **重要数据应用 `{w: "majority"}`，普通数据可以应用 `{w: 1}` 以确保最佳性能**。
{{< /callout >}}

#### readPreference

在读取数据的过程中我们需要关注以下两个问题： 

- 从哪里读？（要不要做读写分离，是优先主节点还是优先从节点）
- 什么样的数据可以读？ 

**`readPreference` 决定使用哪一个节点来满足正在发起的读请求**。可选值包括：

- `primary`: 只选择主节点，默认模式； 
- `primaryPreferred`：优先选择主节点，如果主节点不可用则选择从节点； 
- `secondary`：只选择从节点； 
- `secondaryPreferred`：优先选择从节点， 如果从节点不可用则选择主节点； 
- `nearest`：根据客户端对节点的 Ping 值判断节点的远近，选择从最近的节点读取。

**合理的 ReadPreference 可以极大地扩展复制集的读性能，降低访问延迟**。

#### readConcern

#### 事务的隔离级别

### 事务超时

### 事务的错误处理