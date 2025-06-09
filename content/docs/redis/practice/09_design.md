---
title: 缓存设计
weight: 9
---

## 简单的冷热分离实现

比如一个电商网站，商品可能会有很多，但是真正热门的，每天都有人访问的商品可能不足 `1%`，对于这种热门的商品，可以延长其缓存的有效期，这样可以减少数据库的访问次数，提高系统的性能。

示例代码：

```go
var (
    ValidityDuration = 24 * time.Hour // 缓存有效期
)

func GetProduct(id int) (*Product, error) {
    // 从缓存中获取商品信息
    product, err := GetProductFromCache(id)
    if err == nil {
        // 如果缓存中存在商品信息，延长有效期
        UpdateProductExpireTime(id, ValidityDuration)
        return product, nil
    }   
    // 从数据库中获取商品信息
    product, err = GetProductFromDB(id)
    if err != nil {
        return nil, err
    }   
    // 将商品信息存入缓存，有效期为 24 小时
    SetProductToCache(product, ValidityDuration)
    return product, nil
}
```

上面的代码中，**只要商品被访问过，并且在缓存中，那么就会延长其有效期**，这样可以保证热门的商品一直存在于缓存中。


## 缓存失效（击穿）

缓存失效（击穿）是指由于大批量缓存在同一时间失效可能导致大量请求同时穿透缓存直达数据库，可能会造成数据库瞬间压力过大甚至挂掉。例如电商系统中，如果有一大批商品同时上架，这批商品的缓存数据可能会在同一时间失效。

### 解决方案

可以在批量增加缓存时，对于这一批数据中的每一个 key 的缓存过期时间都增加为一个随机的值，这样的话，每个 key 的过期时间都不同，从而避免了大量缓存同时失效的问题。

```java
String get(String key) {
    // 从缓存中获取数据
    String cacheValue = cache.get(key);
    // 缓存为空
    if (StringUtils.isBlank(cacheValue)) {
        // 从存储中获取
        String storageValue = storage.get(key);
        cache.set(key, storageValue);
        // 设置一个过期时间(300到600之间的一个随机数)
        int expireTime = new Random().nextInt(300)  + 300;
        if (storageValue == null) {
            cache.expire(key, expireTime);
        }
        return storageValue;
    } else {
        // 缓存非空
        return cacheValue;
    }
}
```

## 缓存穿透

缓存穿透是指**查询一个根本不存在的数据**（**缓存击穿的区别就在于数据至少在数据库中还是存在的**，击穿只是击穿了缓存层，**穿透是整个后端都被穿透**了），缓存层和存储层都不会命中，通常出于容错的考虑，如果从存储层查不到数据则不写入缓存层。

缓存穿透将导致不存在的数据每次请求都要到存储层去查询，失去了缓存保护后端存储的意义。

造成缓存穿透的基本原因有两个：

1. 自身业务代码或者数据出现问题。
2. 一些恶意攻击、 爬虫等造成大量空命中。 

### 解决方案

#### 缓存空对象

如果一个查询返回的数据为空（不管是数据是否不存在），仍然把这个空结果（null）进行缓存，并且设置一个过期时间。

```java
String get(String key) {
    // 从缓存中获取数据
    String cacheValue = cache.get(key);
    // 缓存为空
    if (StringUtils.isBlank(cacheValue)) {
        // 从存储中获取
        String storageValue = storage.get(key);
        cache.set(key, storageValue);
        // 如果存储数据为空， 需要设置一个过期时间 300s
        if (storageValue == null) {
            cache.expire(key, 60 * 5);
        }
        return storageValue;
    } else {
        // 缓存非空
        return cacheValue;
    }
}
```

如果是被恶意攻击，每次攻击可能都会换不一样的 key，如果缓存中存储上百万个空值，占用了大量的内存空间。可以**为空值缓存设置一个短的过期时间。对于空值缓存，也需要设置延期，避免同一个空值的 key 被不停的访问**。

#### 布隆过滤器

对于恶意攻击，向服务器请求大量不存在的数据造成的缓存穿透，还可以用布隆过滤器先做一次过滤，对于不存在的数据布隆过滤器一般都能够过滤掉，不让请求再往后端发送。当布隆过滤器说**某个值存在时，这个值可能不存在；当它说不存在时，那就肯定不存在**。

布隆过滤器就是一个大型的位数组和一组的无偏 `hash` 函数。所谓无偏就是能够把元素的 `hash` 值算得比较均匀。

向布隆过滤器中添加 key 时，会使用多个 hash 函数对 key 进行 hash 算得一个整数索引值然后对位数组长度进行取模运算得到一个位置，每个 `hash` 函数都会算得一个不同的位置。再把位数组的这几个位置都置为 1 就完成了 add 操作。

向布隆过滤器询问 key 是否存在时，跟 add 一样，也会把 hash 的几个位置都算出来，看看位数组中这几个位置是否都为 1，只要有一个位为 0，那么说明布隆过滤器中这个 key 不存在。如果都是 1，这并不能说明这个 key 就一定存在，只是极有可能存在，因为这些位被置为 1 可能是因为其它的 key 存在所致。如果这个位数组长度比较大，存在概率就会很大，如果这个位数组长度比较小，存在概率就会降低。

可以用 redisson 实现布隆过滤器：

```java
package com.redisson;

import org.redisson.Redisson;
import org.redisson.api.RBloomFilter;
import org.redisson.api.RedissonClient;
import org.redisson.config.Config;

public class RedissonBloomFilter {

    public static void main(String[] args) {
        Config config = new Config();
        config.useSingleServer().setAddress("redis://localhost:6379");
        // 构造 Redisson
        RedissonClient redisson = Redisson.create(config);

        RBloomFilter<String> bloomFilter = redisson.getBloomFilter("nameList");
        // 初始化布隆过滤器：预计元素为 100000000L,误差率为 3%,根据这两个参数会计算出底层的 bit 数组大小
        bloomFilter.tryInit(100000000L,0.03);
        //将 zhuge 插入到布隆过滤器中
        bloomFilter.add("zhuge");

        // 判断下面号码是否在布隆过滤器中
        System.out.println(bloomFilter.contains("guojia"));//false
        System.out.println(bloomFilter.contains("baiqi"));//false
        System.out.println(bloomFilter.contains("zhuge"));//true
    }
}
```

使用布隆过滤器需要把所有数据提前放入布隆过滤器，并且在增加数据时也要往布隆过滤器里放，布隆过滤器缓存过滤伪代码：

```java
//初始化布隆过滤器
RBloomFilter<String> bloomFilter = redisson.getBloomFilter("nameList");
//初始化布隆过滤器：预计元素为100000000L,误差率为3%
bloomFilter.tryInit(100000000L,0.03);
        
//把所有数据存入布隆过滤器
void init(){
    for (String key: keys) {
        bloomFilter.put(key);
    }
}

String get(String key) {
    // 从布隆过滤器这一级缓存判断下 key 是否存在
    Boolean exist = bloomFilter.contains(key);
    if(!exist){
        return "";
    }
    // 从缓存中获取数据
    String cacheValue = cache.get(key);
    // 缓存为空
    if (StringUtils.isBlank(cacheValue)) {
        // 从存储中获取
        String storageValue = storage.get(key);
        cache.set(key, storageValue);
        // 如果存储数据为空， 需要设置一个过期时间(300秒)
        if (storageValue == null) {
            cache.expire(key, 60 * 5);
        }
        return storageValue;
    } else {
        // 缓存非空
        return cacheValue;
    }
}
```

{{< callout type="info" >}}
这种方法**适用于数据命中不高、数据相对固定（因为添加删除元素需要重建）、实时性低（通常是数据集较大）的应用场景**，代码维护较为复杂，但是缓存空间占用很少。
{{< /callout >}}

## 热点缓存 key 的重建优化

开发人员使用 “缓存+过期时间” 的策略既可以加速数据读写，又保证数据的定期更新，这种模式基本能够满足绝大部分需求。但是有两个问题如果同时出现，可能就会对应用造成致命的危害：

1. 当前 key 是一个热点 key（例如一个热门的娱乐新闻），并发量非常大。
2. 重建缓存不能在短时间完成，可能是一个复杂计算，例如复杂的 SQL、多次 IO、多个依赖等。

**在缓存失效的瞬间，有大量线程来重建缓存，造成后端负载加大**，甚至可能会让应用崩溃。

### 解决方案

要解决这个问题主要就是要避免大量线程同时重建缓存。可以**利用互斥锁，（多个服务实例，使用分布式锁）**来解决，此方法**只允许一个线程重建缓存**，其他线程等待重建缓存的线程执行完，重新从缓存获取数据即可。

```java
String get(String key) {
    // 从 Redis 中获取数据
    String value = redis.get(key);
    // 如果 value 为空， 则开始重构缓存
    if (value == null) {
        // 只允许一个线程重建缓存， 使用 nx， 并设置过期时间 ex
        String mutexKey = "mutex:key:" + key;
        if (redis.set(mutexKey, "1", "ex 180", "nx")) {
             // 从数据源获取数据
            value = db.get(key);
            // 回写 Redis， 并设置过期时间
            redis.setex(key, timeout, value);
            // 删除 key_mutex
            redis.delete(mutexKey);
        } else {
            // 其他线程休息 50 毫秒后重试，重试时缓存中已经有值了
            Thread.sleep(50);
            get(key);
        }
    }
    return value;
}
```

**对于不同的商品，可以使用不同的 key， 避免不同的商品竞争同一把锁**，提高并发度。

{{< callout type="info" >}}
这里要注意，如果使用了缓存空对象来解决缓存穿透的问题，那么这里在判断缓存为空的时候，要区分一下是真的不存在，还是缓存的空对象，如果是缓存的空对象，就不需要去重建缓存了，因为数据库里也没有。
{{< /callout >}}


## 缓存与数据库双写不一致

在大并发下，同时操作数据库与缓存会存在数据不一致性问题：

<img src="https://raw.gitcode.com/shipengqi/illustrations/files/main/db/db-cache-demo.png" width="280px">


上图线程 1 先执行了更新数据库的操作，但是卡了一会还没来得及更新缓存，然后线程 2 也执行了更新数据库的操作并且更新的缓存，最后线程 1 更新了缓存。

最后数据库中的 `stock=6` 而缓存中的 `stock=10`。这就是缓存与数据库的数据不一致问题。

有些业务实现，在写完数据库之后可能不会去更新缓存，而是删除缓存，在查询数据库的时候再去更新缓存。这种方式也有一样的问题：

<img src="https://raw.gitcode.com/shipengqi/illustrations/files/main/db/db-cache-demo2.png" width="380px">

图中，线程 1 写入数据库 `stock=10` 并删除缓存，然后线程 3 查询数据缓存为空，接着查询数据库得到 `stock=10`，这个时候如果在线程 3 更新缓存之前，线程 2 写入数据库 `stock=6` 并删除缓存，最后线程 3 写入缓存 `stock=10`。一样的缓存与数据库的数据不一致问题。

### 解决方案

问题主要是出在了查询数据库和更新缓存之间，在高并发的场景下，可能会出现别的线程更新数据库的操作。

直接使用**分布式锁**就能解决这种问题。

1. 对于并发几率很小的数据(如个人维度的订单数据、用户数据等)，这种几乎不用考虑这个问题，很少会发生缓存不一致，可以给缓存数据加上过期时间，每隔一段时间触发读的主动更新即可。
2. 就算并发很高，如果业务上能容忍短时间的缓存数据不一致(如商品名称，商品分类菜单等)，缓存加上过期时间依然可以解决大部分业务对于缓存的要求。
3. 如果不能容忍缓存数据不一致，可以通过加**分布式读写锁**来保证并发读写或写写的时候按顺序排好队，读读的时候相当于无锁。
4. 也可以用阿里开源的 canal 通过监听数据库的 binlog 日志及时的去修改缓存，但是引入了新的中间件，增加了系统的复杂度。

## 缓存雪崩

缓存雪崩指的是缓存层支撑不住或宕掉后，流量会像奔逃的野牛一样，打向后端存储层。

由于缓存层承载着大量请求，有效地保护了存储层，但是如果缓存层由于某些原因不能提供服务(比如超大并发过来，缓存层支撑不住，或者由于缓存设计不好，类似大量请求访问 bigkey，导致缓存能支撑的并发急剧下降)，于是大量请求都会打到存储层，存储层的调用量会暴增，造成存储层也会级联宕机的情况。 

### 解决方案

预防和解决缓存雪崩问题， 可以从以下三个方面进行着手：

1. 保证缓存层服务**高可用**性，比如使用 Redis Sentinel 或 Redis Cluster。
2. 依赖隔离组件为后端**限流熔断并降级**。比如使用 Sentinel 或 Hystrix **限流降级**组件。比如服务降级，我们可以针对不同的数据采取不同的处理方式。当业务应用访问的是非核心数据（例如电商商品属性，用户信息等）时，暂时停止从缓存中查询这些数据，而是直接返回预定义的默认降级信息、空值或是错误提示信息；当业务应用访问的是核心数据（例如电商商品库存）时，仍然允许查询缓存，如果缓存缺失，也可以继续通过数据库读取。
3. **多级缓存**，`进程内存 -> Redis -> 数据库`。对于进程内存缓存，可以使用一些轻量级的缓存组件，比如 Google 的 Guava Cache 或者 Caffeine，这些组件都实现了进程内缓存，并且支持多种缓存过期策略。可以避免内存泄露的问题。**多级缓存架构也会有数据不一致的问题**，可以通过异步的方式来更新缓存。不过一点点的不一致是可以接受的，没有必要继续增加系统的复杂性。真正实践中会有一个独立的 HotKey 监测系统来监控热点 key 的，然后将热点 key 加入到多级缓存中。
4. **提前演练**。 在项目上线前，演练缓存层宕掉后，应用以及后端的负载情况以及可能出现的问题，在此基础上做一些预案设定。

## 总结

针对**读多写少的情况加入缓存提高性能**，如果**写多读多的情况又不能容忍缓存数据不一致，那就没必要加缓存了**，可以直接操作数据库。当然，如果数据库抗不住压力，还可以把缓存作为数据读写的主存储，异步将数据同步到数据库，数据库只是作为数据的备份。
放入缓存的数据应该是对实时性、一致性要求不是很高的数据。切记不要为了用缓存，同时又要保证绝对的一致性做大量的过度设计和控制，增加系统复杂性。