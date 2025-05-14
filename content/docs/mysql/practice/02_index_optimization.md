---
title: 索引使用和优化
weight: 2
---

## Explain 分析

`Explain` 关键字可以模拟优化器执行 SQL 语句，分析查询语句或是结构的性能瓶颈。

在 `select` 语句之前增加 `explain` 关键字，MySQL 会在查询上设置一个标记，执行查询会返回执行计划的信息，而不是执行这条 SQL。

注意：如果 `from` 中包含子查询，仍会执行该子查询，将结果放入临时表中。

创建示例表：

```sql
DROP TABLE IF EXISTS `actor`; 
CREATE TABLE `actor` (
  `id` int(11) NOT NULL,
  `name` varchar(45) DEFAULT NULL,
  `update_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT INTO `actor` (`id`, `name`, `update_time`) VALUES (1,'a','2017-12-22 15:27:18'), (2,'b','2017-12-22 15:27:18'), (3,'c','2017-12-22 15:27:18');

DROP TABLE IF EXISTS `film`;
CREATE TABLE `film` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(10) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT INTO `film` (`id`, `name`) VALUES (3,'film0'),(1,'film1'),(2,'film2');

DROP TABLE IF EXISTS `film_actor`;
CREATE TABLE `film_actor` (
  `id` int(11) NOT NULL,
  `film_id` int(11) NOT NULL,
  `actor_id` int(11) NOT NULL,
  `remark` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_film_actor_id` (`film_id`,`actor_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT INTO `film_actor` (`id`, `film_id`, `actor_id`) VALUES (1,1,1),(2,1,2),(3,2,1);
```

### Explain 列

#### id

`select` 查询的序列号，有几个 `select` 就有几个 `id`，并且 `id` 的顺序是按 `select` 出现的顺序增长的。`id` 列越大执行优先级越高，`id` 相同则从上往下执行，`id` 为 `NULL` 最后执行。

#### select_type

表示对应行是简单还是复杂的查询。

1. `simple`：简单的 select 查询，查询中不包含子查询或者 `UNION`。

```sql
explain select * from film where id = 2;
```

![explain-simple-select]()


2. `primary`：查询中若包含任何复杂的子部分，最外层查询则被标记为 `primary`。
3. `subquery`：包含在 `select` 中的子查询（**不在 `from` 子句中**）
4. `derived`：包含在 `from` 子句中的子查询。MySQL会将结果存放在一个临时表中，也称为派生表。

```sql
-- 关闭衍生表的合并优化
set session optimizer_switch='derived_merge=off';
explain select (select 1 from actor where id = 1) from (select * from film where id = 1) der;

-- 还原默认配置
set session optimizer_switch='derived_merge=on';
```

![explain-complex-select]()

可以看出来：

-  `(select 1 from actor where id = 1)` 就是 `id` 为 2， 类型为 `subquery` 的 `select`。
- `(select * from film where id = 1)` 就是 `id` 为 3，类型为 `derived` 的 `select`。将查询语句的结果集放到一个临时表中。
- 最外层的 `select` 就是 `id` 为 1，类型为 `primary`。这个语句的 `table` 的值为 `<dirived3>`，表示这个查询是从衍生表中查询的。`<dirived3>` 中的 `3` 表示 `id` 为 3 的 `select`。

5. `union`：在 `union` 中的第二个和随后的 `select`

#### type

这一列表示**关联类型**或**访问类型**，即 MySQL 决定如何查找表中的行，查找数据行记录的大概范围。依次从最优到最差分别为：`system` > `const` > `eq_ref` > `ref` > `range` > `index` > `ALL`。一般来说，要**保证查询达到 range 级别，最好达到 ref**。

1. `NULL`：MySQL 优化器优化查询语句时，判断不需要再扫描表或着索引树。例如：在索引列中选取最小值，`select min(id) from film;`，直接从索引树中获取最小值，不需要扫描表。
2. `system`，`const`：主键索引或唯一二级索引的**等值**查询。MySQL 能对查询的某部分进行优化并将其转化成一个常量，就是说像查询常量一样快。因为表最多有一个匹配行，读取 1 次，所以速度很快。`system` 是 `const` 类型的特列，表示要查询的表或者结果集中只有一条数据。

```sql
EXPLAIN SELECT * FROM (SELECT * FROM film WHERE id = 1) tmp;
SHOW WARNINGS;
```

![expalin-type-const]()

`SHOW WARNINGS` 的 message 可以看到表示查询被优化成了 `select 1 AS 'id','film1' AS 'name' from dual`。


3. `eq_ref`：**主键索引**或**唯一二级索引的所有列**都被被连接使用，最多只会返回一条符合条件的记录。这可能是在 `const` 之外最好的联接类型了，简单的 `select` 查询不会出现这种类型。

```sql
EXPLAIN SELECT * FROM film_actor LEFT JOIN film ON film_actor.film_id = film.id;
```

![expalin-type-eq_ref]()

可以看到，两个 `select` 的 `id` 都是 1，说明关联的两张表没有明确的前后顺序，会一起去查询，不过真正执行的时候，会先执行上面的 `select`。

- `film` 表对应的 type 是 `eq_ref`，因为关联的条件使用的是 `film` 表的主键 `id`，所以会使用主键索引来查询数据。使用主键来查询，只会返回一条记录，速度还是很快的。

4. `ref`：相比 `eq_ref`，不使用唯一索引，而是使用二级索引或者唯一性索引的部分前缀，索引要和某个值相比较，可能会找到多个符合条件的行。

```sql
-- name 是普通索引
EXPLAIN SELECT * FROM film WHERE name = 'film1';
```

5. `range`：利用索引进行范围匹配。范围扫描通常出现在 `in(), between ,> ,<, >=` 等操作中。

```sql
EXPLAIN SELECT * FROM actor WHERE id > 1;
```

6. `index`：**全索引扫描**，一般是扫描某个二级索引，直接对二级索引的叶子节点遍历和扫描，速度还是比较慢的，这种查询一般为使用**覆盖索引**，二级索引一般比较小，所以这种通常比 `ALL` 快一些。

```sql
EXPLAIN SELECT * FROM film;
```

![expalin-type-index]()

这里扫描的是二级索引 `idx_name`，没有去扫描聚簇索引。MySQL 在优化时，**如果查询的字段在二级索引中全部都有，会优先使用二级索引**，而不会去扫描聚簇索引。因为一般情况下，聚簇索引的叶子节点存储的是完整的行数据，所要扫描的数据量会比较大，而二级索引的叶子节点存储的是主键值，所以扫描的行数会比较少。

7. `ALL`：Full Table Scan，即**全表扫描**，直接扫描聚簇索引的所有叶子节点。通常情况下这需要增加索引来进行优化了。

```sql
EXPLAIN SELECT * FROM actor;
```

#### possible_keys

查询时，可能使用的索引。如果 `possible_keys` 有列，而 `key` 为 `NULL`，这种情况是因为表中数据不多，MySQL 认为索引对此查询帮助不大，选择了全表查询。

如果 `possible_keys` 为 `NULL`，则没有相关的索引。在这种情况下，可以考虑创造一个适当的索引来提高查询性能。

#### key

实际使用的索引。如果为 `NULL`，则没有使用索引。如果想强制 MySQL 使用或忽视 `possible_keys` 列中的索引，可以在查询中使用 `force index`、`ignore index`。

#### key_len

显示 MySQL 在索引里使用的字节数，通过这个值可以算出具体使用了索引中的哪些列。例如，`film_actor` 的联合索引 `idx_film_actor_id` 由 `film_id` 和 `actor_id` 两个 `int` 列组成，并且每个 `int` 是 `4 `字节。通过结果中的 `key_len=4` 可推断出查询使用了第一个列：`film_id` 列来执行索引查找。

`key_len` 计算规则如下：

- 字符串，`char(n)` 和 `varchar(n)`，`n` 均代表字符数，而不是字节数，如果是 `utf-8`，一个数字或字母占 `1` 个字节，一个汉字占 `3` 个字节
`char(n)`：如果存汉字长度就是 `3n` 字节
`varchar(n)`：如果存汉字则长度是 `3n + 2` 字节，加的 `2` 字节用来存储字符串长度，因为 `varchar` 是变长字符串
- 数值类型
    - tinyint：1 字节
    - smallint：2 字节
    - int：4 字节
    - bigint：8 字节　　
- 时间类型　
    - date：3 字节
    - timestamp：4 字节
    - datetime：8 字节
- 如果字段允许为 NULL，需要 1 字节记录是否为 NULL
- 索引最大长度是 768 字节，当字符串过长时，MySQL 会做一个类似左前缀索引的处理，将前半部分的字符提取出来做索引

#### ref

显示索引的哪些列或常量被使用了。

#### rows

表示 MySQL 估计要读取并检测的行数，注意这个不是结果集里的行数。

#### filtered 

该列是一个百分比的值，`rows*filtered/100` 可以估算出将要和 explain 中前一个表进行连接的行数。

#### extra

展示的是额外信息：

- Using index：使用**覆盖索引**，避免访问了表的数据行，效率不错。
- Using where：表示使用了 `where` 过滤，并且**查询的列未被索引覆盖**。
- Usering index condition：表示使用了**索引下推**优化。
- Using temporary：要创建一张临时表来处理查询。出现这种情况一般是要进行优化的，首先是想到**用索引来优化**。例如 `EXPLAIN SELECT DISTINCT name FROM actor;` `actor.name` 没有索引，此时创建了张临时表来 `distinct`。可以为 `name` 列创建索引，然后再去重，MySQL 在扫描索引树的过程中就可以直接去重。因为索引是有序的，相同的记录是在一起的，相同的记录直接扔掉就可以了。
- Using filesort：将用外部排序而不是索引排序，数据较小时在内存排序，否则需要在磁盘完成排序。这种情况下一般也是要考虑使用索引来优化。索引本身就是排好序的。
- Using join buffer：使用连接缓存。
- Select tables optimized away：使用某些聚合函数（比如 `max`、`min`）来访问存在索引的某个字段。


#### table

表示 `explain` 的一行正在访问哪个表。

#### partitions

如果查询是基于分区表的话，partitions 字段会显示查询将访问的分区。


## 索引的使用

## 查询优化

```sql
select * from employees limit 10000,10;
```

从表 `employees` 中取出从 10001 行开始的 10 行记录。看似只查询了 10 条记录，实际这条 SQL 是先读取 10010 条记录，然后抛弃前 10000 条记录，然后返回后面 10 条想要的数据。因此要**查询一张大表比较靠后的数据，执行效率是非常低的**。


### 常见的分页场景优化技巧

#### 自增且连续的主键排序的分页查询

```sql
select * from employees limit 90000,5;

-- 优化为
select * from employees where id > 90000 limit 5;
```

但是，这条改写的 SQL 在很多场景并不实用，因为表中可能某些记录被删后，主键空缺，导致结果不一致。

#### 根据非主键字段排序的分页查询

```sql
-- 联合索引 (name,age,position)
-- 该 sql 没有使用 name 字段的索引，因为查找联合索引的结果集太大，并回表的成本比扫描全表的成本更高，所以优化器放弃使用索引。
select * from employees ORDER BY name limit 90000,5;

-- 关键是让排序时返回的字段尽可能少，所以可以让排序和分页操作先查出主键，然后根据主键查到对应的记录，SQL 改写如下
-- 优化为
select * from employees e inner join (select id from employees order by name limit 90000,5) ed on e.id = ed.id;
```

优化后的语句全部都走了索引，其中 `(select id from employees order by name limit 90000,5)` 使用了覆盖索引来优化，查询的字段只有 id 字段，而且排好了序。`(select id from employees order by name limit 90000,5) ed` 产生的临时表只有 5 条记录，然后再根据主键 id 去 `employees` 表中查询对应的记录。

### JOIN 关联查询优化

测试数据：

```sql
-- 示例表：
CREATE TABLE `t1` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `a` int(11) DEFAULT NULL,
  `b` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_a` (`a`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

create table t2 like t1;

-- 插入一些示例数据
-- 往t1表插入1万行记录
DROP PROCEDURE IF EXISTS insert_t1;
DELIMITER ;;

CREATE PROCEDURE insert_t1()
BEGIN
  DECLARE i INT;
  SET i = 1;
  WHILE i <= 10000 DO
    INSERT INTO t1(a, b) VALUES(i, i);
    SET i = i + 1;
  END WHILE;
END;;

DELIMITER ;

CALL insert_t1();


-- 往t2表插入100行记录
DROP PROCEDURE IF EXISTS insert_t2;
DELIMITER ;;

CREATE PROCEDURE insert_t2()
BEGIN
  DECLARE i INT;
  SET i = 1;
  WHILE i <= 100 DO
    INSERT INTO t2(a, b) VALUES(i, i);
    SET i = i + 1;
  END WHILE;
END;;

DELIMITER ;

CALL insert_t2();
```

MySQL 的表关联常见有两种算法：

- Nested-Loop Join 算法
- Block Nested-Loop Join 算法

#### 1. Nested-Loop Join 算法

一次一行循环地从第一张表（称为**驱动表**）中读取行，在这行数据中取到关联字段，根据关联字段在另一张表（**被驱动表**）里取出满足条件的行，然后取出两张表的结果合集。

```sql
EXPLAIN select * from t1 inner join t2 on t1.a= t2.a;
```

![]()

从执行计划中可以看到：

- 驱动表是 t2，被驱动表是 t1。先执行的就是驱动表(执行计划结果的 id 如果一样则按从上到下顺序执行 sql)；**优化器一般会优先选择小表做驱动表。所以使用 inner join 时，排在前面的表并不一定就是驱动表**。
- 如果执行计划 Extra 中未出现 Using join buffer 则表示使用的 join 算法是 NLJ。

SQL 的大致流程如下：
1. 从表 t2 中读取一行数据（如果 t2 表有查询过滤条件的，会从过滤结果里取出一行数据）；
2. 从第 1 步的数据中，取出关联字段 a，到表 t1 中查找；
3. 取出表 t1 中满足条件的行，跟 t2 中获取到的结果合并，作为结果返回给客户端；
4. 重复上面 3 步。

整个过程会读取 t2 表的所有数据(扫描 100 行)，然后遍历这每行数据中字段 a 的值，根据 t2 表中 a 的值索引扫描 t1 表中的对应行(扫描 100 次 t1 表的索引，1 次扫描可以认为最终只扫描 t1 表一行完整数据，也就是总共 t1 表也扫描了 100 行)。因此整个过程扫描了 200 行。

**如果被驱动表的关联字段没索引，使用 NLJ 算法性能会比较低，MySQL 会选择 Block Nested-Loop Join 算法**。

#### 2. Block Nested-Loop Join 算法

**把驱动表的数据读入到 join_buffer 中**，然后扫描被驱动表，把被驱动表每一行取出来跟 **join_buffer 中的所有数据**一起做对比。

```sql
EXPLAIN select * from t1 inner join t2 on t1.b = t2.b;
```

![]()

`Extra` 中 的 `Using join buffer` (Block Nested Loop)说明该关联查询使用的是 BNL 算法。

sql 的大致流程如下：
1. 把 t2 的所有数据放入到 join_buffer 中
2. 把表 t1 中每一行取出来，跟 join_buffer 中的所有数据做对比
3. 返回满足 join 条件的数据


整个过程对表 t1 和 t2 都做了一次全表扫描，因此扫描的总行数为 `10000 (表 t1 的数据总量) + 100 (表 t2 的数据总量) = 10100`。并且 join_buffer 里的数据是无序的，因此对表 t1 中的每一行，都要做 100 次判断，所以内存中的判断次数是 `100 * 10000= 100 万次`。

join_buffer 的大小是由参数 `join_buffer_size` 设定的，默认值是 `256k`。如果放不下表 t2 的所有数据话，策略很简单，就是**分段放**。

比如 t2 表有 1000 行记录， join_buffer 一次只能放 800 行数据，那么执行过程就是先往 join_buffer 里放 800 行记录，然后从 t1 表里取数据跟 join_buffer 中数据对比得到部分结果，然后清空 join_buffer，再放入 t2 表剩余 200 行记录，再次从 t1 表里取数据跟 join_buffer 中数据对比。所以就多扫了一次 t1 表。

**被驱动表的关联字段没索引为什么要选择使用 BNL 算法而不使用 Nested-Loop Join 呢**？

如果上面第二条 sql 使用 Nested-Loop Join，那么扫描行数为 `100 * 10000 = 100 万行`，由于没有 t1.b 是没有索引的，意味这要进行**全表扫描**，这个要在磁盘中扫描 t1 表的所有行。很显然，用 BNL 磁盘扫描次数少很多，相比于磁盘扫描，BNL 的内存计算会快得多。


对于关联 SQL 的优化

- **关联字段加索引**，让 MySQL 做 join 操作时尽量选择 NLJ 算法
- **小表驱动大表**，写多表连接 SQL 时如果明确知道哪张表是小表可以用 straight_join 写法固定连接驱动方式，省去 mysql 优化器自己判断的时间。


`straight_join` 解释：straight_join 功能同 join 类似，但能让左边的表来驱动右边的表，能改表优化器对于联表查询的执行顺序。
比如：`select * from t2 straight_join t1 on t2.a = t1.a`; 代表指定 mysql 选着 t2 表作为驱动表。
**`straight_join`只适用于 inner join**，并不适用于left join，right join。（因为 left join，right join 已经代表指定了表的执行顺序）
尽可能让优化器去判断，因为大部分情况下 mysql 优化器是比人要聪明的。使用 straight_join 一定要慎重，因为部分情况下人为指定的执行顺序并不一定会比优化引擎要靠谱。


**对于小表定义的明确**：

在决定哪个表做驱动表的时候，应该是两个表按照各自的条件过滤，过滤完成之后，计算参与 join 的各个字段的总数据量，数据量小的那个表，就是“小表”，应该作为驱动表。


**join buffer**：

当被驱动表中的数据非常多时，每次访问被驱动表，被驱动表的记录会被加载到内存中，在内存中的每一条记录只会和驱动表结果集的一条记录做匹配，之后就会被从内存中清除掉。然后再从驱动表结果集中拿出另一条记录，再一次把被驱动表的记录加载到内存中一遍，周而复始，**驱动表结果集中有多少条记录，就得把被驱动表从磁盘上加载到内存中多少次**。所以可以在把被驱动表的记录加载到内存的时候，一次性和多条驱动表中的记录做匹配，这样就可以大大减少重复从磁盘上加载被驱动表的代价了。

`join buffer` 就是执行连接查询前申请的一块固定大小的内存，先把**若干条驱动表结果集中的记录装在这个 `join buffer` 中**，然后开始扫描被驱动表，**每一条被驱动表的记录一次性和 join buffer 中的多条驱动表记录做匹配**，因为匹配的过程都是在内存中完成的，所以这样可以显著减少被驱动表的 I/O 代价。

另外需要注意的是，驱动表的记录并不是所有列都会被放到join buffer中，只有查询列表中的列和过滤条件中的列才会被放到join buffer中，所以，**最好不要把 `*` 作为查询列表**，只需要把我们关心的列放到查询列表就好了，这样还可以在 `join buffer` 中放置更多的记录。

####  Hash Join 原理（仅支持等值连接）

MySQL 中的 Hash Join 是一种高效的**等值连接算法**，尤其适合**没有索引、表不太大或临时表**操作的场景。

```sql
SELECT * FROM t1 JOIN t2 ON t1.id = t2.id;
```

Hash Join 分两个阶段进行：

1. Build Phase（构建哈希表）
  - 优化器选择较小的一张表（如 t2）作为构建表（build input）。
  - 把这张表的连接列（如 t2.id）作为 key，构建哈希表，存入内存。
2. Probe Phase（探测匹配项）
  - 遍历另一张较大的表 t1。
  - 以连接键 t1.id 去刚刚构建的哈希表中查找匹配项。

```
-- 探测并输出结果：
for each row in t1:
    if hash_table contains t1.id:
        output (t1, hash_table[t1.id])
```

相对于传统的 Nested Loop Join（嵌套循环），Hash Join 将连接时间复杂度从 O(n*m) 降低到接近 O(n + m)，适合**无索引的中小表等值连接**。


限制：

- **只支持等值连接**，例如 `ON a.id = b.id`，不能用范围条件如 `>`、`<`。
- 大表内存不够会溢出，如果构建的哈希表过大，会使用磁盘上的临时表，性能降低


**in 和 exsits 优化**:
原则：**小表驱动大表**

in：当 B 表的数据集小于 A 表的数据集时，in 优于 exists `select * from A where id in (select id from B)`。

```
#等价于：
for(select id from B){
     select * from A where A.id = B.id
}
```

exists：当 B 表的数据集大于 A 表的数据集时，exists 优于 in `select * from A where exists (select 1 from B where B.id = A.id)`。

`EXISTS` (subquery)只返回TRUE或FALSE,因此子查询中的 `SELECT *` 也可以用 `SELECT 1` 替换,官方说法是实际执行时会忽略 `SELECT` 清单,因此没有区别

```
#等价于：
for(select * from A){
    select * from B where B.id = A.id
}
```



### count(*) 查询优化


```sql
EXPLAIN select count(1) from employees;
EXPLAIN select count(id) from employees;
EXPLAIN select count(name) from employees;
EXPLAIN select count(*) from employees;
```

四个 sql 的执行计划一样，说明这四个 sql 执行效率应该差不多。


字段有索引：`count(*)≈count(1)>count(字段)>count(主键 id)`，字段有索引，`count(字段)` 统计走二级索引，二级索引存储数据比主键索引少，所以 `count(字段)>count(主键 id)`
字段无索引：`count(*)≈count(1)>count(主键 id)>count(字段)`，字段没有索引 `count(字段)` 统计走不了索引，`count(主键 id)` 还可以走主键索引，所以 `count(主键 id)>count(字段)`。

`count(*)` mysql 是专门做了优化，并不会把全部字段取出来，不取值，按行累加，效率很高。

`count(1)` 跟 `count(字段)`执行过程类似，不过 `count(1)` 是用常量 1 做统计，`count(字段)` 还需要取出字段，所以理论上 `count(1)` 比 `count(字段)` 会快一点。

为什么对于 `count(id)`，mysql 最终选择辅助索引而不是主键聚集索引？因为二级索引相对主键索引存储数据更少，检索性能应该更高。


不带 where 条件的常见优化方法：

1. 对于 myisam 存储引擎的表做不带 where 条件的 count 查询性能是很高的，因为 myisam 存储引擎的表的总行数会被 mysql 存储在磁盘上，查询不需要计算。
2. `show table status` 可以看到表的行数，但是这个行数是不准确的。性能很高。例如 `show table status like 'employees'`。
3. 将总数维护到 Redis 里，插入或删除表数据行的时候同时维护 redis 里的表总行数 key 的计数值(用 incr 或 decr 命令)，但是这种方式可能不准，很难保证表操作和redis操作的**事务一致性**
4. 增加数据库计数表，插入或删除表数据行的时候同时维护计数表，让他们在同一个事务里操作