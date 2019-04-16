# 字符集和比较规则
## 字符集和比较规则简介
### 字符集简介
计算机中只能存储二进制数据，那该怎么存储字符串呢？当然是建立字符与二进制数据的映射关系了，建立这个关系最起码要搞清楚两件事儿：
- 你要把哪些字符映射成二进制数据？也就是界定清楚字符范围。
- 怎么映射？将一个字符映射成一个二进制数据的过程也叫做**编码**，将一个二进制数据映射到一个字符的过程叫做**解码**。

**字符集**就是来描述某个字符范围的编码规则。

### 比较规则简介
怎么比较两个字符？最容易的就是直接比较这两个字符对应的二进制编码的大小。二进制比较规则是简单，但有时候并不符合现实需求，比如在很多场合对于英文字符我们都是不区分大小写的。

### 一些重要的字符集
- `ASCII`字符集，共收录128个字符，包括空格、标点符号、数字、大小写字母和一些不可见字符。
- `ISO 8859-1`字符集，共收录256个字符，是在`ASCII`字符集的基础上又扩充了128个西欧常用字符。别名`latin1`。
- `GB2312`字符集，收录了汉字以及拉丁字母、希腊字母、日文平假名及片假名字母、俄语西里尔字母。其中收录汉字6763个，其他文字符号682个。同时这种字符集又兼容ASCII字符集。
- `GBK`字符集，GBK字符集只是在收录字符范围上对`GB2312`字符集作了扩充，编码方式上兼容`GB2312`。
- `utf8`字符集，收录地球上能想到的所有字符，而且还在不断扩充。这种字符集兼容`ASCII`字符集，采用变长编码方式，编码一个字符需要使用1～4个字节。

> `utf8`只是`Unicode`字符集的一种编码方案，`Unicode`字符集可以采用`utf8`、`utf16`、`utf32`这几种编码方案，`utf8`使用1～4个字节编码一个字符，`utf16`使用2个或4个字节编码一个字符，
`utf32`使用4个字节编码一个字符。


## MySQL中支持的字符集和排序规则
### MySQL中的utf8和utf8mb4
`utf8`字符集表示一个字符需要使用1～4个字节，但是常用的一些字符使用1～3个字节就可以表示了。

而在MySQL中字符集表示一个字符所用最大字节长度在某些方面会影响系统的存储和性能，所以MySQL定义了两个概念：
- `utf8mb3`：阉割过的`utf8`字符集，只使用1～3个字节表示字符。
- `utf8mb4`：正宗的`utf8`字符集，使用1～4个字节表示字符。

**在MySQL中`utf8`是`utf8mb3`的别名，所以之后在MySQL中提到`utf8`就意味着使用1~3个字节来表示一个字符，如果大家有使用4字节编码一个字符的情况，
比如存储一些emoji表情啥的，那请使用`utf8mb4`**。

### 字符集的查看
查看当前MySQL中支持的字符集`SHOW (CHARACTER SET|CHARSET) [LIKE 匹配的模式];`。`CHARACTER SET`和`CHARSET`是同义词，用任意一个都可以。
查看MySQL中支持的比较规则`SHOW COLLATION [LIKE 匹配的模式];`。

**每种字符集对应若干种比较规则，每种字符集都有一种默认的比较规则**。

## 字符集和比较规则的应用
MySQL有4个级别的字符集和比较规则，分别是：
- 服务器级别
- 数据库级别
- 表级别
- 列级别

### 服务器级别
MySQL提供了两个系统变量来表示服务器级别的字符集和比较规则：
- `character_set_server	`服务器级别的字符集
- `collation_server`服务器级别的比较规则

### 数据库级别
创建和修改数据库的时候可以指定该数据库的字符集和比较规则，具体语法如下：
```sql
CREATE DATABASE 数据库名
    [[DEFAULT] CHARACTER SET 字符集名称]
    [[DEFAULT] COLLATE 比较规则名称];

ALTER DATABASE 数据库名
    [[DEFAULT] CHARACTER SET 字符集名称]
    [[DEFAULT] COLLATE 比较规则名称];
```
`DEFAULT`可以省略，并不影响语句的语义。比方说我们新创建一个名叫`charset_demo_db`的数据库，在创建的时候指定它使用的字符集为`gb2312`，比较规则为`gb2312_chinese_ci`：
```sql
mysql> CREATE DATABASE charset_demo_db
    -> CHARACTER SET gb2312
    -> COLLATE gb2312_chinese_ci;
Query OK, 1 row affected (0.01 sec)
```

如果想查看当前数据库使用的字符集和比较规则，可以查看下面两个系统变量的值（前提是使用`USE`语句选择当前默认数据库，如果没有默认数据库，
则变量与相应的服务器级系统变量具有相同的值）：
- `character_set_database`当前数据库的字符集
- `collation_database`当前数据库的比较规则

> **`character_set_database`和`collation_database`这两个系统变量是只读的，我们不能通过修改这两个变量的值而改变当前数据库的字符集和比较规则**。

如果数据库的创建或修改语句中**没有指定字符集和比较规则**，那么会**使用服务器级别的字符集和比较规则作为数据库的字符集和比较规则**。

### 表级别
可以在创建和修改表的时候指定表的字符集和比较规则，语法如下：
```sql
CREATE TABLE 表名 (列的信息)
    [[DEFAULT] CHARACTER SET 字符集名称]
    [COLLATE 比较规则名称]]

ALTER TABLE 表名
    [[DEFAULT] CHARACTER SET 字符集名称]
    [COLLATE 比较规则名称]
```
比方说我们创建一个名为`t`的表，并指定这个表的字符集和比较规则：
```sql
mysql> CREATE TABLE t(
    ->     col VARCHAR(10)
    -> ) CHARACTER SET utf8 COLLATE utf8_general_ci;
Query OK, 0 rows affected (0.03 sec)
```

如果表的创建或修改语句中**没有指定字符集和比较规则**，那么会**使用该表所在数据库的字符集和比较规则作为该表的字符集和比较规则**。

### 列级别
**同一个表中的不同的列也可以有不同的字符集和比较规则**。在创建和修改列定义的时候可以指定该列的字符集和比较规则，语法如下：
```sql
CREATE TABLE 表名(
    列名 字符串类型 [CHARACTER SET 字符集名称] [COLLATE 比较规则名称],
    其他列...
);

ALTER TABLE 表名 MODIFY 列名 字符串类型 [CHARACTER SET 字符集名称] [COLLATE 比较规则名称];
```
比如我们修改一下表`t`中列`col`的字符集和比较规则可以这么写：
```sql
mysql> ALTER TABLE t MODIFY col VARCHAR(10) CHARACTER SET gbk COLLATE gbk_chinese_ci;
Query OK, 0 rows affected (0.04 sec)
Records: 0  Duplicates: 0  Warnings: 0
```

对于列，如果表的创建或修改语句中**没有指定该列的字符集和比较规则**，那么会**使用该列所在表的字符集和比较规则**。

> 在转换列的字符集时需要注意，如果转换前列中存储的数据不能用转换后的字符集进行表示会发生错误。比方说原先列使用的字
符集是`utf8`，列中存储了一些汉字，现在把列的字符集转换为`ascii`的话就会出错，因为`ascii`字符集并不能表示汉字字符。

### 仅修改字符集或仅修改比较规则
由于字符集和比较规则是互相有联系的，如果我们只修改了字符集，比较规则也会跟着变化，如果只修改了比较规则，字符集也会跟着变化，具体规则如下：
- 只修改字符集，则比较规则将变为修改后的字符集默认的比较规则。
- 只修改比较规则，则字符集将变为修改后的比较规则对应的字符集。

### 客户端和服务器通信中的字符集
#### 编码和解码使用的字符集不一致的后果
比如字符`'我'`在`utf8`字符集编码下的字节串长这样：`0xE68891`，把这个字节串发送到另一个程序里，另一个程序用不同的字符集去解码这个字节串，
假设使用的是`gbk`字符集来解释这串字节，解码过程就是这样的：
1. 首先看第一个字节`0xE6`，它的值大于`0x7F`（十进制：127），说明是两字节编码，继续读一字节后是`0xE688`，然后从`gbk`编码表中查找字节为`0xE688`对应的字符，发现是字符`'鎴'`
2. 继续读一个字节`0x91`，它的值也大于`0x7F`，再往后读一个字节发现木有了，所以这是半个字符。
3. 所以`0xE68891`被`gbk`字符集解释成一个字符`'鎴'`和半个字符。

#### 字符集转换
如果接收`0xE68891`这个字节串的程序按照`utf8`字符集进行解码，然后又把它按照`gbk`字符集进行编码，最后编码后的字节串就是`0xCED2`，我们把这个过程称为字符集的转换，
也就是字符串`'我'`从`utf8`字符集转换为`gbk`字符集。

#### MySQL中字符集的转换
从客户端发往服务器的请求本质上就是一个字符串，服务器向客户端返回的结果本质上也是一个字符串，而字符串其实是使用某种字符集编码的二进制数据。这个字符串可不是使用一种字符集的
编码方式一条道走到黑的，从发送请求到返回结果这个过程中伴随着多次字符集的转换，在这个过程中会用到3个系统变量：
- `character_set_client`服务器解码请求时使用的字符集
- `character_set_connection	`服务器处理请求时会把请求字符串从`character_set_client`转为`character_set_connection`
- `character_set_results`服务器向客户端返回数据时使用的字符集

这三个系统变量的值可能默认都是`utf-8`。为了体现出字符集在请求处理过程中的变化，这里特意修改一个系统变量的值：
```sh
mysql> set character_set_connection = gbk;
Query OK, 0 rows affected (0.00 sec)
```

所以现在系统变量`character_set_clien`t和`character_set_results`的值还是`utf8`，而`character_set_connection`的值为`gbk`。现在假设我们客户端发送的请求是下边这个字符串：
```sql
SELECT * FROM t WHERE s = '我';
```
我们只分析字符`'我'`在这个过程中字符集的转换。
现在看一下在请求从发送到结果返回过程中字符集的变化：
1. 客户端发送请求所使用的字符集
一般情况下客户端所使用的字符集和当前操作系统一致，不同操作系统使用的字符集可能不一样，如下：
- 类`Unix`系统使用的是`utf8`
- `Windows`使用的是`gbk`

例如在使用的`macOS`操作系统时，客户端使用的就是`utf8`字符集。所以字符`'我'`在发送给服务器的请求中的字节形式就是：`0xE68891`

2. 服务器接收到客户端发送来的请求其实是一串二进制的字节，它会认为这串字节采用的字符集是`character_set_client`，然后把这串字节转换为`character_set_connection`字符集编码的字符。
由于计算机上`chacharacter_set_client`的值是`utf8`，首先会按照`utf8`字符集对字节串`0xE68891`进行解码，得到的字符串就是`'我'`，然后按照`character_set_connection`
代表的字符集，也就是`gbk`进行编码，得到的结果就是字节串`0xCED2`。

3. 因为表`t`的列`col`采用的是`gbk`字符集，与`character_set_connection`一致，所以直接到列中找字节值为`0xCED2`的记录，最后找到了一条记录。
**如果某个列使用的字符集和`character_set_connection`代表的字符集不一致的话，还需要进行一次字符集转换**。

4. 上一步骤找到的记录中的`col`列其实是一个字节串`0xCED2`，`col`列是采用`gbk`进行编码的，所以首先会将这个字节串使用`gbk`进行解码，
得到字符串`'我'`，然后再把这个字符串使用`character_set_results`代表的字符集，也就是`utf8`进行编码，得到了新的字节串：`0xE68891`，然后发送给客户端。

5. 由于客户端是用的字符集是`utf8`，所以可以顺利的将`0xE68891`解释成字符我，从而显示到我们的显示器上，所以我们人类也读懂了返回的结果。

几点需要注意的地方：
- 假设你的客户端采用的字符集和`character_set_client`不一样的话，这就会出现意想不到的情况。
- 假设你的客户端采用的字符集和`character_set_results`不一样的话，这就可能会出现客户端无法解码结果集的情况


> **通常都把`character_set_client`、`character_set_connection`、`character_set_results`这三个系统变量设置成和客户端使用的字符集一致的情况，这样减少了很多无谓的字符集转换**。

MySQL提供了一条非常简便的语句:
```sql
SET NAMES 字符集名;
```
效果与下面的三条语句的执行效果一样：
```sql
SET character_set_client = 字符集名;
SET character_set_connection = 字符集名;
SET character_set_results = 字符集名;
```

也可以写到配置文件里：
```
[client]
default-character-set=utf8
```

> 如果你使用的是`Windows`系统，那应该设置成`gbk`。

#### 比较规则的应用
`比较规则`的作用通常体现在**比较字符串大小的表达式**以及**对某个字符串列进行排序**中，所以有时候也称为**排序规则**。比方说表`t`的列`col`使用的字符集是`gbk`，
使用的比较规则是`gbk_chinese_ci`，我们向里边插入几条记录：
```sh
mysql> INSERT INTO t(col) VALUES('a'), ('b'), ('A'), ('B');
Query OK, 4 rows affected (0.00 sec)
Records: 4  Duplicates: 0  Warnings: 0

mysql>
```
查询的时候按照`col`列排序一下：
```sh
mysql> SELECT * FROM t ORDER BY col;
+------+
| col  |
+------+
| a    |
| A    |
| b    |
| B    |
| 我   |
+------+
5 rows in set (0.00 sec)
```

可以看到在默认的比较规则`gbk_chinese_ci`中是不区分大小写的，我们现在把列`col`的比较规则修改为`gbk_bin`：
```sh
mysql> ALTER TABLE t MODIFY col VARCHAR(10) COLLATE gbk_bin;
Query OK, 5 rows affected (0.02 sec)
Records: 5  Duplicates: 0  Warnings: 0
```

`gbk_bin`是直接比较字符的编码，所以是区分大小写的，我们再看一下排序后的查询结果：
```sh
mysql> SELECT * FROM t ORDER BY col;
+------+
| s    |
+------+
| A    |
| B    |
| a    |
| b    |
| 我   |
+------+
5 rows in set (0.00 sec)

mysql>
```