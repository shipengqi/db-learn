---
title: 写操作
weight: 5
---

## 插入数据

### 插入完整的行

```sql
insert into customers values('xiaoming', 'shanghai', 18);

insert into customers(cust_name, cust_address, cust_age) values('xiaoming', 'shanghai', 18);
```

**不建议使用第一种语法**，因为各个列必须以它们在表定义中出现的次序填充。高度依赖于表中列的定义次序。

> **如果表的定义允许，则可以在 `insert` 操作中省略某些列。省略的列必须满足以下某个条件。该列定义为允许 `NULL` 值（无值或空值）。在表定义中给出默认值。这表示如果不给出值，将使用默认值**。
> **如果不提供列名，`values` 则必须给每个表列提供一个值。如果提供列名，`values` 则必须对每个列出的列给出一个值**。

### 插入多行

```sql
insert into customers(cust_name, cust_address, cust_age) values('xiaoming', 'shanghai', 18), values('xiaoliang', 'shanghai', 18);
```

其中单条 `insert` 语句有多组值，每组值用一对圆括号括起来，用 `,` 分隔。

> MySQL 用**单条 `insert` 语句处理多个插入比使用多条 `insert` 语句快**。

### 插入检索出的数据

`INSERT` 还存在另一种形式，可以利用它将一条 `select` 语句的结果插入表中。

```sql
insert into customers(cust_name, cust_address, cust_age) select cust_name, cust_address, cust_age from custnew;
```

`select` 语句从 `custnew` 检索出要插入的值。

> `insert` 和 `select` 语句中使用了相同的列名。但是，**不一定要求列名匹配**。事实上，MySQL 使用的是列的位置，因此 `select` 中的第一列（不管其列名）将用来填充表列中指定的第一个列

## 更新数据

使用 `update` 语句更新（修改）表中的数据。

`update` 语句由 3 部分组成，分别是：

- 要更新的表
- 列名和它们的新值
- 确定要更新行的过滤条件

```sql
update customers set cust_email = '111111@demo.com' where cust_id = 1005;
```

> **`update` 一定要注意添加过滤条件，避免更新所有行**。
> `udpate` 语句更新多行，并且在更新这些行中的一行或多行时出一个现错误，则整个 `UPDATE` 操作被取消。为即使是发生错误，也继续进行更新，可使用 **`ignore` 关键字**，如：`update ignore customers`。

## 删除数据

```sql
delete from customers where cust_id = 1005;
```

`delete` 不需要列名或通配符。

> **`delete` 一定要注意添加过滤条件，避免删除所有行**。
> **删除表中所有行使用 `truncate table`**，它的速度比使用 `delete` 快很多（**`truncate` 实际是删除原来的表并重新创建一个表，而不是逐行删除表中的数据**）。

**对 `update` 或 `delete` 语句应该先用 `select` 进行测试 `where` 过滤条件，保证它过滤的是正确的记录**。

## 创建和操作表

### 创建

`create table` 创建表，必须给出下列信息：

- 新表的名字，在关键字 `create table` 之后给出；
- 表列的名字和定义，用 `,` 分隔。

```sql
create table customers
(
  cust_id int not null auto_increment,
  cust_name char(50) not null,
  cust_address char(50) not null default 'shanghai',
  primary key (cust_id)
) engine=InnoDB;
```

#### NULL 值

每个表列或者是 `NULL` 列，或者是 `NOT NULL` 列。

**数据库开发人员应该使用默认值而不是使用 `NULL` 列**。

#### 主键

**主键值必须唯一。如果主键使用单个列，则它的值必须唯一。如果使用多个列，则这些列的组合值必须唯一**。

#### auto_increment

`auto_increment` 自增。每次执行一个 `insert` 操作时，给该列值自动加 1。

**每个表只允许一个 `auto_increment` 列，而且它必须被索引**。

`last_insert_id()`，函数返回最后一个 `auto_increment` 值。

#### 设置表的存储引擎

**可以为不同的表设置不同的存储引擎**。不指定表的存储引擎，会使用默认的存储引擎 InnoDB。

修改表的存储引擎：

```sql
alter table 表名 engine = 存储引擎名称;
```

### 更新表

**理想状态下，当表中存储数据以后，该表就不应该再被更新**。

> 使用 `ALTER TABLE` 要极为小心，**应该在进行改动前做一个完整的备份（模式和数据的备份）**。

### 删除表

```sql
drop table customers
```

### 重命名表

```sql
rename table customers to customers2
```
