---
title: 快速入门
weight: 1
---

## 文档操作

SQL to MongoDB Mapping Chart ：https://www.mongodb.com/docs/manual/reference/sql-comparison/

### 插入文档

- `db.collection.insertOne()`：将单个文档插入到集合中。
- `db.collection.insertMany()`：将多个文档插入到集合中。

> **注意**：`save` 和 `insert` 方法已被弃用，建议使用 `insertOne`、`insertMany` 或者 `buldWrite` 方法。

```javascript
db.collection.insertOne(
   <document>,
   {
      writeConcern: <document>
   }
)

db.emps.insertOne(
   { name: "fox", age: 35 }
)

// 设置 writeConcern 参数
db.emps.insertOne(
   { name: "fox", age: 35},
   {
      writeConcern: { w: "majority", j: true, wtimeout: 5000 }
   }
)
```

`writeConcern` 是 MongoDB 中用来控制**写入确认**的选项。以下是 `writeConcern` 参数的一些常见选项：

- `w`：指定**写入确认级别**。如果指定为**数字，则表示要等待写入操作完成的节点数**。如果指定为 **`majority`，则表示等待大多数节点完成写入操作**。默认为 1，表示等待写入操作完成的节点数为 1。
- `j`：表示写入操作**是否要求持久化到磁盘**。如果设置为 **`true`，则表示写入操作必须持久化到磁盘后才返回成功**。如果设置为 `false`，则表示写入操作可能在数据被持久化到磁盘之前返回成功。默认为 `false`。
- `wtimeout`：表示**等待写入操作完成的超时时间**，单位为毫秒。**如果超过指定的时间仍然没有返回确认信息，则返回错误**。默认为 0，表示不设置超时时间。

```javascript
// 插入多条文档数据
db.collection.insertMany(
   [ <document 1> , <document 2>, ... ],
   {
      writeConcern: <document>,
      ordered: <boolean>      
   }
)

db.emps.insertMany(
   [ { name: "fox", age: 35 }, { name: "cat", age: 35 } ]
)
```

- `ordered`：指定**是否按顺序写入**，默认 `true`，按顺序写入。

```javascript
var tags = ["nosql","mongodb","document","developer","popular"];
var types = ["technology","sociality","travel","novel","literature"];
var books=[];
for(var i=0;i<50;i++){
    var typeIdx = Math.floor(Math.random()*types.length);
    var tagIdx = Math.floor(Math.random()*tags.length);
    var favCount = Math.floor(Math.random()*100);
    var book = {
        title: "book-"+i,
        type: types[typeIdx],
        tag: tags[tagIdx],
        favCount: favCount,
        author: "xxx"+i
    };
    books.push(book)
}
db.books.insertMany(books);
```

进入 mongosh，执行：

```bash
> load("books.js")
true
> db.bools.countDocuments()
50
```

### 查询文档

#### 查询多个文档

```javascript
db.collection.find(query, projection)
```

- `query`：可选，使用查询操作符**指定查询条件**。
- `projection`：可选，使用投影操作符**指定返回的键**。查询时返回文档中所有键值，只需省略该参数即可（默认省略）。投影时，`_id` 为 1 的时候，其他字段必须是 1；`_id` 是 0 的时候，其他字段可以是 0；如果没有 `_id` 字段约束，多个其他字段必须同为 0 或同为 1。


如果查询返回的条目数量较多，mongosh 则会自动实现分批显示。默认情况下每次只显示 20 条，可以输入 it 命令读取下一批。


#### 查询集合中的第一个文档

```bash
db.collection.findOne(query, projection)

> db.books.findOne()
{
  _id: ObjectId("6470393633826370200"),
  title: 'book-0',
  type: 'technology',
  tag: 'nosql',
  favCount: 7,
  author: 'xxx0'
}
```

#### pretty

`pretty()` 方法以格式化的方式来显示所有文档：

```javascript
db.collection.find().pretty()
```

#### 条件查询

**查询条件对照表**：

| SQL | MQL |
| --- | --- |
| `a = 1` | `{a: 1}` |
| `a <> 1` | `{a: {$ne: 1}}` |
| `a > 1` | `{a: {$gt: 1}}` |
| `a >= 1` | `{a: {$gte: 1}}` |
| `a < 1` | `{a: {$lt: 1}}` |
| `a <= 1` | `{a: {$lte: 1}}` | 

**查询逻辑对照表**：

| SQL | MQL |
| --- | --- |
| `a = 1 AND b = 2` | `{a: 1, b: 2}` 或者 `{$and: [{a: 1}, {b: 1}]}` |
| `a = 1 OR b = 2` | `{$or: [{a: 1}, {b: 2}]}` |
| `a IS NULL` | `{a: {$exists: false}}` |
| `a IS NOT NULL` | `{a: {$exists: true}}` |
| `a IN (1, 2, 3)` | `{a: {$in: [1, 2, 3]}}` |
| `a NOT IN (1, 2, 3)` | `{a: {$nin: [1, 2, 3]}}` |


```javascript
// 查询带有 nosql 标签的 book 文档：
db.books.find({tag:"nosql"})
// 按照 id 查询单个 book 文档：
db.books.find({_id:ObjectId("61caa09ee0782536660494d9")})
// 查询分类为 “travel”，收藏数超过 60 个的 book 文档：
db.books.find({type:"travel",favCount:{$gt:60}})
```

#### 正则表达式匹配查询

使用 `$regex` 操作符来设置匹配字符串的正则表达式：

```javascript
// 使用正则表达式查找 type 包含 so 字符串的 book
db.books.find({type:{$regex:"so"}})
// 或者
db.books.find({type:/so/})
```

#### 排序

使用 `sort()` 方法对数据进行排序：

```javascript
// 指定按收藏数（favCount）降序返回
db.books.find({type:"travel"}).sort({favCount:-1})
```

- **`1` 为升序排列，而 `-1` 是用于降序排列**。

#### 分页

`skip` 用于指定跳过记录数，`limit` 则用于限定返回结果数量：

```javascript
// 假定每页大小为 8 条，查询第 3 页的 book 文档
// 跳过前面 16 条记录，返回后面 8 条记录
db.books.find().skip(16).limit(8)
```

#### 分页优化

**数据量大的时候，应该避免使用 `skip/limit` 形式的分页**。

替代方案：**使用 `查询条件+唯一排序条件`**。例如： 

第一页：`db.books.find({}).sort({_id: 1}).limit(10)`
第二页：`db.books.find({_id: {$gt: <第一页最后一个_id>}}).sort({_id: 1}).limit(10)`; 
第三页：`db.books.find({_id: {$gt: <第二页最后一个_id>}}).sort({_id: 1}).limit(10)`;

**避免使用 count**：

尽可能不要计算总页数，特别是数据量大和查询条件不能完整命中索引时。 

假设集合总共有 1000w 条数据，在没有索引的情况下考虑以下查询：

```javascript
db.coll.find({x: 100}).limit(50);
db.coll.count({x: 100}); 
```

- 前者只需要遍历前 n 条，直到找到 50 条 `x=100` 的文档即可结束； 
- **后者需要遍历完 1000w 条找到所有符合要求的文档才能得到结果**。为了计算总页数而进行的 `count()` 往往是拖慢页面整体加载速度的原因。

### 更新文档

- `db.collection.updateOne()`：即使多个文档可能与指定的筛选器匹配，也只会更新第一个匹配的文档。
- `db.collection.updateMany()`：更新与指定筛选器匹配的所有文档。

#### 更新操作符

| 操作符 | 格式 | 描述 |
| --- | --- | --- |
| `$set` | `{$set: {field: value}}` | 指定一个键并更新值，**若键不存在则创建** |
| `$unset` | `{$unset: {field: 1}}` | 删除一个键 |
| `$inc` | `{$inc: {field: value}}` | 对数值类型进行增减 |
| `$rename` | `{$rename: {old_field_name: new_field_name}}` | 修改字段名称 |
| `$push` | `{$push: {field: value}}` | 向数组末尾添加一个元素 |
| `$pushAll` | `{$pushAll: {field: [value1, value2, ...]}}` | 向数组末尾添加多个元素 |
| `$pull` | `{$pull: {field: value}}` | 从数组中删除指定的元素 |
| `$addToSet` | `{$addToSet: {field: value}}` | 添加元素到数组中，具有排重功能 |
| `$pop` | `{$pop: {field: 1}}` | 删除数组的第一个或最后一个元素 |

#### 更新单个文档

```javascript
db.collection.updateOne(
   <filter>,
   <update>,
   {
     upsert: <boolean>,
     writeConcern: <document>,
     collation: <document>,
     arrayFilters: [ <filterdocument1>, ... ],
     hint:  <document|string>        // Available starting in MongoDB 4.2.1
   }
)

// favCount 加 1
db.books.updateOne({_id:ObjectId("642e62ec933c0dca8f8e9f60")},{$inc:{favCount:1}})
```

- `<filter>`：必选。一个筛选器对象，用于指定要更新的文档。只有与筛选器对象匹配的第一个文档才会被更新。
- `<update>`：必选。一个更新操作对象，用于指定如何更新文档。可以使用一些操作符，例如 `$set`、`$inc`、`$unset` 等，以更新文档中的特定字段。
- `upsert`：可选。一个布尔值，用于指定**如果找不到与筛选器匹配的文档时是否应插入一个新文档**。如果 `upsert` 为`true`，则会插入一个新文档。默认值为 `false`。
- `writeConcern`：可选。一个文档，用于指定写入操作的安全级别。可以指定写入操作需要到达的节点数或等待写入操作的时间。
- `collation`：可选。一个文档，用于指定用于查询的排序规则。例如，可以通过指定 `locale` 属性来指定语言环境，从而实现基于区域设置的排序。
- `arrayFilters`：可选。一个数组，用于指定要更新的数组元素。数组元素是通过使用更新操作符 `$[]` 和 `$` 来指定的。
- `hint`：一个文档或字符串，**用于指定查询使用的索引**。该参数仅在 MongoDB 4.2.1 及以上版本中可用。

#### 更新多个文档

`updateMany` 更新与集合的指定筛选器匹配的所有文档：

```javascript
// 给分类为 “novel” 的文档添加发布时间
db.books.updateMany({type:"novel"},{$set:{publishedDate:new Date()}})
```

#### findAndModify

`findAndModify` 兼容了查询和修改指定文档的功能，**`findAndModify` 只能更新单个文档**。

```javascript
// 将某个 book 文档的收藏数（favCount）加1
db.books.findAndModify({
    query:{_id:ObjectId("6457a39c817728350ec83b9d")},
    update:{$inc:{favCount:1}}
})
```


**默认情况下，`findAndModify` 会返回修改前的“旧”数据**。如果希望**返回修改后的数据，则可以指定 `new` 选项**：

```javascript
db.books.findAndModify({
    query:{_id:ObjectId("6457a39c817728350ec83b9d")},
    update:{$inc:{favCount:1}},
    new: true
})
```

与 `findAndModify` 语义相近的命令如下：
- `findOneAndUpdate`：更新单个文档并返回更新前（或更新后）的文档。
- `findOneAndReplace`：替换单个文档并返回替换前（或替换后）的文档。

### 删除文档

`deleteOne()` 和 `deleteMany()` 方法用来删除文档。

```javascript
db.books.deleteOne({type:"novel"})  // 删除 type 等于 novel 的第一个文档
db.books.deleteMany({})               // 删除集合下全部文档，最好使用 db.books.drop() 来删除。deleteMany 是一条条的删除，drop 是直接删除集合
db.books.deleteMany({type:"novel"}) // 删除 type 等于 novel 的全部文档
```

#### findOneAndDelete

`deleteOne` 命令在删除文档后只会返回确认性的信息，如果**希望获得被删除的文档**，则可以使用 `findOneAndDelete` 命令：

```javascript
db.books.findOneAndDelete({type:"novel"})
```

`findOneAndDelete` 命令还允许定义**删除的顺序**，即按照指定顺序删除找到的第一个文档。利用这个特性，`findOneAndDelete` 可以实现**队列的先进先出**。

```javascript
db.books.findOneAndDelete({type:"novel"},{sort:{favCount:1}})
```

### 批量操作

`bulkwrite()` 方法提供了执行**批量插入、更新和删除操作**的能力。支持以下写操作:

- `insertOne`：插入单个文档。
- `updateOne`：更新单个文档。
- `updateMany`：更新多个文档。
- `replaceOne`：替换单个文档。
- `deleteOne`：删除单个文档。
- `deleteMany`：删除多个文档。

每个写操作都作为数组中的文档传递给 `bulkWrite()`。

```javascript
db.pizzas.insertMany( [
   { _id: 0, type: "pepperoni", size: "small", price: 4 },
   { _id: 1, type: "cheese", size: "medium", price: 7 },
   { _id: 2, type: "vegan", size: "large", price: 8 }
])

db.pizzas.bulkWrite([
   { insertOne: { document: { _id: 3, type: "beef", size: "medium", price: 6 } } },
   
   { insertOne: { document: { _id: 4, type: "sausage", size: "large", price: 10 } } },
   
   { updateOne: {
      filter: { type: "cheese" },
      update: { $set: { price: 8 } }
   }},

   { deleteOne: { filter: { type: "pepperoni"} } },

   { replaceOne: {
      filter: { type: "vegan" },
      replacement: { type: "tofu", size: "small", price: 4 }
   }}
])
```

#### 可控的执行顺序

- 有序模式（`ordered: true`）：

操作按顺序执行，适合有依赖关系的场景（如先删除再插入）。

- 无序模式（`ordered: false`）：

操作并行执行，最大化吞吐量，适合独立操作。

#### bulkWrite 是非原子性的

默认情况下：**`bulkWrite` 不是原子性的**。

**如果中间某个操作失败，已执行的操作不会回滚（部分成功）**。

例如：批量插入 10 个文档，第 5 个失败时，前 4 个仍会写入。

- **有序模式**（`ordered: true`）：

操作按顺序执行，**遇到错误会停止后续操作**（但**已执行的不会回滚**）。

- **无序模式**（`ordered: false`）：

操作**并行执行**，**失败的操作不影响其他操作**。

#### 批量操作的优势

- 减少网络开销：

将多个操作合并为一个请求发送到服务器，避免多次网络往返延迟（尤其在高延迟环境中优势明显）。

- 批量处理优化：

MongoDB 会对批量操作进行内部优化（如顺序写入、减少锁竞争），比单条操作更高效。