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

![mongodb-pss]()

此模式始终提供两个完整的副本，如果 Primary 节点不可用，则复制集会自动选择一个 Secondary 节点作为新的 Primary 节点。旧的 Primary 节点在可用时重新加入复制集。

### PSA 模式

PSA 模式由一个 Primary 节点、一个 Secondary 节点和**一个仲裁者节点（Arbiter）**组成。

![mongodb-psa]()

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
rs0:SECONDARY> rs.secondaryOk() # 这个方法以弃用，使用 db.getMongo().setReadPref("primaryPreferred") 代替
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

**方式二（强烈推荐）**：通过高可用 Uri 的方式连接 MongoDB，当 Primary 故障切换后，MongoDB Driver 可自动感知并把流量路由到新的 Primary 节点。

```bash
mongosh mongodb://fox:fox@192.168.65.206:28017,192.168.65.206:28018,192.168.65.206:28019/admin?replicaSet=rs0
```

## 复制集成员角色 

复制集里面有多个节点，每个节点拥有不同的职责。 
在看成员角色之前，先了解两个重要属性： 

### Priority = 0 

**当 Priority 等于 0 时，它不可以被复制集选举为主，Priority 的值越高，则被选举为主的概率更大**。通常，在跨机房方式下部署复制集可以使用该特性。假设使用了机房 A 和机房 B，由于主要业务与机房 A 更近，则可以将机房 B 的复制集成员 Priority 设置为 0，这样主节点就一定会是 A 机房的成员。 

### Vote = 0

**Vote 等于 0 时，不可以参与选举投票，此时该节点的 Priority 也必须为 0，即它也不能被选举为主**。由于**一个复制集中最多只有 7 个投票成员**，因此多出来的成员则必须将其 vote 属性值设置为 0，即这些成员将无法参与投票。

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


### 数据同步机制

#### 数据回滚