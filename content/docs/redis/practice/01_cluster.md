---
title: 集群
weight: 1
---

## CAP 原理

- C - Consistent ，一致性
- A - Availability ，可用性
- P - Partition tolerance ，分区容忍性

分布式系统的节点往往都是分布在不同的机器上进行网络隔离开的，这意味着必然会有网络断开的风险，这个网络断开的场景的专业词汇叫着**网络分区**。

在网络分区发生时，两个分布式节点之间无法进行通信，我们对一个节点进行的修改操作将无法同步到另外一个节点，所以数据的**一致性**将无法满足，因为两个分布式节点的数据不再保持一致。除非我们牺牲**可用性**，也就是暂停分布式节点服务，在网络分区发生时，不再提供修改数据的功能，直到网络状况完全恢复正常再继续对外提供服务。

一句话概括 CAP 原理就是——网络分区发生时，一致性和可用性两难全。

### 最终一致性

Redis 保证**最终一致性**，从节点会努力追赶主节点，最终从节点的状态会和主节点的状态将保持一致。如果网络断开了，主从节点的数据将会出现大量不一致，一旦网络恢复，从节点会采用多种策略努力追赶上落后的数据，继续尽力保持和主节点一致。


## 主从同步 

Redis 支持**主从同步**和**从从同步**，从从同步功能是 Redis 后续版本增加的功能，为了减轻主库的同步负担。

### Redis 主从架构搭建

```
1、复制一份redis.conf文件

2、将相关配置修改为如下值：
port 6380
pidfile /var/run/redis_6380.pid  # 把 pid 进程号写入 pidfile 配置的文件
logfile "6380.log"
dir /usr/local/redis-5.0.3/data/6380  # 指定数据存放目录
# 需要注释掉 bind
# bind 127.0.0.1（bind 绑定的是自己机器网卡的 ip，如果有多块网卡可以配多个 ip，代表允许客户端通过机器的哪些网卡 ip 去访问，内网一般可以不配置 bind，注释掉即可）

3、配置主从复制
replicaof 192.168.0.60 6379   # 从本机 6379 的 Redis 实例复制数据，Redis 5.0之前使用 slaveof
replica-read-only yes  # 配置从节点只读

4、启动从节点
redis-server redis.conf   # redis.conf 文件务必用你复制并修改了之后的 redis.conf 文件

5、连接从节点
redis-cli -p 6380

6、测试在 6379 实例上写数据，6380 实例是否能及时同步新修改数据

7、可以自己再配置一个 6381 的从节点
```

### 主从同步原理

1. 如果你为 master 配置了一个 slave，不管这个 slave 是否是第一次连接上 master，它都会发送一个 PSYNC 命令给 master 请求复制数据。
2. master 收到 PSYNC 命令后，会在后台进行数据持久化通过 bgsave 生成最新的 rdb 快照文件（这里的 rdb 与开不开启 rdb 持久化没有关系），持久化期间，master 会继续接收客户端的请求，它会把这些可能**修改数据集的请求缓存在内存中**。
3. 当持久化进行完毕以后，master 会把这份 rdb 文件数据集发送给 slave。
4. slave 会把接收到的数据进行持久化生成 rdb，然后再加载到内存中。
5. 然后，master 再将之前缓存在内存中的命令发送给 slave。
6. 当 master 与 slave 之间的连接由于某些原因而断开时，slave 能够自动重连 master，如果 master 收到了多个 slave 并发连接请求，它**只会进行一次持久化**，而不是一个连接一次，然后再把这一份持久化的数据发送给多个并发连接的 slave。

**为什么不使用 AOF 来做数据同步？**

因为 RDB 更快。

#### 部分复制

就是说一个 slave 之前连接了 master，已经有部分数据了，后面又和 master 断开了连接，然后又重新连接上 master，master 会把断开连接期间修改的数据发送给 slave。

master 会在其内存中创建一个复制数据用的缓存队列，缓存最近一段时间的数据，master 和它所有的 slave 都维护了复制的数据下标 offset 和 master 的进程 ID，因此，当网络连接断开重连后，slave 会请求 master 继续进行未完成的复制，从所记录的数据下标开始。如果 master 进程 ID 变化了，或者从节点数据下标 offset 太旧，已经不在 master 的缓存队列里了，那么将会进行一次全量数据的复制。

### 从从同步

从从同步是从 Redis 3.0 开始支持的功能，它的出现主要是为了分担主节点的同步压力，在主从同步中，从节点也可以作为其他从节点的主节点，从而形成一个树状结构。为了缓解**主从复制风暴** (多个从节点同时复制主节点导致主节点压力过大)。

![redis-master-slave](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redis-master-slave.png)

## Sentinel 哨兵架构

Redis 主从架构虽然可以实现数据的高可用，但是当主节点挂掉后，需要手动将从节点提升为主节点，这是一个比较麻烦的过程。

为了解决这个问题，Redis 引入了 Sentinel 哨兵架构。

Sentinel 是一个分布式架构，它由多个 Sentinel 实例组成，每个 Sentinel 实例都可以监控多个主从节点，它会持续监控主从节点的健康，当主节点挂掉后，它会自动选择一个最优的从节点切换为主节点。

![redis-sentinel](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redis-sentinel.png)

Sentinel 哨兵是特殊的 Redis 服务，**不提供读写服务，主要用来监控 Redis 实例节点**。

哨兵架构下 **Client 第一次要从哨兵获取到 Redis 的主节点，后续就直接访问 Redis 的主节点，不会每次都通过 Sentinel 代理访问 Redis 的主节点**。当 Redis 的主节点发生变化，哨兵会第一时间感知到，并且将新的 Redis
主节点通知给 Client 端 (这里面 Redis 的 Client 一般都实现了订阅功能，订阅 Sentinel 发布的节点变动消息)。

### Redis 哨兵架构搭建

```
1、复制一份 sentinel.conf 文件
cp sentinel.conf sentinel-26379.conf

2、将相关配置修改为如下值：
port 26379
daemonize yes
pidfile "/var/run/redis-sentinel-26379.pid"
logfile "26379.log"
dir "/usr/local/redis-5.0.3/data"
# sentinel monitor <master-redis-name> <master-redis-ip> <master-redis-port> <quorum>
# quorum 是一个数字，指明当有多少个 sentinel 认为一个 master 失效时(值一般为：sentinel总数/2 + 1)，master 才算真正失效
sentinel monitor mymaster 192.168.0.60 6379 2   # mymaster 这个名字随便取，客户端访问时会用到

3、启动 Sentinel 哨兵实例
src/redis-sentinel sentinel-26379.conf

4、查看 Sentinel 的 info 信息
src/redis-cli -p 26379
127.0.0.1:26379>info

可以看到 Sentinel 的 info 里已经识别出了 Redis 的主从

5、可以再配置两个 Sentinel，端口 26380 和 26381，注意上述配置文件里的对应数字都要修改
```

Sentinel 集群都启动完毕后，会将哨兵集群的元数据信息写入所有 Sentinel 的配置文件里去(追加在文件的最下面)，查看下如下配置文件 `sentinel-26379.conf`，如下所示：

```bash
sentinel known-replica mymaster 192.168.0.60 6380 # 代表 Redis 主节点的从节点信息
sentinel known-replica mymaster 192.168.0.60 6381 # 代表 Redis 主节点的从节点信息
sentinel known-sentinel mymaster 192.168.0.60 26380 52d0a5d70c1f90475b4fc03b6ce7c3c56935760f  # 代表感知到的其它哨兵节点
sentinel known-sentinel mymaster 192.168.0.60 26381 e9f530d3882f8043f76ebb8e1686438ba8bd5ca6  # 代表感知到的其它哨兵节点
```

当 Redis 主节点如果挂了，哨兵集群会重新选举出新的 Redis 主节点，同时会修改所有 Sentinel 节点配置文件的集群元数据信息，比如 6379 的 Redis 如果挂了，假设选举出的新主节点是 6380，则 Sentinel 文件里的集群元数据信息会变成如下所示：

```bash
sentinel known-replica mymaster 192.168.0.60 6379 # 代表主节点的从节点信息
sentinel known-replica mymaster 192.168.0.60 6381 # 代表主节点的从节点信息
sentinel known-sentinel mymaster 192.168.0.60 26380 52d0a5d70c1f90475b4fc03b6ce7c3c56935760f  # 代表感知到的其它哨兵节点
sentinel known-sentinel mymaster 192.168.0.60 26381 e9f530d3882f8043f76ebb8e1686438ba8bd5ca6  # 代表感知到的其它哨兵节点
```

同时还会修改 Sentinel 文件里之前配置的 mymaster 对应的 6379 端口，改为 6380

```bash
sentinel monitor mymaster 192.168.0.60 6380 2
```

当 6379 的 Redis 实例再次启动时，哨兵集群根据集群元数据信息就可以将 6379 端口的 Redis 节点作为从节点加入集群。

### 消息丢失

Redis 主从采用异步复制，意味着当主节点挂掉时，从节点可能没有收到全部的同步消息，这部分未同步的消息就丢失了。如果主从延迟特别大，那么丢失的数据就可能会特别多。Sentinel 无法保证消息完全不丢失，但是也尽可能保证消息少丢失。它有两个选项可以限制主从延迟过大。

```
min-slaves-to-write 1  # 表示主节点必须至少有一个从节点在进行正常复制，否则就停止对外写服务
min-slaves-max-lag 10 # 单位是秒，表示如果 10s 没有收到从节点的反馈，就意味着从节点同步异常
```
## Redis Cluster

Redis 3.0 以前的版本要实现集群一般是借助哨兵来监控 master 节点的状态，如果 master 节点异常，则会做主从切换，将某一台 slave 作为 master，哨兵的配置略微复杂，并且性能和高可用性等各方面表现一般，特别是在主从切换的瞬间存在**访问瞬断**的情况，而且**哨兵模式只有一个主节点对外提供服务**，没法支持很高的并发，且**单个主节点内存也不宜设置得过大，否则会导致持久化文件过大，影响数据恢复或主从同步的效率**，一般推荐小于 10G。

### Redis Cluster 架构

Redis Cluster 是 Redis 官方提供的分布式集群方案。

![redis-cluster](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redis-cluster.png)

Redis 集群是一个由**多个主从节点**组成的分布式服务器群，它具有复制、高可用和分片特性。Redis 集群不需要哨兵也能完成节点移除和故障转移的功能。需要将每个节点设置成集群模式，这种集群模式**没有中心节点**，可水平扩展，据官方文档称可以线性扩展到上万个节点 (**官方推荐不超过 1000 个节点**)。Redis 集群的性能和高可用性均优于之前版本的哨兵模式，且集群配置非常简单。

### Redis 集群搭建

Redis 集群需要至少三个 master 节点，这里搭建三个 master 节点，并且给每个 master 再搭建一个 slave 节点，总共 6 个 Redis 节点，这里用三台机器部署 6 个 Redis 实例，每台机器一主一从，搭建集群的步骤如下：

```
第一步：在第一台机器的 /usr/local 下创建文件夹 redis-cluster，然后在其下面分别创建 2 个文件夾如下
（1）mkdir -p /usr/local/redis-cluster
（2）mkdir 8001 8004

第一步：把之前的 redis.conf 配置文件 copy 到 8001 下，修改如下内容：
（1）daemonize yes
（2）port 8001（分别对每个机器的端口号进行设置）
（3）pidfile /var/run/redis_8001.pid  # 把 pid 进程号写入 pidfile 配置的文件
（4）dir /usr/local/redis-cluster/8001/（指定数据文件存放位置，必须要指定不同的目录位置，不然会丢失数据）
（5）cluster-enabled yes（启动集群模式）
（6）cluster-config-file nodes-8001.conf（集群节点信息文件，这里 800x 最好和 port 对应上）
（7）cluster-node-timeout 10000
 (8)# bind 127.0.0.1（bind 绑定的是自己机器网卡的 ip，如果有多块网卡可以配多个 ip，代表允许客户端通过机器的哪些网卡 ip 去访问，内网一般可以不配置 bind，注释掉即可）
 (9) protected-mode  no   （关闭保护模式）
 (10) appendonly yes
如果要设置密码需要增加如下配置：
 (11) requirepass zhuge     (设置 Redis 访问密码)
 (12) masterauth zhuge      (设置集群节点间访问密码，跟上面一致)

第三步：把修改后的配置文件，copy 到 8004，修改第 2、3、4、6 项里的端口号，可以用批量替换：
:%s/源字符串/目的字符串/g 

第四步：另外两台机器也需要做上面几步操作，第二台机器用 8002 和 8005，第三台机器用 8003 和 8006

第五步：分别启动 6 个 Redis 实例，然后检查是否启动成功
（1）/usr/local/redis-5.0.3/src/redis-server /usr/local/redis-cluster/800*/redis.conf
（2）ps -ef | grep redis # 查看是否启动成功
    
第六步：用 redis-cli 创建整个 Redis 集群( Redis 5 以前的版本集群是依靠 ruby 脚本 redis-trib.rb 实现)
# 执行这条命令需要确认三台机器之间的 Redis 实例要能相互访问，可以先简单把所有机器防火墙关掉，如果不关闭防火墙则需要打开 Redis 服务端口和集群节点 gossip 通信端口 16379 (默认是在 Redis 端口号上加 10000)
# 关闭防火墙
# systemctl stop firewalld # 临时关闭防火墙
# systemctl disable firewalld # 禁止开机启动
# 注意：下面这条创建集群的命令大家不要直接复制，里面的空格编码可能有问题导致创建集群不成功
（1）/usr/local/redis-5.0.3/src/redis-cli -a zhuge --cluster create --cluster-replicas 1 192.168.0.61:8001 192.168.0.62:8002 192.168.0.63:8003 192.168.0.61:8004 192.168.0.62:8005 192.168.0.63:8006 

# --cluster-replicas 1 表示为每个创建的主服务器节点创建一个从服务器节点，对于这里 6 个节点来说，就是 3 主 3 从

第七步：验证集群：
（1）连接任意一个客户端即可：./redis-cli -c -h -p (-a 访问服务端密码，-c 表示集群模式，指定 ip 地址和端口号）
    如：/usr/local/redis-5.0.3/src/redis-cli -a zhuge -c -h 192.168.0.61 -p 800*
（2）进行验证： cluster info（查看集群信息）、cluster nodes（查看节点列表）
（3）进行数据操作验证
（4）关闭集群则需要逐个进行关闭，使用命令：
/usr/local/redis-5.0.3/src/redis-cli -a zhuge -c -h 192.168.0.60 -p 800* shutdown
```

![cluster-slots](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/cluster-slots.png)

其中 `slots` 就是分配给每个节点的槽位，只有主节点才会分配槽位。

![redis-cluster-nodes](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redis-cluster-nodes.png)

前面创建集群时，8001 和 8004 是在一个节点上的，8002 和 8005 是在一个节点上的，8003 和 8006 是在一个节点上的，但是上面的节点列表中，8001 是主节点，它的从节点却是 8005，8002 的从节点是 8006，8003 的从节点是 8004，这是为什么？

因为更加安全，避免一个节点挂了导致小的主从集群不可用。

`cluster-config-file nodes-8001.conf` 集群创建好以后，整个集群节点的信息会被保存到这个配置文件中。

**为什么要保存到这个文件中？**

因为如果整个集群如果关掉了，再次启动的时候是不能再使用 `--cluster create` 命令的，只需要把每个节点的 Redis 重新启动即可。Redis 启动的时候会读取这个配置文件中的节点信息，然后再重新组件集群。

### Redis 集群原理

Redis Cluster 将所有数据划分为 16384 个 slots(槽位)，每个节点负责其中一部分槽位。槽位的信息存储于每个节点中。当 Redis Cluster 的客户端来连接集群时，它也会得到一份集群的槽位配置信息并将其缓存在客户端本地。这样当客户端要查找某个 key 时，可以直接定位到目标节点。同时因为槽位的信息可能会存在客户端与服务器不一致的情况，还需要纠正机制来实现槽位信息的校验调整。

**哨兵架构访问瞬断**的问题在集群中也没有完全解决，但是因为集群中的**数据是分散存储在多个节点上的**，所以当客户端访问某个节点时，如果这个节点挂了，并不会影响其他节点的数据。**只有这个集群内的小的主从集群会出现访问瞬断的情况**。

#### 槽位定位算法

Cluster 默认会对 key 值使用 crc16 算法进行 hash 得到一个整数值，然后用这个整数值对 16384 进行取模来得到具体槽位。

Cluster 还允许用户强制某个 key 挂在特定槽位上，通过在 key 字符串里面嵌入 tag 标记，这就可以强制 key 所挂在的槽位等于 tag 所在的槽位。

#### 跳转重定向

当客户端向一个错误的节点发出了指令，该节点会发现指令的 key 所在的槽位并不归自己管理，这时它会向客户端发送一个特殊的跳转指令携带目标操作的节点地址，告诉客户端去连这个节点去获取数据。

```sh
GET x
-MOVED 3999 127.0.0.1:6381
```

`MOVED` 指令的第一个参数 `3999` 是 key 对应的槽位编号，后面是目标节点地址。`MOVED` 指令前面有一个减号，表示该指令是一个错误消息。

客户端收到 `MOVED` 指令后，要立即纠正本地的槽位映射表。后续所有 key 将使用新的槽位映射表。

#### Redis 集群节点间的通信机制

Redis cluster 节点间采取 gossip 协议进行通信维护集群的元数据 (集群节点信息，主从角色，节点数量，各节点共享的数据等) 有两种方式：**集中式**和 **gossip**

- 集中式：
优点在于元数据的更新和读取，时效性非常好，一旦元数据出现变更立即就会更新到集中式的存储中，其他节点读取的时候立即就可以立即感知到；不足在于所有的元数据的更新压力全部集中在一个地方，可能导致元数据的存储压力。很多中间件都会借助 zookeeper 集中式存储元数据。

- gossip：

gossip 协议包含多种消息，包括 ping，pong，meet，fail 等等。 
meet：某个节点发送 meet 给新加入的节点，让新节点加入集群中，然后新节点就会开始与其他节点进行通信；
ping：每个节点都会频繁给其他节点发送 ping，其中包含自己的状态还有自己维护的集群元数据，互相通过 ping 交换元数据(类似自己感知到的集群节点增加和移除，hash slot 信息等)； 
pong: 对 ping 和 meet 消息的返回，包含自己的状态和其他信息，也可以用于信息广播和更新； 
fail: 某个节点判断另一个节点 fail 之后，就发送 fail 给其他节点，通知其他节点，指定的节点宕机了。

gossip 协议的优点在于元数据的更新比较分散，不是集中在一个地方，更新请求会陆陆续续，打到所有节点上去更新，有一定的延时，降低了压力；缺点在于元数据更新有延时可能导致集群的一些操作会有一些滞后。

**gossip 通信的端口**：
 
每个节点都有一个专门用于节点间 gossip 通信的端口，就是自己提供服务的 `端口号+10000`，比如 7001，那么用于节点间通信的就是 17001 端口。每个节点每隔一段时间都会往另外几个节点发送 ping 消息，同时其他几点接收到 ping 消息之后返回 pong 消息。

这也是为什么不推荐集群的节点超过 1000 个的原因，因为**集群内部节点的心跳通知非常频繁，这对网络带宽是一个非常大的消耗**。

#### 网络抖动

真实世界的机房网络往往并不是风平浪静的，它们经常会发生各种各样的小问题。比如网络抖动就是非常常见的一种现象，突然之间部分连接变得不可访问，然后很快又恢复正常。

为解决这种问题，Redis Cluster 提供了一种选项 `cluster-node-timeout`，表示**当某个节点持续 timeout 的时间失联时，才可以认定该节点出现故障，需要进行主从切换**。如果没有这个选项，网络抖动会导致主从频繁切换 (数据的重新复制)。

#### Redis 集群选举原理

当 slave 发现自己的 master 变为 FAIL 状态时，便尝试进行 Failover，以期成为新的 master。由于挂掉的 master 可能会有多个 slave，从而存在多个 slave 竞争成为 master 节点的过程，其过程如下：

1. slave 发现自己的 master 变为 FAIL
2. 将自己记录的集群 currentEpoch 加 1，并广播 `FAILOVER_AUTH_REQUEST` 信息
3. 其他节点收到该信息，只有 master 响应，判断请求者的合法性，并发送 `FAILOVER_AUTH_ACK`，对每一个 epoch 只发送一次 ack。
4. 尝试 failover 的 slave 收集 master 返回的 `FAILOVER_AUTH_ACK`。
5. slave 收到超过半数 master 的 ack 后变成新 master (这里解释了集群为什么至少需要三个主节点，**如果只有两个，当其中一个挂了，只剩一个主节点是不能选举成功的**)。
6. slave 广播 Pong 消息通知其他集群节点。

**为了避免多个从节点在选举获得的票数一样**：

从节点并不是在主节点一进入 FAIL 状态就马上尝试发起选举，而是有一定延迟，一定的延迟确保我们等待 FAIL 状态在集群中传播，slave 如果立即尝试选举，其它 masters 或许尚未意识到 FAIL 状态，可能会拒绝投票。

**延迟计算公式**：`DELAY = 500ms + random(0 ~ 500ms) + SLAVE_RANK * 1000ms`。

`SLAVE_RANK` 表示此 slave 已经从 master 复制数据的总量的 rank。Rank 越小代表已复制的数据越新。这种方式下，持有最新数据的 slave 将会首先发起选举（理论上）。

### 集群脑裂数据丢失问题

脑裂数据丢失问题，网络分区导致脑裂后多个主节点对外提供写服务，一旦网络分区恢复，会将其中一个主节点变为从节点，这时就会有数据丢失。

![redis-brain-split](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redis-brain-split.png)

Redis 可以通过配置 `min-slaves-to-write` 参数来规避脑裂数据丢失问题 （这种方法不可能百分百避免数据丢失，参考集群 master 选举机制）：

```
// 写数据成功最少同步的 slave 数量，这个数量可以模仿大于半数机制配置，比如集群总共三个节点可以配置 1，加上 master 就是 2，超过了半数（也就是说至少要有一个从节点同步成功之后，才会返回客户端写入成功）。
// 该参数在 Redis 最新版本里名字已经换成了 min-replicas-to-write。
min-slaves-to-write 1  
```

{{< callout type="info" >}}
这个配置在一定程度上会影响集群的可用性，比如 slave 要是少于 1 个，这个集群就算 master 正常也不能提供服务了，需要具体场景权衡选择。一般情况下可以不用考虑这个配置，可用性是更重要的，丢一点缓存数据是可以接受的。如果因为 Redis 不可用，导致大量请求打到数据库，数据库可能会直接挂掉，这是无法接受的。
{{< /callout >}}


### 集群是否完整才能对外提供服务

当 `redis.conf` 的配置 `cluster-require-full-coverage` 为 `no` 时，表示当负责一个插槽的 master 节点下线且没有相应的 slave 节点进行故障恢复时，集群仍然可用，如果为 `yes` 则集群不可用。

### Redis 集群为什么至少需要三个 master 节点，并且推荐节点数为奇数？

因为新 master 的选举需要大于半数的集群 master 节点同意才能选举成功，如果只有两个 master 节点，当其中一个挂了，是达不到选举新 master 的条件的。

奇数个 master 节点可以在满足选举该条件的基础上节省一个节点，比如三个 master 节点和四个 master 节点的集群相比，如果都挂了一个 master 节点，三个 master 节点的集群只需要两个节点就可以选举，而四个 master 节点的集群需要三个节点（过半）才能选举 。所以三个 master 节点和四个 master 节点的集群都只能挂一个节点。如果都挂了两个 master 节点都没法选举新 master 节点了，所以奇数的master节点更多的是从节省机器资源角度出发说的。


### Redis 集群对批量操作命令的支持

对于 Redis 集群，批量操作命令一定要在集群上操作，因为在集群中多个 key 可能是不同的 master 节点的 slot 上，或者在同一个 master 节点的不同的 slot 上，**客户端**会直接返回错误。这是由于 Redis 要保证批量操作命令的原子性，要么全部成功，要么全部失败。在不同的 master 节点上操作，如果其中的一个 master 节点挂了，会导致有些 key 写入成功，有些 key 写入失败，这就破坏了原子性。

为了解决这个问题，则可以在 key 的前面加上 `{XX}`，这样参数数据分片 hash 计算的只会是大括号里的值，这样能确保不同的 key 能落到同一 slot 里去，示例如下：

```
mset {user1}:1:name zhuge {user1}:1:age 18
```

假设 name 和 age 计算的 hash slot 值不一样，但是这条命令在集群下执行，Redis 只会用大括号里的 `user1` 做 hash slot 计算，所以算出来的 slot 值肯定相同，最后都能落在同一 slot。


### 为什么是 16384 个槽位？

1. 如果槽位为 65536，发送心跳信息的消息头达 8KB，发送的心跳包过于庞大。

当槽位为 65536 时，这块的大小是 `65536 / 8 / 1024= 8kb`。因为每秒钟，Redis 节点需要发送一定数量的 ping 消息作为心跳包，如果槽位为 65536，这个 ping 消息的消息头太大了，会导致网络拥堵。

2. Redis 的集群主节点数量官方建议不超过 1000 个。

集群节点越多，心跳包的消息体内携带的数据越多。如果节点过 1000 个，也会导致网络拥堵。因此官方不建议 Redis cluster 节点数量超过 1000 个。那么，对于节点数在 1000 以内的 Redis cluster 集群，16384 个槽位够用了。没有必要拓展到 65536 个。

3. 槽位越小，节点少的情况下，压缩率高

Redis 主节点的配置信息中，它所负责的哈希槽是通过一张 bitmap 的形式来保存的，在传输过程中，会对 bitmap 进行压缩，但是如果 bitmap 的填充率 `slots / N` 很高的话(`N` 表示节点数)，bitmap 的压缩率就很低。如果节点数很少，而哈希槽数量很多的话，bitmap 的压缩率就很低。


## Redis 7.0

Redis 7.0 对主从复制进行了优化，性能有了很大的提升。

### Redis 7.0 之前的主从复制的问题

#### 多从库时主库内存占用过多

Redis 的主从复制主要分为两步：

- **全量同步**，主库通过 `fork` 子进程产生内存快照，然后将数据序列化为 RDB 格式同步到从库，使从库的数据与主库某一时刻的数据一致。
- **命令传播**：全量同步期间，master 会继续接收客户端的请求，它会把这些可能**修改数据集的请求缓存在内存中**。当从库与主库完成全量同步后，进入命令传播阶段，主库将变更数据的命令发送到从库，从库将执行相应命令，使从库与主库数据持续保持一致。

![redis-replication]()

**复制积压区**，可以理解为是一个备份，因为主从复制的过程中，如果从库的连接突然断开了，那么从库对应的**从库复制缓冲区**会被释放掉，包括其他的网络资源。等到从库重新连接时，重新开始复制，就刻意从复制积压区找到断开连接时数据复制的位置，从这个断开的位置开始继续复制。

如上图所示，对于 Redis 主库，当用户的写请求到达时，主库会将变更命令分别写入所有**从库复制缓冲区**（OutputBuffer)，以及**复制积压区** (ReplicationBacklog)。

该实现一个明显的问题是内存占用过多，所有从库的连接在主库上是独立的，也就是说**每个从库 OutputBuffer 占用的内存空间也是独立的**，那么**主从复制消耗的内存就是所有从库缓冲区内存大小之和**。如果我们设定从库的 `client-output-buffer-limit` 为 1GB，如果有三个从库，则在主库上可能会消耗 3GB 的内存用于主从复制。另外，真实环境中从库的数量不是确定的，这也导致 Redis 实例的内存消耗不可控。

{{< callout type="info" >}}
当全量复制的时间过长或者 `client-output-buffer-limit` 设置的 buffer 过小，会导致增量的指令在 buffer 中被覆盖，导致全量复制后无法进行增量复制，然后会再次发起快照同步，如此极有可能会陷入快照同步的死循环。
{{< /callout >}}

#### OutputBuffer 拷贝和释放的堵塞问题

Redis 为了提升多从库全量复制的效率和减少 fork 产生 RDB 的次数，会尽可能的让多个从库共用一个 RDB，从代码 (`replication.c`) 上看：

![redis-copy-output-buffer]()

当已经有一个从库触发 RDB BGSAVE 时，后续需要全量同步的从库会共享这次 BGSAVE 的 RDB，为了从库复制数据的完整性，会将第一个触发 RDB BGSAVE 从库的 OutputBuffer 拷贝到后续请求全量同步从库的 OutputBuffer 中。

代码中的 `copyClientOutputBuffer` 可能存在堵塞问题，因为 OutputBuffer 链表上的数据可达数百 MB 甚至数 GB 之多，对其拷贝的耗时可能达到百毫秒甚至秒级的时间，而且该堵塞问题没法通过日志或者 latency 观察到，但对 Redis 性能影响却很大，甚至造成 Redis 阻塞。

同样地，当 OutputBuffer 大小触发 limit 限制时，Redis 就是关闭该从库链接，而在释放 OutputBuffer 时，也需要释放数百 MB 甚至数 GB 的数据，其耗时对 Redis 而言也很长。

而且如果重新设置 ReplicationBacklog 大小时，Redis 会重新申请一块内存，然后将 ReplicationBacklog 中的内容拷贝过去，这也是非常耗时的操作。

#### ReplicationBacklog 的限制

复制积压缓冲区 ReplicationBacklog 是 Redis 实现部分重同步的基础，如果从库可以进行增量同步，则主库会从 ReplicationBacklog 中拷贝从库缺失的数据到其 OutputBuffer。拷贝的数据量最大当然是 ReplicationBacklog 的大小，为了避免拷贝数据过多的问题，通常不会让该值过大，一般百兆左右。但在大容量实例中，为了避免由于主从网络中断导致的全量同步，又希望该值大一些，这就存在矛盾了。

#### Redis 7.0 主从复制的优化

每个从库都有自己的 OutputBuffer，但其存储的内容却是一样的，一个最直观的想法就是主库在命令传播时，将这些命令放在一个全局的复制数据缓冲区中，多个从库共享这份数据。复制积压缓冲区（ReplicationBacklog）中的内容与从库 OutputBuffer 中的数据也是一样的，所以该方案中，ReplicationBacklog 和从库一样共享一份复制缓冲区的数据，也避免了 ReplicationBacklog 的内存开销。

**共享复制缓存区**方案中复制缓冲区 (ReplicationBuffer) 的表示采用**链表**的表示方法，将 ReplicationBuffer 数据切割为多个 16KB 的数据块 (`replBufBlock`)，然后使用链表来维护起来。为了维护不同从库的对 ReplicationBuffer 的使用信息，在 `replBufBlock` 中存在字段：

- `refcount`：block 被引用的次数。
- `id`：block 的 id。
- `repl_offset`：block 中数据的偏移量。

ReplicationBuffer 由多个 `replBufBlock` 组成链表，当**复制积压区**或从库对某个 block 使用时，便对正在使用的 `replBufBlock` 增加引用计数，上图中可以看到，复制积压区正在使用的 replBufBlock `refcount` 是 1，从库 A 和 B 正在使用的 `replBufBlock` 的 `refcount` 是 2。当从库使用完当前的 `replBufBlock`（已经将数据发送给从库）时，就会对其 `refcount` 减 1 而且移动到下一个 `replBufBlock`，并对其 `refcount` 加 1。

##### 堵塞问题和限制问题的解决

多从库消耗内存过多的问题通过共享复制缓存区方案得到了解决，对于 OutputBuffer 拷贝和释放的堵塞问题和 ReplicationBacklog 的限制问题是否解决了？

首先来看 OutputBuffer 拷贝和释放的堵塞问题问题，这个问题很好解决，因为 ReplicationBuffer 是个链表实现，当前从库的 OutputBuffer 只需要维护共享 ReplicationBuffer 的引用信息即可。所以无需进行数据深拷贝，只需要更新引用信息，即对正在使用的 `replBufBlock` 的 `refcount` 加 1，这仅仅是一条简单的赋值操作，非常轻量。

OutputBuffer 释放问题呢？在当前的方案中释放从库 OutputBuffer 就变成了对其正在使用的 `replBufBlock` 的 `refcount` 减 1，也是一条赋值操作，不会有任何阻塞。

对于 ReplicationBacklog 的限制问题也很容易解决了，因为 ReplicatonBacklog 也只是记录了对 ReplicationBuffer 的引用信息，对 ReplicatonBacklog 的拷贝也仅仅成了找到正确的 `replBufBlock`，然后对其 `refcount` 加 1。这样的话就不用担心 ReplicatonBacklog 过大导致的拷贝堵塞问题。而且对 ReplicatonBacklog 大小的变更也仅仅是配置的变更，不会清掉数据。


##### ReplicationBuffer 的裁剪和释放

ReplicationBuffer 不可能无限增长，Redis 有相应的逻辑对其进行裁剪，简单来说，Redis 会从头访问 `replBufBlock` 链表，如果发现 `replBufBlock` 的 `refcount` 为 0，则会释放它，直到迭代到第一个 `replBufBlock` 的 `refcount` 不为 0 才停止。所以想要释放 ReplicationBuffer，只需要减少相应 `replBufBlock` 的 `refcount`，会减少 `refcount` 的主要情况有：

1. 当从库使用完当前的 `replBufBlock` 会对其 `refcount` 减 1；
2. 当从库断开链接时会对正在引用的 `replBufBlock` 的 `refcount` 减 1，无论是因为超过 `client-output-buffer-limit` 导致的断开还是网络原因导致的断开；
3、当 ReplicationBacklog 引用的 `replBufBlock` 数据量超过设置的该值大小时，会对正在引用的 `replBufBlock` 的 `refcount` 减 1，以尝试释放内存；

不过当一个从库引用的 `replBufBlock` 过多，它断开时释放的 `replBufBlock `可能很多，也可能造成堵塞问题，所以 Redis7 里会限制一次释放的个数，未及时释放的内存在系统的定时任务中渐进式释放。

##### 数据结构的选择

当从库尝试与主库进行增量重同步时，会发送自己的 `repl_offset`，主库在每个 `replBufBlock` 中记录了该其第一个字节对应的 `repl_offset`，但如何高效地从数万个 `replBufBlock` 的链表中找到特定的那个？

链表只能直接从头到位遍历链表查找对应的 `replBufBlock`，这个操作必然会耗费较多时间而堵塞服务。

Redis 7 使用 rax 树实现了对 `replBufBlock` 固定区间间隔的索引，每 64 个记录一个索引点。一方面，rax 索引占用的内存较少；另一方面，查询效率也是非常高，理论上查找比较次数不会超过 100，耗时在 1 毫秒以内。


##### RAX 树

Redis 中还有其他地方使用了 Rax 树，比如 streams 这个类型里面的 consumer group (消费者组) 的名称还有和 Redis 集群名称存储。

RAX 叫做**基数树（前缀压缩树）**，就是有相同前缀的字符串，其前缀可以作为一个公共的父节点，什么又叫前缀树？

**Trie 树**

即**字典树**，也有的称为前缀树，是一种树形结构。广泛应用于统计和排序大量的字符串（但不仅限于字符串），所以经常被搜索引擎系统用于文本词频统计。它的优点是最大限度地减少无谓的字符串比较，查询效率比较高。

Trie 的核心思想是空间换时间，利用字符串的公共前缀来降低查询时间的开销以达到提高效率的目的。

先看一下几个场景问题：

1. 我们输入 n 个单词，每次查询一个单词，需要回答出这个单词是否在之前输入的 n 单词中出现过。

答：当然是用 map 来实现。

2. 我们输入 n 个单词，每次查询一个单词的前缀，需要回答出这个前缀是之前输入的 n 单词中多少个单词的前缀？

答：还是可以用 map 做，把输入 n 个单词中的每一个单词的前缀分别存入 map 中，然后计数，这样的话复杂度会非常的高。若有 n 个单词，平均每个单词的长度为 c，那么复杂度就会达到 `n*c`。

因此我们需要更加高效的数据结构，这时候就是 Trie 树的用武之地了。现在我们通过例子来理解什么是 Trie 树。现在我们对 cat、cash、apple、aply、ok 这几个单词建立一颗Trie 树。

![redis-trie]()

从图中可以看出：

1. 每一个节点代表一个字符
2. 有相同前缀的单词在树中就有公共的前缀节点。
3. 整棵树的根节点是空的。
4. 每个节点结束的时候用一个特殊的标记来表示，这里用 `-1` 来表示结束，从根节点到 `-1` 所经过的所有的节点对应一个英文单词。
5. 查询和插入的时间复杂度为 `O(k)`，k 为字符串长度，当然如果大量字符串没有共同前缀时还是很耗内存的。

所以，总的来说，Trie 树把很多的公共前缀独立出来共享了。这样避免了很多重复的存储。想想字典集的方式，一个个的key被单独的存储，即使他们都有公共的前缀也要单独存储。相比字典集的方式，Trie 树显然节省更多的空间。

Trie 树其实依然比较浪费空间，比如前面所说的“如果大量字符串没有共同前缀时”。比如这个字符串列表："deck", "did", "doe", "dog", "doge" , "dogs"。"deck" 这一个分支，有没有必要一直往下来拆分吗？还有 "did"，存在着一样的问题。像这样的不可分叉的单支分支，其实完全可以合并，也就是压缩。

**Radix 树：压缩后的 Trie 树**

所以 Radix 树就是压缩后的 Trie 树，因此也叫**压缩 Trie 树**。比如上面的字符串列表完全可以这样存储：

![redis-rax]()

同时在具体存储上，Radix 树的处理是以 bit（或二进制数字）来读取的。一次被对比 r 个 bit。

比如 "dog", "doge" , "dogs"，按照人类可读的形式，dog 是 dogs 和 doge 的子串。但是如果按照计算机的二进制比对：


dog: 01100100 01101111 01100111

doge: 01100100 01101111 01100111 011<font color="red">0</font>0101

dogs: 01100100 01101111 01100111 011<font color="red">1</font>0011



可以发现 dog 和 doge 是在第二十五位的时候不一样的。dogs 和 doge 是在第二十八位不一样的。也就是说，从二进制的角度还可以进一步进行压缩。把第二十八位前面相同的 `011` 进一步压缩。