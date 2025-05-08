---
title: B+ 树索引
weight: 6
---

B+ 树是目前为止排序最有效率的数据结构。像二叉树，哈希索引、红黑树、SkipList，在海量数据基于磁盘存储效率方面远不如 B+ 树索引高效。

**B+ 树索引的特点是**： 基于磁盘的平衡树，但**树非常矮**，通常为 3~4 层，能存放千万到上亿的排序数据。树矮意味着访问效率高，从千万或上亿数据里查询一条数据，只需要 3、4 次 I/O。

B+ 树（B 树变种）：

- 非叶子节点不存储 data，只存储索引(冗余)，可以放更多的索引。B 树的非叶子节点也存储 data，会占用更多的磁盘空间，每个非叶子节点存储的记录远少于 B+ 树，这会导致树的高度变高，磁盘 I/O 次数变多，查询效率变低。
- 叶子节点包含所有索引字段
- 叶子节点用指针连接，提高区间访问的性能，B 树的叶子节点由于没有这个指针，一次只能访问一个节点，然后再从根节点开始遍历。

为什么要 InnoDB 要建主键？

InnoDB 要求表必须有主键（MyISAM 可以没有），如果没有显式指定，则 MySQL 系统会自动选择一个可以**非空唯一**标识数据记录的列作为主键。如果不存在这种列，则 MySQL 自动为 InnoDB 表生成一个**隐藏字段**作为主键，这个字段长度为 6 个字节，类型为长整型。

为什么主键推荐使用整形的自增主键？

- 自增主键可以保证数据的顺序插入，减少磁盘 I/O 次数。如果不是自增的话，每次插入数据都可能会导致页分裂，和数据的重新排序。而顺序插入的话，只需要插入到当前页的末尾，当页满了之后，再建一个新页并插入到下一个页中。
- 自增主键可以保证数据的唯一性，避免重复插入。

至于为什么推荐使用整形，而不是字符串，是因为：

- 整形的比较操作比字符串的比较操作要快。
- 整形的存储空间比字符串的存储空间要小。

## 索引的数据结构

- 二叉树
- 红黑树
- Hash 表
- B-Tree

### Hash 索引

Hash 索引是一种基于哈希表的数据结构，它将索引键值与存储在表中的数据行的物理地址相关联。

Hash 索引的特点是：

- 哈希索引是一种非聚集索引，它将索引的键值与存储在表中的数据行的物理地址相关联。
- 哈希索引的查询速度非常快，因为它是基于哈希表实现的，一般一次磁盘 I/O 就可以找到对应的记录。
- 哈希索引**不支持范围查询**，因为哈希表是无序的。

1. InnoDB 的各个数据页会组成一个**双向链表**。
2. 每个数据页为 `User Records` 中的记录划分**组**，并生成 `Page Directory`。
3. `Page Directory` 中的每个**槽**对应一个**分组的最后一条记录**。
4. 通过主键查找某条记录的时，在 `Page Directory` 中使用**二分法**快速定位到对应的**槽**，然后再遍历该槽对应分组中的记录即可快速找到指定的记录。

![innodb-data-page-link](../../../images/innodb-data-page-link.jpg)

## 没有索引的查找

没有索引的时候是怎么查找记录的？比如：

```sql
SELECT [列名列表] FROM 表名 WHERE 列名 = xxx;
```

### 在一个页中的查找

如果表中的记录比较少，所有的记录都可以被存放到一个页中，在查找记录的时候可以根据搜索条件的不同分为两种情况：

- 以主键为搜索条件

在页目录中使用二分法快速定位到对应的槽，然后再遍历该槽对应分组中的记录即可快速找到指定的记录。

- 以其他列作为搜索条件

对非主键列，数据页中并**没有对非主键列建立所谓的页目录，所以无法通过二分法快速定位相应的槽。这种情况下只能从最小记录开始依次遍历单链表中的每条记录**，然后对比每条记录是不是符合搜索条件。很显然，这种查找的效率是非常低的。

### 在很多页中查找

大部分情况下表中存放的记录都是非常多的，需要好多的数据页来存储这些记录。在很多页中查找记录的话可以分为两个步骤：

1. 定位到记录所在的页。
2. 从所在的页内中查找相应的记录。

由于并不能快速的定位到记录所在的页，所以只能从第一个页沿着双向链表一直往下找，每一个页中再使用上面一个页中的查找方法。非常低效。

## 索引

```sh
mysql> CREATE TABLE index_demo(
    ->     c1 INT,
    ->     c2 INT,
    ->     c3 CHAR(1),
    ->     PRIMARY KEY(c1)
    -> ) ROW_FORMAT = Compact;
Query OK, 0 rows affected (0.03 sec)
```

### 简单的索引方案

根据某个搜索条件查找一些记录时为什么要遍历所有的数据页？

因为各个页中的记录并没有规律，我们并不知道搜索条件匹配哪些页中的记录，所以不得不依次遍历所有的数据页。

那么，如何快速定位记录所在的数据页？

我们可以创建 `Page directoty` 来快速定位页中的记录，也可以建一个类似的目录来定位记录所在的页：

### B+ 树

#### 为什么是 B+ 树

哈希表是一种以键-值（key-value）存储数据的结构，插入和查询都很快，但是适用于只有等值查询的场景，范围查询很慢。

有序数组在等值查询（二分法）和范围查询场景中的性能就都非常优秀。如果仅仅看查询效率，有序数组就是最好的数据结构了。但是，在需要更新数据的时候就麻烦了，你往中间插入一个记录就必须得挪动后面所有的记录，成本太高。

#### 聚簇索引

1. 使用记录主键值的大小进行记录和页的排序
2. B+ 树的叶子节点存储完整的用户记录（记录中存储了所有列的值，包括隐藏列）。

具有这两种特性的 B+ 树称为**聚簇索引**，所有完整的用户记录都存放在这个聚簇索引的叶子节点处。这就是所谓的**索引即数据，数据即索引**。

#### 二级索引

**非主键列**建立的 B+ 树需要一次**回表**操作才可以定位到完整的用户记录，所以这种 B+ 树也被称为**二级索引**（secondary index），或者**辅助索引**。

二级索引的叶子节点包含的用户记录由 `索引列 + 主键` 组成。

为什么二级索引的叶子节点只存储主键？

- 节省空间，如果所有的列都存储在二级索引的叶子节点中，那么二级索引的叶子节点就会非常大，占用的空间也会非常大。
- 一致性问题，如果二级索引的叶子节点中存储的是完整的用户记录，那么当用户记录发生变化时，所有二级索引的叶子节点也需要发生变化。

#### 联合索引

联合索引，本质上也是一个二级索引。

### B+ 树索引的注意事项

**一个 B+ 树索引的根节点自诞生之日起，便不会再移动**。根节点一旦建立，它的页号便会被记录到某个地方，然后 InnoDB 存储引擎需要用到这个索引的时候，都会从那个固定的地方取出根节点的页号，从而来访问这个索引。

B+ 树的形成过程：

- 当为某个表创建一个 B+ 树索引时，都会为这个索引创建一个**根节点**页。最开始表中没有数据的时候，每个 B+ 树索引对应的根节点中既没有用户记录，也没有目录项记录。
- 随后向表中插入用户记录时，先把用户记录存储到这个根节点中。
- 当根节点中的可用空间用完时继续插入记录，此时会将根节点中的所有记录复制到一个新分配的页，比如 `页 a` 中，然后对这个新页进行**页分裂**的操作，得到另一个新页，比如 `页 b`。这时新插入的记录根据键值（也就是聚簇索引中的主键值，二级索引中对应的索引列的值）的大小就会被分配到 `页 a` 或者 `页 b` 中，而**根节点升级为存储目录项记录的页**。

**B+ 树的同一层内节点的目录项记录除页号这个字段以外是唯一的**。所以对于二级索引的内节点的目录项记录的内容实际上是由三个部分构成的：

- 索引列的值
- 主键值
- 页号

把主键值也添加到二级索引内节点中的目录项记录了，这样就能保证 B+ 树每一层节点中各条目录项记录除页号这个字段外是**唯一**的。

**一个页面最少存储 2 条记录**。

## Explain 分析

Explain 关键字，查询上设置一个标记，执行查询会返回执行计划的信息，而不是执行 SQL。

注意：如果 `from` 中包含子查询，仍会执行该子查询，将结果放入临时表中。

Explain 每个列的信息：

- id：select 的序列号，表示查询中执行 select 子句或操作表的顺序。有几个 select 就有几个 id，并且 id 的顺序是按 select 出现的顺序增长的。id 列越大执行优先级越高，id 相同则从上往下执行，id 为 NULL 最后执行。
- select_type：select 类型，主要是用于区别普通查询、联合查询、子查询等复杂的查询类型。
  - simple：简单的 select 查询，查询中不包含子查询或者 UNION。
  - primary：查询中若包含任何复杂的子部分，最外层查询则被标记为 primary。
  - subquery：在 select 或 where 列表中包含了子查询。
  - derived：在 from 列表中包含的子查询被标记为 derived（衍生），MySQL 会递归执行这些子查询，把结果放到临时表中。
  - union：若第二个 select 出现在 union 之后，则被标记为 union；若 union 包含在 from 子句的子查询中，外层 select 将被标记为 derived。
- table：表示 explain 的一行正在访问哪个表。
- partitions：如果查询是基于分区表的话，partitions 字段会显示查询将访问的分区。
- type：这一列表示**关联类型或访问类型**，即 MySQL 决定如何查找表中的行，查找数据行记录的大概范围。依次从最优到最差分别为：`system` > `const` > `eq_ref` > `ref` > `range` > `index` > `ALL`。一般来说，要**保证查询达到 range 级别，最好达到 ref**。
  - const，system：表示通过索引一次就找到了，const 用于比较 primary key 或者 unique 索引。因为只匹配一行数据，所以很快。system 是 const 类型的特列，平时不会出现，这个也可以忽略不计。
  - eq_ref：primary key 或 unique key 索引的所有部分被连接使用 ，最多只会返回一条符合条件的记录。这可能是在 const 之外最好的联接类型了，简单的 select 查询不会出现这种 type。
  - ref：相比 eq_ref，不使用唯一索引，而是使用普通索引或者唯一性索引的部分前缀，索引要和某个值相比较，可能会找到多个符合条件的行。
  - range：范围扫描通常出现在 `in(), between ,> ,<, >=` 等操作中。使用一个索引来检索给定范围的行。
  - index：扫描全索引就能拿到结果，一般是扫描某个二级索引，直接对二级索引的叶子节点遍历和扫描，速度还是比较慢的，这种查询一般为使用覆盖索引，二级索引一般比较小，所以这种通常比 ALL 快一些。
  - ALL：Full Table Scan，即全表扫描，直接扫描聚簇索引的所有叶子节点。通常情况下这需要增加索引来进行优化了。
- possible_keys：查询时，可能使用的索引。如果 possible_keys 有列，而 key 为 NULL，这种情况是因为表中数据不多，mysql 认为索引对此查询帮助不大，选择了全表查询。
- key：实际使用的索引。如果为 NULL，则没有使用索引。如果想强制 mysql 使用或忽视 possible_keys 列中的索引，在查询中使用 force index、ignore index。
- key_len：显示 mysql 在索引里使用的字节数，通过这个值可以算出具体使用了索引中的哪些列。例如，film_actor 的联合索引 idx_film_actor_id 由 film_id 和 actor_id 两个 int 列组成，并且每个 int 是 4 字节。通过结果中的 `key_len=4` 可推断出查询使用了第一个列：film_id 列来执行索引查找。
  - key_len 计算规则如下：
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
    - 索引最大长度是 768 字节，当字符串过长时，mysql 会做一个类似左前缀索引的处理，将前半部分的字符提取出来做索引
- ref：显示索引的些列或常量被使用了。
- rows：是 mysql 估计要读取并检测的行数，注意这个不是结果集里的行数。
- filtered：该列是一个百分比的值，`rows*filtered/100` 可以估算出将要和 explain 中前一个表进行连接的行数。
- extra：展示的是额外信息。
  - Using index：使用**覆盖索引**，避免访问了表的数据行，效率不错。
  - Using where：表示使用了 where 过滤，并且**查询的列未被索引覆盖**。
  - Usering index condition：表示使用了**索引下推**优化。
  - Using temporary：要创建一张临时表来处理查询。出现这种情况一般是要进行优化的，首先是想到用索引来优化。
  - Using filesort：将用外部排序而不是索引排序，数据较小时在内存排序，否则需要在磁盘完成排序。这种情况下一般也是要考虑使用索引来优化。
  - Using join buffer：使用连接缓存。
  - Select tables optimized away：使用某些聚合函数（比如 max、min）来访问存在索引的某个字段。

## 索引的代价

- 空间上的代价

每建立一个索引都要为它建立一棵 B+ 树，每一棵 B+ 树的每一个节点都是一个数据页，一个页默认会占用 16KB 的存储空间，一棵很大的 B+ 树会消耗很大的一片存储空间。

- 时间上的代价

每次对表中的数据进行增、删、改操作时，都需要去修改各个 B+ 树索引。**B+ 树每层节点都是按照索引列的值从小到大的顺序排序而组成了双向链表**。不论是叶子节点中的记录，还是非叶**节点中的记录都是按照索引列的值从小到大的顺序而形成了一个单向链表**。而增、删、改操作可能会对节点和记录的排序造成破坏，所以存储引擎需要**额外的时间进行一些记录移位，页面分裂、页面回收等操作来维护好节点和记录的排序**。

一个表上索引建的越多，就会占用越多的存储空间，在增删改记录的时候性能就越差。

## 索引的使用

### B+ 树索引适用的条件

联合索引的各个排序列的排序顺序必须是一致的

```sql
CREATE TABLE person_info(
    id INT NOT NULL auto_increment,
    name VARCHAR(100) NOT NULL,
    birthday DATE NOT NULL,
    phone_number CHAR(11) NOT NULL,
    country varchar(100) NOT NULL,
    PRIMARY KEY (id),
    KEY idx_name_birthday_phone_number (name, birthday, phone_number)
);
```

二级索引 `idx_name_birthday_phone_number`，它是由 3 个列组成的联合索引。所以在这个索引对应的 B+ 树的叶子节点处存储的用户记录只保留 `name`、`birthday`、`phone_number` 这三个列的值以及主键 `id` 的值。

这个 `idx_name_birthday_phone_number` 索引对应的 B+ 树中页面和记录的排序方式就是这样的：

- 先按照 name 列的值进行排序。
- 如果 `name` 列的值相同，则按照 `birthday` 列的值进行排序。
- 如果 `birthday` 列的值也相同，则按照 `phone_number` 的值进行排序。

这个排序方式非常重要，因为**只要页面和记录是排好序的，就可以通过二分法来快速定位查找**。

### 全值匹配

当查询条件中的列与索引中的列完全匹配，并且全部使用等值比较（`=`）时，称为**全值匹配**。这种情况下，MySQL 可以最有效地利用索引。

```sql
SELECT * FROM person_info WHERE name = 'Ashburn' AND birthday = '1990-09-27' AND phone_number = '15123983239';
```

`idx_name_birthday_phone_number` 索引包含的 3 个列，查询过程：

1. B+ 树的数据页和记录是先按照 name 列的值进行排序的，可以先按照 name 列来查找。
2. name 列相同的记录又是按照 birthday 进行排序的，可以继续按照 birthday 来查找。
3. 如果 name 和 birthday 都是相同的，会按照 `phone_number` 列的值排序。

`name`、`birthday`、`phone_number` 这几个搜索列的顺序对查询结果没有影响，因为优化器可以优化语句。

### 匹配左边的列

```sql
SELECT * FROM person_info WHERE name = 'Ashburn';

SELECT * FROM person_info WHERE name = 'Ashburn' AND birthday = '1990-09-27';
```

没有包含全部联合索引的列，只要包含左边的一列或者多列，也可以使用索引。

因为 B+ 树的联合索引按照索引从左到右的顺序排序的。

**如果想使用联合索引中尽可能多的列，搜索条件中的各个列必须是联合索引中从最左边连续的列**。

### 匹配列前缀

字符串排序的本质就是比较哪个字符串大一点儿，哪个字符串小一点，比较字符串大小就用到了该列的字符集和比较规则。

比较两个字符串的大小的过程其实是这样的：

- 先比较字符串的第一个字符，第一个字符小的那个字符串就比较小。
- 如果两个字符串的第一个字符相同，那就再比较第二个字符，第二个字符比较小的那个字符串就比较小。
- 如果两个字符串的第二个字符也相同，那就接着比较第三个字符，依此类推。

所以一个排好序的字符串列其实有这样的特点：

- 先按照字符串的第一个字符进行排序。
- 如果第一个字符相同再按照第二个字符进行排序。
- 如果第二个字符相同再按照第三个字符进行排序，依此类推。

也就是说这些字符串的前 n 个字符，也就是前缀都是排好序的，所以对于字符串类型的索引列来说，我们只匹配它的前缀也是可以快速定位记录的，比方说我们想查询名字以'As'开头的记录，那就可以这么写查询语句：

```sql
SELECT * FROM person_info WHERE name LIKE 'As%';
```

但是如果只给出后缀或者中间某个字符串，如 `%As%`，就没办法利用索引了。

### 匹配范围值

**所有记录都是按照索引列的值从小到大的顺序排好序的**，所以查找某个范围的值的记录是很简单的。

```sql
SELECT * FROM person_info WHERE name > 'Asa' AND name < 'Barlow';
```

记录是先按 name 列排序的，所以我们上边的查询过程其实是这样的：

- 找到 name 值为 Asa 的记录。
- 找到 name 值为 Barlow 的记录。
- 由于所有记录都是由链表连起来的（记录之间用单链表，数据页之间用双链表），所以他们之间的记录都可以很容易的取出来。
- 找到这些记录的主键值，再到**聚簇索引中回表**查找完整的记录。

**如果对多个列同时进行范围查找的话，只有对索引最左边的那个列进行范围查找的时候才能用到 B+ 树索引**。

```sql
SELECT * FROM person_info WHERE name > 'Asa' AND name < 'Barlow' AND birthday > '1980-01-01';
```

查询可以分成两个部分：

1. 通过条件 `name > 'Asa' AND name < 'Barlow'` 来对 name 进行范围查找，查找的结果可能有多条 name 值不同的记录，
1. 对这些 name 值不同的记录继续通过 `birthday > '1980-01-01'` 条件继续过滤。

对于联合索引 `idx_name_birthday_phone_number` 来说，**只能用到 name 列的部分，而用不到 birthday 列的部分，因为只有 name 值相同的情况下才能用 birthday 列的值进行排序**。

### 精确匹配某一列并范围匹配另外一列

对于同一个联合索引来说，虽然对多个列都进行范围查找时只能用到最左边那个索引列，但是如果左边的列是精确查找，则右边的列可以进行范围查找，比方说这样：

```sql
SELECT * FROM person_info WHERE name = 'Ashburn' AND birthday > '1980-01-01' AND birthday < '2000-12-31' AND phone_number > '15100000000';
```

`name = 'Ashburn'`，对 name 列进行精确查找，使用 B+ 树索引。

`birthday > '1980-01-01' AND birthday < '2000-12-31'`，由于 name 列是精确查找，所以通过 `name = 'Ashburn'`条件查找后得到的结果的 name 值都是相同的，它们会再按照 birthday 的值进行排序。所以此时对 birthday 列进行范围查找是可以用到 B+ 树索引的。

`phone_number > '15100000000'`，通过 birthday 的范围查找的记录的 birthday 的值可能不同，所以这个条件无法再利用 B+ 树索引了，只能遍历上一步查询得到的记录。

### 用于排序

一般情况下，我们只能把记录都加载到内存中，再用一些排序算法，比如快速排序、归并排序、吧啦吧啦排序等等在内存中对这些记录进行排序，有的时候可能查询的结果集太大以至于不能在内存中进行排序的话，还可能暂时借助磁盘的空间来存放中间结果，排序操作完成后再把排好序的结果集返回到客户端。在 MySQL 中，把这种在内存中或者磁盘上进行排序的方式统称为**文件排序**。

如果 `ORDER BY` 子句里使用到了我们的索引列，就有可能省去在内存或文件中排序的步骤，比如下边这个简单的查询语句：

```sql
SELECT * FROM person_info ORDER BY name, birthday, phone_number LIMIT 10;
```

这个查询的结果集需要先按照 name 值排序，如果记录的 name 值相同，则需要按照 birthday 来排序，如果 birthday 的值相同，则需要按照 `phone_number` 排序。大家可以回过头去看我们建立的 `idx_name_birthday_phone_number` 索引的示意图，因为这个 B+树索引本身就是按照上述规则排好序的，所以直接从索引中提取数据，然后进行回表操作取出该索引中不包含的列就好了。

注意，ORDER BY 的子句后边的列的顺序也必须按照索引列的顺序给出，如果给出 `ORDER BY phone_number, birthday, name` 的顺序，那也是用不了 B+树索引

同理，`ORDER BY name、ORDER BY name, birthday` 这种匹配索引左边的列的形式可以使用部分的 B+树索引。当联合索引左边列的值为常量，也可以使用后边的列进行排序

```sql
SELECT * FROM person_info WHERE name = 'A' ORDER BY birthday, phone_number LIMIT 10;
```

不可以使用索引进行排序:

- ASC、DESC 混用
- 排序列包含非同一个索引的列
- 排序列使用了复杂的表达式 `SELECT * FROM person_info ORDER BY UPPER(name) LIMIT 10;`

### 用于分组

```sql
SELECT name, birthday, phone_number, COUNT(*) FROM person_info GROUP BY name, birthday, phone_number
```

先把记录按照 name 值进行分组，所有 name 值相同的记录划分为一组。

将每个 name 值相同的分组里的记录再按照 birthday 的值进行分组，将 birthday 值相同的记录放到一个小分组里，所以看起来就像在一个大分组里又化分了好多小分组。

再将上一步中产生的小分组按照 phone_number 的值分成更小的分组，所以整体上看起来就像是先把记录分成一个大分组，然后把大分组分成若干个小分组，然后把若干个小分组再细分成更多的小小分组。

如果没有索引的话，这个分组过程全部需要在内存里实现，而如果有了索引的话，恰巧这个分组顺序又和我们的 B+树中的索引列的顺序是一致的，而我们的 B+树索引又是按照索引列排好序的，这不正好么，所以可以直接使用 B+树索引进行分组。

### 回表的代价

`idx_name_birthday_phone_number` 索引为例

```sql
SELECT * FROM person_info WHERE name > 'Asa' AND name < 'Barlow';
```

索引 `idx_name_birthday_phone_number` 对应的 B+ 树用户记录中只包含 `name`、`birthday`、`phone_number`、`id` 这 4 个字段，而查询列表是 `*`，意味着要查询表中所有字段。这时需要把从上一步中获取到的每一条记录的 id 字段都到聚簇索引对应的 B+ 树中找到完整的用户记录，也就是我们通常所说的**回表**，然后把完整的用户记录返回给查询用户。

#### 顺序 I/O

索引 `idx_name_birthday_phone_number` 对应的 B+ 树中的记录首先会按照 `name` 列的值进行排序，所以值在 `Asa～Barlow` 之间的记录在磁盘中的存储是相连的，集中分布在一个或几个数据页中，我们可以很快的把这些连着的记录从磁盘中读出来，这种读取方式我们也可以称为**顺序 I/O**

#### 随机 I/O

根据第 1 步中获取到的记录的 id 字段的值可能并不相连，而在聚簇索引中记录是根据 `id`（也就是主键）的顺序排列的，所以根据这些并不连续的 `id` 值到聚簇索引中访问完整的用户记录可能分布在不同的数据页中，这样读取完整的用户记录可能要访问更多的数据页，这种读取方式我们也可以称为**随机 I/O**

顺序 I/O 比随机 I/O 的性能高很多，所以步骤 1 的执行可能很快，而步骤 2 就慢一些。

**需要回表的记录越多，使用二级索引的性能就越低**。某些查询宁愿使用全表扫描也不使用二级索引。比方说 name 值在 Asa ～ Barlow 之间的用户记录数量占全部记录数量 90% 以上，那么如果使用 `idx_name_birthday_phone_number` 索引的话，有 90% 多的 id 值需要回表，这不是吃力不讨好么，还不如直接去扫描聚簇索引（也就是全表扫描）。

查询优化器会事先对表中的记录计算一些统计数据，然后再利用这些统计数据根据查询的条件来计算一下需要回表的记录数，需要回表的记录数越多，就越倾向于使用全表扫描，反之倾向于使用二级索引 + 回表的方式。

回表的记录特别少，优化器就会倾向于使用二级索引 + 回表的方式执行查询。

### 覆盖索引

为了彻底告别回表操作带来的性能损耗：**最好在查询列表里只包含索引列**

```sql
SELECT name, birthday, phone_number FROM person_info WHERE name > 'Asa' AND name < 'Barlow'
```

只查询 name, birthday, phone_number 这三个索引列的值，所以在通过 idx_name_birthday_phone_number 索引得到结果后就不必到聚簇索引中再查找记录的剩余列，这样就省去了回表操作带来的性能损耗。

**不鼓励用 `*` 号作为查询列表，最好把我们需要查询的列依次标明**。

### 索引下推

索引下推（Index Condition Pushdown, ICP）的基本思想是：改变 WHERE 条件在查询执行过程中的处理位置，在索引遍历过程中，就对索引中包含的字段先做判断，直接过滤掉不满足条件的记录，减少回表次数。优先使用覆盖索引，避免回表。

**传统查询流程（无 ICP）**：

1. 存储引擎通过索引定位记录
2. 通过主键回表获取完整数据行
3. 将所有数据返回给 MySQL 服务器层
4. 服务器层应用 WHERE 条件进行过滤

**使用 ICP 的查询流程**：
1. 存储引擎通过索引定位记录
2. 在存储引擎层就检查 WHERE 条件中可以使用索引评估的部分
3. 只对满足条件的记录执行回表操作
4. 将过滤后的数据返回给服务器层

```sql
-- 联合索引 (name,age,position)
SELECT * FROM employees WHERE name like 'LiLei%' AND age = 22 AND position ='manager';
```

在MySQL 5.6之 前的版本没有 ICP，这个查询只能在联合索引里匹配到名字是 'LiLei' 开头的索引，然后拿这些索引对应的主键逐个回表，到主键索引上找出相应的记录，服务器层再根据 `age` 和 `position` 的过滤条件进行筛选。这种情况只会走 `name` 字段索引，无法很好的利用索引。

MySQL 5.6 引入了索引下推优化 ICP，上面那个查询在联合索引里匹配到名字是 'LiLei' 开头的索引之后，同时还会在索引里过滤 `age` 和 `position` 这两个字段，拿着过滤完剩下的索引对应的主键 `id` 再回表查整行数据。

#### 为什么范围查找 MySQL 没有用索引下推优化？

可能是 MySQL 认为索引下推需要额外的判断，范围查找过滤的结果集过大，会导致更多的计算，`like KK%` 在绝大多数情况来看，过滤后的结果集比较小，所以这里 MySQL 选择给 `like KK%` 用了索引下推优化，当然这也不是绝对的，有时 `like KK%` 也不一定就会走索引下推。

### 不在索引列上做任何操作（计算、函数、（自动 or 手动）类型转换），会导致索引失效而转向全表扫描

```sql
SELECT * FROM person_info WHERE left(name,3) = 'LiLei';
```

索引列上做了函数操作，得到的结果在索引树上是无法匹配的，所以索引失效了。`left(name,3)` 有点类似 `LiL%` 的效果，但是 MySQL 并没有对这种情况做优化，所以索引失效了。

### is null,is not null 一般情况下也无法使用索引

### 范围查询优化

```sql
select * from employees where id >=1 and id <=20000;
```

mysql 内部优化器会根据检索比例、表大小等多个因素整体评估是否使用索引。比如这个例子，可能是由于单次数据量查询过大导致优化器最终选择不走索引。

优化方法：可以将大的范围拆分成多个小范围：

```sql
select * from employees where id >=1 and id <=5000;
select * from employees where id >=5001 and id <=10000;
select * from employees where id >=10001 and id <=15000;
select * from employees where id >=15001 and id <=20000;
```


### 如何挑选索引

索引设计的核心思想就是尽量利用一两个复杂的多字段联合索引，抗下 80% 以上的查询，然后用一两个辅助索引尽量抗下剩余的一些非典型查询，保证这种大数据量表的查询尽可能多的都能充分利用索引，这样就能保证查询速度和性能了。

#### 代码先行，索引后上

一般应该等到主体业务功能开发完毕，把涉及到该表相关sql都要拿出来分析之后再建立索引。

#### 联合索引尽量覆盖条件

可以设计一个或者两三个联合索引(尽量少建单值索引)，让每一个联合索引都尽量去包含 sql 语句里的 where、order by、group by 的字段，还要确保这些联合索引的字段顺序尽量满足 sql 查询的最左前缀原则。

联合索引中的某个字段通常是范围查找，最好把这个字段放在联合索引的最后面。因为一般情况下，范围查找之后的字段就无法走索引了。

示例：

```sql
-- 联合索引 (province,city,sex)
SELECT * FROM users WHERE province = xx AND city = xx AND age <= xx AND age >= xx;

-- 范围查找字段放在联合索引的最后面
-- 联合索引 (province,city,sex,age)
-- 索引修改后，会导致上面的 sql age 无法走索引，可以优化为
SELECT * FROM users WHERE province = xx AND city = xx AND sex in ('female','male') AND age <= xx AND age >= xx;
```

假设可能还有一个筛选条件，比如要筛选最近一周登录过的用户，对应后台 sql 可能是这样：

```sql
where province=xx and city=xx and sex in ('female','male') and age>=xx and age<=xx and
latest_login_time>= xx
```
`latest_login_time` 也是一个范围查找字段，如果把它放在联合索引里，如 `(province,city,sex,hobby,age,latest_login_time)`，显然是不行的。可以换一种思路，设计一个字段 `is_login_in_latest_7_days`，用户如果一周内有登录值就为 1，否则为 0，那么就可以把索引设计成 `(province,city,sex,hobby,is_login_in_latest_7_days,age)` 来满足上面那种场景。



#### 只为用于搜索、排序或分组的列创建索引

#### 不要在小基数字段上建立索引

索引基数是指这个字段在表里总共有多少个不同的值，比如一张表总共 100 万行记录，其中有个性别字段，其值不是男就是女，那么该字段的基数就是 2。对这种小基数字段建立索引的话，还不如全表扫描。因为索引树里就包含男和女两种值，根本没
法进行快速的二分查找，那用索引就没有太大的意义了。

#### 索引列的类型尽量小

尽量对字段类型较小的列设计索引，因为字段类型较小的话，占用磁盘空间也会比较小。

以整数类型为例，有 `TINYINT`、`MEDIUMINT`、`INT`、`BIGINT` 这么几种，它们占用的存储空间依次递增，我们这里所说的**类型大小指的就是该类型表示的数据范围的大小**。

在表示的整数范围允许的情况下，**尽量让索引列使用较小的类型**：

- 数据类型越小，在查询时进行的比较操作越快（这是 CPU 层次的东东）
- 数据类型越小，**索引占用的存储空间就越少，在一个数据页内就可以放下更多的记录，从而减少磁盘 I/O 带来的性能损耗，也就意味着可以把更多的数据页缓存在内存中，从而加快读写效率**。

#### 长字符串我们可以采用前缀索引

假设我们的字符串很长，那存储一个字符串就需要占用很大的存储空间。

例如，`varchar(100)` 这种大字段建立索引，可以稍微优化下，比如针对这个字段的前 20 个字符建立索引，就是说，对这个字段里的每个值的前 20 个字符放在索引树里。

这样在根据 name 字段来搜索记录时虽然不能精确的定位到记录的位置，但是能定位到相应前缀所在的位置，然后根据前缀相同的记录的主键值回表查询完整的字符串值，再对比就好了。

```sql
CREATE TABLE person_info(
    name VARCHAR(100) NOT NULL,
    birthday DATE NOT NULL,
    phone_number CHAR(11) NOT NULL,
    country varchar(100) NOT NULL,
    KEY idx_name_birthday_phone_number (name(10), birthday, phone_number)
);
```

但是对于 order by name，那么此时你的 name 因为在索引树里仅仅包含了前 20 个字符，无法对后边的字符不同的记录进行排序， group by 也是同理。

#### 让索引列在比较表达式中单独出现

整数列 my_col，我们为这个列建立了索引

`WHERE my_col * 2 < 4` 是以 `my_col * 2` 这样的表达式的形式出现的，存储引擎会依次**遍历所有的记录**，计算这个表达式的值是不是小于 4

`WHERE my_col < 4/2` my_col 列并是以单独列的形式出现的，这样的情况可以直接使用 B+ 树索引。

**如果索引列在比较表达式中不是以单独列的形式出现，而是以某个表达式，或者函数调用形式出现的话，是用不到索引的**。

#### 主键插入顺序

据页和记录又是按照记录主键值从小到大的顺序进行排序，所以如果我们插入的记录的主键值是依次增大的话，那我们每插满一个数据页就换到下一个数据页继续插，而如果我们插入的主键值忽大忽小的话，可能需要**页面分裂和记录移位**。意味着：**性能损耗**。

最好让插入的记录的主键值依次递增，这样就不会发生这样的性能损耗了。所以我们建议：**让主键具有 AUTO_INCREMENT，让存储引擎自己为表生成主键**，而不是我们手动插入

#### 冗余和重复索引

## 索引选择异常和处理

一种方法是，采用 force index 强行选择一个索引。

```sql
set long_query_time=0;
select * from t where a between 10000 and 20000; /*Q1*/
select * from t force index(a) where a between 10000 and 20000;/*Q2*/
```

第二种方法就是，可以考虑修改语句，引导 MySQL 使用我们期望的索引。
第三种方法是，在有些场景下，我们可以新建一个更合适的索引，来提供给优化器做选择，或删掉误用的索引。


## Using filesort 文件排序原理详解


文件排序方式：

- 单路排序：是一次性取出（聚簇索引）满足条件行的所有字段，然后在 sort buffer 中进行排序；trace 工具可以看到 sort_mode 信息里显示 `<sort_key, additional_fields>` 或者 `<sort_key,packed_additional_fields>`，sort_key 就表示排序的 key，additional_fields 表示表中的其他字段。
- 双路排序（又叫**回表排序模式**）：是首先根据相应的条件取出相应的排序字段和可以直接定位行数据的主键 ID，然后在 sort buffer 中进行排序，排序完后需要再次取回其它需要的字段；trace 工具可以看到 sort_mode 信息里显示`<sort_key, rowid>`，sort_key 就表示排序的 key，rowid 表示主键 ID。

判断使用哪种排序模式：

- 如果字段的总长度小于 `max_length_for_sort_data` ，那么使用单路排序模式；
- 如果字段的总长度大于 `max_length_for_sort_data` ，那么使用双路排序模式。


如果 MySQL 排序内存 sort_buffer 配置的比较小并且没有条件继续增加了，可以适当把 `max_length_for_sort_data` 配置小点，让优化器选择使用双路排序算法，可以在 sort_buffer 中一次排序更多的行，只是需要再根据主键回到原表取数据。

如果 MySQL 排序内存有条件可以配置比较大，可以适当增大 `max_length_for_sort_data` 的值，让优化器优先选择全字段排序(单路排序)，把需要的字段放到 sort_buffer 中，这样排序后就会直接从内存里返回查
询结果了。

所以，MySQL 通过 `max_length_for_sort_data` 这个参数来控制排序，在不同场景使用不同的排序模式，从而提升排序效率。

> 注意，如果全部使用 sort_buffer 内存排序一般情况下效率会高于磁盘文件排序，但不能因为这个就随便增大 sort_buffer(默认 1M)，mysql 很多参数设置都是做过优化的，不要轻易调整。

在磁盘中排序，最终还是要加载到内存中进行排序的，只不过由于数据量太大，需要先创建临时文件，然后在一块更大的内存中再加载临时文件进行排序，不会在 sort_buffer 中进行排序了。

## 查询优化

```sql
select * from employees limit 10000,10;
```

从表 employees 中取出从 10001 行开始的 10 行记录。看似只查询了 10 条记录，实际这条 SQL 是先读取 10010 条记录，然后抛弃前 10000 条记录，然后返回后面 10 条想要的数据。因此要**查询一张大表比较靠后的数据，执行效率是非常低的**。


### 常见的分页场景优化技巧

1. 根据**自增且连续**的主键排序的分页查询

```sql
select * from employees limit 90000,5;

-- 优化为
select * from employees where id > 90000 limit 5;
```

但是，这条改写的 SQL 在很多场景并不实用，因为表中可能某些记录被删后，主键空缺，导致结果不一致。

2. 根据非主键字段排序的分页查询

```sql
-- 联合索引 (name,age,position)
-- 该 sql 没有使用 name 字段的索引，因为查找联合索引的结果集太大，并回表的成本比扫描全表的成本更高，所以优化器放弃使用索引。
select * from employees ORDER BY name limit 90000,5;

-- 关键是让排序时返回的字段尽可能少，所以可以让排序和分页操作先查出主键，然后根据主键查到对应的记录，SQL 改写如下
-- 优化为
select * from employees e inner join (select id from employees order by name limit 90000,5) ed on e.id = ed.id;
```

优化后的语句全部都走了索引，其中 `(select id from employees order by name limit 90000,5)` 使用了覆盖索引来优化，查询的字段只有 id 字段，而且排好了序。`(select id from employees order by name limit 90000,5) ed` 产生的临时表只有 5 条记录，然后再根据主键 id 去 employees 表中查询对应的记录。

#### JOIN 关联查询优化

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

mysql 的表关联常见有两种算法：
- Nested-Loop Join 算法
- Block Nested-Loop Join 算法

**1. Nested-Loop Join 算法**：

一次一行循环地从第一张表（称为**驱动表**）中读取行，在这行数据中取到关联字段，根据关联字段在另一张表（**被驱动表**）里取出满足条件的行，然后取出两张表的结果合集。

```sql
EXPLAIN select * from t1 inner join t2 on t1.a= t2.a;
```

![]()

从执行计划中可以看到：

- 驱动表是 t2，被驱动表是 t1。先执行的就是驱动表(执行计划结果的 id 如果一样则按从上到下顺序执行 sql)；**优化器一般会优先选择小表做驱动表。所以使用 inner join 时，排在前面的表并不一定就是驱动表**。
- 如果执行计划 Extra 中未出现 Using join buffer 则表示使用的 join 算法是 NLJ。

sql 的大致流程如下：
1. 从表 t2 中读取一行数据（如果 t2 表有查询过滤条件的，会从过滤结果里取出一行数据）；
2. 从第 1 步的数据中，取出关联字段 a，到表 t1 中查找；
3. 取出表 t1 中满足条件的行，跟 t2 中获取到的结果合并，作为结果返回给客户端；
4. 重复上面 3 步。

整个过程会读取 t2 表的所有数据(扫描 100 行)，然后遍历这每行数据中字段 a 的值，根据 t2 表中 a 的值索引扫描 t1 表中的对应行(扫描 100 次 t1 表的索引，1 次扫描可以认为最终只扫描 t1 表一行完整数据，也就是总共 t1 表也扫描了 100 行)。因此整个过程扫描了 200 行。

**如果被驱动表的关联字段没索引，使用 NLJ 算法性能会比较低，mysql会选择 Block Nested-Loop Join 算法**。

**2. Block Nested-Loop Join 算法**：


**把驱动表的数据读入到 join_buffer 中**，然后扫描被驱动表，把被驱动表每一行取出来跟 **join_buffer 中的所有数据**一起做对比。

```sql
EXPLAIN select * from t1 inner join t2 on t1.b = t2.b;
```

![]()

Extra 中 的 Using join buffer (Block Nested Loop)说明该关联查询使用的是 BNL 算法。

sql 的大致流程如下：
1. 把 t2 的所有数据放入到 join_buffer 中
2. 把表 t1 中每一行取出来，跟 join_buffer 中的所有数据做对比
3. 返回满足 join 条件的数据


整个过程对表 t1 和 t2 都做了一次全表扫描，因此扫描的总行数为 `10000 (表 t1 的数据总量) + 100 (表 t2 的数据总量) = 10100`。并且 join_buffer 里的数据是无序的，因此对表 t1 中的每一行，都要做 100 次判断，所以内存中的判断次数是 `100 * 10000= 100 万次`。

join_buffer 的大小是由参数 `join_buffer_size` 设定的，默认值是 `256k`。如果放不下表 t2 的所有数据话，策略很简单，就是**分段放**。

比如 t2 表有 1000 行记录， join_buffer 一次只能放 800 行数据，那么执行过程就是先往 join_buffer 里放 800 行记录，然后从 t1 表里取数据跟 join_buffer 中数据对比得到部分结果，然后清空 join_buffer，再放入 t2 表剩余 200 行记录，再次从 t1 表里取数据跟 join_buffer 中数据对比。所以就多扫了一次 t1 表。

**被驱动表的关联字段没索引为什么要选择使用 BNL 算法而不使用 Nested-Loop Join 呢**？

如果上面第二条 sql 使用 Nested-Loop Join，那么扫描行数为 `100 * 10000 = 100 万行`，由于没有 t1.b 是没有索引的，意味这要进行**全表扫描**，这个要在磁盘中扫描 t1 表的所有行。很显然，用 BNL 磁盘扫描次数少很多，相比于磁盘扫描，BNL 的内存计算会快得多。


对于关联 sql 的优化
- **关联字段加索引**，让 mysql 做 join 操作时尽量选择 NLJ 算法
- **小表驱动大表**，写多表连接 sql 时如果明确知道哪张表是小表可以用 straight_join 写法固定连接驱动方式，省去 mysql 优化器自己判断的时间。


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



#####  Hash Join 原理（仅支持等值连接）

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



#### count(*) 查询优化


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

`count(1)` 跟 `count(字段)`执行过程类似，不过 `count(1)` 不需要取出字段统计，就用常量 1 做统计，`count(字段)` 还需要取出字段，所以理论上 `count(1)` 比 `count(字段)` 会快一点。

为什么对于 `count(id)`，mysql 最终选择辅助索引而不是主键聚集索引？因为二级索引相对主键索引存储数据更少，检索性能应该更高。


不带 where 条件的常见优化方法：

1. 对于 myisam 存储引擎的表做不带 where 条件的 count 查询性能是很高的，因为 myisam 存储引擎的表的总行数会被 mysql 存储在磁盘上，查询不需要计算。
2. `show table status` 可以看到表的行数，但是这个行数是不准确的。性能很高。例如 `show table status like 'employees'`。
3. 将总数维护到 Redis 里，插入或删除表数据行的时候同时维护 redis 里的表总行数 key 的计数值(用 incr 或 decr 命令)，但是这种方式可能不准，很难保证表操作和redis操作的**事务一致性**
4. 增加数据库计数表，插入或删除表数据行的时候同时维护计数表，让他们在同一个事务里操作