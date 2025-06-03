---
title: GeoHash
weight: 8
---

## GeoHash 算法

GeoHash 是一种地理位置编码的方法。GeoHash 算法将二维的经纬度数据映射到一维的整数，这样所有的元素都将在挂载到一条线上，距离靠近的二维坐标映射到一维后，两点之间距离也会很接近。当我们想要计算**附近的人**时，首先将目标位置映射到这条线上，然后在这个一维的线上获取附近的点就行了。

地图元素的位置数据使用**二维的经纬度**表示，经度范围 `(-180, 180]`，纬度范围 `(-90, 90]`，纬度正负以赤道为界，北正南负，经度正负以本初子午线 (英国格林尼治天文台) 为界，东正西负。

如果纬度范围 `[-90, 0)` 用二进制 0 表示，`(0, 90]` 用二进制 1 表示。经度范围 `[-180, 0)` 用二进制 0 表示，`(0, 180]` 用二进制 1 表示。那么地球可以分为 4 个区域：

<img src="https://raw.gitcode.com/shipengqi/illustrations/files/main/db/geo-earth.png" width="280px">

- `00` 第一个 0 表示纬度 `[-90, 0)`，第二个 0 表示经度 `[-180, 0)`。
- `10` 0 表示纬度 `[-90, 0)`，1 表示经度 `(0, 180]`。
- `01` 1 表示纬度 `(0, 90]`，0 表示经度 `[-180, 0)`。
- `11` 第一个 1 表示纬度 `(0, 90]`，第二个 1 表示经度 `(0, 180]`。

分成 4 个区域之后，大概可以知道在地球的哪个方位了。如果想要更精确的定位，就可以继续切分，比如把 `00` 继续分成 4 个区域。

<img src="https://raw.gitcode.com/shipengqi/illustrations/files/main/db/geo-earth2.png" width="280px">

- `00` 分成了 4 个区域，这四个区域前面的 `00` 就是父区域的编码。子区域的四个编码还是 `00`、`01`、`10`、`11`，再加上父区域的编码作为前缀，就得到了 `0000`、`0001`、`0010`、`0011`。
- 有共同前缀的区域，可以理解为距离也是比价近的。

继续切下去，正方形就会越来越小，二进制整数也会越来越长，精确度就会越来越高。

通过上述的过程，最终会得到一串二进制的编码。这串二进制编码通常很长，不便于阅读和传输。为了使得 Geohash 更加紧凑和易于使用，通常会将这串二进制编码转换成 Base32 编码。

在 Redis 里面，将这个编码值放了 `zset` 的 `score` 中。

在使用 Redis 进行 Geo 查询时，通过 `zset` 的 `score` 排序就可以得到坐标附近的其它元素，通过将 `score` 还原成坐标值就可以得到元素的原始坐标。

## Geo 指令

### geoadd

`geoadd` 指令携带集合名称以及多个经纬度名称三元组，注意这里可以加入多个三元组：

```sh
127.0.0.1:6379> geoadd company 116.48105 39.996794 juejin
(integer) 1
127.0.0.1:6379> geoadd company 116.514203 39.905409 ireader
(integer) 1
127.0.0.1:6379> geoadd company 116.489033 40.007669 meituan
(integer) 1
127.0.0.1:6379> geoadd company 116.562108 39.787602 jd 116.334255 40.027400 xiaomi
(integer) 2
```

**geo 存储结构上使用的是 zset，意味着可以使用 zset 相关的指令来操作 geo 数据，所以删除指令可以直接使用 zrem 指令**。

### geodist

`geodist` 指令可以用来计算两个元素之间的距离，携带集合名称、2 个名称和距离单位。

```sh
127.0.0.1:6379> geodist company juejin ireader km
"10.5501"
127.0.0.1:6379> geodist company juejin meituan km
"1.3878"
127.0.0.1:6379> geodist company juejin jd km
"24.2739"
127.0.0.1:6379> geodist company juejin xiaomi km
"12.9606"
127.0.0.1:6379> geodist company juejin juejin km
"0.0000"
```

掘金离美团最近，因为它们都在望京。距离单位可以是 `m`、`km`、`ml`、`ft`，分别代表米、千米、英里和尺。

### geopos

`geopos` 指令可以获取集合中任意元素的经纬度坐标，可以一次获取多个。

```sh
127.0.0.1:6379> geopos company juejin
1) 1) "116.48104995489120483"
   2) "39.99679348858259686"
127.0.0.1:6379> geopos company ireader
1) 1) "116.5142020583152771"
   2) "39.90540918662494363"
127.0.0.1:6379> geopos company juejin ireader
1) 1) "116.48104995489120483"
   2) "39.99679348858259686"
2) 1) "116.5142020583152771"
   2) "39.90540918662494363"
```

获取的经纬度坐标和 `geoadd` 进去的坐标有轻微的误差，原因是 `geohash` 对二维坐标进行的一维映射是有损的，通过映射再还原回来的值会出现较小的差别。

### geohash

`geohash` 可以获取元素的经纬度编码字符串，它是 base32 编码。

```sh
127.0.0.1:6379> geohash company ireader
1) "wx4g52e1ce0"
127.0.0.1:6379> geohash company juejin
1) "wx4gd94yjn0"
```

可以使用 [geohash](http://geohash.org/) 来解编码。格式 `http://geohash.org/wx4g52e1ce0`。

### georadiusbymember

`georadiusbymember` 指令是最为关键的指令，它可以用来查询指定元素附近的其它元素。

```sh
# 范围 20 公里以内最多 3 个元素按距离正排，它不会排除自身
127.0.0.1:6379> georadiusbymember company ireader 20 km count 3 asc
1) "ireader"
2) "juejin"
3) "meituan"
# 范围 20 公里以内最多 3 个元素按距离倒排
127.0.0.1:6379> georadiusbymember company ireader 20 km count 3 desc
1) "jd"
2) "meituan"
3) "juejin"
# 三个可选参数 withcoord withdist withhash 用来携带附加参数
# withdist 很有用，它可以用来显示距离
127.0.0.1:6379> georadiusbymember company ireader 20 km withcoord withdist withhash count 3 asc
1) 1) "ireader"
   2) "0.0000"
   3) (integer) 4069886008361398
   4) 1) "116.5142020583152771"
      2) "39.90540918662494363"
2) 1) "juejin"
   2) "10.5501"
   3) (integer) 4069887154388167
   4) 1) "116.48104995489120483"
      2) "39.99679348858259686"
3) 1) "meituan"
   2) "11.5748"
   3) (integer) 4069887179083478
   4) 1) "116.48903220891952515"
      2) "40.00766997707732031"
```

### georadius

`georadius` 可以用来查询指定坐标附近的其它元素。参数和 `georadiusbymember` 基本一致。

```sh
127.0.0.1:6379> georadius company 116.514202 39.905409 20 km withdist count 3 asc
1) 1) "ireader"
   2) "0.0000"
2) 1) "juejin"
   2) "10.5501"
3) 1) "meituan"
   2) "11.5748"
```

## 注意

地图应用中，车的数据、餐馆的数据可能会有百万千万条，如果使用 Redis 的 Geo 数据结构，它们将全部放在一个 `zset` 集合中。在 Redis 的集群环境中，集合可能会从一个节点迁移到另一个节点，如果单个 key 的数据过大，会对集群的迁移工作造成较大的影响，在集群环境中单个 key 对应的数据量**不宜超过 1M**，否则会导致集群迁移出现卡顿现象，影响线上服务的正常运行。

建议 Geo 的数据使用单独的 Redis 实例部署，不使用集群环境。

如果数据量过亿甚至更大，就需要**对 Geo 数据进行拆分，按国家拆分、按省拆分，按市拆分，在人口特大城市甚至可以按区拆分**。这样就可以显著降低单个 `zset` 集合的大小。
