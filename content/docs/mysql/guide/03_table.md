---
title: 表操作
weight: 3
---

## 查看表

```bash
mysql> SHOW TABLES;
Empty set (0.00 sec)
```

### 查看表结构

下面的语句效果是一样的，可以用来查看定义的表的结构：

```bash
DESCRIBE 表名;
DESC 表名;
EXPLAIN 表名;
SHOW COLUMNS FROM 表名;
SHOW FIELDS FROM 表名;
```

示例：

```bash
mysql> DESC student_info;
+-----------------+-------------------+------+-----+---------+-------+
| Field           | Type              | Null | Key | Default | Extra |
+-----------------+-------------------+------+-----+---------+-------+
| number          | int(11)           | YES  |     | NULL    |       |
| name            | varchar(5)        | YES  |     | NULL    |       |
| sex             | enum('男','女')   | YES  |     | NULL    |       |
| id_number       | char(18)          | YES  |     | NULL    |       |
| department      | varchar(30)       | YES  |     | NULL    |       |
| major           | varchar(30)       | YES  |     | NULL    |       |
| enrollment_time | date              | YES  |     | NULL    |       |
+-----------------+-------------------+------+-----+---------+-------+
7 rows in set (0.00 sec)
```

### 查看表的创建语句

```bash
SHOW CREATE TABLE 表名;
```

示例：

```bash
mysql> SHOW CREATE TABLE student_info;
+--------------+-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| Table        | Create Table                                                                                                                                                                                                                                                                                                                                                          |
+--------------+-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| student_info | CREATE TABLE `student_info` (
  `number` int(11) DEFAULT NULL,
  `name` varchar(5) DEFAULT NULL,
  `sex` enum('男','女') DEFAULT NULL,
  `id_number` char(18) DEFAULT NULL,
  `department` varchar(30) DEFAULT NULL,
  `major` varchar(30) DEFAULT NULL,
  `enrollment_time` date DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='学生基本信息表'          |
+--------------+-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
1 row in set (0.00 sec)
```

{{< callout type="info" >}}
如果输出数据太长，显示效果不好，可以把标记语句结束的**分号 `;` 改为 `\G`**，以垂直的方式展示每一列数据。
{{< /callout >}}

## 创建

基本语法：

```sql
CREATE TABLE 表名 (
    列名1    数据类型    [列的属性],
    列名2    数据类型    [列的属性],
    ...
    列名n    数据类型    [列的属性]
);
```

`列的属性` 用中括号 `[]` 引起来表示是可选的。

示例：

```sql
CREATE TABLE first_table (
    first_column INT,
    second_column VARCHAR(100)
);
```

### 添加表注释

```sql
CREATE TABLE 表名 (
    各个列的信息 ...
) COMMENT '表的注释信息';
```

### IF NOT EXISTS

和重复创建数据库一样，创建表时也可以使用 `IF NOT EXISTS` 关键字，这样如果表不存在，则创建表；如果表已经存在，则不执行创建表的操作。

```sql
CREATE TABLE IF NOT EXISTS first_table (
    first_column INT,
    second_column VARCHAR(100)
);
```

## 删除

基本语法：

```sql
DROP TABLE 表1, 表2, ..., 表n;
```

示例：

```sql
DROP TABLE first_table;
```

### IF EXISTS

和重复删除数据库一样，删除表时也可以使用 `IF EXISTS` 关键字，这样如果表存在，则删除表；如果表不存在，则不执行删除表的操作。

```sql
DROP TABLE IF EXISTS first_table;
```

## 操作其他库的表

例如当前所在的库是 `test`，如果不想使用 `use` 语句切换到其他库，**就必须显式的指定这些表所属的数据库（数据库名.表名）**。

```bash
mysql> SHOW TABLES FROM demo;
+---------------------+
| Tables_in_demo      |
+---------------------+
| first_table         |
| student_info        |
| student_score       |
+---------------------+
3 rows in set (0.00 sec)

mysql> SHOW CREATE TABLE test.first_table\G
```

## 修改

### 修改表名

```sql
ALTER TABLE 旧表名 RENAME TO 新表名;

-- 或者
RENAME TABLE 旧表名1 TO 新表名1, 旧表名2 TO 新表名2, ... 旧表名n TO 新表名n;
```

如果在修改表名的时候指定了数据库名，还可以将该表转移到对应的数据库下：

```sql
ALTER TABLE test.first_table RENAME TO demo.first_table;
```


### 增加列

基本语法：

```sql
ALTER TABLE 表名 ADD COLUMN 列名 数据类型 [列的属性];
```

示例：

```bash
mysql> ALTER TABLE first_table ADD COLUMN third_column CHAR(4) ;
Query OK, 0 rows affected (0.05 sec)
Records: 0  Duplicates: 0  Warnings: 0

mysql> SHOW CREATE TABLE first_table\G
*************************** 1. row ***************************
       Table: first_table
Create Table: CREATE TABLE `first_table` (
  `first_column` int(11) DEFAULT NULL,
  `second_column` varchar(100) DEFAULT NULL,
  `third_column` char(4) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='第一个表'
1 row in set (0.01 sec)
```

### 增加列到特定位置

默认的情况下列都是加到现有列的最后一列后面，但是也可以在添加列的时候指定它的位置：

```sql
-- 添加到第一列
ALTER TABLE 表名 ADD COLUMN 列名 列的类型 [列的属性] FIRST;

-- 添加到指定列的后边
ALTER TABLE 表名 ADD COLUMN 列名 列的类型 [列的属性] AFTER 指定列名;
```

### 删除列

```sql
ALTER TABLE 表名 DROP COLUMN 列名;
```

### 修改列

```sql
ALTER TABLE 表名 MODIFY 列名 新数据类型 [新属性];

-- 或者
-- 这种方式可以再修改数据类型和属性的同时，也可以修改列名。
ALTER TABLE 表名 CHANGE 旧列名 新列名 新数据类型 [新属性];
```

示例：

```bash
mysql> ALTER TABLE first_table MODIFY second_column VARCHAR(2);
Query OK, 0 rows affected (0.04 sec)
Records: 0  Duplicates: 0  Warnings: 0

mysql> SHOW CREATE TABLE first_table\G
*************************** 1. row ***************************
       Table: first_table
Create Table: CREATE TABLE `first_table` (
  `first_column` int(11) DEFAULT NULL,
  `second_column` varchar(2) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='第一个表'
1 row in set (0.00 sec)    

mysql> ALTER TABLE first_table CHANGE second_column second_column1 VARCHAR(2)\G
Query OK, 0 rows affected (0.04 sec)
Records: 0  Duplicates: 0  Warnings: 0

mysql> SHOW CREATE TABLE first_table\G
*************************** 1. row ***************************
       Table: first_table
Create Table: CREATE TABLE `first_table` (
  `first_column` int(11) DEFAULT NULL,
  `second_column1` varchar(2) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='第一个表'
1 row in set (0.00 sec)
```

### 修改列顺序

```sql
-- 将列设为表的第一列
ALTER TABLE 表名 MODIFY 列名 列的类型 列的属性 FIRST;

-- 将列放到指定列的后边
ALTER TABLE 表名 MODIFY 列名 列的类型 列的属性 AFTER 指定列名;
```

### 同时执行多个操作

```sql
ALTER TABLE 表名 操作1, 操作2, ..., 操作n;
```

示例：

```sql
ALTER TABLE first_table DROP COLUMN third_column, DROP COLUMN fourth_column, DROP COLUMN fifth_column;
```

用一条语句删除了 `third_column`、`fourth_column` 和 `fifth_column` 这三个列。

### 修改表的存储引擎

```sql
ALTER TABLE 表名 ENGINE=新的存储引擎;
```