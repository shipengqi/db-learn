---
title: ziplist
weight: 5
---

Redis 为了节约内存空间使用，zset 和 hash 容器对象在元素个数较少的时候，采用压缩列表 (`ziplist`) 进行存储。压缩列表是**一块连续的内存空间，元素之间紧挨着存储，没有任何冗余空隙**。

例如：

```bash
redis> hset testhash name pooky address shanghai f1 v1 f2 v2 f3 v3
(integer) 5
redis> hgetall testhash
1) "name"
2) "pooky"
3) "address"
4) "shanghai"
5) "f1"
6) "v1"
7) "f2"
8) "v2"
9) "f3"
10) "v3"
redis> object encoding testhash
"ziplist"
```

可以看出，hash 中的元素是顺序存放的。这时 hash 底层的存储结构时 `ziplist`。

```bash
redis> hset testhash f4 aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
(integer) 1
redis> hgetall testhash
1) "name"
2) "pooky"
3) "f4"
4) "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
5) "f2"
6) "v3"
7) "f3"
8) "v3"
9) "f1"
10) "v1"
11) "address"
12) "shanghai"
redis> object encoding testhash
"hashtable"
```

可以看出，hash 中的元素时乱序的，这是因为这个时候 hash 底层的存储结构已经从 `ziplist` 变成了 `hashtable`。`f4` 这个元素的长度超过了 `hash-max-ziplist-value` 的 64 字节`。

## 数据结构

Redis 的 `ziplist` 是一个紧凑的 `byte` 数组结构，如下图，每个元素之间都是紧挨着的。

![redis-ziplist](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redis-ziplist.png)

- `zlbytes`：整个压缩列表占用字节数。
- `zltail_offset` 最后一个元素距离压缩列表起始位置的偏移量，用于快速定位到最后一个节点。
- `zllength`：元素个数。
- `entries`：元素内容列表，挨个挨个紧凑存储。
- `zlend`：标志压缩列表的结束。

{{< callout type="info" >}}
`ziplist` 为了支持双向遍历，所以才会有 `ztail_offset` 这个字段，用来快速定位到最后一个元素，然后倒着遍历。
{{< /callout >}}


`entry` 块可以容纳不同的元素类型，也会有不一样的结构：


![redis-ziplist-entry](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redis-ziplist-entry.png)

- `prerawlen`：表示前一个 entry 的字节长度，当压缩列表倒着遍历时，需要通过这个字段来快速定位到下一个元素的位置。它是一个变长的整数，当字符串长度小于 254 时，使用一个字节表示；如果达到或超出 254 那就使用 5 个字节来表示（第一个字节是 254，剩余四个字节表示字符串长度）。
- `len`：除了表示当前元素的字节长度，还有别的含义。`len` 的第一个字节分为 9 种情况：
  - `00xxxxxx`：前两个 bit 是 00，`len` 占 1 个字节。剩余的 6 个 bit 表示字符串长度，即最大的长度是 `2^6 - 1`。
  - `01xxxxxx xxxxxxxx`：前两个 bit 是 01，`len` 占 2 个字节。剩余的 14 个 bit 表示字符串长度，即最大的长度是 `2^14 - 1`。
  - `10xxxxxx xxxxxxxx xxxxxxxx xxxxxxxx xxxxxxxx`：前两个 bit 是 10，`len` 占 5 个字节。剩余的 32 个 bit 表示字符串长度，即最大的长度是 `2^32 - 1`。
  - `11000000`：表示 `int16`，`len` 占 1 个字节。后面的 `data` 占 2 个字节。
  - `11010000`：表示 `int32`，`len` 占 1 个字节。后面的 `data` 占 4 个字节。
  - `11100000`：表示 `int64`，`len` 占 1 个字节。后面的 `data` 占 8 个字节。
  - `11110000` 表示 `int24`，`len` 占 1 个字节。后面的 `data` 占 3 个字节。
  - `11111110` 表示 `int8`，`len` 占 1 个字节。后面的 `data` 占 1 个字节。
  - `1111xxxx` 表示极小整数，`xxxx` 的范围只能是 `0001~1101`, 也就是 `1~13`，因为 `0000`、`1110`、`1111`（`11111111` 表示 `ziplist` 的结束，也就是 `zlend`）都被其他情况占用了。读取到的 `value` 需要将 `xxxx` 减 `1`，也就是整数 `0~12` 就是最终的 `value`。
- `data`：元素的内容。


### 存储 hash 结构

如果 `ziplist` 存储的是 hash 结构，那么 **key 和 value 会作为两个 entry 相邻存在一起**。

![redis-ziplist-hash](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redis-ziplist-hash.png)

#### 配置选项

当数据量比较少，或者单个元素比较小的时候，Redis 会使用 `ziplist` 来存储。数据大小和元素数量的阈值可以通过以下配置项来调整：

```ini
hash-max-ziplist-entries 512  # ziplist 的元素个数超过 512，就使用 hashtable 存储
hash-max-ziplist-value 64  # 单个元素大小超过 64 字节，就使用 hashtable 存储
```

### 存储 zset 结构

`zset` 使用 `ziplist` 来存储元素时，**value 和 score 会作为两个 entry 相邻存在一起**。

![redis-ziplist-zset](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redis-ziplist-zset.png)

```bash
redis> zadd testzset 100 a 200 b 150 c
(integer) 3
redis> zrange testzset 0 -1 withscores
1) "a"
2) "100"
3) "c"
4) "150"
5) "b"
6) "200"
redis> type testzset
"zset"
redis> object encoding testzset
"ziplist"
```

#### 配置选项

```ini
zset-max-ziplist-entries 128  # zset 的元素个数超过 128，使用 skiplist 存储
zset-max-ziplist-value 64  # zset 的任意元素大小超过 64 字节，使用 skiplist 存储
```

## 增加元素

因为 `ziplist` 是紧凑存储，没有冗余空间 (对比一下 Redis 的字符串结构)。意味着插入一个新的元素就需要调用 `realloc` 扩展内存。取决于内存分配器算法和当前的 `ziplist` 内存大小，`realloc` 可能会重新分配新的内存空间，并将之前的内容一次性拷贝到新的地址，也可能在原有的地址上进行扩展，这时就不需要进行旧内容的内存拷贝。

**如果 `ziplist` 占据内存太大，重新分配内存和拷贝内存就会有很大的消耗**。所以 `ziplist` 不适合存储大型字符串，存储的元素也不宜过多。


## intset

当 set 集合容纳的**元素都是整数并且元素个数较少**时，会使用 `intset` 来存储结合元素。`intset` 是紧凑的数组结构，同时支持 16 位、32 位和 64 位整数。

**如果向 set 里存储非整数值时，那么 sadd 立即转变为 hashtable 结构**。

```bash
redis> sadd testset 1 2 3 5 10 9 4 4 4
(integer) 7
redis> smembers testset
1) "1"
2) "2"
3) "3"
4) "4"
5) "5"
6) "9"
7) "10"
redis> type testset
"set"
redis> object encoding testset
"intset"
```

可以看到上面的 `set` 中的元素是有序的，为什么不是无序的？因为 `set` 底层的存储结构是 `intset`，`intset` 是一个紧凑的数组结构。有序的数组查询的时间复杂度是 `O(logn)`，因为可以使用二分查找。

```bash
redis> sadd testset a
(integer) 1
redis> smembers testset
1) "1"
2) "4"
3) "9"
4) "3"
5) "2"
6) "10"
7) "a"
8) "5"
redis> type testset
"set"
redis> object encoding testset
"hashtable"
```

可以看到上面的 `set` 中的元素是无序的，因为 `set` 底层的存储结构已经变成了 `hashtable`。

### 配置选项

`intset` 能存储的元素个数可以通过下面的配置项来调整：

```ini
set-max-intset-entries 512  # set 的整数元素个数超过 512，使用 hashtable 存储
```

### 数据结构

```c
typedef struct intset {
    uint32_t encoding; // 编码类型，决定整数位是 16 位、32 位还是 64 位
    uint32_t length; // 元素个数
    int8_t contents[]; // 元素数组，可以是 16 位、32 位和 64 位
} intset;
```

![redis-intset](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redis-intset.png)