---
title: 开发规范和建模优化
weight: 5
---

## 开发规范

1. **命名原则**。数据库、集合命名需要简单易懂，数据库名使用小写字符，集合名称使用统一命名风格，可以**统一大小写或使用驼峰式命名**。数据库名和集合名称均不能超过 64 个字符。
2. 集合设计。**对少量数据的包含关系，使用嵌套模式有利于读性能和保证原子性的写入**。对于复杂的关联关系，以及后期可能发生演进变化的情况，建议使用引用模式。
3. 文档设计。**避免使用大文档，MongoDB 的文档最大不能超过 16MB**。如果使用了**内嵌的数组对象或子文档，应该保证内嵌数据不会无限制地增长**。在文档结构上，尽可能减少字段名的长度，MongoDB 会保存文档中的字段名，因此字段名称会影响整个集合的大小以及内存的需求。一般建议将字段名称控制在 32 个字符以内。
4. 索引设计。**在必要时使用索引加速查询。避免建立过多的索引，单个集合建议不超过 10 个索引**。MongoDB 对集合的写入操作很可能也会触发索引的写入，从而触发更多的I/O操作。无效的索引会导致内存空间的浪费，因此有必要对索引进行审视，及时清理不使用或不合理的索引。遵循索引优化原则，如覆盖索引、优先前缀匹配等，使用 `explain` 命令分析索引性能。
5. 分片设计。**对可能出现快速增长或读写压力较大的业务表考虑分片**。分片键的设计满足均衡分布的目标，业务上尽量避免广播查询。应尽早确定分片策略，**最好在集合达到 256GB 之前就进行分片**。如果集合中存在唯一性索引，则应该确保该索引覆盖分片键，避免冲突。为了降低风险，单个分片的数据集合大小建议不超过 2TB。
6. 升级设计。应用上需支持对旧版本数据的兼容性，在添加唯一性约束索引之前，对数据表进行检查并及时清理冗余的数据。新增、修改数据库对象等操作需要经过评审，并保持对数据字典进行更新。
7. 考虑数据老化问题，**要及时清理无效、过期的数据，优先考虑为系统日志、历史数据表添加合理的老化策略**。
8. 数据一致性方面，**非关键业务使用默认的 `WriteConcern：1`（更高性能写入）；对于关键业务类，使用 `WriteConcern：majority` 保证一致性（性能下降）。如果业务上严格不允许脏读，则使用 `ReadConcern：majority` 选项**。
9. 使用 `update`、`findAndModify` 对数据进行修改时，**如果设置了 `upsert：true`，则必须使用唯一性索引避免产生重复数据**。
10. 业务上尽量避免短连接，使用官方最新驱动的连接池实现，**控制客户端连接池的大小，最大值建议不超过 200**。
11. 对大量数据写入使用 Bulk Write 批量化 API，建议使用无序批次更新。
12. **优先使用单文档事务保证原子性**，如果需要使用多文档事务，则必须保证事务尽可能小，一个事务的执行时间最长不能超过 60s。
13. **在条件允许的情况下，利用读写分离降低主节点压力**。对于一些统计分析类的查询操作，可优先从节点上执行。
14. 考虑业务数据的隔离，例如将配置数据、历史数据存放到不同的数据库中。微服务之间使用单独的数据库，尽量避免跨库访问。
15. 维护数据字典文档并保持更新，提前按不同的业务进行数据容量的规划。

## 建模案例分析

[官方建模示例](https://www.mongodb.com/zh-cn/docs/manual/tutorial/model-embedded-one-to-one-relationships-between-documents/)。

### 一对一关系模型

例如，模式包含两个实体，一个 patron 和一个 address：

```javscript
// patron document
{
   _id: "joe",
   name: "Joe Bookreader"
}
// address document
{
   street: "123 Fake Street",
   city: "Faketon",
   state: "MA",
   zip: "12345"
}
```

#### 嵌入式文档模式

将 address 信息嵌入 patron 文档：

```javascript
{
   _id: "joe",
   name: "Joe Bookreader",
   address: {
              street: "123 Fake Street",
              city: "Faketon",
              state: "MA",
              zip: "12345"
            }
}
```

一个人的地址信息不会太多，那么使用嵌入式文档就比较合适。可以一次查询拿到所有信息。

#### 子集模式

嵌入式文档模式的一个潜在问题是，例如一个文档包含了很多信息字段，但是通常在使用时，只有其中的几个字段才会被用到。如果将所有的字段都放到一个文档中，那么在查询时，返回所有的字段，会导致网络传输和磁盘空间的浪费。

```javascript
{
  "_id": 1,
  "title": "The Arrival of a Train",
  "year": 1896,
  "runtime": 1,
  "released": ISODate("01-25-1896"),
  "poster": "http://ia.media-imdb.com/images/M/MV5BMjEyNDk5MDYzOV5BMl5BanBnXkFtZTgwNjIxMTEwMzE@._V1_SX300.jpg",
  "plot": "A group of people are standing in a straight line along the platform of a railway station, waiting for a train, which is seen coming at some distance. When the train stops at the platform, ...",
  "fullplot": "A group of people are standing in a straight line along the platform of a railway station, waiting for a train, which is seen coming at some distance. When the train stops at the platform, the line dissolves. The doors of the railway-cars open, and people on the platform help passengers to get off.",
  "lastupdated": ISODate("2015-08-15T10:06:53"),
  "type": "movie",
  "directors": [ "Auguste Lumière", "Louis Lumière" ],
  "imdb": {
    "rating": 7.3,
    "votes": 5043,
    "id": 12
  },
  "countries": [ "France" ],
  "genres": [ "Documentary", "Short" ],
  "tomatoes": {
    "viewer": {
      "rating": 3.7,
      "numReviews": 59
    },
    "lastUpdated": ISODate("2020-01-09T00:02:53")
  }
}
```

`movie` 只存储常用的基本信息：

```javascript
// movie collection

{
  "_id": 1,
  "title": "The Arrival of a Train",
  "year": 1896,
  "runtime": 1,
  "released": ISODate("1896-01-25"),
  "type": "movie",
  "directors": [ "Auguste Lumière", "Louis Lumière" ],
  "countries": [ "France" ],
  "genres": [ "Documentary", "Short" ],
}
```

`movie_details` 包含每部电影的其他不常访问的数据，通过 `movie_id` 来建立关联关系：

```javascript
// movie_details collection

{
  "_id": 156,
  "movie_id": 1, // reference to the movie collection
  "poster": "http://ia.media-imdb.com/images/M/MV5BMjEyNDk5MDYzOV5BMl5BanBnXkFtZTgwNjIxMTEwMzE@._V1_SX300.jpg",
  "plot": "A group of people are standing in a straight line along the platform of a railway station, waiting for a train, which is seen coming at some distance. When the train stops at the platform, ...",
  "fullplot": "A group of people are standing in a straight line along the platform of a railway station, waiting for a train, which is seen coming at some distance. When the train stops at the platform, the line dissolves. The doors of the railway-cars open, and people on the platform help passengers to get off.",
  "lastupdated": ISODate("2015-08-15T10:06:53"),
  "imdb": {
    "rating": 7.3,
    "votes": 5043,
    "id": 12
  },
  "tomatoes": {
    "viewer": {
      "rating": 3.7,
      "numReviews": 59
    },
    "lastUpdated": ISODate("2020-01-29T00:02:53")
  }
}
```

这种方法可以提高读取性能。

### 一对多关系模型

#### 嵌入式文档模式

一个人有多个地址：


```javascript
// patron document
{
   _id: "joe",
   name: "Joe Bookreader"
}

// address documents
{
   patron_id: "joe", // reference to patron document
   street: "123 Fake Street",
   city: "Faketon",
   state: "MA",
   zip: "12345"
}

{
   patron_id: "joe",
   street: "1 Some Other Street",
   city: "Boston",
   state: "MA",
   zip: "12345"
}
```

将 address 数据嵌入到 patron 中：

```javascript
{
   _id: "joe",
   name: "Joe Bookreader",
   addresses: [
      {
         street: "123 Fake Street",
         city: "Faketon",
         state: "MA",
         zip: "12345"
      },
      {
         street: "1 Some Other Street",
         city: "Boston",
         state: "MA",
         zip: "12345"
      }
   ]
}
```

这种方式只适合数据量少的场景。如果数组无限增长，会导致网络拥堵。而且大量的查询请求，肯定会性能下降。

#### 子集模式

嵌入式文档模式的一个潜在问题是，它可能导致文档过大，尤其是在嵌入式字段没有限制的情况下。

考虑一个包含产品评论列表的电商站点：

```javascript
{
  "_id": 1,
  "name": "Super Widget",
  "description": "This is the most useful item in your toolbox.",
  "price": { "value": NumberDecimal("119.99"), "currency": "USD" },
  "reviews": [
    {
      "review_id": 786,
      "review_author": "Kristina",
      "review_text": "This is indeed an amazing widget.",
      "published_date": ISODate("2019-02-18")
    }
    ...
    {
      "review_id": 777,
      "review_author": "Pablo",
      "review_text": "Amazing!",
      "published_date": ISODate("2019-02-16")
    }
  ]
}
```

将该集合拆分为两个集合，而不存储该产品的所有评论：

`product` 集合存储每个产品的信息，包括该产品的 10 条最新评论。评论按时间倒序排列，用户访问产品页面时，应用程序会加载最近十条评论：

```javascript
{
  "_id": 1,
  "name": "Super Widget",
  "description": "This is the most useful item in your toolbox.",
  "price": { "value": NumberDecimal("119.99"), "currency": "USD" },
  "reviews": [
    {
      "review_id": 786,
      "review_author": "Kristina",
      "review_text": "This is indeed an amazing widget.",
      "published_date": ISODate("2019-02-18")
    }
    ...
    {
      "review_id": 777,
      "review_author": "Pablo",
      "review_text": "Amazing!",
      "published_date": ISODate("2019-02-16")
    }
  ]
}
```

`review` 集合存储所有评论。每条评论都包含对相应产品的引用 `product_id`：

```javascript
{
  "review_id": 786,
  "product_id": 1,
  "review_author": "Kristina",
  "review_text": "This is indeed an amazing widget.",
  "published_date": ISODate("2019-02-18")
}
{
  "review_id": 785,
  "product_id": 1,
  "review_author": "Trina",
  "review_text": "Nice product. Slow shipping.",
  "published_date": ISODate("2019-02-17")
}
...
{
  "review_id": 1,
  "product_id": 1,
  "review_author": "Hans",
  "review_text": "Meh, it's okay.",
  "published_date": ISODate("2017-12-06")
}
```

#### 引用模式

将出版商文档嵌入图书文档会导致出版商数据重复：

```javascript
{
   title: "MongoDB: The Definitive Guide",
   author: [ "Kristina Chodorow", "Mike Dirolf" ],
   published_date: ISODate("2010-09-24"),
   pages: 216,
   language: "English",
   publisher: {
              name: "O'Reilly Media",
              founded: 1980,
              location: "CA"
            }
}

{
   title: "50 Tips and Tricks for MongoDB Developer",
   author: "Kristina Chodorow",
   published_date: ISODate("2011-05-06"),
   pages: 68,
   language: "English",
   publisher: {
              name: "O'Reilly Media",
              founded: 1980,
              location: "CA"
            }
}
```

为避免出现重复的出版商数据，使用**引用**将出版商信息保存在图书集合之外的单独集合中。

使用引用时，关系的增长将决定引用的存储方式。如果每个出版商的图书数量较少且增长有限，则将图书引用存储在出版商文档中有时可能十分有用。

```javascript
{
   name: "O'Reilly Media",
   founded: 1980,
   location: "CA",
   books: [123456789, 234567890, ...]
}

{
    _id: 123456789,
    title: "MongoDB: The Definitive Guide",
    author: [ "Kristina Chodorow", "Mike Dirolf" ],
    published_date: ISODate("2010-09-24"),
    pages: 216,
    language: "English"
}

{
   _id: 234567890,
   title: "50 Tips and Tricks for MongoDB Developer",
   author: "Kristina Chodorow",
   published_date: ISODate("2011-05-06"),
   pages: 68,
   language: "English"
}
```

相反，当每个出版商的图书数量没有限制时，此数据模型将导致可变且不断增长的数组，为避免出现可变且不断增长的数组，可以将出版商的引用存储在图书文档中：

```javascript
{
   _id: "oreilly",
   name: "O'Reilly Media",
   founded: 1980,
   location: "CA"
}

{
   _id: 123456789,
   title: "MongoDB: The Definitive Guide",
   author: [ "Kristina Chodorow", "Mike Dirolf" ],
   published_date: ISODate("2010-09-24"),
   pages: 216,
   language: "English",
   publisher_id: "oreilly"
}

{
   _id: 234567890,
   title: "50 Tips and Tricks for MongoDB Developer",
   author: "Kristina Chodorow",
   published_date: ISODate("2011-05-06"),
   pages: 68,
   language: "English",
   publisher_id: "oreilly"
}
```

### 树状结构模型

![mongodb-tree-model](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-tree-model.png)


这种也不复杂，就是每一个节点就是一个文档，只不过增加了 `parent` 字段：

```javascript
db.categories.insertMany( [
   { _id: "MongoDB", parent: "Databases" },
   { _id: "dbm", parent: "Databases" },
   { _id: "Databases", parent: "Programming" },
   { _id: "Languages", parent: "Programming" },
   { _id: "Programming", parent: "Books" },
   { _id: "Books", parent: null }
])
```

检索节点的父节点的查询：

```javascript
db.categories.findOne( { _id: "MongoDB" } ).parent
```

如果是父节点找子节点的方式，增加一个 `children` 字段：

```javascript
db.categories.insertMany( [
   { _id: "MongoDB", children: [] },
   { _id: "dbm", children: [] },
   { _id: "Databases", children: [ "MongoDB", "dbm" ] },
   { _id: "Languages", children: [] },
   { _id: "Programming", children: [ "Databases", "Languages" ] },
   { _id: "Books", children: [ "Programming" ] }
] )
```

### 朋友圈评论内容管理

社交类的APP需求，一般都会引入“朋友圈”功能，这个产品特性有一个非常重要的功能就是评论体系。

先整理下需求：

- 这个 APP 希望点赞和评论信息都要包含头像信息：
  - 点赞列表，点赞用户的昵称，头像；
  - 评论列表，评论用户的昵称，头像；
- 数据查询则相对简单：
  - 根据分享 ID，批量的查询出 10 条分享里的所有评论内容；


#### 建模

好的设计：

```javascript
{
  "_id": 41,
  "uid": "100000",
  "praise_uid_list": [
    "100010",
    "100011",
    "100012"
  ],
  "comment_msg_list": [
    {
      "100013": "good"
    },
    {
      "100014": "bad"
    }
  ]
}
```

昵称和头像，通过 uid 去用户表查询。通常头像，用户名等信息可以做一层缓存，甚至存储在 APP 端。

### 多列数据结构

需求是基于电影票售卖的不同渠道价格存储。某一个场次的电影，不同的销售渠道对应不同的价格。整理需求为：

- 数据字段：
  - 场次信息；
  - 播放影片信息；
  - 渠道信息，与其对应的价格；
  - 渠道数量最多几十个；
- 业务查询有两种：
  - 根据电影场次，查询某一个渠道的价格；
  - 根据渠道信息，查询对应的所有场次信息；

#### 建模

不好的设计：

```javascript
{
  "scheduleId": "0001",
  "movie": "你的名字",
  "price": {
    "gewala": 30,
    "maoyan": 50,
    "taopiao": 20
  }
}
```

数据表达上基本没有字段冗余，非常紧凑。再来看业务查询能力：

- 根据电影场次，查询某一个渠道的价格：
  - 建立 `createIndex({scheduleId:1, movie:1})` 索引，虽然对 `price` 来说没有创建索引优化，但通过前面两个维度，已经可以定位到唯一的文档，查询效率上来说尚可；
- 根据渠道信息，查询对应的所有场次信息：
  - 为了优化这种查询，需要对每个渠道分别建立索引，例如：`createIndex({"price.gewala":1})` 、`createIndex({"price.maoyan":1})`。
  - 但**渠道会经常变化**，并且为了支持此类查询，**肯能需要创建几十个索引，维护困难**。

```javascript
{
  "scheduleId": "0001",
  "movie": "你的名字",
  "channel": "gewala",
  "price": 30
}
 
{
  "scheduleId": "0001",
  "movie": "你的名字",
  "channel": "maoyan",
  "price": 50
}
 
{
  "scheduleId": "0001",
  "movie": "你的名字",
  "channel": "taopiao",
  "price": 20
}
```

与上面的方案相比，把整个存储对象结构进行了平铺展开，变成了一种表结构，传统的关系数据库多数采用这种类型的方案。信息表达上，把一个对象按照渠道维度拆成多个，**其他的字段进行了冗余存储。如果业务需求再复杂点，造成的信息冗余膨胀非常巨大**。膨胀后带来的副作用会有磁盘空间占用上升，内存命中率降低等缺点。

- 根据电影场次，查询某一个渠道的价格：
  - 建立 `createIndex({scheduleId:1, movie:1, channel:1})` 索引；
- 根据渠道信息，查询对应的所有场次信息：
  - 建立 `createIndex({channel:1})` 索引；

好的设计：

```javascript
{
  "scheduleId": "0001",
  "movie": "你的名字",
  "provider": [
    {
      "channel": "gewala",
      "price": 30
    },
    {
      "channel": "maoyan",
      "price": 50
    },
    {
      "channel": "taopiao",
      "price": 20
    }
  ]
}
```

使用了在 MongoDB 建模中非常容易忽略的结构——**数组**。查询方面的处理，是可以建立 **Multikey Index 索引**。

- 根据电影场次，查询某一个渠道的价格：
  - 建立 `createIndex({scheduleId:1, movie:1, "provider.channel":1})` 索引。
- 根据渠道信息，查询对应的所有场次信息：
  - 建立 `createIndex({"provider.channel":1})` 索引；

### 物联网时序数据建模

案例背景是来自真实的业务，美国州际公路的流量统计。数据库需要提供的能力：

- 存储事件数据
- 提供分析查询能力
- 理想的平衡点：
  - 内存使用
  - 写入性能
  - 读取分析性能
- 可以部署在常见的硬件平台上

#### 建模

每个事件用一个独立的文档存储：

```javascript
{
    segId: "I80_mile23",
    speed: 63,
    ts: ISODate("2013-10-16T22:07:38.000-0500")
}
```

每辆车每秒钟都会写入一条信息。多少的信息，就有多少条数据，数据量增长非常快。数据采集操作全部是 `Insert` 语句。

每分钟的信息用一个独立的文档存储（存储平均值）：

```javascript
{
    segId: "I80_mile23",
    speed_num: 18,
    speed_sum: 1134,
    ts: ISODate("2013-10-16T22:07:00.000-0500")
}
```

- 对每分钟的平均速度计算非常友好（`speed_sum/speed_num`）；
- 数据采集操作基本是 `Update` 语句；
- 数据精度降为一分钟；

每分钟的信息用一个独立的文档存储（秒级记录）：

```javascript
{
    segId: "I80_mile23",
    speed: {0: 63, 1: 58, ... , 58: 66, 59: 64},
    ts: ISODate("2013-10-16T22:07:00.000-0500")
}
```

- 每秒的数据都存储在一个文档中；
- 数据采集操作基本是 `Update` 语句；


每小时的信息用一个独立的文档存储（秒级记录）：

```javascript
{
    segId: "I80_mile23",
    speed: {0: 63, 1: 58, ... , 3598: 54, 3599: 55},
    ts: ISODate("2013-10-16T22:00:00.000-0500")
}
```

相比上面的方案更进一步，从分钟到小时：

- 每小时的数据都存储在一个文档中；
- 数据采集操作基本是 `Update` 语句；
- 更新最后一个时间点（第3599秒），需要3599次迭代（虽然是在同一个文档中）


进一步优化：

```javascript
{
    segId: "I80_mile23",
    speed: {
        0:  {0:47, ..., 59:45},
        ...,
        59: {0:65, ... , 59:56}
    }
    ts: ISODate("2013-10-16T22:00:00.000-0500")
}
```

- 用了嵌套的手法把秒级别的数据存储在小时数据里；
- 数据采集操作基本是 `Update` 语句；
- 更新最后一个时间点（第 3599 秒），需要 `59+59` 次迭代；

##### 每小时的信息用一个独立的文档存储 VS 每分钟的信息用一个独立的文档存储

- 从写入上看：因为 `WiredTiger` 是每分钟进行一次刷盘，所以每小时一个文档的方案，在这一个小时内要被反复的 load 到 PageCache 中，再刷盘；所以，按分钟存储更合理些，可以减少 IO 操作。
- 从读取上看：前者的数据信息量较大，正常的业务请求未必需要这么多的数据，有很大一部分是浪费的。业务上一般很少一次取一个小时的数据，统计的时候可能是按分钟级来计算的。
- 从索引上看：前者的索引更小，内存利用率更高。

## 调优

### 导致 MongoDB 性能不佳的原因

1. 慢查询
2. 阻塞等待
3. 硬件资源不足

**1、2 通常是因为模型/索引设计不佳导致的**。

**排查思路：按 1-2-3 依次排查**。

### 影响 MongoDB 性能的因素

<div class="img-zoom">
  <img src="https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-perf-factors.png" alt="mongodb-perf-factors">
</div>

**网络问题是第一个要排查的问题**，网络没有问题再从客户端、服务端去排查问题。

当 WiredTiger 开启压缩功能时，压缩效率会随着压缩算法的不同而变化。

在 MongoDB 中，WiredTiger 支持的压缩算法有 snappy、zlib、zstd 等。

压缩效率 `zstd > zlib > snappy`，压缩效率越高，文件体积越小，传输的时候越快。但是压缩效率越高也以为这 CPU 消耗越大。

**当内存中的数据需要持久化到磁盘的时候，WiredTiger 会先将数据压缩后再进行持久化，可以节约磁盘空间。当从磁盘中加载的时候，也需要进行解压缩**。

### 性能监控工具

#### mongostat

mongostat 是 MongoDB 自带的监控工具，其可以提供数据库节点或者整个集群当前的状态视图。该功能的设计非常类似于 Linux 系统中的 `vmstat` 命令，可以呈现出实时的状态变化。不同的是，**mongostat 所监视的对象是数据库进程。mongostat 常用于查看当前的 QPS/内存使用/连接数，以及多个分片的压力分布**。mongostat 采用 Go 语言实现，其内部使用了 `db.serverStatus()` 命令，**执行用户需具备 `clusterMonitor` 角色权限**。

```bash
mongostat -h 192.168.65.174 --port 28017 -ufox -pfox --authenticationDatabase=admin --discover -n 300 2
```

参数说明:

- `-h`：指定监听的主机，分片集群模式下指定到一个 mongos 实例，也可以指定单个 mongod，或者复制集的多个节点。
- `--port`：接入的端口，如果不提供则默认为 27017。
- `-u`：接入用户名，等同于 `-user`。
- `-p`：接入密码，等同于 `-password`。
- `--authenticationDatabase`：鉴权数据库。
- `--discover`：启用自动发现，可展示集群中所有分片节点的状态。
- `-n 300 2`：表示输出 300 次，每次间隔 2s。也可以不指定 “-n 300”，此时会一直保持输出。

![mongodb-mongostat](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-mongostat.png)

指标说明：

| 指标名 | 说明 |
| ---   | --- |
| inserts | 每秒插入数 |
| query | 每秒查询数 |
| update | 每秒更新数 |
| delete | 每秒删除数 |
| getmore | 每秒 getmore 数 |
| command | 每秒命令数，涵盖了内部的一些操作 |
| **dirty** | **WiredTiger 缓存中脏数据百分比** |
| **used** | **WiredTiger 正在使用的缓存百分比（如果这个百分比过高，说明内存可能不够用了，可以将内存参数配置的大一点）** |
| flushesWiredTiger | 执行 CheckPoint 的次数 |
| vsize | 虚拟内存使用量 | 
| res | 物理内存使用量 | 
| qrw | 客户端读写等待队列数量，高并发时，一般队列值会升高 |
| arw | 客户端读写活跃个数 |
| netIn | 网络接收数据量 |
| netOut | 网络发送数据量 |
| conn | 当前连接数 | 
| set | 所属复制集名称 |
| repl | 复制节点状态（主节点/二级节点……）|
| time | 时间戳 |

需要关注的指标：

- 插入、删除、修改、查询的速率是否产生较大波动，是否超出预期。
- **qrw、arw：代表目前正在排队(queue)的 read 和 write 请求数量。队列是否较高，若长时间大于 0 则说明此时读写速度较慢**。
- conn：连接数是否太多。
- **dirty：如果这个百分比过高，说明内存中的数据刷盘的效率不高，磁盘 IO 可能存在瓶颈**。
- netIn、netOut：是否超过网络带宽阈值。
- repl：状态是否异常，如 PRI、SEC、RTR 为正常，若出现 REC 等异常值则需要修复。

##### 使用交互模式

mongostat 一般采用滚动式输出，即每一个间隔后的状态数据会被追加到控制台中。从 MongoDB 3.4 开始增加了 `--interactive` 选项，用来实现非滚动式的监视，非常方便。

```bash
mongostat -h 192.168.65.174 --port 28017 -ufox -pfox --authenticationDatabase=admin --discover --interactive -n 2
```

#### mongotop

**mongotop 命令可用于查看数据库的热点表**，通过观察 mongotop 的输出，**可以判定是哪些集合占用了大部分读写时间**。mongotop 与 mongostat 的实现原理类似，同样**需要 `clusterMonitor` 角色权限**。

```bash
mongotop -h 192.168.65.174 --port=28017 -ufox -pfox --authenticationDatabase=admin
```

默认情况下，mongotop 会持续地每秒输出当前的热点表：

![mongodb-mongotop](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-mongotop.png)


指标说明：

| 指标名 | 说明 |
| ---   | ---  |
| ns | 集合名称空间 |
| total | 花费在该集合上的时长 |
| read | 花费在该集合上的读操作时长 |
| write | 花费在该集合上的写操作时长 |


需要关注的因素主要包括：

- **热点表操作耗费时长是否过高**。这里的时长是在一定的时间间隔内的统计值，它代表某个集合读写操作所耗费的时间总量。在业务高峰期时，核心表的读写操作一般比平时高一些，通过 mongotop 的输出可以对业务尖峰做出一些判断。
- **是否存在非预期的热点表**。一些慢操作导致的性能问题可以从 mongotop 的结果中体现出来。


mongotop 的统计周期、输出总量都是可以设定的：

```bash
# 最多输出 100 次，每次间隔时间为 2s
mongotop -h 192.168.65.174 --port=28017 -ufox -pfox --authenticationDatabase=admin -n 100 2
```

#### Profiler 模块

Profiler 模块可以用来**记录、分析 MongoDB 的详细操作日志**。默认情况下该功能是关闭的，对**某个业务库开启 Profiler 模块之后，符合条件的慢操作日志会被写入该库的 `system.profile` 集合中**。Profiler 的设计很像代码的日志功能，其提供了几种调试级别：

- `0`：日志关闭，无任何输出。
- `1`：部分开启，仅符合条件（时长大于 slowms）的操作日志会被记录。
- `2`：日志全开，所有的操作日志都被记录。

**对当前的数据库开启 Profiler 模块**：

```bash
# 将 level 设置为 2，此时所有的操作会被记录下来。
db.setProfilingLevel(2)
# 检查是否生效
db.getProfilingStatus()
```

```bash
> db.setProfilingLevel(2)
{
    "was" : 0, 
    "slowms" : 100, # 默认 100 ms
    "sampleRate" : 1,
    # ...
}
```

- `slowms` 是慢操作的阈值，单位是毫秒；
- `sampleRate` 表示日志随机采样的比例，1.0 则表示满足条件的全部输出。


**如果希望只记录时长超过 500ms 的操作，则可以将 `level` 设置为 `1`**：

```javascript
db.setProfilingLevel(1,500)
```

进一步设置随机采样的比例：

```javascript
db.setProfilingLevel(1,{slowms:500,sampleRate:0.5})
```

##### 查看操作日志

开启 Profiler 模块之后，可以通过 `system.profile` 集合查看最近发生的操作日志：

```javascript
db.system.profile.find().limit(5).sort({ts:-1}).pretty()
```

查询结果示例：

```javascript
{
    "op" : "insert",
    "ns" : "test.emp",
    "command" : {
      "insert" : "emp",
      "documents" : [
          {
            "_id" : ObjectId("642c23309900000000000000"),
            "name" : "guanyu",
            "age" : 20
          }
      ],
      "ordered" : true,
      lsid: {
        "id" : UUID("00000000-0000-0000-0000-000000000000")
      },
      txnNumber: NumberLong(1),
      // ...
    },   
}    
```

- `op`：操作类型，描述增加、删除、修改、查询。
- `ns`：名称空间，格式为 `{db}.{collection}`。
- `command`：原始的命令文档。
- `cursorid`：游标ID。
- `numYield`：操作数，大于0表示等待锁或者是磁盘I/O操作。
- `nreturned`：返回条目数。
- `keysExamined`：扫描索引条目数，如果比nreturned大出很多，则说明查询效率不高。docsExamined：扫描文档条目数，如果比nreturned大出很多，则说明查询效率不高。
- `locks`：锁占用的情况。
- `storage`：存储引擎层的执行信息。
- `responseLength`：响应数据大小（字节数），一次性查询太多的数据会影响性能，可以使用limit、batchSize进行一些限制。
- `millis`：命令执行的时长，单位是毫秒。
- `planSummary`：查询计划的概要，如IXSCAN表示使用了索引扫描。
- `execStats`：执行过程统计信息。
- `ts`：命令执行的时间点。

`system.profile` 是一个集合，可以像查询普通集合一样加上过滤条件，例如查询 `shop` 库的 `user` 集合的 `update` 操作：

```javascript
db.system.profile.find({op: "update", ns: "shop.user"})
```

{{< callout type="info" >}}
- **`system.profile` 是一个 1MB 的固定大小的集合**，随着记录日志的增多，一些旧的记录会被滚动删除。
- 在线上开启 Profiler 模块需要非常谨慎，这是因为其对 MongoDB 的性能影响比较大。**建议按需部分开启，同时 `slowms` 的值不要设置太低**。
- `sampleRate` 的默认值是 1.0，该字段可以控制记录日志的命令数比例，但只有在 MongoDB 4.0 版本之后才支持。
- **Profiler 模块的设置是内存级的，重启服务器后会自动恢复默认状态**。
{{< /callout >}}

#### db.currentOp()

**Profiler 模块所记录的日志都是已经发生的事情，`db.currentOp()` 命令则与此相反，它可以用来查看数据库当前正在执行的一些操作**。想象一下，当数据库系统的 CPU 发生骤增时，我们最想做的无非是快速找到问题的根源，这时 `db.currentOp` 就派上用场了。

**`db.currentOp()` 读取的是当前数据库的命令快照，该命令可以返回许多有用的信息**，比如：

- 操作的运行时长，快速发现耗时漫长的低效扫描操作。
- 执行计划信息，用于判断是否命中了索引，或者存在锁冲突的情况。
- 操作 ID、时间、客户端等信息，方便定位出产生慢操作的源头。

```bash
> db.currentOP()
{
  "inprog": [
    {
      "type": "op",
      "host": "mongodb1.example.com:27017",
      "desc": "conn12345",
      "connectionId": 12345,
      "client": "192.168.1.100:54216",
      "clientMetadata": {
        "application": {
          "name": "MyApp"
        },
        "driver": {
          "name": "nodejs",
          "version": "4.1.0"
        }
      },
      "active": true,
      "currentOpTime": "2023-05-15T08:42:17.123Z",
      "opid": 678910,
      "secs_running": 5,
      "microsecs_running": NumberLong(5123456),
      "op": "update",
      "ns": "mydb.users",
      "command": {
        "q": {
          "value": {
            "$gt" : 59.32656132664
          }
        },
        "u": {
          "$inc": {
            "value": 82.3154654541
          }
        },
        "multi": true,
        "upsert": true
      },
      "planSummary": "COLLSCAN",
      "locks": {
        "Global": "r",
        "Database": "r",
        "Collection": "r"
      },
      "waitingForLock": false,
      "numYields": 12,
      "lockStats": {
        "Global": {
          "acquireCount": { "r": NumberLong(13) }
        },
        "Database": {
          "acquireCount": { "r": NumberLong(13) }
        },
        "Collection": {
          "acquireCount": { "r": NumberLong(13) }
        }
      }
    },
    # ...
  ],
  # ...
}
```

对示例操作的解读如下:

1. 从 `ns、op` 字段获知，当前进行的操作正在对 `mydb.users` 集合执行 `update` 命令。
2. `command` 字段显示了其原始信息。其中，`command.q` 和 `command.u` 分别展示了 `update` 的查询条件和更新操作。
3. `"planSummary"："COLLSCAN"` 说明情况并不乐观，`update` 没有利用索引而是正在全表扫描。
4. `microsecs_running：NumberLong（5123456）`表示操作运行了 5123ms，注意这里的单位是微秒。

优化方向：

- value 字段**加上索引**。
- 如果更新的数据集非常大，要避免大范围 `update` 操作，**切分成小批量的操作**。

**`opid` 表示当前操作在数据库进程中的唯一编号**。如果已经发现该操作正在导致数据库系统响应缓慢，则可以考虑将其“杀”死：

```javascript
db.killOp(678910)
```

**`db.currentOp` 默认输出当前系统中全部活跃的操作**，由于返回的结果较多，可以指定一些过滤条件：

- 查看等待锁的增加、删除、修改、查询操作

```javascript
db.currentOp({
    waitingForLock:true,
    $or:[
        {op:{$in:["insert","update","remove"]}},
        {"query.findandmodify":{$exists:true}}
    ]
})
```

- 查看执行时间超过 1s 的操作

```javascript
db.currentOp({
    secs_running:{$gt:1}
})
```

- 查看 `test` 数据库中的操作

```javascript
db.currentOp({
    ns:/test/
})
```


currentOp 命令输出说明
- `currentOp.type`：操作类型，可以是 op、idleSession、idleCursor 的一种，一般的操作信息以 op 表示。其为 MongoDB 4.2 版本新增功能。
- `currentOp.host`：主机的名称。
- `currentOp.desc`：连接描述，包含 connectionId。
- `currentOp.connectionId`：客户端连接的标识符。
- `currentOp.client`：客户端主机和端口。
- `currentOp.appName`：应用名称，一般是描述客户端类型。
- `currentOp.clientMetadata`：关于客户端的附加信息，可以包含驱动的版本。
- `currentOp.currentOpTime`：操作的开始时间。MongoDB 3.6 版本新增功能。
- `currentOp.lsid`：会话标识符。MongoDB 3.6 版本新增功能。
- `currentOp.opid`：操作的标志编号。
- `currentOp.active`：操作是否活跃。如果是空闲状态则为 `false`。
- `currentOp.secs_running`：操作持续时间（以秒为单位）。
- `currentOp.microsecs_running`：操作持续时间（以微秒为单位）。
- `currentOp.op`：标识操作类型的字符串。可能的值是："none" "update" "insert" "query" "command" "getmore" "remove" "killcursors"。其中，command 操作包括大多数命令，如 `createIndexes` 和 `findAndModify`。
- `currentOp.ns`：操作目标的集合命名空间。
- `currentOp.command`：操作的完整命令对象的文档。如果文档大小超过 1KB，则会使用一种 `$truncate` 形式表示。
- `currentOp.planSummary`：查询计划的概要信息。
- `currentOp.locks`：当前操作持有锁的类型和模式。
- `currentOp.waitingForLock`：是否正在等待锁。
- `currentOp.numYields`：当前操作执行 yield（让步）的次数。一些锁互斥或者磁盘 I/O 读取都会导致该值大于 0。
- `currentOp.lockStats`：当前操作持有锁的统计。
- `currentOp.lockStats.acquireCount`：操作以指定模式获取锁的次数。
- `currentOp.lockStats.acquireWaitCount`：操作获取锁等待的次数，等待是因为锁处于冲突模式。`acquireWaitCount` 小于或等于 `acquireCount`。
- `currentOp.lockStats.timeAcquiringMicros`：操作为了获取锁所花费的累积时间（以微秒为单位）。`timeAcquiringMicros` 除以 `acquireWaitCount` 可估算出平均锁等待时间。
- `currentOp.lockStats.deadlockCount`：在等待锁获取时，操作遇到死锁的次数。


{{< callout type="info" >}}

- **`db.currentOp` 返回的是数据库命令的瞬时状态**，因此，如果数据库压力不大，则通常只会返回极少的结果。
- 如果启用了复制集，那么 `currentOp` 还会返回一些复制的内部操作（针对 `local.oplog.rs`），需要做一些筛选。
- `db.currentOp` 的结果是一个 BSON 文档，如果大小超过 16MB，则会被压缩。可以使用聚合操作 `$currentOp` 获得完整的结果。
{{< /callout >}}

### 性能问题排查

- [记一次 MongoDB 占用 CPU 过高问题的排查](https://cloud.tencent.com/developer/article/1495820)
- [MongoDB线上案例：一个参数提升16倍写入速度](https://cloud.tencent.com/developer/article/1857119)