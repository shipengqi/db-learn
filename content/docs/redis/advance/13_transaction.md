---
title: 事务
weight: 13
---

事务表示一组动作，要么全部执行，要么全部不执行。

## Redis 事务

Redis 提供了简单的事务功能，将一组需要一起执行的命令放到 `multi` 和 `exec` 两个命令之间。

`multi` 命令代表事务开始，`exec` 命令代表事务结束，如果要停止事务的执行，可以使用 `discard` 命令代替 `exec` 命令即可。它们之间的命令是原子顺序执行的，例如：

```bash
redis> multi
OK
redis> sadd u:a:follow ub
QUEUED
redis> sadd u:b:fans ua
QUEUED
```

命令返回的是 `QUEUED`，代表命令并**没有真正执行，而是暂时保存在 Redis 中的一个缓存队列**（所以 **`discard` 也只是丢弃这个缓存队列中的未执行命令**，并**不会回滚已经操作过的数据**，这一点要和关系型数据库的 rollback 操作区分开）。

如果此时另一个客户端执行 `sismember u:a:follow ub` 返回结果应该为 0，因为上面的命令还没有执行。

```bash
redis> sismember u:a:follow ub
(integer) 0
```

只有当执行 `exec` 命令后，会将缓存队列中的命令按照顺序执行，并返回执行结果。

```bash
redis> exec
1) (integer) 1
2) (integer) 1
```

另一个客户端：

```bash
redis> sismember u:a:follow ub
(integer) 1
```


如果事务中的命令出现错误,Redis 的处理机制也不尽相同。如果是 MySQL 数据库，事务中的错误会导致整个事务的执行失败，并且会回滚到事务开始之前的状态。

而 Redis 中，事务中的错误分为两种情况：

1. 命令错误

例如下面操作错将 `set` 写成了 `sett`，属于语法错误，会造成整个事务无法执行，key 和counter 的值未发生变化：

```bash
redis> set txkey hello
OK
redis> set txcount 100
OK
redis> mget txkey txcount
1) "hello"
2) "100"


redis> multi
OK
redis> set k v
QUEUED
redis> sett txkey world
(error) ERR unknown command `sett`, with args beginning with: `txkey`, `world`, 
redis> incr txcount
QUEUED
redis> exec
(error) EXECABORT Transaction discarded because of previous errors.
redis> mget txkey txcount
1) "hello"
2) "100"
```

可以看出，对于**命令错误，Redis 会将整个事务的放弃，不执行任何命令**，并且返回错误信息。

2. 运行时错误

例如用户 B 在添加粉丝列表时，误把 `sadd` 命令 (针对集合) 写成了 `zadd` 命令 (针对有序集合)，这种就是运行时命令，因为语法是正确的：

```bash
redis> multi
OK
redis> sadd u:a:follow ub
QUEUED
redis> zadd u:b:fans 1 uc
QUEUED
redis> exec
1) (integer) 1
2) (error) WRONGTYPE Operation against a key holding the wrong kind of value
redis> sismember u:c:follow ub
(integer) 1
```

`u:b:fans` 在前面已经是一个集合了，但是 `zadd` 是操作有序集合的命令，虽然命令没有错，但是运行时会出现错误。

可以看出，命令没有错，在运行时才出现的错误，Redis 会**将其他命令正常执行**，并没有全部回滚。如果碰到这种问题，需要开发人员根据具体情况进行处理。

### watch 命令

有些应用场景需要在事务之前，确保事务中的 key 没有被其他客户端修改过，才执行事务，否则不执行 (类似乐观锁)。

可以使用 `watch` 命令来实现，例如：

客户端 1：

```bash
redis> set testwatch java
OK
redis> watch testwatch
OK
redis> multi
OK
redis>
```

客户端 2：

```bash
redis> append testwatch python
(integer) 10
```

客户端 1：

```bash
redis> append testwatch jedis
QUEUED
redis> exec
(nil)
redis> get testwatch
"javapython"
```

可以看到“客户端-1”在执行 `multi` 之前执行了 `watch` 命令，“客户端-2”在“客户端-1”执行 `exec` 之前修改了 key 值，造成“客户端-1”事务没有执行 ( `exec` 结果为 `nil`，就是因为 `watch` 命令观察到 key 值被修改了，导致事务没有执行)。

{{< callout type="info" >}}
Redis 禁止在 `multi` 和 `exec` 之间执行 `watch` 指令，而必须在 `multi` 之前做好盯住关键变量，否则会出错。
{{< /callout >}}

## Pipeline 和事务的区别

1. pipeline 是客户端的行为，对于服务器来说无法区分客户端发送来的查询命令是以普通命令的形式还是以 pipeline 的形式发送到服务器的。
2. 事务则是实现在服务器端的行为，用户执行 `MULTI` 命令时，服务器会将对应这个用户的客户端对象设置为一个特殊的状态，在这个状态下后续用户执行的查询命令不会被真的执行，而是被服务器缓存起来，直到用户执行 `EXEC` 命令为止，服务器会将这个用户对应的客户端对象中缓存的命令按照提交的顺序依次执行。
3. 应用 pipeline 可以提服务器的吞吐能力，并提高 Redis 处理查询请求的能力。但是无法保证原子性。

## 优化 

可以**将事务和 pipeline 结合起来**使用，**减少事务的命令在网络上的传输时间**，将多次网络 IO 缩减为一次网络 IO。

```python
pipe = redis.pipeline(transaction=true)
pipe.multi()
pipe.incr("books")
pipe.incr("books")
values = pipe.execute()
```

## 总结

Redis 的事务过于简单，可以使用 Lua 脚本实现复杂的事务。