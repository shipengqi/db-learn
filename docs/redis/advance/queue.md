# 延时队列
RabbitMQ 和 Kafka 是常用的消息队列中间件，但是这种专业的消息队列中间件使用复杂。

对于那些只有一组消费者的消息队列，使用 Redis 就可以轻松搞定。Redis 的消息队列没有 ACK 保证，如果追求消息的可靠性，还是使用专业的消息队列。

## 异步消息队列

使用 Redis 的 list(列表) 数据结构来实现异步消息队列，使用`rpush/lpush`操作入队列，使用`lpop/rpop`来出队列。

```sh
> rpush notify-queue apple banana pear
(integer) 3
> llen notify-queue
(integer) 3
> lpop notify-queue
"apple"
> llen notify-queue
(integer) 2
> lpop notify-queue
"banana"
> llen notify-queue
(integer) 1
> lpop notify-queue
"pear"
> llen notify-queue
(integer) 0
> lpop notify-queue
(nil)
```
上面是 `rpush` 和 `lpop` 结合使用的例子。还可以使用 `lpush` 和 `rpop` 结合使用，效果是一样的。这里不再赘述。

## 空队列处理
客户端是通过队列的 pop 操作来获取消息，然后进行处理。可是如果队列空了，客户端就会陷入 pop 的死循环，不停地 pop，没有数据，
接着再 pop，又没有数据。这就是浪费生命的空轮询。空轮询不但拉高了客户端的 CPU，redis 的 QPS 也会被拉高，如果这样空轮询的客户端有几十来个，Redis 的慢查询可能会显著增多。

用 `blpop/brpop` 替代前面的 `lpop/rpop`，就可以解决。这两个指令的前缀字符 `b` 代表的是 `blocking`，也就是**阻塞读**。

## 空闲连接自动断开
使用阻塞读还有其他的问题要解决 —— **空闲连接**的问题。

如果线程一直阻塞在那里，Redis 的客户端连接就成了闲置连接，闲置过久，服务器一般会主动断开连接，减少闲置资源占用。这个时候 `blpop/brpop` 会抛出异常。

所以编写客户端消费者的时候要小心，注意捕获异常，还要重试。

## 锁冲突处理
对于分布式锁，如果加锁失败，异步消息可以使用延时队列，将当前冲突的请求扔到另一个队列延后处理以避开冲突。

## 延时队列的实现
延时队列可以通过 Redis 的 `zset`(有序列表) 来实现。我们将消息序列化成一个字符串作为 `zset` 的 `value`，这个消息的到期处理时间作为 `score`，然后用多个线程轮询 `zset` 获取
到期的任务进行处理，多个线程是为了保障可用性，万一挂了一个线程还有其它线程可以继续处理。因为有多个线程，所以需要考虑并发争抢任务，确保任务不能被多次执行。

```py
def delay(msg):
    msg.id = str(uuid.uuid4())  # 保证 value 值唯一
    value = json.dumps(msg)
    retry_ts = time.time() + 5  # 5 秒后重试
    redis.zadd("delay-queue", retry_ts, value)


def loop():
    while True:
        # 最多取 1 条
        values = redis.zrangebyscore("delay-queue", 0, time.time(), start=0, num=1)
        if not values:
            time.sleep(1)  # 延时队列空的，休息 1s
            continue
        value = values[0]  # 拿第一条，也只有一条
        success = redis.zrem("delay-queue", value)  # 从消息队列中移除该消息
        if success:  # 因为有多进程并发的可能，最终只会有一个进程可以抢到消息
            msg = json.loads(value)
            handle_msg(msg) # 异常捕获，避免因为个别任务处理问题导致循环异常退出
```

Redis 的 **`zrem` 方法是多线程多进程争抢任务的关键，它的返回值决定了当前实例有没有抢到任务，因为 `loop` 方法可能会被多个线程、多个进程调用，
同一个任务可能会被多个进程线程抢到，通过 `zrem` 来决定唯一的属主**。