# HyperLogLog
如果你负责开发维护一个大型的网站，有一天老板找产品经理要网站每个网页每天的 UV 数据，然后让你来开发这个统计模块，如何实现？

如果统计 PV 那非常好办，给每个网页一个独立的 Redis 计数器就可以了，这个计数器的 `key` 后缀加上当天的日期。这样来一个请求，`incrby` 一次，最终就可以统计出所有的 PV 数据。

但是 UV 不一样，它要去重，同一个用户一天之内的多次访问请求只能计数一次。这就要求每一个网页请求都需要带上用户的 ID，无论是登陆用户还是未登陆用户都需要一个唯一 ID 来标识。

你也许已经想到了一个简单的方案，那就是为每一个页面一个独立的 `set` 集合来存储所有当天访问过此页面的用户 ID。当一个请求过来时，我们使用 `sadd` 将用户 ID 塞进去就可以了。
通过 `scard` 可以取出这个集合的大小，这个数字就是这个页面的 UV 数据。没错，这是一个非常简单的方案。

但是，如果你的页面访问量非常大，比如一个爆款页面几千万的 UV，你需要一个很大的 `set` 集合来统计，这就非常浪费空间。如果这样的页面很多，那所需要的存储空间是惊人的。

Redis 提供了 **`HyperLogLog` 数据结构就是用来解决这种统计问题的。`HyperLogLog` 提供不精确的去重计数方案，虽然不精确但是也不是非常不精确，
标准误差是 `0.81%`**，这样的精确度已经可以满足上面的 UV 统计需求了。

`HyperLogLog` 这个数据结构需要占据 `12k` 的存储空间，Redis 对 `HyperLogLog` 的存储进行了优化，在计数比较小时，它的存储空间采用稀疏矩阵存储，空间占用很小，
仅仅在计数慢慢变大，稀疏矩阵占用空间渐渐超过了阈值时才会一次性转变成稠密矩阵，才会占用 `12k` 的空间。

## 使用
`HyperLogLog` 提供了两个指令 **`pfadd` 和 `pfcount`，根据字面意义很好理解，一个是增加计数，一个是获取计数**。
```sh
127.0.0.1:6379> pfadd codehole user1
(integer) 1
127.0.0.1:6379> pfcount codehole
(integer) 1
127.0.0.1:6379> pfadd codehole user2
(integer) 1
127.0.0.1:6379> pfcount codehole
(integer) 2
127.0.0.1:6379> pfadd codehole user3
(integer) 1
127.0.0.1:6379> pfcount codehole
(integer) 3
127.0.0.1:6379> pfadd codehole user4
(integer) 1
127.0.0.1:6379> pfcount codehole
(integer) 4
127.0.0.1:6379> pfadd codehole user5
(integer) 1
127.0.0.1:6379> pfcount codehole
(integer) 5
127.0.0.1:6379> pfadd codehole user6
(integer) 1
127.0.0.1:6379> pfcount codehole
(integer) 6
127.0.0.1:6379> pfadd codehole user7 user8 user9 user10
(integer) 1
127.0.0.1:6379> pfcount codehole
(integer) 10
```

在数据量比较大的时候，`pfcount` 的结果就会出现误差。

## pfmerge
`pfmerge` 用于将多个 pf 计数值累加在一起形成一个新的 pf 值。

比如在网站中我们有两个内容差不多的页面，运营说需要这两个页面的数据进行合并。其中页面的 UV 访问量也需要合并，那这个时候 `pfmerge` 就可以派上用场了。
