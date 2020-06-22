---
title: Buffer Pool
---


不管是用于存储用户数据的索引（包括聚簇索引和二级索引），还是各种系统数据，最终都是存储在磁盘上的。磁盘的读取速度是非常慢的。

所以 InnoDB 存储引擎在处理客户端的请求时，当需要**访问某个页的数据时，就会把完整的页的数据全部加载到内存中**，也就是说即使只需要访问一个页的一条记录，那也需要先把整个页的数据加载到内存中。将整个页加载到内存中后就可以进行读写访问了，在进行完读写访问之后并不着急把该页对应的内存空间释放掉，而是将其缓存起来，这样将来有请求再次访问该页面时，就可以省去磁盘 IO 的开销了。

## Buffer Pool 是什么

MySQL 服务器启动的时候就向操作系统申请了一片连续的内存，叫做 `Buffer Pool`。默认情况下 Buffer Pool 只有 `128M` 大小。服务器配置 `innodb_buffer_pool_size` 这个参数可以设置 Buffer Pool 的值：

```cnf
[server]
innodb_buffer_pool_size = 268435456
```

`268435456` 的单位是字节，也就是 `256M`。Buffer Pool 最小值为 5M(当小于该值时会自动设置成 5M)。