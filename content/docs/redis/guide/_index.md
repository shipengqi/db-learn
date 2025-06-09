---
title: 使用指南
weight: 1
---



## Redis 安装


```bash
# 安装 gcc
yum install gcc


# 下载并解压到 /usr/local
wget http://download.redis.io/releases/redis-5.0.3.tar.gz
tar -zxvf redis-5.0.3.tar.gz
cd redis-5.0.3

# 编译
make

# 运行服务器，daemonize 表示在后台运行
./src/redis-server --daemonize yes

# 或者修改配置文件后运行
# 修改配置
daemonize yes  # 后台启动
protected-mode no  # 关闭保护模式，开启的话，只有本机才可以访问 Redis
# 需要注释掉bind
# bind 127.0.0.1（bind 绑定的是自己机器网卡的 ip，如果有多块网卡可以配多个 ip，代表允许客户端通过机器的哪些网卡 ip 去访问，内网一般可以不配置 bind，注释掉即可）

# 运行服务器并指定配置文件
./src/redis-server redis.conf

# 验证启动是否成功 
ps -ef | grep redis 

# 进入 Redis 客户端 
src/redis-cli

# 退出客户端
quit

# 退出 Redis 服务： 
（1）pkill redis-server 
（2）kill 进程号                       
（3）src/redis-cli shutdown 
```

## 配置密码认证登录

Redis 默认配置是不需要密码认证，需要手动配置启用 Redis 的认证密码，增加 Redis 的安全性。

### 修改配置

打开 `redis.conf` 配置文件，找到下面的内容：

``` conf
################################## SECURITY ###################################

# Require clients to issue AUTH <PASSWORD> before processing any other
# commands.  This might be useful in environments in which you do not trust
# others with access to the host running redis-server.
#
# This should stay commented out for backward compatibility and because most
# people do not need auth (e.g. they run their own servers).
#
# Warning: since Redis is pretty fast an outside user can try up to
# 150k passwords per second against a good box. This means that you should
# use a very strong password otherwise it will be very easy to break.
#
#requirepass foobared
```

去掉注释，设置密码：

``` conf
requirepass {your password}
```

修改后重启 Redis。

### 登录验证

重启后登录时需要使用 `-a` 参数输入密码，否则登录后没有任何操作权限。如下：

``` bash
./src/redis-cli -h 127.0.0.1 -p 6379
127.0.0.1:6379> set testkey
(error) NOAUTH Authentication required.
```

使用密码认证登录：

``` bash
./redis-cli -h 127.0.0.1 -p 6379 -a myPassword
127.0.0.1:6379> set testkey hello
OK
```

或者在连接后进行验证：

``` bash
./redis-cli -h 127.0.0.1 -p 6379
127.0.0.1:6379> auth yourpassword
OK
127.0.0.1:6379> set testkey hello
OK
```

### 客户端配置密码

``` bash
127.0.0.1:6379> config set requirepass yourpassword
OK
127.0.0.1:6379> config get requirepass
1) "requirepass"
2) "yourpassword"
```

> 注意：使用客户端配置密码，重启 Redis 后仍然会使用 `redis.conf` 配置文件中的密码。

### 在集群中配置认证密码

如果 Redis 使用了集群。除了在 `master` 中配置密码外，`slave` 中也需要配置。在 `slave` 的配置文件中找到如下行，去掉注释并修改为与 `master` 相同的密码：

``` conf
# masterauth your-master-password
```