---
title: 字典
weight: 2
---

Redis 中除了 `hash` 结构的数据会用到字典外，整个 Redis 数据库的所有 key 和 value 也组成了一个全局字典，还有带过期时间的 key 集合也是一个字典。`zset` 集合中存储 value 和 score 值的映射关系也是通过字典实现的。

`set` 的结构底层实现也是字典，只不过所有的 value 都是 NULL，其它的特性和字典一模一样。

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

![redis-reshash]()

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
rehash 进行期间，查找某个 key 的操作，Redis 会先去 `ht[0]` 数组上进行查找操作，如果 `ht[0]` 不存在，就会去 `ht[1]` 数组上进行查找操作。如果 `ht[0]` 存在 key，就会把对应的 key 搬到 `ht[1]` 数组上。而且会把 key 所在的桶的整个链表全部迁移到 `ht[1]` 数组上。删除和更新都都依赖于查找，先必须把元素找到，才可以进行数据结构的修改操作。

Redis 还有一个循环定时器去不断的去执行 rehash 操作，即使没有命令执行，也会不断的执行 rehash 操作。
{{< /callout >}}
 
## 扩容条件

正常情况下，当 hash 表中**元素的个数等于数组的长度时**，就会开始扩容，扩容的**新数组是原数组大小的 2 倍**。不过如果 Redis 正在做 bgsave，为了减少内存页的过多分离 (Copy On Write)，Redis 尽量不去扩容，但是如果 hash 表已经非常满了，元素的个数已经达到了数组长度的 5 倍，说明 hash 表已经过于拥挤了，这个时候就会强制扩容。

## 缩容条件

当 hash 表因为元素的逐渐删除变得越来越稀疏时，Redis 会对 hash 表进行缩容来减少 hash 表的数组空间占用。缩容的条件是**元素个数低于数组长度的 10%**。缩容不会考虑 Redis 是否正在做 bgsave。

## RedisObject

![redisdb-dict](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redisdb-dict.jpg)

Redis 在存储 value 时，会把 value 包装成一个 RedisObject 数据结构，RedisObject 是 Redis 中所有 key 和 value 的基础数据结构。比如简单动态字符串（SDS）、双端链表、字典、压缩列表、整数集合，等等。

每个对象都由一个 `redisObject` 结构表示:

```c
typedef struct redisObject {
    // 类型
    unsigned type:4;
    // 编码
    unsigned encoding:4;
    // 指向底层实现数据结构的指针
    void *ptr;
    // ...
} robj;
```


```bash
redis> set str guanyu
OK
redis> set 100 1000
OK
redis> set long aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
OK
redis> type str
string
redis> type 100
string
redis> type long
string
redis> object encoding str
"embstr"
redis> object encoding 100
"int"
redis> object encoding long
"raw"
```

`type` 命令可以看到 value 的类型都是 `string`，`object encoding` 查看编码，可以看到虽然都是字符串，但是编码方式不同。字符串就只有 3 种编码方式：

- `int`
- `embstr`
- `raw`

### raw

`raw` 就是 SDS。

Redis 的字符串有两种存储方式，在长度特别短时，使用 emb 形式存储 (embeded)，当长度超过 44 时，使用 `raw` 形式存储。

### int

`redisObject` 中的 `ptr` 指针是真正存储数据的地方，但是对于 `int` 编码来说，一个 `int` 类型的整数最多是 64 位，也就是 8 个字节，而 `ptr` 指针所占用的存储空间也是 8 个字节。

那么能不能**把 `int` 类型的整数直接存储在 `ptr` 指针中**呢？

答案是可以的，但是为了避免内存的浪费，Redis 在存储 value 时，会判断 value 的长度，如果 value 的长度小于 20 个字符并且**可以转为整形**（`2^64` 能表示的最大的数字不超过 20 位），那么就会把 value 直接存储在 `ptr` 指针中。

这么做的好处有两个：

1. 节省了内存，不需要再为 value 分配内存。
2. 可以直接使用 `ptr` 指针中的值，而不需要再根据 `ptr` 指针地址去取值，省去了一次内存访问的开销。

### embstr

字符串长度小于 44 时，使用 `embstr` 形式存储。

#### 为什么是小于 44 呢？

先看 Redis 对象头结构体：

```c
struct redisObject {
    unsigned type:4; // 4 bits
    unsigned encoding:4; // 4 bits
    unsigned lru:LRU_BITS; // 24 bits
    int refcount; // 4 bytes
    void *ptr; // 8 bytes，64-bit system
} robj;
```

- 不同的对象具有不同的类型 `type`，同一个类型的 `type` 会有不同的编码形式 `encoding`，为了记录对象的 LRU 信息，使用了 24 个 bit 来记录 LRU 信息。
- 每个对象都有个引用计数，当引用计数为零时，对象就会被销毁，内存被回收。
- `ptr` 指针将指向对象内容 (body) 的具体存储位置。这样一个 **`redisObject` 需要占据 16 字节**的存储空间。

CPU 缓冲行 cache line 一般一次性读取 64 字节的数据。

Redis 读取数据时，先从 `dictEntry` 中读取到 value 的指针，拿到 `redisObject` 后，再通过 `redisObject` 中的 `ptr` 去读取 value 的数据。上面已经知道 `redisObject` 占据 16 字节，而 CPU cache line 一般一次性读取 64 字节的数据，还有 48 字节的空间没有被使用。

**那这 48 字节的空间可以用来存储什么数据**呢？

再来看 SDS 结构体：

```c
struct __attribute__ ((__packed__)) sdshdr8 {
    uint8_t len; // 1 byte
    uint8_t alloc; // 1 byte
    unsigned char flags; // 1 byte
    char buf[];
};
```

`len`，`alloc`，`flags` 分别占据 1 字节，而且由于 sds 要兼容 C 语言的函数库，所以 `buf` 后面还要添加一个字符 `\0`，也占据 1 字节。所以这个 sds 对象要占据 4 字节。

48 减去 4 字节，还剩下 44 字节，刚好可以存储一个 44 字节的字符串。这样就可以一次性把 `redisObject` 和 `sds` 一起读取到 CPU cache line 中，减少了内存访问的次数，提高了读取效率。