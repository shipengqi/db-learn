---
title: Change Stream 实战
weight: 6
---

**Change Stream 指数据的变化事件流**，MongoDB 从 3.6 版本开始提供**订阅数据变更的功能**。

## Change Stream 的实现原理

**Change Stream 是基于 oplog 实现的，提供推送实时增量的推送功能**。它在 oplog 上开启一个 tailable cursor 来追踪所有复制集上的变更操作，最终调用应用中定义的回调函数。

被追踪的变更事件主要包括：

- `insert/update/delete`：插入、更新、删除； 
- `drop`：集合被删除； 
- `rename`：集合被重命名；
- `dropDatabase`：数据库被删除；
- `invalidate`：`drop/rename/dropDatabase` 将导致 `invalidate` 被触发，并关闭 change stream；

从 MongoDB 6.0 开始，change stream 支持 DDL 事件的更改通知，如 `createIndexes` 和 `dropIndexes` 事件。

![mongodb-change-stream](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-change-stream.png)


如果只对某些类型的变更事件感兴趣，可以**使用聚合管道的过滤步骤过滤事件**：

```javascript
var cs = db.user.watch([{
    $match:{operationType:{$in:["insert","delete"]}}
}])
```

`db.watch()` 语法：https://www.mongodb.com/docs/manual/reference/method/db.watch/#example。

**Change Stream 会采用 `"readConcern：majority"` 这样的一致性级别，保证写入的变更不会被回滚**。因此：

- 未开启 majority readConcern 的集群无法使用 Change Stream；
- 当集群无法满足 `{w: "majority"}` 时，不会触发 Change Stream（例如 PSA 架构 中的 S 因故障宕机）。

### MongoShell 测试

窗口 1：

```javascript
db.user.watch([],{maxAwaitTimeMS:1000000})
```

窗口 2：

```javascript
db.user.insert({name:"xxxx"})
```

![mongodb-change-event](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-change-event.png)

变更字段说明：

![mongodb-change-event-field](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-change-event-field.png)

## 故障恢复

假设在一系列写入操作的过程中，订阅 Change Stream 的应用在接收到 “写3” 之后 于 t0 时刻崩溃，重启后后续的变更怎么办？

![mongodb-change-event-fail](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/mongodb-change-event-fail.png)


**想要从上次中断的地方继续获取变更流，只需要保留上次变更通知中的 `_id` 即可**。Change Stream 回调所返回的的数据带有 `_id`，这个 `_id` 可以用于断点恢复。例如： 

```javascript
var cs = db.collection.watch([], {resumeAfter: <_id>}) 
```

即可从上一条通知中断处继续获取后续的变更通知。

## 使用场景

- 监控 

用户需要及时获取变更信息（例如账户相关的表），ChangeStreams 可以提供监控功能，一旦相关的表信息发生变更，就会将变更的消息实时推送出去。 

- 分析平台 

例如需要基于增量去分析用户的一些行为，可以基于 ChangeStreams 把数据拉出来，推到下游的计算平台， 比如类似 Flink、Spark 等计算平台等等。

- 数据同步 

基于 ChangeStreams，用户可以搭建额外的 MongoDB 集群，这个集群是从原端的 MongoDB 拉取过来的， 那么这个集群可以做一个热备份，假如源端集群发生网络不通等等之类的变故，备集群就可以接管服务。还可以做一个冷备份，如用户基于 ChangeStreams 把数据同步到文件，万一源端数据库发生不可服务， 就可以从文件里恢复出完整的 MongoDB 数据库，继续提供服务。（当然，此处还需要借助定期全量备份来一同完成恢复）另外数据同步它不仅仅局限于同一地域，可以跨地域，从北京到上海甚至从中国到美国等等。 

- 消息推送 

假如用户想实时了解公交车的信息，那么公交车的位置每次变动，都实时推送变更的信息给想了解的用户，用户能够实时收到公交车变更的数据，非常便捷实用。 

{{< callout type="info" >}}
- Change Stream 依赖于 oplog，因此中断时间不可超过 oplog 回收的最大时间窗；  
- 在执行 `update` 操作时，如果只更新了部分数据，那么 Change Stream 通知的也是增量部分； 
- 删除数据时通知的仅是删除数据的 `_id`。
{{< /callout >}}