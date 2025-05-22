---
title: 集群
---

## CAP 原理

C - Consistent ，一致性
A - Availability ，可用性
P - Partition tolerance ，分区容忍性
分布式系统的节点往往都是分布在不同的机器上进行网络隔离开的，这意味着必然会有网络断开的风险，这个网络断开的场景的专业词汇叫着「网络分区」。

在网络分区发生时，两个分布式节点之间无法进行通信，我们对一个节点进行的修改操作将无法同步到另外一个节点，所以数据的「一致性」将无法满足，因为两个分布式节点的数据不再保持一致。除非我们牺牲「可用性」，也就是暂停分布式节点服务，在网络分区发生时，不再提供修改数据的功能，直到网络状况完全恢复正常再继续对外提供服务。

一句话概括 CAP 原理就是——网络分区发生时，一致性和可用性两难全。

### 最终一致性

Redis 保证「最终一致性」，从节点会努力追赶主节点，最终从节点的状态会和主节点的状态将保持一致。如果网络断开了，主从节点的数据将会出现大量不一致，一旦网络恢复，从节点会采用多种策略努力追赶上落后的数据，继续尽力保持和主节点一致。


## 主从同步 

Redis 同步支持主从同步和从从同步，从从同步功能是 Redis 后续版本增加的功能，为了减轻主库的同步负担。



### redis主从架构搭建

```bash
1、复制一份redis.conf文件

2、将相关配置修改为如下值：
port 6380
pidfile /var/run/redis_6380.pid  # 把pid进程号写入pidfile配置的文件
logfile "6380.log"
dir /usr/local/redis-5.0.3/data/6380  # 指定数据存放目录
# 需要注释掉bind
# bind 127.0.0.1（bind绑定的是自己机器网卡的ip，如果有多块网卡可以配多个ip，代表允许客户端通过机器的哪些网卡ip去访问，内网一般可以不配置bind，注释掉即可）

3、配置主从复制
replicaof 192.168.0.60 6379   # 从本机6379的redis实例复制数据，Redis 5.0之前使用slaveof
replica-read-only yes  # 配置从节点只读

4、启动从节点
redis-server redis.conf   # redis.conf文件务必用你复制并修改了之后的redis.conf文件

5、连接从节点
redis-cli -p 6380

6、测试在6379实例上写数据，6380实例是否能及时同步新修改数据

7、可以自己再配置一个6381的从节点
```

### 主从同步原理

1. 如果你为 master 配置了一个slave，不管这个slave是否是第一次连接上Master，它都会发送一个 PSYNC 命令给 master 请求复制数据。
2. master 收到 PSYNC 命令后，会在后台进行数据持久化通过 bgsave 生成最新的 rdb 快照文件（这里的 rdb 与开不开启 rdb 持久化没有关系），持久化期间，master会继续接收客户端的请求，它会把这些可能**修改数据集的请求缓存在内存中**。
3. 当持久化进行完毕以后，master会把这份rdb文件数据集发送给slave
4. slave会把接收到的数据进行持久化生成rdb，然后再加载到内存中。
5. 然后，master再将之前缓存在内存中的命令发送给slave。
6. 当master与slave之间的连接由于某些原因而断开时，slave能够自动重连Master，如果master收到了多个slave并发连接请求，它**只会进行一次持久化**，而不是一个连接一次，然后再把这一份持久化的数据发送
给多个并发连接的slave。

**为什么不使用 AOF 来做数据同步呢？**

因为 RDB 更快。

#### 部分复制

就是说一个 slave 之前连接了 master，已经有部分数据了，后面又和 master 断开了连接，然后又重新连接上 master，master 会把断开连接期间修改的数据发送给 slave。

master会在其内存中创建一个复制数据用的缓存队列，缓存最近一段时间的数据，master和它所有的slave都维护了复制的数据下标offset和master的进程id，因此，当网络连接断开重连后，slave会请求master继续进行未完成的复制，从所记录的数据下标开始。如果master进程id变化了，或者从节点数据下标offset太旧，已经不在master的缓存队列里了，那么将会进行一次全量数据的复制。

### 从从同步

从从同步是从 Redis 3.0 开始支持的功能，它的出现主要是为了分担主节点的同步压力，在主从同步中，从节点也可以作为其他从节点的主节点，从而形成一个树状结构。为了缓解**主从复制风暴**(多个从节点同时复制主节点导致主节点压力过大)

![redis-master-salve]()

## Sentinel 哨兵架构

Redis 主从架构虽然可以实现数据的高可用，但是当主节点挂掉后，需要手动将从节点提升为主节点，这是一个比较麻烦的过程。

为了解决这个问题，Redis 引入了 Sentinel 哨兵架构。

Sentinel 是一个分布式架构，它由多个 Sentinel 实例组成，每个 Sentinel 实例都可以监控多个主从节点，它会持续监控主从节点的健康，当主节点挂掉后，它会自动选择一个最优的从节点切换为主节点。

![redis-sentinel]()

sentinel哨兵是特殊的redis服务，不提供读写服务，主要用来监控redis实例节点。
哨兵架构下client端第一次从哨兵找出redis的主节点，后续就直接访问redis的主节点，不会每次都通过
sentinel代理访问redis的主节点，当redis的主节点发生变化，哨兵会第一时间感知到，并且将新的redis
主节点通知给client端(这里面redis的client端一般都实现了订阅功能，订阅sentinel发布的节点变动消息)

### redis哨兵架构搭建

```bash
1、复制一份sentinel.conf文件
cp sentinel.conf sentinel-26379.conf

2、将相关配置修改为如下值：
port 26379
daemonize yes
pidfile "/var/run/redis-sentinel-26379.pid"
logfile "26379.log"
dir "/usr/local/redis-5.0.3/data"
# sentinel monitor <master-redis-name> <master-redis-ip> <master-redis-port> <quorum>
# quorum是一个数字，指明当有多少个sentinel认为一个master失效时(值一般为：sentinel总数/2 + 1)，master才算真正失效
sentinel monitor mymaster 192.168.0.60 6379 2   # mymaster这个名字随便取，客户端访问时会用到

3、启动sentinel哨兵实例
src/redis-sentinel sentinel-26379.conf

4、查看sentinel的info信息
src/redis-cli -p 26379
127.0.0.1:26379>info
可以看到Sentinel的info里已经识别出了redis的主从

5、可以自己再配置两个sentinel，端口26380和26381，注意上述配置文件里的对应数字都要修改
```

sentinel集群都启动完毕后，会将哨兵集群的元数据信息写入所有sentinel的配置文件里去(追加在文件的最下面)，查看下如下配置文件 `sentinel-26379.conf`，如下所示：

```bash
sentinel known-replica mymaster 192.168.0.60 6380 #代表redis主节点的从节点信息
sentinel known-replica mymaster 192.168.0.60 6381 #代表redis主节点的从节点信息
sentinel known-sentinel mymaster 192.168.0.60 26380 52d0a5d70c1f90475b4fc03b6ce7c3c56935760f  #代表感知到的其它哨兵节点
sentinel known-sentinel mymaster 192.168.0.60 26381 e9f530d3882f8043f76ebb8e1686438ba8bd5ca6  #代表感知到的其它哨兵节点
```

当redis主节点如果挂了，哨兵集群会重新选举出新的redis主节点，同时会修改所有sentinel节点配置文件的集群元数据信息，比如6379的redis如果挂了，假设选举出的新主节点是6380，则sentinel文件里的集群元数据信息会变成如下所示：

```bash
sentinel known-replica mymaster 192.168.0.60 6379 #代表主节点的从节点信息
sentinel known-replica mymaster 192.168.0.60 6381 #代表主节点的从节点信息
sentinel known-sentinel mymaster 192.168.0.60 26380 52d0a5d70c1f90475b4fc03b6ce7c3c56935760f  #代表感知到的其它哨兵节点
sentinel known-sentinel mymaster 192.168.0.60 26381 e9f530d3882f8043f76ebb8e1686438ba8bd5ca6  #代表感知到的其它哨兵节点
```

同时还会修改sentinel文件里之前配置的mymaster对应的6379端口，改为6380

```bash
sentinel monitor mymaster 192.168.0.60 6380 2
```

当6379的redis实例再次启动时，哨兵集群根据集群元数据信息就可以将6379端口的redis节点作为从节点加入集群

### 消息丢失

Redis 主从采用异步复制，意味着当主节点挂掉时，从节点可能没有收到全部的同步消息，这部分未同步的消息就丢失了。如果主从延迟特别大，那么
丢失的数据就可能会特别多。Sentinel 无法保证消息完全不丢失，但是也尽可能保证消息少丢失。它有两个选项可以限制主从延迟过大。

```
min-slaves-to-write 1  # 表示主节点必须至少有一个从节点在进行正常复制，否则就停止对外写服务
min-slaves-max-lag 10 # 单位是秒，表示如果 10s 没有收到从节点的反馈，就意味着从节点同步异常
```
## Redis Cluster

redis3.0 以前的版本要实现集群一般是借助哨兵sentinel工具来监控master节点的状态，如果master节点异
常，则会做主从切换，将某一台slave作为master，哨兵的配置略微复杂，并且性能和高可用性等各方面表现
一般，特别是在主从切换的瞬间存在**访问瞬断**的情况，而且哨兵模式只有一个主节点对外提供服务，没法支持
很高的并发，且**单个主节点内存也不宜设置得过大，否则会导致持久化文件过大，影响数据恢复或主从同步的效率**，一般推荐小于 10G。


### Redis Cluster 架构

Redis Cluster 是 Redis 官方提供的分布式集群方案。

![redis-cluster]()
redis 集群是一个由**多个主从节点**组成的分布式服务器群，它具有复制、高可用和分片特性。Redis 集群不需要 sentinel 哨兵也能完成节点移除和故障转移的功能。需要将每个节点设置成集群模式，这种集群模式**没有中心节点**，可水平扩展，据官方文档称可以线性扩展到上万个节点(**官方推荐不超过1000个节点**)。redis集群的性能和高可用性均优于之前版本的哨兵模式，且集群配置非常简单。

### Redis 集群搭建

redis 集群需要至少三个 master 节点，这里搭建三个 master 节点，并且给每个 master 再搭建一个 slave 节点，总共6个redis节点，这里用三台机器部署6个redis实例，每台机器一主一从，搭建集群的步骤如下：

```bash
第一步：在第一台机器的/usr/local下创建文件夹redis-cluster，然后在其下面分别创建2个文件夾如下
（1）mkdir -p /usr/local/redis-cluster
（2）mkdir 8001 8004

第一步：把之前的redis.conf配置文件copy到8001下，修改如下内容：
（1）daemonize yes
（2）port 8001（分别对每个机器的端口号进行设置）
（3）pidfile /var/run/redis_8001.pid  # 把pid进程号写入pidfile配置的文件
（4）dir /usr/local/redis-cluster/8001/（指定数据文件存放位置，必须要指定不同的目录位置，不然会丢失数据）
（5）cluster-enabled yes（启动集群模式）
（6）cluster-config-file nodes-8001.conf（集群节点信息文件，这里800x最好和port对应上）
（7）cluster-node-timeout 10000
 (8)# bind 127.0.0.1（bind绑定的是自己机器网卡的ip，如果有多块网卡可以配多个ip，代表允许客户端通过机器的哪些网卡ip去访问，内网一般可以不配置bind，注释掉即可）
 (9)protected-mode  no   （关闭保护模式）
 (10)appendonly yes
如果要设置密码需要增加如下配置：
 (11)requirepass zhuge     (设置redis访问密码)
 (12)masterauth zhuge      (设置集群节点间访问密码，跟上面一致)

第三步：把修改后的配置文件，copy到8004，修改第2、3、4、6项里的端口号，可以用批量替换：
:%s/源字符串/目的字符串/g 

第四步：另外两台机器也需要做上面几步操作，第二台机器用8002和8005，第三台机器用8003和8006

第五步：分别启动6个redis实例，然后检查是否启动成功
（1）/usr/local/redis-5.0.3/src/redis-server /usr/local/redis-cluster/800*/redis.conf
（2）ps -ef | grep redis 查看是否启动成功
    
第六步：用redis-cli创建整个redis集群(redis5以前的版本集群是依靠ruby脚本redis-trib.rb实现)
# 下面命令里的1代表为每个创建的主服务器节点创建一个从服务器节点
# 执行这条命令需要确认三台机器之间的redis实例要能相互访问，可以先简单把所有机器防火墙关掉，如果不关闭防火墙则需要打开redis服务端口和集群节点gossip通信端口16379(默认是在redis端口号上加1W)
# 关闭防火墙
# systemctl stop firewalld # 临时关闭防火墙
# systemctl disable firewalld # 禁止开机启动
# 注意：下面这条创建集群的命令大家不要直接复制，里面的空格编码可能有问题导致创建集群不成功
（1）/usr/local/redis-5.0.3/src/redis-cli -a zhuge --cluster create --cluster-replicas 1 192.168.0.61:8001 192.168.0.62:8002 192.168.0.63:8003 192.168.0.61:8004 192.168.0.62:8005 192.168.0.63:8006 

# --cluster-replicas 1 表示为每个创建的主服务器节点创建一个从服务器节点，对于这里 6 个节点来说，就是 3 主 3 从

第七步：验证集群：
（1）连接任意一个客户端即可：./redis-cli -c -h -p (-a访问服务端密码，-c表示集群模式，指定ip地址和端口号）
    如：/usr/local/redis-5.0.3/src/redis-cli -a zhuge -c -h 192.168.0.61 -p 800*
（2）进行验证： cluster info（查看集群信息）、cluster nodes（查看节点列表）
（3）进行数据操作验证
（4）关闭集群则需要逐个进行关闭，使用命令：
/usr/local/redis-5.0.3/src/redis-cli -a zhuge -c -h 192.168.0.60 -p 800* shutdown
```

![cluster-slots]()

其中 `slots` 就是分配给每个节点的槽位，只有主节点才会分配槽位。

![redis-cluster-nodes]()

前面创建集群时，8001 和 8004 是在一个节点上的，8002 和 8005 是在一个节点上的，8003 和 8006 是在一个节点上的，但是上面的节点列表中，8001 是主节点，它的从节点却是 8005，8002 的从节点是 8006，8003 的从节点是 8004，这是为什么呢？

因为更加安全，避免一个节点挂了导致小的主从集群不可用。


`cluster-config-file nodes-8001.conf` 集群创建好以后，整个集群节点的信息会被保存到这个配置文件中。

**为什么要保存到这个文件中呢？**

因为如果整个集群如果关掉了，再次启动的时候是不能再使用 `--cluster create` 命令的，只需要把每个节点的 Redis 重新启动即可。Redis 启动的时候会读取这个配置文件中的节点信息，然后再重新组件集群。


### Redis 集群原理

Redis Cluster 将所有数据划分为 16384 个 slots(槽位)，每个节点负责其中一部分槽位。槽位的信息存储于每个节点中。当 Redis Cluster 的客户端来连接集群时，它也会得到一份集群的槽位配置信息并将其缓存在客户端本地。这样当客户端要查找某个 key 时，可以直接定位到目标节点。同时因为槽位的信息可能会存在客户端与服务器不一致的情况，还需要纠正机制来实现槽位信息的校验调整。

哨兵架构访问瞬断的问题在集群中也没有完全解决，但是因为集群中的数据是分散存储在多个节点上的，所以当客户端访问某个节点时，如果这个节点挂了，并不会影响其他节点的数据。只有这个小的主从集群会出现访问瞬断的情况。

#### 槽位定位算法

Cluster 默认会对 key 值使用 crc16 算法进行 hash 得到一个整数值，然后用这个整数值对 16384 进行取模来得到具体槽位。

Cluster 还允许用户强制某个 key 挂在特定槽位上，通过在 key 字符串里面嵌入 tag 标记，这就可以强制 key 所挂在的槽位等于 tag 所在的槽位。

#### 跳转重定向

当客户端向一个错误的节点发出了指令，该节点会发现指令的 key 所在的槽位并不归自己管理，这时它会向客户端发送一个特殊的跳转指令携带目标操作的节点地址，告诉客户端去连这个节点去获取数据。

```sh
GET x
-MOVED 3999 127.0.0.1:6381
```

MOVED 指令的第一个参数 3999 是 key 对应的槽位编号，后面是目标节点地址。MOVED 指令前面有一个减号，表示该指令是一个错误消息。

客户端收到 MOVED 指令后，要立即纠正本地的槽位映射表。后续所有 key 将使用新的槽位映射表。

#### Redis 集群节点间的通信机制

redis cluster 节点间采取 gossip 协议进行通信维护集群的元数据(集群节点信息，主从角色，节点数量，各节点共享的数据等)有两种方式：**集中式**和 **gossip**

- 集中式：
优点在于元数据的更新和读取，时效性非常好，一旦元数据出现变更立即就会更新到集中式的存储中，其他节点读取的时候立即就可以立即感知到；不足在于所有的元数据的更新压力全部集中在一个地方，可能导致元数据的存储压力。很多中间件都会借助 zookeeper 集中式存储元数据。

- gossip：

gossip协议包含多种消息，包括ping，pong，meet，fail等等。 
meet：某个节点发送meet给新加入的节点，让新节点加入集群中，然后新节点就会开始与其他节点进行通信；
ping：每个节点都会频繁给其他节点发送ping，其中包含自己的状态还有自己维护的集群元数据，互相通过ping交换元数据(类似自己感知到的集群节点增加和移除，hash slot信息等)； 
pong: 对ping和meet消息的返回，包含自己的状态和其他信息，也可以用于信息广播和更新； 
fail: 某个节点判断另一个节点fail之后，就发送fail给其他节点，通知其他节点，指定的节点宕机了。

gossip协议的优点在于元数据的更新比较分散，不是集中在一个地方，更新请求会陆陆续续，打到所有节点上去更新，有一定的延时，降低了压力；缺点在于元数据更新有延时可能导致集群的一些操作会有一些滞后。

gossip通信的10000端口 
每个节点都有一个专门用于节点间gossip通信的端口，就是自己提供服务的 `端口号+10000`，比如7001，那么用于节点间通信的就是17001端口。 每个节点每隔一段时间都会往另外几个节点发送ping消息，同时其他几点接收到ping消息之后返回pong消息。

这也是为什么不推荐集群的节点超过1000个的原因，因为集群内部节点的心跳通知非常频繁，这对网络带宽是一个非常大的消耗。

#### 网络抖动

真实世界的机房网络往往并不是风平浪静的，它们经常会发生各种各样的小问题。比如网络抖动就是非常常见的一种现象，突然之间部分连接变得不可访问，然后很快又恢复正常。
为解决这种问题，Redis Cluster 提供了一种选项 `cluster-node-timeout`，表示当某个节点持续 timeout 的时间失联时，才可以认定该节点出现故障，需要进行主从切换。如果没有这个选项，网络抖动会导致主从频繁切换 (数据的重新复制)。

#### Redis 集群选举原理

当slave发现自己的master变为FAIL状态时，便尝试进行Failover，以期成为新的master。由于挂掉的master可能会有多个slave，从而存在多个slave竞争成为master节点的过程， 其过程如下：
1. slave发现自己的master变为FAIL
2. 将自己记录的集群 currentEpoch 加1，并广播F `AILOVER_AUTH_REQUEST` 信息
3. 其他节点收到该信息，只有master响应，判断请求者的合法性，并发送 `FAILOVER_AUTH_ACK`，对每一个epoch只发送一次ack
4. 尝试failover的slave收集master返回的 `FAILOVER_AUTH_ACK`
5. slave收到超过半数master的ack后变成新Master(这里解释了集群为什么至少需要三个主节点，如果只有两个，当其中一个挂了，只剩一个主节点是不能选举成功的)
6. slave广播Pong消息通知其他集群节点。

为了避免多个从节点在选举获得的票数一样：

从节点并不是在主节点一进入 FAIL 状态就马上尝试发起选举，而是有一定延迟，一定的延迟确保我们等待FAIL状态在集群中传播，slave 如果立即尝试选举，其它masters或许尚未意识到FAIL状态，可能会拒绝投票
•延迟计算公式：
` DELAY = 500ms + random(0 ~ 500ms) + SLAVE_RANK * 1000ms`
•SLAVE_RANK表示此 slave 已经从 master 复制数据的总量的 rank。Rank 越小代表已复制的数据越新。这种方式下，持有最新数据的 slave 将会首先发起选举（理论上）。

### 集群脑裂数据丢失问题

脑裂数据丢失问题，网络分区导致脑裂后多个主节点对外提供写服务，一旦网络分区恢复，会将其中一个主节点变为从节点，这时会有大量数据丢失。

![redis-brain-split]()

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

当 redis.conf 的配置 `cluster-require-full-coverage` 为 `no` 时，表示当负责一个插槽的 master 节点下线且没有相应的 slave 节点进行故障恢复时，集群仍然可用，如果为 `yes` 则集群不可用。

### Redis 集群为什么至少需要三个 master 节点，并且推荐节点数为奇数？

因为新 master 的选举需要大于半数的集群 master 节点同意才能选举成功，如果只有两个 master 节点，当其中一个挂了，是达不到选举新 master 的条件的。

奇数个 master 节点可以在满足选举该条件的基础上节省一个节点，比如三个 master 节点和四个 master 节点的集群相比，如果都挂了一个 master 节点，三个 master 节点的集群只需要两个节点就可以选举，而四个 master 节点的集群需要三个节点（过半）才能选举 。所以三个 master 节点和四个 master 节点的集群都只能挂一个节点。如果都挂了两个master节点都没法选举新master节点了，所以奇数的master节点更多的是从节省机器资源角度出发说的。


### Redis 集群对批量操作命令的支持

对于 Redis 集群，批量操作命令一定要在集群上操作，因为在集群中多个 key 可能是不同的 master 节点的 slot 上，或者在同一个 master 节点的不同的 slot 上，**客户端**会直接返回错误。这是由于 Redis 要保证批量操作命令的原子性，要么全部成功，要么全部失败。在不同的 master 节点上操作，如果其中的一个 master 节点挂了，会导致有些 key 写入成功，有些 key 写入失败，这就破坏了原子性。

为了解决这个问题，则可以在 key 的前面加上 `{XX}`，这样参数数据分片 hash 计算的只会是大括号里的值，这样能确保不同的 key 能落到同一 slot 里去，示例如下：

```
mset {user1}:1:name zhuge {user1}:1:age 18
```

假设 name 和 age 计算的 hash slot 值不一样，但是这条命令在集群下执行，Redis 只会用大括号里的 `user1` 做 hash slot 计算，所以算出来的 slot 值肯定相同，最后都能落在同一 slot。