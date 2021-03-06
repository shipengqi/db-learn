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

也就是说可以**直接通过遍历 `idx_key_part` 索引的叶子节点的记录来比较** `key_part2 = 'abc'` 这个条件是否成立，把匹配成功的二级索引记录的 `key_part1`, `key_part2`, `key_part3` 列的值直接加到结果集中就行了。由于**二级索引记录比聚簇索记录小的多**（聚簇索引记录要存储所有用户定义的列以及所谓的隐藏列，而二级索引记录只需要存放索引列和主键，所以每个页中的记录要多很多），而且这个过程也**不用进行回表**操作，所以直接遍历二级索引比直接遍历聚簇索引的成本要小很多，把这种采用**遍历二级索引记录的执行方式**称之为：`index`。

## all

全表扫描

## 索引合并

使用到多个索引来完成一次查询的执行方法称之为：index merge，索引合并算法有下边三种

### intersection

```sql
SELECT * FROM single_table WHERE key1 = 'a' AND key3 = 'b';
```

假设这个查询使用Intersection合并的方式执行的话，那这个过程就是这样的：

- 从idx_key1二级索引对应的B+树中取出`key1 = 'a'`的相关记录。
- 从idx_key3二级索引对应的B+树中取出`key3 = 'b'`的相关记录。
- 二级索引的记录都是由`索引列 + 主键`构成的，所以我们可以计算出这两个结果集中`id`值的交集。
- 按照上一步生成的`id`值列表进行回表操作，也就是从聚簇索引中把指定`id`值的完整用户记录取出来，返回给用户。

为啥不直接使用idx_key1或者idx_key3只根据某个搜索条件去读取一个二级索引，然后回表后再过滤另外一个搜索条件呢？

下两种查询执行方式之间需要的成本代价:

只读取一个二级索引的成本：

- 按照某个搜索条件读取一个二级索引
- 根据从该二级索引得到的主键值进行回表操作，然后再过滤其他的搜索条件

读取多个二级索引之后取交集成本：

- 按照不同的搜索条件分别读取不同的二级索引
- 将从多个二级索引得到的主键值取交集，然后进行回表操作

虽然读取多个二级索引比读取一个二级索引消耗性能，但是**读取二级索引的操作是顺序I/O，而回表操作是随机I/O**，所以如果只读取一个二级索引时需要回表的记录数特别多，而读取多个二级索引之后取交集的记录数非常少，当节省的因为回表而造成的性能损耗比访问多个二级索引带来的性能损耗更高时，读取多个二级索引后取交集比只读取一个二级索引的成本更低。

在某些特定的情况下才可能会使用到Intersection索引合并：

- 情况一：**二级索引列是等值匹配的情况，对于联合索引来说，在联合索引中的每个列都必须等值匹配**，不能出现只匹配部分列的情况。

```sql
SELECT * FROM single_table WHERE key1 = 'a' AND key_part1 = 'a' AND key_part2 = 'b' AND key_part3 = 'c';
```

下边这两个查询就不能进行Intersection索引合并：

```sql
SELECT * FROM single_table WHERE key1 > 'a' AND key_part1 = 'a' AND key_part2 = 'b' AND key_part3 = 'c';

SELECT * FROM single_table WHERE key1 = 'a' AND key_part1 = 'a';
```

- 情况二：主键列可以是范围匹配
比方说下边这个查询可能用到主键和idx_key1进行Intersection索引合并的操作：

```sql
SELECT * FROM single_table WHERE id > 100 AND key1 = 'a';
```

为啥呢？
对于InnoDB的二级索引来说，记录先是按照索引列进行排序，如果该二级索引是一个联合索引，那么会按照联合索引中的各个列依次排序。而二级索引的用户记录是由`索引列 + 主键`构成的，**二级索引列的值相同的记录可能会有好多条，这些索引列的值相同的记录又是按照主键的值进行排序的**。所以重点来了，之所以在二级索引列都是等值匹配的情况下才可能使用Intersection索引合并，是因为**只有在这种情况下根据二级索引查询出的结果集是按照主键值排序的**。

如果从各个二级索引中查询的到的结果集本身就是已经按照主键排好序的，那么求交集的过程就很easy啦。假设某个查询使用Intersection索引合并的方式从idx_key1和idx_key2这两个二级索引中获取到的主键值分别是：

- 从idx_key1中获取到已经排好序的主键值：1、3、5
- 从idx_key2中获取到已经排好序的主键值：2、3、4

那么求交集的过程就是这样：逐个取出这两个结果集中最小的主键值，如果两个值相等，则加入最后的交集结果中，否则丢弃当前较小的主键值，再取该丢弃的主键值所在结果集的后一个主键值来比较，直到某个结果集中的主键值用完了，如果还是觉得不太明白那继续往下看：

- 先取出这两个结果集中较小的主键值做比较，因为1 < 2，所以把idx_key1的结果集的主键值1丢弃，取出后边的3来比较。
- 因为 `3 > 2`，所以把idx_key2的结果集的主键值2丢弃，取出后边的3来比较。
- 因为 `3 = 3`，所以把3加入到最后的交集结果中，继续两个结果集后边的主键值来比较。
- 后边的主键值也不相等，所以最后的交集结果中只包含主键值3。

别看我们写的啰嗦，这个过程其实可快了，时间复杂度是O(n)，但是如果从各个二级索引中查询出的结果集并不是按照主键排序的话，那就要先把结果集中的主键值排序完再来做上边的那个过程，就比较耗时了。

在搜索条件中有主键的范围匹配的情况下也可以使用Intersection索引合并索引合并。为啥主键这就可以范围匹配了？

```sql
SELECT * FROM single_table WHERE key1 = 'a' AND id > 100;
```

二级索引的记录中都带有主键值的，所以可以在从idx_key1中获取到的主键值上直接运用条件id > 100过滤就行了，这样多简单。所以涉及主键的搜索条件只不过是为了从别的二级索引得到的结果集中过滤记录罢了，是不是等值匹配不重要。

情况一、情况二成立，也不一定发生Intersection索引合并，这得看优化器的心情。优化器只有在单独根据搜索条件从某个二级索引中获取的记录数太多，导致回表开销太大，而通过Intersection索引合并后需要回表的记录数大大减少时才会使用Intersection索引合并。

### union

有时候OR关系的不同搜索条件会使用到不同的索引：

```sql
SELECT * FROM single_table WHERE key1 = 'a' OR key3 = 'b'
```

Union是并集的意思，适用于使用不同索引的搜索条件之间使用OR连接起来的情况。与Intersection索引合并类似，MySQL在某些特定的情况下才可能会使用到Union索引合并：

- 情况一：二级索引列是等值匹配的情况，对于联合索引来说，在联合索引中的每个列都必须等值匹配，不能出现只出现匹配部分列的情况。

```sql
SELECT * FROM single_table WHERE key1 = 'a' OR ( key_part1 = 'a' AND key_part2 = 'b' AND key_part3 = 'c');
```

下边这两个查询就不能进行Union索引合并：

```sql
SELECT * FROM single_table WHERE key1 > 'a' OR (key_part1 = 'a' AND key_part2 = 'b' AND key_part3 = 'c');

SELECT * FROM single_table WHERE key1 = 'a' OR key_part1 = 'a';
```

- 情况二：主键列可以是范围匹配

- 情况三：使用Intersection索引合并的搜索条件

就是搜索条件的某些部分使用Intersection索引合并的方式得到的主键集合和其他方式得到的主键集合取交集

```sql
SELECT * FROM single_table WHERE key_part1 = 'a' AND key_part2 = 'b' AND key_part3 = 'c' OR (key1 = 'a' AND key3 = 'b');
```

优化器可能采用这样的方式来执行这个查询：

- 先按照搜索条件key1 = 'a' AND key3 = 'b'从索引idx_key1和idx_key3中使用Intersection索引合并的方式得到一个主键集合。
- 再按照搜索条件key_part1 = 'a' AND key_part2 = 'b' AND key_part3 = 'c'从联合索引idx_key_part中得到另一个主键集合。
- 采用Union索引合并的方式把上述两个主键集合取并集，然后进行回表操作，将结果返回给用户。

通过Union索引合并后进行访问的代价比全表扫描更小时才会使用Union索引合并。

### sort-union

Union索引合并的使用条件太苛刻，必须保证各个二级索引列在进行等值匹配的条件下才可能被用到，比方说下边这个查询就无法使用到Union索引合并：

```sql
SELECT * FROM single_table WHERE key1 < 'a' OR key3 > 'z'
```

这是因为根据 `key1 < 'a'` 从 idx_key1 索引中获取的二级索引记录的主键值不是排好序的，根据 `key3 > 'z'` 从idx_key3索引中获取的二级索引记录的主键值也不是排好序的，但是`key1 < 'a'`和`key3 > 'z'`这两个条件又特别让我们动心，所以我们可以这样：

- 先根据`key1 < 'a'`条件从idx_key1二级索引中获取记录，并按照记录的主键值进行排序
- 再根据`key3 > 'z'`条件从idx_key3二级索引中获取记录，并按照记录的主键值进行排序
- 因为上述的两个二级索引主键值都是排好序的，剩下的操作和Union索引合并方式就一样了。

我们把上述这种先按照二级索引记录的主键值进行排序，之后按照Union索引合并方式执行的方式称之为**Sort-Union索引合并**，很显然，这种Sort-Union索引合并比单纯的Union索引合并多了一步对二级索引记录的主键值排序的过程。

Sort-Union的适用场景是单独根据搜索条件从某个二级索引中获取的记录数比较少，这样即使对这些二级索引记录按照主键值进行排序的成本也不会太高

### 联合索引替代Intersection索引合并

```sql
SELECT * FROM single_table WHERE key1 = 'a' AND key3 = 'b';
```

这个查询之所以可能使用Intersection索引合并的方式执行，还不是因为idx_key1和idx_key3是两个单独的B+树索引，你要是把这两个列搞一个联合索引，那直接使用这个联合索引就把事情搞定了，何必用啥索引合并呢，就像这样：

```sql
ALTER TABLE single_table drop index idx_key1, idx_key3, add index idx_key1_key3(key1, key3);
```

## 索引下推

满足最左前缀原则的时候，最左前缀可以用于在索引中定位记录。这时，你可能要问，那些不符合最左前缀的部分，会怎么样呢？

我们还是以市民表的联合索引（name, age）为例。如果现在有一个需求：检索出表中“名字第一个字是张，而且年龄是10岁的所有男孩”。那么，SQL语句是这么写的：

```sql
select * from tuser where name like '张%' and age=10 and ismale=1;
```

这个语句在搜索索引树的时候，只能用 “张”，找到第一个满足条件的记录ID3。当然，这还不错，总比全表扫描要好。

然后呢？

当然是判断其他条件是否满足。

在MySQL 5.6之前，只能从ID3开始一个个回表。到主键索引上找出数据行，再对比字段值。

而MySQL 5.6 引入的**索引下推**优化（index condition pushdown)， 可以在**索引遍历过程中，对索引中包含的字段先做判断，直接过滤掉不满足条件的记录，减少回表次数**。

![](../../../images/index-down.png)

![](../../../images/index-down2.png)
