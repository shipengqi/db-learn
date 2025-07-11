---
title: mtools 搭建集群
weight: 99
draft: true
---

[mtools](https://github.com/rueckstiess/mtools) 是一套基于 Python 实现的 MongoDB 工具集，其包括MongoDB 日志分析、报表生成及简易的数据库安装等功能。它由 MongoDB 原生的工程师单独发起并做开源维护，目前已经有大量的使用者。

mtools 所包含的一些常用组件如下：

- `mlaunch`：支持快速搭建本地测试环境，可以是单机、副本集、分片集群。
- `mlogfilter`：日志过滤组件，支持按时间检索慢查询、全表扫描操作，支持通过多个属性进行信息过滤，支持输出为 JSON 格式。
- `mplotqueries`：支持将日志分析结果转换为图表形式，依赖 tkinter（Python 图形模块）和 matplotlib 模块。
- `mlogvis`：支持将日志分析结果转换为一个独立的 HTML 页面，实现与 mplotqueries 同样的功能。

## 安装 mtools

- mtools 需要调用 MongoDB 的二进制程序来启动数据库，因此**需保证环境变量 Path 路径中包含 `{MONGODB_HOME}/bin` 这个目录**。
- **需要安装 Python 环境**，需选用 Python 3.7、3.8、3.9版本。 Centos7安装Python3.9

pip 安装：

```bash
# 安装依赖
pip3 install python-dateutil
pip3 install psutil pymongo
# 安装 mtools
pip3 install mtools
```

源码安装：

```bash
wget https://github.com/rueckstiess/mtools/archive/refs/tags/v1.6.4.tar.gz
# 解压后进入 mtools
python setup.py install
```

## 使用 mtools 创建复制集

```bash
# 准备复制集使用的工作目录
mkdir -p /data/mongo
cd /data/mongo
# 初始化 3 节点复制集
mlaunch init --replicaset --nodes 3
```

端口默认从 27017 开始，依次为 27017，27018，27019。

查看复制集状态：

```bash
mongo --port 27017
replset:PRIMARY> rs.status()
```

## 使用 mtools 创建分片集群

```bash
# 准备分片集群使用的工作目录
mkdir /data/mongo-cluster
cd /data/mongo-cluster/
# 执行 mlaunch init 初始化集群
mlaunch init --sharded 2 --replicaset --node 3 --config 3 --csrs --mongos 3 --port 27050 
```

- `--sharded 2`：启用分片集群模式，分片数为 2。
- `--replicaset --nodes 3`：采用 3 节点的复制集架构，即每个分片为一致的复制集模式。
- `--config 3 --csrs`：配置服务器采用3节点的复制集架构模式，`--csrs` 是指 Config Server as a Replica Set。
- `--mongos 3`：启动 3 个 mongos 实例进程。
- `--port 27050`：集群将以 27050 作为起始端口，集群中的各个实例基于该端口向上递增。
- `--noauth`：不启用鉴权。
- `--arbiter`：向复制集中添加一个额外的仲裁器。
- `--single`：创建单个独立节点。
- `--dir`：数据目录，默认是 `./data`。
- `--binarypath`：如果环境有二进制文件，则不用指定。


如果执行成功，那么片刻后可以看到如下输出：

```bash
# 启动分片实例
launching "mongod" on port 27053
launching "mongod" on port 27054
launching "mongod" on port 27055
launching "mongod" on port 27056
launching "mongod" on port 27057
launching "mongod" on port 27058
# 启动 config server
launching config server on port 27059
launching config server on port 27060
launching config server on port 27061
# 初始化复制集
replica set "configRepl" initialized.
replica set "shard01" initialized.
replica set "shard02" initialized.
# 启动 mongos
launching "mongos" on port 27050
launching "mongos" on port 27051
launching "mongos" on port 27052
# 执行 addShards
adding shards. can take up to 30 seconds.
```

**检查分片实例**

`mlaunch list`命令可以对当前集群的实例状态进行检查。

```bash
# 显示标签
mlaunch  list --tags 
# 显示启动命令
mlaunch  list --startup
```

**连接 mongos，查看分片实例的情况**：

```bash
mongo --port 27050
mongos> db.adminCommand({listShards:1})
```

**停止、启动**

```bash
# 停止
mlaunch stop
# 再次启动集群
mlaunch start
```