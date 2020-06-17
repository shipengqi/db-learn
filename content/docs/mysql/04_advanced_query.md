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

**子查询总是从内向外处理**。在处理上面的 `SELECT` 语句时，MySQL **实际上执行了两个操作**。

进一步根据客户 id 查出客户信息：

```sql
select cust_name from customers where cust_id in (
select cust_id from orders where order_num in (select order_num from orderitems where pod_id= 'TNT2'));
```

> 对于能嵌套的子查询的数目没有限制，不过在实际使用时由于性能的限制，不能嵌套太多的子查询。
> **在 `WHERE` 子句中使用子查询，应该保证 `SELECT` 语句具有与 `WHERE` 子句中相同数目的列**。通常，子查询将返回单个列并且与
单个列匹配，但如果需要也可以使用多个列。

### 作为计算字段使用子查询

子查询的另一方法是创建计算字段。

假如需要显示 `customers` 表中每个客户的订单总数。订单与相应的客户ID存储在 `orders` 表中。

```sql
select cust_name, (select COUNT(*) from orders where orders.cust_id = customers.cust_id) as orders from customers order by cust_name;
```

`orders` 是一个计算字段，它是由圆括号中的子查询建立的。该子查询对检索出的每个客户执行一次。子查询中的 `WHERE` 子句使
用了**完全限定列名**。

## 联结表

SQL 最强大的功能之一就是能在数据检索查询的执行中联结（`join`）表。

### 关系表

假如有一个包含产品目录的数据库表，其中每种类别的物品占一行。对于每种物品要存储的信息包括产品描述和价格，以及生产该产品的供应商信息。

现在，假如有由同一供应商生产的多种物品，那么在何处存储供应商信息（如，供应商名、地址、联系方法等）呢？将这些数据与产品信息分开存储的理
由如下。

- 因为同一供应商生产的每个产品的供应商信息都是相同的，对每个产品重复此信息既浪费时间又浪费存储空间。
- 如果供应商信息改变（例如，供应商搬家或电话号码变动），只需改动一次即可。
- 如果有重复数据（即每种产品都存储供应商信息），很难保证每次输入该数据的方式都相同。不一致的数据在报表中很难利用。
数据无重复，显然数据是一致的，这使得处理数据更简单。

**相同数据出现多次决不是一件好事，此因素是关系数据库设计的基础**。关系表的设计就是要保证把信息分解成多个表，一类数据一个表。**各表通过某
些常用的值（即关系设计中的关系（relational））互相关联**。

#### 为什么使用联结表

分解数据为多个表能更有效地存储，更方便地处理，并且具有更大的可伸缩性。但这些好处是有代价的。如果数据存储在多个表中，怎样用单条 `SELECT`
语句检索出数据？

使用**联结**。

### 创建联结

```sql
select vend_name, prod_name, prod_price from vendors, products where vendors.vend_id = products.vend_id order by vend_name, prod_name;
```

列 `prod_name` 和 `prod_price` 在一个表中，而列 `vend_name` 在另一个表中。`FROM` 子句列出了两个表，分别是 `vendors` 和 `products`。
这两个表用 `WHERE` 子句联结。

> **笛卡儿积**（cartesian product）由**没有联结条件**的表关系返回的结果为笛卡儿积。检索出的行的数目将是第一个表中的行数乘以第二个表中
的行数。

#### 内部联结

目前为止所用的联结称为**等值联结（equijoin）**，它基于两个表之间的相等测试。这种联结也称为**内部联结**。其实，对于这种联结可以使用稍
微不同的语法来明确指定联结的类型。

```sql
select vend_name, prod_name, prod_price from vendors inner join products on vendors.vend_id = products.vend_id;
```

此语句中的 `FROM` 子句与上面的不同。两个表之间的关系是 `FROM` 子句的组成部分，以 **`INNER JOIN`** 指定。在使用这种语法时，联
结条件用特定的 **`ON`子句** 而不是 `WHERE` 子句给出。传递给 `ON` 的实际条件与传递给 `WHERE` 的相同。

> **SQL 规范首选 `INNER JOIN` 语法。此外，尽管使用 `WHERE` 子句定义联结的确比较简单，但是使用明确的联结语法能够确保不会忘记联结条件，
有时候这样做也能影响性能。不要联结不必要的表。联结的表越多，性能下降越厉害**。

当两张表的数据量比较大，又需要连接查询时，应该使用 `FROM table1 JOIN table2 ON xxx` 的语法，避免使
用 `FROM table1,table2 WHERE xxx` 的语法，因为后者会在内存中先生成一张数据量比较大的笛卡尔积表，增加了内存的开销。

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
- 允许在单条 `SELECT` 语句中多次使用相同的表

**使用表别名的主要原因之一是能在单条 `SELECT` 语句中不止一次引用相同的表**。

```sql
select cust_name from customers as c, orders as o, orderitems as oi
where c.cust_id = o.cust_id
and oi.order_num = o.order_num
and pod_id = 'TNT2';
```

与上面的例子功能一样，却更简短。

#### 不同类型的联结

内部联结或等值联结（equijoin）是简单联结。其他联结类型有自联结、自然联结和外部联结。

##### 自联结

假如你发现某物品（其 ID 为 DTNTR）存在问题，因此想知道生产该物品的供应商生产的其他物品是否也存在这些问题。
此查询要求首先找到生产 ID 为 DTNTR 的物品的供应商，然后找出这个供应商生产的其他物品：

```sql
select prod_id, prod_name from products where vend_id = (select vend_id from products where prod_id = 'DTNTR');

select p1.prod_id, p1.prod_name from products as p1, products as p2 where p1.vend_id = p2.vend_id an dp2.prod_id = 'DTNTR';
```

上面第一条使用了子查询，第二条使用了联结，并且一个表使用了两次。

> 自联结通常作为外部语句用来替代从相同表中检索数据时使用的子查询语句。虽然最终的结果是相同的，但有时候处理联结远比处理子查询快得多。应
该试一下两种方法，以确定哪一种的性能更好。

##### 自然联结

标准的联结返回所有数据，甚至相同的列多次出现。**自然联结**排除多次出现，使每个列只返回一次。

事实上，很可能永远都不会用到不是自然联结的内部联结。

##### 外部联结

许多联结将一个表中的行与另一个表中的行相关联。但有时候会需要包含没有关联行的那些行。例如，可能需要使用联结来完成以下工作：

- 对每个客户下了多少订单进行计数，包括那些至今尚未下订单的客户；
- 列出所有产品以及订购数量，包括没有人订购的产品；
- 计算平均销售规模，包括那些至今尚未下订单的客户。

在上述例子中，联结包含了那些在相关表中没有关联行的行。这种类型的联结称为**外部联结**。

```sql
select customers.cust_id, orders.order_num from customers inner join orders on customers.cust_id = orders.cust_id;
```

上面的语句很简单，就是检索所有客户及其订单。那么如果想要检索所有客户，包括没有订单的客户，如下：

```sql
select customers.cust_id, orders.order_num from customers left outer join orders on customers.cust_id = orders.cust_id;
```

这条语句使用了关键字 `OUTER JOIN` 来指定联结的类型。外部联结还包括没有关联行的行。在使用 `OUTER JOIN` 语法时，必须使用 **`RIGHT`
或 `LEFT` 关键字指定包括其所有行的表**（`RIGHT` 指的是 `OUTER JOIN` 右边的表，而 `LEFT` 指的是 `OUTER JOIN` 左边的表）。

上面的例子使用 `LEFT OUTER JOIN` 从 `FROM` 子句的左边表（`customers` 表）中选择所有行。

> 可通过颠倒 FROM 或 WHERE 子句中表的顺序，来转换外部联结形式。两种类型的外部联结可互换使用，而究竟使用哪一种纯粹是根据方便而定。

#### 带聚集函数的联结

检索所有客户及每个客户所下的订单数：

```sql
select customers.cust_id, COUNT(orders.order_num) as num_ord from customers left outer join orders on customers.cust_id = orders.cust_id group by customers.cust_id;
```

## 组合查询

MySQL 也允许执行多个查询（多条 `SELECT` 语句），并将结果作为单个查询结果集返回。这些组合查询通常称为**并**（`union`）或
**复合查询**（`compound query`）。

需要使用组合查询的情况：

- 在单个查询中从不同的表返回类似结构的数据；
- 对单个表执行多个查询，按单个查询返回数据。

```sql
select vend_id, prod_id, prod_price from products where prod_price >= 5
union
select vend_id, prod_id, prod_price from products where vend_id in (1001,1002);
```

转成多条 `where` 子句的写法：

```sql
select vend_id, prod_id, prod_price from products where prod_price >= 5 or vend_id in (1001,1002);
```

> 任何具有多个 `WHERE` 子句的 `SELECT` 语句都可以作为一个组合查询给出。这两种技术在不同的查询中性能也不同。因此，应该试一下这
两种技术，以确定对特定的查询哪一种性能更好。

### UNION 规则

- `UNION` 必须由两条或两条以上的 `SELECT` 语句组成，语句之间用关键字 `UNION` 分隔。
- `UNION` 中的每个查询**必须包含相同的列、表达式或聚集函数**（不过各个列不需要以相同的次序列出）。
- **列数据类型必须兼容**：类型不必完全相同，但必须是 DBMS 可以隐含地转换的类型（例如，不同的数值类型或不同的日期类型）。

### 包含或取消重复的行

`UNION` 从查询结果集中**自动去除了重复的行**。如果想返回所有匹配行，可使用 `UNION ALL` 而不是 `UNION`。

### 组合查询结果排序

**用 `UNION` 组合查询时，只能使用一条 `ORDER BY` 子句，它必须出现在最后一条 `SELECT` 语句之后**。对于结果集，不存在用一种方式排序一
部分，而又用另一种方式排序另一部分的情况，因此不允许使用多条 `ORDER BY` 子句。

## 全文本搜索

两个最常使用的引擎为 `MyISAM` 和 `InnoDB`，前者支持全文本搜索，而后者不支持。

使用 MySQL 的 `Match()` 和 `Against()` 函数进行全文本搜索。
