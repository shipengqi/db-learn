---
title: 访问方法
---

MySQL 查询的执行方式大致分为两种：

1. 使用全表扫描进行查询

这种执行方式就是把表的每一行记录都扫一嘛，把符合搜索条件的记录加入到结果集就完了。

2. 使用索引进行查询

如果查询语句中的搜索条件可以使用到某个索引，那直接使用索引来执行查询可能会加快查询执行的时间。使用索引来执行查询的方式可以细分为许多种类：

- 针对主键或唯一二级索引的等值查询
- 针对普通二级索引的等值查询
- 针对索引列的范围查询
- 直接扫描整个索引

```sql
CREATE TABLE single_table (
    id INT NOT NULL AUTO_INCREMENT,
    key1 VARCHAR(100),
    key2 INT,
    key3 VARCHAR(100),
    key_part1 VARCHAR(100),
    key_part2 VARCHAR(100),
    key_part3 VARCHAR(100),
    common_field VARCHAR(100),
    PRIMARY KEY (id),
    KEY idx_key1 (key1),
    UNIQUE KEY idx_key2 (key2),
    KEY idx_key3 (key3),
    KEY idx_key_part(key_part1, key_part2, key_part3)
) Engine=InnoDB CHARSET=utf8;
```

- 为 `id`列建立的聚簇索引。
- 为 `key1` 列建立的 `idx_key1` 二级索引。
- 为 `key2` 列建立的 `idx_key2` 二级索引，而且该索引是唯一二级索引。
- 为 `key3` 列建立的 `idx_key3` 二级索引。
- 为 `key_part1`、`key_part2`、`key_part3` 列建立的 `idx_key_part` 二级索引，也是一个联合索引

## const

主键或唯一二级索引的等值查询，

B+ 树叶子节点中的记录是按照索引列排序的，对于的聚簇索引来说，它对应的 B+ 树叶子节点中的记录就是按照 `id` 列排序的。

如果是主键等值查询，那么可以通过聚簇索引直接找到

如果通过唯一二级索引列来定位一条记录，也只比主键多一次**回表**

这种通过主键或者唯一二级索引列来定位一条记录的访问方法定义为：`const`，意思是常数级别的，代价是可以忽略不计的。

对于唯一二级索引来说，查询该列为 NULL 值的情况比较特殊，因为唯一二级索引列并不限制 NULL 值的数量，

```sql
SELECT * FROM single_table WHERE key2 IS NULL;
```

所以上述语句可能访问到多条记录，也就是说上边这个语句不可以使用 const 访问方法来执行

## ref

对某个普通的二级索引列与常数进行等值比较，比如：

```sql
SELECT * FROM single_table WHERE key1 = 'abc';
```

对于这个查询，可以选择全表扫描来逐一对比搜索条件是否满足要求，也可以先使用二级索引找到对应记录的 `id` 值，然后再回表到聚簇索引中查找完整的用户记录。

由于**普通二级索引并不限制索引列值的唯一性，所以可能找到多条对应的记录**，也就是说使用**二级索引来执行查询的代价取决于等值匹配到的二级索引记录条数**。如果匹配的记录较少，则回表的代价还是比较低的，所以 MySQL 可能选择使用索引而不是全表扫描的方式来执行查询。这种搜索条件为二级索引列与常数等值比较，采用二级索引来执行查询的访问方法称为：`ref`。

这种 `ref` 访问方法比 `const` 差了一点，但是在二级索引等值比较时匹配的记录数较少时的效率还是很高的。

## ref_or_null

找出某个二级索引列的值等于某个常数的记录，还想把该列的值为 NULL 的记录也找出来，比如：

```sql
SELECT * FROM single_table WHERE key1 = 'abc' OR key1 IS NULL;
```

当使用二级索引而不是全表扫描的方式执行该查询时，这种类型的查询使用的访问方法就称为 `ref_or_null`

## range

利用索引进行范围匹配的访问方法称之为：`range`。

## index

```sql
SELECT key_part1, key_part2, key_part3 FROM single_table WHERE key_part2 = 'abc';
```

由于 `key_part2` 并不是联合索引 `idx_key_part` 最左索引列，所以我们无法使用 `ref` 或者 `range` 访问方法来执行这个语句。但是这个查询符合下边这两个条件：

它的查询列表只有3个列：`key_part1`, `key_part2`, `key_part3`，而索引 `idx_key_part` 又包含这三个列。

搜索条件中只有 `key_part2` 列。这个列也包含在索引 `idx_key_part` 中。

也就是说可以**直接通过遍历 `idx_key_part` 索引的叶子节点的记录来比较** `key_part2 = 'abc'` 这个条件是否成立，把匹配成功的二级索引记录的 `key_part1`, `key_part2`, `key_part3` 列的值直接加到结果集中就行了。由于二级索引记录比聚簇索记录小的多（聚簇索引记录要存储所有用户定义的列以及所谓的隐藏列，而二级索引记录只需要存放索引列和主键），而且这个过程也**不用进行回表**操作，所以直接遍历二级索引比直接遍历聚簇索引的成本要小很多，把这种采用**遍历二级索引记录的执行方式**称之为：`index`。

## all

全表扫描

## 索引合并

### intersection

### union

### sort-union
