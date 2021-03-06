---
title: Redis Key 操作
---

在日常开发中，查找某个，或某些铁定前缀的 `key`，修改他们的值，删除 `key`，都是很常用的操作。 Redis 如何从海量的 `key` 中找出满足特
定前缀的`key`列表？

Redis 的 `keys` 指令。

Redis 允许的最大 `Key` 长度是 `512MB`（对 `Value` 的长度限制也是 `512MB`），**但是尽量不要使用过长的 `key`，不仅会消耗更多的
内存，还会导致查找的效率降低。`key` 也不应该过于短，开发中应该使用统一的规范来设计 `key`，可读性好，也易于维护。比
如 `user:<user id>:followers`**。

## 查找删除

### KEYS

按指定的正则匹配模式 `pattern` 查找 `key`。

```bash
KEYS pattern
```

- `KEYS *` 匹配数据库中所有的 `key`
- `KEYS h?llo` 匹配 hello、hallo、hxllo 等
- `KEYS h*llo` 匹配 hllo、heeeeello等
- `KEYS h[ae]llo` 匹配 hello、hallo，但不匹配 hillo

`KEYS` 指令非常简单，但是有两个缺点：

- 没有 `offset`、`limit` 参数，会返回所有匹配到的 `key`。
- **执行 `KEYS` 会遍历所有的 `key`，如果 Redis 存储了海量的 `key`，由于 Redis 是单线程，`KEYS` 指令就会阻塞其他指令**，直
到 `KEYS` 执行完毕。

所以在数据量很大的情况下，不建议使用 `KEYS`，会造成 Redis 服务卡顿，导致其他的指令延时甚至超时报错。
Redis 提供了 `SCAN` 指令来解决这个问题。

```bash
# 4 个测试数据
redis> MSET one 1 two 2 three 3 four 4
OK

redis> KEYS *o*
1) "four"
2) "two"
3) "one"

redis> KEYS t??
1) "two"

redis> KEYS t[w]*
1) "two"

# 匹配数据库内所有 key
redis> KEYS *
1) "four"
2) "three"
3) "two"
4) "one"
```

### EXISTS

判断 `key` 是否存在。

```bash
EXISTS key
```

存在返回 `1`，不存在返回 `0`。

```bash
redis> SET db "redis"
OK

redis> EXISTS db
(integer) 1

redis> DEL db
(integer) 1

redis> EXISTS db
(integer) 0
```

### RANDOMKEY

随机返回一个 `key`

```bash
# 设置多个 key
redis> MSET fruit "apple" drink "beer" food "cookies"
OK

redis> RANDOMKEY
"fruit"

redis> RANDOMKEY
"food"

# 返回 key 但不删除
redis> KEYS *
1) "food"
2) "drink"
3) "fruit"

# 删除当前数据库所有 key，数据库为空
redis> FLUSHDB
OK
redis> RANDOMKEY
(nil)
```

### TYPE

返回 `key` 的值的类型。`key` 不存在返回 `none`，否则返回值得类型 `string`，`list`，`set`，`zset`，`hash`。

```bash
# 字符串
redis> SET weather "sunny"
OK
redis> TYPE weather
string

# 列表
redis> LPUSH book_list "programming in scala"
(integer) 1
redis> TYPE book_list
list

# 集合
redis> SADD pat "dog"
(integer) 1
redis> TYPE pat
set
```

### SORT

返回指定 `key` 中元素,并对元素进行排序，`key` 的类型是列表、集合、有序集合。排序默认以数字作为对象，值被解释为双精度浮点数，然后进
行比较。

```bash
SORT key [BY pattern] [LIMIT offset count] [GET pattern [GET pattern ...]] [ASC | DESC] [ALPHA] [STORE destination]
```

#### 简单用法

- `SORT key`，按从大到小顺序排序
- `SORT key DESC`，按从小到大的顺序排序

```bash
# 开销金额列表
redis> LPUSH today_cost 30 1.5 10 8
(integer) 4

# 排序
redis> SORT today_cost
1) "1.5"
2) "8"
3) "10"
4) "30"

# 倒序
redis> SORT today_cost DESC
1) "30"
2) "10"
3) "8"
4) "1.5"
```

#### ALPHA 排序

`SORT` 默认以数字作为对象排序，如果需要对字符串进行排序，使用 `ALPHA` 参数：

```bash
# 网址

redis> LPUSH website "www.reddit.com"
(integer) 1

redis> LPUSH website "www.slashdot.com"
(integer) 2

redis> LPUSH website "www.infoq.com"
(integer) 3

# 默认（按数字）排序

redis> SORT website
1) "www.infoq.com"
2) "www.slashdot.com"
3) "www.reddit.com"

# 按字符排序

redis> SORT website ALPHA
1) "www.infoq.com"
2) "www.reddit.com"
3) "www.slashdot.com"
```

#### 使用 LIMIT

类似 SQL 的分页查询，两个参数：

- offset，指定偏移量
- count，指定返回数量

```bash
# 添加测试数据，列表值为 1 指 10
redis> RPUSH rank 1 3 5 7 9
(integer) 5

redis> RPUSH rank 2 4 6 8 10
(integer) 10

# 返回列表中最小的 5 个值
redis> SORT rank LIMIT 0 5
1) "1"
2) "2"
3) "3"
4) "4"
5) "5"

# 使用排序
redis> SORT rank LIMIT 0 5 DESC
1) "10"
2) "9"
3) "8"
4) "7"
5) "6"
```

#### 使用外部 key 进行排序

可以使用外部 key 的数据作为权重，代替默认的直接对比键值的方式来进行排序。

假设现在有用户数据如下：

| uid | user_name_{uid} | user_level_{uid} |
| ------ | ------ | ------ |
| 1   | admin         | 9999             |
| 2   | jack         | 10               |
| 3   | peter         | 25               |
| 4   | mary         | 70               |

以下代码将数据输入到 Redis 中：

```bash
# admin

redis 127.0.0.1:6379> LPUSH uid 1
(integer) 1

redis 127.0.0.1:6379> SET user_name_1 admin
OK

redis 127.0.0.1:6379> SET user_level_1 9999
OK

# jack

redis 127.0.0.1:6379> LPUSH uid 2
(integer) 2

redis 127.0.0.1:6379> SET user_name_2 jack
OK

redis 127.0.0.1:6379> SET user_level_2 10
OK

# peter

redis 127.0.0.1:6379> LPUSH uid 3
(integer) 3

redis 127.0.0.1:6379> SET user_name_3 peter
OK

redis 127.0.0.1:6379> SET user_level_3 25
OK

# mary

redis 127.0.0.1:6379> LPUSH uid 4
(integer) 4

redis 127.0.0.1:6379> SET user_name_4 mary
OK

redis 127.0.0.1:6379> SET user_level_4 70
OK
```

##### BY 选项

默认情况下，`SORT uid` 直接按 `uid` 中的值排序：

```bash
redis 127.0.0.1:6379> SORT uid
1) "1"      # admin
2) "2"      # jack
3) "3"      # peter
4) "4"      # mary
```

通过使用 `BY` 选项，可以让 `uid` 按其他键的元素来排序。

比如说， 以下代码让 `uid` 键按照 `user_level_{uid}` 的大小来排序：

```bash
redis 127.0.0.1:6379> SORT uid BY user_level_*
1) "2"      # jack , level = 10
2) "3"      # peter, level = 25
3) "4"      # mary, level = 70
4) "1"      # admin, level = 9999
```

`user_level_*` 是一个占位符， 它先取出 `uid` 中的值， 然后再用这个值来查找相应的键。
比如在对 `uid` 列表进行排序时， 程序就会先取出 `uid` 的值 1 、 2 、 3 、 4 ， 然后使用 `user_level_1`、`user_level_2` 、
`user_level_3` 和 `user_level_4` 的值作为排序 `uid` 的权重。

##### GET 选项

使用 `GET` 选项， 可以根据排序的结果来取出相应的键值。

比如说， 以下代码先排序 `uid`， 再取出键 `user_name_{uid}`的值：

```bash
redis 127.0.0.1:6379> SORT uid GET user_name_*
1) "admin"
2) "jack"
3) "peter"
4) "mary"
```

##### 组合使用 BY 和 GET

通过组合使用 `BY` 和 `GET`， 可以让排序结果以更直观的方式显示出来。

比如说， 以下代码先按 `user_level_{uid}` 来排序 `uid` 列表， 再取出相应的 `user_name_{uid}` 的值：

```bash
redis 127.0.0.1:6379> SORT uid BY user_level_* GET user_name_*
1) "jack"       # level = 10
2) "peter"      # level = 25
3) "mary"       # level = 70
4) "admin"      # level = 9999
```

现在的排序结果要比只使用  `SORT uid BY user_level_*` 要直观得多。

##### 获取多个外部键

可以同时使用多个 `GET` 选项， 获取多个外部键的值。

以下代码就按 `uid` 分别获取 `user_level_{uid}` 和 `user_name_{uid}`：

```bash
redis 127.0.0.1:6379> SORT uid GET user_level_* GET user_name_*
1) "9999"       # level
2) "admin"      # name
3) "10"
4) "jack"
5) "25"
6) "peter"
7) "70"
8) "mary"
```

`GET` 有一个额外的参数规则，那就是 `——` 可以用 `#` 获取被排序键的值。

以下代码就将 `uid` 的值、及其相应的 `user_level_*` 和 `user_name_*` 都返回为结果：

```bash
redis 127.0.0.1:6379> SORT uid GET # GET user_level_* GET user_name_*
1) "1"          # uid
2) "9999"       # level
3) "admin"      # name
4) "2"
5) "10"
6) "jack"
7) "3"
8) "25"
9) "peter"
10) "4"
11) "70"
12) "mary"
```

##### 获取外部键，但不进行排序

通过将一个不存在的键作为参数传给 `BY` 选项， 可以让 `SORT` 跳过排序操作， 直接返回结果：

```bash
redis 127.0.0.1:6379> SORT uid BY not-exists-key
1) "4"
2) "3"
3) "2"
4) "1"
```

这种用法在单独使用时，没什么实际用处。

不过，通过将这种用法和 `GET` 选项配合， 就可以在不排序的情况下， 获取多个外部键， 相当于执行一个整合的获取操作（类似于 `SQL` 数
据库的 `join` 关键字）。

以下代码演示了，如何在不引起排序的情况下，使用 `SORT` 、 `BY` 和 `GET` 获取多个外部键：

```bash
redis 127.0.0.1:6379> SORT uid BY not-exists-key GET # GET user_level_* GET user_name_*
1) "4"      # id
2) "70"     # level
3) "mary"   # name
4) "3"
5) "25"
6) "peter"
7) "2"
8) "10"
9) "jack"
10) "1"
11) "9999"
12) "admin"
```

##### 将哈希表作为 GET 或 BY 的参数

除了可以将字符串键之外， 哈希表也可以作为 `GET` 或 `BY` 选项的参数来使用。

比如说，对于前面给出的用户信息表：

| uid | user_name_{uid} | user_level_{uid} |
| ------ | ------ | ------ |
| 1   | admin         | 9999             |
| 2   | jack         | 10               |
| 3   | peter         | 25               |
| 4   | mary         | 70               |

我们可以不将用户的名字和级别保存在 `user_name_{uid}` 和 `user_level_{uid}` 两个字符串键中， 而是用一个带有 `name` 域和 `level` 域
的哈希表 `user_info_{uid}` 来保存用户的名字和级别信息：

```bash
redis 127.0.0.1:6379> HMSET user_info_1 name admin level 9999
OK

redis 127.0.0.1:6379> HMSET user_info_2 name jack level 10
OK

redis 127.0.0.1:6379> HMSET user_info_3 name peter level 25
OK

redis 127.0.0.1:6379> HMSET user_info_4 name mary level 70
OK
```

之后， BY 和 GET 选项都可以用 `key->field` 的格式来获取哈希表中的域的值， 其中 `key` 表示哈希表键， 而 `field` 则表示哈希表的域：

```bash
redis 127.0.0.1:6379> SORT uid BY user_info_*->level
1) "2"
2) "3"
3) "4"
4) "1"

redis 127.0.0.1:6379> SORT uid BY user_info_*->level GET user_info_*->name
1) "jack"
2) "peter"
3) "mary"
4) "admin"
```

#### 保存排序结果

我们可以把 `SORT` 命令返回的排序结果，保存到指定 `key` 上。如果 `key` 已存在，会覆盖。
没有使用 `STORE`，`SORT`命令返回列表形式的排序结果；使用 `STORE` 参数，`SORT` 命令返回排序结果的元素数量。

```bash
redis> RPUSH numbers 1 3 5 7 9
(integer) 5

redis> RPUSH numbers 2 4 6 8 10
(integer) 10

redis> LRANGE numbers 0 -1
1) "1"
2) "3"
3) "5"
4) "7"
5) "9"
6) "2"
7) "4"
8) "6"
9) "8"
10) "10"

redis> SORT numbers STORE sorted-numbers
(integer) 10

# 排序后的结果
redis> LRANGE sorted-numbers 0 -1
1) "1"
2) "2"
3) "3"
4) "4"
5) "5"
6) "6"
7) "7"
8) "8"
9) "9"
10) "10"
```

### DEL

删除一个或多个 `key`。

```bash
DEL key [key ...]
```

返回被删除的 `key` 的数量。

```bash
#  删除单个 key

redis> SET name huangz
OK

redis> DEL name
(integer) 1


# 删除一个不存在的 key

redis> EXISTS phone
(integer) 0

redis> DEL phone # 失败，没有 key 被删除
(integer) 0


# 同时删除多个 key

redis> SET name "redis"
OK

redis> SET type "key-value store"
OK

redis> SET website "redis.com"
OK

redis> DEL name type website
(integer) 3
```

## 重命名

### RENAME

将 `key` 重命名为 `newkey`。

```bash
RENAME key newkey
```

如果 `key` 和 `newkey` 相同，或 `key` 不存在，返回一个错误。当 `newkey` 已存在，覆盖 `newkey`。

```bash
# key 存在且 newkey 不存在
redis> SET message "hello world"
OK
redis> RENAME message greeting
OK

# message 不复存在
redis> EXISTS message
(integer) 0
# 已被重命名为 greeting
redis> EXISTS greeting
(integer) 1


# 当 key 不存在时，返回错误
redis> RENAME fake_key never_exists
(error) ERR no such key


# newkey 已存在时， RENAME 会覆盖旧 newkey
redis> SET pc "lenovo"
OK
redis> SET personal_computer "dell"
OK
redis> RENAME pc personal_computer
OK
redis> GET pc
(nil)
# 原来的值 dell 被覆盖了
redis:1> GET personal_computer
"lenovo"
```

### RENAMENX

与 `RENAME` 类似，不同的是 `RENAMENX` 只有在 `newkey` 不存在的时候，才会重命名。

```bash
RENAMENX key newkey
```

如果 `newkey` 已经存在返回 `0`。

```bash
# newkey 不存在时，重命名成功
redis> SET player "MPlyaer"
OK
redis> EXISTS best_player
(integer) 0
redis> RENAMENX player best_player
(integer) 1

# newkey存在时，失败
redis> SET animal "bear"
OK
redis> SET favorite_animal "butterfly"
OK
redis> RENAMENX animal favorite_animal
(integer) 0
redis> get animal
"bear"
redis> get favorite_animal
"butterfly"
```

## 序列化和反序列化

### DUMP

序列化指定的 `key` 的值，并返回被序列化的值。

```bash
redis> SET greeting "hello, dumping world!"
OK

redis> DUMP greeting
"\x00\x15hello, dumping world!\x06\x00E\xa0Z\x82\xd8r\xc1\xde"

redis> DUMP not-exists-key
(nil)
```

### RESTORE

将序列化的值反序列化，并将反序列化的值存储到指定的 `key`。

```bash
RESTORE key ttl serialized-value
```

`ttl` 表示以毫秒为单位设置 `key` 的生存时间；如果 `ttl` 值为 `0`，表示不设置生存时间。
Redis 在进行反序化前，首先会对序列化值进行 `RDB` 较验，如果版本不符或数据不完整，会拒绝反序列化并返回一个错误

```bash
redis> SET greeting "hello, dumping world!"
OK

redis> DUMP greeting
"\x00\x15hello, dumping world!\x06\x00E\xa0Z\x82\xd8r\xc1\xde"

redis> RESTORE greeting-again 0 "\x00\x15hello, dumping world!\x06\x00E\xa0Z\x82\xd8r\xc1\xde"
OK

redis> GET greeting-again
"hello, dumping world!"

# 使用错误的值进行反序列化
redis> RESTORE fake-message 0 "hello moto moto blah blah"   ;
(error) ERR DUMP payload version or checksum are wrong
```

## 生存时间

### EXPIRE

为指定的 `key` 设置生存时间。当生存时间为 `0` 时，`key` 会自动删除。
`key` 设置生存时间后，可以再次执行 `EXPIRE` 命令更新生存时间。
**注意对 `key` 的值进行修改甚至使用 `RENAME` 对 `key` 进行重命名时，都不会修改 `key` 的生存时间**

```bash
EXPIRE key seconds
```

如果 `key` 不存在或者不能设置生存时间时，返回 `0`。

```bash
redis> SET cache_page "www.google.com"
OK

# 设置过期时间为 30 秒
redis> EXPIRE cache_page 30
(integer) 1

# 查看剩余生存时间
redis> TTL cache_page
(integer) 23

# 更新过期时间
redis> EXPIRE cache_page 30000
(integer) 1

redis> TTL cache_page
(integer) 29996
```

### EXPIREAT

与 `EXPIRE` 命令类似，不同的是 `EXPIREAT` 设置的生存时间是 `UNIX` 时间戳，以秒为单位。

```bash
EXPIREAT key timestamp
```

如果 `key` 不存在返回 `0`。

```bash
redis> SET mykey "Hello"
OK
redis> EXISTS mykey
(integer) 1
redis> EXPIREAT mykey 1293840000
(integer) 1
redis> EXISTS mykey
(integer) 0
```

### PERSIST

移除 `key` 的生存时间，将 `key` 持久化(永不过期的 `key`)。

```bash
# 设置一个 key
redis> SET mykey "Hello"
OK

# 为 key 设置生存时间
redis> EXPIRE mykey 10
(integer) 1
redis> TTL mykey
(integer) 10

# 移除 key 的生存时间
redis> PERSIST mykey
(integer) 1
redis> TTL mykey
(integer) -1
```

### PERSISTAT

与 `PERSIST` 命令类似，不同的是它以毫秒为单位设置 `key` 的过期 UNIX 时间戳。

```bash
PEXPIREAT key milliseconds-timestamp
```

如果 `key` 不存在返回 `0`。

```bash
redis> SET mykey "Hello"
OK
redis> PEXPIREAT mykey 1555555555005
(integer) 1
redis> TTL mykey
(integer) 192569170
redis> PTTL mykey
(integer) 192569169649
```

### TTL

获取指定 `key` 的剩余生存时间。

```bash
TTL key
```

如果 `key` 不存在时返回 `-2`，如果 `key` 但没有生存时间时，返回 `-1`。

```bash
# 不存在的 key
redis> FLUSHDB
OK
redis> TTL key
(integer) -2

# key 存在，但没有设置剩余生存时间
redis> SET key value
OK
redis> TTL key
(integer) -1

# 有剩余生存时间的 key
redis> EXPIRE key 10086
(integer) 1
redis> TTL key
(integer) 10084
```

### PTTL

与 `TTL` 命令类似，不同的是剩余生存时间以毫秒为单位。

```bash
PTTL key
```

如果 `key` 不存在返回 `0`。

```bash
# 不存在的 key
redis> FLUSHDB
OK
redis> PTTL key
(integer) -2


# key 存在，但没有设置剩余生存时间
redis> SET key value
OK
redis> PTTL key
(integer) -1


# 有剩余生存时间的 key
redis> PEXPIRE key 10086
(integer) 1
redis> PTTL key
(integer) 6179
```

## 迁移

### MIGRATE

将指定 `key` 从当前实例迁移到到目标实例，并从当前实例删除。**原子操作**，由于 Redis 是单线程，所以该指令会造成阻塞，直到迁移完成，
失败或者超时。

```bash
MIGRATE host port key destination-db timeout [COPY] [REPLACE]
```

- `timeout`，超时时间，以毫秒为单位。 Redis 会在指定时间内完成 IO 操作，如果传送时间内发送 IO 错误或达到了超时时间，命令就会停止，并
返回一个 `IOERR` 错误。
可选参数：
- `COPY`：不移除源实例上的 `key`。
- `REPLACE`：替换目标实例上已存在的 `key`。
迁移流程：

1. 源实例执行 `DUMP` 命令进行序列化，并将序列化数据传送到目标实例。
2. 目标实例使用 `RESTORE` 命令进行反序列化，并存储数据。
3. 当前实例和目标实例一样，收到 `RESTORE` 命令返回的 `ok`，当前实例就执行 `DEL` 命令删除 `key`。

```bash
#启动实例，使用默认的 6379 端口
$ ./redis-server &
[1] 3557

#启动实例，使用 7777 端口
$ ./redis-server --port 7777 &
[2] 3560

#连接 6379 端口的实例
$ ./redis-cli

redis> flushdb
OK

redis> SET greeting "Hello from 6379 instance"
OK

redis> MIGRATE 127.0.0.1 7777 greeting 0 1000
OK

# 迁移成功后 key 会被删除
redis> EXISTS greeting
(integer) 0

#查看 7777 端口的实例
$ ./redis-cli -p 7777

redis 127.0.0.1:7777> GET greeting
"Hello from 6379 instance"
```

### MOVE

移动当前数据库中指定的 `key` 到指定数据库 `db` 中，`MOVE` 指令是在同一个实例中的迁移。

```bash
MOVE key db
```

如果源数据库中 `key` 不存在，或者目标数据库中存在相同的 `key`，`MOVE` 命令无效。

```bash
# key 存在于当前数据库

redis> SELECT 0                             # redis默认使用数据库 0，为了清晰起见，这里再显式指定一次。
OK

redis> SET song "secret base - Zone"
OK

redis> MOVE song 1                          # 将 song 移动到数据库 1
(integer) 1

redis> EXISTS song                          # song 已经被移走
(integer) 0

redis> SELECT 1                             # 使用数据库 1
OK

redis:1> EXISTS song                        # 证实 song 被移到了数据库 1 (注意命令提示符变成了"redis:1"，表明正在使用数据库 1)
(integer) 1


# 当 key 不存在的时候

redis:1> EXISTS fake_key
(integer) 0

redis:1> MOVE fake_key 0                    # 试图从数据库 1 移动一个不存在的 key 到数据库 0，失败
(integer) 0

redis:1> select 0                           # 使用数据库0
OK

redis> EXISTS fake_key                      # 证实 fake_key 不存在
(integer) 0


# 当源数据库和目标数据库有相同的 key 时

redis> SELECT 0                             # 使用数据库0
OK
redis> SET favorite_fruit "banana"
OK

redis> SELECT 1                             # 使用数据库1
OK
redis:1> SET favorite_fruit "apple"
OK

redis:1> SELECT 0                           # 使用数据库0，并试图将 favorite_fruit 移动到数据库 1
OK

redis> MOVE favorite_fruit 1                # 因为两个数据库有相同的 key，MOVE 失败
(integer) 0

redis> GET favorite_fruit                   # 数据库 0 的 favorite_fruit 没变
"banana"
redis> SELECT 1
OK
redis:1> GET favorite_fruit                 # 数据库 1 的 favorite_fruit 也是
"apple"
```

## SCAN

迭代当前数据库中的数据库键。

```bash
SCAN cursor [MATCH pattern] [COUNT count]
```

选项:

- `cursor`，整数值，游标参数。
- `MATCH`，指定正则匹配模式，对元素的模式匹配工作是在命令从数据集中取出元素之后， 向客户端返回元素之前的这段时间内进行的，
所以如果被迭代的数据集中只有少量元素和模式相匹配， 那么迭代命令或许会在多次执行中都不返回任何元素。
- `COUNT`，指定每次迭代中从数据集里返回的元素数量，默认值为 `10`。`COUNT` 只是一个 `hint`，返回的结果可多可少。

相关命令：

- `HSCAN`，迭代哈希类型中的键值对。
- `SSCAN`，迭代集合中的元素。
- `ZSCAN`，迭代有序集合中的元素（包括元素成员和元素分值）。
- `SSCAN`，`HSCAN`和`ZSCAN`与`SCAN`都返回一个包含两个元素的 `multi-bulk` 回复，第一个元素是游标，第二个元素也是一个 `multi-bulk`
回复，包含了本次被迭代的元素。
- `SSCAN`，`HSCAN` 和 `ZSCAN` 与 `SCAN` 类似，不同的是这三个命令的的第一个参数是一个数据库键。
`SCAN` 它迭代的是当前数据库中的所有数据库键，所以不需要提供数据库键。
- `SSCAN`，`HSCAN` 和 `ZSCAN` 与 `SCAN` 的返回值也不相同：
  - `SCAN` 返回的每个元素都是一个数据库键。
  - `SSCAN` 返回的每个元素都是一个集合成员。
  - `HSCAN` 返回的每个元素都是一个键值对，一个键值对由一个键和一个值组成。
  - `ZSCAN` 返回的每个元素都是一个有序集合元素，一个有序集合元素由一个成员（member）和一个分值（score）组成。

`SCAN` 是一个基于游标的迭代器：`SCAN` 每次被调用之后，都会向用户返回一个新的游标，用户在下次迭代时需要使用这个新游标作为 `SCAN` 的游
标参数，以此来延续之前的迭代过程。

**当 `SCAN` 命令的游标参数被设置为 `0` 时，服务器将开始一次新的迭代，而当服务器向用户返回值为 `0` 的游标时，表示迭代已结束**。
当一个数据集不断地变大时，想要访问这个数据集中的所有元素就需要做越来越多的工作，能否结束一个迭代取决于用户执行迭代的速度是否比数据集
增长的速度更快。

对于 `SCAN` 这类增量式迭代命令来说，因为在对键进行增量式迭代的过程中，键可能会被修改，所以增量式迭代命令只能对被返回的元素提供有限的
保证（offer limited guarantees about the returned elements）。

**注意返回的结果可能会有重复，需要客户端去重复。**

### SCAN 对比 KEYS

- 复杂度虽然也是 `O(n)`，但是它是通过游标分步进行的，不会阻塞线程;
- 提供 `limit` 参数，可以控制每次返回结果的最大条数，`limit` 只是一个 hint，返回的结果可多可少;
- 同 `keys` 一样，它也提供模式匹配功能;
- 服务器不需要为游标保存状态，游标的唯一状态就是 `scan` 返回给客户端的游标整数;
- 返回的**结果可能会有重复，需要客户端去重复**;
- 遍历的过程中如果有数据修改，改动后的数据能不能遍历到是不确定的;
- 单次返回的结果是空的并不意味着遍历结束，而要看返回的游标值是否为零;

```bash
redis 127.0.0.1:6379> scan 0
1) "17"
2)  1) "key:12"
    2) "key:8"
    3) "key:4"
    4) "key:14"
    5) "key:16"
    6) "key:17"
    7) "key:15"
    8) "key:10"
    9) "key:3"
    10) "key:7"
    11) "key:1"

redis 127.0.0.1:6379> scan 17
1) "0"
2) 1) "key:5"
   2) "key:18"
   3) "key:0"
   4) "key:2"
   5) "key:19"
   6) "key:13"
   7) "key:6"
   8) "key:9"
   9) "key:11"

# 使用 MATCH
redis 127.0.0.1:6379> sadd myset 1 2 3 foo foobar feelsgood
(integer) 6

redis 127.0.0.1:6379> sscan myset 0 match f*
1) "0"
2) 1) "foo"
   2) "feelsgood"
   3) "foobar"

# 匹配不到元素
redis 127.0.0.1:6379> scan 0 MATCH *11*
1) "288"
2) 1) "key:911"

redis 127.0.0.1:6379> scan 288 MATCH *11*
1) "224"
2) (empty list or set)

redis 127.0.0.1:6379> scan 224 MATCH *11*
1) "80"
2) (empty list or set)

redis 127.0.0.1:6379> scan 80 MATCH *11*
1) "176"
2) (empty list or set)

# cursor 值为 0 遍历结束
redis 127.0.0.1:6379> scan 176 MATCH *11* COUNT 1000
1) "0"
2)  1) "key:611"
    2) "key:711"
    3) "key:118"
    4) "key:117"
    5) "key:311"
    6) "key:112"
    7) "key:111"
    8) "key:110"
    9) "key:113"
   10) "key:211"
   11) "key:411"
   12) "key:115"
   13) "key:116"
   14) "key:114"
   15) "key:119"
   16) "key:811"
   17) "key:511"
   18) "key:11"
```

上面的示例中提供的 limit 是 1000，但是返回的结果只有 18 个。因为这个 **limit 不是限定返回结果的数量，而是限定服务器单次遍历的字典
槽位数量**(约等于)。如果将 limit 设置为 10，发现返回结果是空的，但是游标值不为零，意味着遍历还没结束。

### 字典结构

Redis 中所有的 key 都存储在一个很大的字典中，这个字典的结构是一维数组 + 二维链表结构，第一维数组的大小总是 2^n(n>=0)，扩容一次数组
大小空间加倍，也就是 n++。

scan 指令返回的游标就是**一维数组的位置索引**，将这个位置索引称为**槽** (slot)。如果不考虑字典的扩容缩容，直接按数组下标挨个遍历就
行了。**limit 参数就表示需要遍历的槽位数**，之所以返回的结果可能多可能少，是因为不是所有的槽位上都会挂接链表，有些槽位可能是空的，还
有些槽位上挂接的链表上的元素可能会有多个。每一次遍历都会将 limit 数量的槽位上挂接的所有链表元素进行模式匹配过滤后，一次性返回给客户端。

### scan 遍历顺序

scan 的遍历顺序不是从第一维数组的第 0 位一直遍历到末尾，而是采用了**高位进位加法**来遍历。之所以使用这样特殊的方式进行遍历，是考虑
到字典的扩容和缩容时避免槽位的遍历重复和遗漏。

### 大 key 扫描

如果在 Redis 实例中形成很大的对象，比如一个很大的 hash，一个很大的 zset。这样的对象对 Redis 的集群数据迁移带来了很大的问题，因为在
集群环境下，如果某个 key 太大，会数据导致迁移卡顿。另外在内存分配上，如果一个 key 太大，那么当它需要扩容时，会一次性申请更大的一
块内存，这也会导致卡顿。如果这个大 key 被删除，内存会一次性回收，卡顿现象会再一次产生。

**尽量避免大 key 的产生**。

**如果观察到 Redis 的内存大起大落，这极有可能是因为大 key 导致的**。

#### 那如何定位大 key

用 scan 指令，对于扫描出来的每一个 key，使用 type 指令获得 key 的类型，然后使用相应数据结构的 size 或者 len 方法来得到它的大小，对
于每一种类型，保留大小的前 N 名作为扫描结果展示出来。

上面这样的过程需要编写脚本，比较繁琐，不过 Redis 官方已经在 redis-cli 指令中提供了这样的扫描功能。

```sh
redis-cli -h 127.0.0.1 -p 7001 –-bigkeys
```

如果你担心这个指令会大幅抬升 Redis 的 ops 导致线上报警，还可以增加一个休眠参数。

```sh
redis-cli -h 127.0.0.1 -p 7001 –-bigkeys -i 0.1
```

上面这个指令每隔 100 条 scan 指令就会休眠 0.1s，ops 就不会剧烈抬升，但是扫描的时间会变长。
