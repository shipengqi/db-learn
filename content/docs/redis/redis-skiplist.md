---
title: 跳表
---
# 跳表

zset 的内部实现是一个 hash 字典加一个跳跃列表 (skiplist)。

```c
struct zset {
    dict *dict; // all values  value=>score
    zskiplist *zsl;
}
```

dict 结构存储 value 和 score 值的映射关系。

参考 [跳表](https://www.shipengqi.top/algorithm-learn/10_skip_list.html)。