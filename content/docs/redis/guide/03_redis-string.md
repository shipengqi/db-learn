---
title: Redis 数据类型 String
---

String 类型是最常用，也是最简单的的一种类型，string 类型是二进制安全的。也就是说 string 可以包含任何数据。比如 `jpg图片`
或者 `序列化的对象` 。一个键**最大能存储 512MB**。Redis 所有的数据结构都是以唯一的 `key` 字符串作为名称，然后通过这个唯
一 `key` 值来获取相应的 `value` 数据。不同类型的数据结构的差异就在于 `value` 的结构不一样。

字符串结构使用非常广泛，不仅限于字符串，通常会使用 JSON 序列化成字符串，然后将序列化后的字符串塞进 Redis 来缓存。

## 键值对存取

```bash
redis> set testkey hello
OK
redis> get testkey
"hello"

//EX
redis> set testkey hello2 EX 60
OK
redis> get testkey
"hello2"
redis> TTL testkey
(integer) 55

//PX
redis> SET testkey hello3 PX 60000
OK
redis> GET testkey
"hello3"
redis> PTTL testkey
(integer) 55000

//NX
redis> SET testkey hello4 NX
OK # 键不存在，设置成功
redis> GET testkey
"hello4"
redis> SET testkey hello4 NX
(nil) # 键已经存在，设置失败
redis> GET testkey
"hello4"

//XX
redis> SET testkey hello5 XX
OK # 键已经存在，设置成功
redis> GET testkey
"hello5"
redis> SET testkey2 hello XX
(nil) # 键不存在，设置失败

//EX 和 PX 同时使用，后面的选项会覆盖前面设置的选项
redis> set testkey hello2 EX 10 PX 50000
OK
redis> TTL testkey
(integer) 45000 # PX 参数覆盖了 EX
redis> set testkey hello2 PX 50000 EX 10
OK
redis> TTL testkey
(integer) 8 # EX 参数覆盖了 PX
```

### SET

```bash
SET [key] [value] [EX seconds] [PX milliseconds] [NX|XX]
```

- EX seconds - 设置过期时间，单位为秒。
- PX millisecond - 设置过期时间，单位毫秒。
- NX - 只在 `key` 不存在时才进行设置。
- XX - 只在 `key` 存在时才进行设置。

### SETEX

设置 `key` 值并指定过期时间，单位秒。
`SETEX key second value` 等同于 `SET key value EX second`

```bash
redis> SETEX name 60 xiaoming
OK
redis> GET name
"10086"
redis> TTL name
(integer) 49
```

### PSETEX

设置 `key` 值并指定过期时间，单位毫秒。
`SET key value PX millisecond` 等同于 `PSETEX key millisecond value`

```bash
redis> PSETEX mykey 1000 "Hello"
OK
redis> PTTL mykey
(integer) 999
redis> GET mykey
"Hello"
```

### SETNX

如果 `key` 不存在，则设置其值。设置成功，返回 `1`。失败，返回 `0`。
`SET key value NX` 等同于 `SETNX key value`

### GET

获取 `key` 对应的 `value`。如果 `key` 不存在，则返回 nil；如果 `key` 不是字符串类型，则返回错误。

### GETSET

设置 `key` 的值，并返回其旧值。也就是执行了 `set` 操作和 `get` 操作。如果 `key` 不是字符串类型，则返回错误。

```bash
redis> GETSET testkey3 hello3
(nil)    # 没有旧值，返回 nil

redis> GETSET testkey3 hello4
"hello3"    # 返回旧值
```

## 批量操作键值对

同时设置或获取多个字符串，可以节省网络耗时开销。

```bash
> SET name xiaoming
OK
> SET age 18
OK
> MGET name age phone
1) "xiaoming"
2) "18"
3) (nil)
> MSET name xiaoming age 18 phone 17235617235
> MGET name age phone
1) "xiaoming"
2) "18"
3) "17235617235"
```

### MSET

**`MSET` 操作具有原子性**，所有 `key` 设置要么全成功，要么全部失败。

#### MSETNX

`MSETNX` 和 `SETNX` 类似，当 `key` 不存在时，才会设置其值。`MSETNX` 一样具有原子性。

```bash
# 对不存在的 key 进行 MSETNX
redis> MSETNX rmdbs "MySQL" nosql "MongoDB" key-value-store "redis"
(integer) 1
redis> MGET rmdbs nosql key-value-store
1) "MySQL"
2) "MongoDB"
3) "redis"

# MSET 的给定 key 当中有已存在的 key
redis> MSETNX rmdbs "Sqlite" language "python"
(integer) 0
# 因为 MSET 是原子性操作，language 没有被设置
redis> EXISTS language
(integer) 0
# rmdbs 也没有被修改
redis> GET rmdbs
"MySQL"
```

### MGET

返回一个或多个 `key` 值。

## 自增/自减

在 Redis 中，**数值也会也字符串形式存储。**
**注意，执行自增或自减时，如果 `key` 不存在，会被初始化为 `0` 再执行自增或自减操作。如果 `key` 值为非数字，那么会返回一个错误。数
字值的有效范围为 64 位(bit)有符号数字。**

```bash
redis> SET age 18
OK
redis> INCR age
(integer) 19
redis> GET age
"19"
redis> DECR age
(integer) 18
redis> INCRBY age 20
(integer) 38
```

### INCR

将 `key` 的值加 1。

### INCRBY

将 `key` 的值增加指定的值。

### INCRBYFLOAT

将 `key` 的值增加指定的浮点值。

```bash
redis> SET floatkey 9.5
OK
redis> INCRBYFLOAT floatkey 0.1
"9.6"
```

### DECR

将 `key` 的值减 `1`。如果 `key` 不存在，

```bash
redis> DECR count #count 不存在，初始化为 0，再减一
(integer) -1
```

### DECRBY

将 `key` 的值减去指定的值。

```bash
redis> SET count 100
OK
redis> DECRBY count 20
(integer) 80
```

## APPEND

向 `key` 值字符串的末尾追加指定的 `value`；如果 `key` 不存在，则执行 `SET` 操作，设置 `key` 值

```bash
redis> APPEND notexistkey hello
(integer) 5

redis> APPEND notexistkey " - redis"
(integer) 13

redis> GET myphone
"hello - redis"
```

## STRLEN

返回 key 的值的长度。如果 `key` 不存在，则返回 `0`。果 key 的值不是字符串，则返回一个错误。

```bash
redis> SET existkey "hello redis"
OK
redis> STRLEN existkey
(integer) 11

redis> STRLEN notexistkey
(integer) 0
```

## SETRANGE

使用 `value` 覆盖 `key` 以偏移量 `offset` 开始的字符串。

```bash
SETRANGE key offset value
```

如果 key 原来储存的字符串长度比偏移量小，那么原字符和偏移量之间的空白将用零字节("\x00")来填充。

```bash
# 对非空字符串进行 SETRANGE
redis> SET greeting "hello redis"
OK
redis> SETRANGE greeting 6 "Redis"
(integer) 11
redis> GET greeting
"hello Redis"

# 对空字符串/不存在的 key 进行 SETRANGE
redis> EXISTS notexistkey
(integer) 0
redis> SETRANGE notexistkey 5 "Redis"
(integer) 10
redis> GET notexistkey
"\x00\x00\x00\x00\x00Redis"
```

## GETRANGE

`GETRANGE` 类似 `javascript` 中的 `substring`。提取字符串中两个指定的索引号之间的字符。

```bash
GETRANGE key start end
```

`start` 提取字符串的起始位置，`end` 为结束位置。
如果是负数，那么该参数声明从字符串的尾部开始算起的位置。也就是说，-1 指字符串中最后一个字符，-2 指倒数第二个字符，以此类推。

```bash
redis> SET greeting "hello, redis"
OK

redis> GETRANGE greeting 0 4
"hello"

# 不支持回绕操作
redis> GETRANGE greeting -1 -5
""

redis> GETRANGE greeting -3 -1
"dis"

# 从第一个到最后一个
redis> GETRANGE greeting 0 -1
"hello, redis"

# 取值范围超过实际字符串长度范围，会被自动忽略
redis> GETRANGE greeting 0 1008611
"hello, redis"
```
