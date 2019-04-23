# INFO 指令
`Info` 指令，可以清晰地知道 Redis 内部一系列运行参数。

`Info` 指令显示的信息非常繁多，分为 9 大块，每个块都有非常多的参数，这 9 个块分别是:
1. Server 服务器运行的环境参数
2. Clients 客户端相关信息
3. Memory 服务器运行内存统计数据
4. Persistence 持久化信息
5. Stats 通用统计数据
6. Replication 主从复制相关信息
7. CPU CPU 使用情况
8. Cluster 集群信息
9. KeySpace 键值对统计数量信息

`Info` 可以一次性获取所有的信息，也可以按块取信息。
```sh
# 获取所有信息
> info
# 获取内存相关信息
> info memory
# 获取复制相关信息
> info replication
```