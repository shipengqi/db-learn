---
title: MySQL 日志机制
weight: 7
---

<div class="img-zoom">
  <img src="https://raw.gitcode.com/shipengqi/illustrations/files/main/db/arch.png" alt="arch">
</div>

## redo log

redo log 只是为了系统崩溃后恢复脏页用的，先写 redo log，再写数据库表文件的机制，也被称之为 `WAL（Write Ahead Logging）`。

**WAL**：对数据库做任何修改之前，必须先将对应的修改操作记录到日志文件中，并保证日志已经持久化到磁盘，才能真正对数据页进行修改。

WAL 的好处：

- 保证数据的**持久性**，只要日志已经写入磁盘，重启后可以通过日志重放恢复数据。
- 日志是**顺序 I/O**（追加写入），比随机 I/O 快得多（尤其是机械硬盘时代）。
  - 先顺序 I/O 写入日志文件，数据页可后续批量写入数据表文件，减少 I/O 频率（innodb 中是以页为单位来进行磁盘 IO 的，一个页面默认是 16KB，也就是说在该事务提交时不得不将一个完整的页面从内存中刷新到磁盘，哪怕只修改一个字节也要刷新 16KB 的数据到磁盘上）。

### redo log 关键配置

#### innodb_log_buffer_size

设置 redo log buffer 大小参数，默认 `16M`，最大值是 `4096M`，最小值为 `1M`。

```sql
show variables like '%innodb_log_buffer_size%';
```

#### innodb_log_group_home_dir

设置 redo log 文件存储位置，默认值为 `./`，即 innodb 数据文件存储位置，其中的 `ib_logfile0` 和 `ib_logfile1` 即为 redo log 文件。

```sql
show variables like '%innodb_log_group_home_dir%';
```

#### innodb_log_files_in_group

磁盘上的 redo log 文件不只一个，而是以一个日志文件组的形式出现的。`innodb_log_files_in_group` 设置 redo log 文件组中文件个数，默认值为 2，命名方式如: `ib_logfile0`, `iblogfile1` ... `iblogfileN`。


```sql
show variables like '%innodb_log_files_in_group%';
```

##### redo log 写入磁盘过程

在将 redo log 写入日志文件组时，是从 `ib_logfile0` 开始写，如果 `ib_logfile0` 写满了，就接着 `ib_logfile1` 写，同理，`ib_logfile1` 写满了就去写 `ib_logfile2`，依此类推。如果写到最后一个文件该咋办？那就重新转到 `ib_logfile0` 继续写。类似环形的写入，假设现在有四个 redo log 文件，如下图，

![redo-log-write](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redo-log-write.png)

- **write pos** 是当前记录的位置，一边写一边后移，写到第 3 号文件末尾后就回到 0 号文件开头。
- **checkpoint** 是当前要擦除的位置，也是往后推移并且循环的，擦除记录前要把记录更新到**数据文件**。

write pos 和 checkpoint 之间的部分就是空着的**可写部分**，可以用来记录新的操作。如果 write pos 追上 checkpoint，表示 redo log 写满了，这时候不能再执行新的更新，得停下来先擦掉一些记录，把 checkpoint 推进一下。

#### innodb_log_file_size

设置 redo log 文件大小，默认值为 `48M`，最大值为 `512G`，注意**最大值指的是所有 redo log 文件之和**，即 `innodb_log_files_in_group * innodb_log_file_size` 不能大于最大值 `512G`。

```sql
show variables like '%innodb_log_file_size%';
```

#### innodb_flush_log_at_trx_commit

这个参数控制 redo log 的写入策略，可选值：

- `0`：表示每次事务提交时都只是把 redo log 留在 redo log buffer 中，效率最高，但是如果数据库宕机可能会丢失数据。
- `1`：默认值，表示每次事务提交时都将 redo log 直接持久化到磁盘，数据最安全，不会因为数据库宕机丢失数据，但是效率稍微差一点，线上系统推荐这个设置。
- `2`：这是一个折中的选择，表示每次事务提交时都只是把 redo log 写到操作系统的缓存 Page Cache 里，这种情况如果数据库宕机是不会丢失数据的，但是操作系统如果宕机了，Page Cache 里的数据还没来得及写入磁盘文件的话就会丢失数据。

![redo-log-policy](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redo-log-policy.png)

查看 `innodb_flush_log_at_trx_commit` 的值：

```sql
show variables like 'innodb_flush_log_at_trx_commit';
```

设置 `innodb_flush_log_at_trx_commit` 的值（也可以在 `my.ini` 或 `my.cnf` 文件里配置）：

```sql
set global innodb_flush_log_at_trx_commit = 1;
```

#### redo 日志刷盘时机

`innodb_flush_log_at_trx_commit` 是控制**事务提交**时 redo 日志写入磁盘的策略，还有其他的一些时机也会触发 redo 日志写入磁盘：

- redo log buffer 空间不足时。redo log buffer 的大小是有限的（通过系统变量 `innodb_log_buffer_size` 指定），如果不停的往这个 buffer 里塞入日志，很快它就会被填满。如果当前写入的 redo log 已经占满了 buffer 总容量的**大约一半左右**，就会把日志刷新到磁盘上。
- 事务提交时。
- 后台线程不停的刷刷刷。后台有一个线程，大约每秒都会刷新一次 redo log buffer 中的 redo 日志到磁盘。
- 正常关闭服务器时。


## binlog

MySQL 整体来看，其实就有两块：一块是 Server 层，它主要做的是 MySQL 功能层面的事情；还有一块是引擎层。

**redo log 是 InnoDB 引擎特有的日志**，而 **Server 层也有自己的日志，称为 binlog**（归档日志）。

binlog **二进制**日志记录**保存了所有执行过的修改操作语句，不保存查询操作**。

启动 binlog 记录功能，会影响服务器性能，但如果需要恢复数据或主从复制功能，则好处则大于对服务器性能的影响。

MySQL 5.7 版本中，binlog 默认是关闭的，8.0 版本默认是打开的。

开启 binlog 记录功能，需要修改配置文件 `my.ini(windows)` 或 `my.cnf(linux)`，然后重启数据库。

```ini
[mysqld]
...
# 设置 binlog 的存放位置，可以是绝对路径，也可以是相对路径，这里写的相对路径，则 binlog 文件默认会放在 data 数据目录下
log_bin = mysql‐binlog
# Server Id 是数据库服务器 id，随便写一个数都可以，这个 id 用来在 MySQL 集群环境中标记唯一 MySQL 服务器，集群环境中每台 MySQL 服务器的 id 不能一样，否则启动会报错
server‐id=1
# 记录 binlog 的格式，有三种格式：STATEMENT、ROW、MIXED，默认是 STATEMENT 格式
binlog_format = ROW
# 执行自动删除 binlog 日志文件的天数，也就是每个 binlog 文件最多保存的时间。默认为 0，表示不自动删除。一般情况下，需要根据业务设置一个合理的值，这样可以保证 binlog 日志文件不会占用太多的磁盘空间。
expire_logs_days = 7
# 单个 binlog 日志文件的大小限制，默认为 1GB
max_binlog_size = 200M 
```

查看 binlog 配置：

```sql
show variables like '%log_bin%';
```

![log-bin-variables](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/log-bin-variables.png)

- `log_bin`：binlog 日志是否打开状态
- `log_bin_basename`：是 binlog 日志的基本文件名，后面会追加标识来表示每一个文件，binlog 日志文件会滚动增加
- `log_bin_index`：指定的是 binlog 文件的索引文件，这个文件管理了所有的 binlog 文件的目录。
- `sql_log_bin`：SQL 语句是否写入 binlog 文件，`ON` 代表需要写入，`OFF` 代表不需要写入。

**`sql_log_bin` 有什么用？**

既然开启了 binlog 日志，那为什么还要设置 `sql_log_bin` 呢？

`sql_log_bin` 如果修改为 `OFF`，则代表所有的 SQL 语句都不会写入 binlog 文件，但是 binlog 文件还是有的。这样就可以在主库上执行一些操作，但不复制到 slave 库上。可以用来模拟主从同步复制异常。

### binlog 的日志格式

`binlog_format` 可以设置 binlog 日志的记录格式，MySQL 支持三种格式类型：

- `STATEMENT`：基于 SQL 语句的复制，每一条会修改数据的 SQL 都会记录到 master 机器的 binlog 中，这种方式日志量小，节约 IO 开销，提高性能，但是对于一些执行过程中才能确定结果的函数，比如 `UUID()`、`SYSDATE()` 等函数如果随 SQL 同步到 slave 机器去执行，则结果跟 master 机器执行的不一样。
- `ROW`：基于行的复制，日志中会记录成每一行数据被修改的形式，然后在 slave 端再对相同的数据进行修改记录下每一行数据修改的细节，可以解决函数、存储过程等在 slave 机器的复制问题，但这种方式日志量较大，性能不如 `STATEMENT`。举个例子，假设 `update` 语句更新 `10` 行数据，`STATEMENT` 方式就记录这条 `update` 语句，而 `Row` 方式会记录被修改的 10 行数据。
- `MIXED`：混合模式复制，实际就是前两种模式的结合，在 `MIXED` 模式下，MySQL 会根据执行的每一条具体的 SQL 语句来区分对待记录的日志形式，也就是在 `STATEMENT` 和 `ROW` 之间选择一种，如果 SQL 里有函数或一些在执行时才知道结果的情况，会选择 `ROW`，其它情况选择 `STATEMENT`，推荐使用这一种。

### binlog 的写入机制

binlog 写入磁盘机制主要通过 `sync_binlog` 参数控制。与 redo log 类似，`sync_binlog` 也有三种取值：

- 设置为 `0`，默认值，表示每次提交事务都只 write 到 Page Cache，由系统自行判断什么时候执行 `fsync` 写入磁盘。虽然性能得到提升，但是机器宕机，Page Cache 里面的 binlog 会丢失。
- 设置为 `1`，表示每次提交事务都会执行 `fsync` 写入磁盘，这种方式最安全。
- 设置为 `N` （`N>1`），表示每次提交事务都 write 到 Page Cache，但累积 `N` 个事务后才 `fsync` 写入磁盘，这种如果机器宕机会丢失 `N` 个事务的 binlog。

binlog 日志文件重新生成的时机：

- 服务器启动或重新启动。
- 服务器刷新日志，执行命令 `flush logs`。
- 日志文件大小达到 `max_binlog_size` 值，默认值为 `1GB`。

### 删除 binlog 日志文件

```sql
-- 删除当前的 binlog 文件
reset master;
-- 删除指定日志文件之前的所有日志文件，下面这个是删除 6 之前的所有日志文件，当前这个文件不删除
purge master logs to 'mysql‐binlog.000006';
-- 删除指定日期前的日志索引中binlog日志文件
purge master logs before '2023‐01‐21 14:00:00';
```

### 查看 binlog 日志文件

MySQL 提供了 `mysqlbinlog` 命令，可以查看 binlog 日志：

```bash
# 查看 binlog（命令行方式，不用登录 MySQL）
$ mysqlbinlog ‐‐no‐defaults ‐v \
‐‐base64‐output=decode‐rows D:/dev/mysql‐5.7.25‐winx64/data/mysql‐binlog.000007

# 查看 binlog（带查询条件）
$ mysqlbinlog ‐‐no‐defaults ‐v \
‐‐base64‐output=decode‐rows D:/dev/mysql‐5.7.25‐winx64/data/mysql‐binlog.000007 \
start‐datetime="2023‐01‐21 00:00:00" \
stop‐datetime="2023‐02‐01 00:00:00" \
startposition="5000" \
stop‐position="20000"
```

查出来的 binlog 日志文件内容如下：

```binlog
/*!50530 SET @@SESSION.PSEUDO_SLAVE_MODE=1*/;
/*!50003 SET @OLD_COMPLETION_TYPE=@@COMPLETION_TYPE,COMPLETION_TYPE=0*/;
DELIMITER /*!*/;
# at 4
#230127 21:13:51 server id 1 end_log_pos 123 CRC32 0x084f390f Start: binlog v 4, server v 5.7.25‐
log created 230127 21:13:51 at startup
# Warning: this binlog is either in use or was not closed properly.
ROLLBACK/*!*/;
# at 123
#230127 21:13:51 server id 1 end_log_pos 154 CRC32 0x672ba207 Previous‐GTIDs
# [empty]
# at 154
#230127 21:22:48 server id 1 end_log_pos 219 CRC32 0x8349d010 Anonymous_GTID last_committed=0 sequence_number=1 rbr_only=yes
/*!50718 SET TRANSACTION ISOLATION LEVEL READ COMMITTED*//*!*/;
SET @@SESSION.GTID_NEXT= 'ANONYMOUS'/*!*/;
# at 219
#230127 21:22:48 server id 1 end_log_pos 291 CRC32 0xbf49de02 Query thread_id=3 exec_time=0 erro
r_code=0
SET TIMESTAMP=1674825768/*!*/;
SET @@session.pseudo_thread_id=3/*!*/;
SET @@session.foreign_key_checks=1, @@session.sql_auto_is_null=0, @@session.unique_checks=1, @@s
ession.autocommit=1/*!*/;
SET @@session.sql_mode=1342177280/*!*/;
SET @@session.auto_increment_increment=1, @@session.auto_increment_offset=1/*!*/;
/*!\C utf8 *//*!*/;
SET @@session.lc_time_names=0/*!*/;
SET @@session.collation_database=DEFAULT/*!*/;
BEGIN
/*!*/;
# at 291
#230127 21:22:48 server id 1 end_log_pos 345 CRC32 0xc4ab653e Table_map: `test`.`account` mapped to number 99
# at 345
#230127 21:22:48 server id 1 end_log_pos 413 CRC32 0x54a124bd Update_rows: table id 99 flags: ST
MT_END_F
### UPDATE `test`.`account`
### WHERE
### @1=1
### @2='lilei'
### @3=1000
### SET
### @1=1
### @2='lilei'
### @3=2000
# at 413
#230127 21:22:48 server id 1 end_log_pos 444 CRC32 0x23355595 Xid = 10
COMMIT/*!*/;
# at 444
...
```

包含具体执行的 SQL 语句以及执行时的相关情况。

### binlog 日志文件恢复数据

```sql
-- 先执行刷新日志的命令生成一个新的 binlog 文件 mysql‐binlog.000008，后面修改操作日志都会记录在最新的这个文件里
flush logs;
-- 执行两条插入语句
INSERT INTO `test`.`account` (`id`, `name`, `balance`) VALUES ('4', 'zhuge', '666');
INSERT INTO `test`.`account` (`id`, `name`, `balance`) VALUES ('5', 'zhuge1', '888');
-- 假设现在误操作执行了一条删除语句把刚新增的两条数据删掉了
delete from account where id > 3;
```

现在需要恢复被删除的两条数据，先查看 binlog 日志文件：

```bash
$ mysqlbinlog ‐‐no‐defaults ‐v \
‐‐base64‐output=decode‐rows D:/dev/mysql‐5.7.25‐winx64/data/mysql‐binlog.000008
```

文件内容：

```binlog
...
SET @@SESSION.GTID_NEXT= 'ANONYMOUS'/*!*/;
# at 219
#230127 23:32:24 server id 1  end_log_pos 291 CRC32 0x4528234f  Query   thread_id=5     exec_time=0     error_code=0
SET TIMESTAMP=1674833544/*!*/;
SET @@session.pseudo_thread_id=5/*!*/;
SET @@session.foreign_key_checks=1, @@session.sql_auto_is_null=0, @@session.unique_checks=1, @@session.autocommit=1/*!*/;
SET @@session.sql_mode=1342177280/*!*/;
SET @@session.auto_increment_increment=1, @@session.auto_increment_offset=1/*!*/;
/*!\C utf8 *//*!*/;
SET @@session.character_set_client=33,@@session.collation_connection=33,@@session.collation_server=33/*!*/;
SET @@session.lc_time_names=0/*!*/;
SET @@session.collation_database=DEFAULT/*!*/;
BEGIN
/*!*/;
# at 291
#230127 23:32:24 server id 1  end_log_pos 345 CRC32 0x7482741d  Table_map: `test`.`account` mapped to number 99
# at 345
#230127 23:32:24 server id 1  end_log_pos 396 CRC32 0x5e443cf0  Write_rows: table id 99 flags: STMT_END_F
### INSERT INTO `test`.`account`
### SET
###   @1=4
###   @2='zhuge'
###   @3=666
# at 396
#230127 23:32:24 server id 1  end_log_pos 427 CRC32 0x8a0d8a3c  Xid = 56
COMMIT/*!*/;
# at 427
#230127 23:32:40 server id 1  end_log_pos 492 CRC32 0x5261a37e  Anonymous_GTID  last_committed=1        sequence_number=2       rbr_only=yes
/*!50718 SET TRANSACTION ISOLATION LEVEL READ COMMITTED*//*!*/;
SET @@SESSION.GTID_NEXT= 'ANONYMOUS'/*!*/;
# at 492
#230127 23:32:40 server id 1  end_log_pos 564 CRC32 0x01086643  Query   thread_id=5     exec_time=0     error_code=0
SET TIMESTAMP=1674833560/*!*/;
BEGIN
/*!*/;
# at 564
#230127 23:32:40 server id 1  end_log_pos 618 CRC32 0xc26b6719  Table_map: `test`.`account` mapped to number 99
# at 618
#230127 23:32:40 server id 1  end_log_pos 670 CRC32 0x8e272176  Write_rows: table id 99 flags: STMT_END_F
### INSERT INTO `test`.`account`
### SET
###   @1=5
###   @2='zhuge1'
###   @3=888
# at 670
#230127 23:32:40 server id 1  end_log_pos 701 CRC32 0xb5e63d00  Xid = 58
COMMIT/*!*/;
# at 701
#230127 23:34:23 server id 1  end_log_pos 766 CRC32 0xa0844501  Anonymous_GTID  last_committed=2        sequence_number=3       rbr_only=yes
/*!50718 SET TRANSACTION ISOLATION LEVEL READ COMMITTED*//*!*/;
SET @@SESSION.GTID_NEXT= 'ANONYMOUS'/*!*/;
# at 766
#230127 23:34:23 server id 1  end_log_pos 838 CRC32 0x687bdf88  Query   thread_id=7     exec_time=0     error_code=0
SET TIMESTAMP=1674833663/*!*/;
BEGIN
/*!*/;
# at 838
#230127 23:34:23 server id 1  end_log_pos 892 CRC32 0x4f7b7d6a  Table_map: `test`.`account` mapped to number 99
# at 892
#230127 23:34:23 server id 1  end_log_pos 960 CRC32 0xc47ac777  Delete_rows: table id 99 flags: STMT_END_F
### DELETE FROM `test`.`account`
### WHERE
###   @1=4
###   @2='zhuge'
###   @3=666
### DELETE FROM `test`.`account`
### WHERE
###   @1=5
###   @2='zhuge1'
###   @3=888
# at 960
#230127 23:34:23 server id 1  end_log_pos 991 CRC32 0x386699fe  Xid = 65
COMMIT/*!*/;
SET @@SESSION.GTID_NEXT= 'AUTOMATIC' /* added by mysqlbinlog */ /*!*/;
DELIMITER ;
# End of log file
...
```

找到两条插入数据的 SQL，每条 SQL 的上下都有 `BEGIN` 和 `COMMIT`，找到第一条 SQL `BEGIN` 前面的文件位置标识 `at 219`（这是文件的位置标识），再找到第二条 SQL `COMMIT` 后面的文件位置标识 `at 701` 。然后可以根据文件位置标识来恢复数据，执行如下 SQL：

```bash
mysqlbinlog  --no-defaults --start-position=219 --stop-position=701 --database=test D:/dev/mysql-5.7.25-winx64/data/mysql-binlog.000009 | mysql -uroot -p123456 -v test

# 补充一个根据时间来恢复数据的命令，找到第一条 SQL BEGIN 前面的时间戳标记 SET TIMESTAMP=1674833544，再找到第二条 SQL COMMIT 后面的时间戳标记 SET TIMESTAMP=1674833663，转成 datetime 格式
mysqlbinlog  --no-defaults --start-datetime="2023-1-27 23:32:24" --stop-datetime="2023-1-27 23:34:23" --database=test D:/dev/mysql-5.7.25-winx64/data/mysql-binlog.000009 | mysql -uroot -p123456 -v test
```

执行后删除数据被恢复。如果数据后面还有别的更新，那么就接着找到对应 SQL 的位置标识，然后执行恢复数据的命令即可。如果要恢复整个 binlog 日志文件，那么就不需要过滤参数，直接执行恢复数据的命令即可。


{{< callout type="info" >}}
注意，如果要恢复大量数据，假设数据库所有数据都被删除了要怎么恢复，如果数据库之前没有备份，所有的 binlog 日志都在的话，就从 binlog 第一个文件开始逐个恢复每个 binlog 文件里的数据，这种一般不太可能，因为 binlog 日志比较大，早期的 binlog 文件会定期删除的，所以一般不可能用 binlog 文件恢复整个数据库的。
{{< /callout >}}

{{< callout type="info" >}}
一般推荐的是**每天(在凌晨后)需要做一次全量数据库备份**，那么恢复数据库可以用最近的一次全量备份再加上备份时间点之后的 binlog 来恢复数据。备份数据库一般可以用 `mysqldump` 命令工具。

```bash
mysqldump -u root 数据库名>备份文件名;   # 备份整个数据库
mysqldump -u root 数据库名 表名字>备份文件名;  # 备份整个表

mysql -u root test < 备份文件名 # 恢复整个数据库，test 为数据库名称，需要自己先建一个数据库 test
```

binlog 的 `expire_logs_days` 如何设置，一般最好比全量备份的间隔时间长，避免全量备份某天备份失败了，后面的 binlog 日志也丢失了，无法恢复数据。
{{< /callout >}}


### redo log 和 binlog 的区别

- redo log 是 InnoDB 引擎特有的；binlog 是 MySQL 的 Server 层实现的，所有引擎都可以使用。
- redo log 是物理日志，记录的是“在某个数据页上做了什么修改”；binlog 是逻辑日志，记录的是这个语句的原始逻辑，比如 “给 ID=2 这一行的 c 字段加 1”。
- redo log 是循环写的，空间固定会用完；binlog 是可以追加写入的。“追加写”是指 binlog 文件写到一定大小后会切换到下一个，并不会覆盖以前的日志。

### 为什么需要两阶段提交？

如果只有 redo log 或者只有 binlog，那么事务就不需要两阶段提交。但是如果同时使用了 redo log 和 binlog，那么就需要保证这两种日志之间的一致性。否则，在数据库发生异常重启或者主从切换时，可能会出现数据不一致的情况。

假设我们有一个事务 T，它修改了两行数据 A 和 B，并且同时开启了 redo log 和 binlog。

- 如果先写 redo log 再写 binlog，并且在写完 redo log 后数据库发生了宕机，那么在重启后，根据 redo log 可以恢复 A 和 B 的修改，但是 binlog 中没有记录数据 A 和 B 的修改信息，导致备份或者从库中没有 A 和 B 的修改。
- 如果先写 binlog 再写 redo log，并且在写完 binlog 后数据库发生了宕机，那么在重启后，根据 redo log 无法恢复 A 和 B 的修改，但是 binlog 中有记录 A 和 B 的修改信息，导致备份或者从库中有 A 和 B 的修改。

所以如果不使用“两阶段提交”，那么数据库的状态就有可能和用它的日志恢复出来的库的状态不一致。

### 为什么会有 redo log 和 binlog 两份日志？

因为最开始 MySQL 里并没有 InnoDB 引擎。MySQL 自带的引擎是 MyISAM，但是 MyISAM 没有 crash-safe 的能力， binlog 只在 Server 层记录逻辑操作，不参与存储引擎的数据页写入和 crash recovery 过程。而 InnoDB 是以插件形式引入 MySQL 的，既然只依靠 binlog 是没有 crash-safe 能力的，所以 InnoDB 使用另外一套日志系统——也就是 redo log 来实现 crash-safe 能力。有了 redo log，InnoDB 就可以保证即使数据库发生异常重启，之前提交的记录都不会丢失，这个能力称为 crash-safe。


## undo log

**事务**需要保证**原子性**，也就是事务中的操作要么全部完成，要么什么也不做。但是有时候就是会在事务执行到一半会出现一些的情况，比如：

1. 事务执行过程中可能遇到各种错误，比如服务器本身的错误，操作系统错误，甚至是突然断电导致的错误。
2. 手动输入 `ROLLBACK` 语句结束当前的事务的执行。

为了保证事务的原子性，就需要把东西改回原先的样子，这个过程就称之为**回滚**。

每当对一条记录做改动时（`INSERT`、`DELETE`、`UPDATE`），都需要把回滚时所需的东西都给记下来。例如：

- **插入一条记录时，至少要把这条记录的主键值记下来**，之后回滚的时候只需要把这个主键值对应的记录删掉就好了。
- **删除了一条记录，至少要把这条记录中的内容都记下来**，这样之后回滚时再把由这些内容组成的记录插入到表中就好了。
- **修改了一条记录，至少要把修改这条记录前的旧值都记录下来**，这样之后回滚时再把这条记录更新为旧值就好了。

这些为了回滚而记录的这些日志就称之为 **undo log**。

### 事务 id 分配的时机

事务执行过程中，只有在第一次真正修改记录时（执行 `INSERT`、`DELETE`、`UPDATE` 这些语句或加排它锁操作比如 `select...for update` 语句时），才会被分配一个单独的事务 id，否则在一个只读事务中的事务 id 值都默认为 0。

### trx_id 和 roll_pointer 隐藏列

**聚簇索引的记录除了会保存完整的用户数据以外，而且还会自动添加名为 `trx_id`、`roll_pointer` 的隐藏列**。如果用户没有在表中定义主键以及 UNIQUE 键，还会自动添加一个名为 `row_id` 的隐藏列。

- `trx_id` 就是某个聚簇索引记录被修改时所在的事务对应的**事务 id**。
- `roll_pointer` 就是一个**指向记录对应的 undo log 的一个指针**。

### undo log 回滚段

InnoDB 对 undo log 文件的管理采用段的方式，也就是**回滚段**（rollback segment）。**每个回滚段记录了 `1024` 个 undo log segment，每个事务只会使用一个 undo log segment**。

在 MySQL 5.5 时只有一个回滚段，那么最大同时支持的事务数量为 1024 个。MySQL 5.6 以后，InnoDB 支持最大 128 个回滚段，故其支持同时在线的事务限制提高到了 `128*1024`。

查看 undo log 相关参数：

```sql
show variables like '%innodb_undo%';
show variables like 'innodb_rollback_segments';
```

- `innodb_undo_directory`：undo log 文件所在的路径。默认值为 `./`，即 innodb 数据文件存储位置，目录下 `ibdata1` 文件就是 undo log 存储的位置。
- `innodb_undo_tablespaces`: undo log 文件的数量，这样回滚段可以较为平均地分布在多个文件中。设置该参数后，会在路径 `innodb_undo_directory` 看到 `undo` 为前缀的文件。
- `innodb_rollback_segments`: undo log 文件内部回滚段的个数，默认值为 `128`。

### undo log 日志什么时候删除

- **新增类型**的，在**事务提交之后就可以清除掉了**。
- **修改类型**的，事务提交之后不能立即清除掉，这些**日志会用于 MVCC。只有当没有事务用到该版本信息时才可以清除**。

## 错误日志

MySQL 有一个比较重要的日志是错误日志，它记录了数据库启动和停止，以及运行过程中发生任何严重错误时的相关信息。当数据库出现任何故障导致无法正常使用时，建议首先查看此日志。在 MySQL 数据库中，错误日志功能是默认开启的，而且无法被关闭。

```sql
show variables like '%log_error%';
```


## 通用查询日志

通用查询日志**记录用户的所有操作**，包括启动和关闭 MySQL 服务、所有用户的连接开始时间和截止时间、发给 MySQL 数据库服务器的所有 SQL 指令等，如 `select`、`show` 等，无论 SQL 的语法正确还是错误、也无论 SQL 执行成功还是失败，MySQL 都会将其记录下来。

一般不建议开启，只在需要调试查询问题时开启。因为开启会消耗系统资源并且占用大量的磁盘空间。

```sql
show variables like '%general_log%';
-- 打开通用查询日志
SET GLOBAL general_log=on;
```

## 慢查询日志

慢查询日志是 MySQL 提供的一种日志记录，它用来记录在 MySQL 中响应时间超过阈值的语句，具体指运行时间超过 `long_query_time` 值的 SQL 才会被记录到慢查询日志中。

**如果不是调优需要的话，不建议启动该参数**，因为开启慢查询日志会或多或少带来一定的性能影响。

```sql
show variables like '%slow_query%';
```

- `long_query_time`：慢查询日志的阈值，默认值为 `10`，单位为秒。意思是记录运行 10 秒以上的语句。
- `log_queries_not_using_indexes`：未使用索引的查询也被记录到慢查询日志中（可选项）。
- `log_output`：日志存储方式。`log_output='FILE'` 表示将日志存入文件，默认值是 `'FILE'`。`log_output='TABLE'` 表示将日志存入数据库。
- `log_slow_admin_statements`：表示，是否将慢管理语句例如ANALYZE TABLE和ALTER TABLE等记入慢查询日志。

开启慢查询日志：

```sql  
-- 1 表示开启，0 表示关闭
SET GLOBAL slow_query_log=1; 
```

### mysqldumpslow

MySQL 提供了日志分析工具 `mysqldumpslow`，可以用来分析慢查询日志，找出执行时间比较长的 SQL 语句。

比如，得到返回记录集最多的 10 个 SQL。

```bash
mysqldumpslow -s r -t 10 /database/mysql/mysql06_slow.log
```

得到访问次数最多的 10 个 SQL：

```bash
mysqldumpslow -s c -t 10 /database/mysql/mysql06_slow.log
```

得到按照时间排序的前 10 条里面含有左连接的查询语句：

```bash
mysqldumpslow -s t -t 10 -g “left join” /database/mysql/mysql06_slow.log
```

建议在使用这些命令时结合 `|` 和 `more` 使用，否则有可能出现刷屏的情况：

```bash
mysqldumpslow -s r -t 20 /mysqldata/mysql/mysql06-slow.log | more
```