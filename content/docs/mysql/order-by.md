---
title: order by
---

```sql
select city,name,age from t where city='杭州' order by name limit 1000  ;
```

## 全字段排序

MySQL会给每个线程分配一块内存用于排序，称为 sort_buffer。在MySQL中，把这种在内存中或者磁盘上进行排序的方式统称为**文件排序**（英文名：filesort）

这个语句执行流程如下所示 ：

1. 初始化sort_buffer，确定放入name、city、age这三个字段；
2. 从索引city找到第一个满足city='杭州’条件的主键id，也就是图中的ID_X；
3. 到主键id索引取出整行，取name、city、age三个字段的值，存入sort_buffer中；
4. 从索引city取下一个记录的主键id；
5. 重复步骤3、4直到city的值不满足查询条件为止，对应的主键id也就是图中的ID_Y；
6. 对sort_buffer中的数据按照字段name做快速排序；

按照排序结果取前1000行返回给客户端。

按name排序这个动作，**可能在内存中完成，也可能需要使用外部排序，这取决于排序所需的内存和参数sort_buffer_size**。

sort_buffer_size，就是MySQL为排序开辟的内存（sort_buffer）的大小。如果要排序的数据量小于sort_buffer_size，排序就在内存中完成。但如果排序数据量太大，内存放不下，则不得不利用磁盘临时文件辅助排序。

利用磁盘临时文件辅助排序时，**MySQL将需要排序的数据分成多份，每一份单独排序后存在这些临时文件中。然后把这12个有序文件再合并成一个有序的大文件**。

## rowid排序

如果查询要返回的字段很多的话，那么sort_buffer里面要放的字段数太多，这样内存里能够同时放下的行数很少，要分成很多个临时文件，排序的性能会很差。

如果MySQL认为排序的单行长度太大会怎么做呢？

接下来，我来修改一个参数，让MySQL采用另外一种算法。

```sql
SET max_length_for_sort_data = 16;
```

max_length_for_sort_data，是MySQL中专门控制用于排序的行数据的长度的一个参数。它的意思是，如果单行的长度超过这个值，MySQL就认为单行太大，要换一个算法。

city、name、age 这三个字段的定义总长度是36，我把max_length_for_sort_data设置为16，我们再来看看计算过程有什么改变。

**新的算法放入sort_buffer的字段，只有要排序的列（即name字段）和主键id**。

但这时，**排序的结果就因为少了city和age字段的值，不能直接返回了**，整个执行流程就变成如下所示的样子：

1. 初始化sort_buffer，确定放入两个字段，即name和id；
2. 从索引city找到第一个满足city='杭州’条件的主键id，也就是图中的ID_X；
3. 到主键id索引取出整行，取name、id这两个字段，存入sort_buffer中；
4. 从索引city取下一个记录的主键id；
5. 重复步骤3、4直到不满足city='杭州’条件为止，也就是图中的ID_Y；
6. 对sort_buffer中的数据按照字段name进行排序；

遍历排序结果，取前1000行，并按照id的值回到原表中取出city、name和age三个字段返回给客户端。

**如果内存够，就要多利用内存，尽量减少磁盘访问**。

并不是所有的order by语句，都需要排序操作的。从上面分析的执行过程，我们可以看到，MySQL之所以需要生成临时表，并且在临时表上做排序操作，其原因是原来的数据都是无序的。

## 不需要排序的情况

可以在这个市民表上创建一个city和name的联合索引，对应的SQL语句是：

```sql
alter table t add index city_user(city, name);
```

查询过程的流程就变成了：

1. 从索引(city,name)找到第一个满足city='杭州’条件的主键id；
2. 到主键id索引取出整行，取name、city、age三个字段的值，作为结果集的一部分直接返回；
3. 从索引(city,name)取下一个记录主键id；
4. 重复步骤2、3，直到查到第1000条记录，或者是不满足city='杭州’条件时循环结束。
