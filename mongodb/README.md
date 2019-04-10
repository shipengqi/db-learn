# MongoDB
MongoDB 有各种语言的 [官方驱动](https://docs.mongodb.com/ecosystem/drivers/)。

## MongoDB 相比 RDBMS 的优势
- 模式较少：MongoDB 是一种文档数据库，一个集合可以包含各种不同的文档。每个文档的字段数、内容以及文档大小都可以各不相同。
- 采用单个对象的模式，清晰简洁。
- 没有复杂的连接功能。
- 深度查询功能。MongoDB 支持对文档执行动态查询，使用的是一种不逊色于 SQL 语言的基于文档的查询语言。
- 具有调优功能。
- 易于扩展。MongoDB 非常易于扩展。
- 不需要从应用对象到数据库对象的转换/映射。
- 使用内部存储存储（窗口化）工作集，能够更快地访问数据。

## 为何选择使用 MongoDB
- 面向文档的存储：以 JSON 格式的文档保存数据。
- 任何属性都可以建立索引。
- 复制以及高可扩展性。
- 自动分片。
- 丰富的查询功能。
- 快速的即时更新。

## 适用场景
### 无模式(Flexible Schema)
面向文档数据库经常吹嘘的一个好处就是，它不需要一个固定的模式。这使得他们比传统的数据库表要灵活得多。无模式是酷，可是大多数情况下你的数据结构还是应当好好设计的。

### 写操作
MongoDB 可以胜任的一个特殊角色是在日志领域。有两点使得 MongoDB 的写操作非常快。
- 发送了写操作命令之后立刻返回，而无须等到操作完成。
- 可以控制数据持久性的写行为。

#### 受限集合
MongoDB 还提供了 [受限集合(capped collection)](https://docs.mongodb.com/manual/core/capped-collections/)。可以通过 `db.createCollection` 命令来创建一个受限集合并标记它的限制:
```js
//limit our capped collection to 1 megabyte
db.createCollection('logs', {capped: true, size: 1048576})
```
另外一种限制可以基于文档个数，而不是大小，用 `max` 标记。

#### TTL索引
如果想让你的数据 "过期" ，基于时间而不是整个集合的大小，可以用 [TTL 索引](https://docs.mongodb.com/manual/tutorial/expire-data/) ，所谓 TTL 是 "time-to-live" 的缩写。

### 持久性(Durability)
从 2.0 版的 MongoDB 开始，日志是默认启动的，该功能允许快速恢复服务器，比如遭遇到了服务器崩溃或者停电的情况。

### 全文检索(Full Text Search)
全文检索是在最近加入到 MongoDB 中的。它支持十五国语言，支持词形变化(stemming)和干扰字(stop words)。除了原生的 MongoDB 的全文检索支持，如果你
需要一个更强大更全面的全文检索引擎的话，你需要另找方案。

### 事务(Transactions)
MongoDB 不支持事务。这有两个代替案，一个很好用但有限制，另外一个比较麻烦但灵活。

#### 原子更新操作
MongoDB 提供了针对单个文档的原子操作。比如 `$inc` 和 `$set`。还有像 `findAndModify` 命令，可以更新或删除文档之后，自动返回修改过的文档。

维持原子性的建议方法是利用**内嵌文档**（embedded document）将所有经常更新的相关信息都保存在一个文档中。这能确保所有针对单一文档的更新具有原子性。

#### 两段提交
两段提交实际上在关系型数据库世界中非常常用，用来实现多数据库之间的事务。 MongoDB 网站 [有个例子](https://docs.mongodb.com/manual/core/transactions/) 演示了最典型的
场合 (资金转账)。通常的想法是，把事务的状态保存到实际的原子更新的文档中，然后手工的进行 `init-pending-commit/rollback` 处理。

MongoDB 支持内嵌文档以及它灵活的 schema 设计，让两步提交没那么痛苦，但是它仍然不是一个好处理。

### 数据处理(Data Processing)
MongoDB 依赖 MapReduce 来解决大部分数据处理工作。

在 2.2 版本，它追加了一个强力的功能，叫做 [aggregation framework or pipeline](https://docs.mongodb.com/manual/core/aggregation-pipeline/)，因此你只
要对那些尚未支持管道的，需要使用复杂方法的，不常见的聚合使用 MapReduce。

### 地理空间查询(Geospatial)
MongoDB 支持 [geospatial 索引](https://docs.mongodb.com/manual/geospatial-queries/)。这
允许你保存 `geoJSON` 或者 `x` 和 `y` 坐标到文档，并查询文档，用如 `$near` 来获取坐标集，或者 `$within` 来获取一个矩形或圆中的点。
