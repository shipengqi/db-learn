---
title: 查询优化
---

## 成本

MySQL执行一个查询可以有不同的执行方案，它会选择其中成本最低，或者说代价最低的那种方案去真正的执行查询。

### I/O成本

我们的表经常使用的MyISAM、InnoDB存储引擎都是将数据和索引都存储到磁盘上的，当我们想查询表中的记录时，需要先把数据或者索引加载到内存中然后再操作。这个从磁盘到内存这个加载的过程损耗的时间称之为I/O成本。

### CPU成本

读取以及检测记录是否满足对应的搜索条件、对结果集进行排序等这些操作损耗的时间称之为CPU成本。

对于InnoDB存储引擎来说，页是磁盘和内存之间交互的基本单位，MySQL规定读取一个页面花费的成本默认是 `1.0`，读取以及检测一条记录是否符合搜索条件的成本默认是 `0.2`。`1.0`、`0.2`这些数字称之为**成本常数**。

## 单表查询的成本

在一条单表查询语句真正执行之前，MySQL的查询优化器会找出执行该语句所有可能使用的方案，对比之后找出成本最低的方案，这个成本最低的方案就是所谓的**执行计划**，之后才会调用存储引擎提供的接口真正的执行查询，这个过程总结一下就是这样：

1. 根据搜索条件，找出所有可能使用的索引
2. 计算全表扫描的代价
3. 计算使用不同索引执行查询的代价
4. 对比各种执行方案的代价，找出成本最低的那一个

## 连接查询的成本

## 基于规则的优化

MySQL 依据一些规则，竭尽全力的把这个很糟糕的语句转换成某种可以比较高效执行的形式，这个过程也可以被称作**查询重写**

### 外连接消除

我们前边说过，内连接的驱动表和被驱动表的位置可以相互转换，而左（外）连接和右（外）连接的驱动表和被驱动表是固定的。这就导致内连接可能通过优化表的连接顺序来降低整体的查询成本，而外连接却无法优化表的连接顺序。为了故事的顺利发展，我们还是把之前介绍连接原理时用过的t1和t2表请出来，为了防止大家早就忘掉了，我们再看一下这两个表的结构：

CREATE TABLE t1 (
    m1 int,
    n1 char(1)
) Engine=InnoDB, CHARSET=utf8;

CREATE TABLE t2 (
    m2 int,
    n2 char(1)
) Engine=InnoDB, CHARSET=utf8;
为了唤醒大家的记忆，我们再把这两个表中的数据给展示一下：

mysql> SELECT * FROM t1;
+------+------+
| m1   | n1   |
+------+------+
|    1 | a    |
|    2 | b    |
|    3 | c    |
+------+------+
3 rows in set (0.00 sec)

mysql> SELECT * FROM t2;
+------+------+
| m2   | n2   |
+------+------+
|    2 | b    |
|    3 | c    |
|    4 | d    |
+------+------+
3 rows in set (0.00 sec)
我们之前说过，外连接和内连接的本质区别就是：对于外连接的驱动表的记录来说，如果无法在被驱动表中找到匹配ON子句中的过滤条件的记录，那么该记录仍然会被加入到结果集中，对应的被驱动表记录的各个字段使用NULL值填充；而内连接的驱动表的记录如果无法在被驱动表中找到匹配ON子句中的过滤条件的记录，那么该记录会被舍弃。查询效果就是这样：

mysql> SELECT * FROM t1 INNER JOIN t2 ON t1.m1 = t2.m2;
+------+------+------+------+
| m1   | n1   | m2   | n2   |
+------+------+------+------+
|    2 | b    |    2 | b    |
|    3 | c    |    3 | c    |
+------+------+------+------+
2 rows in set (0.00 sec)

mysql> SELECT * FROM t1 LEFT JOIN t2 ON t1.m1 = t2.m2;
+------+------+------+------+
| m1   | n1   | m2   | n2   |
+------+------+------+------+
|    2 | b    |    2 | b    |
|    3 | c    |    3 | c    |
|    1 | a    | NULL | NULL |
+------+------+------+------+
3 rows in set (0.00 sec)
对于上边例子中的（左）外连接来说，由于驱动表t1中m1=1, n1='a'的记录无法在被驱动表t2中找到符合ON子句条件t1.m1 = t2.m2的记录，所以就直接把这条记录加入到结果集，对应的t2表的m2和n2列的值都设置为NULL。

小贴士： 右（外）连接和左（外）连接其实只在驱动表的选取方式上是不同的，其余方面都是一样的，所以优化器会首先把右（外）连接查询转换成左（外）连接查询。我们后边就不再唠叨右（外）连接了。

我们知道WHERE子句的杀伤力比较大，凡是不符合WHERE子句中条件的记录都不会参与连接。只要我们在搜索条件中指定关于被驱动表相关列的值不为NULL，那么外连接中在被驱动表中找不到符合ON子句条件的驱动表记录也就被排除出最后的结果集了，也就是说：在这种情况下：外连接和内连接也就没有什么区别了！比方说这个查询：

mysql> SELECT * FROM t1 LEFT JOIN t2 ON t1.m1 = t2.m2 WHERE t2.n2 IS NOT NULL;
+------+------+------+------+
| m1   | n1   | m2   | n2   |
+------+------+------+------+
|    2 | b    |    2 | b    |
|    3 | c    |    3 | c    |
+------+------+------+------+
2 rows in set (0.01 sec)
由于指定了被驱动表t2的n2列不允许为NULL，所以上边的t1和t2表的左（外）连接查询和内连接查询是一样一样的。当然，我们也可以不用显式的指定被驱动表的某个列IS NOT NULL，只要隐含的有这个意思就行了，比方说这样：

mysql> SELECT * FROM t1 LEFT JOIN t2 ON t1.m1 = t2.m2 WHERE t2.m2 = 2;
+------+------+------+------+
| m1   | n1   | m2   | n2   |
+------+------+------+------+
|    2 | b    |    2 | b    |
+------+------+------+------+
1 row in set (0.00 sec)
在这个例子中，我们在WHERE子句中指定了被驱动表t2的m2列等于2，也就相当于间接的指定了m2列不为NULL值，所以上边的这个左（外）连接查询其实和下边这个内连接查询是等价的：

mysql> SELECT * FROM t1 INNER JOIN t2 ON t1.m1 = t2.m2 WHERE t2.m2 = 2;
+------+------+------+------+
| m1   | n1   | m2   | n2   |
+------+------+------+------+
|    2 | b    |    2 | b    |
+------+------+------+------+
1 row in set (0.00 sec)
我们把这种在外连接查询中，指定的WHERE子句中包含被驱动表中的列不为NULL值的条件称之为空值拒绝（英文名：reject-NULL）。在被驱动表的WHERE子句符合空值拒绝的条件后，外连接和内连接可以相互转换。这种转换带来的好处就是查询优化器可以通过评估表的不同连接顺序的成本，选出成本最低的那种连接顺序来执行查询。

## 子查询在MySQL中是怎么执行的

### 标量子查询、行子查询的执行方式

我们经常在下边两个场景中使用到标量子查询或者行子查询：

SELECT子句中，我们前边说过的在查询列表中的子查询必须是标量子查询。

子查询使用=、>、<、>=、<=、<>、!=、<=>等操作符和某个操作数组成一个布尔表达式，这样的子查询必须是标量子查询或者行子查询。

对于上述两种场景中的不相关标量子查询或者行子查询来说，它们的执行方式是简单的，比方说下边这个查询语句：

SELECT * FROM s1
    WHERE key1 = (SELECT common_field FROM s2 WHERE key3 = 'a' LIMIT 1);
它的执行方式和年少的我想的一样：

先单独执行(SELECT common_field FROM s2 WHERE key3 = 'a' LIMIT 1)这个子查询。

然后在将上一步子查询得到的结果当作外层查询的参数再执行外层查询SELECT * FROM s1 WHERE key1 = ...。

也就是说，对于**包含不相关的标量子查询或者行子查询的查询语句来说，MySQL会分别独立的执行外层查询和子查询，就当作两个单表查询就好了**。

对于相关的标量子查询或者行子查询来说，比如下边这个查询：

SELECT * FROM s1 WHERE
    key1 = (SELECT common_field FROM s2 WHERE s1.key3 = s2.key3 LIMIT 1);
事情也和年少的我想的一样，它的执行方式就是这样的：

先从外层查询中获取一条记录，本例中也就是先从s1表中获取一条记录。

然后从上一步骤中获取的那条记录中找出子查询中涉及到的值，本例中就是从s1表中获取的那条记录中找出s1.key3列的值，然后执行子查询。

最后根据子查询的查询结果来检测外层查询WHERE子句的条件是否成立，如果成立，就把外层查询的那条记录加入到结果集，否则就丢弃。

再次执行第一步，获取第二条外层查询中的记录，依次类推～


### IN子查询优化