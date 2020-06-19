---
title: 字符集和比较规则
---

## 字符集和比较规则简介

### 字符集简介

计算机中只能存储二进制数据，如何存储字符串？当然是建立字符与二进制数据的映射关系，建立这个关系要搞清楚两件事儿：

- 要把哪些字符映射成二进制数据？也就是界定清楚字符范围。
- 怎么映射？将一个字符映射成一个二进制数据的过程也叫做**编码**，将一个二进制数据映射到一个字符的过程叫做**解码**。

**字符集**就是来描述某个字符范围的编码规则。

### 比较规则简介

怎么比较两个字符？最容易的就是直接比较这两个字符对应的二进制编码的大小。如，字符 'a' 的编码为 `0x01`，字符 'b' 的编码为 `0x02`，所
以 'a' 小于 'b'。

二进制比较规则很简单，但有时候并不符合现实需求，比如在有些场景对于英文字符不区分大小写。这时候可以这样指定比较规则：

1. 将两个大小写不同的字符全都转为大写或者小写。
2. 再比较这两个字符对应的二进制数据。

但是实际生活中的字符不止英文字符一种，比如汉字有几万之多，**同一种字符集可以有多种比较规则**。

## MySQL 中支持的字符集和排序规则

`utf8` 只是 `Unicode` 字符集的一种编码方案，`Unicode` 字符集可以采用 `utf8`、`utf16`、`utf32` 这几种编码方案，`utf8` 使用 1～4 个字节编码一个字符，`utf16` 使用 2 个或 4 个字节编码一个字符，`utf32` 使用 4 个字节编码一个字符。

### utf8 和 utf8mb4

`utf8` 字符集表示一个字符需要使用 1～4 个字节，但是常用的一些字符使用 1～3 个字节就可以表示了。

字符集表示一个字符所用最大字节长度在某些方面会影响系统的存储和性能，所以 MySQL 定义了两个概念：

- `utf8mb3`：阉割过的 `utf8` 字符集，只使用 1～3个 字节表示字符。
- `utf8mb4`：正宗的 `utf8`字符集，使用 1～4 个字节表示字符。

**MySQL 中 `utf8` 只的是 `utf8mb3`**，如果有使用 4 字节编码一个字符的情况，比如存储一些 emoji 表情啥的，要使用 `utf8mb4`。

### 字符集的查看

查看当前 MySQL 中支持的字符集 `show (character set|charset) [like 匹配的模式];`。`character set` 和 `charset` 是一个意思。

### 比较规则查看

查看 MySQL 中支持的比较规则 `show collation [like 匹配的模式];`。

查看 `utf8` 字符集下的比较规则：

```sh
mysql> show collation like 'utf8\_%';
+--------------------------+---------+-----+---------+----------+---------+
| Collation                | Charset | Id  | Default | Compiled | Sortlen |
+--------------------------+---------+-----+---------+----------+---------+
| utf8_general_ci          | utf8    |  33 | Yes     | Yes      |       1 |
| utf8_bin                 | utf8    |  83 |         | Yes      |       1 |
| utf8_unicode_ci          | utf8    | 192 |         | Yes      |       8 |
| utf8_icelandic_ci        | utf8    | 193 |         | Yes      |       8 |
| utf8_latvian_ci          | utf8    | 194 |         | Yes      |       8 |
| utf8_romanian_ci         | utf8    | 195 |         | Yes      |       8 |
| utf8_slovenian_ci        | utf8    | 196 |         | Yes      |       8 |
| utf8_polish_ci           | utf8    | 197 |         | Yes      |       8 |
| utf8_estonian_ci         | utf8    | 198 |         | Yes      |       8 |
| utf8_spanish_ci          | utf8    | 199 |         | Yes      |       8 |
| utf8_swedish_ci          | utf8    | 200 |         | Yes      |       8 |
| utf8_turkish_ci          | utf8    | 201 |         | Yes      |       8 |
| utf8_czech_ci            | utf8    | 202 |         | Yes      |       8 |
| utf8_danish_ci           | utf8    | 203 |         | Yes      |       8 |
| utf8_lithuanian_ci       | utf8    | 204 |         | Yes      |       8 |
| utf8_slovak_ci           | utf8    | 205 |         | Yes      |       8 |
| utf8_spanish2_ci         | utf8    | 206 |         | Yes      |       8 |
| utf8_roman_ci            | utf8    | 207 |         | Yes      |       8 |
| utf8_persian_ci          | utf8    | 208 |         | Yes      |       8 |
| utf8_esperanto_ci        | utf8    | 209 |         | Yes      |       8 |
| utf8_hungarian_ci        | utf8    | 210 |         | Yes      |       8 |
| utf8_sinhala_ci          | utf8    | 211 |         | Yes      |       8 |
| utf8_german2_ci          | utf8    | 212 |         | Yes      |       8 |
| utf8_croatian_ci         | utf8    | 213 |         | Yes      |       8 |
| utf8_unicode_520_ci      | utf8    | 214 |         | Yes      |       8 |
| utf8_vietnamese_ci       | utf8    | 215 |         | Yes      |       8 |
| utf8_general_mysql500_ci | utf8    | 223 |         | Yes      |       1 |
+--------------------------+---------+-----+---------+----------+---------+
27 rows in set (0.00 sec)
```

比较规则名称都是以 utf8 开头的，后边紧跟着该比较规则主要作用于哪种语言，名称后缀意味着该比较规则是否区分语言中的重音、大小写啥的：

- `_ai`，accent insensitive，不区分重音
- `_as`，accent sensitive，区分重音
- `_ci`，case insensitive，不区分大小写
- `_cs`，case sensitive，区分大小写
- `_bin`，binary，以二进制方式比较

**每种字符集对应多种种比较规则，每种字符集都有一种默认的比较规则**（`Default` 列的值为 `YES` 的）。

## 字符集和比较规则的应用

MySQL 有 4 个级别的字符集和比较规则，分别是：

- 服务器级别
- 数据库级别
- 表级别
- 列级别

### 服务器级别

MySQL 提供了两个系统变量来表示服务器级别的字符集和比较规则：

- `character_set_server`服务器级别的字符集
- `collation_server` 服务器级别的比较规则

### 数据库级别

创建和修改数据库的时候可以指定该数据库的字符集和比较规则，具体语法如下：

```sql
create database 数据库名
    [[DEFAULT] character set 字符集名称]
    [[DEFAULT] collate 比较规则名称];

alter database 数据库名
    [[DEFAULT] character set 字符集名称]
    [[DEFAULT] collate 比较规则名称];
```

`DEFAULT` 可以省略。

```sql
mysql> CREATE DATABASE charset_demo_db
    -> CHARACTER SET gb2312
    -> COLLATE gb2312_chinese_ci;
Query OK, 1 row affected (0.01 sec)
```

上面语句表示创建一个名叫 `charset_demo_db` 的数据库，指定它的字符集为 `gb2312`，比较规则为 `gb2312_chinese_ci`。

查看当前数据库使用的字符集和比较规则（先使用 `use` 语句选择当前数据库）：

```sql
show variables like 'character_set_database';
```

- `character_set_database` 当前数据库的字符集
- `collation_database` 当前数据库的比较规则

> `character_set_database` 和 `collation_database` 这两个系统变量是**只读**的，不能修改。

如果数据库**没有指定字符集和比较规则**，那么会**使用服务器级别的字符集和比较规则**。

### 表级别

指定表的字符集和比较规则，语法如下：

```sql
create table 表名 (列的信息)
    [[DEFAULT] character set 字符集名称]
    [collate 比较规则名称]]

alter table 表名
    [[DEFAULT] character set 字符集名称]
    [collate 比较规则名称]
```

比如创建一个名为 `t` 的表，并指定这个表的字符集和比较规则：

```sql
mysql> create table t(
    ->     col varchar(10)
    -> ) character set utf8 collate utf8_general_ci;
Query OK, 0 rows affected (0.03 sec)
```

如果表**没有指定字符集和比较规则**，那么**使用该表所在数据库的字符集和比较规则**。

### 列级别

**同一个表中的不同的列也可以有不同的字符集和比较规则**。指定列的字符集和比较规则，语法如下：

```sql
cretae table 表名(
    列名 字符串类型 [character set 字符集名称] [collate 比较规则名称],
    其他列...
);

alter table 表名 modify 列名 字符串类型 [character set 字符集名称] [collate 比较规则名称];
```

比如我们修改一下表 `t` 中列 `col` 的字符集和比较规则可以这么写：

```sql
mysql> alter table t modify col varchar(10) character set gbk collate gbk_chinese_ci;
Query OK, 0 rows affected (0.04 sec)
Records: 0  Duplicates: 0  Warnings: 0
```

如果**没有指定该列的字符集和比较规则**，那么会**使用该列所在表的字符集和比较规则**。

> 在转换列的字符集时，如果转换前列中存储的数据不能用转换后的字符集进行表示，就会发生错误。如，之前列的字符集是 `utf8`，并存储了一些汉字，现在把列的字符集转换为 `ascii` 的话就会出错，因为 `ascii` 字符集并不能表示汉字字符。

### 仅修改字符集或仅修改比较规则

由于字符集和比较规则是互相有联系的，如果只修改了字符集，比较规则也会跟着变化，如果只修改了比较规则，字符集也会跟着变化，具体规则如下：

- 只修改字符集，则比较规则将变为修改后的字符集默认的比较规则。
- 只修改比较规则，则字符集将变为修改后的比较规则对应的字符集。

**不论哪个级别的字符集和比较规则，这两条规则都适用**。

### 客户端和服务器通信中的字符集

#### 字符集转换

字符 `'我'` 在 `utf8` 字符集编码下的字节串长这样：`0xE68891`。

如果接收 `0xE68891` 这个字节串的程序按照 `utf8` 字符集进行解码，然后又把它按照 `gbk` 字符集进行编码，最后编码后的字节串就是 `0xCED2`，把这
个过程称为**字符集的转换**，也就是字符串 `'我'` 从 `utf8` 字符集转换为 `gbk` 字符集。

#### MySQL 中字符集的转换

从客户端发往服务端的请求本质上就是一个字符串，服务端向客户端返回的结果本质上也是一个字符串，而字符串其实是使用某种字符集编码的二进制数据。这个字符
串可不是使用一种字符集的编码方式一条道走到黑的，从发送请求到返回结果这个过程中伴随着多次字符集的转换，在这个过程中会用到 3 个系统变量：

- `character_set_client` 服务端解码请求时使用的字符集
- `character_set_connection` 服务端处理请求时会把请求字符串从 `character_set_client` 转为 `character_set_connection`
- `character_set_results` 服务端向客户端返回数据时使用的字符集

这三个系统变量的值可能默认都是 `utf8`。为了体现出字符集在请求处理过程中的变化，这里特意修改一个系统变量的值：

```sh
mysql> set character_set_connection = gbk;
Query OK, 0 rows affected (0.00 sec)
```

所以现在系统变量 `character_set_client` 和 `character_set_results` 的值还是 `utf8`，而 `character_set_connection` 的值为 `gbk`。
现在假设客户端发送的请求是下边这个字符串：

```sql
SELECT * FROM t WHERE s = '我';
```

分析字符 `'我'` 在这个过程中字符集的转换。请求从发送到结果返回过程中字符集的变化：

1. 客户端发送请求所使用的字符集
一般情况下客户端所使用的字符集和当前操作系统一致，不同操作系统使用的字符集可能不一样，如下：

- 类 `Unix` 系统使用的是 `utf8`
- `Windows` 使用的是 `gbk`

例如在使用的 `linux` 操作系统时，客户端使用的就是 `utf8` 字符集。所以字符 `'我'` 在发送给服务端的请求中的字节形式就是：`0xE68891`

2. 服务端接收到客户端发送来的请求其实是一串二进制的字节，它会认为这串字节采用的字符集是 `character_set_client`，然后把这串字节转换
为 `character_set_connection` 字符集编码的字符。由于计算机上 `chacharacter_set_client` 的值是 `utf8`，首先会按照 `utf8` 字符集
对字节串 `0xE68891` 进行解码，得到的字符串就是 `'我'`，然后按照 `character_set_connection` 代表的字符集，也就是 `gbk` 进行编码，得
到的结果就是字节串 `0xCED2`。

3. 因为表 `t` 的列 `col` 采用的是 `gbk` 字符集，与 `character_set_connection` 一致，所以直接到列中找字节值为 `0xCED2` 的记录，最后找
到了一条记录。

**如果某个列使用的字符集和 `character_set_connection` 代表的字符集不一致的话，还需要进行一次字符集转换**。

4. 上一步骤找到的记录中的 `col` 列其实是一个字节串 `0xCED2`，`col` 列是采用 `gbk` 进行编码的，所以首先会将这个字节串使用 `gbk` 进行解码，
得到字符串`'我'`，然后再把这个字符串使用 `character_set_results` 代表的字符集，也就是 `utf8` 进行编码，得到了新的字节串：`0xE68891`，然
后发送给客户端。

5. 由于客户端是用的字符集是 `utf8`，所以可以顺利的将 `0xE68891` 解释成字符我，从而显示到我们的显示器上，所以我们人类也读懂了返回的结果。

几点需要注意的地方：

- 假设你的客户端采用的字符集和 `character_set_client` 不一样的话，这就会出现意想不到的情况。
- 假设你的客户端采用的字符集和 `character_set_results` 不一样的话，这就可能会出现客户端无法解码结果集的情况

> **通常都把 `character_set_client`、`character_set_connection`、`character_set_results` 这三个系统变量设置成和客户端使用的字符集一致的情况，这样减少了很多无谓的字符集转换**。

MySQL 提供了一条非常简便的语句:

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

> 如果你使用的是 Windows 系统，那应该设置成 `gbk`。

#### 比较规则的应用

**比较规则**的作用通常体现在**比较字符串大小的表达式**以及**对某个字符串列进行排序**中，所以有时候也称为**排序规则**。比方说表 `t` 的列 `col` 使用的字符集是 `gbk`，使用的比较规则是 `gbk_chinese_ci`，我们向里边插入几条记录：

```sh
mysql> INSERT INTO t(col) VALUES('a'), ('b'), ('A'), ('B');
Query OK, 4 rows affected (0.00 sec)
Records: 4  Duplicates: 0  Warnings: 0

mysql>
```

查询的时候按照 `col` 列排序一下：

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

可以看到在默认的比较规则 `gbk_chinese_ci` 中是不区分大小写的，我们现在把列 `col` 的比较规则修改为 `gbk_bin`：

```sh
mysql> ALTER TABLE t MODIFY col VARCHAR(10) COLLATE gbk_bin;
Query OK, 5 rows affected (0.02 sec)
Records: 5  Duplicates: 0  Warnings: 0
```

`gbk_bin` 是直接比较字符的编码，所以是区分大小写的，我们再看一下排序后的查询结果：

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
