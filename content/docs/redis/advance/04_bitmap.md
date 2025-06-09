---
title: 位图
weight: 4
---

位图不是特殊的数据结构，它的内容其实就是普通的字符串，也就是 **`byte` 数组**。

在日常开发中，可能会有一些 `bool` 型数据需要存取，如果使用普通的 `key/value` 方式存储，会浪费很多存储空间,比如签到记录，签了是 `true`，没签是 `false`，记录 365 天。如果每个用户存储 365 条记录，当用户量很庞大的时候，需要的存储空间是惊人的。

对于这种操作，Redis 提供了位操作，这样每天的签到记录只占据一个位，用 1 表示已签到，0 表示没签，那么 365 天就是 365 个 bit，46 个字节就可以完全容纳下，大大节约了存储空间。

```
11001101010010
```

当我们要统计月活的时候，因为需要去重，需要使用 `set` 来记录所有活跃用户的 id，这非常浪费内存。这时就可以考虑使用位图来标记用户的活跃状态。每个用户会都在这个位图的一个确定位置上，0 表示不活跃，1 表示活跃。然后到月底遍历一次位图就可以得到月度活跃用户数。不过这个方法也是有条件的，那就是 `userid` 是整数连续的（用户 id 作为 offset），并且活跃占比较高，否则可能得不偿失。

## 基本使用

Redis 的位数组是**自动扩展**的，如果设置了某个偏移位置超出了现有的内容范围，就会自动将位数组进行零扩充。

### setbit

设置指定偏移量上的 bit 的值：

```
setbit key offset 0|1
```

- 当 key 不存在时，**自动创建一个新的 key**。
- `offset` 参数的取值范围为大于等于 `0`，小于 `2^32`(bit 映射限制在 `512MB` 以内)。

```bash
# 日活场景，设置 11 月 11 日用户 100 的登录状态为 1
redis> setbit login_11_11 100 1
(integer) 0
redis> getbit login_11_11 100
(integer) 1
redis> getbit login_11_11 101
(integer) 0 # bit 默认被初始化为 0
redis> strlen login_11_11 # 查看长度
(integer) 13 # 13 个字节
redis> type login_11_11 # 查看类型
string # 字符串类型
redis> get login_11_11 # 查看值
"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01" # 13 个字节的字符串
```

{{< callout type="info" >}}
上面的 length 之所以是 13 个字节，是因为 `100` 表示的是偏移量为 `100` 的 bit，而一个字节有 8 个 bit，所以需要 13 个字节才能表示偏移量为 `100` 的 bit。
{{< /callout >}}

### getbit

获取指定偏移量上的 bit 的值：

```bash
redis> exists testbit2
(integer) 0
redis> getbit testbit2 100
(integer) 0
redis> setbit testbit2 100 1
(integer) 0
redis> getbit testbit2 100
(integer) 1
```

**`offset` 比字符串值的长度大，或者 key 不存在时，返回 `0`**。

### 统计

- 统计指令 `bitcount` 用来统计指定位置范围内 1 的个数。
- `bitop` 指令用来对多个 `bit` 数组进行位运算。

比如可以通过 `bitcount` 统计用户一共签到了多少天，如果指定了范围参数 `[start, end]`，就可以统计在某个时间范围内用户签到了多少天。

**`start` 和 `end` 参数是字节索引，也就是说指定的位范围必须是 8 的倍数**，而不能任意指定。

```bash
redis> setbit login_11_11 100 1 # 模拟用户 100 签到
(integer) 0
redis> setbit login_11_11 101 1 # 模拟用户 101 签到
(integer) 0
redis> setbit login_11_11 102 1 # 模拟用户 102 签到
(integer) 0
redis> setbit login_11_11 103 1 # 模拟用户 103 签到
(integer) 0
redis> bitcount login_11_11 # 统计 bit 为 1 的数量，没有指定范围，默认是整个字符串
(integer) 4
redis> strlen login_11_11 # 查看长度
(integer) 13 # 13 个字节
redis> bitcount login_11_11 0 12 # 0 是第一个字节，12 是最后一个字节，统计 0-12 字节范围内 bit 为 1 的数量
(integer) 4
```

假设要统计用户连续登录的情况，比如用户 100 连续登录了 5 天，用户 101 连续登录了 3 天。

可以将几天的数据进行**位运算**，然后再统计结果中 1 的个数。

```
login_11_07: 0 1 0 1 1 1 0 1 1 1 0 1 0 0 0 1
login_11_08: 0 0 0 1 0 1 0 1 1 1 0 1 0 0 0 1
login_11_09: 1 0 0 1 1 0 0 1 1 1 0 1 0 0 0 1
login_11_10: 1 1 0 1 0 1 0 1 1 1 0 1 0 0 0 1
login_11_11: 1 1 0 1 0 0 0 1 1 1 0 1 0 0 0 1
----------------------------------------------
             0 0 0 1 0 0 0 1 1 1 0 1 0 0 0 1 # 按位与运算，连续登录的人数
             1 1 0 1 1 1 0 1 1 1 0 1 0 0 0 1 # 按位或运算，只要有一天登录，就为 1，可以用来统计周活，月活等
```

`bitop` 示例：

```bash
redis> setbit login_11_10 100 1 # 模拟用户 100 签到
(integer) 0
redis> setbit login_11_10 101 1 # 模拟用户 101 签到
(integer) 0
redis> setbit login_11_10 102 1 # 模拟用户 102 签到
(integer) 0
redis> bitop and login_11_10-11 login_11_10 login_11_11 # 按位与运算，连续登录的人数，and 表示按位与运算，结果保存在 login_11_10-11 中
(integer) 13
redis> bitcount login_11_10-11 # 统计连续登录的人数
(integer) 3
redis> bitop or login_11_10-11-active login_11_10 login_11_11 # 按位或运算，只要有一天登录，就为 1，or 表示按位或运算，结果保存在 login_11_10-11-active 中
(integer) 13
redis> bitcount login_11_10-11-active
(integer) 4
```