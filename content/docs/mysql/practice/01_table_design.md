---
title: 表设计
weight: 1
---

在选择数据类型时，一般应该遵循下面两步：

1. 确定合适的大类型：数字、字符串、时间、二进制；
2. 确定具体的类型：有无符号、取值范围、变长定长等。

在 MySQL 数据类型设置方面，尽量用**更小的数据类型，它们通常有更好的性能，花费更少的硬件资源**。

## 数值

### signed 和 unsigned

在整型类型中，有 `signed` 和 `unsigned` 属性，默认为 `signed`。

表结构设计中时：

1. 如果整形数据确定没有负数，如主键 ID，建议指定为 `unsigned` 无符号类型，容量可以扩大一倍。
2. 其他情况不建议刻意去用 `unsigned` 属性，因为在做一些数据分析时，`unsigned` 会导致一些问题。

**MySQL 要求 `unsigned` 数值相减之后依然为 `unsigned`，否则就会报错**。在一些统计计算的 SQL 中，可能会出现 `unsigned` 数值相减之后为负数的情况，这时就会报错。

为了避免这个错误，需要对数据库参数 `sql_mode` 设置为 `NO_UNSIGNED_SUBTRACTION`，允许相减的结果为 `signed`：

```sql
SET sql_mode='NO_UNSIGNED_SUBTRACTION';
```

### 整型类型与自增设计

- 整型类型最常见的就是在业务中用来表示某件物品的数量。例如销售数量、库存数量、购买次数等。
- 另一个重要的使用用法是作为表的主键。

整型结合属性 `auto_increment`，可以实现**自增**功能，但在表结构设计时用自增做主键，要注意以下两点：

- 用 `BIGINT` 做主键，而不是 `INT`，不要为了节省 4 个字节使用 `INT`，当达到上限时，再进行表结构的变更，要付出巨大的代价。当达到 `INT` 上限后，再次进行自增插入时，会报重复错误，MySQL 数据库并不会自动将其重置为 `1`。
- MySQL 8.0 版本前自增值是**不持久化的**，可能会有回溯现象。

#### 回溯现象

```bash
mysql> SELECT * FROM t;
+---+
| a |
+---+
| 1 |
| 2 |
| 3 |
+---+
3 rows in set (0.01 sec)

mysql> DELETE FROM t WHERE a = 3;
Query OK, 1 row affected (0.02 sec)

mysql> SHOW CREATE TABLE t\G
*************************** 1. row ***************************
Table: t
Create Table: CREATE TABLE `t` (
  `a` int NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`a`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci
1 row in set (0.00 sec)
```

可以看到，在删除自增为 3 的这条记录后，下一个自增值依然为 4（`AUTO_INCREMENT=4`），这里并没有错误。但若这时数据库发生重启，那数据库启动后，表 t 的自增起始值将再次变为 3，即自增值发生回溯。

若要彻底解决这个问题，有 2 种方法：

1. 升级 MySQL 版本到 8.0 版本，每张表的自增值会持久化；
2. 为了之后更好的分布式架构扩展性，不建议使用自增整型类型做主键，更为推荐的是字符串类型，例如 UUID。

### 资金字段设计

通常在表结构设计中，用户的工资、账户的余额等精确到小数点后 2 位的业务（精确到**分**），可以使用 `DECIMAL` 类型（例如 `DECIMAL(8,2)`）。

**在海量互联网业务的设计标准中，并不推荐用 `DECIMAL` 类型，更推荐将 DECIMAL 转化为整型类型**。也就是说，资金类型更推荐使用用分单位存储，而不是用元单位存储。如 1 元在数据库中用整型类型 100 存储。

#### 金额字段为什么不使用 DECIMAL 类型？

金额字段的取值范围如果用 DECIMAL 表示的，如何定义长度？因为类型 **DECIMAL 是个变长字段**，如果定义为 `DECIMAL(8,2)` ，那么存储最大值为 `999999.99`，百万级的资金存储，这是远远不够的。

用户的金额至少要存储百亿的字段，而统计局的 GDP 金额字段则可能达到数十万亿级别。用类型 `DECIMAL` 定义，不好统一。而如果使用 `BIGINT` 来存储金额，用**分**来做单位（小数点中的数据呢，可以交由前端进行处理并展示），可以轻松存储千兆级别的金额。

- 存储高效：所有金额相关字段都是定长字段，占用 8 个字节。
- 扩展性强：轻松支持百亿级甚至万亿级金额存储。
- 计算高效：整型运算比 DECIMAL 的二进制转换计算更快。

**在数据库设计中，非常强调定长存储，因为定长存储的性能更好**。

### IP 地址字段设计

IP 也是可以使用整型来存储的，因为 IP 地址本身是一个**变长字段**，如果使用 `INT` 存储，就是占用固定的 4 个字节。这种方法比使用字符串更节省空间且查询效率更高。

存储 IP 地址的表结构：

```sql
CREATE TABLE `network_logs` (
    id INT AUTO_INCREMENT PRIMARY KEY,
    ip_address INT UNSIGNED,  -- 存储转换后的IP整数值
    -- 其他字段...
    INDEX (ip_address)        -- 为提高查询效率添加索引
);
```

**一定要使用 `INT UNSIGNED` 类型，使用 `INT UNSIGNED` 可以存储 0 到 4294967295 的值，正好对应 IPv4 的 32 位地址空间**。所以这种方法只适用于 IPv4 地址，IPv6 地址需要其他存储方式。

```sql
-- 将IP字符串转为整型
SELECT INET_ATON('192.168.1.1');  -- 返回: 3232235777

-- 将整型转为 IP 字符串
SELECT INET_NTOA(3232235777);     -- 返回: '192.168.1.1'

-- 插入记录
INSERT INTO network_logs (ip_address) VALUES (INET_ATON('192.168.1.1'));

-- 查询记录
-- 查询特定IP
SELECT * FROM network_logs WHERE ip_address = INET_ATON('192.168.1.1');

-- 查询时显示IP字符串格式
SELECT id, INET_NTOA(ip_address) AS ip FROM network_logs;
```

也可以在**应用程序中先转换再存储，减少数据库的计算负担**。


#### 性能优势

- 存储空间：整型(4 字节) vs 字符串(`7-15`字节)
- 查询效率：整型比较比字符串比较快得多
- 索引效率：整型索引更小更高效

### 小数类型

浮点类型 `Float` 和 `Double`，不是高精度，也不是 SQL 标准的类型，**不推荐使用**。

小数类型使用 `DECIMAL` 类型。如果存储的数值范围超过 `DECIMAL` 的范围，可以**将数值拆成整数和小数并分开存储**。

## 字符串

大部分场景中，对于字符串使用类型 `VARCHAR` 就足够了。

- 字符串的长度相差较大用 `VARCHAR`；字符串短，且所有值都接近一个长度用 `CHAR`。
- 尽量少用 `BLOB` 和 `TEXT`，如果实在要用可以考虑将 `BLOB` 和 `TEXT` 字段单独存一张表，用 `id` 关联。

### utf8

**`utf8` 字符集**：是 Unicode 字符集的一种编码方案，收录地球上能想到的所有字符，而且还在不断扩充。这种字符集兼容 ASCII 字符集，采用变长编码方式，编码一个字符需要使用 `1~4` 个字节。

`utf8` 字符集表示一个字符需要使用 `1~4` 个字节，但是我们常用的一些字符使用 `1~3` 个字节就可以表示了。而在一个字符所用最大字节长度在某些方面会影响系统的存储和性能，所以MySQL 定义了两个概念：

- `utf8mb3`：阉割过的 `utf8` 字符集，只使用 `1~3` 个字节表示字符。无法存储需要 4 字节编码的字符，如表情符号、其他补充字符等。
- `utf8mb4`：正宗的 `utf8` 字符集，使用 `1~4` 个字节表示字符。

MySQL **8.0 和以上版本**：

- 字符集默认为 `utf8mb4`。
- `utf8` 默认指向的也是 `utf8mb4`。

**8.0 之前的版本**：

- 字符集默认为 `latin1`。
- `utf8` 默认指向的也是 `utf8mb3`。

**如果主要涉及英文和少量特殊符号，并且不打算使用表情符号或任何特殊 Unicode 字符，那么使用 `utf8mb3` 就足够了**。

**如果需要支持多种语言（国际化），包括那些使用大量特殊字符（如表情符号）的语言，那么使用 `utf8mb4`**。

### 字符集排序规则

MySQL 支持**多种字符集，每个字符集都会有默认的排序规则**。可以用命令 `SHOW CHARSET` （`CHARACTER SET` 和 `CHARSET` 是同义词）来查看：

```bash
mysql> SHOW CHARSET LIKE 'utf8%';

+---------+---------------+--------------------+--------+
| Charset | Description   | Default collation  | Maxlen |
+---------+---------------+--------------------+--------+
| utf8    | UTF-8 Unicode | utf8_general_ci    |      3 |
| utf8mb4 | UTF-8 Unicode | utf8mb4_0900_ai_ci |      4 |
+---------+---------------+--------------------+--------+
2 rows in set (0.01 sec)

mysql> SHOW COLLATION LIKE 'utf8mb4%';
+----------------------------+---------+-----+---------+----------+---------+---------------+
| Collation                  | Charset | Id  | Default | Compiled | Sortlen | Pad_attribute |
+----------------------------+---------+-----+---------+----------+---------+---------------+
| utf8mb4_0900_ai_ci         | utf8mb4 | 255 | Yes     | Yes      |       0 | NO PAD        |
| utf8mb4_0900_as_ci         | utf8mb4 | 305 |         | Yes      |       0 | NO PAD        |
| utf8mb4_0900_as_cs         | utf8mb4 | 278 |         | Yes      |       0 | NO PAD        |
| utf8mb4_0900_bin           | utf8mb4 | 309 |         | Yes      |       1 | NO PAD        |
| utf8mb4_bin                | utf8mb4 |  46 |         | Yes      |       1 | PAD SPACE     |
......
```

其中 **Default collation** 列表示这种字符集中一种默认的比较规则。排序规则

- 以 `_ci` 结尾，表示不区分大小写（Case Insentive）
- `_cs` 表示大小写敏感
- `_bin` 表示通过存储字符的二进制进行比较。

**绝大部分业务的表结构设计无须设置排序规则为大小写敏感**。

#### 正确修改字符集

```sql
ALTER TABLE emoji_test CHARSET utf8mb4;
```

上面的 SQL 将表的字符集修改为 `utf8mb4`，下次**新增列时，若不显式地指定字符集，新列的字符集会变更为 `utf8mb4`**，但对于**已经存在的列，其默认字符集并不做修改**。

```bash
mysql> SHOW CREATE TABLE emoji_test\G

*************************** 1. row ***************************
Table: emoji_test
Create Table: CREATE TABLE `emoji_test` (
  `a` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  PRIMARY KEY (`a`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
1 row in set (0.00 sec)
```

表的列 `a` 的字符集依然是 `utf8`。

所以，正确修改列字符集的命令应该使用 `ALTER TABLE ... CONVERT TO`：

```bash
mysql> ALTER TABLE emoji_test CONVERT TO CHARSET utf8mb4;
Query OK, 0 rows affected (0.94 sec)
Records: 0  Duplicates: 0  Warnings: 0

mysql> SHOW CREATE TABLE emoji_test\G
*************************** 1. row ***************************
Table: emoji_test
Create Table: CREATE TABLE `emoji_test` (
  `a` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  PRIMARY KEY (`a`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
1 row in set (0.00 sec)
```

### 性别或状态等枚举类型设计

设计表结构时，你会遇到一些固定选项值的字段。例如，性别字段（Sex），只有男或女；又或者状态字段（State），有效的值为运行、停止、重启等有限状态。

**不推荐使用 `ENUM` 类型**，因为 `ENUM` 类型并非 SQL 标准的数据类型，而是 MySQL 所独有的一种字符串类型。抛出的错误提示也并不直观。

通常情况下，**枚举**类型，可以使用 `unsigned tinyint` 类型替代，占用 1 个字节，取值范围是 `0~255`。

例如，性别字段，可以使用 `unsigned tinyint` 类型替代，其中 `0` 表示男，`1` 表示女。

####  CHECK 约束功能

MySQL 8.0.16 版本开始，数据库原生提供了 CHECK 约束功能，用来检查字段值是否符合指定的条件。**通过 CHECK 约束，可以实现枚举类型的功能**。

```bash
mysql> SHOW CREATE TABLE User\G
*************************** 1. row ***************************
Table: User
Create Table: CREATE TABLE `User` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `sex` char(1) COLLATE utf8mb4_general_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `user_chk_1` CHECK (((`sex` = _utf8mb4'M') or (`sex` = _utf8mb4'F')))
) ENGINE=InnoDB
1 row in set (0.00 sec)

mysql> INSERT INTO User VALUES (NULL,'M');
Query OK, 1 row affected (0.07 sec)

mysql> INSERT INTO User VALUES (NULL,'Z');
ERROR 3819 (HY000): Check constraint 'user_chk_1' is violated.
```

示例中的约束定义 `user_chk_1` 表示列 sex 的取值范围，只能是 `M` 或者 `F`。插入非法数据 Z 时，看到 MySQL 显式地抛出了违法约束的提示。

通过 CHECK 约束，实现枚举类型可以避免 `tinyint` 实现的两个问题：

1. **表达不清**：在具体存储时，0 表示女，还是 1 表示女呢？每个业务可能有不同的潜规则；
2. **脏数据**：因为是 `tinyint`，因此除了 0 和 1，用户完全可以插入 2、3、4 这样的数值，最终表中可能存在无效数据，如果后期再进行清理，代价就非常大了。

### 账户密码存储设计

密码不能明文存储，通常的做法是对密码进行加密存储。

不要直接使用 MD5 算法，虽然 MD5 算法并不可逆，但是可以通过暴力破解的方式，计算出所有可能的字符串对应的 MD5 值。

所以，在设计密码存储使用，还需要加**盐**（salt），每个公司的盐值都是不同的，因此计算出的值也是不同的。若盐值为 `psalt`，则密码 `12345678` 在数据库中的值为：

```
password = MD5（‘psalt12345678’）
```

这是一种固定盐值的加密算法，其中存在三个主要问题：

1. 若 salt 值被（离职）员工**泄漏**，则外部黑客依然存在暴利破解的可能性；
2. 对于**相同密码，其密码存储值相同**，一旦一个用户密码泄漏，其他相同密码的用户的密码也将被泄漏；
3. 固定使用 MD5 加密算法，一旦 MD5 算法被破解，则影响很大。

所以一个真正好的密码存储设计，应该是：**动态盐 + 非固定加密算法**。例如密码的存储格式为：

```
$salt$cryption_algorithm$value
```

- `$salt`：表示动态盐，**每次用户注册时业务产生不同的盐值**，并存储在数据库中。若做得再精细一点，可以**动态盐值 + 用户注册日期**合并为一个更为动态的盐值。
- `$cryption_algorithm`：表示加密的算法，如 v1 表示 MD5 加密算法，v2 表示 AES256 加密算法，v3 表示 AES512 加密算法等。
- `$value`：表示加密后的字符串。

```bash
CREATE TABLE User (
    id BIGINT NOT NULL AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    sex CHAR(1) NOT NULL,
    password VARCHAR(1024) NOT NULL,
    CHECK (sex = 'M' OR sex = 'F'),
    PRIMARY KEY(id)
);


SELECT * FROM User\G
*************************** 1. row ***************************
id: 1
name: David
sex: M
password: $fgfaef$v1$2198687f6db06c9d1b31a030ba1ef074
*************************** 2. row ***************************
id: 2
name: Amy
sex: F
password: $zpelf$v2$0x860E4E3B2AA4005D8EE9B7653409C4B133AF77AEF53B815D31426EC6EF78D882
```

上面的例子中，用户 David 和 Amy 密码都是 12345678，然而由于使用了动态盐和动态加密算法，两者存储的内容完全不同。

## 日期和时间

在做表结构设计时，对日期字段的存储，开发人员通常会有 3 种选择：`DATETIME`、`TIMESTAMP`、`INT`。

**建议使用类型 `DATETIME`**。

**为什么不使用 `INT` 类型？**

`INT` 类型就是直接存储 '1970-01-01 00:00:00' 到现在的毫秒数，本质和 `TIMESTAMP` 一样。而且 `INT` 存储日期**可读性，运维性太差**了。

有些同学会认为 `INT` 比 `TIMESTAMP` 性能更好。但是，当前每个 CPU 每秒可执行上亿次的计算，所以无须为这种转换的性能担心。

**为什么不使用 `TIMESTAMP` 类型？**

虽然 `TIMESTAMP` 占用 4 个字节，比 `DATETIME` 小一半的存储空间。但是 `TIMESTAMP` 的最大值 '2038-01-19 03:14:07' 已经很接近了，要慎重考虑。

`TIMESTAMP` 虽然有时区属性的优势，对于 `DATETIME` 的时区问题，可以由前端或者服务端做一次转化，不一定非要在数据库中解决。

**`TIMESTAMP` 的性能问题**，

虽然从毫秒数转换到类型 `TIMESTAMP` 本身需要的 CPU 指令并不多，这并不会带来直接的性能问题。但是如果使用默认的操作系统时区，则每次通过时区计算时间时，要调用操作系统底层系统函数 `__tz_convert()`，而这个函数需要额外的加锁操作，以确保这时操作系统时区没有修改。所以，当大规模并发访问时，由于热点资源竞争，会产生两个问题。

- **性能不如 `DATETIME`**： `DATETIME` 不存在时区转化问题。
- **性能抖动**： 海量并发时，存在性能抖动问题。

可以通过**设置显式的时区**，来优化 `TIMESTAMP`。比如在配置文件中显示地设置时区，而不要使用系统时区：

```ini
[mysqld]
time_zone = "+08:00"
```

**日期字段推荐使用 `DATETIME`，没有时区转化。即便使用 `TIMESTAMP`，也需要在数据库中显式地配置时区，而不是用系统时区**。


## JSON 数据类型

### 用户登录设计

**JSON 类型比较适合存储一些修改较少、相对静态的数据**，比如用户登录信息的存储：

```sql
DROP TABLE IF EXISTS UserLogin;

CREATE TABLE UserLogin (
    userId BIGINT NOT NULL,
    loginInfo JSON,
    PRIMARY KEY(userId)
);
```

由于现在业务的登录方式越来越多样化，如同一账户支持手机、微信、QQ 账号登录，所以这里可以用 JSON 类型存储登录的信息。

插入下面的数据：

```sql
SET @a = '
{
  "cellphone" : "13918888888",
  "wxchat" : "破产码农",
  "QQ" : "82946772"
}
';

INSERT INTO UserLogin VALUES (1,@a);

SET @b = '
{
  "cellphone" : "15026888888"
}
';

INSERT INTO UserLogin VALUES (2,@b);
```

上面的例子中，用户 1 登录有三种方式：手机验证码登录、微信登录、QQ 登录，而用户 2 只有手机验证码登录。

而如果不采用 JSON 数据类型，就要用下面的方式建表：

```sql
CREATE TABLE UserLogin (
    userId		BIGINT NOT NULL,
    cellphone	VARCHAR(255),
    wechat		VARCHAR(255)
    QQ			VARCHAR(255),
    PRIMARY KEY(userId)
);
```

上面传统的方式存在两个问题：

1. 有些列可能是比较稀疏的，一些列可能大部分都是 NULL 值；
2. 如果要新增一种登录类型，如微博登录，则需要添加新列，而 JSON 类型无此烦恼。

使用 JSON 类型，再配合 JSON 字段处理函数，实现更加简便。其中，最常见的 JSON 字段处理函数就是 **`JSON_EXTRACT`，它用来从 JSON 数据中提取所需要的字段内容**。

例如查询用户的手机和微信信息：

```sql
SELECT
    userId,
    JSON_UNQUOTE(JSON_EXTRACT(loginInfo,"$.cellphone")) cellphone,
    JSON_UNQUOTE(JSON_EXTRACT(loginInfo,"$.wxchat")) wxchat
FROM UserLogin;
+--------+-------------+--------------+
| userId | cellphone   | wxchat       |
+--------+-------------+--------------+
|      1 | 13918888888 | 破产码农     |
|      2 | 15026888888 | NULL         |
+--------+-------------+--------------+
2 rows in set (0.01 sec)
```

MySQL 还提供了 `->>` 表达式，和 `JSON_EXTRACT`、`JSON_UNQUOTE` 的效果完全一样：

```sql
SELECT
    userId,
    loginInfo->>"$.cellphone" cellphone,
    loginInfo->>"$.wxchat" wxchat
FROM UserLogin;
```

当 JSON 数据量非常大，用户希望对 JSON 数据进行有效检索时，可以利用 MySQL 的**函数索引**功能对 JSON 中的某个字段进行索引。

比如在上面的用户登录示例中，假设用户必须绑定唯一手机号，且希望未来能用手机号码进行用户检索时，可以创建下面的索引：

```sql
ALTER TABLE UserLogin ADD COLUMN cellphone VARCHAR(255) AS (loginInfo->>"$.cellphone");

ALTER TABLE UserLogin ADD UNIQUE INDEX idx_cellphone(cellphone);
```

1. 首先创建了一个虚拟列 `cellphone`，这个列是由函数 `loginInfo->>"$.cellphone"` 计算得到的。
2. 在这个虚拟列上创建一个唯一索引 `idx_cellphone`。

通过虚拟列 `cellphone` 进行查询，就可以看到优化器会使用到新创建的 `idx_cellphone` 索引：

```sql
EXPLAIN SELECT  *  FROM UserLogin 
WHERE cellphone = '13918888888'\G
*************************** 1. row ***************************
           id: 1
  select_type: SIMPLE
        table: UserLogin
   partitions: NULL
         type: const
possible_keys: idx_cellphone
          key: idx_cellphone
      key_len: 1023
          ref: const
         rows: 1
     filtered: 100.00
        Extra: NULL
1 row in set, 1 warning (0.00 sec)
```

### 用户画像设计

用户画像（也就是对用户打标签）比如：

- 在电商行业中，根据用户的穿搭喜好，推荐相应的商品；
- 在音乐行业中，根据用户喜欢的音乐风格和常听的歌手，推荐相应的歌曲；
- 在金融行业，根据用户的风险喜好和投资经验，推荐相应的理财产品。

就可以使用 JSON 类型在数据库中存储用户画像信息，并结合 JSON 数组类型和多值索引的特点进行高效查询。假设有张画像定义表：

```sql
CREATE TABLE Tags (
    tagId bigint auto_increment,
    tagName varchar(255) NOT NULL,
    primary key(tagId)
);

SELECT * FROM Tags;
+-------+--------------+
| tagId | tagName      |
+-------+--------------+
|     1 | 70后         |
|     2 | 80后         |
|     3 | 90后         |
|     4 | 00后         |
|     5 | 爱运动       |
|     6 | 高学历       |
|     7 | 小资         |
|     8 | 有房         |
|     9 | 有车         |
|    10 | 常看电影     |
|    11 | 爱网购       |
|    12 | 爱外卖       |
+-------+--------------+
```

表 Tags 是一张画像定义表，用于描述当前定义有多少个标签，接着给每个用户打标签，比如用户 David，他的标签是 80 后、高学历、小资、有房、常看电影；用户 Tom，90 后、常看电影、爱外卖。

若不用 JSON 数据类型进行标签存储，通常会将用户标签通过字符串，加上分割符的方式，在一个字段中存取用户所有的标签：

```bash
+-------+---------------------------------------+
|用户    |标签                                   |
+-------+---------------------------------------+
|David  |80后 ； 高学历 ； 小资 ； 有房 ；常看电影   |
|Tom    |90后 ；常看电影 ； 爱外卖                 |
+-------+---------------------------------------+
```

**缺点**：不好搜索特定画像的用户，另外分隔符也是一种自我约定，在数据库中其实可以任意存储其他数据，最终产生脏数据。

JSON 数据类型就能很好解决这个问题：

```sql
DROP TABLE IF EXISTS UserTag;
CREATE TABLE UserTag (
    userId bigint NOT NULL,
    userTags JSON,
    PRIMARY KEY (userId)
);

INSERT INTO UserTag VALUES (1,'[2,6,8,10]');
INSERT INTO UserTag VALUES (2,'[3,10,12]');
```

MySQL 8.0.17 版本开始支持 Multi-Valued Indexes，用于在 JSON 数组上创建索引，并通过函数 member of、json_contains、json_overlaps 来快速检索索引数据。

```sql
ALTER TABLE UserTag
ADD INDEX idx_user_tags ((cast((userTags->"$") as unsigned array)));
```

如果想要查询用户画像为常看电影的用户，可以使用函数 MEMBER OF：

```sql
EXPLAIN SELECT * FROM UserTag 
WHERE 10 MEMBER OF(userTags->"$")\G
*************************** 1. row ***************************
           id: 1
  select_type: SIMPLE
        table: UserTag
   partitions: NULL
         type: ref
possible_keys: idx_user_tags
          key: idx_user_tags
      key_len: 9
          ref: const
         rows: 1
     filtered: 100.00
        Extra: Using where
1 row in set, 1 warning (0.00 sec)

SELECT * FROM UserTag 
WHERE 10 MEMBER OF(userTags->"$");
+--------+---------------+
| userId | userTags      |
+--------+---------------+
|      1 | [2, 6, 8, 10] |
|      2 | [3, 10, 12]   |
+--------+---------------+
2 rows in set (0.00 sec)
```

如果想要查询画像为 80 后，且常看电影的用户，可以使用函数 JSON_CONTAINS：

```sql
EXPLAIN SELECT * FROM UserTag 
WHERE JSON_CONTAINS(userTags->"$", '[2,10]')\G
*************************** 1. row ***************************
           id: 1
  select_type: SIMPLE
        table: UserTag
   partitions: NULL
         type: range
possible_keys: idx_user_tags
          key: idx_user_tags
      key_len: 9
          ref: NULL
         rows: 3
     filtered: 100.00
        Extra: Using where
1 row in set, 1 warning (0.00 sec)

SELECT * FROM UserTag 
WHERE JSON_CONTAINS(userTags->"$", '[2,10]');
+--------+---------------+
| userId | userTags      |
+--------+---------------+
|      1 | [2, 6, 8, 10] |
+--------+---------------+
1 row in set (0.00 sec)
```

如果想要查询画像为 80 后或 90 后，且常看电影的用户，则可以使用函数 JSON_OVERLAP：

```sql
EXPLAIN SELECT * FROM UserTag 
WHERE JSON_OVERLAPS(userTags->"$", '[2,3,10]')\G
*************************** 1. row ***************************
           id: 1
  select_type: SIMPLE
        table: UserTag
   partitions: NULL
         type: range
possible_keys: idx_user_tags
          key: idx_user_tags
      key_len: 9
          ref: NULL
         rows: 4
     filtered: 100.00
        Extra: Using where
1 row in set, 1 warning (0.00 sec)

SELECT * FROM UserTag 
WHERE JSON_OVERLAPS(userTags->"$", '[2,3,10]');
+--------+---------------+
| userId | userTags      |
+--------+---------------+
|      1 | [2, 6, 8, 10] |
|      2 | [3, 10, 12]   |
+--------+---------------+
2 rows in set (0.01 sec)
```







```sql
SET @a = '{
  "cellphone": "188888888888",
  "wxchat": "this.",
  "QQ": "166666666"
}';

INSERT INTO UserLogin VALUES (1, @a);
```

```sql
SELECT 
    user_id,
    JSON_UNQUOTE(JSON_EXTRACT(loginInfo,"$.cellphone")) cellphone,
    JSON_UNQUOTE(JSON_EXTRACT(loginInfo,"$.wxchat")) wxchat
FROM UserLogin;
```

使用 `->>` 表达式，效果一样：

```sql
SELECT
    userId,
    loginInfo->>"$.cellphone" cellphone,
    loginInfo->>"$.wxchat" wxchat
FROM UserLogin;
```

## 大类型

尽量少用 `BLOB` 和 `TEXT` 等大类型，如果实在要用可以考虑将 `BLOB` 和 `TEXT` 字段单独存一张表，用 `id` 关联。