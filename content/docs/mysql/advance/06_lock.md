---
title: 锁
weight: 6
---

## 锁分类

- 从性能上分为**乐观锁**（用版本对比或 CAS 机制）和**悲观锁**，
  - 乐观锁适合读操作较多的场景。
  - 悲观锁适合写操作较多的场景，如果在写操作较多的场景使用乐观锁会导致大量的重试，长时间占用 CPU，降低性能。
- 从对数据操作的粒度分，分为**表锁**、**页锁**、**行锁**。
- 从对数据库操作的类型分，分为**读锁**和**写锁**(都属于悲观锁)，还有**意向锁**。
  - 读锁（共享锁，S 锁（**S**hared））：针对同一份数据，多个读操作可以同时进行而不会互相影响，比如：`select * from T where id=1 lock in share mode`
  - 写锁（排它锁，X 锁(e**X**clusive)）：当前写操作没有完成前，它会阻断其他写锁和读锁，数据修改操作都会加写锁，比如：`select * from T where id=1 for update`
  - 意向锁（Intention Lock）：又称 I 锁，针对**表锁**，主要是为了提高加表锁的效率，是 MySQL 数据库自己加的。**当有事务给表的数据行加了共享锁或排他锁时，同时也会给表设置一个标识，代表已经有行锁了，其他事务要想对表加表锁时，就不必逐行判断有没有行锁可能跟表锁冲突了，直接读这个标识就可以确定自己该不该加表锁**。特别是表中的记录很多时，逐行判断加表锁的方式效率很低。而这个标识就是意向锁。
    - 意向共享锁，IS 锁（**I**ntention **S**hared）：对整个表加共享锁之前，需要先获取到意向共享锁。
    - 意向排他锁，IX 锁（**I**ntention **X**clusive）：对整个表加排他锁之前，需要先获取到意向排他锁。

### 表锁

每次操作锁住整张表。**开销小，加锁快**；不会出现死锁；锁定粒度大，发生锁冲突的概率最高，并发度最低；一般用在整表数据迁移的场景。

```sql
‐‐手动增加表锁
lock table 表名称 read(write),表名称2 read(write),表名称3 read;
‐‐查看表上加过的锁
show open tables;
‐‐删除表锁
unlock tables;
```

### 页锁

只有 BDB 存储引擎支持页锁，页锁就是在页的粒度上进行锁定，锁定的数据资源比行锁要多，因为一个页中可以有多个行记录。当使用页锁的时候，会出现数据浪费的现象，但这样的浪费最多也就是一个页上的数据行。页锁的开销介于表锁和行锁之间，会出现死锁。锁定粒度介于表锁和行锁之间，并发度一般。

### 行锁

每次操作锁住一行数据。**开销大，加锁慢**（行锁需要先到表中找到对应的那行记录，所以说开销大；而表锁是直接在整张表上加一个标记）；会出现死锁；锁定粒度最小，发生锁冲突的概率最低，并发度最高。

InnoDB 相对于 MYISAM 的最大不同有两点：

- InnoDB 支持事务
- InnoDB 支持行级锁

注意，InnoDB 的行锁实际上是针**对索引字段加的锁**（在索引对应的索引项上做标记），**不是针对整个行记录加的锁**。并且该索引不能失效（或者不存在索引），否则会从行锁升级为表锁。**不管是一级索引还是二级索引，只要更新时使用了索引，就会对索引字段加锁，否则就会升级为表锁**。（**RR 级别会升级为表锁**，RC 级别不会升级为表锁）

比如在 RR 级别执行如下 SQL：

```sql
‐‐ name 字段无索引
select * from account where name = 'lilei' for update; 
```

由于 `name` 无索引，升级为表锁，则**其它 session 对该表任意一行记录做修改，插入，删除的操作都会被阻塞住**。

#### 关于 RR 级别行锁升级为表锁的原因分析

因为在 RR 隔离级别下，需要解决幻读问题，当**字段没有索引时，MySQL 无法有效使用间隙锁精确锁定范围**，只能退而求其次使用表锁来**保证隔离性**。所以在遍历扫描聚簇索引记录时，为了防止扫描过的索引间隙被其它事务插入记录（幻读问题），从而导致数据不一致，MySQL 的解决方案就是把所有扫描过的索引记录和间隙都锁上，这里要**注意，并不是直接将整张表加表锁，因为不一定能加上表锁，可能会有其它事务锁住了表里的其它行记录**。

RC 读已提交级别不会升级为表锁，因为 RC 级别下，**不需要解决幻读问题**，所以不需要加锁。

### 间隙锁（Gap Lock）

间隙锁，锁的就是**两个值之间的空隙**，间隙锁是在**可重复读隔离级别下才会生效**。MySQL 默认隔离级别是 RR，有幻读问题，**间隙锁是可以解决幻读问题的**。

![gap-lock](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/gap-lock.png)

上面的表中，间隙就有 id 为 `(1,4)`，`(4,112)`，`(115,正无穷)` 这三个区间，执行 sql：

```sql
select * from account where id = 10 for update;
```

其他 Session 没法在这个 `(4,112)` 这个间隙范围里插入任何数据。

也就是说，**只要在间隙范围内锁了一条不存在的记录，整个间隙范围都会被锁住**，这样就能防止其它 Session 在这个间隙范围内插入数据，就解决了可重复读隔离级别的幻读问题。

![gap-lock1](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/gap-lock1.png)

#### 为什么说间隙锁可以解决幻读问题

例如还是下面这些数据：

![gap-lock](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/gap-lock.png)

```sql
set tx_isolation='repeatable‐read';


-- 事务1
START TRANSACTION;
SELECT * FROM account; -- 快照读，看到 6 条记录

-- 此时事务 2 插入一条 name=小郭 的新记录并提交
insert into account values(5,'小郭',12256487569,'洛阳');

SELECT * FROM account; -- 仍然看到 6 条记录（快照读）
update account set address='信阳' where id = 5；; -- 当前读
SELECT * FROM account; -- 看到 7 条记录
COMMIT;
```

可以看到，事务 1 最后看到了 7 条记录。还是有幻读问题的。

但是如果在 `(4,112)` 这个区间加一把 **间隙锁**，这个间隙就不会允许其他事务再插入数据，就不会出现幻读问题了。

### 临键锁（Next-Key Lock）

**临键锁 = 间隙锁 + 行锁**，既能锁住某条记录，又能阻止别的事务将新记录插入被锁记录前边的间隙。

```sql
select * from account where id > 4 and id <= 122 for update;
```

`(4,112)` 这个间隙和 `112` 这条记录都会被锁住，相当于 `(4,112]`。

### 锁定读

采用加锁方式解决脏读、不可重复读、幻读这些问题时，读取一条记录时需要获取一下该记录的 S 锁，其实这是不严谨的，有时候想在读取记录时就获取记录的 X 锁，来禁止别的事务读写该记录，MySQL 提供了两种比较特殊的 `SELECT` 语句格式：

- 对读取的记录加 S 锁：

```sql
SELECT ... LOCK IN SHARE MODE;
```

- 对读取的记录加 X 锁：

```sql
SELECT ... FOR UPDATE;
```

也就是在普通的 `SELECT` 语句后边加 `FOR UPDATE`，如果当前事务执行了该语句，那么它会为读取到的记录加 X 锁，这样既不允许别的事务获取这些记录的S锁（比方说别的事务使用 `SELECT ... LOCK IN SHARE MODE` 语句来读取这些记录），也不允许获取这些记录的 X 锁。

### MyISAM 和 InnoDB 的锁

MyISAM 在执行查询语句 `SELECT` 前，会自动给涉及的所有表加读锁，在执行 `update`、`insert`、`delete` 操作会自动给涉及的表加写锁。MyISAM 没有事务，就算是加的表锁，也只是执行一条 SQL 语句时才会生效，执行完 SQL 语句后，会自动释放锁。

InnoDB 在执行查询语句 `SELECT` 时(非串行隔离级别)，不会加锁。但是 `update`、`insert`、`delete` 操作会加行锁。

读锁会阻塞写，但是不会阻塞读。而写锁则会把读和写都阻塞。

InnoDB 存储引擎由于实现了行级锁定，虽然在锁定机制的实现方面所带来的性能损耗可能比表级锁定会要更高一下，但是在整体并发处理能力方面要远远优于MYISAM的表级锁定的。当系统并发量高的时候，InnoDB 的整体性能和 MYISAM 相比就会有比较明显的优势了。

## 锁等待分析

通过检查 `innodb_row_lock` 状态变量来分析系统上的行锁的争夺情况

```sql
show status like 'innodb_row_lock%';
```

- `Innodb_row_lock_current_waits`：当前正在等待锁定的数量。
- `Innodb_row_lock_time`：从系统启动到现在锁定总时间长度。
- `Innodb_row_lock_time_avg`：每次等待所花平均时间。
- `Innodb_row_lock_time_max`：从系统启动到现在等待最久的一次所花的时间。
- `Innodb_row_lock_waits`：系统启动后到现在总共等待的次数。

在 MySQL 8.0，这些统计变量不再更新或直接消失。

在 MySQL 8.0 之后，可以通过 `performance_schema` 数据库的表来查看锁的相关信息。

```sql
‐‐ 查看事务
select * from information_schema.INNODB_TRX;
‐‐ 查看锁，8.0 之后需要换成表 performance_schema.data_locks
select * from information_schema.INNODB_LOCKS;
‐‐ 查看锁等待，8.0 之后需要换成表 performance_schema.data_lock_waits
select * from information_schema.INNODB_LOCK_WAITS;


‐‐ 释放锁，trx_mysql_thread_id 可以从 INNODB_TRX 表里查看到
kill <trx_mysql_thread_id>;
```

## 死锁分析

```sql
‐‐ 查看死锁，status 字段中的内容可以帮助进行死锁分析
show engine innodb status;
```

示例：

```sql
-- 插入测试数据
CREATE TABLE hero (
    id INT,
    name VARCHAR(100),
    country varchar(100),
    PRIMARY KEY (id),
    KEY idx_name (name)
) Engine=InnoDB CHARSET=utf8;

INSERT INTO hero VALUES
    (1, 'l刘备', '蜀'),
    (3, 'z诸葛亮', '蜀'),
    (8, 'c曹操', '魏'),
    (15, 'x荀彧', '魏'),
    (20, 's孙权', '吴');


-- 8.0 之后需要换成变量 transaction_isolation
set tx_isolation='repeatable‐read';

-- session 1 执行：
begin;
select * from hero where id=1 for update;
-- session 2 执行：
begin;
select * from hero where id=3 for update;
-- session 1 执行：
select * from hero where id=3 for update;
-- session 2 执行：
select * from hero where id=1 for update;

-- 查看近期死锁日志信息，根据 DEADLOCK 关键字搜索，分析：
show engine innodb status;
```

大多数情况 MySQL 可以自动检测死锁并回滚产生死锁的那个事务，但是有些情况 MySQL 没法自动检测死锁，这种情况可以通过日志分析找到对应事务线程 id，可以通过 kill 杀掉。

```perl
mysql> SHOW ENGINE INNODB STATUS\G
...
------------------------
LATEST DETECTED DEADLOCK
------------------------
2025-05-14 14:53:26 0x2414
*** (1) TRANSACTION:
TRANSACTION 121629, ACTIVE 8 sec starting index read
mysql tables in use 1, locked 1
LOCK WAIT 3 lock struct(s), heap size 1128, 2 row lock(s)
MySQL thread id 163, OS thread handle 7184, query id 332329 localhost ::1 root statistics
select * from hero where id=3 for update

*** (1) HOLDS THE LOCK(S):
RECORD LOCKS space id 26 page no 4 n bits 72 index PRIMARY of table `asdb`.`hero` trx id 121629 lock_mode X locks rec but not gap
Record lock, heap no 2 PHYSICAL RECORD: n_fields 5; compact format; info bits 0
 0: len 4; hex 80000001; asc     ;;
 1: len 6; hex 00000001daf9; asc       ;;
 2: len 7; hex 81000000a10110; asc        ;;
 3: len 7; hex 6ce58898e5a487; asc l      ;;
 4: len 3; hex e89c80; asc    ;;


*** (1) WAITING FOR THIS LOCK TO BE GRANTED:
RECORD LOCKS space id 26 page no 4 n bits 72 index PRIMARY of table `asdb`.`hero` trx id 121629 lock_mode X locks rec but not gap waiting
Record lock, heap no 3 PHYSICAL RECORD: n_fields 5; compact format; info bits 0
 0: len 4; hex 80000003; asc     ;;
 1: len 6; hex 00000001daf9; asc       ;;
 2: len 7; hex 81000000a1011d; asc        ;;
 3: len 10; hex 7ae8afb8e8919be4baae; asc z         ;;
 4: len 3; hex e89c80; asc    ;;


*** (2) TRANSACTION:
TRANSACTION 121630, ACTIVE 5 sec starting index read
mysql tables in use 1, locked 1
LOCK WAIT 3 lock struct(s), heap size 1128, 2 row lock(s)
MySQL thread id 164, OS thread handle 9416, query id 332330 localhost ::1 root statistics
select * from hero where id=1 for update

*** (2) HOLDS THE LOCK(S):
RECORD LOCKS space id 26 page no 4 n bits 72 index PRIMARY of table `asdb`.`hero` trx id 121630 lock_mode X locks rec but not gap
Record lock, heap no 3 PHYSICAL RECORD: n_fields 5; compact format; info bits 0
 0: len 4; hex 80000003; asc     ;;
 1: len 6; hex 00000001daf9; asc       ;;
 2: len 7; hex 81000000a1011d; asc        ;;
 3: len 10; hex 7ae8afb8e8919be4baae; asc z         ;;
 4: len 3; hex e89c80; asc    ;;


*** (2) WAITING FOR THIS LOCK TO BE GRANTED:
RECORD LOCKS space id 26 page no 4 n bits 72 index PRIMARY of table `asdb`.`hero` trx id 121630 lock_mode X locks rec but not gap waiting
Record lock, heap no 2 PHYSICAL RECORD: n_fields 5; compact format; info bits 0
 0: len 4; hex 80000001; asc     ;;
 1: len 6; hex 00000001daf9; asc       ;;
 2: len 7; hex 81000000a10110; asc        ;;
 3: len 7; hex 6ce58898e5a487; asc l      ;;
 4: len 3; hex e89c80; asc    ;;

*** WE ROLL BACK TRANSACTION (2)
...

```

死锁信息，就是 `LATEST DETECTED DEADLOCK` 这一部分。

第一句：

```yaml
2025-05-14 14:53:26 0x2414
```

死锁发生的时间是：`2025-05-14 14:53:26`，后边的一串十六进制 `0x2414` 表示的**操作系统为当前 session 分配的线程的线程 id**。

然后是关于死锁发生时第一个事务的有关信息：

```yaml
*** (1) TRANSACTION:

# 事务 id 为 121629，事务处于 ACTIVE 状态已经 8 秒了，事务现在正在做的操作就是：“starting index read”
TRANSACTION 121629, ACTIVE 8 sec starting index read
# 此事务使用了 1 个表，为 1 个表上了锁
mysql tables in use 1, locked 1
# 此事务处于 LOCK WAIT 状态，拥有 3 个锁结构，heap size 是为了存储锁结构而申请的内存大小（可以忽略），其中有 2 个行锁的结构
LOCK WAIT 3 lock struct(s), heap size 1128, 2 row lock(s)
# 本事务所在线程的 id 是 163（MySQL 自己命名的线程 id），该线程在操作系统级别的 id 就是那一长串数字，当前查询的 id 为 332329（MySQL 内部使用，可以忽略），还有用户名主机信息
MySQL thread id 163, OS thread handle 7184, query id 332329 localhost ::1 root statistics
# 本事务发生阻塞的语句
select * from hero where id=3 for update

# 表示该事务获取到的锁信息
*** (1) HOLDS THE LOCK(S):
RECORD LOCKS space id 26 page no 4 n bits 72 index PRIMARY of table `asdb`.`hero` trx id 121629 lock_mode X locks rec but not gap
Record lock, heap no 2 PHYSICAL RECORD: n_fields 5; compact format; info bits 0
# 主键值为 1
 0: len 4; hex 80000001; asc     ;;
 1: len 6; hex 00000001daf9; asc       ;;
 2: len 7; hex 81000000a10110; asc        ;;
 3: len 7; hex 6ce58898e5a487; asc l      ;;
 4: len 3; hex e89c80; asc    ;;

# 本事务当前在等待获取的锁：
*** (1) WAITING FOR THIS LOCK TO BE GRANTED:
# 等待获取的表空间 ID 为 26，页号为 4，也就是表 hero 的P RIMAY 索引中的某条记录的锁
# 该锁的类型是 X 锁（rec but not gap）
RECORD LOCKS space id 26 page no 4 n bits 72 index PRIMARY of table `asdb`.`hero` trx id 121629 lock_mode X locks rec but not gap waiting
# 记录在页面中的 heap_no 为 3，具体的记录信息如下：
Record lock, heap no 3 PHYSICAL RECORD: n_fields 5; compact format; info bits 0
# 这是主键值
 0: len 4; hex 80000003; asc     ;;
# 这是 trx_id 隐藏列 
 1: len 6; hex 00000001daf9; asc       ;;
# 这是 roll_pointer 隐藏列 
 2: len 7; hex 81000000a1011d; asc        ;;
# 这是 name 列 
 3: len 10; hex 7ae8afb8e8919be4baae; asc z         ;;
# 这是 country 列 
 4: len 3; hex e89c80; asc    ;;
```

从这个信息中可以看出，Session 1 中的事务为 2 条记录生成了锁结构，但是其中有一条记录上的 X 锁（rec but not gap）并没有获取到，没有获取到锁的这条记录主键值为 `80000003`，这其实是 InnoDB 内部存储使用的格式，其实就代表数字 `3`，也就是该事务在等待获取 `hero` 表聚簇索引主键值为 `3` 的那条记录的 X 锁。

然后是关于死锁发生时第二个事务的有关信息：

```yaml
*** (2) TRANSACTION:
TRANSACTION 121630, ACTIVE 5 sec starting index read
mysql tables in use 1, locked 1
LOCK WAIT 3 lock struct(s), heap size 1128, 2 row lock(s)
MySQL thread id 164, OS thread handle 9416, query id 332330 localhost ::1 root statistics
select * from hero where id=1 for update

# 表示该事务获取到的锁信息
*** (2) HOLDS THE LOCK(S):
RECORD LOCKS space id 26 page no 4 n bits 72 index PRIMARY of table `asdb`.`hero` trx id 121630 lock_mode X locks rec but not gap
Record lock, heap no 3 PHYSICAL RECORD: n_fields 5; compact format; info bits 0
# 主键值为 3
 0: len 4; hex 80000003; asc     ;;
 1: len 6; hex 00000001daf9; asc       ;;
 2: len 7; hex 81000000a1011d; asc        ;;
 3: len 10; hex 7ae8afb8e8919be4baae; asc z         ;;
 4: len 3; hex e89c80; asc    ;;

*** (2) WAITING FOR THIS LOCK TO BE GRANTED:
RECORD LOCKS space id 26 page no 4 n bits 72 index PRIMARY of table `asdb`.`hero` trx id 121630 lock_mode X locks rec but not gap waiting
Record lock, heap no 2 PHYSICAL RECORD: n_fields 5; compact format; info bits 0
# 主键值为 1
 0: len 4; hex 80000001; asc     ;;
 1: len 6; hex 00000001daf9; asc       ;;
 2: len 7; hex 81000000a10110; asc        ;;
 3: len 7; hex 6ce58898e5a487; asc l      ;;
 4: len 3; hex e89c80; asc    ;;

# 回滚事务
*** WE ROLL BACK TRANSACTION (2)
```

Session 2 中的事务获取了 `hero` 表聚簇索引主键值为 `3` 的记录的 X 锁，等待获取 `hero` 表聚簇索引主键值为 `1` 的记录的 X 锁。

### 分析的思路

1. 首先看一下发生死锁的事务等待获取锁的语句是什么。
2. 找到发生死锁的事务中所有的语句之后，对照着事务获取到的锁和正在等待的锁的信息来分析死锁发生过程。

## 死锁检测和锁超时

处理死锁的方式有两种：

- 死锁检测
- 锁超时

### 死锁检测

```yaml
innodb_deadlock_detect=ON
```

`innodb_deadlock_detect` 的默认值是 `on`，表示开启死锁检测。死锁检测能够在发生死锁的时候，快速发现并进行处理，但是它也是有额外负担的。

可以想象一下这个过程：每当一个事务被锁的时候，就要看看它所依赖的线程有没有被别人锁住，如此循环，最后判断是否出现了循环等待，也就是死锁。

每个新来的被堵住的线程，都要判断会不会由于自己的加入导致了死锁，这是一个时间复杂度是 `O(n)` 的操作。假设有 1000 个并发线程要同时更新同一行，那么死锁检测操作就是 100 万这个量级的。虽然最终检测的结果是没有死锁，但是这期间要消耗大量的 CPU 资源。因此，你就会看到 CPU 利用率很高，但是每秒却执行不了几个事务。

怎么解决由这种热点行更新导致的性能问题？问题的症结在于，死锁检测要耗费大量的 CPU 资源。

一种就是**如果你能确保这个业务一定不会出现死锁，可以临时把死锁检测关掉**。但是这种操作本身带有一定的风险，因为业务设计的时候一般不会把死锁当做一个严重错误，毕竟出现死锁了，就回滚，然后通过业务重试一般就没问题了，这是业务无损的。而关掉死锁检测意味着可能会出现大量的超时，这是业务有损的。

另一个思路是**控制并发度**。根据上面的分析，如果并发能够控制住，比如同一行同时最多只有 10 个线程在更新，那么死锁检测的成本很低，就不会出现这个问题。一个直接的想法就是，在客户端做并发控制。但是，你会很快发现这个方法不太可行，因为客户端很多。我见过一个应用，有 600 个客户端，这样即使每个客户端控制到只有 5 个并发线程，汇总到数据库服务端以后，峰值并发数也可能要达到 3000。

可以考虑通过将一行改成逻辑上的多行来减少锁冲突。以一个影院账户为例，可以考虑放在多条记录上，比如 10 个记录，影院的账户总额等于这 10 个记录的值的总和。这样每次要给影院账户加金额的时候，随机选其中一条记录来加。这样每次冲突概率变成原来的 `1/10`，可以减少锁等待个数，也就减少了死锁检测的 CPU 消耗。

### 锁超时

行锁的超时时间可以通过参数 `innodb_lock_wait_timeout` 来设置。

`innodb_lock_wait_timeout` 的默认值是 `50s`。当出现死锁以后，第一个被锁住的线程要过 `50s` 才会超时退出，然后其他线程才有可能继续执行。对于在线服务来说，这个等待时间往往是无法接受的。

但是，又不可能直接把这个时间设置成一个很小的值，比如 `1s`。这样当出现死锁的时候，确实很快就可以解开，但如果不是死锁，而是简单的锁等待呢？所以，超时时间设置太短的话，会出现很多误伤。要根据业务来设置一个合理的值。


## 锁优化实践

- 尽可能让所有数据检索都通过索引来完成，**避免无索引行锁升级为表锁**，
- 合理设计索引，尽量缩小锁的范围。
- 尽可能**减少检索条件范围**，避免间隙锁。
- 尽量**控制事务大小**，减少锁定资源量和时间长度，涉及事务加锁的 sql 尽量放在事务最后执行。
- 尽可能用低的事务隔离级别。