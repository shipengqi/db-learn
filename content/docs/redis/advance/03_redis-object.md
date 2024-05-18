---
title: 对象
wight: 3
---

Redis 用到的所有主要**数据结构**，比如简单动态字符串（SDS）、双端链表、字典、压缩列表、整数集合，等等。

Redis 并没有直接使用这些数据结构来实现键值对数据库，而是基于这些数据结构创建了一个**对象系统**，这个系统包含字符串对象、列表对象、哈希
对象、集合对象和有序集合对象这五种类型的对象。

## 对象的类型与编码

Redis 使用对象来表示数据库中的键和值，每次在 Redis 的数据库中新创建一个键值对时，至少会创建两个对象，一个对象用作键值对的键（键对象），
另一个对象用作键值对的值（值对象）。

比如：

```sh
redis> SET msg "hello world"
OK
```

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

### type

对象的 `type` 属性记录了对象的类型，这个属性的值可以是下面列表中的任意一个：

- `REDIS_STRING`，字符串对象
- `REDIS_LIST`，列表对象
- `REDIS_HASH`，哈希对象
- `REDIS_SET`，集合对象
- `REDIS_ZSET`，有序集合对象

对于 Redis 数据库保存的键值对来说，**键总是一个字符串对象，而值则可以是字符串对象、列表对象、哈希对象、集合对象或者有序集合对象的其中
一种**，因此：

- 当我们称呼一个数据库键为“字符串键”时， 我们指的是“这个数据库键所对应的值为字符串对象”
- 当我们称呼一个键为“列表键”时， 我们指的是“这个数据库键所对应的值为列表对象”

### 编码和底层实现

**对象的 `ptr` 指针指向对象的底层实现数据结构，而这些数据结构由对象的 `encoding` 属性决定**。

`encoding` 属性记录了对象所使用的编码，也即是说这个对象使用了什么数据结构作为对象的底层实现：

| 编码 | 底层实现 |
| --- | --- |
| `REDIS_ENCODING_INT`（`int`） | `long` 类型的整数 |
| `REDIS_ENCODING_EMBSTR`（`embstr`） | `embstr` 编码的简单动态字符串 |
| `REDIS_ENCODING_RAW`（`raw`） | 简单动态字符串 |
| `REDIS_ENCODING_HT`（`hashtable`） | 字典 |
| `REDIS_ENCODING_LINKEDLIST`（`linkedlist`） | 双端链表 |
| `REDIS_ENCODING_ZIPLIST`（`ziplist`） | 压缩列表 |
| `REDIS_ENCODING_INTSET`（`intset`） | 整数集合 |
| `REDIS_ENCODING_SKIPLIST`（`skiplist`） | 跳跃表和字典 |

每种类型的对象都至少使用了两种不同的编码:

| 类型 | 编码 | 对象 |
| --- | --- | --- |
| `REDIS_STRING` | `REDIS_ENCODING_INT` | 使用整数值实现的字符串对象。 |
| `REDIS_STRING` | `REDIS_ENCODING_EMBSTR` | 使用 `embstr` 编码的简单动态字符串实现的字符串对象。 |
| `REDIS_STRING` | `REDIS_ENCODING_RAW` | 使用简单动态字符串实现的字符串对象。 |
| `REDIS_LIST` | `REDIS_ENCODING_ZIPLIST` | 使用压缩列表实现的列表对象。 |
| `REDIS_LIST` | `REDIS_ENCODING_LINKEDLIST` | 使用双端链表实现的列表对象。 |
| `REDIS_HASH` | `REDIS_ENCODING_ZIPLIST` | 使用压缩列表实现的哈希对象。 |
| `REDIS_HASH` | `REDIS_ENCODING_HT` | 使用字典实现的哈希对象。 |
| `REDIS_SET` | `REDIS_ENCODING_INTSET` | 使用整数集合实现的集合对象。 |
| `REDIS_SET` | `REDIS_ENCODING_HT`  | 使用字典实现的集合对象。 |
| `REDIS_ZSET` | `REDIS_ENCODING_ZIPLIST` | 使用压缩列表实现的有序集合对象。 |
| `REDIS_ZSET` | `REDIS_ENCODING_SKIPLIST` | 使用跳跃表和字典实现的有序集合对象。 |

## 字典遍历

Redis 对象树的主干是一个字典，如果对象很多，这个主干字典也会很大。当我们使用 keys 命令搜寻指定模式的 key 时，它会遍历整个主干字典。

## 重复遍历

字典在扩容的时候要进行渐进式迁移，会存在新旧两个 hashtable。遍历需要对这两个 hashtable 依次进行，先遍历完旧的 hashtable，再继续
遍历新的 hashtable。如果在遍历的过程中进行了 rehashStep，将已经遍历过的旧的 hashtable 的元素迁移到了新的 hashtable 中，那么会不
会出现元素的重复？

为了解决这个问题，Redis 为字典的遍历提供了 2 种迭代器，一种是安全迭代器，另一种是不安全迭代器。

## 迭代器

```c
typedef struct dictIterator {
    dict *d; // 目标字典对象
    long index; // 当前遍历的槽位置，初始化为-1
    int table; // ht[0] or ht[1]
    int safe; // 这个属性非常关键，它表示迭代器是否安全
    dictEntry *entry; // 迭代器当前指向的对象
    dictEntry *nextEntry; // 迭代器下一个指向的对象
    long long fingerprint; // 迭代器指纹，放置迭代过程中字典被修改
} dictIterator;

// 获取非安全迭代器，只读迭代器，允许 rehashStep
dictIterator *dictGetIterator(dict *d)
{
    dictIterator *iter = zmalloc(sizeof(*iter));

    iter->d = d;
    iter->table = 0;
    iter->index = -1;
    iter->safe = 0;
    iter->entry = NULL;
    iter->nextEntry = NULL;
    return iter;
}

// 获取安全迭代器，允许触发过期处理，禁止 rehashStep
dictIterator *dictGetSafeIterator(dict *d) {
    dictIterator *i = dictGetIterator(d);

    i->safe = 1;
    return i;
}
```

- **安全的迭代器**，指的是在遍历过程中可以对字典进行查找和修改，会触发过期判断，删除内部元素。迭代过程中不会出现元素重复，为了保证不重
复，就会禁止 rehashStep。
- **不安全的迭代器**，是指遍历过程中字典是只读的，不可以修改，只能调用 dictNext 对字典进行持续遍历，不得调用任何可能触发过期判断的函数。
不过好处是不影响 rehash，代价就是遍历的元素可能会出现重复。

安全迭代器在刚开始遍历时，会给字典打上一个标记，有了这个标记，rehashStep 就不会执行，遍历时元素就不会出现重复。

```c
typedef struct dict {
    dictType *type;
    void *privdata;
    dictht ht[2];
    long rehashidx;
    // 这个就是标记，它表示当前加在字典上的安全迭代器的数量
    unsigned long iterators;
} dict;

// 如果存在安全的迭代器，就禁止rehash
static void _dictRehashStep(dict *d) {
    if (d->iterators == 0) dictRehash(d,1);
}
```

## 迭代器的选择

除了 `keys` 指令使用了安全迭代器，因为结果不允许重复。那还有其它的地方使用了安全迭代器么，什么情况下遍历适合使用非安全迭代器？

如果遍历过程中不允许出现重复，那就使用 SafeIterator：

- `bgaofrewrite` 需要遍历所有对象转换称操作指令进行持久化，绝对不允许出现重复
- `bgsave` 也需要遍历所有对象来持久化，同样不允许出现重复

如果遍历过程中需要处理元素过期，需要对字典进行修改，那也必须使用 SafeIterator，因为非安全的迭代器是只读的。

其它情况下，也就是允许遍历过程中出现个别元素重复，不需要对字典进行结构性修改的情况下一律使用非安全迭代器。
