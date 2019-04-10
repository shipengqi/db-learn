# 高级部分
## 关系
MongoDB 中的**关系表示文档之间的逻辑相关方式**。关系可以通过**内嵌**（Embedded）或**引用**（Referenced）两种方式建模。
这样的**关系可能是 1：1、1：N、N：1，也有可能是 N：N**。

例如，一个用户可能有多个地址，这是一个 1：N 的关系。

```js
// user 文档结构
{
   "_id":ObjectId("52ffc33cd85242f436000001"),
   "name": "Tom Hanks",
   "contact": "987654321",
   "dob": "01-01-1991"
}
// address 文档结构
{
   "_id":ObjectId("52ffc4a5d85242602e000000"),
   "building": "22 A, Indiana Apt",
   "pincode": 123456,
   "city": "Los Angeles",
   "state": "California"
}

// 内嵌关系的建模
{
   "_id":ObjectId("52ffc33cd85242f436000001"),
   "contact": "987654321",
   "dob": "01-01-1991",
   "name": "Tom Benzamin",
   "address": [
      {
         "building": "22 A, Indiana Apt",
         "pincode": 123456,
         "city": "Los Angeles",
         "state": "California"
      },
      {
         "building": "170 A, Acropolis Apt",
         "pincode": 456789,
         "city": "Chicago",
         "state": "Illinois"
      }]
}
```

**该方法会将所有相关数据都保存在一个文档中**，从而易于检索和维护。**缺点是，如果内嵌文档不断增长，会对读写性能造成影响**。

### 引用关系的建模
这是一种设计归一化关系的方法。按照这种方法，这种引用关系也被称作**手动引用**，所有的用户和地址文档都将分别存放，而用户文档会包含一个字段，用来引用地址文档 id 字段。
```js
{
   "_id":ObjectId("52ffc33cd85242f436000001"),
   "contact": "987654321",
   "dob": "01-01-1991",
   "name": "Tom Benzamin",
   "address_ids": [
      ObjectId("52ffc4a5d85242602e000000"),
      ObjectId("52ffc4a5d85242602e000001")
   ]
}
```

数组字段 `address_ids` 含有相应地址的 `ObjectId` 对象。利用这些 `ObjectId`，能够查询地址文档，从而获取地址细节信息。利用这种方法时，需要进行两种查询：
首先从 `user` 文档处获取 `address_ids`，其次从 `address` 集合中获取这些地址。
```js
var result = db.users.findOne({"name":"Tom Benzamin"},{"address_ids":1})
var addresses = db.address.find({"_id":{"$in":result["address_ids"]}})
```

## 数据库引用
上一节中，我们使用引用关系实现了归一化的数据库结构，这种引用关系也被称作**手动引用**，即可以手动地将引用文档的 id 保存在其他文档中。
有些情况下，文档包含其他集合的引用时，我们可以使用**数据库引用**（MongoDB DBRefs）。

如何用数据库引用代替手动引用。假设一个数据库中存储有多个类型的地址（家庭地址、办公室地址、邮件地址，等等），这些地址保存在不同的集合中
（`address_home`、`address_office`、`address_mailing`，等等）。当 `user` 集合的文档引用了一个地址时，它还需要按照地址类型来指定所需要查看的集合。
这种情况下，一个文档引用了许多结合的文档，所以就应该使用 `DBRef`。

### 使用数据库引用
`DBRef` 中有三个字段：
- `$ref` 该字段指定所引用文档的集合。
- `$id` 该字段指定引用文档的 `_id` 字段
- `$db` 该字段是可选的，包含引用文档所在数据库的名称。

假如在一个简单的 `user` 文档中包含着 `DBRef` 字段 `address`，如下所示：
```js
{
   "_id":ObjectId("53402597d852426020000002"),
   "address": {
   "$ref": "address_home",
   "$id": ObjectId("534009e4d852427820000002"),
   "$db": "tutorialspoint"},
   "contact": "987654321",
   "dob": "01-01-1991",
   "name": "Tom Benzamin"
}
```
数据库引用字段 `address` 指定出，引用地址文档位于 `tutorialspoint` 数据库的 `address_home` 集合中，并且它的 `id` 为 `534009e4d852427820000002`。

在由 `$ref` 所指定的集合（本例中为 `address_home`）中，如何动态查找由 `$id` 所确定的文档。
```js
var user = db.users.findOne({"name":"Tom Benzamin"})
var dbRef = user.address
db[dbRef.$ref].findOne({"_id":(dbRef.$id)})
```

## 查询分析
对于衡量数据库及索引设计的效率来说，分析查询是一个很重要的衡量方式。经常使用的查询有 `$explain` 和 `$hint`。

### $explain
`$explain` 操作提供的消息包括：查询消息、查询所使用的索引以及其他的统计信息。

例如：
```js
// 创建索引
db.users.ensureIndex({gender:1,user_name:1})

// 查询
db.users.find({gender:"M"},{user_name:1,_id:0}).explain()

// 输出
{
   "cursor" : "BtreeCursor gender_1_user_name_1",
   "isMultiKey" : false,
   "n" : 1,
   "nscannedObjects" : 0,
   "nscanned" : 1,
   "nscannedObjectsAllPlans" : 0,
   "nscannedAllPlans" : 1,
   "scanAndOrder" : false,
   "indexOnly" : true,
   "nYields" : 0,
   "nChunkSkips" : 0,
   "millis" : 0,
   "indexBounds" : {
      "gender" : [
         [
            "M",
            "M"
         ]
      ],
      "user_name" : [
         [
            {
               "$minElement" : 1
            },
            {
               "$maxElement" : 1
            }
         ]
      ]
   }
}
```

- `indexOnly` 为`true`代表该查询使用了索引。
- `cursor` 字段指定了游标所用的类型。`BTreeCursor` 类型代表了使用了索引并且提供了所用索引的名称。`BasicCursor` 表示进行了完整扫描，没有使用任何索引。
- `n` 代表所返回的匹配文档的数量。
- `nscannedObjects` 表示已扫描文档的总数。
- `nscanned` 所扫描的文档或索引项的总数。

### $hint
`$hint` 操作符**强制索引优化器使用指定的索引运行查询**。这尤其适用于测试带有多个索引的查询性能。比如，下列查询指定了用于该查询的 `gender` 和 `user_name` 字段的索引：
```js
db.users.find({gender:"M"},{user_name:1,_id:0}).hint({gender:1,user_name:1})
```


## Map Reduce
`Map-Reduce`（映射归约）是一种将大量数据压缩成有用的聚合结果的数据处理范式。MongoDB 使用 `mapReduce` 命令来实现映射归约操作。映射归约通常用来处理大型数据。

`mapReduce` 命令的基本格式为：
```js
db.collection.mapReduce(
   function() {emit(key,value);},  //map function
   function(key,values) {return reduceFunction},   //reduce function
   {
      out: collection,
      query: document,
      sort: document,
      limit: number
   }
)
```
**`mapReduce` 函数首先查询集合，然后将结果文档利用 `emit` 函数映射为键值对，然后再根据有多个值的键来简化**。

- `map` 一个 JavaScript 函数，将一个值与键对应起来，并生成键值对。
- `reduce` 一个 JavaScript 函数，用来减少或组合所有拥有同一键的文档。
- `out` 指定映射归约查询结果的位置。
- `query` 指定选择文档所用的选择标准（可选的）。
- `sort` 指定可选的排序标准。
- `limit` 指定返回的文档的最大数量值（可选的）。

### 使用
例如，下面这个存储用户发帖的文档结构。该文档存储用户的用户名（user_name）和发帖状态（status）。
```js
{
   "post_text": "tutorialspoint is an awesome website for tutorials",
   "user_name": "mark",
   "status":"active"
}
```

在 `posts` 集合上使用 `mapReduce` 函数选择所有的活跃帖子，将它们基于用户名组合起来，然后计算每个用户的发帖量。代码如下：
```js
db.posts.mapReduce(
  function() { emit(this.user_id,1); },
  function(key, values) {return Array.sum(values)},
  {
     query:{status:"active"},
     out:"post_total"
  }
)
```

输出：
```js
{
   "result" : "post_total",
   "timeMillis" : 9,
   "counts" : {
      "input" : 4,
      "emit" : 4,
      "reduce" : 2,
      "output" : 2
   },
   "ok" : 1,
}
```
结果显示，只有 4 个文档符合查询条件（`status:"active"`），于是 `map` 函数就生成了 4 个带有键值对的文档，而最终 `reduce` 函数将具有相同键值的映射文档变为了 2 个。

### 查看 mapReduce 查询的结果
使用 `find` 操作符。
```js
db.posts.mapReduce(
   function() { emit(this.user_id,1); },
   function(key, values) {return Array.sum(values)},
      {
         query:{status:"active"},
         out:"post_total"
      }
).find()
```

## 全文检索
### 启用文本搜索
最初的文本搜索只是一种试验性功能，但从 2.6 版本起就成为默认功能了。但如果使用的是之前的 MongoDB，则需要使用下列代码启用文本搜索：
```js
db.adminCommand({setParameter:true,textSearchEnabled:true})
```

### 创建文本索引
```js
db.posts.ensureIndex({post_text:"text"})
```

上面的代码在 `post_text` 字段上创建文本索引，以便搜索帖子文本之内的内容。

在 `post_text` 字段上创建了文本索引，接下来搜索包含 `tutorialspoint` 文本内容的帖子。
```js
db.posts.find({$text:{$search:"tutorialspoint"}})
```

### 删除文本索引
```js
// 找到索引名称
db.posts.getIndexes()

// 删掉
db.posts.dropIndex("post_text_text")
```

## 正则表达式
正则表达式在所有语言当中都是经常会用到的一个功能，可以用来搜索模式或字符串中的单词。MongoDB 也提供了这一功能，使用 $regex 运算符来匹配字符串模式。
MongoDB 使用 PCRE（可兼容 Perl 的正则表达式）作为正则表达式语言。

使用正则表达式不需要使用任何配置或命令。

假如 posts 集合有下面这个文档，它包含着帖子文本及其标签。
```js
{
   "post_text": "enjoy the mongodb articles on tutorialspoint",
   "tags": [
      "mongodb",
      "tutorialspoint"
   ]
}
```

使用下列正则表达式来搜索包含 tutorialspoint 的所有帖子。
```js
db.posts.find({post_text:{$regex:"tutorialspoint"}})

// 或者
db.posts.find({post_text:/tutorialspoint/})
```

### 不区分大小写
要想使搜索不区分大小写，使用 `$options` 参数和值 `$i`。
```js
db.posts.find({post_text:{$regex:"tutorialspoint",$options:"$i"}})
```

### 使用正则表达式来处理数组元素
还可以在数组字段上使用正则表达式。在实现标签的功能时，这尤为重要。假如想搜索标签以 “tutorial” 开始（tutorial、tutorials、tutorialpoint 或 tutorialphp）的帖子，可以使用下列代码：

```js
db.posts.find({tags:{$regex:"tutorial"}})
```

### 优化正则表达式查询
- 如果文档字段已经设置了索引，查询将使用索引值来匹配正则表达式，从而使查询效率相对于扫描整个集合的正则表达式而言大大提高。
- 如果正则表达式为前缀表达式，所有的匹配结果都要在前面带有特殊的前缀字符串。比如，如果正则表达式为 `^tut`，那么查询将搜索所有以 `tut` 开始的字符串。

## ID自增
默认情况下，MongoDB 将 `_id` 字段（使用 12 字节的 `ObjectId`）来作为文档的唯一标识。但在有些情况下，我们希望 `_id` 字段值能够自动增长，而不是固守在 `ObjectId` 值上。

使用 `counters` 集合来程序化地实现该功能。

1. 使用 counters 集合
假设存在下列文档 `products`，我们希望 `_id` 字段值是一个能够自动增长的整数序列（1、2、3、4 …… n）。
```js
{
  "_id":1,
  "product_name": "Apple iPhone",
  "category": "mobiles"
}
```
创建一个 `counters` 集合，其中包含了所有序列字段最后的序列值。
现在，将文档（`productid` 是它的键）插入到 `counters` 集合中：
```js
db.counters.insert({_id:"productid",sequence_value:0})
```
`sequence_value` 字段保存了序列的最后值。

2. 创建一个 `getNextSequenceValue` 函数
创建一个 `getNextSequenceValue` 函数，该函数以序列名为输入，按照 1 的幅度增加序列数，返回更新的序列数。在该例中，序列名称为 `productid`。
```js
function getNextSequenceValue(sequenceName){
   var sequenceDocument = db.counters.findAndModify(
      {
         query:{_id: sequenceName },
         update: {$inc:{sequence_value:1}},
         new:true
      });
   return sequenceDocument.sequence_value;
}
```

3. 使用 `getNextSequenceValue` 函数

```js
db.products.insert({"_id":getNextSequenceValue("productid"),"product_name":"Apple iPhone","category":"mobiles"})
db.products.insert({"_id":getNextSequenceValue("productid"),"product_name":"Samsung S3","category":"mobiles"})
```