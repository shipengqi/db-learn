---
title: 核心设计与原理
weight: 2
---

## 的单线程和高性能

### Redis 是单线程吗？

Redis 的单线程主要是指 Redis 的网络 IO 和键值对读写是由一个线程来完成的，这也是 Redis 对外提供键值存储服务的主要流程。但 Redis 的其他功能，比如持久化、异步删除、集群数据同步等，其实是由额外的线程执行的。

### Redis 单线程为什么还能这么快？

因为它所有的数据都在内存中，所有的运算都是**内存级别的运算**，而且**单线程避免了多线程的切换性能损耗**问题。正因为 Redis 是单线程，所以要小心使用 Redis 指令，对于那些耗时的指令(比如 keys)，一定要谨慎使用，一不小心就可能会导致 Redis 卡顿。

### Redis 单线程如何处理那么多的并发客户端连接？

Redis 利用 epoll 来实现 IO 多路复用，将连接信息和事件放到队列中，依次放到文件事件分派器，事件分派器将事件分发给事件处理器。

```bash
# 查看 Redis 支持的最大连接数，在 redis.conf 文件中可修改，# maxclients 10000
127.0.0.1:6379> CONFIG GET maxclients
##1) "maxclients"
##2) "10000"
```

## 其他高级命令

### keys

**全量遍历键**，用来列出所有满足特定正则字符串规则的 key，当 Redis 数据量比较大时，性能比较差，要避免使用

### scan

**渐进式遍历键**，


### info

查看 redis 服务运行信息，分为 9 大块，每个块都有非常多的参数，这 9 个块分别是:
Server 服务器运行的环境参数
Clients 客户端相关信息
Memory 服务器运行内存统计数据
Persistence 持久化信息
Stats 通用统计数据
Replication 主从复制相关信息
CPU CPU 使用情况
Cluster 集群信息
KeySpace 键值对统计数量信息

```
connected_clients:2                  # 正在连接的客户端数量

instantaneous_ops_per_sec:789        # 每秒执行多少次指令

used_memory:929864                   # Redis分配的内存总量(byte)，包含redis进程内部的开销和数据占用的内存
used_memory_human:908.07K            # Redis分配的内存总量(Kb，human会展示出单位)
used_memory_rss_human:2.28M          # 向操作系统申请的内存大小(Mb)（这个值一般是大于used_memory的，因为Redis的内存分配策略会产生内存碎片）
used_memory_peak:929864              # redis的内存消耗峰值(byte)
used_memory_peak_human:908.07K       # redis的内存消耗峰值(KB)

maxmemory:0                         # 配置中设置的最大可使用内存值(byte),默认0,不限制，一般配置为机器物理内存的百分之七八十，需要留一部分给操作系统
maxmemory_human:0B                  # 配置中设置的最大可使用内存值
maxmemory_policy:noeviction         # 当达到maxmemory时的淘汰策略
```