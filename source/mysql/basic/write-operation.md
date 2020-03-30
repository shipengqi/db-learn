---
title: 写操作
---
# 写操作

## 插入数据
### 插入完整的行
```sql
insert into customers values('xiaoming', 'shanghai', 18);

insert into customers(cust_name, cust_address, cust_age) values('xiaoming', 'shanghai', 18);
```
第二条语句更安全。第一种语法不建议使用，因为各个列必须以它们在表定义中出现的次序填充。高度依赖于表中列的定义次序，并且还依赖于其次序容
易获得的信息。

> **如果表的定义允许，则可以在 `INSERT` 操作中省略某些列。省略的列必须满足以下某个条件。该列定义为允许 NULL 值（无值或空值）。在表定义
中给出默认值。这表示如果不给出值，将使用默认值**。

> **不管使用哪种 `INSERT` 语法，都必须给出 `VALUES` 的正确数目。如果不提供列名，则必须给每个表列提供一个值。如果提供列名，则必须对
每个列出的列给出一个值**。

### 插入多行
```sql
insert into customers(cust_name, cust_address, cust_age) values('xiaoming', 'shanghai', 18), values('xiaoliang', 'shanghai', 18);
```
其中单条 `INSERT` 语句有多组值，每组值用一对圆括号括起来，用逗号分隔。

> MySQL 用单条 `INSERT` 语句处理多个插入比使用多条 `INSERT` 语句快。

### 插入检索出的数据
`INSERT` 还存在另一种形式，可以利用它将一条 `SELECT` 语句的结果插入表中。
```sql
insert into customers(cust_name, cust_address, cust_age) select cust_name, cust_address, cust_age from custnew;
```
使用 `INSERT SELECT` 从 `custnew` 中将所有数据导入 `customers`。`SELECT` 语句从 `custnew` 检索出要插入的值，而不是列出它们。

> `INSERT` 和 `SELECT` 语句中使用了相同的列名。但是，不一定要求列名匹配。事实上，MySQL 甚至不关心 `SELECT` 返回的列名。
它使用的是列的位置，因此 `SELECT` 中的第一列（不管其列名）将用来填充表列中指定的第一个列

## 更新数据
使用 `UPDATE` 语句更新（修改）表中的数据。

`UPDATE` 语句由 3 部分组成，分别是：
- 要更新的表；
- 列名和它们的新值
- 确定要更新行的过滤条件

```sql
update customers set cust_email = '111111@demo.com' where cust_id = 1005;
```

> **使用 `UPDATE `时一定要注意添加过滤条件，避免更新所有行**。

> `UPDATE` 语句更新多行，并且在更新这些行中的一行或多行时出一个现错误，则整个 `UPDATE` 操作被取消。为即使是发生错误，
也继续进行更新，可使用 **`IGNORE` 关键字**，如：`UPDATE IGNORE customers`。

## 删除数据

```sql
delete from customers where cust_id = 1005;
```

`DELETE` 不需要列名或通配符。

> **使用 `DELETE` 时一定要注意添加过滤条件，避免删除所有行**。

> **删除表中所有行使用 `TRUNCATE TABLE` 语句**，它的速度比使用 `delete` 快很多（`TRUNCATE` 实际是删除原来的表并重新创建一个表，而
不是逐行删除表中的数据）。


**对 `UPDATE` 或 `DELETE` 语句使用 `WHERE` 子句前，应该先用 `SELECT` 进行测试，保证它过滤的是正确的记录，以防编写的 `WHERE` 子
句不正确**。

## 创建和操作表
### 创建
`CREATE TABLE` 创建表，必须给出下列信息：
- 新表的名字，在关键字 `CREATE TABLE` 之后给出；
- 表列的名字和定义，用逗号分隔。

```sql
create table customers
(
  cust_id int NOT NULL AUTO_INCREMENT,
  cust_name char(50) NOT_NULL,
  cust_address char(50) NOT NULL DEFAULT 'shanghai',
  primary key (cust_id)
) engine=InnoDB;
```

#### NULL值
允许 `NULL` 值的列也允许在插入行时不给出该列的值。每个表列或者是 `NULL` 列，或者是 `NOT NULL` 列，这种状态在创建时由表的定义规定。

**数据库开发人员应该使用默认值而不是 `NULL` 列**。

#### 主键
**主键值必须唯一。即，表中的每个行必须具有唯一的主键值。如果主键使用单个列，则它的值必须唯一。如果使用多个列，则这些列的组合值必
须唯一**。

#### AUTO_INCREMENT
`AUTO_INCREMENT` 告诉 MySQL，本列每当增加一行时自动增量。每次执行一个 `INSERT` 操作时，MySQL 自动对该列增量，给该列赋予下一个可
用的值。

**每个表只允许一个 `AUTO_INCREMENT` 列，而且它必须被索引**。

`last_insert_id()`，函数返回最后一个 `AUTO_INCREMENT` 值。

### 更新表
**理想状态下，当表中存储数据以后，该表就不应该再被更新**。

使用 `ALTER TABLE` 更改表结构，必须给出下面的信息：
- 在 `ALTER TABLE` 之后给出要更改的表名（该表必须存在，否则将出错）；
- 所做更改的列表

> 使用 `ALTER TABLE` 要极为小心，**应该在进行改动前做一个完整的备份（模式和数据的备份）**。数据库表的更改不能撤销，如果增加了不需
要的列，可能不能删除它们。类似地，如果删除了不应该删除的列，可能会丢失该列中的所有数据。

### 删除表
```sql
drop table customers
```

### 重命名表
```sql
rename table customers to customers2
```