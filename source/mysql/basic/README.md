---
title: 关系型数据基础
---
# 关系型数据基础
## 基本概念
### 表
数据库可以理解是一个文件柜。此文件柜是一个存放数据的物理位置，不管数据是什么以及如何组织的。
在将资料放入文件柜时，并不是随便将它们扔进某个抽屉就完事了，而是在文件柜中创建文件，然后将相关的资料放入特定的文件中。
这种文件称为**表**。是一种结构化的文件，可用来存储某种特定类型的数据。

### 列和数据类型
列是表中的一个字段。所有表都是由一个或多个列组成的。

数据类型（datatype）所容许的数据的类型。每个表列都有相应的数据类型，它限制（或容许）该列中存储的数据。

### 行
行是表中的一个记录。

### 主键
主键是表中的一列（或一组列），其值能够唯一区分表中每个行。

表中的任何列都可以作为主键，只要它满足以下条件：
- **任意两行都不具有相同的主键值**
- **每个行都必须具有一个主键值**（主键列不允许`NULL`值）。

主键通常定义在表的一列上，但这并不是必需的。

### 外键
外键为某个表中的一列，它包含另一个表的主键值，定义了两个表之间的关系。

### 索引
索引是对数据库表中一列或多列的值进行排序的一种结构，使用索引可快速访问数据库表中的特定信息。

#### 索引的原理：
- 对要查询的字段建立索引其实就是把该字段按照一定的方式排序
- 建立的索引只对该字段有用，如果查询的字段改变，那么这个索引也就无效了，比如图书馆的书是按照书名的第一个字母排序的，那么你想要找作者叫张
三的书就不能用这个索引了

#### 索引是优缺点
首先明白为什么索引会增加速度，DB 在执行一条 SQL 语句的时候，默认的方式是根据搜索条件进行全表扫描，遇到匹配条件的就加入搜索结果集合。
如果对某一字段增加索引，查询时就会先去索引列表中一次定位到特定值的行数，大大减少遍历匹配的行数，所以能明显增加查询的速度。不应该加索引
的场景：
- 如果每次都需要取到所有表记录，无论如何都必须进行全表扫描了，那么是否加索引也没有意义了。
- 对非唯一的字段，例如“性别”这种大量重复值的字段，增加索引也没有什么意义。
- 对于记录比较少的表，增加索引不会带来速度的优化反而浪费了存储空间，因为索引是需要存储空间的，而且有个致命缺点是
对于 `update/insert/delete` 的每次执行，字段的索引都必须重新计算更新。

### 可伸缩性
可伸缩性（scale）能够适应不断增加的工作量而不失败。设计良好的数据库或应用程序称之为可伸缩性好（scale well）。

## MySQL 简单查询

**注意每条语句后面都要以`;`结尾。SQL语句是不区分大小写的。**

## `USE` 选择数据库

## `SHOW`
- `SHOW DATABASES;`，查看数据库列表。
- `SHOW TABLES;`，查看数据库中的表。
- `SHOW COLUMNS`，显示某个表中的列，比如 `SHOW COLUMNS FROM users`。也可以使用 `DESCRIBE users`，效果
和 `SHOW COLUMNS FROM users` 是一样的。
- `SHOW STATUS`，用于显示广泛的服务器状态信息。
- `SHOW CREATE DATABASE` 和 `SHOW CREATE TABLE`，分别用来显示创建特定数据库或表的语句。
- `SHOW GRANTS`，用来显示授予用户（所有用户或特定用户）的安全权限。
- `SHOW ERRORS` 和 `SHOW WARNINGS`，用来显示服务器错误或警告消息。
- `HELP SHOW;` 查看 `SHOW` 的用法

## `SELECT`
为了使用 `SELECT` 所搜表数据，必须至少给出两条信息——**想选择什么，以及从什么地方选择**。比如`select name from users;`，会找
出 `users` 表中的所有 `name` 列。

检索多个列：
```sql
select name, age, phone from users;
```

检索所有列，使用星号 `*` 通配符：
```sql
select * from users;
```

### `DISTINCT`
`DISTINCT` 关键字用来去重。比如下面的语句，只会返回 `name` 不同的用户：
```sql
select distinct name from users;
```

> `DISTINCT` 关键字应用于所有列而不仅是前置它的列。如果给出 `SELECT DISTINCT vend_id, prod_price`，会分别作用于了 `vend_id` 
和 `prod_price` 列。

### `LIMIT`
`SELECT` 语句返回所有匹配的行，为了返回第一行或前几行，可使用 `LIMIT` 子句。
```sql
select name from users limit 5;
```
上面的语句最多返回5行。

```sql
select name from users limit 5,5;
```

上面的语句的意思是从第五行开始，最多返回 5 行。

> **MySQL 5 支持 LIMIT 的另一种替代语法。`LIMIT 4 OFFSET 3` 意为从行 3 开始取 4 行，就像 `LIMIT 3, 4` 一样**。

### 完全限定的表名和列名
```sql
select users.name from demo.users;
```
上面的语句和 `select name from users` 没什么区别。但是有一些情形需要完全限定名。比如在涉及外部子查询的语句中，会使用完全限定列名，
避免列名可能存在多义性。

> **完全限定列名**在引用的列可能出现二义性时，必须使用完全限定列名（用一个点分隔的表名和列名）。

### `ORDER BY`
使用 `ORDER BY` 子句对输出进行排序。比如 `select name from users order by age`，按照 `age` 排序。

#### 按多个列排序
```sql
select name from users order by age, weight;
```

会先按照 `age` 排序，如果有多行有相同的 `age`，再按照 `weight` 排序。

#### 排序方向
升序排序（从 A 到 Z）是默认的排序顺序，如果要进行降序排序，需要指定 `DESC` 关键字。
```sql
select name, age from users order by age desc;
```

结果会是
```
+--------+------+
| name   | age  |
+--------+------+
| ming   | 18   |
| qiang  | 17   |
| liang  | 16   |
| long   | 16   |
+--------+------+
```

**多个列降序排序**：
```sql
select name from users order by age desc, weight;
```

上面的示例，对 `age` 列降序排序，`weight` 还是默认的升序排序。可以看出 **`DESC` 关键字只会作用到其前面的列**。

> **如果想在多个列上进行降序排序，必须对每个列指定 `DESC` 关键字。**
> **`ASC`** 是升序排列的关键字，没什么用，因为默认就是升序。

### 过滤数据
通常数据库检索数据都会指定过滤条件（filter condition）。

在 `SELECT` 语句中，数据根据 `WHERE` 子句中指定的搜索条件进行过滤。

```sql
select name from users where age = 18;

select name from users where name = 'ming';
```

会找到 `age` 等于 18 的行。

> **如果同时使用 `ORDER BY` 和 `WHERE` 子句时，`ORDER BY` 必须位于 `WHERE` 之后，否则将会产生错误。**

#### `WHERE`条件操作符

| 操作符 | 描述 |
| --- |  --- |
| `=` | 等于  |
| `<>` | 不等于 |
| `!=` | 不等于 |
| `<`  | 小于   |
| `<=` |  小于等于 |
| `>`  | 大于   |
| `>=` | 大于等于 |
| `BETWEEN` | 在指定的两个值之间 |


**`BETWEEN` 操作符使用**：
```sql
select name from users where age between 15 and 18;
```

检索年龄在 15 和 18 之间的用户。

`AND` 关键字前后分别是开始值和结束值。查询到的数据**包括指定的开始值和结束值**。

#### 空值检查
创建表时，列的值可以为**空值 `NULL`**。

`IS NULL` 子句可用来检查具有 `NULL` 值的列。
```sql
select name from users where phone IS NULL;
```

#### AND
可使用 `AND` 操作符给 `WHERE` 子句附加条件。
```sql
select name from users where age = 18 and weight = 60;
```

#### OR
和 `AND` 差不多，只不过是匹配任一满足的条件。

#### 操作符优先级
```sql
select name from users where age = 18 or agr = 19 and weight >= 60;
```
上面的语句是什么结果？
是找出年龄是 18 或者 19，体重在 60 以上的行？并不是。

SQL在处理 `OR` 操作符前，优先处理 `AND` 操作符。当 SQL 看到上述 `WHERE` 子句时，它理解为由年龄为 19，并且体重在 60 以上的用户，或者
年龄为 18，不管体重多少的用户。

正确的语法：
```sql
select name from users where (age = 18 or agr = 19) and weight >= 60;
```

**SQL 会首先过滤圆括号内的条件**。

#### IN
`IN` 操作符用来指定条件范围，匹配圆括号中的值。
```sql
select name from users where age in (18, 19);
```

`IN` 操作符与 `OR` 有相同的功能。

#### `NOT`
`NOT` 操作否定条件。
```sql
select name from users where age not in (18, 19);
```

#### `LIKE`
如果要使用通配符，需要使用 `LIKE` 操作符。

**通配符可在搜索模式中任意位置使用，并且可以使用多个通配符**。

##### `%`
`%` 表示任何字符出现任意次数。例如找出所有以词 `m` 开头的用户：
```sql
select name from users where name like 'm%';
```
使用两个通配符，表示匹配任何位置包含文本 `in` 的值：
```sql
select name from users where name like '%in%';
```

##### `_`
下划线 `_` 的用途与 `%` 一样，但下划线只匹配单个字符而不是多个字符。

#### 使用正则表达式
```sql
select name from users where name REGEX 'ing';
```

`REGEXP` 后跟的是**正则表达式**。

正则表达式并没有什么优势，但是有些场景下可以考虑使用:
```sql
select name from users where weight REGEX '.6';
```

`.` 是正则表达式语言中一个特殊的字符。它表示匹配任意一个字符，因此，`56` 和 `66` 都匹配且返回。

##### OR 匹配
```sql
select name from users where weight REGEX '46|56';
```
`|` 为正则表达式的 `OR` 操作符。

##### 匹配几个字符之一
想匹配特定的字符可以指定一组用 `[` 和 `]` 括起来的字符来完成。
```sql
select name from users where weight REGEX '[456]6';
```
`[456]` 表示匹配 4，5，6。`[]` 是另一种形式的 `OR` 语句。可以使用一些正则的语法例如 `[^123]`，匹配除了 1，2，3 之外的字符，
`[1-9]` 匹配 1 到 9 范围的字符。匹配特殊字符钱加 `\\` 比如 `.` 要用 `\\.` 来查找。


## 计算字段
存储在数据库表中的数据一般不是应用程序所需要的格式。

### 拼接字段
`Concat()` 拼接串，即把多个串连接起来形成一个较长的串。`Concat()` 需要一个或多个指定的串，各个串之间用逗号分隔。

```sh
mysql> select Concat(c1, '(', c2, ')') from record_format_demo;
+--------------------------+
| Concat(c1, '(', c2, ')') |
+--------------------------+
| aaaa(bbb)                |
| eeee(fff)                |
+--------------------------+
2 rows in set (0.01 sec)
```

`RTrim()` 函数可以删除数据右侧多余的空格，还有 `LTrim()` 和 `Trim()`，分别是删除左边空格和删除左右空格。

### 别名
别名（alias）是一个字段或值的替换名。用 `AS` 关键字：
```sh
mysql> select Concat(c1, '(', c2, ')') as c5 from record_format_demo;
+-----------+
| c5        |
+-----------+
| aaaa(bbb) |
| eeee(fff) |
+-----------+
2 rows in set (0.01 sec)
```


### 算术计算
```sql
select price, name, quantity from orders where order_num = 2005;
```

`price` 是物品的价格，`quantity` 是数量，如果想汇总物品总价;
```sql
select price, name, quantity, quantity*price as total_price from orders where order_num = 2005;
```
`total_price` 就是总价。

**MySQL 支持基本算术操作符 `+`，`-`，`*`，`/`。此外，圆括号可用来区分优先顺序**。


## 数据处理函数
`Trim()` 就是一个数据处理函数。

### 文本函数
- `Upper()` 函数将文本转换为大写，`Upper(name) as newName`
- `Left()`  返回串左边的字符
- `Length()` 返回串的长度
- `Locate()` 找出串的一个子串
- `Lower()` 将串转换为小写
- `LTrim()` 去掉串左边的空格
- `Right()` 返回串右边的字符
- `RTrim()` 去掉串右边的空格
- `Soundex()` 返回串的 `SOUNDEX` 值
- `SubString()` 返回子串的字符

`SOUNDEX` 是一个将任何文本串转换为描述其语音表示的字母数字模式的算法。

### 日期和时间处理函数
- `AddDate()` 增加一个日期（天、周等）
- `AddTime()` 增加一个时间（时、分等）
- `CurDate()` 返回当前日期
- `CurTime()` 返回当前时间
- `Date()` 返回日期时间的日期部分
- `DateDiff()` 计算两个日期之差
- `Date_Add()` 高度灵活的日期运算函数
- `Date_Format()` 返回一个格式化的日期或时间串
- `Day()` 返回一个日期的天数部分
- `DayOfWeek()` 对于一个日期，返回对应的星期几
- `Hour()` 返回一个时间的小时部分
- `Minute()` 返回一个时间的分钟部分
- `Month()` 返回一个日期的月份部分
- `Now()` 返回当前日期和时间
- `Second()` 返回一个时间的秒部分
- `Time()` 返回一个日期时间的时间部分
- `Year()` 返回一个日期的年份部分

### 数值函数
- `Abs()` 返回一个数的绝对值
- `Cos()` 返回一个角度的余弦
- `Exp()` 返回一个数的指数值
- `Mod()` 返回除操作的余数
- `Pi()` 返回圆周率
- `Rand()` 返回一个随机数
- `Sin()` 返回一个角度的正弦
- `Sqrt()` 返回一个数的平方根
- `Tan()` 返回一个角度的正切


## 聚合函数
### AVG
`AVG` 函数可用来返回所有列的平均值，也可以用来返回特定列或行的平均值。
```sql
select AVG(price) as avg_price from orders;
```
返回订单的平均价格。

**`AVG()` 函数忽略列值为`NULL` 的行**。

### COUNT
`COUNT(*)` 对表中行的数目进行计数。`COUNT(column)` **对特定列中具有值的行进行计数，忽略 `NULL` 值**。

### MAX
返回指定列中的最大值。忽略列值为 `NULL` 的行。例如 `select max(price) as max_price from products;` 返回 products 表中
最贵的物品的价格。

### MIN
与 `MAX()` 功能相反。

### SUM
`SUM()` 函数返回指定列值的和。忽略列值为 `NULL` 的行。例如 `select SUM(item_price*quantity) as total_price from products;`

### 聚合不同值
上面的几个函数都可以使用 `DISTINCT`，比如 `AVG(DISTINCT price) as avg_total_price`

**`DISTINCT` 只能用于 `COUNT()`。`DISTINCT` 不能用于 `COUNT(*)`**。

## 分组
`GROUP BY` 子句用来创建分组。
```sql
select vend_id, COUNT(*) as prod_num from products group by vend_id;
```
上面的语句按 `vend_id` 排序并分组数据。

注意：
- `GROUP BY` 句必须出现在 **`WHERE` 子句之后，`ORDER BY` 子句之前**。
- 如果分组列中具有 `NULL` 值，则 `NULL` 将作为一个分组返回。如果列中有多行 `NULL` 值，它们将分为一组。
- `GROUP BY` 子句中列出的每个列都必须是检索列或有效的表达式（但不能是聚集函数）。如果在 `SELECT` 中使用表达式，则必须在 `GROUP BY` 
子句中指定相同的表达式。不能使用别名。
- 除了聚集计算语句，`SELECT` 语句中的每个列都必须在 `GROUP BY` 子句中给出。

### 过滤分组
`HAVING` 子句**过滤分组**。`HAVING` 非常类似于 `WHERE`（`WHERE` 过滤的是**行**）。它们的句法是相同的，只是关键字有差别。

也可以这么理解：**`WHERE` 在数据分组前进行过滤，`HAVING` 在数据分组后进行过滤**。

```sql
select vend_id, COUNT(*) as prod_num from products group by vend_id having COUNT(*) >= 2;
```

它过滤 `COUNT(*) >=2` 的那些分组。

### 分组和排序
`GROUP BY` 和 `ORDER BY` 的差别：

| `order by` | `group by` |
| ------  | --------- |
| 排序产生的输出 | 分组行。但输出可能不是分组的顺序 |
| 任意列都可以使用（甚至非选择的列也可以使用） | 只可能使用选择列或表达式列，而且**必须使用每个选择列表达式** |
| 不一定需要 | 如果与聚集函数一起使用列（或表达式），则必须使用 |

> 一般在使用 `GROUP BY` 子句时，应该也给出 `ORDER BY` 子句。这是保证数据正确排序的唯一方法。千万不要仅依赖 `GROUP BY` 排序数据。

检索总计订单价格大于等于 50 的订单的订单号和总计订单价格：
```sql
select order_num, SUM(quantity*price) as order_total from orders group by order_num having SUM(quantity*price) >= 50;
```

按总计订单价格排序输出：
```sql
select order_num, SUM(quantity*price) as order_total from orders group by order_num having SUM(quantity*price) >= 50 order by order_total;
```

## SELECT子句顺序
| 子句 | 是否必须使用 |
| --- | ---- |
| `SELECT` | 是 |
| `FROM` | 仅在从表选择数据时使用 |
| `WHERE`| 否 |
| `GROUP BY` | 仅在按组计算聚集时使用 |
| `HAVING` | 否 |
| `ORDER BY` | 否 |
| `LIMIT` | 否 |

