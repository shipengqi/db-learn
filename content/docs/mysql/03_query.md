---
title: 简单查询语句
---


**SQL 语句要以 `;` 结尾。不区分大小写。**

## `use` 选择数据库

## `show`

- `show databases;`，查看数据库列表。
- `show tables;`，查看数据库中的表。
- `show colums`，显示某个表中的列，比如 `show colums from users`。也可以使用 `describe users`，效果一样。
- `show status`，用于显示广泛的服务器状态信息。
- `show create database` 和 `show create table`，分别用来显示创建特定数据库或表的语句。
- `show grants`，用来显示授予用户（所有用户或特定用户）的安全权限。
- `show erros` 和 `show warnings`，用来显示服务器错误或警告消息。
- `help show;` 查看 `show` 的用法

## `select`

为了使用 `select` 查询语句，如 `select name from users;`，找出 `users` 表中的所有 `name` 列。

查询多个列：

```sql
select name, age, phone from users;
```

查询所有列，使用星号 `*`

### `distinct`

`distinct` 去重。

```sql
select distinct name from users;
```

上面的示例，会对 `name` 去重，返回 `name` 不同的行

> `distinct` 关键字，会对其后面的所有列去重。如 `select distinct vend_id, prod_price from products`，`vend_id` 和 `prod_price` 列。

### `limit`

使用 `limit` 子句，限制返回的行数。

```sql
select name from users limit 5;
```

上面的语句最多返回 5 行。

```sql
select name from users limit 5,5;
```

上面的语句的意思是从第 5 行开始，最多返回 5 行。

> **MySQL 5 支持 limit 的另一种替代语法。`limit 4 offset 3` 表示从第 3 行开始取 4 行，和 `limit 3, 4` 一样**。

### 限定的表名和列名

```sql
select users.name from demo.users;
```

上面的语句和 `select name from users` 没什么区别。但是有一些情形需要限定名。比如 join 多个表时，都包含同样的列名，就可以使用限定表名。

### `order by`

使用 `order by` 子句对输出进行排序。比如 `select name from users order by age`，按照 `age` 排序。

#### 按多个列排序

```sql
select name from users order by age, weight;
```

会先按照 `age` 排序，如果有多行有相同的 `age`，再按照 `weight` 排序。

#### 排序方向

查询默认是**升序排序**（`asc`），如果要进行降序排序，使用 `desc` 关键字。

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

上面的示例，对 `age` 列降序排序，`weight` 还是默认的升序排序。

**`desc` 关键字只会作用到其前面的列**。

**如果想在多个列上进行降序排序，必须对每个列指定 `desc` 关键字。**

### `where`

查询时使用 `where` 子句中指定过滤条件。

```sql
select name from users where age = 18;

select name from users where name = 'ming';
```

> **注意 `order by` 必须位于 `where` 之后。**

#### `where` 条件操作符

| 操作符 | 描述 |
| --- |  --- |
| `=` | 等于  |
| `<>`, `!=` | 不等于 |
| `<`  | 小于   |
| `<=` |  小于等于 |
| `>`  | 大于   |
| `>=` | 大于等于 |
| `between` | 在指定的两个值之间 |

**`between` 操作符使用**：

```sql
select name from users where age between 15 and 18;
```

检索年龄在 15 和 18 之间的用户。

`and` 关键字前后分别是开始值和结束值。查询到的数据**包括指定的开始值和结束值**。

#### 空值检查

创建表时，列的值可以为**空值 `NULL`**。

`IS NULL` 子句可用来检查具有 `NULL` 值的列。

```sql
select name from users where phone IS NULL;
```

#### and

`and` 添加过滤条件。

```sql
select name from users where age = 18 and weight = 60;
```

#### or

和 `and` 差不多，只不过是匹配任一满足的条件。

#### 操作符优先级

```sql
select name from users where age = 18 or agr = 19 and weight >= 60;
```

上面的语句是什么结果？

是找出年龄是 18 或者 19，体重在 60 以上的行？并不是。

SQL 在处理 `or` 操作符前，优先处理 `and` 操作符。当 SQL 看到上述 `where` 子句时，它理解为由年龄为 19，并且体重在 60 以上的用户，或者
年龄为 18，不管体重多少的用户，相当于 `age = 18 or (agr = 19 and weight >= 60)`。

正确的语法：

```sql
select name from users where (age = 18 or agr = 19) and weight >= 60;
```

**SQL 会首先过滤圆括号内的条件**。

#### in

`in` 操作符用来指定条件范围，匹配圆括号中的值。

```sql
select name from users where age in (18, 19);
```

`in` 操作符与 `or` 有相同的功能。

#### `not`

`not` 操作否定条件。

```sql
select name from users where age not in (18, 19);
```

#### `like`

通配符，使用 `like` 操作符。

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

**`_` 匹配单个字符，`%` 匹配多个字符**。

#### 使用正则表达式

```sql
select name from users where name regexp 'ing';
```

`regexp` 后面跟 **正则表达式**。

正则表达式并没有什么优势，但是有些场景下可以考虑使用:

```sql
select name from users where weight regexp '.6';
```

`.` 是正则表达式语言中一个特殊的字符。它表示匹配任意一个字符，因此，`56` 和 `66` 都匹配且返回。

##### or 匹配

```sql
select name from users where weight regexp '46|56';
```

`|` 为正则表达式的 `or` 操作符。

##### 匹配几个字符之一

想匹配特定的字符可以指定一组用 `[` 和 `]` 括起来的字符来完成。

```sql
select name from users where weight REGEX '[456]6';
```

`[456]` 表示匹配 4，5，6。`[]` 是另一种形式的 `or` 语句。可以使用一些正则的语法例如 `[^123]`，匹配除了 1，2，3 之外的字符，
`[1-9]` 匹配 1 到 9 范围的字符。匹配特殊字符钱加 `\\` 比如 `.` 要用 `\\.` 来查找。

## 计算字段

存储在数据库表中的数据一般不是应用程序所需要的格式。

### 拼接字段

`concat()` 拼接串，多个串之间用 `,` 分隔。

```sh
mysql> select concat(c1, '(', c2, ')') from record_format_demo;
+--------------------------+
| Concat(c1, '(', c2, ')') |
+--------------------------+
| aaaa(bbb)                |
| eeee(fff)                |
+--------------------------+
2 rows in set (0.01 sec)
```

### 别名

别名（alias） `as` 关键字：

```sh
mysql> select concat(c1, '(', c2, ')') as c5 from record_format_demo;
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
select price, name, quantity, quantity*price as total_price from orders where order_num = 2005;
```

`price` 是物品的价格，`quantity` 是数量， `total_price` （quantity*price）就是总价。

**基本算术操作符 `+`，`-`，`*`，`/`。`()` 用来区分优先顺序**。

## 数据处理函数

### 文本函数

- `upper()` 文本转大写，`upper(name) as newName`
- `lower()` 文本转小写
- `left()`  返回串左边的字符
- `right()` 返回串右边的字符
- `length()` 返回串的长度
- `locate()` 找出串的一个子串
- `trim()`，删除两边空格。
- `ltrim()` 去掉串左边的空格
- `rtrim()` 去掉串右边的空格
- `soundex()` 返回串的 `SOUNDEX` 值
- `substring()` 返回子串的字符

`SOUNDEX` 是一个将任何文本串转换为描述其语音表示的字母数字模式的算法。

### 日期和时间处理函数

- `adddate()` 增加一个日期（天、周等）
- `addtime()` 增加一个时间（时、分等）
- `curdate()` 返回当前日期
- `curtime()` 返回当前时间
- `date()` 返回日期时间的日期部分
- `datediff()` 计算两个日期之差
- `date_add()` 高度灵活的日期运算函数
- `date_format()` 返回一个格式化的日期或时间串
- `day()` 返回一个日期的天数部分
- `dayofweek()` 对于一个日期，返回对应的星期几
- `hour()` 返回一个时间的小时部分
- `minute()` 返回一个时间的分钟部分
- `month()` 返回一个日期的月份部分
- `now()` 返回当前日期和时间
- `second()` 返回一个时间的秒部分
- `time()` 返回一个日期时间的时间部分
- `year()` 返回一个日期的年份部分

### 数值函数

- `abs()` 返回一个数的绝对值
- `cos()` 返回一个角度的余弦
- `exp()` 返回一个数的指数值
- `mod()` 返回除操作的余数
- `pi()` 返回圆周率
- `rand()` 返回一个随机数
- `sin()` 返回一个角度的正弦
- `sqrt()` 返回一个数的平方根
- `tan()` 返回一个角度的正切

## 聚合函数

### avg

`avg` 函数可用来返回所有列的平均值，也可以用来返回特定列或行的平均值。

```sql
select avg(price) as avg_price from orders;
```

返回订单的平均价格。

**`avg()` 函数忽略列值为 `NULL` 的行**。

### count

`count(*)` 对行数进行计数。`count(column)` **对特定列中具有值的行进行计数，忽略 `NULL` 值**。

### max

返回指定列中的最大值。忽略列值为 `NULL` 的行。例如 `select max(price) as max_price from products;` 返回 products 表中
最贵的物品的价格。

### min

与 `max()` 功能相反。

### sum

`sum()` 函数返回指定列值的和。忽略列值为 `NULL` 的行。例如 `select sum(item_price*quantity) as total_price from products;`

### 聚合不同值

上面的几个函数都可以使用 `distinct`，比如 `avg(distinct price) as avg_total_price`

**`distinct` 只能用于 `count()`。`distinct` 不能用于 `count(*)`**。

## 分组

`group by` 创建分组。

```sql
select vend_id, count(*) as prod_num from products group by vend_id;
```

上面的语句按 `vend_id` 排序并分组数据。

注意：

- `group by` 句必须出现在 **`where` 子句之后，`order by` 子句之前**。
- 如果分组列中具有 `NULL` 值，则 `NULL` 将作为一个分组返回。
- `group by` 子句中列出的每个列都必须是检索列或有效的表达式（但不能是聚集函数）。如果在 `select` 中使用表达式，则必须在 `group by`
子句中指定相同的表达式。不能使用别名。
- 除了聚集计算语句，`select` 语句中的每个列都必须在 `group by` 子句中给出。

### 过滤分组

`having` 子句**过滤分组**。`having` 类似于 `where`，不过 `where` 过滤的是**行**。它们的句法是相同的。

**`where` 在数据分组前进行过滤，`having` 在数据分组后进行过滤**。

```sql
select vend_id, count(*) as prod_num from products group by vend_id having count(*) >= 2;
```

它过滤 `count(*) >=2` 的那些分组。

## SELECT 子句顺序

| 子句 | 是否必须使用 |
| --- | ---- |
| `select` | 是 |
| `from` | 仅在从表选择数据时使用 |
| `where`| 否 |
| `group by` | 仅在按组计算聚集时使用 |
| `having` | 否 |
| `order by` | 否 |
| `limit` | 否 |
