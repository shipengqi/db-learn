---
title: Group By
weight: 11
---

准备数据：

```sql
CREATE TABLE student_score (
  number INT(11) NOT NULL,
  name VARCHAR(30) NOT NULL,
  subject VARCHAR(30) NOT NULL,
  score TINYINT(4) DEFAULT NULL,
  PRIMARY KEY (number,subject)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- 插入一些数据后，表中数据如下：
mysql> SELECT * FROM student_score;
+----------+-----------+-----------------------------+-------+
| number   | name      | subject                     | score |
+----------+-----------+-----------------------------+-------+
| 20180101 | 杜子腾    | 母猪的产后护理              |    78 |
| 20180101 | 杜子腾    | 论萨达姆的战争准备          |    88 |
| 20180102 | 杜琦燕    | 母猪的产后护理              |   100 |
| 20180102 | 杜琦燕    | 论萨达姆的战争准备          |    98 |
| 20180103 | 范统      | 母猪的产后护理              |    59 |
| 20180103 | 范统      | 论萨达姆的战争准备          |    61 |
| 20180104 | 史珍香    | 母猪的产后护理              |    55 |
| 20180104 | 史珍香    | 论萨达姆的战争准备          |    46 |
+----------+-----------+-----------------------------+-------+
8 rows in set (0.00 sec)
```

## GROUP BY 是在干什么？

MySQL 中提供了一系列的聚集函数，例如：

- `COUNT`：统计记录数。
- `MAX`：查询某列的最大值。
- `MIN`：查询某列的最小值。
- `SUM`：某列数据的累加总和。
- `AVG`：某列数据的平均数。

如果想查看一下 `student_score` 表中所有人的平均成绩就可以这么写：

```bash
mysql> SELECT AVG(score) FROM student_score;
+------------+
| AVG(score) |
+------------+
|    73.1250 |
+------------+
1 row in set (0.00 sec)
```

如果只想查看 `《母猪的产后护理》` 这个科目的平均成绩，那加个 `WHERE` 子句就好了：

```bash
mysql> SELECT AVG(score) FROM student_score WHERE subject = '母猪的产后护理';
+------------+
| AVG(score) |
+------------+
|    73.0000 |
+------------+
1 row in set (0.00 sec)
```

那么如果这个 `student_score` 表中存储了 20 门科目的成绩信息，那如何得到这 20 门课程的平均成绩呢？单独写 20 个查询语句？

可以按照某个列将表中的数据进行分组，比方说现在按照 `subject` 列对表中数据进行分组，那么所有的记录就会被分成 2 组：

```bash
# 母猪的产后护理 组
+----------+-----------+-----------------------------+-------+
| number   | name      | subject                     | score |
+----------+-----------+-----------------------------+-------+
| 20180101 | 杜子腾    | 母猪的产后护理              |    78 |
| 20180102 | 杜琦燕    | 母猪的产后护理              |   100 |
| 20180103 | 范统      | 母猪的产后护理              |    59 |
| 20180104 | 史珍香    | 母猪的产后护理              |    55 |
+----------+-----------+-----------------------------+-------+

# 论萨达姆的战争准备 组
+----------+-----------+-----------------------------+-------+
| number   | name      | subject                     | score |
+----------+-----------+-----------------------------+-------+
| 20180101 | 杜子腾    | 论萨达姆的战争准备          |    88 |
| 20180102 | 杜琦燕    | 论萨达姆的战争准备          |    98 |
| 20180103 | 范统      | 论萨达姆的战争准备          |    61 |
| 20180104 | 史珍香    | 论萨达姆的战争准备          |    46 |
+----------+-----------+-----------------------------+-------+
```

分组的子句就是 `GROUP BY`：

```bash
mysql> SELECT subject, AVG(score) FROM student_score GROUP BY subject;
+-----------------------------+------------+
| subject                     | AVG(score) |
+-----------------------------+------------+
| 母猪的产后护理              |    73.0000 |
| 论萨达姆的战争准备          |    73.2500 |
+-----------------------------+------------+
2 rows in set (0.00 sec)
```

## 报错

```bash
mysql> SELECT subject, name, AVG(score) FROM student_score GROUP BY subject;

ERROR 1055 (42000): Expression #2 of SELECT list is not in GROUP BY clause and contains nonaggregated column 'test.student_score.name' which is not functionally dependent on columns in GROUP BY clause; this is incompatible with sql_mode=only_full_group_by

mysql>
```

为什么会报错？**使用 `GROUP BY` 子句是想把记录分为若干组，然后再对各个组分别调用聚集函数去做一些统计工作**。

本例 SQL 的查询列表中有一个 `name` 列，**既不是分组的列，也不是聚集函数的列**。

那么从各个分组中的记录中取一个记录的 `name` 列？该取哪条记录为好呢？比方说对于`'母猪的产后护理'`这个分组中的记录来说，`name` 列的值应该取 `杜子腾`，还是`杜琦燕`，还是`范统`，还是 `史珍香` 呢？这个不知道，所以把非分组列放到查询列表中会引起争议，导致结果不确定，所以 MySQL 才会报错。

如果一定要把非分组列也放到查询列表中，可以修改 `sql_mode` 的系统变量：

```bash
mysql> SHOW VARIABLES LIKE 'sql_mode';
+---------------+-------------------------------------------------------------------------------------------------------------------------------------------+
| Variable_name | Value                                                                                                                                     |
+---------------+-------------------------------------------------------------------------------------------------------------------------------------------+
| sql_mode      | ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_AUTO_CREATE_USER,NO_ENGINE_SUBSTITUTION |
+---------------+-------------------------------------------------------------------------------------------------------------------------------------------+
1 row in set (0.02 sec)
```

其中一个称之为 `ONLY_FULL_GROUP_BY` 的值，把这个值从 `sql_mode` 系统变量中移除，就不会报错了：

```bash
mysql> set sql_mode='STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_AUTO_CREATE_USER,NO_ENGINE_SUBSTITUTION';

Query OK, 0 rows affected (0.00 sec)
```