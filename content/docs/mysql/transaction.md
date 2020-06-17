---
title: 事务
---

# 事务
事务的四个特性：ACID（Atomicity、Consistency、Isolation、Durability，即原子性、一致性、隔离性、持久性）。

事务处理是一种机制，用来管理必须成批执行的 MySQL 操作，以保证数据库不包含不完整的操作结果。利用事务处理，可以保证一组操作不会中途停止，
它们或者作为整体执行，或者完全不执行（除非明确指示）。如果没有错误发生，整组语句提交给（写到）数据库表。如果发生错误，则进行回退（撤销）
以恢复数据库到某个已知且安全的状态。

关于事务处理需要知道的几个术语：
- **事务**（transaction）指一组 SQL 语句；
- **回退**（rollback）指撤销指定 SQL 语句的过程；
- **提交**（commit）指将未存储的 SQL 语句结果写入数据库表；
- **保留点**（savepoint）指事务处理中设置的临时占位符（`place- holder`），你可以对它发布回退（与回退整个事务处理不同）

## 语法


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