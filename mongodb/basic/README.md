# MongoDB 基础
## 基本概念
- `database`，和"数据库"一样的概念 (对 Oracle 来说就是 `schema`)。一个 MongoDB 实例中，可以有零个或多个数据库
- `collections`，数据库中可以有零个或多个 collections (集合)。和传统意义上的`table`基本一致。
- `documents`，集合是由零个或多个 `documents` (文档)组成。一个文档可以看成是一 `row`。
- `fields`，文档是由零个或多个 `fields` (字段)组成。可以看成是 `columns`。
- `Indexes` (索引)在 MongoDB 中扮演着和它们在 RDBMS(Relational Database Management System 关系数据库管理系统) 中一样的角色。
- `Cursors` (游标)，游标是，当你问 MongoDB 拿数据的时候，它会给你返回一个结果集的指针而不是真正的数据，这个指针我们叫它**游标**，
我们可以拿游标做我们想做的任何事情，比如说计数或者跨行之类的，而无需把真正的数据拖下来，在真正的数据上操作。

这些概念和关系型数据中的概念类似，但是还是有差异的。

**核心差异在于，关系型数据库是在 `table` 上定义的 `columns`，而面向文档数据库是在 `document` 上定义的 `fields`。
也就是说，在 `collection` 中的每个 `document` 都可以有它自己独立的 `fields`**。

要点就是，集合不对存储内容严格限制 (所谓的**无模式**(`schema-less`))。

## mongo shell
mongo shell 用的是 JavaScript。比如 `db.help()` 或者 `db.stats()`。大多数情况下我们会操作集合而不是数据库，
用 `db.COLLECTION_NAME` ，比如 `db.unicorns.help()` 或者 `db.unicorns.count()`。**如果输入 `db.help` (不带括号), 你会看到 `help` 方法的内部实现**。

## `_id`
每个文档都会有一个唯一 `_id` 字段。你可以自己生成一个，或者让 MongoDB 帮你生成一个 `ObjectId` 类型的。默认的 `_id` 字段是已被索引的。
`_id` 是一个 12 字节长的十六进制数，头 4 个字节代表的是当前的时间戳，接着的后 3 个字节表示的是机器 id 号，接着的 2 个字节表示 MongoDB 服务器进程 id，最后的 3 个字节代表递增值。

## 数据类型
- **String**：字符串。存储数据常用的数据类型。在 MongoDB 中，UTF-8 编码的字符串才是合法的。
- **Integer**：整型数值。用于存储数值。根据你所采用的服务器，可分为 32 位或 64 位。
- **Boolean**：布尔值。用于存储布尔值（真/假）。
- **Double**：双精度浮点值。用于存储浮点值。
- **Min/Max keys**：将一个值与 BSON（二进制的 JSON）元素的最低值和最高值相对比。
- **Arrays**：用于将数组或列表或多个值存储为一个键。
- **Timestamp**：时间戳。记录文档修改或添加的具体时间。
- **Object**：用于内嵌文档。
- **Null**：用于创建空值。
- **Symbol**：符号。该数据类型基本上等同于字符串类型，但不同的是，它一般用于采用特殊符号类型的语言。
- **Date**：日期时间。用 UNIX 时间格式来存储当前日期或时间。你可以指定自己的日期时间：创建 Date 对象，传入年月日信息。
- **Object ID**：对象 ID。用于创建文档的 ID。
- **Binary Data**：二进制数据。用于存储二进制数据。
- **Code**：代码类型。用于在文档中存储 JavaScript 代码。
- **Regular expression**：正则表达式类型。用于存储正则表达式。

## 常用命令
### `use`
`use` 会创建一个新的数据库，如果该数据库存在，则返回这个数据库。格式 `use DATABASE_NAME`。

### 删除数据库
`dropDatabase()`用于删除已有数据库。格式 `db.dropDatabase()`。
它将删除选定的数据库。如果没有选定要删除的数据库，则它会将默认的 `test` 数据库删除。

```sh
>use mydb
switched to db mydb
>db.dropDatabase()
>{ "dropped" : "mydb", "ok" : 1 }
```

### 创建集合
`db.createCollection(name, options)` 创建集合。`name` 是所要创建的集合名称。`options` 是一个用来指定集合配置的文档。

参数 `options` 是可选的，可用选项：
- `capped`，（可选）如果为 `true`，则创建固定集合。固定集合是指有着固定大小的集合，当达到最大值时，它会自动覆盖最早的文档。当该值为 `true` 时，必须指定 `size` 参数。
- `autoIndexID`，（可选）如为 `true`，自动在 `_id` 字段创建索引。默认为 `false`。
- `size`，（可选）为固定集合指定一个最大值（以字节计）。如果 `capped` 为 `true`，也需要指定该字段。
- `max`，（可选）指定固定集合中包含文档的最大数量。

**在插入文档时，MongoDB 首先检查固定集合的 `size` 字段，然后检查 `max` 字段**。

#### 固定集合
##### isCapped()
`isCapped()`检查集合是否是固定集合。格式 `db.COLLECTION_NAME.isCapped()`

##### 将现有集合转化为固定集合
```js
db.runCommand({"convertToCapped":"posts",size:10000})
```
将现有的 `posts` 集合转化为固定集合。

##### 值得注意的点

- 无法从固定集合中删除文档。
- 固定集合没有默认索引，甚至在 `_id` 字段中也没有。
- 在插入新的文档时，MongoDB 并不需要寻找磁盘空间来容纳新文档。它只是盲目地将新文档插入到集合末尾。这使得固定集合中的插入操作是非常快速的。
- 同样的，在读取文档时，MongoDB 会按照插入磁盘的顺序来读取文档，从而使读取操作也非常快。

### 删除集合
`db.collection.drop()` 来删除数据库中的集合。格式 `db.COLLECTION_NAME.drop()`。

### 插入文档
使用 `insert()` 或 `save()` 方法。格式 `db.COLLECTION_NAME.insert(document)`。

### 查询
使用 `find()` 方法。格式 `db.COLLECTION_NAME.find()`。

#### `pretty()` 方法
用格式化方式显示结果，使用的是 `pretty()` 方法。

#### `findOne()`
`findOne()` 方法，它只返回一个文档。

#### 类似于 WHERE 子句的语句
`$lt`小于，`$lte`小于等于，`$gt`大于，`$gte`大于等于，`$ne`不等于。

```js
db.mycol.find({"likes":{$lt:50}}).pretty()
```

#### AND 和 OR
```js
// 逗号分隔看成是 AND 条件
db.mycol.find({key1:value1, key2:value2}).pretty()
```

基于 `OR` 条件来查询文档，可以使用关键字 `$or`。
```js
db.mycol.find({$or: [{key1: value1}, {key2:value2}]}).pretty()
```

#### 映射（Projection）
映射（Projection）指的是只选择文档中的必要数据，而非全部数据。如果文档有 5 个字段，而你只需要显示 3 个，则只需选择 3 个字段即可。
执行 find() 方法时，可以利用 0 或 1 来设置字段列表。** 1 用于显示字段，0 用于隐藏字段**。

```js
db.mycol.find({},{"title":1, _id:0})
```
`_id` 字段是一直显示的。如果不想显示该字段，则可以将其设为 0。

#### limit()
`limit()` 方法接受一个数值类型的参数，其值为想要显示的文档数。

```js
db.mycol.find({},{"title":1, _id:0}).limit(2)
```

#### skip()
```js
db.mycol.find({},{"title":1,_id:0}).skip(1).limit(1)
```

#### sort()
`sort()` 方法可以通过一些参数来指定要进行排序的字段，并使用 1 和 -1 来指定排序方式，其中** 1 表示升序，而 -1 表示降序**。

```js
db.mycol.find({},{"title":1,_id:0}).sort({"title":-1})
```


### 更新
`update()` 方法更新已有文档中的值，而 `save()` 方法则是用传入该方法的文档来替换已有文档。格式 `db.COLLECTION_NAME.update(SELECTIOIN_CRITERIA, UPDATED_DATA)`。

```js
db.mycol.update({'title':'MongoDB Overview'},{$set:{'title':'New MongoDB Tutorial'}})
```

### 删除
`remove()` 方法 清除集合中的文档。格式 `db.COLLECTION_NAME.remove(DELLETION_CRITTERIA)`。 2 个可选参数：
- `deletion criteria`：（可选）删除文档的标准。
- `justOne`：（可选）如果设为 `true` 或 `1`，则只删除一个文档。

```js
db.mycol.remove({'title':'MongoDB Overview'}, 1)
```

## 索引
`ensureIndex()` 方法创建索引。格式 `db.COLLECTION_NAME.ensureIndex({KEY:1})`。**1 代表按升序排列字段值。-1 代表按降序排列**。

```js
db.mycol.ensureIndex({"title":1})

// 为多个字段创建索引
db.mycol.ensureIndex({"title":1,"description":-1})
```

`ensureIndex()` 方法也可以接受一些可选参数：
- `background`，在后台构建索引，从而不干扰数据库的其他活动。取值为 `true` 时，代表在后台构建索引。默认值为 `false`
- `unique`，创建一个唯一的索引，从而当索引键匹配了索引中一个已存在值时，集合不接受文档的插入。取值为 `true` 代表创建唯一性索引。默认值为 `false`。
- `name`，索引名称。如果未指定，MongoDB 会结合索引字段名称和排序序号，生成一个索引名称。
- `dropDups`，在可能有重复的字段内创建唯一性索引。MongoDB 只在某个键第一次出现时进行索引，去除该键后续出现时的所有文档。
- `sparse`，如果为 `true`，索引只引用带有指定字段的文档。这些索引占据的空间较小，但在一些情况下的表现也不同（特别是排序）。默认值为 `false`。
- `expireAfterSeconds`，指定一个秒数值，作为 TTL 来控制 MongoDB 保持集合中文档的时间。
- `v`，索引版本号。默认的索引版本跟创建索引时运行的 MongoDB 版本号有关。
- `weights`，文档数值，范围从 1 到 99, 999。表示就字段相对于其他索引字段的重要性。
- `default_language`，对文本索引而言，用于确定停止词列表，以及词干分析器（stemmer）与断词器（tokenizer）的规则。默认值为 `english`。
- `language_override`，对文本索引而言，指定了文档所包含的字段名，该语言将覆盖默认语言。默认值为 `language`。

### 覆盖索引查询

在每一个 MongoDB 官方文档中，覆盖查询都具有以下两个特点：

- 查询中的所有字段都属于一个索引；
- 查询所返回的所有字段也都属于同一索引内。

既然**查询中的所有字段都属于一个索引，MongoDB 就会利用同一索引，匹配查询集合并返回结果，而不需要实际地查看文档。因为索引存在于 RAM 中，从索引中获取数据要比通过扫描文档获取数据快得多**。

#### 使用覆盖查询
假设在一个 `users` 集合中包含下列文档：
```js
{
   "_id": ObjectId("53402597d852426020000002"),
   "contact": "987654321",
   "dob": "01-01-1991",
   "gender": "M",
   "name": "Tom Benzamin",
   "user_name": "tombenzamin"
}
```
为 `users` 集合中的 `gender` 和 `user_name` 字段创建一个复合索引：
```js
db.users.ensureIndex({gender:1,user_name:1})
```

这一索引将覆盖下列查询：
```js
db.users.find({gender:"M"},{user_name:1,_id:0})
```

也就是说，对于上面的查询，MongoDB 不会去查看文档，转而从索引数据中获取所需的数据。

下面的查询就不会被覆盖，因为`_id`会默认返回，而`_id`和`user_name`，`gender`不在同一个索引。
```js
db.users.find({gender:"M"},{user_name:1})
```

如果出现下列情况，索引不能覆盖查询：
- 索引字段是数组
- 索引字段是子文档

### 高级索引
例如下面的 `user` 集合文档:
```js
{
   "address": {
      "city": "Los Angeles",
      "state": "California",
      "pincode": "123"
   },
   "tags": [
      "music",
      "cricket",
      "blogs"
   ],
   "name": "Tom Benzamin"
}
```
上述文档包含一个地址**子文档**（address sub-document）与一个标签**数组**（tags array）。

#### 索引数组字段
假设我们想要根据标签来搜索用户文档。首先在集合中创建一个标签数组的索引。

反过来说，**在标签数组上创建一个索引，也就为每一个字段创建了单独的索引项**。因此在该例中，当我们创建了标签数组的索引时，
也就为它的music（音乐）、cricket（板球）以及 blog（博客）值创建了独立的索引。

```js
// 创建标签数据的索引
db.users.ensureIndex({"tags":1})

// 搜索集合中的标签字段
db.users.find({tags:"cricket"})

// 使用 explain 命令验证所使用索引的正确性
db.users.find({tags:"cricket"}).explain()
```

上述 `explain` 命令的执行结果是 `"cursor" : "BtreeCursor tags_1"`，表示使用了正确的索引。

#### 索引子文档字段
假设需要根据市（city）、州（state）、个人身份号码（pincode）字段来搜索文档。因为所有这些字段都属于地址子文档字段的一部分，
所以我们将在子文档的所有字段上创建索引。

```js
// 在子文档的所有三个字段上创建索引
db.users.ensureIndex({"address.city":1,"address.state":1,"address.pincode":1})

// 搜索子文档字段
db.users.find({"address.city":"Los Angeles"})

// 查询
db.users.find({"address.city":"Los Angeles","address.state":"California"})

// 也支持如下这样的查询
db.users.find({"address.city":"LosAngeles","address.state":"California","address.pincode":"123"})
```

**查询表达式必须遵循指定索引的顺序**。

### 索引限制
#### 额外开销
每个索引都会占据一些空间，从而也会在每次插入、更新与删除操作时产生一定的开销。所以如果**集合很少使用读取操作，就尽量不要使用索引**。

#### 内存使用
因为**索引存储在内存**中，所以应**保证索引总体的大小不超过内存的容量**。如果索引总体积超出了内存容量，就会删除部分索引，从而降低性能。

#### 查询限制
当查询使用以下元素时，不能使用索引：
- 正则表达式或否定运算符（`$nin`、`$not`，等等）
- 算术运算符（比如 `$mod`）
- `$where` 子句
因此，经常检查查询使用的索引是一个明智的做法。

#### 索引键限制
自 MongoDB 2.6 版本起，如果已有索引字段的值超出了索引键限制，则无法创建索引。
- 插入文档超过索引键限制
- **如果文档的索引字段值超出了索引键的限制，MongoDB 不会将任何文档插入已索引集合**。类似于使用 `mongorestore` 和 `mongoimport` 工具时的情况。

#### 最大范围
- 集合索引数不能超过 64 个。
- 索引名称长度不能大于 125 个字符。
- 复合索引最多能有 31 个被索引的字段。

## 聚合
聚合操作能将多个文档中的值组合起来，对成组数据执行各种操作，返回单一的结果。使用 `aggregate()` 方法。相当于 SQL 中的 `count(*)` 组合 `group by`。
```js
db.mycol.aggregate([{$group : {_id : "$by_user", num_tutorial : {$sum : 1}}}])
```
上例使用 `by_user` 字段来组合文档，每遇到一次 `by_user`，就递增之前的合计值。

| 表达式 | 描述 | 范例 |
| --- | --- | --- |
| `$sum` | 对集合中所有文档的定义值进行加和操作 | `db.mycol.aggregate([{$group : {_id : "$by_user", num_tutorial : {$sum : "$likes"}}}])` |
| `$avg` | 对集合中所有文档的定义值进行平均值 | `db.mycol.aggregate([{$group : {_id : "$by_user", num_tutorial : {$avg : "$likes"}}}])` |
| `$min` | 计算集合中所有文档的对应值中的最小值 | `db.mycol.aggregate([{$group : {_id : "$by_user", num_tutorial : {$min : "$likes"}}}])` |
| `$max` | 计算集合中所有文档的对应值中的最大值 | `db.mycol.aggregate([{$group : {_id : "$by_user", num_tutorial : {$max : "$likes"}}}])` |
| `$push` | 将值插入到一个结果文档的数组中 | `db.mycol.aggregate([{$group : {_id : "$by_user", url : {$push: "$url"}}}])` |
| `$addToSet` | 将值插入到一个结果文档的数组中，但不进行复制 | `db.mycol.aggregate([{$group : {_id : "$by_user", url : {$addToSet : "$url"}}}])` |
| `$first` | 根据成组方式，从源文档中获取第一个文档。但只有对之前应用过 `$sort` 管道操作符的结果才有意义。 | `db.mycol.aggregate([{$group : {_id : "$by_user", first_url : {$first : "$url"}}}])` |
| `$last` | 根据成组方式，从源文档中获取最后一个文档。但只有对之前进行过 `$sort` 管道操作符的结果才有意义。 | `db.mycol.aggregate([{$group : {_id : "$by_user", last_url : {$last : "$url"}}}])` |


### 管道
管道（pipeline）概念指的是能够在一些输入上执行一个操作，然后将输出结果用作下一个命令的输入。MongoDB 的聚合架构也支持这种概念。管道中有很多阶段（stage），
在每一阶段中，管道操作符都会将一组文档作为输入，产生一个结果文档（或者管道终点所得到的最终 JSON 格式的文档），然后再将其用在下一阶段。

聚合架构中可能采取的管道操作符有：

- `$project` 用来选取集合中一些特定字段。
- `$match` 过滤操作。减少用作下一阶段输入的文档的数量。
- `$group` 如上所述，执行真正的聚合操作。
- `$sort` 对文档进行排序。
- `$skip` 在一组文档中，跳过指定数量的文档。
- `$limit` 将查看文档的数目限制为从当前位置处开始的指定数目。
- `$unwind` 解开使用数组的文档。当使用数组时，数据处于预连接状态，通过该操作，数据重新回归为各个单独的文档的状态。利用该阶段性操作可增加下一阶段性操作的文档数量。

```js
db.test.aggregate([
  {$match: {uuid: 'sfsdfsfd'}},
  {$project: {completeNum: 1, failedNum: 1, createTime: 1, _id: 0}},
  {$group: {
    _id: '$createTime',
    completeTotal: {$sum: '$completeNum'},
    failedTotal: {$sum: '$failedNum'}}}
]);
```