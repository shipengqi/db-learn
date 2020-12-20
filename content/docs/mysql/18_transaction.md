---
title: 事务
---

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

```sql
BEGIN
```

`START TRANSACTION` 语句后边可以跟随几个修饰符，就是它们几个，`START TRANSACTION READ ONLY;`，`START TRANSACTION READ ONLY, WITH CONSISTENT SNAPSHOT;`：

`READ ONLY`：标识当前事务是一个只读事务，也就是属于该事务的数据库操作只能读取数据，而不能修改数据。

其实只读事务中只是不允许修改那些其他事务也能访问到的表中的数据，对于临时表来说（我们使用`CREATE TMEPORARY TABL`E创建的表），由于它们只能在当前会话中可见，所以只读事务其实也是可以对临时表进行增、删、改操作的。

- `READ WRITE`：标识当前事务是一个读写事务，也就是属于该事务的数据库操作既可以读取数据，也可以修改数据。
- `WITH CONSISTENT SNAPSHOT`：启动一致性读。

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

##### 哪些语句可以回退

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

```bash
mysql> SELECT * FROM account;
+----+--------+---------+
| id | name   | balance |
+----+--------+---------+
|  1 | 狗哥   |      11 |
|  2 | 猫爷   |       2 |
+----+--------+---------+
2 rows in set (0.00 sec)

mysql> BEGIN;
Query OK, 0 rows affected (0.00 sec)

mysql> UPDATE account SET balance = balance - 10 WHERE id = 1;
Query OK, 1 row affected (0.01 sec)
Rows matched: 1  Changed: 1  Warnings: 0

mysql> SAVEPOINT s1;    # 一个保存点
Query OK, 0 rows affected (0.00 sec)

mysql> SELECT * FROM account;
+----+--------+---------+
| id | name   | balance |
+----+--------+---------+
|  1 | 狗哥   |       1 |
|  2 | 猫爷   |       2 |
+----+--------+---------+
2 rows in set (0.00 sec)

mysql> UPDATE account SET balance = balance + 1 WHERE id = 2; # 更新错了
Query OK, 1 row affected (0.00 sec)
Rows matched: 1  Changed: 1  Warnings: 0

mysql> ROLLBACK TO s1;  # 回滚到保存点s1处
Query OK, 0 rows affected (0.00 sec)

mysql> SELECT * FROM account;
+----+--------+---------+
| id | name   | balance |
+----+--------+---------+
|  1 | 狗哥   |       1 |
|  2 | 猫爷   |       2 |
+----+--------+---------+
2 rows in set (0.00 sec)
```

## 自动提交

MySQL中有一个系统变量 `autocommit`：

默认情况下，如果我们不显式的使用START TRANSACTION或者BEGIN语句开启一个事务，那么每一条语句都算是一个独立的事务，这种特性称之为事务的**自动提交**。

如果我们想关闭这种自动提交的功能，可以使用下边两种方法：

- 显式的的使用 `START TRANSACTION` 或者 `BEGIN` 语句开启一个事务。

这样在本次事务提交或者回滚前会暂时关闭掉自动提交的功能。

把系统变量 `autocommit` 的值设置为 `OFF`，就像这样：

```sql
SET autocommit = OFF;
```

这样的话，我们写入的多条语句就算是属于同一个事务了，直到我们显式的写出COMMIT语句来把这个事务提交掉，或者显式的写出ROLLBACK语句来把这个事务回滚掉。

## 隐式提交

当我们使用START TRANSACTION或者BEGIN语句开启了一个事务，或者把系统变量 autocommit的 值设置为OFF时，事务就不会进行自动提交，但是如果我们输入了某些语句之后就会悄悄的提交掉，就像我们输入了COMMIT语句了一样，这种因为某些特殊的语句而导致事务提交的情况称为**隐式提交**。

隐式提交的语句包括：

- 定义或修改数据库对象的数据定义语言（Data definition language，缩写为：DDL）。

所谓的数据库对象，指的就是数据库、表、视图、存储过程等等这些东西。当我们使用`CREATE、ALTER、DROP`等语句去修改这些所谓的数据库对象时，就会隐式的提交前边语句所属于的事务

- 隐式使用或修改mysql数据库中的表
当我们使用 `ALTER USER`、`CREATE USER`、`DROP USER`、`GRANT`、`RENAME USER`、`REVOKE`、`SET PASSWORD` 等语句时也会隐式的提交前边语句所属于的事务。

- 事务控制或关于锁定的语句
当我们在一个事务还没提交或者回滚时就又使用START TRANSACTION或者BEGIN语句开启了另一个事务时，会隐式的提交上一个事务

或者当前的autocommit系统变量的值为OFF，我们手动把它调为ON时，也会隐式的提交前边语句所属的事务。

- 加载数据的语句

比如我们使用LOAD DATA语句来批量往数据库中导入数据时，也会隐式的提交前边语句所属的事务。

- 关于MySQL复制的一些语句

使用START SLAVE、STOP SLAVE、RESET SLAVE、CHANGE MASTER TO等语句时也会隐式的提交前边语句所属的事务。

- 其它的一些语句

使用ANALYZE TABLE、CACHE INDEX、CHECK TABLE、FLUSH、 LOAD INDEX INTO CACHE、OPTIMIZE TABLE、REPAIR TABLE、RESET等语句也会隐式的提交前边语句所属的事务。
