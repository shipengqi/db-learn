---
title: Redis 入门
weight: 1
---

## 基础数据类型

### 字符串 (string)

字符串 `string` 是 Redis 最简单的数据结构。Redis 所有的数据结构都是以唯一的 key 字符串作为名称，然后通过这个唯一 key 值来获取相应的 value 数据。不同类型的数据结构的差异就在于 value 的结构不一样。

`string` 类型是二进制安全的。也就是说 `string` 可以包含任何数据。比如 `jpg` 图片或者 `序列化的对象` 。一个键**最大能存储 `512MB`**。

``` bash
redis> set testkey hello
OK
redis> get testkey
"hello"
```

#### 批量操作

``` bash
redis> set name1 codehole
OK
redis> set name2 holycoder
OK
redis> mget name1 name2 name3 # 返回一个列表
1) "codehole"
2) "holycoder"
3) (nil)
redis> mset name1 boy name2 girl name3 unknown
redis> mget name1 name2 name3
1) "boy"
2) "girl"
3) "unknown"
```

#### 过期时间

可以对 key 设置过期时间，到点自动删除，这个功能常用来控制缓存的失效时间。

``` bash
redis> set name codehole
redis> get name
"codehole"
redis> expire name 5  # 5s 后过期
...  # wait for 5s
redis> get name
(nil)

redis> setex name 5 codehole  # 5s 后过期，等价于 set+expire
redis> get name
"codehole"
... # wait for 5s
redis> get name
(nil)

redis> setnx name codehole  # 如果 name 不存在就执行 set 创建
(integer) 1
redis> get name
"codehole"
redis> setnx name holycoder
(integer) 0  # 因为 name 已经存在，所以 set 创建不成功
redis> get name
"codehole"  # 没有改变
```

#### 计数

如果 value 值是一个整数，还可以对它进行自增操作。自增是有范围的，它的范围是 signed long 的最大最小值，超过了这个值，Redis 会报错。

``` bash
redis> set age 30
OK
redis> incr age
(integer) 31
redis> incrby age 5
(integer) 36
redis> incrby age -5
(integer) 31
redis> set codehole 9223372036854775807  # Long.Max
OK
redis> incr codehole
(error) ERR increment or decrement would overflow
```

### 列表（list）

Redis 的列表相当于 Java 语言里面的 `LinkedList`，注意**它是链表而不是数组**。这意味着 list 的插入和删除操作非常快，时间复杂度为 `O(1)`，但是索引定位很慢，时间复杂度为 `O(n)`，这点让人非常意外。

**当列表弹出了最后一个元素之后，该数据结构自动被删除，内存被回收**。

Redis 的列表结构常用来做**异步队列**使用。将需要延后处理的任务结构体序列化成字符串塞进 Redis 的列表，另一个线程从这个列表中轮询数据进行处理。

底层使用 `quicklist + ziplist` 存储。

#### 队列

右边进左边出（先进先出）：

``` bash
redis> rpush books python java golang
(integer) 3
redis> llen books
(integer) 3
redis> lpop books
"python"
redis> lpop books
"java"
redis> lpop books
"golang"
redis> lpop books
(nil)
```

还可以使用 `lpush` 和 `rpop` 来实现队列，效果是一样的。

#### 栈

右边进右边出（先进后出）：

``` bash
redis> rpush books python java golang
(integer) 3
redis> rpop books
"golang"
redis> rpop books
"java"
redis> rpop books
"python"
redis> rpop books
(nil)
```

#### ltrim

`lindex` 相当于 Java 链表的 `get(int index)` 方法，它需要对链表进行遍历，性能随着参数 `index` 增大而变差。

`ltrim` 和字面上的含义不太一样，叫它 `lretain` (保留) 可能更合适一些，因为 `ltrim` 跟的两个参数 `start_index` 和 `end_index` 定义了一个区间，在这个区间内的值，`ltrim` 要保留，区间之外统统砍掉。可以通过 `ltrim` 来实现一个定长的链表，这一点非常有用。

**`index` 可以为负数，`index=-1` 表示倒数第一个元素，同样 `index=-2` 表示倒数第二个元素**。

``` bash
redis> rpush books python java golang
(integer) 3
redis> lindex books 1  # O(n) 慎用
"java"
redis> lrange books 0 -1  # 获取所有元素，O(n) 慎用
1) "python"
2) "java"
3) "golang"
redis> ltrim books 1 -1 # O(n) 慎用
OK
redis> lrange books 0 -1
1) "java"
2) "golang"
redis> ltrim books 1 0 # 这其实是清空了整个列表，因为区间范围长度为负
OK
redis> llen books
(integer) 0
```

### 哈希（hash）

Redis 的字典相当于 Java 语言里面的 HashMap，它是无序字典。内部实现结构上同 Java 的 HashMap 也是一致的，同样的 `数组 + 链表` 二维结构。第一维 hash 的数组位置碰撞时，就会将碰撞的元素使用链表串接起来。

不同的是，Redis 的字典的值只能是字符串，另外它们 rehash 的方式不一样，因为 Java 的 HashMap 在字典很大时，rehash 是个耗时的操作，需要一次性全部 rehash。Redis 为了高性能，不能堵塞服务，所以采用了渐进式 rehash 策略。

**当 hash 移除了最后一个元素之后，该数据结构自动被删除，内存被回收**。

Hash 结构也可以用来存储 `JSON` 数据，不同于字符串一次性需要全部序列化整个对象，Hash 可以对 `JSON` 数据中的每个字段单独存取。而以整个字符串的形式去保存 `JSON` 数据的话就只能一次性存取，这样就会比较浪费网络流量。

底层使用 `ziplist` 或者 `hashtable` 存储。在数据量比较小，或者单个元素比较小的时候，会使用 `ziplist` 来存储。

``` bash
redis> hset books java "think in java"  # 命令行的字符串如果包含空格，要用引号括起来
(integer) 1
redis> hset books golang "concurrency in go"
(integer) 1
redis> hset books python "python cookbook"
(integer) 1
redis> hgetall books  # entries()，key 和 value 间隔出现
1) "java"
2) "think in java"
3) "golang"
4) "concurrency in go"
5) "python"
6) "python cookbook"
redis> hlen books
(integer) 3
redis> hget books java
"think in java"
redis> hset books golang "learning go programming"  # 因为是更新操作，所以返回 0
(integer) 0
redis> hget books golang
"learning go programming"
redis> hmset books java "effective java" python "learning python" golang "modern golang programming"  # 批量 set
OK
```

#### hincrby

Hash 结构中的单个子 key 也可以进行计数，它对应的指令是 hincrby，和 incr 使用基本一样。

```bash
> hincrby user-xiaoqiang age 1
(integer) 30
```

### 集合（set）

Redis 的集合相当于 Java 语言里面的 HashSet，它内部的键值对是无序的唯一的。它的内部实现相当于一个**特殊的字典，字典中所有的 value 都是一个值 NULL**。

**当集合中最后一个元素移除之后，数据结构自动删除，内存被回收**。

底层使用 `intset` 或者 `hashtable` 存储。**元素都是整数并且元素个数较小**的时候，会使用 `intset` 来存储。

``` bash
redis> sadd books python
(integer) 1
redis> sadd books python  #  重复
(integer) 0
redis> sadd books java golang
(integer) 2
redis> smembers books  # 注意顺序，和插入的并不一致，因为 set 是无序的
1) "java"
2) "python"
3) "golang"
redis> sismember books java  # 查询某个 value 是否存在，相当于 contains(o)
(integer) 1
redis> sismember books rust
(integer) 0
redis> scard books  # 获取长度相当于 count()
(integer) 3
redis> spop books  # 弹出一个
"java"
```

### 有序集合（zset）

zset 可能是 Redis 提供的最为特色的数据结构，它也是在面试中面试官最爱问的数据结构。它类似于 Java 的 SortedSet 和 HashMap 的结合体，一方面它是一个 set，保证了内部 value 的唯一性，另一方面它可以给每个 value 赋予一个 score，代表这个 value 的排序权重。它的内部实现用的是一种叫做**跳跃列表**的数据结构。

**zset 中最后一个 value 被移除后，数据结构自动删除，内存被回收**。

zset 可以用来存粉丝列表，value 值是粉丝的用户 ID，score 是关注时间。我们可以对粉丝列表按关注时间进行排序。

zset 还可以用来存储学生的成绩，value 值是学生的 ID，score 是他的考试成绩。我们可以对成绩按分数进行排序就可以得到他的名次。

底层使用 `ziplist` 或者 `skiplist + hashtable` 存储。当元素个数比较少的时候，会使用 `ziplist` 来存储。

``` bash
redis> zadd books 9.0 "think in java"
(integer) 1
redis> zadd books 8.9 "java concurrency"
(integer) 1
redis> zadd books 8.6 "java cookbook"
(integer) 1
redis> zrange books 0 -1  # 按 score 排序列出，参数区间为排名范围
1) "java cookbook"
2) "java concurrency"
3) "think in java"
redis> zrevrange books 0 -1  # 按 score 逆序列出，参数区间为排名范围
1) "think in java"
2) "java concurrency"
3) "java cookbook"
redis> zcard books  # 相当于 count()
(integer) 3
redis> zscore books "java concurrency"  # 获取指定 value 的 score
"8.9000000000000004"  # 内部 score 使用 double 类型进行存储，所以存在小数点精度问题
redis> zrank books "java concurrency"  # 排名
(integer) 1
redis> zrangebyscore books 0 8.91  # 根据分值区间遍历 zset
1) "java cookbook"
2) "java concurrency"
redis> zrangebyscore books -inf 8.91 withscores # 根据分值区间 (-∞, 8.91] 遍历 zset，同时返回分值。inf 代表 infinite，无穷大的意思。
1) "java cookbook"
2) "8.5999999999999996"
3) "java concurrency"
4) "8.9000000000000004"
redis> zrem books "java concurrency"  # 删除 value
(integer) 1
redis> zrange books 0 -1
1) "java cookbook"
2) "think in java"
```

### 容器型数据结构

`list/set/hash/zset` 这四种都属于容器型数据结构，他们有两条通用规则：

- 如果容器不存在，那就创建一个，再进行操作。比如 `RPUSH`，如果列表不存在，Redis 就会自动创建一个，然后再执行 `RPUSH`。
- 如果容器里元素没有了，那么立即删除 key，释放内存。比如 `LPOP` 操作到最后一个元素，列表 key 就会自动删除。

### 过期时间

Redis 所有的数据结构都可以设置过期时间，时间到了，Redis 会自动删除相应的对象。需要注意的是**过期是以对象为单位**，比如一个 hash 结构的过期是整个 hash 对象的过期，而不是其中的某个子 key。

还有一个需要特别注意的地方是如果一个字符串已经设置了过期时间，然后调用了 **set 方法修改了它，它的过期时间会消失**。

```bash
127.0.0.1:6379> set codehole yoyo
OK
127.0.0.1:6379> expire codehole 600
(integer) 1
127.0.0.1:6379> ttl codehole
(integer) 597
127.0.0.1:6379> set codehole yoyo
OK
127.0.0.1:6379> ttl codehole
(integer) -1
```