---
title: 索引
weight: 1
---

**MongoDB 索引数据结构是 B-Tree 还是 B+Tree？**

先说结论，是 **B+Tree**。

为什么会产生这个问题呢？

起源于早期的官方文档的一句话 `MongoDB indexes use a B-tree data structure.`，然后就导致了分歧：有人说 MongoDB 索引数据结构使用的是 B-Tree,有的人又说是 B+Tree。

MongoDB 从 3.2 开始就默认使用 WiredTiger 作为存储引擎，所以 MongoDB 内部存储的数据结构由 WiredTiger 决定。而 **WiredTiger 官方文档明确说了底层用的 B+Tree**。

> WiredTiger 官方文档：https://source.wiredtiger.com/3.0.0/tune_page_size_and_comp.html
>
> WiredTiger maintains a table's data in memory using a data structure called a **B-Tree ( B+ Tree to be specific)**, referring to the nodes of a B-Tree as pages. **Internal pages carry only keys. The leaf pages store both keys and values**.


## 索引操作

### 创建索引

```javascript
db.collection.createIndex(keys, options)
```

- `keys`：指定要创建索引的字段，**可以是单个字段或多个字段**。如果是多个字段，那么就会创建一个复合索引。**1 按升序创建索引， -1  按降序创建索引**。
- `options`：可选参数。

可选参数列表：

| 参数名         | 类型 | 描述                                                         |
| -------------- | ---- | ------------------------------------------------------------ |
| `background`    | 布尔值 | 建索引过程会阻塞其它数据库操作，`background` 可指定以后台方式创建索引。 `background` 默认值为 `false`。                   |
| `unique`         | 布尔值 | 是否创建唯一索引，默认为 `false`。                       |
| `name`           | 字符串 | 索引的名称，如果未指定，MongoDB 的通过连接索引的字段名和排序顺序生成一个索引名称。                              |
| `sparse`        | 布尔值 | 对文档中不存在的字段数据不启用索引；这个参数需要特别注意，如果设置为 `true` 的话，在索引字段中不会查询出不包含对应字段的文档。默认值为 `false`。                    |
| `expireAfterSeconds` | 数字 | 指定一个以秒为单位的数值，完成 TTL 设定，设定集合的生存时间。                             |
| `v`            | 数字 | 指定索引的版本号，默认的索引版本取决于 mongod 创建索引时运行的版本。                         |
| `dropDups`        | 布尔值 | 3.0 版本已废弃。在建立唯一索引时是否删除重复记录，指定 `true` 创建唯一索引。默认值为 `false`。                   |
| `weights`        | 对象 | 索引权重值，数值在 1 到 99,999 之间，表示该索引相对于其他索引字段的得分权重。                     |
| `default_language` | 字符串 | 对于文本索引，该参数决定了停用词及词干和词器的规则的列表。 默认为英语                 |
| `language_override `| 字符串 | 对于文本索引，该参数指定了包含在文档中的字段名，语言覆盖默认的language，默认值为 language。                 |

> 3.0.0 版本前创建索引方法为 `db.collection.ensureIndex()`。

```javascript
// 创建索引后台执行
db.values.createIndex({open: 1, close: 1}, {background: true})
// 创建唯一索引
db.values.createIndex({title:1},{unique:true})
```

### 查看索引

```javascript
// 查看索引信息
db.books.getIndexes()
// 查看索引键
db.books.getIndexKeys()
```

### 删除索引

```javascript
// 删除索引
db.collection.dropIndex(indexName)
// 删除所有索引，不能删除主键索引
db.collection.dropIndexes()
```

## 索引类型

MongoDB 支持各种丰富的索引类型，包括**单键索引、复合索引，唯一索引**等一些常用的结构。由于采用了灵活可变的文档类型，因此它也同样**支持对嵌套字段、数组进行索引**。通过建立合适的索引，可以极大地提升数据的检索速度。在一些特殊应用场景，MongoDB 还支持地理空间索引、文本检索索引、TTL 索引等不同的特性。

### 单键索引（Single Field Indexes）

**在某一个特定的字段上建立索引**。MongoDB 在 ID 上建立了唯一的单键索引,所以经常会使用 id 来进行查询； 在索引字段上进行精确匹配、排序以及范围查找都会使用此索引。

```javascript
db.books.createIndex({title:1}) // 升序索引
// 对内嵌文档的字段进行索引
db.books.createIndex({"author.name":1})
```

### 复合索引（Compound Index）

**复合索引是多个字段组合而成的索引**，其性质和单字段索引类似。但不同的是，**复合索引中字段的顺序、字段的升降序对查询性能有直接的影响**，因此在设计复合索引时则需要考虑不同的查询场景。

```javascript
db.books.createIndex({type:1,favCount:1})
// 查看执行计划
db.books.find({type:"novel",favCount:{$gt:50}}).explain()
```

### 多键(数组)索引（Multikey Index）

**在数组的属性上建立索引**。针对这个数组的任意值的查询都会定位到这个文档,既多个索引入口或者键值引用同一个文档。

准备数据：

```javascript
db.inventory.insertMany([
    { _id: 5, type: "food", item: "aaa", ratings: [ 5, 8, 9 ] },
    { _id: 6, type: "food", item: "bbb", ratings: [ 5, 9 ] },
    { _id: 7, type: "food", item: "ccc", ratings: [ 9, 5, 8 ] },
    { _id: 8, type: "food", item: "ddd", ratings: [ 9, 5 ] },
    { _id: 9, type: "food", item: "eee", ratings: [ 5, 9, 5 ] }
])

// 创建多键索引
db.inventory.createIndex( { ratings: 1 } )
```

{{< callout type="info" >}}
多键索引很容易与复合索引产生混淆，**复合索引是多个字段的组合**，而**多键索引则仅仅是在一个字段上出现了多键（multi key）**。而实质上，**多键索引也可以出现在复合字段上**。
{{< /callout >}}

```javascript
// 创建复合多值索引
db.inventory.createIndex( { item:1,ratings: 1 } )
```

{{< callout type="warn" >}}
MongoDB 并不支持一个复合索引中同时出现多个数组字段
{{< /callout >}}

**在包含嵌套对象的数组字段上创建多键索引**：

```javascript
db.inventory.insertMany([
{
  _id: 1,
  item: "abc",
  stock: [
    { size: "S", color: "red", quantity: 25 },
    { size: "S", color: "blue", quantity: 10 },
    { size: "M", color: "blue", quantity: 50 }
  ]
},
{
  _id: 2,
  item: "def",
  stock: [
    { size: "S", color: "blue", quantity: 20 },
    { size: "M", color: "blue", quantity: 5 },
    { size: "M", color: "black", quantity: 10 },
    { size: "L", color: "red", quantity: 2 }
  ]
},
{
  _id: 3,
  item: "ijk",
  stock: [
    { size: "M", color: "blue", quantity: 15 },
    { size: "L", color: "blue", quantity: 100 },
    { size: "L", color: "red", quantity: 25 }
  ]
}
])

// 在包含嵌套对象的数组字段上创建多键索引
db.inventory.createIndex( { "stock.size": 1, "stock.quantity": 1 } )

db.inventory.find({"stock.size":"S","stock.quantity":{$gt:20}}).explain()
```

### Hash 索引（Hashed Index）

不同于传统的 BTree 索引,哈希索引**使用 `hash` 函数来创建索引**。在索引字段上进行**精确匹配**，但**不支持范围查询**，**不支持多键 hash**，Hash 索引上的入口是均匀分布的，**在分片集合中非常有用**。

```javascript
db.users.createIndex({username : 'hashed'})
```

对于用户名，电话号码之类的字段，可以使用哈希索引。**哈希索引的主要优势在于数据分布更均匀**。

### 地理空间索引（Geospatial Index）

MongoDB 为地理空间检索提供了非常方便的功能。**地理空间索引（2dsphereindex）就是专门用于实现位置检索的一种特殊索引**。

#### 案例

如何实现“查询附近商家"？

假设商家的数据模型如下：

```javascript
db.restaurant.insert({
    restaurantId: 0,
    restaurantName:"兰州牛肉面",
    location : {
        type: "Point",
        coordinates: [ -73.97, 40.77 ]
    }
})
```

创建一个 2dsphere 索引：

```javascript
db.restaurant.createIndex({location : "2dsphere"})
```

查询附近 10000 米商家信息：

```javascript
db.restaurant.find( { 
    location:{ 
        $near :{
            $geometry :{ 
                type : "Point" ,
                coordinates : [ -73.88, 40.78 ] 
            } ,
            $maxDistance: 10000 
        } 
    } 
})
```

- `$near`：查询操作符，用于实现附近商家的检索，返回数据结果会按距离排序。
- `$geometry`：操作符用于指定一个 **GeoJSON 格式的地理空间对象**，`type=Point` 表示地理坐标点，`coordinates` 则是用户当前所在的**经纬度位置**；
- `$maxDistance`：限定了**最大距离**，单位是**米**。

### 全文索引（Text Indexes）

MongoDB 支持**全文检索**功能，可通过建立文本索引来实现**简易的分词检索**。

```javascript
db.reviews.createIndex( { comments: "text" } )
```

**`$text` 操作符可以在有 `text index` 的集合上执行文本检索**。`$text` 将会使用空格和标点符号作为分隔符对检索字符串进行分词，并且对检索字符串中所有的分词结果进行一个逻辑上的 OR 操作。

全文索引能**解决快速文本查找的需求**，比如有一个博客文章集合，需要根据博客的内容来快速查找，则可以针对博客内容建立文本索引。

#### 案例

准备数据：

```javascript
db.stores.insertMany(
   [
     { _id: 1, name: "Java Hut", description: "Coffee and cakes" },
     { _id: 2, name: "Burger Buns", description: "Gourmet hamburgers" },
     { _id: 3, name: "Coffee Shop", description: "Just coffee" },
     { _id: 4, name: "Clothes Clothes Clothes", description: "Discount clothing" },
     { _id: 5, name: "Java Shopping", description: "Indonesian goods" }
   ]
)

// 创建 name 和 description 的全文索引
db.stores.createIndex({name: "text", description: "text"})
```

通过 `$text` 操作符来查寻数据中所有包含 “coffee”,”shop”，“java” 列表中任何词语的商店：

```javascript
db.stores.find({$text: {$search: "java coffee shop"}})
```

MongoDB 的文本索引功能存在诸多限制，而官方**并未提供中文分词的功能**，这使得该功能的应用场景十分受限。

### 通配符索引（Wildcard Indexes）

MongoDB 的文档模式是动态变化的，而**通配符索引可以建立在一些不可预知的字段上**，以此实现查询的加速。MongoDB 4.2 引入了通配符索引来支持对未知或任意字段的查询。

#### 案例

准备商品数据，不同商品属性不一样：

```javascript
db.products.insert([
    {
      "product_name" : "Spy Coat",
      "product_attributes" : { // 每个商品的 product_attributes 中包含的属性是不一样的
        "material" : [ "Tweed", "Wool", "Leather" ],
        "size" : {
          "length" : 72,
          "units" : "inches"
        }
      }
    },
    {
      "product_name" : "Spy Pen",
      "product_attributes" : {
         "colors" : [ "Blue", "Black" ],
         "secret_feature" : {
           "name" : "laser",
           "power" : "1000",
           "units" : "watts",
         }
      }
    },
    {
      "product_name" : "Spy Book"
    }
])

// 创建通配符索引
// $** 表示任意字段
db.products.createIndex( { "product_attributes.$**" : 1 } )
```

通配符索引可以支持任意单字段查询 `product_attributes` 或其嵌入字段：

```javascript
db.products.find( { "product_attributes.size.length": { $gt : 60 } } )
db.products.find( { "product_attributes.material": "Leather" } )
db.products.find( { "product_attributes.secret_feature.name": "laser" } )
```

注意：

1. 通配符索引不兼容的索引类型或属性：
   - Compound
   - TTL
   - Hashed
   - 2dsphere
   - Text
   - Unique

2. 通配符索引不能支持查询字段不存在的文档。因为通配符索引是稀疏的，**稀疏索引只索引集合中存在该字段的文档，对于缺少该字段或该字段值为 `null` 的文档，索引中不会有对应的条目**。

```javascript
// 通配符索引不能支持以下查询
db.products.find( {"product_attributes" : { $exists : false } } )
db.products.aggregate([
  { $match : { "product_attributes" : { $exists : false } } }
])
```

3. 通配符索引会为文档或数组的内容生成条目，而**是文档或数组本身**。因此**通配符索引不能支持精确的文档/数组相等匹配**。通配符索引可以支持查询字段等于空文档 `{}` 的情况。

```javascript
// 精确匹配查数组中的其中一个条目，这条查询是支持的
db.products.find({ "product_attributes.colors" : "Blue" } )

// 通配符索引不能支持以下查询

// 匹配数组中的所有条目，这条查询是不支持的
db.products.find({ "product_attributes.colors" : [ "Blue", "Black" ] } )

db.products.aggregate([{
  $match : { "product_attributes.colors" : [ "Blue", "Black" ] } 
}])
```


## 索引属性

### 唯一索引（Unique Indexes）

在现实场景中，唯一性是很常见的一种索引约束需求，重复的数据记录会带来许多处理上的麻烦，比如订单的编号、用户的登录名等。通过建立唯一性索引，可以保证集合中文档的指定字段拥有唯一值。

```javascript
// 创建唯一索引
db.values.createIndex({title:1},{unique:true})
// 复合索引支持唯一性约束
db.values.createIndex({title:1，type:1},{unique:true})
// 多键索引支持唯一性约束
db.inventory.createIndex( { ratings: 1 },{unique:true} )
```

- **唯一性索引对于文档中缺失的字段，会使用 `null` 值代替**，因此**不允许存在多个文档缺失索引字段的情况**。
- **对于分片的集合，唯一性约束必须匹配分片规则**。换句话说，**为了保证全局的唯一性，分片键必须作为唯一性索引的前缀字段**。

### 部分索引（Partial Indexes）

**部分索引仅包含集合中满足指定过滤条件的文档，其他文档不会被包含在索引中**。部分索引具有更低的存储需求和更低的索引创建和维护的性能成本。3.2 新版功能。

部分索引提供了稀疏索引功能的超集，应该优先于稀疏索引。

例如：

```javascript
db.restaurants.createIndex(
   { cuisine: 1, name: 1 }, // 索引字段
   { partialFilterExpression: { rating: { $gt: 5 } } } // 增加过滤条件，只对符合条件的文档才会加索引
)
```

`partialFilterExpression` 选项接受指定过滤条件的文档:

- `$eq`：等于
- `$gt`：大于
- `$gte`：大于等于
- `$lt`：小于
- `$lte`：小于等于
- `$type`：类型
- 顶层的 `$and`

```javascript
// 符合条件，使用索引
db.restaurants.find( { cuisine: "Italian", rating: { $gte: 8 } } )
// 不符合条件，不能使用索引
db.restaurants.find( { cuisine: "Italian" } )
```

#### 案例

准备数据：

```javascript
db.restaurants.insert({
   "_id" : ObjectId("5641f6a7522545bc535b5dc9"),
   "address" : {
      "building" : "1007",
      "coord" : [
         -73.856077,
         40.848447
      ],
      "street" : "Morris Park Ave",
      "zipcode" : "10462"
   },
   "borough" : "Bronx",
   "cuisine" : "Bakery",
   "rating" : { "date" : ISODate("2014-03-03T00:00:00Z"),
                "grade" : "A",
                "score" : 2
              },
   "name" : "Morris Park Bake Shop",
   "restaurant_id" : "30075445"
})

// 创建索引
db.restaurants.createIndex(
   { borough: 1, cuisine: 1 },
   { partialFilterExpression: { 'rating.grade': { $eq: "A" } } }
)
```

查询：

```javascript
// 符合条件，使用索引
db.restaurants.find( { borough: "Bronx", 'rating.grade': "A" } )
// 不符合条件，不能使用索引
db.restaurants.find( { borough: "Bronx", cuisine: "Bakery" } )
```

#### 唯一约束结合部分索引使用导致唯一约束失效的问题

**如果同时指定了 `partialFilterExpression` 和唯一约束，那么唯一约束只适用于满足筛选器表达式的文档**。如果文档不满足筛选条件，那么带有惟一约束的部分索引不会阻止插入不满足惟一约束的文档。

准备数据：

```javascript
db.users.insertMany( [
   { username: "david", age: 29 },
   { username: "amanda", age: 35 },
   { username: "rajiv", age: 57 }
] )

// 创建索引，指定 username 字段
// 唯一约束
// 部分过滤器表达式 age: {$gte: 21}，表示只有年龄大于等于 21 的文档才会有索引
// 因此，只有年龄大于等于 21 的文档并且 username 是唯一的，才会有索引
db.users.createIndex(
   { username: 1 },
   { unique: true, partialFilterExpression: { age: { $gte: 21 } } }
)
```

索引防止了以下文档的插入，因为文档 username 都已经存在，虽然满足了年龄字段大于 21 的条件:

```javascript
db.users.insertMany( [
   { username: "david", age: 27 },
   { username: "amanda", age: 25 },
   { username: "rajiv", age: 32 }
])
```

但是，以下具有重复用户名的文档是允许的，因为唯一约束只适用于年龄大于或等于 21 岁的文档。也就是说下面的数据不满足上面的索引的条件，所以不会受到索引的约束：

```javascript
db.users.insertMany( [
   { username: "david", age: 20 },
   { username: "amanda" },
   { username: "rajiv", age: null }
])
```

### 稀疏索引（Sparse Indexes）

索引的稀疏属性确保**索引只包含具有索引字段的文档的条目，索引将跳过没有索引字段的文档**。

特性：**只对存在字段的文档进行索引（包括字段值为 `null` 的文档）**

```javascript
// 不索引不包含 xmpp_id 字段的文档
db.addresses.createIndex( { "xmpp_id": 1 }, { sparse: true } )
```

{{< callout type="info" >}}
**如果稀疏索引会导致查询和排序操作的结果集不完整，MongoDB 将不会使用该索引**，除非使用 `hint()` 明确指定索引。
{{< /callout >}}


#### 案例

准备数据：

```javascript
db.scores.insertMany([
    {"userid" : "newbie"},
    {"userid" : "abby", "score" : 82},
    {"userid" : "nina", "score" : 90}
])

// 创建稀疏索引
db.scores.createIndex( { score: 1 } , { sparse: true } )
```

查询：

```javascript
// 使用稀疏索引
db.scores.find( { score: { $lt: 90 } } )

// 即使排序是通过索引字段，MongoDB 也不会选择稀疏索引来完成查询，以返回完整的结果
db.scores.find().sort( { score: -1 } )

// 要使用稀疏索引，使用 hint() 显式指定索引
db.scores.find().sort( { score: -1 } ).hint( { score: 1 } )
```

#### 同时具有稀疏性和唯一性的索引

同时具有稀疏性和唯一性的索引**可以防止集合中存在字段值重复的文档，但允许不包含此索引字段的文档插入**。

```javascript
db.scores.dropIndex({score:1})
// 创建具有唯一约束的稀疏索引
db.scores.createIndex( { score: 1 } , { sparse: true, unique: true } )
```

这个索引将允许插入具有唯一的分数字段值或不包含分数字段的文档。因此，给定 `scores` 集合中的现有文档，索引允许以下插入操作:

```javascript
db.scores.insertMany( [
   { "userid": "AAAAAAA", "score": 50 },
   { "userid": "BBBBBBB", "score": 64 },
   { "userid": "CCCCCCC" },
   { "userid": "CCCCCCC" }
] )
```

索引不允许添加下列文件，因为已经存在评分为 50 的文档：

```javascript
db.scores.insertMany( [
   { "userid": "DDDDDDD", "score": 50 },
])
```

### TTL 索引（TTL Indexes）

在一般的应用系统中，并非所有的数据都需要永久存储。例如一些系统事件、用户消息等，这些数据随着时间的推移，其重要程度逐渐降低。更重要的是，存储这些大量的历史数据需要花费较高的成本，因此项目中通常会**对过期且不再使用的数据进行老化处理**。

通常有两种方案：

1. **为每个数据记录一个时间戳**，应用侧开启一个定时器，按时间戳定期删除过期的数据。
2. 数据**按日期进行分表，同一天的数据归档到同一张表**，同样使用定时器删除过期的表。

对于数据老化，MongoDB 提供了一种更加便捷的做法：TTL（Time To Live）索引。T**TL 索引需要声明在一个日期类型的字段中，TTL 索引是特殊的单字段索引**，MongoDB 可以使用它**在一定时间或特定时钟时间后自动从集合中删除文档**。

```javascript
// 创建 TTL 索引，TTL 值为 3600 秒
db.eventlog.createIndex( { "lastModifiedDate": 1 }, { expireAfterSeconds: 3600 } )
```

对集合创建 TTL 索引之后，MongoDB 会在周期性运行的后台线程中对该集合进行检查及数据清理工作。除了数据老化功能，**TTL 索引具有普通索引的功能**，同样可以用于加速数据的查询。

**TTL 索引不保证过期数据会在过期后立即被删除**。文档过期和 MongoDB 从数据库中删除文档的时间之间可能存在延迟。**删除过期文档的后台任务每 60 秒运行一次**。因此，在文档到期和后台任务运行之间的时间段内，文档可能会保留在集合中。

#### 案例

准备数据：

```javascript
db.log_events.insertOne( {
   "createdAt": new Date(),
   "logEvent": 2,
   "logMessage": "Success!"
})

// 创建 TTL 索引
db.log_events.createIndex( { "createdAt": 1 }, { expireAfterSeconds: 20 } )
```

#### 可变的过期时间

TTL 索引在创建之后，仍然**可以对过期时间进行修改**。这需要使用 `collMod` 命令对索引的定义进行变更：

```javascript
db.runCommand({collMod:"log_events",index:{keyPattern:{createdAt:1},expireAfterSeconds:600}})
```

#### TTL 索引的限制

TTL 索引的确可以减少开发的工作量，而且通过数据库自动清理的方式会更加高效、可靠，但是在使用 TTL 索引时需要注意以下的限制：
- **TTL 索引只能支持单个字段，并且必须是非 `_id` 字段**。
- **TTL 索引不能用于固定集合**。 
- **TTL 索引无法保证及时的数据老化**，MongoDB 会通过后台的 TTL Monitor 定时器来清理老化数据，默认的间隔时间是 1 分钟。当然如果在数据库负载过高的情况下，TTL 的行为则会进一步受到影响。
- **TTL 索引对于数据的清理仅仅使用了 `remove` 命令，这种方式并不是很高效**。因此 TTL Monitor 在运行期间对系统 CPU、磁盘都会造成一定的压力。**相比之下，按日期分表的方式操作会更加高效**。

### 隐藏索引（Hidden Indexes）

**隐藏索引对查询规划器不可见，不能用于支持查询**。通过对规划器隐藏索引，用户可以在不实际删除索引的情况下评估删除索引的潜在影响。如果影响是负面的，用户可以取消隐藏索引，而不必重新创建已删除的索引。4.4 新版功能。

```javascript
// 创建隐藏索引
db.restaurants.createIndex({ borough: 1 },{ hidden: true });
// 隐藏现有索引
db.restaurants.hideIndex( { borough: 1} );
db.restaurants.hideIndex( "索引名称" )
// 取消隐藏索引
db.restaurants.unhideIndex( { borough: 1} );
db.restaurants.unhideIndex( "索引名
```

#### 案例

```javascript
db.scores.insertMany([
    {"userid" : "newbie"},
    {"userid" : "abby", "score" : 82},
    {"userid" : "nina", "score" : 90}
])

// 创建隐藏索引
db.scores.createIndex(
   { userid: 1 },
   { hidden: true }
)

// 查看索引信息
db.scores.getIndexes()

[
    // ...
    {
        "v": 2,
        "hidden": true,
        "key": {
            "userid": 1
        }
        // ...
    }
]
```

查询：

```javascript
// 不使用索引
db.scores.find({userid:"abby"}).explain()

// 取消隐藏索引
db.scores.unhideIndex( { userid: 1} )
// 使用索引
db.scores.find({userid:"abby"}).explain()
```

## 索引实践

### 为每一个查询建立合适的索引

这个是针对于数据量较大比如说超过几十上百万（文档数目）数量级的集合。如果没有索引 MongoDB 需要把所有的 Document 从盘上读到内存，这会对 MongoDB 服务器造成较大的压力并影响到其他请求的执行。

### 创建合适的复合索引，不要依赖于交叉索引

如果你的查询会使用到多个字段，MongoDB 有两个索引技术可以使用：交叉索引和复合索引。**交叉索引就是针对每个字段单独建立一个单字段索引，然后在查询执行时候使用相应的单字段索引进行索引交叉而得到查询结果**。交叉索引目前触发率较低，所以如果你有一个多字段查询的时候，**建议使用复合索引能够保证索引正常的使用**。

```javascript
// 查找所有年龄小于30岁的深圳市马拉松运动员
db.athelets.find({sport: "marathon", location: "sz", age: {$lt: 30}}})
// 创建复合索引
db.athelets.createIndex({sport:1, location:1, age:1})
```

### 复合索引字段顺序

复合索引字段顺序：**匹配条件在前，范围条件在后（Equality First, Range After）**。

前面的例子，在创建复合索引时如果条件有匹配和范围之分，那么匹配条件 `(sport: "marathon")` 应该在复合索引的前面。范围条件 `(age: < 30)` 字段应该放在复合索引的后面。

### 尽可能使用覆盖索引（Covered Index）

建议只返回需要的字段，同时，利用覆盖索引来提升性能。

### 建索引要在后台运行

在对一个**集合创建索引时，该集合所在的数据库将不接受其他读写操作**。对大数据量的集合建索引，建议使用后台运行选项 `{background: true}`。

### 避免设计过长的数组索引

**数组索引是多值的，在存储时需要使用更多的空间**。如果索引的数组长度特别长，或者数组的增长不受控制，则可能导致索引空间急剧膨胀。

## Explain 执行计划

需要关心的问题：

- 查询是否使用了索引
- 索引是否减少了扫描的记录数量
- 是否存在低效的内存排序

MongoDB 提供了 `explain` 命令，它可以评估指定查询模型（querymodel）的执行计划，根据实际情况进行调整，然后提高查询效率。

```javascript
db.collection.find().explain(<verbose>)
```

- `verbose` 可选参数，表示执行计划的输出模式，默认 `queryPlanner`。
  - `queryPlanner`：执行计划的详细信息，包括查询计划、集合信息、查询条件、最佳执行计划、查询方式和 MongoDB 服务信息等。
  - `exectionStats`：最佳执行计划的执行情况和被拒绝的计划等信息。
  - `allPlansExecution`：选择并执行最佳执行计划，并返回最佳执行计划和其他执行计划的执行情况。

### queryPlanner

```javascript
// title 字段无索引
db.books.find({title:"book-1"}).explain("queryPlanner")
```

输出结果：

```javascript
{
   "queryPlanner" : {
      "plannerVersion" : 1,
      "namespace" : "test.books",
      "indexFilterSet" : false,
      "parsedQuery" : {
         "title" : {
            "$eq" : "book-1"
         }
      },
      "winningPlan" : {
         "stage" : "COLLSCAN", // 全表扫描
         "direction" : "forward"
      },
      "rejectedPlans" : []
   }
   // ...
}
```


| 字段名称  |  描述 | 
| --- | --- |
| plannerVersion | 执行计划的版本 |
| namespace | 查询的集合 |
| **indexFilterSet** | 是否使用索引 |
| parsedQuery | 查询条件 |
| **winningPlan** | 最佳执行计划 |
| **stage** | 查询方式 |
| filter | 过滤条件 | 
| direction | 查询顺序 |
| rejectedPlans | 拒绝的执行计划 |
| serverInfo | mongodb 服务器信息 |


### exectionStats

`executionStats` 模式的返回信息中包含了 `queryPlanner` 模式的所有字段，并且还**包含了最佳执行计划的执行情况**。

```javascript
// 创建索引
db.books.createIndex({title:1})

db.books.find({title:"book-1"}).explain("executionStats")
```

运行结果：

```javascript
{
   "queryPlanner" : {
      // ...
      "winningPlan" : {
         "stage" : "FETCH", // 使用索引
         "inputStage" : {
            "stage" : "IXSCAN", // 索引扫描
            "keyPattern" : {
               "title" : 1
            },
            "indexName" : "title_1",
            "isMultiKey" : false,
            "multiKeyPaths" : {
               "title" : [ ]
            },
            "isUnique" : false,
            "isSparse" : false,
            "isPartial" : false,
            "indexVersion" : 2,
            "direction" : "forward",
            "indexBounds" : {
               "title" : [
                  "[\"book-1\", \"book-1\"]"
               ]
            }
         }
      },
      "rejectedPlans" : [ ]
   },
   "executionStats" : {
      "executionSuccess" : true,
      "nReturned" : 1, // 返回记录数
      "executionTimeMillis" : 1, // 执行时间
      "totalKeysExamined" : 1, // 索引扫描的次数
      "totalDocsExamined" : 1, // 扫描的文档数
      // ...
   }
}
```

| 字段名称 | 描述 |
| --- | --- |
| `winningPlan.inputStage` | 用来描述子 stage，并且为其父 stage 提供文档和索引关键字 |
| `winningPlan.inputStage.stage` | 子查询方式 |
| `winningPlan.inputStage.keyPattern` | 所扫描的 index 内容 |
| `winningPlan.inputStage.indexName` | 索引名称 |
| `winningPlan.inputStage.isMultiKey` | 是否是多键索引。如果索引建立在 array 上，则是 `true` |
| `executionStats.executionSuccess` | 是否执行成功 |
| `executionStats.nReturned` | 返回记录数 |
| `executionStats.executionTimeMillis` | 语句执行时间 |
| `executionStats.executionStages.executionTimeMillisEstimate` | 检索文档获取数据的时间 |
| `executionStats.totalKeysExamined` | 索引扫描的次数 |
| `executionStats.totalDocsExamined` | 扫描的文档数 |
| `executionStats.executionStages.isEOF` | 是否到达 steam 结尾，1 或者 true 代表已到达结尾 |
| `executionStats.executionStages.works` | 工作单元数，一个查询会分解成小的工作单元 |
| `executionStats.executionStages.advanced` | 优先返回的结果数 |
| `executionStats.executionStages.docsExamined` | 文档检查数 |

### allPlansExecution

`allPlansExecution` 返回的信息包含 `executionStats` 模式的内容，且包含 `allPlansExecution:[]` 块

```javascript
"allPlansExecution" : [
      {
         "nReturned" : <int>,
         "executionTimeMillisEstimate" : <int>,
         "totalKeysExamined" : <int>,
         "totalDocsExamined" :<int>,
         "executionStages" : {
            "stage" : <STAGEA>,
            "nReturned" : <int>,
            "executionTimeMillisEstimate" : <int>,
            ...
            }
         }
      },
      ...
   ]
```

### stage 状态

| 状态 | 描述 |
| --- | --- |
| COLLSCAN | 全表扫描 |
| IXSCAN | 索引扫描 |
| FETCH | 根据索引检索指定文档 |
| SHARD_MERGE | 将各个分片返回数据进行合并 |
| SORT | 在内存中进行了排序 |
| LIMIT | 使用 limit 限制返回数 |
| SKIP | 使用 skip 进行跳过 |
| IDHACK | 对 `_id` 进行查询 |
| SHARDING_FILTER | 通过 mongos 对分片数据进行查询 |
| COUNTSCAN | count 不使用 Index 进行 count 时的 stage 返回 |
| COUNT_SCAN | count 使用了 Index 进行 count 时的 stage 返回 |
| SUBPLA | 未使用到索引的 `$or` 查询的 stage 返回 |
| TEXT | 使用全文索引进行查询时候的 stage 返回 |
| PROJECTION | 限定返回字段时候 stage 的返回 |

执行计划的返回结果中尽量不要出现以下 stage:

- COLLSCAN (全表扫描)
- SORT (使用 sort 但是无 index)
- 不合理的 SKIP
- SUBPLA (未用到 index 的 `$or`)
- COUNTSCAN (不使用 index 进行 count)