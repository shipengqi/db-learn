---
title: 搭建分片集群
weight: 88
draft: true
---

## 环境准备

- 3 台 Linux 虚拟机，准备 MongoDB 环境，配置环境变量。
- **一定要版本一致（重点）**

![mongodb-shareds-cluster](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-shareds-cluster.png)

## 配置域名解析

在 3 台虚拟机上执行以下命令，注意替换实际 IP 地址：

```bash
echo "192.168.65.97  mongo1 mongo01.com mongo02.com" >> /etc/hosts
echo "192.168.65.190 mongo2 mongo03.com mongo04.com" >> /etc/hosts
echo "192.168.65.200 mongo3 mongo05.com mongo06.com" >> /etc/hosts 
```

## 准备分片目录

在各服务器上创建数据目录，我们使用 `/data`，请按自己需要修改为其他目录：

在 `mongo01.com`、`mongo03.com`、`mongo05.com` 上分别执行以下命令： 

```bash
mkdir -p /data/shard1/db  /data/shard1/log   /data/config/db  /data/config/log
```

在 `mongo02.com`、`mongo04.com`、`mongo06.com` 上分别执行以下命令：

```bash
mkdir -p /data/shard2/db  /data/shard2/log   /data/mongos/
```

## 创建第一个分片用的复制集

在 `mongo01.com`、`mongo03.com`、`mongo05.com` 上分别执行以下命令： 

```bash
mongod --bind_ip 0.0.0.0 --replSet shard1 --dbpath /data/shard1/db \
--logpath /data/shard1/log/mongod.log --port 27010 --fork \
--shardsvr --wiredTigerCacheSizeGB 1
```

- `--shardsvr`：声明这是集群的一个分片
- `--wiredTigerCacheSizeGB`：设置内存大小，生产环境可以不用设置，默认占一半内存。

### 初始化第一个分片复制集
 
```bash
# 进入 mongosh
mongosh mongo01.com:27010
# shard1 复制集节点初始化
rs.initiate({
    _id: "shard1",
    "members" : [
    {
        "_id": 0,
        "host" : "mongo01.com:27010"
    },
    {
        "_id": 1,
        "host" : "mongo03.com:27010"
    },
    {
        "_id": 2,
        "host" : "mongo05.com:27010"
    }
    ]
})
# 查看复制集状态
rs.status()
```

## 创建 config server 复制集

在 `mongo01.com`、`mongo03.com`、`mongo05.com` 上分别执行以下命令： 

```bash
mongod --bind_ip 0.0.0.0 --replSet config --dbpath /data/config/db \
--logpath /data/config/log/mongod.log --port 27019 --fork \
--configsvr --wiredTigerCacheSizeGB 1
```

- `--configsvr` ：声明这是集群的一个配置服务器

#### 初始化 config server 复制集

```bash
# 进入 mongosh
mongosh mongo01.com:27019
# config 复制集节点初始化
rs.initiate({
    _id: "config",
    "members" : [
    {
        "_id": 0,
        "host" : "mongo01.com:27019"
    },
    {
        "_id": 1,
        "host" : "mongo03.com:27019"
    },
    {
        "_id": 2,
        "host" : "mongo05.com:27019"
    }
    ]
})
```

## 搭建 mongos

在 `mongo01.com`、`mongo03.com`、`mongo05.com` 上分别执行以下命令： 

```bash
# 启动 mongos,指定 config 复制集
mongos --bind_ip 0.0.0.0 --logpath /data/mongos/mongos.log --port 27017 --fork \
--configdb config/mongo01.com:27019,mongo03.com:27019,mongo05.com:27019
```

### mongos 加入第 1 个分片

mongos 虽然和 config server 建立了连接，但是还没有分片集群的信息，需要手动添加。

```bash
# 连接到 mongos
mongosh mongo01.com:27017
# 添加分片
mongos>sh.addShard("shard1/mongo01.com:27010,mongo03.com:27010,mongo05.com:27010")
# 查看 mongos 状态
mongos>sh.status()
```

## 创建分片集合

```bash
# 连接到 mongos, 创建分片集合
mongosh mongo01.com:27017
mongos>sh.status()
# 为了使集合支持分片，需要先开启 database "company" 的分片功能
mongos>sh.enableSharding("company")
# 执行 shardCollection 命令，对集合 "emp" 执行分片初始化
mongos>sh.shardCollection("company.emp", {_id: 'hashed'})
mongos>sh.status()

# 插入测试数据
use company
var emps=[]
for (var i = 0; i < 10000; i++) {
    var emp = {i:i}
    emps.push(emp);
}
db.emp.insertMany(emps);
# 查询数据分布
db.emp.getShardDistribution()
```

## 创建第 2 个分片的复制集

在 `mongo02.com`、`mongo04.com`、`mongo06.com` 上分别执行以下命令：

```bash
mongod --bind_ip 0.0.0.0 --replSet shard2 --dbpath /data/shard2/db  \
--logpath /data/shard2/log/mongod.log --port 27011 --fork \
--shardsvr --wiredTigerCacheSizeGB 1
```

### 初始化第 2 个分片复制集

```bash
# 进入 mongosh
mongosh mongo06.com:27011
# shard2 复制集节点初始化
rs.initiate({
    _id: "shard2",
    "members" : [
    {
        "_id": 0,
        "host" : "mongo06.com:27011"
    },
    {
        "_id": 1,
        "host" : "mongo02.com:27011"
    },
    {
        "_id": 2,
        "host" : "mongo04.com:27011"
    }
    ]
})
# 查看复制集状态
rs.status()
```

## mongos 加入第 2 个分片

```bash
# 连接到 mongos
mongosh mongo01.com:27017
# 添加分片
mongos>sh.addShard("shard2/mongo02.com:27011,mongo04.com:27011,mongo06.com:27011")
# 查看 mongos 状态
mongos>sh.status()
# 查询数据分布
db.emp.getShardDistribution()
```