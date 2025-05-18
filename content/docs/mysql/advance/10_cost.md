---
title: 查询成本和扫描区间
weight: 10
---

MySQL 查询优化器决策**是否使用某个索引执行查询时的依据是使用该索引的成本是否足够低**，而**成本很大程度上取决于需要扫描的二级索引记录数量占表中所有记录数量的比例**。

需要扫描的二级索引记录越多，需要执行的回表操作也就越多。如果需要扫描的二级索引记录占全部记录的比例达到某个范围，那优化器就可能选择使用全表扫描的方式执行查询（一个极端的例子就是扫描全部的二级索引记录，那么将对所有的二级索引记录执行回表操作，显然还不如直接全表扫描）。

## 扫描区间和边界条件

示例：

```sql
CREATE TABLE t (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    key1 INT,
    common_field VARCHAR(100),
    PRIMARY KEY (id),
    KEY idx_key1 (key1)
) Engine=InnoDB CHARSET=utf8;

-- 插入一些数据
INSERT INTO t VALUES
    (1, 30, 'b'),
    (2, 80, 'b'),
    (3, 23, 'b'),
    (4, NULL, 'b'),
    (5, 11, 'b'),
    (6, 53, 'b'),
    (7, 63, 'b'),
    (8, NULL, 'b'),
    (9, 99, 'b'),
    (10, 12, 'b'),
    (11, 66, 'b'),
    (12, NULL, 'b'),
    (13, 66, 'b'),
    (14, 30, 'b'),
    (15, 11, 'b'),
    (16, 90, 'b');


SELECT * FROM t WHERE key1 > 20 AND key1 < 50;
```

如果使用 `idx_key1` 执行该查询的话，那么就需要扫描 `key1` 值在 `(20, 50)` 这个区间中的所有二级索引记录，`(20, 50)` 就是 `idx_key1` 执行上述查询时的**扫描区间**，把 `key1 > 20 AND key1 < 50` 称作形成该扫描区间的**边界条件**。

只要索引列和常数使用 `=`、`<=>`、`IN`、`NOT IN`、`IS NULL`、`IS NOT NULL`、`>`、`<`、`>=`、`<=`、`BETWEEN`、`!=` 或者 `LIKE` 操作符连接起来，就可以产生所谓的扫描区间。

- `IN` 操作符的语义和若干个等值匹配操作符 `=` 之间用 `OR` 连接起来的语义是一样的，它们都会产生多个单点扫描区间，比如下边这两个语句的语义上的效果是一样的：

```sql
SELECT * FROM single_table WHERE key1 IN ('a', 'b');

SELECT * FROM single_table WHERE key1 = 'a' OR key1 = 'b';
```

优化器会将 `IN` 子句中的条件看成是 2 个范围区间（虽然这两个区间中都仅仅包含一个值）：

```
['a', 'a']
['b', 'b']
```

- `!=` 产生的扫描区间：

```sql
SELECT * FROM single_table WHERE key1 != 'a';
```

对应的扫描区间就是：`(-∞, 'a')` 和 `('a', +∞)`。

- `LIKE` 操作符比较特殊，例如 `key1 LIKE 'a%'` 形成的扫描区间相当于是 `['a', 'b')`。

## IS NULL、IS NOT NULL、!= 到底能不能用索引？

`s1` 表的聚簇索引示意图：

![id-index-demo](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/id-index-demo.png)

`idx_key1` 二级索引示意图：

![id-key-index-demo](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/id-key-index-demo.png)

图中可以看出，**值为 `NULL` 的二级索引记录都被放到了 B+ 树的最左边**，InnoDB 的设计：`We define the SQL null to be the smallest possible value of a field.` ，也就是认为 **`NULL`值是最小的**。

### IS NULL

示例：

```sql
SELECT * FROM t WHERE key1 IS NULL;
```

优化器在真正执行查询前，会首先少量的访问一下索引，调查一下 `key1` 在 `[NULL, NULL]` 这个区间的记录有多少条。上面的示意图中可以看出需要扫描的二级索引记录占总记录条数的比例是 `3/16`。它觉得这个查询使用二级索引来执行比较靠谱，所以会使用这个 `idx_key1` 索引来执行查询。

### IS NOT NULL

```sql
SELECT * FROM t WHERE key1 IS NOT NULL;
```

`NULL` 作为最小值对待，上面的扫描区间就是 `(NULL, +∞)` 是开区间，也就意味这不包括 `NULL` 值。需要扫描的二级索引记录占总记录条数的比例是 `13/16`，跟显然这个比例已经非常大了，所以会使用全表扫描的方式来执行查询。

现在更新一下数据：

```sql
UPDATE t SET key1 = NULL WHERE key1 < 80;
```

更新后的 `idx_key1` 索引示意图：

![id-key-index-demo2](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/id-key-index-demo.png)

再执行查询：

```sql
SELECT * FROM t WHERE key1 IS NOT NULL;
```

优化器经过调查得知，需要扫描的二级索引记录占总记录条数的比例是 `3/16`，它觉得这个查询使用二级索引来执行比较靠谱，所以会使用这个 `idx_key1` 索引来执行查询。

### !=

```sql
SELECT * FROM t WHERE key1 != 80;
```

优化器在真正执行查询前，会首先少量的访问一下索引，调查一下 `key1` 在 `(NULL, 80)` 和 `(80, +∞)` 这两个区间内记录有多少条：

![id-key-index-demo2](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/id-key-index-demo.png)

可以看出，需要扫描的二级索引记录占总记录条数的比例是 `2/16`，它觉得这个查询使用二级索引来执行比较靠谱，所以会使用这个 `idx_key1` 索引来执行查询。


## 总结

**成本决定执行计划，跟使用什么查询条件并没有什么关系**。优化器会首先针对可能使用到的二级索引划分几个扫描区间，然后分别调查这些区间内有多少条记录，在这些扫描区间内的二级索引记录的总和占总共的记录数量的比例达到某个值时，优化器将放弃使用二级索引执行查询，转而采用全表扫描。