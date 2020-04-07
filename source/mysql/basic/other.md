---
title: 其他
---
# 其他
## 视图
```sql
select cust_name from customers, orders, orderitems
where customers.cust_id = orders.cust_id
and orderitems.order_num = orders.order_num
and pod_id = 'TNT2';
```

上面的语句，涉及到三个表，用来检索订购了某个特定产品的客户。任何需要这个数据的人都必须理解相关表的结构，
并且知道如何创建查询和对表进行联结。为了检索其他产品（或多个产品）的相同数据，必须修改最后的 `WHERE` 子句。

假如可以把整个查询包装成一个名为 `productcustomers` 的虚拟表，则可以如下轻松地检索出相同的数据：
```sql
select cust_name from productcustomers where pod_id = 'TNT2';
```

这就是**视图**的作用。`productcustomers` 是一个视图，作为**视图，它不包含表中应该有的任何列或数据，它包含的是一个 SQL 查询**（与上
面用以正确联结表的相同的查询）。

### 为什么使用视图
- 重用 SQL 语句。
- 简化复杂的 SQL 操作。在编写查询后，可以方便地重用它而不必知道它的基本查询细节。
- 使用表的组成部分而不是整个表。
- 保护数据。可以给用户授予表的特定部分的访问权限而不是整个表的访问权限。
- 更改数据格式和表示。视图可返回与底层表的表示和格式不同的数据。

视图创建之后，可以用与表基本相同的方式利用它们。可以对视图执行 `SELECT` 操作，过滤和排序数据，将视图联结到其他视图或表，甚至能添加
和更新数据。

**视图本身不包含数据，因此它们返回的数据是从其他表中检索出来的。在添加或更改这些表中的数据时，视图将返回改变过的数据**。

### 性能问题
因为视图不包含数据，所以每次使用视图时，都必须处理查询执行时所需的任一个检索。如果你用多个联结和过滤创建了复杂的视图或者嵌套了视图，可能
会发现性能下降得很厉害。因此，在部署使用了大量视图的应用前，应该进行测试。

### 规则和限制
- 与表一样，视图必须唯一命名（不能给视图取与别的视图或表相同的名字）。
- 对于可以创建的视图数目没有限制。
- 为了创建视图，必须具有足够的访问权限。这些限制通常由数据库管理人员授予。
- 视图可以嵌套，即可以利用从其他视图中检索数据的查询来构造一个视图。
- `ORDER BY` 可以用在视图中，但如果从该视图检索数据 `SELECT` 中也含有 `ORDER BY`，那么该视图中的 `ORDER BY` 将被覆盖。
- 视图不能索引，也不能有关联的触发器或默认值。
- 视图可以和表一起使用。例如，编写一条联结表和视图的 `SELECT` 语句。

### 使用视图
- `CREATE VIEW` 语句创建视图。
- `SHOW CREATE VIEW viewname` 查看创建视图的语句。
- `DROP` 删除视图，其语法为 `DROP VIEW viewname`。
- 更新视图时，可以先用 `DROP` 再用 `CREATE`，也可以直接用 `CREATE OR REPLACE VIEW`。如果要更新的视图不存在，
则第 2 条更新语句会创建一个视图；如果要更新的视图存在，则第2条更新语句会替换原有视图

#### 用视图重新格式化检索出的数据
视图的另一常见用途是重新格式化检索出的数据。
```sql
select Concat(RTrim(vend_name)), '(' , RTrim(vend_country), ')') as vend_title from vendors order by vend_name;
```
上面的语句，格式化了结果，如果要经常用，可以创建一个视图：
```sql
create view vendorlocations as
select Concat(RTrim(vend_name)), '(' , RTrim(vend_country), ')') as vend_title from vendors order by vend_name;
```

然后可以直接使用：
```sql
select * from vendorlocations;
```

视图也可以用来过滤数据，或者计算字段。

#### 更新视图
并非所有视图都是可更新的。基本上可以说，如果 MySQL 不能正确地确定被更新的基数据，则不允许更新（包括插入和删除）。
意味着，如果视图定义中有以下操作，则不能进行视图的更新：
- 分组（使用 `GROUP BY` 和` HAVING`）；
- 联结；
- 子查询；
- 并；
- 聚集函数（`Min()`、`Count()`、`Sum()` 等）；
- `DISTINCT`；
- 导出（计算）列。

## 存储过程
以下的情形。
- 为了处理订单，需要核对以保证库存中有相应的物品。
- 如果库存有物品，这些物品需要预定以便不将它们再卖给别的人，并且要减少可用的物品数量以反映正确的库存量。
- 库存中没有的物品需要订购，这需要与供应商进行某种交互。
- 关于哪些物品入库（并且可以立即发货）和哪些物品退订，需要通知相应的客户。

执行这个处理需要针对许多表的多条 MySQL 语句。此外，需要执行的具体语句及其次序也不是固定的，它们可能会（和将）根据哪些物品在库存中哪
些不在而变化。

可以创建存储过程。**存储过程简单来说，就是为以后的使用而保存的一条或多条 MySQL 语句的集合**。

### 为什么要使用存储过程
- 简化复杂的操作
- 由于不要求反复建立一系列处理步骤，这保证了数据的完整性。所有开发人员和应用程序都使用同一（试验和测试）存储过程，则所使用的代码都是
相同的。这一点的延伸就是防止错误。防止错误保证了数据的一致性。
- 简化对变动的管理。如果表名、列名或业务逻辑（或别的内容）有变化，只需要更改存储过程的代码。使用它的人员甚至不需要知道这些变化。
- 提高性能。存储过程比使用单独的 SQL 语句要快。

总结就是，简单、安全、高性能。

**缺陷**：
- 存储过程的编写比基本 SQL 语句复杂
- 你可能没有创建存储过程的安全访问权限。

### 使用
MySQL 称存储过程的执行为调用，因此 MySQL 执行存储过程的语句为 `CALL`。`CALL` 接受存储过程的名字以及需要传递给它的任意参数。
```sql
call productpricing(@pricelow, @pricehigh, @priceacerage);
```
执行名为 `productpricing` 的存储过程，它计算并返回产品的最低、最高和平均价格。

**因为存储过程实际上是一种函数，所以存储过程名后需要有 `()` 符号（即使不传递参数也需要）**。


#### 创建
```sql
create procedure productpricing()
begin
select Avg(prod_price) as priceacerage from products;
end;
```
此存储过程名为 `productpricing`，用 `CREATE PROCEDURE productpricing()` 语句定义。如果存储过程接受参数，它们将在 `()` 中列举
出来。此存储过程没有参数，但后跟的 `()` 仍然需要。`BEGIN` 和 `END` 语句用来限定存储过程体，过程体本身仅是一个简单的 `SELECT` 语句。

#### 删除
```sql
drop procedure productpricing;
```

#### 使用参数
```sql
create procedure productpricing(
  out pl DECIMAL(8,2),
  out ph DECIMAL(8,2),
  out pa DECIMAL(8,2)
)
begin
  select Min(prod_price) into pl from products;
  select Max(prod_price) into ph from products;
  select Avg(prod_price) into pa from products;
end;
```

此存储过程接受 3 个参数：`pl` 存储产品最低价格，`ph` 存储产品最高价格，`pa` 存储产品平均价格。每个参数必须具有指定的类型，这里使用十
进制值。关键字 `OUT` 指出相应的参数用来从存储过程传出一个值（返回给调用者）。

MySQL 支持三种类型参数：
- `IN` 传递给存储过程
- `OUT` 从存储过程传出 
- `INOUT` 对存储过程传入和传出

存储过程的代码位于 `BEGIN` 和 `END` 语句内，它们是一系列 `SELECT` 语句，用来检索值，然后保存到相应的变量（通过指定 `INTO` 关键字）。

调用此修改过的存储过程，必须指定 3 个变量名，如下所示：
```sql
call productpricing(@pricelow, @pricehigh, @priceacerage);
```
此存储过程要求 3 个参数，因此必须正好传递 3 个参数。存储过程将保存结果到这 3 个变量。

> **所有 MySQL 变量都必须以 `@` 开始**。

调用这条语句并不显示任何数据。为了显示检索出的产品平均价格，可使用下面的语句：
```sql
select @priceacerage;

select @pricelow, @pricehigh, @priceacerage;
```

##### 使用 IN 和 OUT
```sql
create procedure ordertotal(
  in onumber int,
  inout ototal DECIMAL(8,2)
)
begin
  select Sum(item_price*quantity) from orderitems where order_num = onumber into otital;
end;
```
`onumber` 定义为 `IN`，因为需要传订单号给存储过程。`ototal` 定义为 `OUT`，因为要从存储过程返回合计。`SELECT` 语句使用这两个参
数，`WHERE` 子句使用 `onumber` 选择正确的行，`INTO` 使用 `ototal` 存储计算出来的合计。

```sql
call ordertotal(2005, @total);
```

必须给 `ordertotal` 传递两个参数；第一个参数为订单号，第二个参数为包含计算出来的合计的变量名。

显示此合计：
```sql
select @total;
```
#### 智能存储过程
存储过程只有在包含业务规则和智能处理时，才真正显现出来他的作用。

例如，需要获得一份订单合计，但需要对合计增加营业税，不过只针对某些顾客。
- 获得合计（与以前一样）
- 把营业税有条件地添加到合计
- 返回合计（带或不带税）

```sql
-- Name: ordertotal
-- Parameters: onumber = order number
--             taxable = 0 if not taxable, 1 if taxable
--             ototal  = order total variable
create procedure odertotal(
  in onumber int,
  in taxable boolean,
  out ototal decimal(8, 2)
) comment 'Obtain order total, optionally adding tax'
begin 
-- Declare variable for total
declare total decimal(8, 2);
-- Declare tax percentage
declare taxrate int default 6;
-- Get the order total
select sum(item_price*quantity) from orderitems where order_num = onumber into total;
-- Is this taxable?
IF taxable THEN 
   -- Yes, so add taxrate to the total
   select total+(total/100*taxrate) into total;
END IF;
-- And finally, save to out variable
select total into ototal;
end;
```

`--` 表示注释。参数 `taxable`，它是一个布尔值，表示是否增加税。
`DECLARE` 语句定义了两个局部变量。`DECLARE` 要求指定变量名和数据类型，它也支持可选的默认值（这里的 `taxrate` 的默认被设置为 `6%`）

`IF` 语句检查 `taxable` 是否为真，如果为真，则用另一 `SELECT` 语句增加营业税到局部变量 `total`。
最后，用另一 `SELECT` 语句将 `total` 保存到 `ototal`。
#### 检查存储过程
`SHOW CREATE PROCEDURE name` 和 `SHOW PROCEDURE STATUS name`。

## 游标
MySQL 检索操作返回一组称为结果集的行。这组返回的行都是与 SQL 语句相匹配的行（零行或多行）。

有时，需要在检索出来的行中前进或后退一行或多行。这就是使用游标的原因。**游标**（cursor）是一个存储在 MySQL 服务器上的数据库查询，
它**不是一条 SELECT 语句，而是被该语句检索出来的结果集**。在存储了游标之后，应用程序可以根据需要滚动或浏览其中的数据。

> MySQL 游标只能用于存储过程（和函数）。

### 使用游标
- 在能够使用游标前，必须声明（定义）它。这个过程实际上没有检索数据，它只是定义要使用的 `SELECT` 语句。
- 一旦声明后，必须打开游标以供使用。这个过程用前面定义的 `SELECT` 语句把数据实际检索出来。
- 对于填有数据的游标，根据需要取出（检索）各行。
- 在结束游标使用时，必须关闭游标。

#### 创建游标
`DECLARE` 语句创建游标。`DECLARE` 命名游标，并定义相应的 `SELECT` 语句，根据需要带 `WHERE` 和其他子句。

```sql
create procedure processorders()
begin
  declare ordernumbers cursor
  for
  select order_num from orders;
end;
```
存储过程处理完成后，游标就消失。

#### 打开关闭
打开使用：`OPEN ordernumbers;`
关闭使用：`CLOSE ordernumbers;`

> 使用声明过的游标不需要再次声明，用 `OPEN` 语句打开它就可以了。
> 如果你不明确关闭游标，MySQL 将会在到达 `END` 语句时自动关闭它。

#### 使用游标数据
在一个游标被打开后，可以使用 `FETCH` 语句分别访问它的每一行。`FETCH` 指定检索什么数据（所需的列），检索出来的数据存储在什么地方。

```sql
create procedure processorders()
begin
  -- Declare local variables
  declare o int;

  -- Delare the cursor
  declare ordernumbers cursor
  for
  select order_num from orders;

  -- open the cursor
  open ordernumbers;

  -- get order number
  fetch ordernumbers into o;

  -- close the cursor
  close ordernumbers;
end;
```

## 触发器
如果你想要某条语句（或某些语句）在事件发生时自动执行，怎么办呢？使用**触发器**。

触发器是 MySQL 响应以下任意语句而自动执行的一条 MySQL 语句（或位于 `BEGIN` 和 `END` 语句之间的一组语句）：
- `DELETE`
- `INSERT`
- `UPDATE`

创建触发器时，需要给出 4 条信息：
- 唯一的触发器名；
- 触发器关联的表；
- 触发器应该响应的活动（`DELETE`、`INSERT` 或 `UPDATE`）；
- 触发器何时执行（处理之前或之后）

### 创建
用 `CREATE TRIGGER` 语句创建。
```sql
create trigger newproduct after insert on products for each row select 'Procduct added';
```
创建名为 `newproduct` 的新触发器。触发器可在一个操作发生之前或之后执行，这里给出了 `AFTER INSERT`，所以此触发器将在 `INSERT` 语句成
功执行后执行。这个触发器还指定 `FOR EACH ROW`，因此代码对每个插入行执行。在这个例子中，文本 `Product added` 将对每个插入的行显示一次。
使用 `INSERT` 语句添加一行或多行到 `products` 中，你将看到对每个成功的插入，显示 `Product added` 消息。

> **每个表每个事件每次只允许一个触发器。因此，每个表最多支持 6 个触发器（每条 `INSERT`、`UPDATE` 和 `DELETE` 的之前和之后）**。

> 如果 `BEFORE` 触发器失败，则 MySQL 将不执行请求的操作。此外，如果 `BEFORE` 触发器或语句本身失败，MySQL 将不执行 `AFTER` 触发
器（如果有的话）。

> MySQL 触发器中不支持 `CALL` 语句。也就是不能从触发器内调用存储过程。

### 删除
删除触发器使用：`DROP TRIGGER newproduct;`。为了**修改一个触发器，必须先删除它，然后再重新创建**。

## 事务
事务处理是一种机制，用来管理必须成批执行的 MySQL 操作，以保证数据库不包含不完整的操作结果。利用事务处理，可以保证一组操作不会中途停止，
它们或者作为整体执行，或者完全不执行（除非明确指示）。如果没有错误发生，整组语句提交给（写到）数据库表。如果发生错误，则进行回退（撤销）
以恢复数据库到某个已知且安全的状态。

关于事务处理需要知道的几个术语：
- **事务**（transaction）指一组 SQL 语句；
- **回退**（rollback）指撤销指定 SQL 语句的过程；
- **提交**（commit）指将未存储的 SQL 语句结果写入数据库表；
- **保留点**（savepoint）指事务处理中设置的临时占位符（`place- holder`），你可以对它发布回退（与回退整个事务处理不同）

### 控制事务处理
管理事务处理的关键在于将 SQL 语句组分解为逻辑块，并明确规定数据何时应该回退，何时不应该回退。

下面的语句来标识事务的开始：
```sql
START TRANSACTION
```

#### ROLLBACK
`ROLLBACK` 命令用来回退（撤销）MySQL 语句：
```sql
select * from ordertotals;
start transaction;
delete from ordertotals;
select * from ordertotals;
rollback;
select * from ordertotals;
```

先执行一条 `SELECT` 以显示该表不为空。然后开始一个事务处理，用一条 `DELETE` 语句删除 `ordertotals` 中的所有行。
另一条 `SELECT` 语句验证 `ordertotals` 确实为空。这时用一条 `ROLLBACK` 语句回退 `START TRANSACTION` 之后的所有语句，最后
一条 `SELECT` 语句显示该表不为空。

**`ROLLBACK` 只能在一个事务处理内使用（在执行一条 `START TRANSACTION` 命令之后）**。

##### 哪些语句可以回退？
事务处理用来管理 `INSERT`、`UPDATE` 和 `DELETE` 语句。不能回退 `SELECT` 语句。（这样做也没有什么意义）不能回退 `CREATE` 或 `DROP` 
操作。事务处理块中可以使用这两条语句，但如果你执行回退，它们不会被撤销。

#### COMMIT
一般的 MySQL 语句都是直接针对数据库表执行和编写的。这就是所谓的**隐含提交**（implicit commit），即提交（写或保存）操作是自动进行的。

**在事务处理块中，提交不会隐含地进行**。为进行明确的提交，使用 `COMMIT` 语句：
```sql
start transaction;
delete from orderitems where order_num = 20005;
delete from orders where order_num = 20005;
commit;
```

> 当 `COMMIT` 或 `ROLLBACK` 语句执行后，事务会自动关闭（将来的更改会隐含提交）。

#### 保留点
简单的 `ROLLBACK` 和 `COMMIT` 语句就可以写入或撤销整个事务处理。复杂的事务处理可能需要部分提交或回退。

为了支持回退部分事务处理，必须能在事务处理块中合适的位置放置占位符。这样，如果需要回退，可以回退到某个占位符。这些占位符称为**保留点**。

创建占位符，可使用 `SAVEPOINT` 语句：`SAVEPOINT delete1;`。
每个保留点都取标识它的唯一名字，以便在回退时，MySQL 知道要回退到何处。

回退到本例给出的保留点，可执行：`ROLLBACK TO delete1;`

> 保留点在事务处理完成（执行一条 `ROLLBACK` 或 `COMMIT`）后自动释放。

## 用户管理
**在现实世界的日常工作中，决不能使用 `root`**。应该创建一系列的账号，有的用于管理，有的供用户使用，有的供开发人员使用，等等。

MySQL 用户账号和信息存储在名为 `mysql` 的库中。一般不需要直接访问 `mysql` 数据库和表，但有时需要直接访问。需要直接访问它
的时机之一是在需要获得所有用户账号列表时。

`mysql` 库有一个名为 `user` 的表，它包含所有用户账号。`user` 表有一个名为 `user` 的列，它存储用户登录名。