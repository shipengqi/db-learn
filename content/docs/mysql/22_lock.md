---
title: 锁
---

## 并发事务

并发事务访问相同记录的情况大致可以划分为3种：

- 读-读情况：即并发事务相继读取相同的记录。不会引起什么问题，所以允许这种情况的发生。
- 写-写情况：即并发事务相继对相同的记录做出改动。脏写的问题，任何一种隔离级别都不允许这种问题的发生。所以在多个未提交事务相继对一条记录做改动时，需要让它们排队执行，这个排队的过程其实是通过锁来实现的。
- 读-写或写-读情况：也就是一个事务进行读取操作，另一个进行改动操作。这种情况下可能发生脏读、不可重复读、幻读的问题。

方案一：读操作利用**多版本并发控制**（MVCC），写操作进行**加锁**。

ReadView找到符合条件的记录版本（历史版本是由undo日志构建的），其实就像是在生成ReadView的那个时刻做了一次时间静止（就像用相机拍了一个快照），查询语句只能读到在生成ReadView之前已提交事务所做的更改，在生成ReadView之前未提交的事务或者之后才开启的事务所做的更改是看不到的。而写操作肯定针对的是最新版本的记录，读记录的历史版本和改动记录的最新版本本身并不冲突，也就是采用MVCC时，读-写操作并不冲突。

方案二：读、写操作都采用加锁的方式。

如果我们的一些业务场景不允许读取记录的旧版本，而是每次都必须去读取记录的最新版本，比方在银行存款的事务中，你需要先把账户的余额读出来，然后将其加上本次存款的数额，最后再写到数据库中。在将账户余额读取出来后，就不想让别的事务再访问该余额，直到本次存款事务执行完成，其他事务才可以访问账户的余额。这样在读取记录的时候也就需要对其进行加锁操作，这样也就意味着**读操作和写操作也像写-写操作那样排队执行**。

采用MVCC方式的话，读-写操作彼此并不冲突，性能更高，采用加锁方式的话，读-写操作彼此需要排队执行，影响性能。一般情况下我们当然愿意采用MVCC来解决读-写操作并发执行的问题，但是业务在某些特殊情况下，要求必须采用加锁的方式执行，那也是没有办法的事。

### 一致性读（Consistent Reads）

事务利用MVCC进行的读取操作称之为**一致性读**，也称之为**快照读**。

### 锁定读（Locking Reads）

#### 共享锁和独占锁

- **共享锁**，英文名：Shared Locks，简称S锁。在事务要读取一条记录时，需要先获取该记录的S锁。
- **独占锁**，也常称**排他锁**，英文名：Exclusive Locks，简称X锁。在事务要改动一条记录时，需要先获取该记录的X锁。

假如事务T1首先获取了一条记录的S锁之后，事务T2接着也要访问这条记录：

- 如果事务T2想要再获取一个记录的S锁，那么事务T2也会获得该锁，也就意味着事务T1和T2在该记录上同时持有S锁。
- 如果事务T2想要再获取一个记录的X锁，那么此操作会被阻塞，直到事务T1提交之后将S锁释放掉。

如果事务T1首先获取了一条记录的X锁之后，那么不管事务T2接着想获取该记录的S锁还是X锁都会被阻塞，直到事务T1提交。

#### 锁定读的语句

采用加锁方式解决脏读、不可重复读、幻读这些问题时，读取一条记录时需要获取一下该记录的S锁，其实这是不严谨的，有时候想在读取记录时就获取记录的X锁，来禁止别的事务读写该记录，为此设计MySQL的大叔提出了两种比较特殊的SELECT语句格式：

- 对读取的记录加S锁：

```sql
SELECT ... LOCK IN SHARE MODE;
```

- 对读取的记录加X锁：

```sql
SELECT ... FOR UPDATE;
```

也就是在普通的SELECT语句后边加FOR UPDATE，如果当前事务执行了该语句，那么它会为读取到的记录加X锁，这样既不允许别的事务获取这些记录的S锁（比方说别的事务使用`SELECT ... LOCK IN SHARE MODE`语句来读取这些记录），也不允许获取这些记录的X锁

### 多粒度锁

**提高共享资源并发性的方式就是让锁定对象更有选择性。尽量只锁定需要修改的部分数据，而不是多有的资源**。只对会修改的数据进行精确的锁定。在给定的资源上，锁定的数据量越少，则系统的并发程度越高，只要不发生冲突即可。

针对记录的锁，也可以被称之为**行级锁**或者**行锁**，对一条记录加锁影响的也只是这条记录而已，我们就说这个锁的粒度比较细；其实一个事务也可以在表级别进行加锁，自然就被称之为**表级锁**或者**表锁**。

**IS、IX锁是表级锁，它们的提出仅仅为了在之后加表级别的S锁和X锁时可以快速判断表中的记录是否被上锁，以避免用遍历的方式来查看表中有没有上锁的记录，也就是说其实IS锁和IX锁是兼容的，IX锁和IX锁是兼容的**。

## MySQL中的行锁和表锁

对于MyISAM、MEMORY、MERGE这些存储引擎来说，它们只支持表级锁，而且这些引擎并不支持事务，所以使用这些存储引擎的锁一般都是针对当前会话来说的。

**因为使用MyISAM、MEMORY、MERGE这些存储引擎的表在同一时刻只允许一个会话对表进行写操作，所以这些存储引擎实际上最好用在只读，或者大部分都是读操作，或者单用户的情景下**。

### InnoDB中的表级锁

InnoDB存储引擎提供的表级S锁或者X锁是相当鸡肋，只会在一些特殊情况下，比方说崩溃恢复过程中用到。不过我们还是可以手动获取一下的，比方说在系统变量autocommit=0，innodb_table_locks = 1时，手动获取InnoDB存储引擎提供的表t的S锁或者X锁可以这么写：

LOCK TABLES t READ：InnoDB存储引擎会对表t加表级别的S锁。

LOCK TABLES t WRITE：InnoDB存储引擎会对表t加表级别的X锁。

不过请尽量避免在使用InnoDB存储引擎的表上使用LOCK TABLES这样的手动锁表语句

表级别的IS锁、IX锁

当我们在对使用InnoDB存储引擎的表的某些记录加S锁之前，那就需要先在表级别加一个IS锁，当我们在对使用InnoDB存储引擎的表的某些记录加X锁之前，那就需要先在表级别加一个IX锁。IS锁和IX锁的使命只是为了后续在加表级别的S锁和X锁时判断表中是否有已经被加锁的记录，以避免用遍历的方式来查看表中有没有上锁的记录。

表级别的**AUTO-INC锁**

在使用MySQL过程中，我们可以为表的某个列添加AUTO_INCREMENT属性，之后在插入记录时，可以不指定该列的值，系统会自动为它赋上递增的值

系统实现这种自动给AUTO_INCREMENT修饰的列递增赋值的原理主要是两个：

- 采用AUTO-INC锁，也就是在执行插入语句时就在表级别加一个AUTO-INC锁，然后为每条待插入记录的AUTO_INCREMENT修饰的列分配递增的值，在**该语句执行结束后，再把AUTO-INC锁释放掉**。这样一个事务在持有AUTO-INC锁的过程中，其他事务的插入语句都要被阻塞，可以保证一个语句中分配的递增值是连续的。
- 采用一个**轻量级的锁**，在**为插入语句生成AUTO_INCREMENT修饰的列的值时获取一下这个轻量级锁，然后生成本次插入语句需要用到的AUTO_INCREMENT列的值之后，就把该轻量级锁释放掉**，并不需要等到整个插入语句执行完才释放锁。

当innodb_autoinc_lock_mode值为0时，一律采用AUTO-INC锁；当innodb_autoinc_lock_mode值为2时，一律采用轻量级锁；当innodb_autoinc_lock_mode值为1时，两种方式混着来（也就是在插入记录数量确定时采用轻量级锁，不确定时使用AUTO-INC锁）。不过当innodb_autoinc_lock_mode值为2时，可能会造成不同事务中的插入语句为AUTO_INCREMENT修饰的列生成的值是交叉的，在有主从复制的场景中是不安全的。

### InnoDB中的行级锁

行锁，也称为记录锁，顾名思义就是在记录上加的锁。InnoDB行锁分成了各种类型。

- Record Locks，正经记录锁，当一个事务获取了一条记录的S型正经记录锁后，其他事务也可以继续获取该记录的S型正经记录锁，但不可以继续获取X型正经记录锁；当一个事务获取了一条记录的X型正经记录锁后，其他事务既不可以继续获取该记录的S型正经记录锁，也不可以继续获取X型正经记录锁
- Gap Locks，MySQL在REPEATABLE READ隔离级别下是可以解决幻读问题的，解决方案有两种，可以使用MVCC方案解决，也可以采用加锁方案解决。但是在使用加锁方案解决时有个大问题，就是事务在第一次执行读取操作时，那些幻影记录尚不存在，我们无法给这些幻影记录加上正经记录锁。不过这难不倒设计InnoDB的大叔，他们提出了一种称之为Gap Locks的锁，官方的类型名称为：LOCK_GAP，我们也可以简称为gap锁。比方说我们把number值为8的那条记录加一个gap锁的示意图如下：

![](../../../images/lock-gap-locks.jpg)

如图中为number值为8的记录加了gap锁，意味着不允许别的事务在number值为8的记录前边的间隙插入新记录，其实就是number列的值(3, 8)这个区间的新记录是不允许立即插入的。比方说有另外一个事务再想插入一条number值为4的新记录，它定位到该条新记录的下一条记录的number值为8，而这条记录上又有一个gap锁，所以就会阻塞插入操作，直到拥有这个gap锁的事务提交了之后，number列的值在区间(3, 8)中的新记录才可以被插入。

这个**gap锁的提出仅仅是为了防止插入幻影记录而提出的**

- Next-Key Locks，本质就是一个正经记录锁和一个gap锁的合体，它既能保护该条记录，又能阻止别的事务将新记录插入被保护记录前边的间隙。
- Insert Intention Locks，我们说一个事务在插入一条记录时需要判断一下插入位置是不是被别的事务加了所谓的gap锁（next-key锁也包含gap锁，后边就不强调了），如果有的话，插入操作需要等待，直到拥有gap锁的那个事务提交。但是设计InnoDB的大叔规定事务在等待的时候也需要在内存中生成一个锁结构，表明有事务想在某个间隙中插入新记录，但是现在在等待。设计InnoDB的大叔就把这种类型的锁命名为Insert Intention Locks，官方的类型名称为：LOCK_INSERT_INTENTION，我们也可以称为插入意向锁。
**插入意向锁并不会阻止别的事务继续获取该记录上任何类型的锁**

- 隐式锁
一个事务在执行INSERT操作时，如果即将插入的间隙已经被其他事务加了gap锁，那么本次INSERT操作会阻塞，并且当前事务会在该间隙上加一个插入意向锁，否则一般情况下INSERT操作是不加锁的。

如果一个事务首先插入了一条记录（此时并没有与该记录关联的锁结构），然后另一个事务：

立即使用SELECT ... LOCK IN SHARE MODE语句读取这条记录，也就是在要获取这条记录的S锁，或者使用SELECT ... FOR UPDATE语句读取这条记录，也就是要获取这条记录的X锁，该咋办？

如果允许这种情况的发生，那么可能产生脏读问题。

立即修改这条记录，也就是要获取这条记录的X锁，该咋办？

如果允许这种情况的发生，那么可能产生脏写问题。

情景一：对于聚簇索引记录来说，有一个trx_id隐藏列，该隐藏列记录着最后改动该记录的事务id。那么如果在当前事务中新插入一条聚簇索引记录后，该记录的trx_id隐藏列代表的的就是当前事务的事务id，如果其他事务此时想对该记录添加S锁或者X锁时，首先会看一下该记录的trx_id隐藏列代表的事务是否是当前的活跃事务，如果是的话，那么就帮助当前事务创建一个X锁（也就是为当前事务创建一个锁结构，is_waiting属性是false），然后自己进入等待状态（也就是为自己也创建一个锁结构，is_waiting属性是true）。

情景二：对于二级索引记录来说，本身并没有trx_id隐藏列，但是在二级索引页面的Page Header部分有一个PAGE_MAX_TRX_ID属性，该属性代表对该页面做改动的最大的事务id，如果PAGE_MAX_TRX_ID属性值小于当前最小的活跃事务id，那么说明对该页面做修改的事务都已经提交了，否则就需要在页面中定位到对应的二级索引记录，然后回表找到它对应的聚簇索引记录，然后再重复情景一的做法。

一个事务对新插入的记录可以不显式的加锁（生成一个锁结构），但是由于事务id这个牛逼的东东的存在，相当于加了一个隐式锁。别的事务在对这条记录加S锁或者X锁时，由于隐式锁的存在，会先帮助当前事务生成一个锁结构，然后自己再生成一个锁结构后进入等待状态。

## 两阶段锁

在InnoDB事务中，**行锁是在需要的时候才加上的，但并不是不需要了就立刻释放，而是要等到事务结束时才释放。这个就是两阶段锁协议**。

**如果你的事务中需要锁多个行，要把最可能造成锁冲突、最可能影响并发度的锁尽量往后放**。

假设你负责实现一个电影票在线交易业务，顾客A要在影院B购买电影票。我们简化一点，这个业务需要涉及到以下操作：

从顾客A账户余额中扣除电影票价；

给影院B的账户余额增加这张电影票价；

记录一条交易日志。

也就是说，要完成这个交易，我们需要update两条记录，并insert一条记录。当然，为了保证交易的原子性，我们要把这三个操作放在一个事务中。那么，你会怎样安排这三个语句在事务中的顺序呢？

试想如果同时有另外一个顾客C要在影院B买票，那么这两个事务冲突的部分就是语句2了。因为它们要更新同一个影院账户的余额，需要修改同一行数据。

根据两阶段锁协议，不论你怎样安排语句顺序，所有的操作需要的行锁都是在事务提交的时候才释放的。所以，如果你把语句2安排在最后，比如按照3、1、2这样的顺序，那么影院账户余额这一行的锁时间就最少。这就最大程度地减少了事务之间的锁等待，提升了并发度。

## 死锁和死锁检测

![](../../../images/dead-lock.jpg)

事务A在等待事务B释放id=2的行锁，而事务B在等待事务A释放id=1的行锁。 事务A和事务B在互相等待对方的资源释放，就是进入了死锁状态。当出现死锁以后，有两种策略：

1. 一种策略是，**直接进入等待，直到超时。这个超时时间可以通过参数innodb_lock_wait_timeout来设置**。
2. 另一种策略是，**发起死锁检测，发现死锁后，主动回滚死锁链条中的某一个事务，让其他事务得以继续执行**。将参数innodb_deadlock_detect设置为on，表示开启这个逻辑。

innodb_lock_wait_timeout的默认值是50s，意味着如果采用第一个策略，当出现死锁以后，第一个被锁住的线程要过50s才会超时退出，然后其他线程才有可能继续执行。对于在线服务来说，这个等待时间往往是无法接受的。

但是，我们又不可能直接把这个时间设置成一个很小的值，比如1s。这样当出现死锁的时候，确实很快就可以解开，但如果不是死锁，而是简单的锁等待呢？所以，超时时间设置太短的话，会出现很多误伤。

正常情况下我们还是要采用第二种策略，即：**主动死锁检测**，而且innodb_deadlock_detect的默认值本身就是on。主动死锁检测在发生死锁的时候，是能够快速发现并进行处理的，但是它也是有额外负担的。

你可以想象一下这个过程：每当一个事务被锁的时候，就要看看它所依赖的线程有没有被别人锁住，如此循环，最后判断是否出现了循环等待，也就是死锁。

那如果是我们上面说到的所有事务都要更新同一行的场景呢？

每个新来的被堵住的线程，都要判断会不会由于自己的加入导致了死锁，这是一个时间复杂度是O(n)的操作。假设有1000个并发线程要同时更新同一行，那么死锁检测操作就是100万这个量级的。虽然最终检测的结果是没有死锁，但是这期间要消耗大量的CPU资源。因此，你就会看到CPU利用率很高，但是每秒却执行不了几个事务。

怎么解决由这种热点行更新导致的性能问题呢？问题的症结在于，死锁检测要耗费大量的CPU资源。

一种就是**如果你能确保这个业务一定不会出现死锁，可以临时把死锁检测关掉**。但是这种操作本身带有一定的风险，因为业务设计的时候一般不会把死锁当做一个严重错误，毕竟出现死锁了，就回滚，然后通过业务重试一般就没问题了，这是业务无损的。而关掉死锁检测意味着可能会出现大量的超时，这是业务有损的。

另一个思路是**控制并发度**。根据上面的分析，你会发现如果并发能够控制住，比如同一行同时最多只有10个线程在更新，那么死锁检测的成本很低，就不会出现这个问题。一个直接的想法就是，在客户端做并发控制。但是，你会很快发现这个方法不太可行，因为客户端很多。我见过一个应用，有600个客户端，这样即使每个客户端控制到只有5个并发线程，汇总到数据库服务端以后，峰值并发数也可能要达到3000。

因此，这个并发控制要做在数据库服务端。如果你有中间件，可以考虑在中间件实现；如果你的团队有能修改MySQL源码的人，也可以做在MySQL里面。基本思路就是，**对于相同行的更新，在进入引擎之前排队。这样在InnoDB内部就不会有大量的死锁检测工作了**。

你可以考虑通过将一行改成逻辑上的多行来减少锁冲突。还是以影院账户为例，可以考虑放在多条记录上，比如 10 个记录，影院的账户总额等于这 10 个记录的值的总和。这样每次要给影院账户加金额的时候，随机选其中一条记录来加。这样每次冲突概率变成原来的 1/10，可以减少锁等待个数，也就减少了死锁检测的 CPU 消耗。
