---
title: MySQL 复杂查询语句
---

## 子查询

SQL 还允许创建子查询（subquery），即嵌套在其他查询中的查询。

### 利用子查询过滤

订单存储在两个表中。对于包含订单号、客户 ID、订单日期的每个订单，`orders` 表存储一行。各订单的物品存储在相关的 `orderitems` 表中。
`orders` 表不存储客户信息。它只存储客户的 ID。实际的客户信息存储在 `customers` 表中。

假如需要列出订购物品 TNT2 的所有客户，需要下面几步：

1. 检索包含物品 TNT2 的所有订单的编号。
2. 检索具有前一步骤列出的订单编号的所有客户的 ID。
3. 检索前一步骤返回的所有客户ID的客户信息。

```sh
# 检索包含物品 TNT2 的所有订单的编号
mysql> select order_num from orderitems where pod_id= 'TNT2';
+-------------+
| order_num   |
+-------------+
|      2005   |
|      2007   |
+-------------+

# 查询具有订单 20005 和 20007 的客户 ID
mysql> select cust_id from orders where order_num in (2005,2007);
+-------------+
|   cust_id   |
+-------------+
|      1001   |
|      1004   |
+-------------+
```

把第一个查询（返回订单号的那一个）变为子查询组合两个查询:

```sh
mysql> select cust_id from orders where order_num in (select order_num from orderitems where pod_id= 'TNT2');
+-------------+
|   cust_id   |
+-------------+
|      1001   |
|      1004   |
+-------------+
```

**子查询总是从内向外处理**。

进一步根据客户 id 查出客户信息：

```sql
select cust_name from customers where cust_id in (
select cust_id from orders where order_num in (select order_num from orderitems where pod_id= 'TNT2'));
```

> 对于能嵌套的子查询的数目没有限制，但是不要嵌套太多的子查询，会影响性能。
> **在 `where` 子句中使用子查询，应该保证 `select` 语句具有与 `where` 子句中的列必须匹配**。

### 作为计算字段使用子查询

子查询的另一方法是创建计算字段。

假如需要显示 `customers` 表中每个客户的订单总数。订单与相应的客户 ID 存储在 `orders` 表中。

```sql
select cust_name, (select COUNT(*) from orders where orders.cust_id = customers.cust_id) as orders from customers order by cust_name;
```

`orders` 是一个计算字段，它是由圆括号中的子查询建立的。该子查询对检索出的每个客户执行一次。子查询中的 `where` 子句使
用了**完全限定列名**。

## 联结表

`join` 表。

### 关系表

**相同数据出现多次决不是一件好事，此因素是关系数据库设计的基础**。关系表的设计就是要保证把信息分解成多个表，一类数据一个表。**各表通过某些常用的值（即关系设计中的关系（relational））互相关联**。

### 创建联结

```sql
select vend_name, prod_name, prod_price from vendors, products where vendors.vend_id = products.vend_id order by vend_name, prod_name;
```

列 `prod_name` 和 `prod_price` 在一个表中，而列 `vend_name` 在另一个表中。`from` 子句列出了两个表，分别是 `vendors` 和 `products`。
这两个表用 `where` 子句联结。

> **笛卡儿积**（cartesian product）由**没有联结条件**的表关系返回的结果为笛卡儿积。检索出的行的数目将是第一个表中的行数乘以第二个表中
的行数。

#### 内部联结

目前为止所用的联结称为**等值联结（equijoin）**（也叫**内部联结**），使用 `join` 来表示：

```sql
select vend_name, prod_name, prod_price from vendors inner join products on vendors.vend_id = products.vend_id;
```

此语句中的 `from` 子句与上面的不同。两个表之间的关系是 `from` 子句的组成部分，以 **`INNER JOIN`** 指定。在使用这种语法时，联
结条件用特定的 **`on` 子句** 而不是 `where` 子句给出。传递给 `on` 的实际条件与传递给 `where` 的相同。

> **SQL 规范首选 `INNER JOIN` 语法**。此外，尽管使用 `where` 子句定义联结的确比较简单，但是使用明确的联结语法能够确保不会忘记联结条件，有时候这样做也能影响性能。不要联结不必要的表。联结的表越多，性能下降越厉害。

当两张表的数据量比较大，又需要连接查询时，应该使用 `FROM table1 JOIN table2 ON xxx` 的语法，避免使用 `FROM table1,table2 WHERE xxx` 的语法，因为后者会在内存中先生成一张数据量比较大的笛卡尔积表，增加了内存的开销。

#### 联结多个表

SQL 对一条 SELECT 语句中可以联结的表的数目没有限制。

使用前面子查询的例子：

```sql
select cust_name from customers where cust_id in (
select cust_id from orders where order_num in (select order_num from orderitems where pod_id = 'TNT2'));
```

可以改造为：

```sql
select cust_name from customers, orders, orderitems
where customers.cust_id = orders.cust_id
and orderitems.order_num = orders.order_num
and pod_id = 'TNT2';
```

### 高级联结

#### 表别名

使用表别名：

- 缩短 SQL 语句
- 允许在单条 `select` 语句中多次使用相同的表

**使用表别名的主要原因之一是能在单条 `select` 语句中不止一次引用相同的表**。

```sql
select cust_name from customers as c, orders as o, orderitems as oi
where c.cust_id = o.cust_id
and oi.order_num = o.order_num
and pod_id = 'TNT2';
```

#### 外部联结

```sql
select customers.cust_id, orders.order_num from customers inner join orders on customers.cust_id = orders.cust_id;
```

上面的语句很简单，就是检索所有客户及其订单。

那么如果想要检索所有客户，包括没有订单的客户，如下：

```sql
select customers.cust_id, orders.order_num from customers left outer join orders on customers.cust_id = orders.cust_id;
```

这条语句使用了关键字 `OUTER JOIN` 来指定联结的类型。外部联结还包括没有关联行的行。在使用 `OUTER JOIN` 语法时，必须使用 **`RIGHT`或 `LEFT` 关键字指定包括其所有行的表**（`RIGHT` 指的是 `OUTER JOIN` 右边的表，而 `LEFT` 指的是 `OUTER JOIN` 左边的表）。

上面的例子使用 `LEFT OUTER JOIN` 从 `from` 子句的左边表（`customers` 表）中选择所有行。

## 组合查询

MySQL 允许执行多个查询（多条 `select` 语句），并将结果作为单个查询结果集返回。这些组合查询通常称为**并**（`union`）或
**复合查询**（`compound query`）。

```sql
select vend_id, prod_id, prod_price from products where prod_price >= 5
union
select vend_id, prod_id, prod_price from products where vend_id in (1001,1002);
```

转成多条 `where` 子句的写法：

```sql
select vend_id, prod_id, prod_price from products where prod_price >= 5 or vend_id in (1001,1002);
```

> 实际中应该试一下这两种方式，以确定对特定的查询哪一种性能更好。

### union 规则

- `union` 必须由两条或两条以上的 `select` 语句组成，语句之间用关键字 `union` 分隔。
- `union` 中的每个查询**必须包含相同的列、表达式或聚集函数**（列的次序无所谓）。
- **列数据类型必须兼容**：类型不必完全相同，但必须是 DBMS 可以隐含地转换的类型（例如，不同的数值类型或不同的日期类型）。

### 包含或取消重复的行

`union` 从查询结果集中**自动去除了重复的行**。如果想**返回所有匹配行，使用 `union all`** 而不是 `union`。

### 组合查询结果排序

**用 `union` 组合查询时，只能使用一条 `order by` 子句，它必须出现在最后一条 `select` 语句之后**。对于结果集，不存在用一种方式排序一
部分，而又用另一种方式排序另一部分的情况，因此不允许使用多条 `order by` 子句。
