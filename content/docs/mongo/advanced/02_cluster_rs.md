---
title: 集群架构 - 复制集
weight: 2
---

MongoDB 有两种集群架构，分别是：

- 主从复制集
- 分片集群

## 复制集介绍

复制集是 MongoDB 最基本的集群架构，它是由一组 MongodB 实例组成的集群，包含一个 Primary 节点，多个 Secondary 节点。

所有数据都写入 Primary，Secondary 从 Primary 同步写入的数据，以保持复制集内所有成员存储相同的数据集，提供数据的高可用。

复制集的高可用依赖于两个方面：

- 数据写入时将数据迅速复制到另一个独立节点上。
- 在接受写入的节点发生故障时自动选举出一个新的替代节点。

复制集其他几个附加作用：

- 数据分发: 将数据从一个区域复制到另一个区域，减少另一个区域的读延迟。
- 读写分离: 不同类型的压力分别在不同的节点上执行。
- 异地容灾: 在数据中心故障时候快速切换到异地。

早期版本的 MongoDB 使用了一种 Master-Slave 的主从架构，该做法在 MongoDB 3.4 版本之后已经废弃。

{{< callout type="info" >}}
MySQL 和 Redis 都实现了 Master-Slave 的主从架构，**普通的主从架构是没有自动故障切换的能力的**，一旦 Master 节点发生故障，需要手动切换 Slave 节点为 Master 节点。
{{< /callout >}}

## 三节点复制集模式

常见的复制集架构由 3 个成员节点组成，官方提供了两种方案。

### PSS 模式（官方推荐）

PSS模式由一个 Primary 节点和两个 Secondary 节点所组成。 

![mongodb-pss](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-pss.png)

此模式始终提供两个完整的副本，如果 Primary 节点不可用，则复制集会自动选择一个 Secondary 节点作为新的 Primary 节点。旧的 Primary 节点在可用时重新加入复制集。

### PSA 模式

PSA 模式由一个 Primary 节点、一个 Secondary 节点和**一个仲裁者节点**（Arbiter）组成。

![mongodb-psa](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-psa.png)

其中，**Arbiter 节点不存储数据副本，也不提供业务的读写操作**。Arbiter 节点**发生故障不影响业务，仅影响选举投票**。此模式仅提供数据的一个完整副本，如果主节点不可用，则复制集将选择备节点作为主节点。

### 典型三节点复制集环境搭建

**环境准备**

- 安装 MongoDB 并配置好环境变量
- 确保有 10GB 以上的硬盘空间

#### 准备配置文件

复制集的每个 mongod 进程应该位于不同的服务器。现在是在一台机器上运行 3 个进程来模拟，因此要为它们各自配置：

- 不同的端口（28017/28018/28019）
- 不同的数据目录，`mkdir -p /data/db{1,2,3}` 
- 不同日志文件路径 (例如：`/data/db1/mongod.log`)

如果是在三台不同的机器上，那就不用创建不同的数据目录了和使用不同的端口了。


创建配置文件 `/data/db1/mongod.conf`，内容如下：

```yaml
systemLog:
  destination: file
  path: /data/db1/mongod.log # log path
  logAppend: true
storage:   
  dbPath: /data/db1 # data directory      
net:
  bindIp: 0.0.0.0
  port: 28017 # port
replication:
  replSetName: rs0  
processManagement:
  fork: true
```

参考上面配置修改端口，路径，依次配置 db2，db3。注意必须是 `yaml` 格式。

**启动 MongoDB 进程**：

```bash
mongod -f /data/db1/mongod.conf 
mongod -f /data/db2/mongod.conf 
mongod -f /data/db3/mongod.conf
```

如果启用了 SELinux，可能阻止上述进程启动。简单起见关闭 SELinux：

```bash
# 永久关闭,将 SELINUX=enforcing 改为 SELINUX=disabled,设置后需要重启才能生效
vim /etc/selinux/config
# 查看 SELINUX
/usr/sbin/sestatus -v
```

#### 配置复制集

**复制集通过 mongosh 的 `rs.initiate()` 进行初始化**，初始化后各个成员间开始发送心跳消息，并发起 Priamry 选举操作，获得**大多数**成员投票支持的节点，会成为 Primary，其余节点成为 Secondary。

```bash
# mongosh --port 28017 
# 初始化复制集
> rs.initiate({
    _id: "rs0", # 复制集的名称，必须唯一
    members: [{
        _id: 0, # 成员的编号，必须唯一，不能重复
        host: "192.168.65.174:28017"
    },{
        _id: 1,
        host: "192.168.65.174:28018"
    },{
        _id: 2,
        host: "192.168.65.174:28019"
    }]
})
```

也可以是用 `rs.add()` 来添加成员。可以使用 `rs.help()` 除了当前节点角色信息，是一个更精简化的信息，也返回整个复制集的成员列表。

#### 验证主从节点读写操作

`rs.status()` 可以查看复制集的状态。`db.isMaster()` 

- MongoDB 主节点进行写入

```javascript
// mongosh --port 28017
db.user.insertMany([{name:"fox"},{name:"monkey"}])
```

- 切换到从节点写入

```javascript
// mongosh --port 28018  
db.user.insertMany([{name:"fox"},{name:"monkey"}])
```

抛出异常 `MongoBulkWriteError: not primary`，从节点是不能写入的。

- MongoDB 从节点进行读

```bash
# mongo --port 28018
rs0:SECONDARY> db.user.find()
```

抛出异常 `MongoBulkWriteError: not primary and secondaryOk=false`，**从节点默认是不可读的**。

```bash
# 设置从节点可读
rs0:SECONDARY> rs.secondaryOk()
rs0:SECONDARY> db.user.find()
```

## 复制集常用命令

| 命令 | 描述 |
| --- | --- |
| `rs.add()` | 为复制集新增节点 |
| `rs.addArb()` | 为复制集新增一个仲裁者（arbiter） |
| `rs.conf()` | 返回复制集配置信息 |
| `rs.freeze()` | 防止当前节点在一段时间内选举成为主节点 |
| `rs.help()` | 返回 replica set 的命令帮助 |
| `rs.initiate()` | 初始化一个新的复制集 |
| `rs.printReplicationInfo()` | 以主节点的视角返回复制的状态报告 |
| `rs.printSecondaryReplicationInfo()` | 以从节点的视角返回复制状态报告 |
| `rs.reconfig()` | 通过重新应用复制集配置来为复制集更新配置 |
| `rs.remove()` | 从复制集中移除一个节点 |
| `rs.secondaryOk()` | 为当前的连接设置从节点可读 |
| `rs.status()` | 返回复制集状态信息 |
| `rs.stepDown()` | 让当前的 primary 变为从节点并触发选举 |
| `rs.syncFrom()` | 设置复制集节点从哪个节点处同步数据，将会覆盖默认选取逻辑 |


## 安全认证

复制集集群就不需要使用 `--auth` 参数了，直接创建 keyFile 文件：

keyFile 文件的作用：集群之间的安全认证，增加安全认证机制 KeyFile（开启 keyfile 认证就默认开启了 auth 认证了）。

```bash
# mongo.key 采用随机算法生成，用作节点内部通信的密钥文件。
openssl rand -base64 756 > /data/mongo.key
# 权限必须是 600
chmod 600 /data/mongo.key  
```

**注意：创建 keyFile 前，需要先停掉复制集中所有主从节点的 mongod 服务，然后再创建，否则有可能出现服务启动不了的情况**。

将主节点中的 keyfile 文件拷贝到复制集其他从节点服务器中，路径地址对应 `mongo.conf` 配置文件中的 keyFile 字段地址，并设置 keyfile 权限为 600。

启动 mongod：

```bash
mongod -f /data/db1/mongod.conf --keyFile /data/mongo.key
mongod -f /data/db2/mongod.conf --keyFile /data/mongo.key
mongod -f /data/db3/mongod.conf --keyFile /data/mongo.key

# 进入主节点
mongosh --port 28017 -u fox -p fox --authenticationDatabase=admin
```

## 复制集连接方式

方式一：直接连接 Primary 节点，正常情况下可读写 MongoDB，但主节点故障切换后，无法正常访问。


```bash
# 这种写死了 host 和 port 的方式，一旦主节点故障切换，就无法正常访问了。
mongosh -u fox -p fox 192.168.65.206:28018
```

**方式二（强烈推荐）**：通过高可用 Uri 的方式连接 MongoDB，当 Primary 故障切换后，**MongoDB Driver 可自动感知并把流量路由到新的 Primary 节点**。

```bash
mongosh mongodb://fox:fox@192.168.65.206:28017,192.168.65.206:28018,192.168.65.206:28019/admin?replicaSet=rs0
```

## 复制集成员角色 

复制集里面有多个节点，每个节点拥有不同的职责。 

在看成员角色之前，先了解两个重要属性： 

### Priority = 0 

**当 Priority 等于 0 时，它不可以被复制集选举为主，Priority 的值越高，则被选举为主的概率更大**。通常，在跨机房方式下部署复制集可以使用该特性。假设使用了机房 A 和机房 B，由于主要业务与机房 A 更近，则可以将机房 B 的复制集成员 Priority 设置为 0，这样主节点就一定会是 A 机房的成员。 

### Vote = 0

**Vote 等于 0 时，不可以参与选举投票。`priority=0 + vote=0`：节点永不参与选举，仅作备份**。由于**一个复制集中最多只有 7 个投票成员**，因此多出来的成员则必须将其 vote 属性值设置为 0，即这些成员将无法参与投票。

### 成员角色

- **Primary：主节点**，其接收所有的写请求，然后把修改同步到所有备节点。一个复制集只能有一个主节点，当主节点“挂掉”后，其他节点会重新选举出来一个主节点。
- **Secondary：备节点**，与主节点保持同样的数据集。当主节点“挂掉”时，参与竞选主节点。分为以下三个不同类型：
  - Hidden = false：正常的只读节点，是否可选为主，是否可投票，取决于 Priority，Vote 的值； 
  - **Hidden = true**：隐藏节点，对客户端不可见，**可以参与选举，但是 Priority 必须为 0，即不能被提升为主**。由于隐藏节点不会接受业务访问，因此**可通过隐藏节点做一些数据备份、离线计算的任务**，这并不会影响整个复制集。
  - **Delayed**：延迟节点，必须**同时具备隐藏节点和 `Priority = 0` 的特性**，会**延迟一定的时间（`secondaryDelaySecs` 配置决定）从上游复制增量**，常用于快速回滚场景。 
- **Arbiter：仲裁节点**，只用于参与选举投票，**本身不承载任何数据，只作为投票角色**。比如你部署了 2 个节点的复制集，1 个 Primary，1 个 Secondary，任意节点宕机，复制集将不能提供服务了（无法选出 Primary），这时可以给复制集添加⼀个 Arbiter 节点，即使有节点宕机，仍能选出 Primary。 Arbiter 本身不存储数据，是非常轻量级的服务，**当复制集成员为偶数时，最好加入⼀个 Arbiter 节点，以提升复制集可用性**。


#### 配置隐藏节点

很多情况下将节点设置为隐藏节点是用来协助 delayed members 的。如果仅仅需要防止该节点成为主节点，可以通过 priority 0 member 来实现。

```javascript
cfg = rs.conf()
cfg.members[1].priority = 0
cfg.members[1].hidden = true
rs.reconfig(cfg)
```

设置完毕后，该从节点的优先级将变为 0 来防止其升职为主节点，同时其也是对应用程序不可见的。在其他节点上执行 `db.isMaster()` 将不会显示隐藏节点。

#### 配置延时节点

当配置一个延时节点的时候，复制过程与该节点的 oplog 都将延时。延时节点中的数据集将会比复制集中主节点的数据延后。

```javascript
cfg = rs.conf()
cfg.members[1].priority = 0
cfg.members[1].hidden = true
// 延迟 1 分钟
cfg.members[1].secondaryDelaySecs = 60
rs.reconfig(cfg)
```

查看复制延迟：

在节点上执行 `rs.printSecondaryReplicationInfo()` 命令，可以一并列出所有备节点成员的同步延迟情况：

```bash
> rs.printSecondaryReplicationInfo()
source: 192.168.65.174:28019
{
    syncedTo: 'Fri May 19 2023 15:27:36 GMT+0800 (中国标准时间)',
    replLag: '-53 secs (-0.01 hrs) behind the primary'
}
```

延时节点**通常用于数据保护和灾难恢复场景**：

1. 人为错误防护

当管理员误删除数据或错误更新时，延时节点可以提供"时间缓冲"。在延迟时间内发现问题，可以从延时节点恢复数据。

2. 数据回滚点

提供一个人为设置的"数据快照点"，相当于一个时间机器。在应用逻辑错误导致数据污染时特别有用。

3. 灾难恢复

防范逻辑损坏（如 Bug 导致的数据错误）传播到所有节点。比常规备份更快速的恢复方式。

#### 添加投票节点

```bash
# 为仲裁节点创建数据目录，存放配置数据。该目录将不保存数据集
mkdir /data/arb
# 启动仲裁节点，指定数据目录和复制集名称
mongod --port 30000 --dbpath /data/arb --replSet rs0 
# 进入 mongo shell,添加仲裁节点到复制集
rs.addArb("ip:30000")
```

如果添加节点遇到下面的错误：

```
MongoServerError: Reconfig attempted to install a config that would change the implicit default write concern. Use the setDefaultRWConcern command to set a cluster-wide write concern and try the reconfig again.
```

```bash
# 执行命令
db.adminCommand( {"setDefaultRWConcern" : 1, "defaultWriteConcern" : { "w" : 2 } } )
```

#### 移除复制集节点

使用 `rs.remove()` 来移除节点

```bash
# 1.关闭节点实例
# 2.连接主节点，执行下面命令
rs.remove("ip:port")
```

通过 `rs.reconfig()` 来移除节点

```bash
# 1.关闭节点实例
# 2.连接主节点，执行下面命令
cfg = rs.conf()
cfg.members.splice(2,1)  #从2开始移除1个元素
rs.reconfig(cfg)
```

#### 更改复制集节点

```javascript
cfg = rs.conf()
cfg.members[0].host = "ip:port"
rs.reconfig(cfg)
```

## 复制集原理

### 高可用

#### 选举

MongoDB 的复制集选举使用 [Raft 算法](https://raft.github.io/) 来实现，**选举成功的必要条件是大多数投票节点存活**。在具体的实现中，MongoDB 对 raft 协议添加了一些自己的扩展，这包括：

- **支持 chainingAllowed 链式复制**，即**备节点不只是从主节点上同步数据，还可以选择一个离自己最近（心跳延时最小）的节点来复制数据**。因为可能两个从节点是在一个数据中心，而主节点在另一个数据中心，所以可以选择离自己最近的节点来复制数据，这样可以减少网络延时。
- **增加了预投票阶段**，即 preVote，这主要是用来避免网络分区时产生 Term (任期) 值激增的问题（**一般使用 Raft 算法的中间件都要优化这个点**）。任期激增问题是指：
  - 当网络分区时，如果一个分区只有两个节点，那么这两个节点是无法选举的，因为它们都无法获得大多数的投票。这会导致这两个节点会不停的重新发起选举，Term (任期) 值会不断增加，直到网络恢复。
  - 预投票机制就是指在网络分区时，**节点会先发起预投票，先看一下能不能拿到大多数的投票**，如果能拿到，那么就会发起正式的投票，否则就不会发起正式的投票。这样就可以避免 Term (任期) 值激增的问题。
- **支持投票优先级**，如果备节点发现自己的优先级比主节点高，则会主动发起投票并尝试成为新的主节点。例如，在同一个节点，如果主节点挂了，通过设置的优先级可以让同一个机房的备节点成为新的主节点。

**一个复制集最多可以有 50 个成员，但只有 7 个投票成员**。这是因为一旦过多的成员参与数据复制、投票过程，将会带来更多可靠性方面的问题。

**当复制集内存活的成员数量不足大多数时，整个复制集将无法选举出主节点**，此时**无法提供写服务**，这些**节点都将处于只读状态**。此外，如果希望避免平票结果的产生，最好使用奇数个节点成员，比如 3 个或 5 个。当然，在 MongoDB 复制集的实现中，对于平票问题已经提供了解决方案：

- 为选举定时器增加少量的随机时间偏差，这样避免各个节点在同一时刻发起选举，提高成功率。
- 使用仲裁者角色，该角色不做数据复制，也不承担读写业务，仅仅用来投票。

{{< callout type="info" >}}
在 Raft 协议中，任期是一个关键概念，它代表了一次选举的周期。每个节点都会记录一个任期号，任期号单调递增。当节点参与选举时，**任期号较大的节点会被优先接受为领导者，这有助于解决选举过程中的冲突问题‌**。
{{< /callout >}}



#### 自动故障转移

在故障转移场景中，我们所关心的问题是：

- **备节点是怎么感知到主节点已经发生故障的**？
- **如何降低故障转移对业务产生的影响**？

**一个影响检测机制的因素是心跳，在复制集组建完成之后，各成员节点会开启定时器，持续向其他成员发起心跳**，这里涉及的参数为 **`heartbeatIntervalMillis`，即心跳间隔时间，默认值是 2s**。如果心跳成功，则会持续以 2s 的频率继续发送心跳；如果心跳失败，则会立即重试心跳，一直到心跳恢复成功。

另一个重要的因素是**选举超时检测，一次心跳检测失败并不会立即触发重新选举**。实际上除了心跳，成员节点还会启动一个选举超时检测定时器，该定时器默认以 10s 的间隔执行，具体可以通过 `electionTimeoutMillis` 参数指定：
- 如果心跳响应成功，则取消上一次的 `electionTimeout` 调度（保证不会发起选举），并发起新一轮  `electionTimeout` 调度。
- 如果心跳响应迟迟不能成功，那么 `electionTimeout` 任务被触发，进而导致备节点发起选举并成为新的主节点。

在 MongoDB 的实现中，**选举超时检测的周期要略大于 `electionTimeoutMillis` 设定**。该周期会加入一个**随机偏移量**，大约在 `10～11.5s`，如此的设计是**为了错开多个备节点主动选举的时间，提升成功率**。

{{< callout type="info" >}}
因此，在 `electionTimeout` 任务中触发选举必须要满足以下条件：
1. 当前节点是备节点。
2. 当前节点具备选举权限。
3. 在检测周期内仍然没有与主节点心跳成功。
{{< /callout >}}

#### 业务影响评估

- **在复制集发生主备节点切换的情况下，会出现短暂的无主节点阶段，此时无法接受业务写操作（访问瞬断）**。如果是因为主节点故障导致的切换，则对于该节点的所有读写操作都会产生超时。如果使用 MongoDB 3.6 及以上版本的驱动，则可以通过**开启 `retryWrite` 来降低影响**。

```bash
# MongoDB Drivers 启用可重试写入
mongodb://localhost/?retryWrites=true
# mongo shell
mongosh --retryWrites
```

- **如果主节点属于强制掉电，那么整个 Failover 过程将会变长**，很可能需要在 Election 定时器超时后才被其他节点感知并恢复，这个时间窗口一般会在 12s 以内。然而实际上，对于业务呼损的考量还应该加上客户端或 mongos 对于复制集角色的监视和感知行为（真实的情况可能需要长达 30s 以上）。
- 对于非常重要的业务，**建议在业务层面做一些防护策略，比如设计重试机制**。

#### 优雅的重启复制集

重启复制集，不能直接把主节点干掉，因为会导致再次选举一个主节点。

如果想不丢数据重启复制集，更优雅的打开方式应该是这样的：

1. 逐个重启复制集里所有的 Secondary 节点。
2. 对 Primary 发送 `rs.stepDown()` 命令，等待 Primary 降级为 Secondary，这之后复制集会自动选出一个新的 Primary。
3. 重启降级后的 Primary。

### 数据同步机制

在复制集架构中，**主节点与备节点之间是通过 oplog 来同步数据的**，这里的 **oplog 是一个特殊的固定集合**，当主节点上的一个写操作完成后，会向 oplog 集合写入一条对应的日志，而备节点则通过这个 oplog 不断拉取到新的日志，在本地进行回放以达到数据同步的目的。

![mongodb-oplog](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-oplog.png)

#### 什么是 oplog

- MongoDB oplog 是 Local 库下的一个集合，用来**保存写操作所产生的增量日志（类似于 MySQL 中 的 Binlog）**。
- 它是一个 **Capped Collection（固定集合）**，即超出配置的最大值后，会自动删除最老的历史数据，MongoDB 针对 oplog 的删除有特殊优化，以提升删除效率。
- 主节点产生新的 oplog Entry，从节点通过复制 oplog 并应用来保持和主节点的状态一致；

#### 查看 oplog

```bash
use local
db.oplog.rs.find().sort({$natural:-1}).pretty()
```

查询结果示例：

```javascript
{
    "op": "i",
    "ns": "test.user",
    "ui": UUID("3aa4b72b-2985-4abf-a40f-fce52sjfysjf6"),
    "ts": Timestamp(1720772421, 1),
    "t": NumberLong(17),
    "v": NumberLong(2),
    "wall": ISODate("2024-04-25T09:27:01.000Z"),
    "o": {
        "_id": ObjectId("65f388906f84793487900000"),
        "name": "fox"
    }
}
```

- op：操作类型。
  - i：插入。
  - u：更新。
  - d：删除。
  - c：执行命令，如 `createDatabse`、`dropDatabse`。
  - n：无操作。
- ns：操作的数据库和集合。
- o：操作的文档。
- o2：操作查询条件
- ts：操作的时间戳。当前 `timestamp + 计数器`，计数器每秒都被重置。
- v：oplog 的版本号。

`ts` 字段描述了 oplog 产生的时间戳，可称之为 **optime**。**optime 是备节点实现增量日志同步的关键**，它保证了 oplog 是节点有序的，其由两部分组成：
- 当前的系统时间，即 UNIX 时间至现在的秒数，32 位。
- 整数计时器，不同时间值会将计数器进行重置，32 位。

optime 属于 BSON 的 Timestamp 类型，这个类型一般在 MongoDB 内部使用。既然 oplog 保证了节点级有序，那么备节点便可以通过轮询的方式进行拉取，这里会用到可持续追踪的游标（tailable cursor）技术。

**每个备节点都分别维护了自己的一个 offset，也就是从主节点拉取的最后一条日志的 optime，在执行同步时就通过这个 optime 向主节点的 oplog 集合发起查询**。为了避免不停地发起新的查询链接，在启动第一次查询后可以将 cursor 挂住（通过将 cursor 设置为 tailable）。这样只要 oplog 中产生了新的记录，备节点就能使用同样的请求通道获得这些数据。tailable cursor 只有在查询的集合为固定集合时才允许开启。

#### oplog 集合的大小

oplog 集合的大小可以通过参数 `replication.oplogSizeMB` 设置，对于 64 位系统来说，oplog 的默认值为：

```bash
oplogSizeMB = min(磁盘可用空间*5%，50GB)
```

对于大多数业务场景来说，很难在一开始评估出一个合适的 oplogSize，所幸的是 MongoDB 在 4.0 版本之后提供了 **`replSetResizeOplog` 命令，可以实现动态修改 oplogSize 而不需要重启服务器**。

```bash
# 将复制集成员的 oplog 大小修改为 60GB  
db.adminCommand({replSetResizeOplog: 1, size: 60000})
# 查看 oplog 大小
use local
db.oplog.rs.stats().maxSize
```

#### oplog 幂等性

每一条 oplog 记录都描述了一次数据的原子性变更，**对于 oplog 来说，必须保证是幂等性的**。也就是说，对于同一个 oplog，无论进行多少次回放操作，数据的最终状态都会保持不变。

某文档 `x` 字段当前值为 `100`，用户向 Primary 发送一条 `{$inc: {x: 1}}`，记录 oplog 时会转化为一条 `{$set: {x: 101}` 的操作，才能保证幂等性。

**幂等性的代价**

简单元素的操作，`$inc` 转化为 `$set` 并没有什么影响，执行开销上也差不多，但当遇到数组元素操作时，情况就不一样了。

```javascript
db.coll.insert({_id:1,x:[1,2,3]})
```

在数组尾部 push 2 个元素，查看 oplog 发现 `$push` 操作被转换为了 `$set` 操作（设置数组指定位置的元素为某个值）

```bash
rs0:PRIMARY> db.coll.update({_id: 1}, {$push: {x: { $each: [4, 5] }}})
WriteResult({ "nMatched" : 1, "nUpserted" : 0, "nModified" : 1 })
rs0:PRIMARY> db.coll.find()
{ "_id" : 1, "x" : [ 1, 2, 3, 4, 5 ] }
rs0:PRIMARY> use local
switched to db local
rs0:PRIMARY> db.oplog.rs.find({ns:"test.coll"}).sort({$natural:-1}).pretty()
{
    "op" : "u",
    "ns" : "test.coll",
    "ui" : UUID("69c871e8-8f99-4734-be5f-c9c5d8565198"),
    "o" : {
        "$v" : 1,
        "$set" : {
            "x.3" : 4,
            "x.4" : 5
        }
    },
    "o2" : {
        "_id" : 1
    },
    "ts" : Timestamp(1646223051, 1),
    "t" : NumberLong(4),
    "v" : NumberLong(2),
    "wall" : ISODate("2022-03-02T12:10:51.882Z")
}

```

`$push` 转换为带具体位置的 `$set` 开销上也差不多，但接下来再看看往数组的头部添加 2 个元素：

```bash
rs0:PRIMARY> use test
switched to db test
rs0:PRIMARY> db.coll.update({_id: 1}, {$push: {x: { $each: [6, 7], $position: 0 }}})
WriteResult({ "nMatched" : 1, "nUpserted" : 0, "nModified" : 1 })
rs0:PRIMARY> db.coll.find()
{ "_id" : 1, "x" : [ 6, 7, 1, 2, 3, 4, 5 ] }
rs0:PRIMARY> use local
switched to db local
rs0:PRIMARY> db.oplog.rs.find({ns:"test.coll"}).sort({$natural:-1}).pretty()
{
    "op" : "u",
    "ns" : "test.coll",
    "ui" : UUID("69c871e8-8f99-4734-be5f-c9c5d8565198"),
    "o" : {
        "$v" : 1,
        "$set" : {
            "x" : [
                6,
                7,
                1,
                2,
                3,
                4,
                5
            ]
        }
    },
    "o2" : {
        "_id" : 1
    },
    "ts" : Timestamp(1646223232, 1),
    "t" : NumberLong(4),
    "v" : NumberLong(2),
    "wall" : ISODate("2022-03-02T12:13:52.076Z")
}
```

可以发现，**当向数组的头部添加元素时，oplog 里的 `$set` 操作不再是设置数组某个位置的值（因为基本所有的元素位置都调整了）**，而是 `$set` 数组最终的结果，即**整个数组的内容都要写入 oplog**。当 `push` 操作指定了 `$slice` 或者 `$sort` 参数时，oplog 的记录方式也是一样的，会将整个数组的内容作为 `$set` 的参数。`$pull`,`$addToSet` 等更新操作符也是类似，更新数组后，oplog 里会转换成 `$set` 数组的最终内容，才能保证幂等性。

#### oplog 的写入被放大，导致同步追不上 - 大数组更新

当数组非常大时，对数组的一个小更新，可能就需要把整个数组的内容记录到 oplog 里，我遇到一个实际的生产环境案例，用户的文档内包含一个很大的数组字段，1000 个元素总大小在 64KB 左右（数组元素是很大的文档），这个数组里的元素按时间反序存储，新插入的元素会放到数组的最前面 (`$position: 0`)，然后保留数组的前 1000 个元素（`$slice: 1000`）。

上述场景导致，Primary 上的每次往数组里插入一个新元素(请求大概几百字节)，oplog 里就要记录整个数组的内容，Secondary 同步时会拉取 oplog 并重放，Primary 到 Secondary 同步 oplog 的流量是客户端到 Primary 网络流量的上百倍，导致主备间网卡流量跑满，而且由于 **oplog 的量太大，旧的内容很快被删除掉，最终导致 Secondary 追不上**，转换为 RECOVERING 状态。

在文档里使用数组时，一定得注意上述问题，避免数组的更新导致同步开销被无限放大的问题。使用数组时，尽量注意：

1. **数组的元素个数不要太多，总的大小也不要太大**。
2. 尽量避免对数组进行更新操作。
3. **如果一定要更新，尽量只在尾部插入元素，复杂的逻辑可以考虑在业务层面上来支持**。

#### 复制延迟

由于 oplog 集合是有固定大小的，因此存放在里面的 **oplog 随时可能会被新的记录冲掉。如果备节点的复制不够快，就无法跟上主节点的步伐，从而产生复制延迟（replication lag）问题**。这是不容忽视的，一旦备节点的**延迟过大，则随时会发生复制断裂的风险**，这意味着备节点的 optime（最新一条同步记录）已经被主节点老化掉，于是备节点将无法继续进行数据同步。

为了尽量避免复制延迟带来的风险，我们可以采取一些措施，比如：

- **增加 oplog 的容量大小**，并保持对复制窗口的监视。
- 通过一些扩展手段降低主节点的写入速度。
- 优化主备节点之间的网络。
- **避免字段使用太大的数组**（可能导致 oplog 膨胀）。

#### 主从复制数据丢失问题

由于复制延迟是不可避免的，这意味着主备节点之间的数据无法保持绝对的同步。**当复制集中的主节点宕机时，备节点会重新选举成为新的主节点。那么，当旧的主节点重新加入时，必须回滚掉之前的一些“脏日志数据”，以保证数据集与新的主节点一致**（因为旧的主节点宕机时，它的数据可能是比从节点新的）。主备复制集合的差距越大，发生大量数据回滚的风险就越高。

**对于写入的业务数据来说，如果已经被复制到了复制集的大多数节点，则可以避免被回滚的风险**。应用上可以通过设定更高的写入级别 `writeConcern：majority`（数据写入到大多数节点后才返回成功，这些节点最好在同一个机房，优先级高一点）来保证数据的持久性。类似于 Redis 的 `min-slaves-to-write` 配置。

这些由旧主节点回滚的数据会被写到单独的 rollback 目录下，必要的情况下仍然可以恢复这些数据。

当 rollback 发生时，MongoDB 将把 rollback 的数据以 BSON 格式存放到 dbpath 路径下 rollback 文件夹中，BSON 文件的命名格式如下： `<database>.<collection>.<timestamp>.bson`。

```bash
mongorestore --host 192.168.192:27018 --db test --collection emp -ufox -pfox 
--authenticationDatabase=admin rollback/emp_rollback.bson
```

#### 同步源选择

MongoDB 是**允许通过备节点进行复制**的，这会发生在以下的情况中：

- **在 `settings.chainingAllowed` 开启的情况下，备节点自动选择一个最近的节点（ping 命令时延最小）进行同步**。`settings.chainingAllowed` 选项默认是开启的，也就是说默认情况下备节点并不一定会选择主节点进行同步，这个**副作用就是会带来延迟的增加**，你可以通过下面的操作进行关闭：

```javascript
cfg = rs.config()
cfg.settings.chainingAllowed = false
rs.reconfig(cfg)
```

- **使用 `replSetSyncFrom` 命令临时更改当前节点的同步源**，比如在初始化同步时将同步源指向备节点来降低对主节点的影响。

```javascript
db.adminCommand({ replSetSyncFrom: "hostname:port" })
```