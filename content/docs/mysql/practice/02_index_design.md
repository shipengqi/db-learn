---
title: 索引设计
weight: 2
---


索引设计的核心思想就是**尽量利用一两个复杂的多字段联合索引，抗下 80% 以上的查询**，然后用一两个辅助索引尽量抗下剩余的一些非典型查询，保证这种大数据量表的查询尽可能多的都能充分利用索引，这样就能保证查询速度和性能了。

## 代码先行，索引后上

一般应该等到主体业务功能开发完毕，把涉及到该表相关 SQL 都要拿出来分析之后再建立索引。

## 联合索引尽量覆盖条件

可以设计一个或者两三个联合索引(尽量少建单值索引)，让每一个联合索引都尽量去包含 SQL 语句里的 `where`、`order by`、`group by` 的字段，还要确保这些联合索引的字段顺序尽量满足 SQL 查询的最左前缀原则。

联合索引中的某个字段如果是**范围查找，最好把这个字段放在联合索引的最后面**。因为**一般情况下，范围查找之后的字段就无法走索引了**。

示例：

```sql
-- 联合索引 (province,city,sex)
SELECT * FROM users WHERE province = xx AND city = xx AND age <= xx AND age >= xx;
```

上面的语句 `age` 是一个范围查找字段，所以最好把它放在联合索引的最后面，即 `(province,city,sex,age)`。但是由于上面的 SQL 并没有用到 `sex` 字段，会导致 `age` 无法走索引，所以可以优化为：

```sql
-- 加上 sex in ('female','male') 的条件
SELECT * FROM users WHERE province = xx AND city = xx AND sex in ('female','male') AND age <= xx AND age >= xx;
```

假设还有一个筛选条件，要筛选最近一周登录过的用户，对应后台 SQL 可能是这样：

```sql
where province=xx and city=xx and 
sex in ('female','male') and 
age>=xx and age<=xx and 
latest_login_time>= xx
```

`latest_login_time` 也是一个范围查找字段，如果把它放在联合索引里，如 `(province,city,sex,hobby,age,latest_login_time)`，`age` 和 `latest_login_time` 两个范围查找，显然是不行的。可以换一种思路，设计一个字段 `is_login_in_latest_7_days`，用户如果一周内有登录值就为 1，否则为 0，那么就可以把索引设计成 `(province,city,sex,hobby,is_login_in_latest_7_days,age)` 来满足上面那种场景。

## 只为用于搜索、排序或分组的列创建索引

## 不要在小基数字段上建立索引

索引基数是指这个字段在表里总共有多少个不同的值，比如一张表总共 100 万行记录，其中有个性别字段，其值不是‘男’就是‘女’，那么该字段的基数就是 2。对这种小基数字段建立索引的话，还不如全表扫描。因为索引树里就包含‘男’和‘女’两种值，根本没法进行快速的二分查找，那用索引就没有太大的意义了。

## 索引列的类型尽量小

尽量对字段类型较小的列设计索引，因为字段类型较小的话，占用磁盘空间也会比较小。

以整数类型为例，有 `TINYINT`、`MEDIUMINT`、`INT`、`BIGINT` 这么几种，它们占用的存储空间依次递增，我们这里所说的**类型大小指的就是该类型表示的数据范围的大小**。

在表示的整数范围允许的情况下，**尽量让索引列使用较小的类型**：

- 数据类型越小，在查询时进行的比较操作越快。
- 数据类型越小，**索引占用的存储空间就越少，在一个数据页内就可以放下更多的记录，从而减少磁盘 I/O 带来的性能损耗，也就意味着可以把更多的数据页缓存在内存中，从而加快读写效率**。

## 长字符串可以采用前缀索引

假设字符串很长，那存储一个字符串就需要占用很大的存储空间。

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

但是对于 `order by name`，那么此时 `name` 因为在索引树里仅仅包含了前 20 个字符，无法对后边的字符不同的记录进行排序，`group by` 也是同理。

## 主键插入顺序

据页和记录又是按照记录主键值从小到大的顺序进行排序，所以如果插入的记录的主键值是依次增大的话，那每插满一个数据页就换到下一个数据页继续插，而如果插入的主键值忽大忽小的话，可能需要**页面分裂和记录移位**。意味着：**性能损耗**。

最好让插入的记录的主键值依次递增，这样就不会发生这样的性能损耗了。所以建议：**让主键具有 AUTO_INCREMENT，让存储引擎自己为表生成主键**，而不是手动插入。

## 优先 where
在 `where` 和 `order by` 出现索引设计冲突时，到底是针对 `where` 去设计索引，还是针对 `order by` 设计索引？

一般是让 `where` 条件去使用索引来快速筛选出来一部分指定的数据，接着再进行排序。**因为大多数情况基于索引进行 `where` 筛选往往可以最快速度筛选出你要的少部分数据，然后做排序的成本可能会小很多**。

## 基于慢 SQL 查询做优化

可以根据监控后台的一些慢 SQL，针对这些慢 SQL 查询做特定的索引优化。参考 [SQL 慢查询](https://note.youdao.com/ynoteshare/index.html?id=c71f1e66b7f91dab989a9d3a7c8ceb8e&type=note&_time=1747366858606)。