---
title: 字典
weight: 2
---

Redis 中除了 `hash` 结构的数据会用到字典外，整个 Redis 数据库的所有 key 和 value 也组成了一个全局字典，还有带过期时间的 key 集合也是一个字典。`zset` 集合中存储 value 和 score 值的映射关系也是通过字典实现的。

**`set` 的结构底层实现也是字典，只不过所有的 value 都是 NULL**，其它的特性和字典一模一样。

```c
struct RedisDb {
    dict* dict; // all keys  key=>value
    dict* expires; // all expired keys key=>long(timestamp)
    ...
}

struct zset {
    dict *dict; // all values  value=>score
    zskiplist *zsl;
}
```

## 字典的结构

```c
struct dict {
    ...
    dictht ht[2];
}

struct dictEntry {
    void* key;
    void* val;
    dictEntry* next; // 链接下一个 entry，用来解决 hash 冲突
}
struct dictht {
    dictEntry** table; // 二维
    long size; // 第一维数组的长度
    long used; // hash 表中的元素个数
    ...
}
```

![redis-reshash](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redis-reshash.png)

`dict` 结构内部包含两个 hashtable，通常情况下只有一个 hashtable 是有值的。但是在 `dict` 扩缩容时，需要分配新的 hashtable，然后进行**渐进式**搬迁，这时候两个 hashtable 存储的分别是旧的 hashtable 和新的 hashtable。待搬迁结束后，旧的 hashtable 被删除，新的 hashtable 取而代之。

## 渐进式 rehash

**rehash 动作并不是一次性、集中式地完成的，而是分多次、渐进式地完成的**。

原因在于，Redis 是单线程的，如果哈希表里保存的键值对数量非常庞大，一次性 rehash 庞大的计算量会导致服务器一段时间内停止服务。

渐进式 rehash 的详细步骤：

1. 为 `ht[1]` 分配空间，让字典同时持有 `ht[0]` 和 `ht[1]` 两个数组。
2. 在字典中维持一个索引计数器变量 `rehashidx` ，并将它的值设置为 0， 表示 rehash 工作正式开始。
3. 在 rehash 进行期间，每次对字典执行添加、删除、查找或者更新操作时，程序除了执行指定的操作以外，还会顺带将 `ht[0]` 在 `rehashidx` 索引上的所有键值对 rehash 到 `ht[1]` ，当 rehash 工作完成之后，程序将 `rehashidx` 属性的值增 1。
4. 随着字典操作的不断执行，最终在某个时间点上，`ht[0]` 的所有键值对都会被 rehash 至 `ht[1]`，然后将 `ht[0]` 指向新的数组，`ht[1]` 指向 `NULL`，同时程序将 `rehashidx` 属性的值设为 `-1`，表示 rehash 操作已完成。

渐进式 rehash 的过程中，字典会同时使用 `ht[0]` 和 `ht[1]` 两个数组，所以在**渐进式 rehash 进行期间，字典的删除（delete）、查找（find）、更新（update）等操作会在两个哈希表上进行**。


{{< callout type="info" >}}
rehash 进行期间，查找某个 key 的操作，Redis 会先去 `ht[0]` 数组上进行查找操作，如果 `ht[0]` 不存在，就会去 `ht[1]` 数组上进行查找操作。如果 `ht[0]` 存在 key，就会把对应的 key 搬到 `ht[1]` 数组上。而且会把 key 所在的桶的整个链表全部迁移到 `ht[1]` 数组上。删除和更新都依赖于查找，先必须把元素找到，才可以进行数据结构的修改操作。

**Redis 还有一个循环定时器去不断的去执行 rehash 操作，即使没有命令执行，也会不断的执行 rehash 操作**。

**rehash 由主线程控制，不会有并发安全的问题**。
{{< /callout >}}

<div class="img-zoom">
  <img src="https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redisdb-dict.jpg" alt="redisdb-dict">
</div>
 
## 扩容条件

正常情况下，当 hash 表中**元素的个数等于数组的长度时**，就会开始扩容，扩容的**新数组是原数组大小的 2 倍**。不过如果 Redis 正在做 bgsave，为了减少内存页的过多分离 (Copy On Write)，Redis 尽量不去扩容，但是如果 hash 表已经非常满了，元素的个数已经达到了数组长度的 5 倍，说明 hash 表已经过于拥挤了，这个时候就会强制扩容。

## 缩容条件

当 hash 表因为元素的逐渐删除变得越来越稀疏时，Redis 会对 hash 表进行缩容来减少 hash 表的数组空间占用。缩容的条件是**元素个数低于数组长度的 10%**。缩容不会考虑 Redis 是否正在做 bgsave。