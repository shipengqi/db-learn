---
title: 一些命令行技巧
---
# 一些命令行技巧
## 直接模式
一般使用 redis-cli 都会进入交互模式，然后一问一答来读写服务器，这是**交互模式**。还有一种**直接模式**，通过将命令参数直接
传递给 redis-cli 来执行指令并获取输出结果。

```sh
$ redis-cli incrby foo 5
(integer) 5

# 输出的内容较大，可以将输出重定向到外部文件
$ redis-cli info > info.txt
$ wc -l info.txt
     120 info.txt

# 如果想指向特定的服务器
# -n 2 表示使用第 2 个库，相当于 select 2
$ redis-cli -h localhost -p 6379 -n 2 ping
PONG
```

## 批量执行命令

```sh
$ cat cmds.txt
set foo1 bar1
set foo2 bar2
set foo3 bar3
......
$ cat cmds.txt | redis-cli
OK
OK
OK
...

# 或者
$ redis-cli < cmds.txt
OK
OK
OK
...
```

## set 多行字符串
如果一个字符串有多行，如何传入 set 指令？使用 `-x` 选项，该选项会使用标准输入的内容作为最后一个参数。
```sh
$ cat str.txt
Ernest Hemingway once wrote,
"The world is a fine place and worth fighting for."
I agree with the second part.
$ redis-cli -x set foo < str.txt
OK
$ redis-cli get foo
"Ernest Hemingway once wrote,\n\"The world is a fine place and worth fighting for.\"\nI agree with the second part.\n"
```

## 重复执行指令
redis-cli 还支持重复执行指令多次，每条指令执行之间设置一个间隔时间，如此便可以观察某条指令的输出内容随时间变化。
```sh
// 间隔 1s，执行 5 次，观察 qps 的变化
$ redis-cli -r 5 -i 1 info | grep ops
instantaneous_ops_per_sec:43469
instantaneous_ops_per_sec:47460
instantaneous_ops_per_sec:47699
instantaneous_ops_per_sec:46434
instantaneous_ops_per_sec:47216
```

如果将次数设置为 `-1` 那就是重复无数次永远执行下去。如果不提供 `-i` 参数，那就没有间隔，连续重复执行。

在交互模式下也可以重复执行指令，形式上比较怪异，在指令前面增加次数
```sh
127.0.0.1:6379> 5 ping
PONG
PONG
PONG
PONG
PONG
```


## 监控服务器状态
可以使用 `--stat` 参数来实时监控服务器的状态，间隔 1s 实时输出一次。
```sh
$ redis-cli --stat
------- data ------ --------------------- load -------------------- - child -
keys       mem      clients blocked requests            connections
2          6.66M    100     0       11591628 (+0)       335
2          6.66M    100     0       11653169 (+61541)   335
2          6.66M    100     0       11706550 (+53381)   335
2          6.54M    100     0       11758831 (+52281)   335
2          6.66M    100     0       11803132 (+44301)   335
2          6.66M    100     0       11854183 (+51051)   335
```

可以使用 `-i` 参数调整输出间隔。

## 扫描大 KEY
遇到 Redis 偶然卡顿问题，第一个想到的就是实例中是否存在大 KEY，大 KEY 的内存扩容以及释放都会导致主线程卡顿。
`--bigkeys` 参数可以很快扫出内存里的大 KEY，使用 `-i` 参数控制扫描间隔，避免扫描指令导致服务器的 ops 陡增报警。

```sh
$ ./redis-cli --bigkeys -i 0.01
# Scanning the entire keyspace to find biggest keys as well as
# average sizes per key type.  You can use -i 0.1 to sleep 0.1 sec
# per 100 SCAN commands (not usually needed).

[00.00%] Biggest zset   found so far 'hist:aht:main:async_finish:20180425:17' with 1440 members
[00.00%] Biggest zset   found so far 'hist:qps:async:authorize:20170311:27' with 2465 members
[00.00%] Biggest hash   found so far 'job:counters:6ya9ypu6ckcl' with 3 fields
[00.01%] Biggest string found so far 'rt:aht:main:device_online:68:{-4}' with 4 bytes
[00.01%] Biggest zset   found so far 'machine:load:20180709' with 2879 members
[00.02%] Biggest string found so far '6y6fze8kj7cy:{-7}' with 90 bytes
```

redis-cli 对于每一种对象类型都会记录长度最大的 KEY，对于每一种对象类型，刷新一次最高记录就会立即输出一次。它能保证输出长度
为 Top1 的 KEY，但是 Top2、Top3 等 KEY 是无法保证可以扫描出来的。一般的处理方法是多扫描几次，或者是消灭了 Top1 的 KEY 之后再扫
描确认还有没有次大的 KEY。

## 采样服务器指令
现在线上有一台 Redis 服务器的 OPS 太高，有很多业务模块都在使用这个 Redis，如何才能判断出来是哪个业务导致了 OPS 异常的高。这
时可以对线上服务器的指令进行采样，观察采样的指令大致就可以分析出 OPS 占比高的业务点。这时就要使用 monitor 指令，它会将服务器瞬间执行的指令全
部显示出来。不过使用的时候要注意即使使用 ctrl+c 中断，否则你的显示器会噼里啪啦太多的指令瞬间让你眼花缭乱。

```sh
$ redis-cli --host 192.168.x.x --port 6379 monitor
1539853410.458483 [0 10.100.90.62:34365] "GET" "6yax3eb6etq8:{-7}"
1539853410.459212 [0 10.100.90.61:56659] "PFADD" "growth:dau:20181018" "2klxkimass8w"
1539853410.462938 [0 10.100.90.62:20681] "GET" "6yax3eb6etq8:{-7}"
1539853410.467231 [0 10.100.90.61:40277] "PFADD" "growth:dau:20181018" "2kei0to86ps1"
1539853410.470319 [0 10.100.90.62:34365] "GET" "6yax3eb6etq8:{-7}"
1539853410.473927 [0 10.100.90.61:58128] "GET" "6yax3eb6etq8:{-7}"
1539853410.475712 [0 10.100.90.61:40277] "PFADD" "growth:dau:20181018" "2km8sqhlefpc"
1539853410.477053 [0 10.100.90.62:61292] "GET" "6yax3eb6etq8:{-7}"
```

## 诊断服务器时延
平时诊断两台机器的时延一般是使用 Unix 的 ping 指令。Redis 也提供了时延诊断指令，不过它的原理不太一样，它是诊断当前机器和 Redis 服务
器之间的指令(PING 指令)时延，它不仅仅是物理网络的时延，还和当前的 Redis 主线程是否忙碌有关。如果你发现 Unix 的 ping 指令时延很小，
而 Redis 的时延很大，那说明 Redis 服务器在执行指令时有微弱卡顿。

```sh
$ redis-cli --host 192.168.x.x --port 6379 --latency
min: 0, max: 5, avg: 0.08 (305 samples)
```

时延单位是 ms。redis-cli 还能显示时延的分布情况，而且是图形化输出。
```sh
$ redis-cli --latency-dist
```

## 远程 rdb 备份
执行下面的命令就可以将远程的 Redis 实例备份到本地机器，远程服务器会执行一次 bgsave 操作，然后将 rdb 文件传输到客户端。
```sh
$ ./redis-cli --host 192.168.x.x --port 6379 --rdb ./user.rdb
SYNC sent to master, writing 2501265095 bytes to './user.rdb'
Transfer finished with success.
```

## 模拟从库
如果你想观察主从服务器之间都同步了那些数据，可以使用 redis-cli 模拟从库。
```sh
$ ./redis-cli --host 192.168.x.x --port 6379 --slave
SYNC with master, discarding 51778306 bytes of bulk transfer...
SYNC done. Logging commands from master.
...
```

从库连上主库的第一件事是全量同步，所以看到上面的指令卡顿这很正常，待首次全量同步完成后，就会输出增量的 aof 日志。