---
title: 集群架构
weight: 5
---

## 主从集群

### 搭建 MySQL 服务

准备两台服务器，用来搭建一个 MySQL 的服务集群。两台服务器均安装 CentOS7 操作系统。MySQL 版本采用 `mysql-8.0.20` 版本。

两台服务器的 IP 分别为 `192.168.232.128` 和 `192.168.232.129`。其中 128 服务器规划为 MySQL 主节点，129 服务器规划为 MySQL 的从节点。

在两台服务器上分别安装 MySQL 服务。

```bash
groupadd mysql
useradd -r -g mysql -s /bin/false mysql  # 这里是创建一个 mysql 用户用于承载 mysql 服务，但是不需要登录权限。
tar -zxvf mysql-8.0.20-el7-x86_64.tar.gz #解压
ln -s mysql-8.0.20-el7-x86_64 mysql #建立软链接
cd mysql
mkdir mysql-files
chown mysql:mysql mysql-files
chmod 750 mysql-files
bin/mysqld --initialize --user=mysql # 初始化 mysql 数据文件 注意点1
bin/mysql_ssl_rsa_setup
bin/mysqld_safe --user=mysql 

cp support-files/mysql.server /etc/init.d/mysql.server
```

> 注意点：
>
> 1. 初始化过程中会初始化一些 MySQL 的数据文件，经常会出现一些文件或者文件夹权限不足的问题。如果有文件权限不足的问题，需要根据他的报错信息，创建对应的文件或者文件夹，并配置对应的文件权限。
> 
> 2. 初始化过程如果正常完成，日志中会打印出一个 root 用户的默认密码。这个密码需要记录下来。
>
> `2020-12-10T06:05:28.948043Z 6 [Note] [MY-010454] [Server] A temporary password is generated for root@localhost: P6kigsT6Lg>=`

启动 MySQL 服务：

```bash
bin/mysqld --user=mysql &
```

连接 MySQL：

**默认是只能从本机登录，远程是无法访问的**。所以需要用 root 用户登录下，配置远程访问的权限。

```bash
cd /root/mysql
bin/mysql -uroot -p # 然后用之前记录的默认密码登录
```

如果遇到 `ERROR 2002 (HY000): Can't connect to local MySQL server through socket '/tmp/mysql.sock'` 报错，可以参照下面的配置，修改下 `/etc/my.cnf` 配置文件，来配置下 sock 连接文件的地址。主要是下面 `client` 部分：

```ini
[mysqld]
datadir=/var/lib/mysql
socket=/var/lib/mysql/mysql.sock
user=mysql
# Disabling symbolic-links is recommended to prevent assorted security risks
symbolic-links=0
# Settings user and group are ignored when systemd is used.
# If you need to run mysqld under a different user or group,
# customize your systemd unit file for mariadb according to the
# instructions in http://fedoraproject.org/wiki/Systemd

[mysqld_safe]
log-error=/var/log/mariadb/mariadb.log
pid-file=/var/run/mariadb/mariadb.pid

#
# include all files from the config directory
#
!includedir /etc/my.cnf.d

[client]
port=3306
socket=/var/lib/mysql/mysql.sock
```

​登录进去后，需要配置远程登录权限：

```bash
alter user 'root'@'localhost' identified by '123456'; #修改 root 用户的密码
use mysql;
update user set host='%' where user='root';
flush privileges;
```

搭建完成，可以使用 navicat 等连接工具远程访问 MySQL 服务了。

搭建主从集群的多个服务，有两个必要的条件：

1. **MySQL 版本必须一致**。
2. **集群中各个服务器的时间需要同步**。

### 配置主从集群

#### 主从同步原理

数据库的主从同步，就是为了要保证多个数据库之间的数据保持一致。最简单的方式就是使用数据库的导入导出工具，定时将主库的数据导出，再导入到从库当中。这是一种很常见，也很简单易行的数据库集群方式。也有很多的工具帮助我们来做这些事情。但是这种方式进行数据同步的实时性比较差。

要保证数据能够实时同步，对于 MySQL，通常就要用到他自身提供的**一套通过 Binlog 日志在多个 MySQL 服务之间进行同步的集群方案**。基于这种集群方案，一方面可以提高数据的安全性，另外也可以以此为基础，提供读写分离、故障转移等其他高级的功能。

![mysql-cluster](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mysql-cluster.png)

**在主库上打开 Binlog 日志，记录对数据的每一步操作。然后在从库上打开 RelayLog 日志，用来记录跟主库一样的 Binlog 日志，并将 RelayLog 中的操作日志在自己数据库中进行重放**。这样就能够更加实时的保证主库与从库的数据一致。

1. 在从库上启动一系列 IO 线程，负责与主库建立 TCP 连接，请求主库在写入 Binlog 日志时，也往从库传输一份。
2. 主库上会有一个 IO Dump 线程，负责将 Binlog 日志通过这些 TCP 连接传输给从库的 IO 线程。
3. 从库为了保证日志接收的稳定性，**并不会立即重演 Binlog 数据操作**，而是先将接收到的 Binlog 日志写入到自己的 RelayLog 日志当中。然后再**异步的重演 RelayLog 中的数据操作**。

​MySQL 的 BinLog 日志能够比较实时的记录主库上的所有操作，因此他也被很多其他工具用来实时监控 MySQL 的数据变化。例如 Canal 框架，可以模拟一个 slave 节点，同步 MySQL 的 Binlog，然后将具体的数据操作按照定制的逻辑进行转发。例如转发到 Redis 实现缓存一致，转发到 Kafka 实现数据实时流转等。甚至像 ClickHouse，还支持将自己模拟成一个 MySQL 的从节点，接收 MySQL 的 Binlog 日志，实时同步 MySQL 的数据。

​接下来就在这两个 MySQL 服务的基础上，搭建一个主从集群。

#### 配置 Master 服务

配置主节点的 MySQL 配置文件： `/etc/my.cnf` (没有的话就手动创建一个)

主要是需要打开 binlog 日志，以及指定 `server-id`。打开 MySQL 主服务的 `my.cnf` 文件，在文件中一行 `server-id` 以及一个关闭域名解析的配置。然后重启服务。

```ini
[mysqld]
server-id=47
# 开启binlog
log_bin=master-bin
log_bin-index=master-bin.index
skip-name-resolve
# 设置连接端口
port=3306
# 设置 MySQL 的安装目录
basedir=/usr/local/mysql
# 设置 MySQL 数据库的数据的存放目录
datadir=/usr/local/mysql/mysql-files
# 允许最大连接数
max_connections=200
# 允许连接失败的次数。
max_connect_errors=10
# 服务端使用的字符集默认为 UTF8
character-set-server=utf8
# 创建新表时将使用的默认存储引擎
default-storage-engine=INNODB
# 默认使用 “mysql_native_password” 插件认证
#mysql_native_password
default_authentication_plugin=mysql_native_password
```

- `server-id`：服务节点的唯一标识。需要给集群中的每个服务分配一个单独的 ID。
- `log_bin`：打开 Binlog 日志记录，并指定文件名。
- `log_bin-index`：Binlog 日志文件


重启 MySQL 服务： `service mysqld restart`。

给 root 用户分配一个 replication slave 的权限：

```bash
# 登录主数据库
mysql -u root -p
GRANT REPLICATION SLAVE ON *.* TO 'root'@'%';
flush privileges;
# 查看主节点同步状态
show master status;

# 输出
+------------------+-----------+--------------+------------------+
| File             | Position  | Binlog_Do_DB | Binlog_Ignore_DB |
+------------------+-----------+--------------+------------------+
| master-bin.000004 | 156       |              |                  |
+------------------+-----------+--------------+------------------+

```

> 通常不会直接使用 root 用户，而会创建一个拥有全部权限的用户来负责主从同步。

#### 配置 Slave 服务

修改配置文件：

```ini
[mysqld]
# 主库和从库需要不一致
server-id=48
# 打开 MySQL 中继日志
relay-log-index=slave-relay-bin.index
relay-log=slave-relay-bin
# 打开从服务二进制日志
log-bin=mysql-bin
# 使得更新的数据写进二进制日志中
log-slave-updates=1
# 设置 3306 端口
port=3306
# 设置 MySQL 的安装目录
basedir=/usr/local/mysql
# 设置 MySQL 数据库的数据的存放目录
datadir=/usr/local/mysql/mysql-files
# 允许最大连接数
max_connections=200
# 允许连接失败的次数。
max_connect_errors=10
# 服务端使用的字符集默认为 UTF8
character-set-server=utf8
# 创建新表时将使用的默认存储引擎
default-storage-engine=INNODB
# 默认使用 “mysql_native_password” 插件认证
# mysql_native_password
default_authentication_plugin=mysql_native_password
```

启动 MySQL 服务，并设置他的主节点同步状态：

```bash
# 登录从服务
mysql -u root -p;
# 设置同步主节点
CHANGE MASTER TO
MASTER_HOST='192.168.232.128',
MASTER_PORT=3306,
MASTER_USER='root',
MASTER_PASSWORD='root',
MASTER_LOG_FILE='master-bin.000004',
MASTER_LOG_POS=156,
GET_MASTER_PUBLIC_KEY=1;
# 开启 slave
start slave;
# 查看主从同步状态
show slave status;
# 或者用 show slave status \G; 这样查看比较简洁
```

> 注意，`CHANGE MASTER` 指令中需要指定的 `MASTER_LOG_FILE` 和 `MASTER_LOG_POS` 必须与主服务中命令 `show master status;` 查到的保持一致。
>
> 并且后续如果要检查主从架构是否成功，也可以通过检查主服务与从服务之间的 File 和 Position 这两个属性是否一致来确定。

```bash
mysql> show slave status \G
*************************** 1. row ***************************
               Slave_IO_State: Waiting for master to send event
                  Master_Host: 192.168.232.128
                  Master_User: root
                  Master_Port: 3306
                Connect_Retry: 60
              Master_Log_File: master-bin.000004 # 主节点的 binlog 文件名
          Read_Master_Log_Pos: 156               # 主节点的 binlog 位置
           Relay_Log_File: slave-relay-bin.000006
            Relay_Log_Pos: 373
    Relay_Master_Log_File: master-bin.000004
         Slave_IO_Running: Yes
        Slave_SQL_Running: Yes
          Replicate_Do_DB:  # 从节点需要同步的数据库
      Replicate_Ignore_DB:  # 从节点不需要同步的数据库
       Replicate_Do_Table:  # 从节点需要同步的数据库中的表
   Replicate_Ignore_Table:  # 从节点不需要同步的数据库中的表
  Replicate_Wild_Do_Table:  # 从节点需要同步的数据库中的通配符表，例如 test.%
Replicate_Wild_Ignore_Table: # 从节点不需要同步的数据库中的通配符表，例如 test.% 

```

{{< callout type="info" >}}
**`Replicate_` 开头这些属性指定了两个服务之间要同步哪些数据库、哪些表的配置**。只是在这个示例中全都**没有进行配置，就标识是全库进行同步**。
{{< /callout >}}


#### 主从集群测试

先用 `showdatabases`，查看下两个 MySQL 服务中的数据库情况：

```bash
# master
mysql> show databases;

+--------------------+
| Database           |
+--------------------+
| information_schema |
| masterdemo         |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
5 rows in set (0.00 sec)

# slave
mysql> show databases;

+--------------------+
| Database           |
+--------------------+
| information_schema |
| masterdemo         |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
5 rows in set (0.01 sec)
```

​ 然后在主服务器上创建一个数据库：

```bash
mysql> create database syncdemo;
Query OK, 1 row affected (0.00 sec)
```

再用 `show databases`，来看下这个 `syncdemo` 的数据库是不是已经同步到了从服务：

```bash
# slave
mysql> show databases;

+--------------------+
| Database           |
+--------------------+
| information_schema |
| masterdemo         |
| mysql              |
| performance_schema |
| syncdemo           |
| sys                |
+--------------------+
6 rows in set (0.00 sec)
```

继续在 `syncdemo` 这个数据库中创建一个表，并插入一条数据：

```bash
mysql> use syncdemo;
Database changed
mysql> create table demoTable(id int not null);
Query OK, 0 rows affected (0.02 sec)

mysql> insert into demoTable value(1);
Query OK, 1 row affected (0.01 sec)
```

查一下这个 `demoTable` 是否同步到了从服务：

```bash
# slave
mysql> use syncdemo;
Database changed
mysql> show tables;
+--------------------+
| Tables_in_syncdemo |
+--------------------+
| demoTable          |
+--------------------+
1 row in set (0.00 sec)

# 查看 demoTable 表中的数据
mysql> select * from demoTable;
+----+
| id |
+----+
|  1 |
+----+
1 row in set (0.00 sec)

```

这样，一个主从集群就搭建完成了。

{{< callout type="warn" >}}
另外，这个主从架构是有可能失败的，如果在 slave 从服务上查看 slave 状态，发`现Slave_SQL_Running=no`，就表示**主从同步失败**了。这有可能是因为在从数据库上进行了写操作，与同步过来的 SQL 操作冲突了，也有可能是 slave 从服务重启后有事务回滚了。

如果是因为 slave 从服务事务回滚的原因，可以按照以下方式重启主从同步：

```bash
mysql> stop slave ;
mysql> set GLOBAL SQL_SLAVE_SKIP_COUNTER=1;
mysql> start slave ;
```

另一种解决方式就是重新记录主节点的 binlog 文件消息：

```bash
mysql> stop slave ;
mysql> change master to .....
mysql> start slave ;
```

这种方式要注意 binlog 的文件和位置，如果修改后和之前的同步接不上，那就会丢失部分数据。所以不太常用。
{{< /callout >}}


#### 全库同步与部分同步

目前配置的主从同步是针对全库配置的，而实际环境中，一般并不需要针对全库做备份，而只需要对一些特别重要的库或者表来进行同步。那如何针对库和表做同步配置？

在 Master 端：在 `my.cnf` 中，可以通过以下这些属性指定需要针对哪些库或者哪些表记录 binlog：

```ini
# 需要同步的二进制数据库名
binlog-do-db=masterdemo
# 只保留 7 天的二进制日志，以防磁盘被日志占满(可选)
expire-logs-days  = 7
# 不备份的数据库
binlog-ignore-db=information_schema
binlog-ignore-db=performation_schema
binlog-ignore-db=sys
```

在 Slave 端：在 `my.cnf` 中，需要配置备份库与主服务的库的对应关系：

```ini
# 如果 salve 库名称与 master 库名相同，使用本配置
replicate-do-db = masterdemo 
# 如果 master 库名 [mastdemo] 与 salve 库名 [mastdemo01] 不同，使用以下配置 [需要做映射]
replicate-rewrite-db = masterdemo -> masterdemo01
# 如果不是要全部同步 [默认全部同步]，则指定需要同步的表
replicate-wild-do-table=masterdemo01.t_dict
replicate-wild-do-table=masterdemo01.t_num
```

配置完成了之后，在 `show master status` 指令中，就可以看到 `Binlog_Do_DB` 和 `Binlog_Ignore_DB` 两个参数的作用了。


#### GTID 同步集群

上面搭建的集群方式，是基于 Binlog 日志记录点的方式来搭建的，这也是最为传统的 MySQL 集群搭建方式。

另外一种搭建主从同步的方式，即 GTID 搭建方式。这种模式是从 MySQL 5.6 版本引入的。

**GTID 的本质也是基于 Binlog 来实现主从同步，只是他会基于一个全局的事务 ID 来标识同步进度。GTID 即全局事务 ID，全局唯一并且趋势递增，他可以保证为每一个在主节点上提交的事务在复制集群中可以生成一个唯一的 ID**。

在基于 GTID 的复制中，首先从服务器会告诉主服务器已经在从服务器执行完了哪些事务的 GTID 值，然后主库会有把所有没有在从库上执行的事务，发送到从库上进行执行，并且使用 GTID 的复制可以保证同一个事务只在指定的从库上执行一次，这样可以避免由于偏移量的问题造成数据不一致。

搭建方式跟主从架构整体搭建方式差不多。只是需要在 `my.cnf` 中修改一些配置：

主节点：

```ini
gtid_mode=on
enforce_gtid_consistency=on
log_bin=on
server_id=单独设置一个
binlog_format=row
```

从节点：

```ini
gtid_mode=on
enforce_gtid_consistency=on
log_slave_updates=1
server_id=单独设置一个
```

然后分别重启主服务和从服务，就可以开启 GTID 同步复制方式。

## 集群扩容与 MySQL 数据迁移

如果要扩展到一主多从的集群架构，其实就比较简单了，只需要增加一个 binlog 复制就行了。

但是如果我们的集群是已经运行过一段时间，这时候如果要扩展新的从节点就有一个问题，之前的数据没办法从 binlog 来恢复了。这时候在扩展新的 slave 节点时，就需要增加一个数据复制的操作。

​MySQL 的数据备份恢复操作相对比较简单，可以通过 SQL 语句直接来完成。具体操作可以使用 MySQL 的 `bin` 目录下的 `mysqldump` 工具：

```bash
mysqldump -u root -p --all-databases > backup.sql
#输入密码
```

通过这个指令，就可以将整个数据库的所有数据导出成 `backup.sql`，然后把这个 `backup.sql` 分发到新的 MySQL 服务器上，并执行下面的指令将数据全部导入到新的 MySQL 服务中：

```bash
mysql -u root -p < backup.sql
#输入密码
```

这样新的 MySQL 服务就已经有了所有的历史数据，然后就可以再按照上面的步骤，配置 Slave 从服务的数据同步了。

## 搭建半同步复制

### 半同步复制的原理

#### 主从复制数据丢失问题

MySQL 的主从集群，是有丢失数据的风险的。

MySQL 主从集群默认采用的是一种异步复制的机制。主服务在执行用户提交的事务后，写入 binlog 日志，然后就给客户端返回一个成功的响应了。而 binlog 会由一个 dump 线程异步发送给 Slave 从服务。

![mysql-copy-async](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mysql-copy-async.png)


由于这个发送 binlog 的过程是异步的。主服务在向客户端反馈执行结果时，是不知道 binlog 是否同步成功了的。这时候如果**主服务宕机了，而从服务还没有备份到新执行的 binlog，那就有可能会丢数据**。

怎么解决这个问题，这就要靠 **MySQL 的半同步复制机制来保证数据安全**。

​半同步复制机制是一种介于异步复制和全同步复制之前的机制。主库在执行完客户端提交的事务后，并不是立即返回客户端响应，而是**等待至少 N 个从库接收并写到 relay log 中，才会返回给客户端**。MySQL 在等待确认时，默认会等 10 秒，如果超过 10 秒没有收到 ack，就会降级成为异步复制（也就是说超时了以后 master 就不等了，直接返回客户端去了，也不会当做失败来处理）。

![mysql-copy-halfsync](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mysql-copy-halfsync.png)

这种半同步复制相比异步复制，能够有效的提高数据的安全性。但是这种安全性也不是绝对的，他只保证事务提交后的 binlog 至少传输到了一个从库，并且并不保证从库应用这个事务的 binlog 是成功的。另一方面，半同步复制机制也会造成一定程度的延迟，这个延迟时间最少是一个 TCP/IP 请求往返的时间。整个服务的性能是会有所下降的。而当从服务出现问题时，主服务需要等待的时间就会更长，要等到从服务的服务恢复或者请求超时才能给用户响应。

#### 无损半同步

MySQL 5.7 版本前的半同步复制机制是**有损的**，这种半同步复制在 Master 发生宕机时，**Slave 会丢失最后一批提交的数据**

![mysql-lossy-semi-sync](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mysql-lossy-semi-sync.png)


有损半同步是在 Master 事务提交后，即步骤 4 后，等待 Slave 返回 ACK，表示至少有 Slave 接收到了二进制日志，如果这时二进制日志还未发送到 Slave，Master 就发生宕机，则此时 Slave 就会丢失 Master 已经提交的数据。

MySQL 5.7 版本之后，引入了无损半同步复制的机制。

![mysql-lossless-semi-sync](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mysql-lossless-semi-sync.png)

无损半同步复制 WAIT ACK 发生在事务提交之前，这样即便 Slave 没有收到二进制日志，但是 Master 宕机了，***由于最后一个事务还没有提交，所以本身这个数据对外也不可见**，不存在丢失的问题。

#### 搭建半同步复制集群

半同步复制需要基于特定的扩展模块来实现。而 MySQL 从 5.5 版本开始，往上的版本都默认自带了这个模块。这个模块包含在 MySQL 安装目录下的 `lib/plugin` 目录下的 `semisync_master.so` 和 `semisync_slave.so` 两个文件中。需要在主服务上安装 `semisync_master` 模块，在从服务上安装 `semisync_slave` 模块。

登陆主节点，安装 `semisync_master` 模块：

```bash
# 安装半同步复制模块，指定扩展库的文件名
mysql> install plugin rpl_semi_sync_master soname 'semisync_master.so';
Query OK, 0 rows affected (0.01 sec)

# 查看系统全局参数，rpl_semi_sync_master_timeout 就是半同步复制时等待应答的最长等待时间，默认是 10 秒，可以根据情况自行调整。
mysql> show global variables like 'rpl_semi%';
+-------------------------------------------+------------+
| Variable_name                             | Value      |
+-------------------------------------------+------------+
| rpl_semi_sync_master_enabled              | OFF        |
| rpl_semi_sync_master_timeout              | 10000      |
| rpl_semi_sync_master_trace_level          | 32         |
| rpl_semi_sync_master_wait_for_slave_count | 1          |
| rpl_semi_sync_master_wait_no_slave        | ON         |
| rpl_semi_sync_master_wait_point           | AFTER_SYNC |
+-------------------------------------------+------------+
6 rows in set, 1 warning (0.02 sec)

# 打开半同步复制的开关
mysql> set global rpl_semi_sync_master_enabled=ON;
Query OK, 0 rows affected (0.00 sec)
```


> 最后的一个参数 `rpl_semi_sync_master_wait_point` 其实表示一种半同步复制的方式。
>
> 半同步复制有两种方式，一种是我们现在看到的这种默认的 `AFTER_SYNC` 方式。这种方式下，主库把日志写入 binlog，并且复制给从库，然后开始等待从库的响应。从库返回成功后，主库再提交事务，接着给客户端返回一个成功响应。
>
> 而另一种方式是叫做 `AFTER_COMMIT` 方式。他不是默认的。这种方式，在主库写入 binlog 后，等待 binlog 复制到从库，主库就提交自己的本地事务，再等待从库返回给自己一个成功响应，然后主库再给客户端返回响应。

登陆从节点，安装 `smeisync_slave` 模块：

```bash
mysql> install plugin rpl_semi_sync_slave soname 'semisync_slave.so';
Query OK, 0 rows affected (0.01 sec)

mysql> show global variables like 'rpl_semi%';
+---------------------------------+-------+
| Variable_name                   | Value |
+---------------------------------+-------+
| rpl_semi_sync_slave_enabled     | OFF   |
| rpl_semi_sync_slave_trace_level | 32    |
+---------------------------------+-------+
2 rows in set, 1 warning (0.01 sec)

mysql> set global rpl_semi_sync_slave_enabled = on;
Query OK, 0 rows affected (0.00 sec)

mysql> show global variables like 'rpl_semi%';
+---------------------------------+-------+
| Variable_name                   | Value |
+---------------------------------+-------+
| rpl_semi_sync_slave_enabled     | ON    |
| rpl_semi_sync_slave_trace_level | 32    |
+---------------------------------+-------+
2 rows in set, 1 warning (0.00 sec)

mysql> stop slave;
Query OK, 0 rows affected (0.01 sec)

mysql> start slave;
Query OK, 0 rows affected (0.01 sec)
```

## 主从复制延迟优化

MySQL 数据库中，大事务除了会导致提交速度变慢，还会导致主从复制延迟。为什么说会导致主从复制延迟？假设一个大事务运行了十分钟，那么在从服务器上也需要运行十分钟回放这个大事务。也就是说 从服务器可能需要十分钟才能追上主服务器。主从服务器之间的数据就产生了延迟。

优化：

- 把大事务拆分成小事务，可以避免主从复制延迟，
- 设置复制回放相关的配置参数。

要彻底避免 MySQL 主从复制延迟，数据库版本至少要升级到 5.7，因为之前的 **MySQL 版本从机回放二进制都是单线程的（5.6 是基于库级别的单线程）**。从 MySQL 5.7 版本开始，**MySQL 支持了从机多线程回放二进制日志的方式**，通常把它叫作**并行复制**，官方文档中称为 “Multi-Threaded Slave（MTS）”。

MySQL 的从机并行复制有两种模式。

- **COMMIT ORDER**：主机怎么并行，从机就怎么并行。从机完全根据主服务的并行度进行回放。理论上来说，主从延迟极小。但如果主服务器上并行度非常小，事务并不小，比如单线程每次插入 1000 条记录，则从机单线程回放，也会存在一些复制延迟的情况。
- **WRITESET**：基于每个事务，只要事务更新的记录不冲突，就可以并行。以 “单线程每次插入 1000 条记录” 为例，如果插入的记录没有冲突，比如唯一索引冲突，那么**虽然主机是单线程，但从机可以是多线程并行回放！！！**

启用 WRITESET 复制模式：

```ini
binlog_transaction_dependency_tracking = WRITESET
transaction_write_set_extraction = XXHASH64
slave-parallel-type = LOGICAL_CLOCK
slave-parallel-workers = 16
```

## 读写分离

只能从主节点同步到从节点，而从节点的数据表更是无法同步到主节点的。为了保证数据一致，通常会需要保证数据只在主节点上写，而从节点只进行数据读取。这就是读写分离。

**MySQL 主从本身是无法提供读写分离的服务的，需要由业务自己来实现**。

一种常见的业务读写分离的架构设计：

![mysql-read-write-separation](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mysql-read-write-separation.png)

引入了 Load Balance 负载均衡的组件，这样 Server 对于数据库的请求不用关心后面有多少个从机，对于业务来说也就是透明的，只需访问 Load Balance 服务器的 IP 或域名就可以。

> 要限制用户写数据，我们可以在从服务中将 `read_only` 参数的值设为 `1` ( `set global read_only=1;`)。这样就可以限制用户写入数据。但是这个属性有两个需要注意的地方：
>
> 1. `read_only=1` 设置的只读模式，不会影响 slave 同步复制的功能。 所以在 MySQL slave 库中设定了 `read_only=1` 后，通过 `show slave status\G` 命令查看 salve 状态，可以看到 salve 仍然会读取 master 上的日志，并且在 slave 库中应用日志，保证主从数据库同步一致；
> 2. `read_only=1` 设置的只读模式，限定的是普通用户进行数据修改的操作，但**不会限定具有 `super` 权限的用户的数据修改操作**。在 MySQL 中设置 `read_only=1` 后，普通的应用用户进行 `insert`、`update`、`delete` 等会产生数据变化的 DML 操作时，都会报出数据库处于只读模式不能发生数据变化的错误，但具有 `super` 权限的用户，例如在本地或远程通过 root 用户登录到数据库，还是可以进行数据变化的 DML 操作； 如果**需要限定 `super` 权限的用户写数据，可以设置  `super_read_only=0`**。另外 如果要想连 `super` 权限用户的写操作也禁止，就**使用 `flush tables with read lock;`，这样设置也会阻止主从同步复制**！

{{< callout type="info" >}}
**读写分离设计的前提是从机不能落后主机很多，最好是能准实时数据同步，务必一定要开始并行复制，并确保线上已经将大事务拆成小事务**。
{{< /callout >}}

{{< callout type="info" >}}
在 Load Balance 服务器，可以配置较小比例的读取请求访问主机，如主服务器权重为的 1% 的读请求，其余三台从服务器各自承担 33% 的读取请求。

如果发生严重的主从复制延迟情况，可以设置下面从机权重为 0，将主机权重设置为 100%，这样就不会因为数据延迟，导致对于业务的影响了。
{{< /callout >}}

## 多源复制

无论是异步复制还是半同步复制，都是 1 个 Master 对应 N 个 Slave。其实 MySQL 也支持 N 个 Master 对应 1 个 Slave，这种架构就称之为**多源复制**。

多源复制允许在不同 MySQL 实例上的数据同步到 1 台 MySQL 实例上，方便在 1 台 Slave 服务器上进行一些统计查询，如常见的 OLAP 业务查询。

![mysql-multi-source](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mysql-multi-source.png)


上图显示了订单库、库存库、供应商库，通过多源复制同步到了一台 MySQL 实例上，接着就可以通过 MySQL 8.0 提供的复杂 SQL 能力，对业务进行深度的数据分析和挖掘。

## 延迟复制

前面介绍的复制架构，Slave 在接收二进制日志后会尽可能快地回放日志，这样是为了避免主从之间出现延迟。而**延迟复制却允许 Slave 延迟回放接收到的二进制日志，为了避免主服务器上的误操作，马上又同步到了从服务器，导致数据完全丢失**。

可以通过以下命令设置延迟复制：

```sql
# 设置了 Slave 落后 Master 服务器 1 个小时
CHANGE MASTER TO master_delay = 3600
```

延迟复制在数据库的备份架构设计中非常常见，比如可以设置一个延迟一天的延迟备机，这样本质上说，用户可以有 1 份 24 小时前的快照。

那么当线上发生误操作，如 `DROP TABLE`、`DROP DATABASE` 这样灾难性的命令时，用户有一个 24 小时前的快照，数据可以快速恢复。

## 高可用方案

MySQL **主从复制是高可用的技术基础**，高可用套件是 MySQL 高可用实现的解决方案，负责 Failover 操作。

为了不让业务感知到数据库的宕机切换，还要要用到 VIP（Virtual IP）技术。VIP 不是真实的物理 IP，而是可以随意绑定在任何一台服务器上。

业务访问数据库，不是服务器上与网卡绑定的物理 IP，而是这台服务器上的 VIP。当数据库服务器发生宕机时，高可用套件会把 VIP 漂移到新的服务器上。数据库 Failover后，业务依旧访问的还是 VIP，所以**使用 VIP 可以做到对业务透明**。

但是 VIP 也是有局限性的，仅限于同机房同网段的 IP 设定。如果是跨机房容灾架构，VIP 就不可用了。这时就要用 DNS（Domain Name Service）服务。

上层业务通过域名进行访问。当发生宕机，进行机房级切换后，高可用套件会把域名指向为新的 MySQL 主服务器，这样也实现了对于上层服务的透明性。

### MHA

[MHA（Master High Availability）](https://github.com/yoshinorim/mha4mysql-manager) 是一款开源的 MySQL 高可用程序，它为 MySQL 数据库主从复制架构提供了 automating master failover 的功能。它由两大组件所组成，MHA Manger 和 MHA Node。

MHA Manager 可以单独部署在一台独立的机器上管理多个 master-slave 集群，也可以部署在一台 slave 节点上。MHA Manager会定时探测集群中的 master 节点，当 master 出现故障时，它可以自动将最新数据的 slave 提升为新的 master，然后将所有其他的 slave 重新指向新的 master。整个故障转移过程对应用程序完全透明。

而 MHA Node 部署在每台 MySQL 服务器上，MHA Manager 通过执行 Node 节点的脚本完成 failover 切换操作。

![mysql-mha](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mysql-mha.png)


MHA Manager 和 MHA Node 的通信是采用 ssh 的方式，也就是**需要在生产环境中打通 MHA Manager 到所有 MySQL 节点的 ssh 策略，那么这里就存在潜在的安全风险**。

另外，**ssh 通信，效率也不是特别高**。所以，MHA 比较适合用于规模不是特别大的公司，所有 MySQL 数据库的服务器数量不超过 20 台。

### MGR

MGR：MySQL Group Replication。 是 MySQL 官方在 5.7.17 版本正式推出的一种组复制机制。主要是解决传统异步复制和半同步复制的数据一致性问题。

不要简单认为它是一种新的数据同步技术，而是应该把它理解为高可用解决方案，而且特别适合应用于对于数据一致性要求极高的金融级业务场景。

首先，MGR 之间的数据同步并没有采用复制技术，而是采用 **GCS（Group Communication System）协议**的日志同步技术。

**GSC 本身是一种类似 Paxos 算法的协议，要求组中的大部分节点都接收到日志，事务才能提交**。所以，**MRG 是严格要求数据一致的**，特别适合用于金融级的环境。由于是类 Paxos 算法，集群的节点要求数量是奇数个，这样才能满足大多数的要求。

之前介绍的无损半同步也能保证数据强一致的要求吗？

是的，**虽然通过无损半同步复制也能保证主从数据的一致性，但通过 GCS 进行数据同步有着更好的性能**：当**启用 MGR 插件时，MySQL 会新开启一个端口用于数据的同步，而不是如复制一样使用 MySQL 服务端口，这样会大大提升复制的效率**。

MGR 有两种模式：
- 单主（Single Primary）模式；
- 多主（Multi Primary）模式。

**单主模式只有 1 个节点可以写入，多主模式能让每个节点都可以写入**。而多个节点之间写入，**如果存在变更同一行的冲突，MySQL 会自动回滚其中一个事务，自动保证数据在多个节点之间的完整性和一致性**。

最后，在单主模式下，**MGR 可以自动进行 Failover 切换**，不用依赖外部的各种高可用套件，所有的事情都由数据库自己完成，比如最复杂的选主（Primary Election）逻辑，都是由 MGR 自己完成，用户不用部署额外的 Agent 等组件。

#### MGR 的限制

- 仅支持 InnoDB 表，并且每张表一定要有一个主键；
- 目前一个 MGR 集群，最多只支持 9 个节点；
- 有一个节点网络出现抖动或不稳定，会影响集群的性能。

第 1、2 点问题不大，因为目前用 MySQL 主流的就是使用 InnoDB 存储引擎，9 个节点也足够用了。

而第 3 点要注意，和复制不一样的是，由于 MGR 使用的是 Paxos 协议，对于网络极其敏感，如果其中一个节点网络变慢，则会影响整个集群性能。而半同步复制，比如 ACK 为1，则 1 个节点网络出现问题，不影响整个集群的性能。所以，在决定使用 MGR 后，切记一定要严格保障网络的质量。

{{< callout type="info" >}}
1. 依赖网络通信

Paxos 算法的核心是通过节点之间的消息传递来实现一致性。它需要多个阶段的通信，包括提案阶段（Prepare 阶段）、接受阶段（Accept 阶段）和学习阶段（Learn 阶段）。每个阶段都需要节点之间发送和接收消息。
- Prepare 阶段：提议者（Proposer）向其他节点发送提案请求（Prepare 消息），**询问是否可以接受新的提案**。
- Accept 阶段：提议者在收到足够多的响应后，向其他节点发送接受请求（Accept 消息），**要求其他节点接受该提案**。
- Learn 阶段：接受者（Acceptor）将接受的提案告知其他节点，最终达成一致。

如果网络延迟过高或消息丢失，这些阶段的通信可能会被阻塞或失败，导致整个协议无法正常推进。

2. 对网络延迟敏感

Paxos 算法的性能高度依赖网络延迟。在每个阶段，节点都需要等待其他节点的响应。例如：
在 Prepare 阶段，提议者需要等待大多数节点的响应才能进入下一个阶段。
在 Accept 阶段，提议者需要等待大多数节点接受提案后才能认为提案成功。

如果网络延迟过高，节点等待响应的时间就会增加，导致整个协议的执行时间变长。在极端情况下，如果延迟过高，可能会导致**协议超时，从而需要重新发起一轮新的协商过程**。
{{< /callout >}}


## 应用层提供管理多个数据源的能力

有了集群化的后端数据库之后，接下来在应用层面，就需要能够随意访问多个数据库的能力。

多数据源访问的实现方式有很多，例如基 Spring 提供的 `AbstractRoutingDataSource` 组件，就可以快速切换后端访问的实际数据库。

可以配置两个不同的目标数据库，然后通过 `DynamicDataSource` 组件中的一个 `ThreadLocal` 变量实现快速切换目标数据库，从而让写接口与读接口分别操作两个不同的数据库。

由于主库与从库之间可以同步数据，虽然写接口与读接口是访问的不同的数据库，但是由于两个数据库之间可以通过主从集群进行数据同步，所以看起来，课程管理的两个接口就像是访问同一个数据库一样。这其实就是对于数据库非常常见的一种分布式的优化方案**读写分离**。

### 读写分离

数据库读写分离是一种常见的数据库优化方案，其基础思想是**将对数据的读请求和写请求分别分配到不同的数据库服务器上，以提高系统的性能和可扩展性**。

一般情况下，数据库的读操作比写操作更为频繁，而且读操作并不会对数据进行修改，因此**可以将读请求分配到多个从数据库服务器上进行处理**。这样，即使一个从数据库服务器故障或者过载，仍然可以使用其他从数据库服务器来处理读请求，保证系统的稳定性和可用性。同时，将写操作分配到主数据库服务器上，可以保证数据的一致性和可靠性。主数据库服务器负责所有的写操作，而从数据库服务器只需要从主数据库服务器同步数据即可。由于**主数据库服务器是唯一的写入点，可以保证数据的正确性和一致性**。

### 将多个数据源抽象成一个统一的数据源

`DynamicDataSource` 的实现方式其实对开发方式的侵入是挺大的，每次进行数据库操作之前，都需要先选择要操作那个数据库。有没有更为自然的多数据源管理方式？就是让业务真正像操作单数据源一样访问多个数据。MyBatis-Plus 框架的开发者就开发了这样的一个框架 `DynamicDataSource`，可以简化多数据源访问的过程。

这个开源的 `DynamicDataSource` 小框架会自行在 Spring 容器当中注册一个具备多数据源切换能力的 `DataSource` 数据源，这样，在应用层面，只需要按照 `DynamicDataSource` 框架的要求修改配置接口，其他地方几乎感知不到与传统操作单数据源有什么区别。

应用只需要像访问单个数据源一样，访问 `DynamicDataSource` 框架提供的一个**逻辑数据库**。而这个逻辑数据库会帮我们将实际的 SQL 语句转发到后面的**真实数据库**当中去执行。

其实也是 ShardingSphere 需要做的事情。只不过，ShardingSphere 提供的逻辑库功能要强大很多，也复杂很多。