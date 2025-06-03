---
title: 热点缓存探测系统
weight: 10
---

## 多级缓存

一般多级缓存分为：

1. 本地缓存
2. 远程缓存

本地缓存的优势：

1. 可以减少网络请求，提高性能。
2. 减少远程缓存的压力。

本地缓存的缺点：

1. 进程空间的大小有限，不能存储大量的数据。
2. 进程重启后，本地缓存会丢失。
3. 分布式场景下，本地缓存会出现数据不一致的问题。
4. 和远程缓存的一致性问题。

对于数据不一致的问题，其实只要保证最终一致性即可。缩短本地缓存的过期时间，根据业务更够接收不一致的时间来设置，比如 10s 或者更短的过期时间。

### 多级缓存的使用场景

- 热点的商品详情页
- 热搜
- 热门帖子
- 热门用户主页

一般都是在高并发的场景下使用。

热点产生的条件：

1. 有限时间
2. 流量高聚

在互联网领域，热点被分为 2 类：

1. 有预期的热点：比如在电商活动中退出的爆款联名限量款商品，又或者是秒杀会场活动等。
2. 无预期的热点：比如受到了黑客的恶意攻击，网络爬虫的频繁访问，又或者突发新闻带来的流量冲击等。

对于有预期的热点，我们可以通过提前预热，提前把数据加载到缓存中，或者提前扩容，降级等方式来解决。

对于无预期的热点，就需要热点探测系统来探测热点，在热点还没有爆火之前探测出来，提前把数据加载到缓存中，进行扩容等。

### 热点探测使用场景

- MySQL 中被频繁访问的数据 ，如热门商品的主键 id
- Redis 缓存中被密集访问的 Key，如热门商品的详情需要 `get goods$id`
- 恶意攻击或机器人爬虫的请求信息，如特定标识的 userId、机器 IP
- 频繁被访问的接口地址，如获取用户信息接口 `/userInfo/ + userId`

#### 使用热点探测的好处

提升性能，规避风险。

对于无预期的热数据（即突发场景下形成的热 Key），可能会对业务系统带来极大的风险，可将风险分为两个层次：

1. 对数据层的风险

正常情况下，Redis 缓存单机就可支持十万左右 QPS，并能通过集群部署提高整体负载能力。对于并发量一般的系统，用 Redis 做缓存就足够了。但是对于瞬时过高并发的请求，因为 Redis 单线程原因会导致正常请求排队，或者因为热点集中导致分片集群压力过载而瘫痪，从而击穿到 DB 引起服务器雪崩。

2. 对应用服务的风险

每个应用在单位时间所能接受和处理的请求量是有限的，如果受到恶意请求的攻击，让恶意用户独自占用了大量请求处理资源，就会导致其他正常用户的请求无法及时响应。


因此，需要一套动态热 Key 检测机制，通过对需要检测的热 Key 规则进行配置，实时监听统计热 Key 数据，当无预期的热点数据出现时，第一时间发现他，并针对这些数据进行特殊处理。如本地缓存、拒绝恶意用户、接口限流/降级等

### 如何实现热点探测

热点产生的条件是 2 个：一个**时间**，一个**流量**。那么根据这个条件可以简单定义一个规则：比如 1 秒内访问 1000 次的数据算是热数据，当然这个数据需要根据具体的业务场景和过往数据进行具体评估。

对于单机应用，检测热数据很简单，直接在本地为每个 Key 创建一个滑动窗口计数器，统计单位时间内的访问总数（频率），并通过一个集合存放检测到的热 Key。

![hotkey-window]()

对于分布式应用，对热 Key 的访问是分散在不同的机器上的，无法在本地独立地进行计算，因此，需要一个独立的、集中的热 Key 计算单元。

可以分为五个步骤：

1. 热点规则：配置热 Key 的上报规则，圈出需要重点监测的 Key
2. 热点上报：应用服务将自己的热 Key 访问情况上报给集中计算单元
3. 热点统计：收集各应用实例上报的信息，使用滑动窗口算法计算 Key 的热度
4. 热点推送：当 Key 的热度达到设定值时，推送热 Key 信息至所有应用实例
5. 热点缓存：各应用实例收到热 Key 信息后，对 Key 值进行本地缓存

#### 单机应用示例

```java
public class HotKeyDetector {
    private final int WINDOW_SIZE = 10; // 滑动窗口大小
    private final int THRESHOLD = 5; // 阈值，达到该条件时即判定为热 Key
    private final Cache<String, Obejct> hotCache = CacheBuilder.newBuilder() // 本地缓存
            .expireAfterWrite(5, TimeUnit.SECONDS)
            .maximumSize(1000) // 缓存最大容量
            .build();

    private Map<String, Queue<Long>> window = new HashMap<>(); // 滑动窗口        
    private Map<String, Integer> counts = new HashMap<>(); // 用来计数，用来和阈值比较

    // 判断是否为热 Key
    public boolean isHotKey(String data) {
        // 如果缓存中有数据，说明已经是 hot key，直接返回 true
        if (hotCache.getIfPresent(data) != null) {
            return true;
        }
        // 获取当前数据在计数器中的统计次数
        int count = counts.getOrDefault(data, 0);
        // 如果次数大于阈值，说明是热 Key，将数据加入本地缓存，清空队列并返回 true
        if (count > THRESHOLD) {
            hotCache.put(data, data); // 加入本地缓存
            clear(data) // 清空滑动窗口中相应的队列
            return true;
        } else {
            // 如果次数小于阈值，说明不是热 Key，将数据加入滑动窗口，并返回 false
            counts.put(data, count + 1); // 次数加 1
            // 获取对应数据的时间队列
            Queue<Long> queue = window.get(data);
            // 如果队列不存在，就创建一个新的队列
            if (queue == null) {
                queue = new LinkedList<Long>();
                window.put(data, queue);
            }
            // 获取当前时间（秒）
            long currentTime = System.currentTimeMillis() / 1000;
            queue.add(currentTime); // 将当前时间加入队列，用于后面数据滑动窗口的统计
            // 如果队列中数据的时间超过了滑动窗口的时间区间，则将该时间从队列中移除
            while (!queue.isEmpty() && currentTime - queue.peek() > WINDOW_SIZE) {
                queue.poll(); // 移除队列头部的时间
                counts.put(data, counts.get(data) - 1); // 统计次数减 1
            }
            return false; // 不是热 Key，返回 false
        }
    }

    // 清除指定数据的队列和计数
    public void clear(String data) {
        window.remove(data); // 移除指定数据的队列
        counts.remove(data); // 移除指定数据的计数
    }

    // 添加数据到本地缓存
    public void set(String key, Object value) {
        hotCache.put(key, value); // 将数据加入本地缓存
    }

    // 从本地缓存中获取数据
    public Object get(String key) {
        return hotCache.getIfPresent(key); // 从本地缓存中获取数据
    }
}
```

上面并没有考虑并发安全的问题，只是简单的示例。


#### 分布式应用

- [JD-hotkey](https://gitee.com/jd-platform-opensource/hotkey)。
- [https://my.oschina.net/1Gk2fdm43/blog/4331985](https://my.oschina.net/1Gk2fdm43/blog/4331985)。

![hotkey-arch]()

该框架主要由 4 个部分组成：

1. etcd 集群

etcd 作为一个高性能的配置中心，可以以极小的资源占用，提供高效的监听订阅服务。主要用于存放规则配置，各 worker 的 ip 地址，以及探测出的热 key、手工添加的热 key 等。

2. client 端 jar 包

就是在服务中添加的引用 jar，引入后，就可以以便捷的方式去判断某 key 是否热 key。同时，该 jar 完成了 key 上报、监听 etcd 里的 rule 变化、worker 信息变化、热 key 变化，对热 key 进行本地 caffeine 缓存等。

3. worker 端集群

worker 端是一个独立部署的 Java 程序，启动后会连接 etcd，并定期上报自己的 ip 信息，供 client 端获取地址并进行长连接。之后，主要就是对各个 client 发来的待测 key 进行累加计算，当达到 etcd 里设定的 rule 阈值后，将热 key 推送到各个 client。

4. dashboard 控制台

控制台是一个带可视化界面的 Java 程序，也是连接到 etcd，之后在控制台设置各个 APP 的 key 规则，譬如 2 秒出现 20 次算热 key。然后当 worker 探测出来热 key 后，会将 key 发往 etcd，dashboard 也会监听热 key 信息，进行入库保存记录。同时，dashboard 也可以手工添加、删除热 key，供各个 client 端监听。



![hotkey-arch2]()

上图中的第一步，其实是不需要的，因为这里是热点探测系统主动将热 Key 推送给应用实例，不需要应用实例去拉取。

写操作通过 MQ 或者长连接的方式将数据推送给热点探测系统。