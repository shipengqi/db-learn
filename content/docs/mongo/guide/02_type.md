---
title: 数据类型
weight: 2
---

## BSON 协议与数据类型

### 为什么会使用 BSON？

JSON 是当今非常通用的一种跨语言Web数据交互格式，属于 ECMAScript 标准规范的一个子集。JSON（JavaScript Object Notation, JS 对象简谱）即 JavaScript 对象表示法，它是 **JavaScript 对象的一种文本表现形式**。

大多数情况下，使用 JSON 作为数据交互格式已经是理想的选择，但是 **JSON 基于文本的解析效率并不是最好的，在某些场景下往往会考虑选择更合适的编/解码格式**，一些做法如：

- 在微服务架构中，使用 gRPC（基于 Google 的 Protobuf）可以获得更好的网络利用率。
- 分布式中间件、数据库，使用私有定制的 TCP 数据包格式来提供高性能、低延时的计算能力。

BSON（Binary JSON）是二进制版本的JSON，其在性能方面有更优的表现。BSON 在许多方面和 JSON 保持一致，其同样也支持内嵌的文档对象和数组结构。**二者最大的区别在于 JSON 是基于文本的，而 BSON 则是二进制（字节流）编/解码的形式**。在空间的使用上，BSON 相比 JSON 并没有明显的优势。

**MongoDB 在文档存储、命令协议上都采用了 BSON 作为编/解码格式**，主要具有如下优势：

- 类 JSON 的轻量级语义，支持简单清晰的嵌套、数组层次结构，可以实现模式灵活的文档结构。
- 更高效的遍历，BSON在编码时会记录每个元素的长度，可以直接通过 seek 操作进行元素的内容读取，相对 JSON 解析来说，遍历速度更快。
- 更丰富的数据类型，除了 JSON 的基本数据类型，BSON 还提供了 MongoDB 所需的一些扩展类型，比如日期、二进制数据等，这更加方便数据的表示和操作。

### BOSN 和 JSON 存储结构对比

JSON存储示例（UTF-8 编码）：

```json
{
  "name": "张三",
  "age": 30,
  "scores": [95.5, 89.0],
  "meta": {"id": "A123"}
}
```

实际存储（十六进制表示）：

```hex
7B 0A 20 20 22 6E 61 6D 65 22 3A 20 22 E5 BC A0 E4 B8 89 22 2C 0A 20 20 22 61 67 65 22 3A 20 33 30 2C 0A 20 20 22 73 63 6F 72 65 73 22 3A 20 5B 39 35 2E 35 2C 20 38 39 2E 30 5D 2C 0A 20 20 22 6D 65 74 61 22 3A 20 7B 22 69 64 22 3A 20 22 41 31 32 33 22 7D 0A 7D
```

相同数据的 BSON 编码（十六进制）：

```bson
5A 00 00 00                // 总长度 90 字节
02                         // 字符串类型
6E 61 6D 65 00             // 字段名 "name"
07 00 00 00                // 字符串长度7字节
E5 BC A0 E4 B8 89 00       // UTF-8 值"张三"
10                         // 32 位整数类型
61 67 65 00                // 字段名 "age"
1E 00 00 00                // 值 30
04                         // 数组类型
73 63 6F 72 65 73 00       // 字段名"scores"
26 00 00 00                // 数组长度38字节
01                         // 双精度浮点类型
30 00                      // 元素索引 "0"
00 00 00 00 00 C0 57 40    // 值 95.5
01                         // 双精度浮点类型
31 00                      // 元素索引"1"
00 00 00 00 00 60 56 40    // 值 89.0
00                         // 数组结束
03                         // 文档类型
6D 65 74 61 00             // 字段名 "meta"
12 00 00 00                // 内嵌文档长度 18 字节
02                         // 字符串类型
69 64 00                   // 字段名 "id"
05 00 00 00                // 字符串长度 5 字节
41 31 32 33 00             // 值 "A123"
00                         // 内嵌文档结束
00                         // 主文档结束
```

### BOSN 和 JSON 解析过程对比

**JSON 解析流程**

1. 字符解码：将字节流解码为UTF-8字符串
2. 词法分析：识别标记字符和值
   - 遇到 `{` 开始对象
   - 遇到 `"` 开始字符串
   - 遇到 `:` 分隔键值
3. 语法分析：构建内存数据结构
4. 类型转换：将字符串表示的值转为对应类型

```javascript
// 伪代码示例
function parseJSON(input) {
  let index = 0;
  function parseValue() {
    skipWhitespace();
    const ch = input[index];
    if (ch === '{') return parseObject();
    if (ch === '[') return parseArray();
    if (ch === '"') return parseString();
    if (ch === '-' || isDigit(ch)) return parseNumber();
    // ...其他类型
  }
  // 其他解析函数...
  return parseValue();
}
```

**BSON 解析流程**

1. 长度检查：读取前 4 字节获取文档总长度
2. 类型驱动解析：根据类型标记决定解析方式
   - `0x01`：双精度浮点（直接读取 8 字节）
   - `0x02`：字符串（先读长度再读内容）
   - `0x10`：32 位整数（直接读取 4 字节）
3. 直接内存映射：数值类型无需转换直接拷贝
4. 长度前缀跳转：通过长度前缀快速跳过不需要的字段

```cpp
// 伪代码示例
void parseBSON(const uint8_t* data) {
  uint32_t length = readInt32(data);
  data += 4;
  
  while (*data != 0x00) { // 直到遇到结束符
    uint8_t type = *data++;
    char* fieldName = readCString(data);
    
    switch(type) {
      case 0x01: // Double
        double value = readDouble(data);
        data += 8;
        break;
      case 0x02: // String
        uint32_t strLen = readInt32(data);
        data += 4;
        char* str = readString(data, strLen);
        data += strLen;
        break;
      // ...其他类型处理
    }
  }
}
```

### BSON 的数据类型

MongoDB 中，一个 BSON 文档最大大小为 16M，文档嵌套的级别不超过 100。

| Type | Number | Alias | Description |
| --- | --- | --- | --- |
| Double | 1 | "double" |  |
| String | 2 | "string" |  |
| Object | 3 | "object" |  |
| Array | 4 | "array" |  |
| Binary data | 5 | "binData" | 二进制数据 |
| Undefined | 6 | "undefined" | **Deprecated** |
| ObjectId | 7 | "objectId" | 对象 ID，用于创建**文档 ID，这个 ID 是由客户端生成** |
| Boolean | 8 | "bool" |  |
| Date | 9 | "date" |  |
| Null | 10 | "null" |  |
| Regular expression | 11 | "regex" | 正则表达式 |
| DBPointer | 12 | "dbPointer" | **Deprecated** |
| JavaScript | 13 | "javascript" |  |
| Symbol | 14 | "symbol" | **Deprecated** |
| JavaScript code with scope | 15 | "javascriptWithScope" | **Deprecated** in MongoDB 4.4. |
| 32-bit integer | 16 | "int" |  |
| Timestamp | 17 | "timestamp" |  |
| 64-bit integer | 18 | "long" |  |
| Decimal128 | 19 | "decimal" | 3.4 新类型 |
| Min key | -1 | "minKey" | 表示一个最小值 |
| Max key | 127 | "maxKey" | 表示一个最大值 |

#### $type 操作符

`$type` 操作符基于 BSON 类型来**检索集合中匹配的数据类型**，并返回结果。

```javascript
// 这里的 2 就是类型的 Number，就是 "string"
db.books.find({"title" : {$type : 2}})
// 或者
db.books.find({"title" : {$type : "string"}})
```

#### 日期类型

MongoDB 的**日期类型使用 UTC（Coordinated Universal Time，即世界协调时）进行存储**，也就是 **`+0` 时区**的时间。

```javascript
db.dates.insertMany([{data1:Date()},{data2:new Date()},{data3:ISODate()}])
db.dates.find().pretty()
```

- `Date()` 生成的 JavaScript 的时间字符串。
- `new Date()`与 `ISODate()` 最终都会生成 ISODate 类型的字段（对应于 UTC 时间）。


#### ObjectId 生成器

MongoDB 集合中所有的文档都有一个**唯一的 `_id` 字段，作为集合的主键**。在默认情况下，`_id` 字段使用 `ObjectId` 类型，采用 16 进制编码形式，共 **12 个字节**。

为了避免文档的 `_id` 字段出现重复，**`ObjectId` 被定义为 3 个部分**：

- 4 字节表示 Unix **时间戳**（秒）。
- 5 字节表示随机数（**`机器号+进程号唯一`**，3 个字节机器号，2 个字节进程号）。 
- 3 字节表示**计数器**（初始化时随机）。

生成一个新的 `ObjectId`，可以直接调用 `ObjectId()` 函数。

```javascript
db.books.insertOne({_id: ObjectId(),title:"MongoDB"})
db.books.find().pretty()
```

`ObjectId` 由几个属性方法：

- `ObjectId.getTimestamp()`：将对象的时间戳部分作为日期返回。
- `ObjectId.toString()`：以字符串形式返回 `ObjectId`。
- `ObjectId.valueOf()`：将对象的表示形式返回为十六进制字符串。返回的字符串是 str 属性。

#### 内嵌文档和数组

一个文档中可以包含作者的信息，包括作者名称、性别、家乡所在地，一个显著的优点是，当我们查询 book 文档的信息时，作者的信息也会一并返回。

```javascript
db.books.insert({
    title: "撒哈拉的故事",
    author: {
        name:"三毛",
        gender:"女",
        hometown:"重庆"
    }
})

// 查询三毛的作品
db.books.find({"author.name":"三毛"})

// 修改三毛的家乡所在地
db.books.update({"author.name":"三毛"},{$set:{"author.hometown":"北京"}})
```

除了作者信息，文档中还包含了若干个标签，这些标签可以用来表示文档所包含的一些特征：

```javascript
db.books.updateOne({"author.name":"三毛"},{$set:{tags:["旅行","随笔","散文","爱情","文学"]}})

// 会查询到所有的 tags
db.books.find({"author.name":"三毛"},{title:1,tags:1})

// 利用 $slice 获取最后一个 tag
db.books.find({"author.name":"三毛"},{title:1,tags:{$slice:-1}})
```

- `{title:1,tags:1}` 表示只返回 `title` 和 `tags` 字段。
- 默认 `_id` 字段会被返回，如果**不想要返回 `_id` 字段，可以使用 `{_id:0,title:1,tags:1}`**。
- `{tags:{$slice:-1}}` 表示只返回 `tags` 数组的最后一个元素。`-2` 表示返回最后两个元素，`-3` 表示返回最后三个元素，以此类推。

数组末尾追加元素，可以使用 `$push` 操作符：

```javascript
db.books.updateOne({"author.name":"三毛"},{$push:{tags:"科幻"}})
```

`$push` 操作符可以配合其他操作符，一起实现不同的数组修改操作，比如和 `$each` 操作符配合可以用于添加多个元素：

```javascript
db.books.updateOne({"author.name":"三毛"},{$push:{tags:{$each:["悬疑","推理"]}}})
```

如果加上 `$slice` 操作符，那么**只会保留经过切片后的元素**，下面的例子中，只会保留最后三个元素：

```javascript
db.books.updateOne({"author.name":"三毛"},{$push:{tags:{$each:["悬疑","推理"],$slice: -3}}})
```

根据元素查询：

```javascript
// 查询 tags 数组中包含科幻的文档
db.books.find({"tags":"科幻"})

// 查询 tags 数组中同时包含科幻和推理的文档
db.books.find({tags:{$all:["悬疑","推理"]}})
```

#### 嵌套型的数组

数组元素可以是基本类型，也可以是内嵌的文档结构：

```javascript
{
    tags:[
        {tagKey:xxx,tagValue:xxxx},
        {tagKey:xxx,tagValue:xxxx}
    ]
}
```

这种结构非常灵活，一个很适合的场景就是商品的多属性，例如一个商品可以同时包含多个维度的属性，比如颜色、尺寸、重量等。

```javascript
db.goods.insertMany([{
    name:"羽绒服",
    tags:[
        {tagKey:"size",tagValue:["M","L","XL","XXL","XXXL"]},
        {tagKey:"color",tagValue:["黑色","宝蓝"]},
        {tagKey:"style",tagValue:"韩风"}
    ]
},{
    name:"羊毛衫",
    tags:[
        {tagKey:"size",tagValue:["L","XL","XXL"]},
        {tagKey:"color",tagValue:["蓝色","杏色"]},
        {tagKey:"style",tagValue:"韩风"}
    ]
}])
```

当需要根据属性进行检索时，需要用到 `$elemMatch` 操作符：

```javascript
// 筛选出 color=黑色 的商品信息
db.goods.find({
    tags:{
        $elemMatch:{tagKey:"color",tagValue:"黑色"}
    }
})
```

如果进行组合式的条件检索，可以使用多个 `$elemMatch` 操作符：

```javascript
// 筛选出 color=黑色，style=韩范 的商品信息
db.goods.find({
    tags:{
        $elemMatch:{tagKey:"color",tagValue:"黑色"},
        $elemMatch:{tagKey:"style",tagValue:"韩范"}
    }
})
```

## 固定封顶集合

**固定集合（capped collection）是一种限定大小的集合**，其中 `capped` 是覆盖、限额的意思。跟普通的集合相比，数据写入这种集合时遵循 FIFO（先进先出）的原则，即**当集合达到最大容量时，最早写入的数据会被自动删除**。可以将这种集合理解为一个**环形缓冲区**，当集合满时，新的数据会覆盖最早写入的数据。通过固定集合的大小，可以**保证数据库的存储空间不会无限增长，超过限额的旧数据会被丢弃**。

### 创建固定集合

```javascript
db.createCollection("logs",{capped:true,size:4096,max:10})
```

- `capped`：表示创建的是固定集合。
- `size`：指集合**占用空间的最大值**，这里是 4096 字节，即 4KB。
- `max`：表示集合的**文档数量的最大值**，这里是 10 条。

`size` 是必选的，`max` 是可选的。如果同时指定了 `size` 和 `max`，只要满足其中一个条件，就会认为集合已满。

`collection.stats()` 可以**查看文档的占用空间大小**：

```javascript
db.logs.stats()
```

**将普通集合转换为固定集合**：

```javascript
db.runCommand({"convertToCapped": "mycoll", size: 100000})
```

### 适用场景

**固定集合适合用来存储一些“临时态”的数据**，临时态意味着数据在一定程度上可以被丢弃。同时用户还应该更关注最新的数据，**随着时间的推移，数据的重要性逐渐降低，直至被淘汰**。

- 日志数据：比如网站的访问日志、错误日志等。
- 存储少量文档：比如最新发布的 TopN 文章信息。比如集合就设置 `max` 为 10 条，这样就可以保持只存储最新的 10 条文档，查询的时候就可以直接使用 `find()` 方法。

#### 存储股票价格变动信息

在股票实时系统中，大家往往最关心股票的价格变动。而应用系统中也需要根据这些实时的变化数据来分析当前行情。若将股票的价格变化看作是一个事件，而股票交易所则是价格变动的**发布者**，股票 APP，应用系统则是事件的**消费者**。这样就可以将股票价格的发布、通知抽象为一种数据的消费行为，此时需要一个消息队列来实现。

**利用固定集合实现存储股票价格变动的消息队列**：

1. 创建 `stock_queue` 消息队列，可以容纳 10MB 的数据。

```javascript
db.createCollection("stock_queue",{capped:true,size:10485760})
```

2. 定义消息格式：

```javascript
{
    timestamped:new Date(), // 股票动态消息的产生时间
    stock: "MongoDB Inc",   // 股票名称
    price: 20.33           // 股票价格，double 类型
}
```

为了支持按时间检索，比如查询某个时间点之后的数据，可以为 `timestamped` 字段添加索引：

```javascript
db.stock_queue.createIndex({timestamped:1})
```

3. 构建生产者，发布股票动态：

```javascript
// 每隔 1 秒向队列中插入一条股票价格变动信息
function pushEvent(){
    while(true){
        db.stock_queue.insert({
            timestamped:new Date(),
            stock: "MongoDB Inc",
            price: 100*Math.random(1000)
        });
        print("publish stock changed");
        sleep(1000);
    }
}

// 执行 pushEvent 函数
pushEvent();
```

4. 构建消费者，消费股票动态

对于消费者来说，更关心的是最新的数据，同时还应该保持持续进行拉取，以便知晓实时发生的变化。

```javascript
function listen(){
    var cursor = db.stock_queue.find({timestamped:{$gte:new Date()}}).tailable();
    while(true){
        if(cursor.hasNext()){
                print(JSON.stringify(cursor.next(),null,2));
        }
        sleep(1000);
    }
}
```

`find` 方法中使用了 `tailable` 选项，这个选项表示采用读取**游标**的方式，**如果没有新的数据，那么就会阻塞等待**，直到有新的数据插入。类似 Linux 中的 `tail -f` 命令。一旦发现新的数据 `cursor.hasNext()` 就会返回 `true`，然后调用 `cursor.next()` 方法获取新的数据。