---
title: quicklist
weight: 6
---

Redis 早期版本存储 list 列表数据结构使用的是压缩列表 `ziplist` 和普通的双向链表 `linkedlist`，也就是元素少时用 `ziplist`，元素多时用 `linkedlist`。

但是由于 `linkedlist` 结构，要存储 `pre` 和 `next` 指针，一个指针的大小为 8 字节，两个指针的大小为 16 字节（如果节点本身存储的 value 很小，比如 1 个字节，但是还是需要 16 个字节来存储指针，非常浪费空间）。当**链表中的节点非常多时，指针占用的空间就会非常大**，而且链表的内存是不连续的，会产生内存碎片，影响内存的利用率。

Redis 使用 `quicklist` 代替了 `ziplist` 和 `linkedlist`。

`quicklist` 是 `ziplist` 和 `linkedlist` 的混合体，它将 `linkedlist` 按段切分，每一段使用 `ziplist` 来紧凑存储，每个节点之间用双向指针串接起来，组成一个**双向链表**。

![redis-quicklist](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redis-quicklist.png)

`quicklist` 虽然还有 `pre` 和 `next` 指针，但是节点少了很多。


## 配置选项

可以设置每个 `ziplist` 的最大容量，`quicklist` 的数据压缩范围，提升数据存取效率。

```ini
list-max-ziplist-size -2  # -2 表示每个 ziplist 最多存储 8kb 大小，超过则会分裂，将数据存储在新的 ziplist 节点中
list-compress-depth 0  # 压缩深度。0 表示不压缩，1 表示头部的一个，尾部的一个，一共两个 ziplist 节点不压缩，中间的节点全部压缩。依次类推，2 表示头部的两个，尾部的两个，一共四个节点不压缩。
```

**`list-max-ziplist-size` 不建议设置太大，因为 `ziplist` 增加元素时会重新分配新的内存空间，并将之前的内容一次性拷贝到新的地址**。如果 `ziplist` 占据内存太大，重新分配内存和拷贝内存就会有很大的消耗。

所以 `ziplist` 不适合存储大型字符串，存储的元素也不宜过多。

`quicklist` 默认的压缩深度是 0。

## 增加元素

增加元素时，只需要修改其中一个 `ziplist` 节点即可。当 `ziplist` 节点的大小超过了 `list-max-ziplist-size` 时，就会分裂出新的 `quicklistNode`，将数据存储在新的 `ziplist` 节点中。

## 压缩深度

当某一个 list 列表中，存储了一部分热点数据，如果头尾节点数据是热点数据经常被访问，中间的数据访问不是那么频繁，就可以设置压缩深度，将中间的数据压缩，减少内存的占用。
