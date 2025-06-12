---
title: 聚合操作
weight: 3
---

聚合操作允许用户**处理多个文档并返回计算结果**。

聚合操作分为三类：

- **单一作用聚合**：一种简单的聚合操作，用于**对单个集合中的文档进行操作**。常用的 `db.collection.estimateDocumentCount()`、`db.collection.countDocument()`、`db.collection.distinct()` 等这类单一作用的聚合函数。
- **聚合管道**：是一个**数据聚合的框架，基于数据处理流水线的概念**。文档进入多级管道，每个管道可以对文档进行各种操作，包括过滤、投影、分组、排序等，将文档转为聚合结果。
- **MapReduce**：已经被废弃，不再使用。

## 聚合管道

**MongoDB 聚合框架（Aggregation Framework）是一个计算框架**，它可以：

- 作用在一个或几个集合上
- 对集合中的数据进行一系列运算
- 将这些数据转化为期望的形式

类似于 SQL 查询的 `GROUP BY`、`LEFT JOIN`、`AS` 等。

### 管道（Pipeline）和阶段（Stage）

**整个聚合运算过程称为管道（Pipeline），管道由多个阶段（Stage）组成**，每个管道：

- 接受一系列文档作为输入（原始数据）
- 每个阶段对这些文档进行一系列运算
- 结果文档作为下一个阶段的输入

通过将多个操作符组合到聚合管道中，用户可以构建出足够复杂的数据处理管道以提取数据并进行分析。

![mongodb-aggregation](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-aggregation.png)


聚合管道的语法：

```javascript
pipeline = [$stage1, $stage2, ...$stageN];
db.collection.aggregate(pipeline, {options})
```

- `pipeline`：**一组聚合阶段**。除了 `$out`、`$merge` 和 `$geoNear` 之外，其他阶段都可以在管道中多次出现。
- `options`：可选参数，包含：查询计划、是否使用临时文件、游标、最大操作时间、读写策略、强制索引等。

例如，对订单做一个聚合操作：

```javascript
db.students.aggregate([
    { $match: { status: "A" } }, // match stage，过滤订单状态为 A 的
    { $group: { _id: "$cus_id", total: { $sum: "$amount" } } } // group stage，类似 SQL 的 group by，按照 cus_id 分组，对 amount 求和，放到 total 字段中
]);
```

![mongodb-aggregate-demo](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-aggregate-demo.png)

#### 常用的聚合阶段运算符

| 阶段运算符 | 描述 | SQL 等价运算符 |
| --- | --- | --- |
| `$match` | 过滤文档 | `WHERE` |
| `$project` | 投影 | `AS` |
| `$lookup` | 左连接 | `LEFT JOIN` |
| `$sort` | 排序 | `ORDER BY` |
| `$group` | 分组 | `GROUP BY` |
| `$skip/$limit` | 跳过/限制 | `LIMIT` |
| `$unwind` | 展开数组 |  |
| `$graphLookup` | 图查询 |  |
| `$facet/$bucket` | 分面搜索 |  |

官方文档：[Aggregation Pipeline Stages](https://www.mongodb.com/zh-cn/docs/manual/reference/operator/aggregation-pipeline/)

#### 聚合表达式

获取字段信息：

```javascript
$<field>  ： 用 $ 指示字段路径
$<field>.<sub field>  ： 使用 $  和 .  来指示内嵌文档的路径
```

常量表达式：

```javascript
$literal :<value> ： 指示常量 <value>
```

系统变量表达式：

```javascript
$$<variable>  使用 $$ 指示系统变量
$$CURRENT  指示管道中当前操作的文档
```

### 使用聚合

准备数据：

```javascript
var tags = ["nosql","mongodb","document","developer","popular"];
var types = ["technology","sociality","travel","novel","literature"];
var books=[];
for(var i=0;i<50;i++){
    var typeIdx = Math.floor(Math.random()*types.length);
    var tagIdx = Math.floor(Math.random()*tags.length);
    var tagIdx2 = Math.floor(Math.random()*tags.length);
    var favCount = Math.floor(Math.random()*100);
    var username = "xx00"+Math.floor(Math.random()*10);
    var age = 20 + Math.floor(Math.random()*15);
    var book = {
        title: "book-"+i,
        type: types[typeIdx],
        tag: [tags[tagIdx],tags[tagIdx2]],
        favCount: favCount,
        author: {name:username,age:age}
    };
    books.push(book)
}
db.books.insertMany(books);
```

可以使用 `load()` 方法将数据加载到当前数据库中，也可以直接在 MongoDB 客户端中执行上面的代码。

查看数据：

```javascript
db.books.countDocuments(); // 50
```

#### $project

投影操作，**将原始字段投影成指定名称**，如将集合中的 `title` 字段投影成 `name` 字段：

```javascript
db.books.aggregate([{$project:{name:"$title"}}])
```

`$project` 可以灵活控制输出文档的格式，也可以**剔除不需要的字段**：

```javascript
db.books.aggregate([{$project:{name:"$title",_id:0,type:1,author:1}}]) // 输出文档中不包含 _id 字段，只包含 name、type 和 author 字段
```

**从嵌套文档中排除字段**:

```javascript
db.books.aggregate([
    {$project:{name:"$title",_id:0,type:1,"author.name":1}}
])
// 或者
db.books.aggregate([
    {$project:{name:"$title",_id:0,type:1,author:{name:1}}}
])
```

#### $match

`$match` 用于**对文档进行过滤**，之后可以在得到的文档子集上做聚合，`$match` 可以使用除了地理空间之外的所有常规查询操作符。

在实际应用中**尽可能将 `$match` 放在管道的前面位置**。这样有两个好处：

1. 可以**快速将不需要的文档过滤掉，减少后续管道操作符要操作的文档数，提升效率**。
2. 如果**在投射和分组之前执行 `$match`，查询可以使用索引**。

```javascript
db.books.aggregate([{$match:{type:"technology"}}])

// 将 $match 放在管道的前面位置，减少后续管道操作符要操作的文档数，提升效率
db.books.aggregate([
    {$match:{type:"technology"}},
    {$project:{name:"$title",_id:0,type:1,author:{name:1}}}
])
```

#### $count

计数并返回与查询匹配的结果数：

```javascript
db.books.aggregate([
    {$match:{type:"technology"}}, // match 阶段筛选出 type 为 technology 的文档
    {$count: "type_count"} // count 阶段返回聚合管道中剩余文档的计数，，并将该值分配给 type_count 字段
])
```

#### $group

**按指定的表达式对文档进行分组，并将每个不同分组的文档输出到下一个阶段**。输出文档包含一个 `_id` 字段，该字段按键包含不同的组。

输出文档还可以包含计算字段，该字段保存由 `$group` 的 `_id` 字段分组的一些 accumulator 表达式的值。`$group` 不会输出具体的文档而只是统计信息。

```javascript
{$group: { _id: <expression>, <field1>: { <accumulator1> : <expression1> }, ... } }
```

- `_id` 字段是必填的。但是，可以指定 `_id` 值为 `null` 来为整个输入文档计算累计值。
- 剩余的计算字段是可选的，并使用 `<accumulator>` 运算符进行计算。
- `_id` 和 `<accumulator>` 表达式可以接受任何有效的表达式。

##### accumulator 操作符

| 名称 | 描述 | 类比 SQL |
| --- | --- | --- |
| `$sum` | 计算指定表达式的总和。 | `sum()` |
| `$avg` | 计算平均值 | `AVG` |
| `$first` | 返回每组第一个文档，如果有排序，按照排序，如果没有按照默认的存储的顺序的第一个文档。 | `LIMIT 0,1` |
| `$last` | 返回每组最后一个文档，如果有排序，按照排序，如果没有按照默认的存储的顺序的最后个文档。 |  |
| `$max` | 根据分组，获取集合中所有文档对应值得最大值。 | `max()` |
| `$min` | 根据分组，获取集合中所有文档对应值得最小值。 | `min()` |
| `$push` | 将指定的表达式的值添加到一个数组中。 | `array_append()` |
| `$addToSet` | 将表达式的值添加到一个集合中（无重复值，无序）。 | `array_append()` |
| `$stdDevPop` | 返回输入值的总体标准偏差（population standard deviation）。 |  |
| `$stdDevSamp` | 返回输入值的样本标准偏差（the sample standard deviation）。 |  |

**`$group` 阶段的内存限制为 100M。默认情况下，如果 stage 超过此限制，`$group` 将产生错误**。但是，要允许处理大型数据集，可以**将 `allowDiskUse` 选项设置为 `true` 以启用 `$group` 操作以写入临时文件**。

##### 示例

book 的数量，收藏总数和平均值：

```javascript
db.books.aggregate([
    // _id 类似于 SQL 的 count(*)，使用 $group 需要一个分组字段，但是这里不需要分组，所以可以使用 null
    {$group:{_id:null,count:{$sum:1},pop:{$sum:"$favCount"},avg:{$avg:"$favCount"}}}
])
```

统计每个作者的 book 收藏总数：

```javascript
db.books.aggregate([
    {$group:{_id:"$author.name",pop:{$sum:"$favCount"}}}
])
```

统计每个作者的每本 book 的收藏数，**多个字段分组**：

```javascript
db.books.aggregate([
    // _id 可以是多个字段，这里是 author.name 表示每个作者， title 表示每本 book
    {$group:{_id:{name:"$author.name",title:"$title"},pop:{$sum:"$favCount"}}}
])
```

每个作者的 book 的 type 合集：

```javascript
db.books.aggregate([
    {$group:{_id:"$author.name",types:{$addToSet:"$type"}}}
])
```

#### $unwind

**可以将数组拆分为单独的文档**。

```javascript
{
  $unwind:
    {
      // 要指定字段路径，在字段名称前加上$符并用引号括起来。
      path: <field path>,
      // 可选,一个新字段的名称用于存放元素的数组索引。该名称不能以 $ 开头。
      includeArrayIndex: <string>,  
      // 可选，default: false，若为 true,如果路径为空，缺少或为空数组，则 $unwind 输出文档
      preserveNullAndEmptyArrays: <boolean> 
 } }
```

姓名为 xx006 的作者的 book 的 tag 数组拆分为多个文档：

```javascript
db.books.aggregate([
    {$match:{"author.name":"xx006"}},
    {$unwind:"$tag"}
])
```

每个作者的 book 的 tag 合集：

```javascript
db.books.aggregate([
    {$unwind:"$tag"},
    {$group:{_id:"$author.name",types:{$addToSet:"$tag"}}}
])
```

##### 案例

准备数据：

```javascript
db.books.insert([
{
	"title" : "book-51",
	"type" : "technology",
	"favCount" : 11,
     "tag":[],
	"author" : {
		"name" : "fox",
		"age" : 28
	}
},{
	"title" : "book-52",
	"type" : "technology",
	"favCount" : 15,
	"author" : {
		"name" : "fox",
		"age" : 28
	}
},{
	"title" : "book-53",
	"type" : "technology",
	"tag" : [
		"nosql",
		"document"
	],
	"favCount" : 20,
	"author" : {
		"name" : "fox",
		"age" : 28
	}
}])
```

测试：

```javascript
// 使用 includeArrayIndex 选项来输出数组元素的数组索引
db.books.aggregate([
    {$match:{"author.name":"fox"}},
    {$unwind:{path:"$tag", includeArrayIndex: "arrayIndex"}}
])
```

运行结果：

`unwind` 按照 `tag` 数组中的元素拆分文档，每个元素都有一个 `arrayIndex` 字段，用于指示元素在数组中的位置。

```bash
[
    {
        "_id": ObjectId("65776072401f11001831313e"),
        "title": "book-53",
        "type": "technology",
        "tag": "nosql",
        "favCount": 20,
        "author": {
            "name": "fox",
            "age": 28
        },
        "arrayIndex": 0
    },
    {
        "_id": ObjectId("65776072401f11001831313e"),
        "title": "book-53",
        "type": "technology",
        "tag": "document",
        "favCount": 20,
        "author": {
            "name": "fox",
            "age": 28
        },
        "arrayIndex": 1
    }
]
```

上面的示例拆出了两个文档，分别是 `nosql` 和 `document`。并没有都取出来，这种来做统计肯定是由问题的，因为由的文档没有 `tag` 字段。这时候就可以使用 `preserveNullAndEmptyArrays` 选项来输出缺少 `tag` 字段，`null` 或空数组的文档。

```javascript
// 使用 preserveNullAndEmptyArrays 选项在输出中包含缺少 tag 字段，null 或空数组的文档
// 因为 MognoDB 的结构很灵活，有些字段在文档中可能不存在
db.books.aggregate([
    {$match:{"author.name":"fox"}},
    {$unwind:{path:"$tag", preserveNullAndEmptyArrays: true}}
])
```

#### $limit

限制传递到管道中下一阶段的文档数：

```javascript
db.books.aggregate([
    {$limit : 5 }
])
```

{{< callout type="info" >}}
这里 MongoDB 对 `$limit` 进行了优化，**如果 `$limit` 之前的阶段有 `$sort`，则 `$limit` 不会对所有的数据进行排序，`$sort` 操作只会在过程中维持前 n 个结果，其中 n 是指定的限制，而 MongoDB 只需要将 n 个项存储在内存中**。
{{< /callout >}}

#### $skip

跳过进入 stage 的指定数量的文档，并将其余文档传递到管道中的下一个阶段：

```javascript
db.books.aggregate([
    {$skip : 5 }
])
```

此操作将跳过管道传递给它的前 5 个文档。`$skip` 对沿着管道传递的文档的内容没有影响。

#### $sort

对所有输入文档进行排序，并按排序顺序将它们返回到管道。

```javascript
{$sort: { <field1>: <sort order>, <field2>: <sort order> ... } }
```

要对字段进行排序，请将排序顺序设置为 **`1` 或 `-1`，以分别指定升序或降序排序**，如下例所示：

```javascript
db.books.aggregate([
    {$sort : {favCount:-1,"author.age":1}}
])
```

#### $lookup

Mongodb 3.2 版本新增，主要用来实现**多表关联查询**，相当关系型数据库中多表关联查询。每个输入待处理的文档，经过 `$lookup` 阶段的处理，输出的新文档中会包含一个新生成的数组（可根据需要命名新 key）。数组列存放的数据是来自被 Join 集合的适配文档，如果没有，集合为空（即 为`[]`）。

```javascript
db.collection.aggregate([{
      $lookup: {
             from: "<collection to join>",
             localField: "<field from the input documents>",
             foreignField: "<field from the documents of the from collection>",
             as: "<output array field>"
           }
})
```

- `from`：待连接的集合名称。
- `localField`：源集合中的 match 值，如果输入的集合中，某文档没有 `localField` 这个 Key（Field），在处理的过程中，会默认为此文档含有 `localField：null` 的键值对。
- `foreignField`：待连接的集合的 match 值，如果待连接的集合中，文档没有 `foreignField` 值，在处理的过程中，会默认为此文档含有 `foreignField：null` 的键值对。
- `as`：为输出文档的新增值命名。如果输入的集合中已存在该值，则会覆盖掉。

其语法功能类似于下面的伪 SQL 语句：

```sql
SELECT *, <output array field>
FROM collection
WHERE <output array field> IN (SELECT *
                               FROM <collection to join>
                               WHERE <foreignField>= <collection.localField>);
```

##### 案例

准备数据：

```javascript
// customer 客户集合，customerCode 是客户 id
db.customer.insert({customerCode:1,name:"customer1",phone:"13112345678",address:"test1"})
db.customer.insert({customerCode:2,name:"customer2",phone:"13112345679",address:"test2"})

// order 订单集合，orderId 是订单 id，customerCode 是客户 id，用来管来关联 customer 集合
db.order.insert({orderId:1,orderCode:"order001",customerCode:1,price:200})
db.order.insert({orderId:2,orderCode:"order002",customerCode:2,price:400})

// orderItem 订单详情集合，orderId 是订单 id，用来关联 order 集合
db.orderItem.insert({itemId:1,productName:"apples",qutity:2,orderId:1})
db.orderItem.insert({itemId:2,productName:"oranges",qutity:2,orderId:1})
db.orderItem.insert({itemId:3,productName:"mangoes",qutity:2,orderId:1})
db.orderItem.insert({itemId:4,productName:"apples",qutity:2,orderId:2})
db.orderItem.insert({itemId:5,productName:"oranges",qutity:2,orderId:2})
db.orderItem.insert({itemId:6,productName:"mangoes",qutity:2,orderId:2})
```

关联查询：

```javascript
// 查询客户信息并包含客户相关的订单信息
db.customer.aggregate([        
    {$lookup: {
       from: "order",
       localField: "customerCode",
       foreignField: "customerCode",
       as: "customerOrder"
     }
    } 
])
```

运行结果：

```bash
[
    {
        "_id": ObjectId("65776072401f11001831313e"),
        "customerCode": 1,
        "name": "customer1",
        "phone": "13112345678",
        "address": "test1",
        "customerOrder": [
            {
                "_id": ObjectId("65776072401f11001831313e"),
                "orderId": 1,
                "orderCode": "order001",
                "customerCode": 1,
                "price": 200
            }
        ]
    },
    {
        "_id": ObjectId("65776072401f11001831313e"),
        "customerCode": 2,
        "name": "customer2",
        "phone": "13112345679",
        "address": "test2",
        "customerOrder": [ 
            {
                "_id": ObjectId("65776072401f11001831313e"),
                "orderId": 2,
                "orderCode": "order002",
                "customerCode": 2,
                "price": 400
            }
        ]
    }
]
```

```javascript
// 查询订单信息并包含订单相关的客户信息和订单详情信息
db.order.aggregate([
    {$lookup: {
               from: "customer",
               localField: "customerCode",
               foreignField: "customerCode",
               as: "curstomer"
             }
        
    },
    {$lookup: {
               from: "orderItem",
               localField: "orderId",
               foreignField: "orderId",
               as: "orderItem"
             }
    }
])
```

运行结构：

```bash
[
    {
        "_id": ObjectId("65776072401f11001831313e"),
        "orderId": 1,
        "orderCode": "order001",
        "customerCode": 1,
        "price": 200,
        "curstomer": [
            {
                "_id": ObjectId("65776072401f11001831313e"),
                "customerCode": 1,
                "name": "customer1",
                "phone": "13112345678",
                "address": "test1"
            }
        ],
        "orderItem": [
            {
                "_id": ObjectId("65776072401f11001831313e"),
                "itemId": 1,
                "productName": "apples",
                "qutity": 2,
                "orderId": 1    
            },
            {
                "_id": ObjectId("65776072401f11001831313e"),
                "itemId": 2,
                "productName": "oranges",
                "qutity": 2,
                "orderId": 1
            },
            {
                "_id": ObjectId("65776072401f11001831313e"),
                "itemId": 3,
                "productName": "mangoes",
                "qutity": 2,
                "orderId": 1
            }
        ]
    },
    {
        "_id": ObjectId("65776072401f11001831313e"),
        "orderId": 2,
        "orderCode": "order002",
        "customerCode": 2,
        "price": 400,
        "curstomer": [
            {
                "_id": ObjectId("65776072401f11001831313e"),
                "customerCode": 2,
                "name": "customer2",
                "phone": "13112345679",
                "address": "test2" 
            }
        ],
        "orderItem": [
            {
                "_id": ObjectId("65776072401f11001831313e"),
                "itemId": 4,
                "productName": "apples",
                "qutity": 2,
                "orderId": 2
            },
            {
                "_id": ObjectId("65776072401f11001831313e"),
                "itemId": 5,
                "productName": "oranges",
                "qutity": 2,
                "orderId": 2    
            },
            {
                "_id": ObjectId("65776072401f11001831313e"),
                "itemId": 6,
                "productName": "mangoes",
                "qutity": 2,
                "orderId": 2
            }
        ]
    }
]
```

#### 聚合操作案例 1

统计每个分类的 book 文档数量：

```javascript
db.books.aggregate([
    {$group:{_id:"$type",total:{$sum:1}}},
    {$sort:{total:-1}}
])
```

标签的热度排行，标签的热度则按其关联 book 文档的收藏数（favCount）来计算：

```javascript
db.books.aggregate([
    {$match:{favCount:{$gt:0}}}, // match 阶段：用于过滤去除掉 favCount=0 的文档
    {$unwind:"$tag"},            // unwind 阶段：用于将标签数组进行展开，这样一个包含 3 个标签的文档会被拆解为 3 个条目
    {$group:{_id:"$tag",total:{$sum:"$favCount"}}}, // group 阶段：对拆解后的文档进行分组计算，$sum："$favCount" 表示按 favCount 字段进行累加。
    {$sort:{total:-1}} // sort 阶段：接收分组计算的输出，按 total 得分进行排序。
])
```

统计 book 文档收藏数 `[0,10)`,`[10,60)`,`[60,80)`,`[80,100)`,`[100,+∞)`：

```javascript
db.books.aggregate([{
    // $bucket 阶段：用于将文档按照指定的边界值进行分组，将文档的 favCount 字段的值分配到不同的桶中。
    $bucket:{
        groupBy:"$favCount",
        boundaries:[0,10,60,80,100],
        default:"other",
        output:{"count":{$sum:1}}
    }
}])
```

#### 聚合操作案例 2

使用 `mongoimport` 工具导入数据。

导入邮政编码数据集: https://media.mongodb.org/zips.json

使用 `mongoimport` 工具导入数据：

```bash
mongoimport -h 192.168.65.174 -d test -u fox -p fox --authenticationDatabase=admin -c zips --file D:\ProgramData\mongodb\import\zips.json  
```

```
h,--host ：代表远程连接的数据库地址，默认连接本地Mongo数据库；
--port：代表远程连接的数据库的端口，默认连接的远程端口27017；
-u,--username：代表连接远程数据库的账号，如果设置数据库的认证，需要指定用户账号；
-p,--password：代表连接数据库的账号对应的密码；
-d,--db：代表连接的数据库；
-c,--collection：代表连接数据库中的集合；
-f, --fields：代表导入集合中的字段；
--type：代表导入的文件类型，包括csv和json,tsv文件，默认json格式；
--file：导入的文件名称
--headerline：导入csv文件时，指明第一行是列名，不需要导入；
```

返回人口超过 1000 万的州：

```javascript
db.zips.aggregate( [
   { $group: { _id: "$state", totalPop: { $sum: "$pop" } } },
   { $match: { totalPop: { $gte: 10*1000*1000 } } }
])
```

等价 SQL 是：

```sql
SELECT state, SUM(pop) AS totalPop
FROM zips
GROUP BY state
HAVING totalPop >= (10*1000*1000)
```

返回各州平均城市人口：

```javascript
db.zips.aggregate( [
   // group 阶段：_id: { state: "$state", city: "$city" } 用于按州和城市进行分组，cityPop: { $sum: "$pop" } 用于计算每个城市的人口总和。
   { $group: { _id: { state: "$state", city: "$city" }, cityPop: { $sum: "$pop" } } },
   // group 阶段：_id: "$_id.state" 用于按州进行分组，avgCityPop: { $avg: "$cityPop" } 用于计算每个州的平均城市人口。 $_id.state 表示上一个阶段的 _id 字段中的 state 字段的值。
   { $group: { _id: "$_id.state", avgCityPop: { $avg: "$cityPop" } } }
])
```

按州返回最大和最小的城市：

```javascript
db.zips.aggregate( [
   // group 阶段：_id: { state: "$state", city: "$city" } 用于按州和城市进行分组，pop: { $sum: "$pop" } 用于计算每个城市的人口总和。 
   { $group:
      {
        _id: { state: "$state", city: "$city" },
        pop: { $sum: "$pop" }
      }
   },
   // sort 阶段：对每个州的城市人口进行排序，pop: 1 表示按人口升序排序。
   { $sort: { pop: 1 } },
   // group 阶段：_id: "$_id.state" 用于按州进行分组，biggestCity: { $last: "$_id.city" } 用于获取每个州中人口最大的城市，biggestPop: { $last: "$pop" } 用于获取每个州中人口最大的城市的人口。
   { $group:
      {
        _id : "$_id.state",
        biggestCity:  { $last: "$_id.city" },
        biggestPop:   { $last: "$pop" },
        smallestCity: { $first: "$_id.city" },
        smallestPop:  { $first: "$pop" }
      }
   },
   // project 阶段：选择要返回的字段，_id: 0 表示不返回 _id 字段，state: "$_id" 表示将 _id 字段的值作为 state 字段的值，biggestCity: { name: "$biggestCity", pop: "$biggestPop" } 表示将 biggestCity 和 biggestPop 字段的值作为 biggestCity 字段的值的嵌套对象。
   {$project:
      {
        _id: 0,
        state: "$_id",
        biggestCity:  { name: "$biggestCity",  pop: "$biggestPop" },
        smallestCity: { name: "$smallestCity", pop: "$smallestPop" }
      }
   }
])
```

运行结果，最终可以拿到每个州最大和最小的城市：

```bash
[
    {
        "state": "AK",
        "biggestCity": {
            "name": "Anchorage",
            "pop": 29730
        },
        "smallestCity": {
            "name": "Fairbanks",
            "pop": 583
        }
    },
    {
        "state": "AL",
        "biggestCity": {
            "name": "Mobile",
            "pop": 1498337
        },
        "smallestCity": {
            "name": "Birmingham",
            "pop": 48758
        }
    }
    // ...
]
```

## 聚合优化

官方文档：[聚合管道优化](https://www.mongodb.com/zh-cn/docs/manual/core/aggregation-pipeline-optimization/)

优化的三个目标：

- **尽可能利用索引完成搜索和排序 -> 快速找到数据，快速排序**
- **尽早尽多的减少数据量 -> 减少 CPU 消耗，减少内存消耗**
- **尽可能减少执行步骤 -> 减少内存消耗，缩短响应时间**

### 执行顺序

#### $match/$sort vs $project/$addFields

**为了使查询能够命中索引，`$match/$sort` 步骤需要在最前面**，该原则适用于 MongoDB <=3.4。**MongoDB 3.6 开始具备一定的自动优化能力**。

#### $project + $skip/$limit

**`$skip/$limit` 应该尽可能放在 `$project` 之前，减少 `$project` 的工作量**。**3.6 开始自动完成这个优化**。

### 内存排序

**在没有索引支持的情况下**，MongoDB 最多只支持使用 **100MB 内存进行排序**。假设总共可用内存为 16GB，一个请求最多可以使用 100MB 内存排序，总共可以有 `16000/ 100= 160` 个请求同时执行。

**内存排序消耗的不仅是内存，还有大量 CPU**。

- **方案一： `$sort + $limit`**：只排 Top N ，只要 N 条记录总和不超过 100MB 即可。
- **方案二： `{allowDiskUse: true}`**：使用磁盘作为交换空间完成全量，超出 100MB 部分与磁盘交换排序。
- **方案三： `索引排序`**：使用索引完成排序，没有内存限制