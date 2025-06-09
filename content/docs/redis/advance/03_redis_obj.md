---
title: Redis Object
weight: 3
---

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

Redis 的字符串有两种存储方式，在长度特别短时，使用 emb 形式存储 (embeded)，当**长度超过 44 时，使用 `raw` 形式存储**。

### int

`redisObject` 中的 `ptr` 指针是真正存储数据的地方，但是对于 `int` 编码来说，一个 `int` 类型的整数最多是 64 位，也就是 8 个字节，而 `ptr` 指针所占用的存储空间也是 8 个字节。

那么能不能**把 `int` 类型的整数直接存储在 `ptr` 指针中**呢？

答案是可以的，但是为了避免内存的浪费，Redis 在存储 value 时，会判断 value 的长度，如果 value 的**长度小于 20 个字符**并且**可以转为整形**（`2^64` 能表示的最大的数字是 20 位），那么就会把 value 直接存储在 `ptr` 指针中。

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

**CPU 缓冲行 cache line 一般一次性读取 64 字节的数据**。

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

`len`，`alloc`，`flags` 分别占据 1 字节，而且由于 sds 要兼容 C 语言的函数库，所以 **`buf` 后面还要添加一个字符 `\0`**，也占据 1 字节。所以这个 sds 对象要占据 4 字节。

48 减去 4 字节，还剩下 44 字节，刚好可以存储一个 44 字节的字符串。这样就可以**一次性把 `redisObject` 和 `sds` 一起读取到 CPU cache line 中，减少了内存访问的次数，提高了读取效率**。