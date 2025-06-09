---
title: MongoDB
weight: 3
---

MongoDB 是一个文档数据库（以 JSON 为数据模型），C++ 编写，旨在为 WEB 应用提供可扩展的高性能数据存储解决方案。

MongoDB 是一个**介于关系数据库和非关系数据库之间的产品**，是非关系数据库当中功能最丰富，最像关系数据库的。它支持的数据结构非常松散，数据格式是 **BSON，一种类似 JSON 的二进制形式的存储格式**，简称 **Binary JSON**，和 JSON 一样**支持内嵌的文档对象和数组对象**，因此可以存储比较复杂的数据类型。

Mongo 最大的特点是它支持的查询语言非常强大，其语法有点类似于面向对象的查询语言，几乎可以实现类似关系数据库单表查询的绝大部分功能，而且还支持对数据建立索引。原则上 Oracle 和 MySQL 能做的事情，MongoDB 都能做（包括 ACID 事务）。

## MongoDB vs 关系型数据库

MongoDB 概念与关系型数据库（RDBMS）非常类似：

- **数据库**（database）：最外层的概念，可以理解为逻辑上的名称空间，一个数据库包含多个不同名称的集合。
- **集合**（collection）：相当于 SQL 中的**表**，一个集合可以存放多个不同的文档。
- **文档**（document）：一个文档相当于数据表中的一**行**，由多个不同的字段组成。
- **字段**（field）：文档中的一个属性，等同于**列**（column）。
- **索引**（index）：独立的检索式数据结构，与 SQL 概念一致。
- **`_id`**：每个文档中都拥有一个唯一的 `_id` 字段，相当于 SQL 中的**主键**（primary key）。
- **视图**（view）：可以看作一种虚拟的（非真实存在的）集合，与 SQL 中的视图类似。从 MongoDB 3.4 版本开始
提供了视图功能，其通过聚合管道技术实现。
- **聚合操作**（`$lookup`）：MongoDB 用于实现“类似”表连接（tablejoin）的聚合操作符。


MongoDB 与传统 RDBMS 仍然存在不少差异：

- **半结构化**，在**一个集合中，文档所拥有的字段并不需要是相同的**，而且也**不需要对所用的字段进行声明**。因此，MongoDB 具有很明显的半结构化特点。除了松散的表结构，文档还可以**支持多级的嵌套、数组**等灵活的数据类型，非常契合面向对象的编程模型。
- **弱关系**，MongoDB 没有外键的约束，也没有非常强大的表连接能力。类似的功能需要使用聚合管道技术来弥补。

## MongoDB 的优势

**基于灵活的 JSON 文档模型，非常适合敏捷式的快速开发**。

![mongodb-rdb]()

- 简单直观：从错综复杂的关系模型到一目了然的对象模型。
- 快速：最简单快速的开发方式。JSON 结构和对象模型接近，开发代码量低。通常后端返回给前端的数据，都是 JSON 格式，基本上可以直接使用，不用再做转换。
- 灵活：快速响应业务变化，JSON 的动态模型意味着更容易响应新的业务需求。例如，业务里需要实现一个功能，在传统的关系型数据库中，可能需要建很多张表，关联表之类的，然后再关联查询，而在 MongoDB 中，只需要新加一个字段就可以了，这个字段可以是一个内嵌文档。
- 原生的高可用：复制集提供 `99.999%` 高可用。
- 原生的高水平扩展能力：分片架构支持海量数据和无缝扩容。

{{< callout type="info" >}}
MongoDB 的集合结构可灵活的修改，可以在集合中添加一个 `version` 字段，给每个文档添加一个版本号，每当业务变更，对集合结构进行修改时，增加版本号。例如 `1.0`，`1.1`，`1.2` 等等，这样就可以实现版本控制。在后面业务可以根据版本号来进行过滤。
{{< /callout >}}

## 应用场景

- 游戏场景，使用 MongoDB 存储游戏用户信息，**用户的装备、积分等直接以内嵌文档的形式存储，方便查询、更新**；
- 物流场景，使用 MongoDB 存储订单信息，订单状态在运送过程中会不断更新，以 MongoDB **内嵌数组的形式来存储，一次查询就能将订单所有的变更读取出来**；
- 社交场景，使用 MongoDB 存储存储用户信息，以及用户发表的朋友圈信息，通过地理位置索引实现附近的人、地点等功能；
- 物联网场景，使用 MongoDB 存储所有接入的智能设备信息，以及设备汇报的日志信息，并对这些信息进行多维度的分析；
- 视频直播，使用 MongoDB 存储用户信息、礼物信息等；
- 大数据应用，使用云数据库 MongoDB 作为大数据的云存储系统，随时进行数据提取分析。

### 当前业务是否适合使用 MongoDB

**没有某个业务场景必须要使用 MongoDB 才能解决，但使用 MongoDB 通常能让你以更低的成本解决问题**。如果你不清楚当前业务是否适合使用 MongoDB,可以通过做几道选择题来辅助决策。

![mongodb-select]()

只要有一项需求满足就可以考虑使用 MongoDB，匹配越多，选择 MongoDB 越合适。基本上 MongoDB 已经可以取代 MySQL 了。

## 安装配置

环境准备：

- Linux 系统： centos7
- 安装 MongoDB 社区版

```bash
# 查看 Linux 版本
[root@hadoop01 soft]# cat /etc/redhat-release 
CentOS Linux release 7.9.2009 (Core)
```

下载地址：https://www.mongodb.com/try/download/community，其中 MongoDB Atlas 是 MongoDB 的云服务，可以提供一个 512MB 的免费数据库。

MongoDB Community Edition 是一个开源的、免费的社区版。Tools 里面是一些工具，例如 MongoDB Compass 是一个可视化工具，可以方便的查看数据库中的数据。

```bash
# 下载 MongoDB
wget https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-rhel70-6.0.5.tgz
tar -zxvf mongodb-linux-x86_64-rhel70-6.0.5.tgz
```

**启动 MongoDB Server**：

```bash
# 创建 dbpath 和 logpath
mkdir -p /mongodb/data /mongodb/log  
# 进入 mongodb 目录，启动 mongodb 服务
bin/mongod --port=27017 --dbpath=/mongodb/data --logpath=/mongodb/log/mongodb.log \
--bind_ip=0.0.0.0 --fork
```

- `--dbpath`: 指定数据文件存放目录。
- `--logpath`: 指定日志文件，注意是指定文件不是目录。
- `--logappend`: 使用追加的方式记录日志。
- `--port`: 指定端口，默认为 27017。
- `--bind_ip`: 默认只监听 `localhost` 网卡。
- `--fork`: 后台启动。
- `--auth`: 开启认证模式。


**添加环境变量**：

修改 `/etc/profile`，添加环境变量,方便执行 MongoDB 命令：

```bash
export MONGODB_HOME=/usr/local/soft/mongodb
PATH=$PATH:$MONGODB_HOME/bin
```

`source /etc/profile` 重新加载环境变量。

**使用配置文件启动服务**，编辑 `/mongodb/conf/mongo.conf` 文件，**必须是 yaml 格式**，内容如下：

```yaml
systemLog:
  destination: file
  path: /mongodb/log/mongod.log # log path
  logAppend: true
storage:
  dbPath: /mongodb/data # data directory
  engine: wiredTiger  #存储引擎
  journal:            #是否启用journal日志
    enabled: true
net:
  bindIp: 0.0.0.0
  port: 27017 # port
processManagement:
  fork: true
```

系统日志 `systemLog`：

- **logAppend**：是否启用日志追加模式。

存储 `storage`：

- **dbPath**：数据文件存放目录。
- **engine**：存储引擎，默认为 `wiredTiger`。
- **journal**：是否启用 journal 日志。journal 日志是 MongoDB 的一种日志机制，**类似于 MySQL 的 redo log**。MongoDB 在写入数据时，先写入缓冲区，默认是 100ms 刷盘一次，这就意味着在这 100ms 内，如果服务器宕机，那么数据就会丢失。如果业务对数据的可靠性要求比较高，那么最好开启 journal 日志。如果是存储日志之类的，那么可以不开启 journal 日志。

```bash
# 启动 mongodb 服务
bin/mongod -f /mongodb/conf/mongo.conf
```

**关闭 MongoDB 服务**：

```bash
mongod --port=27017 --dbpath=/mongodb/data --shutdown

# 或者
# 进入 mongosh
bin/mongosh
# 关闭服务
> use admin
# 关闭 MongoDB server 服务
admin> db.shutdownServer()
```

### mongosh 使用

mongosh 是 MongoDB 的交互式 JavaScript Shell 界面，它为系统管理员提供了强大的界面，并为开发人员提供了直接测试数据库查询和操作的方法。

下载地址：https://www.mongodb.com/try/download/shell

```bash
# centos7 安装 mongosh
wget https://downloads.mongodb.com/compass/mongodb-mongosh-1.8.0.x86_64.rpm
yum install -y mongodb-mongosh-1.8.0.x86_64.rpm

# 连接 mongodb server端
# --port:指定端口，默认为 27017
# --host:连接的主机地址，默认 127.0.0.1
mongosh --host=192.168.65.206 --port=27017 
mongosh 192.168.65.206:27017
# 指定 uri 方式连接
mongosh mongodb://192.168.65.206:27017/test
```

#### 常用命令

- `show dbs | show databases`：显示所有数据库。
- `use <dbname>`：切换到指定数据库。
- `db.dropDatabase()`：删除当前数据库。
- `show collections | show tables`：显示当前数据库中的所有集合。
- `db.<collectionName>.stat()`：显示集合详情。
- `db.<collectionName>.drop()`：删除集合。
- `show users`：显示当前数据库中的所有用户。
- `show roles`：显示当前数据库中的所有角色。
- `show profile`：显示最近发生的操作。
- `load('<filename>')`：加载并执行 JavaScript 文件。
- `exit | quit`：退出 mongosh。
- `help`：显示帮助信息。
- `db.help()`：显示当前数据库的帮助信息。
- `db.<collectionName>.help()`：显示集合的帮助信息。
- `db.version()`：显示 MongoDB 版本信息。

```bash
#查看所有库
show dbs
# 切换到指定数据库，不存在则创建
use test
# 删除当前数据库  
db.dropDatabase()

#查看集合
show collections
#创建集合
db.createCollection("emp")
#删除集合
db.emp.drop()
```

创建集合语法：

```bash
db.createCollection(name, options)
```

options 参数：

- `capped`：布尔类型，（可选）如果为 `true`，则创建固定集合。固定集合是指有着固定大小的集合，当达到最大值时，它会自动覆盖最早的文档。
- autoIndexId：指定是否自动创建 _id 字段的索引。
- size：（可选）为固定集合指定一个最大值（以字节计）。如果 `capped` 为 `true`，需要指定该字段。

### 安全认证

用用户名和密码来认证用户身份是 MongoDB 中最常用的安全认证方式。可以通过以下步骤实现：

- 创建一个管理员用户（root）并设置密码，具有所有数据库的管理权限。
- 创建一个或多个普通用户，指定相应的数据库和集合权限，并设置密码。

启用认证后，客户端连接 MongoDB 服务器时需要提供用户名和密码才能成功连接。

```bash
# 设置管理员用户名密码需要切换到 admin 库
use admin  
# 创建管理员
db.createUser({user:"fox",pwd:"fox",roles:["root"]})
# 查看当前数据库所有用户信息 
show users 
# 显示可设置权限
show roles 
# 显示所有用户
db.system.users.find() 
```

#### 常用权限

| 权限名                  | 描述                                                                         |
| ---------------------- | --------------------------------------------------------------------------- |
| read                   | 允许用户读取指定数据库                                                          |
| readWrite              | 允许用户读写指定数据库                                                          |
| dbAdmin                | 允许用户在指定数据库中执行管理函数，如索引创建、删除，查看统计或访问 `system.profile`   |
| dbOwner                | 允许用户在指定数据库中执行任意操作，增、删、改、查等                                 |
| userAdmin              | 允许用户向 `system.users` 集合写入，可以在指定数据库里创建、删除和管理用户            |
| clusterAdmin           | 只在 `admin` 数据库中可用，赋予用户所有分片和复制集相关函数的管理权限                  |
| readAnyDatabase        | 只在 `admin` 数据库中可用，赋予用户所有数据库的读权限                               |
| readWriteAnyDatabase   | 只在 `admin` 数据库中可用，赋予用户所有数据库的读写权限                              |
| userAdminAnyDatabase   | 只在 `admin` 数据库中可用，赋予用户所有数据库的 `userAdmin` 权限                    |
| dbAdminAnyDatabase     | 只在 `admin` 数据库中可用，赋予用户所有数据库的 `dbAdmin` 权限                      |
| root                   | 只在 `admin` 数据库中可用。超级账号，超级权限                                      |


修改用户操作权限：

```bash
db.grantRolesToUser( "fox" , [ 
    { role: "clusterAdmin", db: "admin" } ,
    { role: "userAdminAnyDatabase", db: "admin"},
    { role: "readWriteAnyDatabase", db: "admin"} 
])
```

删除用户：

```bash
db.dropUser("fox")

# 删除当前数据库所有用户
db.dropAllUser()
```

用户认证，返回 1 表示认证成功：

```bash
db.auth("fox","fox")    # 认证成功返回 1
```

创建应用数据库用户：

```bash
# 创建数据库
use mydb
# 创建用户
db.createUser({user:"fox",pwd:"fox",roles:["readWrite"]})    # 读写权限
```

#### MongoDB 启用鉴权

默认情况下，MongoDB 不会启用鉴权，以鉴权模式启动 MongoDB：

```bash
# 启动 mongodb 服务
bin/mongod -f /mongodb/conf/mongo.conf --auth
```

启用鉴权之后，连接 MongoDB 的相关操作都需要提供身份认证：

```bash
mongosh 192.168.65.206:27017 -u fox -p fox --authenticationDatabase=admin
```

### Docker 安装

```bash
# 拉取 mongo 镜像
docker pull mongo:6.0.5
# 运行 mongo 镜像
docker run --name mongo-server -p 29017:27017 \
-e MONGO_INITDB_ROOT_USERNAME=fox \
-e MONGO_INITDB_ROOT_PASSWORD=fox \
-d mongo:6.0.5 --wiredTigerCacheSizeGB 1

# 连接 mongo 容器
mongosh ip:29017 -u fox -p fox
```

{{< callout type="info" >}}
`--wiredTigerCacheSizeGB` 选项用于设置 WiredTiger 引擎的缓存大小。

WiredTiger 引擎使用缓存来提高读取性能。通过设置 `--wiredTigerCacheSizeGB`，可以指定 WiredTiger 引擎的缓存大小，以提高读取性能。

默认情况下，MongoDB 会将 `wiredTigerCacheSizeGB` 设置为与主机总内存成比例的值（`RAM - 1/2`，大概占机器一半的运行内存），而不考虑你可能对容器施加的内存限制。

`MONGO_INITDB_ROOT_USERNAME` 和 `MONGO_INITDB_ROOT_PASSWORD` 都存在就会启用身份认证（`mongod --auth`）
{{< /callout >}}

## 常用工具

官方 GUI COMPASS，MongoDB 图形化管理工具。下载地址：https://www.mongodb.com/zh-cn/products/compass。

MongoDB Database Tools，下载地址：https://www.mongodb.com/try/download/database-tools。

| 文件名称 | 描述 |
| --- | --- |
| mongostat | 数据库性能监控工具 |
| mongotop | 热点表监控工具 |
| mongodump | 数据库逻辑备份工具 |
| mongorestore | 数据库逻辑恢复工具 |
| mongoimport | 数据导入工具 |
| mongoexport | 数据导出工具 |
| bsondump | BSON 格式转换工具 |
| mongofiles | GridFS 文件管理工具 |