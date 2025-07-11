---
title: 分布式锁
weight: 8
---

分布式锁是用来解决并发问题的。比如一个操作要修改用户的状态，修改状态需要先读出用户的状态，在内存里进行修改，改完了再存回去。如果这样的操作同时进行了，就会出现并发问题，因为读取和保存状态这两个操作不是原子的。

分布式锁本质上要实现的目标就是在 Redis 里面**占一个坑，当别的进程也要来占时，发现已经有人蹲在那里了，就只好放弃或者稍后再试**。

分布式锁一般是使用 `SETNX` 命令实现，`SETNX` 命令的作用是设置一个键值对，如果键不存在，则设置成功，返回 `1`；如果键已经存在，则设置失败，返回 `0`。释放锁可以使用 `DEL` 命令删除键值对。

```bash
# 加锁
SETNX lock:order:{id} true
# 解锁
DEL lock:order:{id}
```

仅仅这么设置是不够的，因为如果逻辑执行到中间出现异常了，`DEL` 没有被调用那么锁就会一直存在，导致其他线程无法获取锁，导致死锁。

为了避免这种情况，可以使用 `EXPIRE` 命令设置锁的过期时间，比如 5s，这样即使中间出现异常也可以保证 5 秒之后锁会自动释放：

```bash
# 加锁并设置过期时间
SETNX lock:order:{id} true
EXPIRE lock:order:{id} 5
```

但是这样也会有问题，因为 `SETNX` 和 `EXPIRE` 是两个命令，它们不是原子性的，如果 `SETNX` 成功了，但是 `EXPIRE` 失败了，那么锁就会一直存在，导致死锁。

为了避免这种情况，Redis 2.8 版本中加入了 `SET` 指令的扩展参数，使得 `SETNX` 和 `EXPIRE` 可以一起执行。


```bash
# 加锁并设置过期时间
SET lock:order:{id} true EX 5 NX
```

这样就可以保证 `SETNX` 和 `EXPIRE` 是原子性的，要么都执行成功，要么都执行失败。

## 超时问题

上面的加锁方式，还是有超时问题的。

假设第一个进程逻辑执行时间执行了 15s，锁的过期时间为 10s，那么第一个 10s 后锁就会自动释放，第二个进程就可以拿到锁。

如果第二个进程在执行 5s 后还没有结束，这个时候第一个进程的逻辑执行完了，释放了锁。那这个时候第一个进程就会把第二个进程的锁释放了。在高并发场景下，会出现大量的锁被错误释放的情况，也就意味会有大量的进程可以拿到这个锁。

1. 进程 1 获取锁成功。
2. 进程 1 在某个操作上阻塞了很长时间。
3. 过期时间到了，锁自动释放了。
4. 进程 2 获取到了对应同一个资源的锁。
5. 进程 1 从阻塞中恢复过来，释放掉了进程 2 持有的锁。

### 解决方案

这个问题的根源就是错误的释放锁。可以 `set` 的 `value` **设置为一个随机数或者唯一的 uuid，释放锁时先匹配随机数是否一致**，然后再删除 `key`，这是为了**确保当前线程占有的锁不会被其它线程释放，除非这个锁是过期了被服务器自动释放的**。

```py
tag = random.nextint()  # 随机数
if redis.set(key, tag, nx=True, ex=5):
    do_something()
    redis.delifequals(key, tag)  # 假想的 delifequals 指令
```

上面的方案还有一点小问题，就是 `delifequals` 指令，这是一个自定义的指令，匹配 `value` 和删除 `key` 并不是一个原子操作，还是会有原子性问题。例如在匹配 `value` 后，还没来得及删除 `key`，锁就过期了，此时其它线程就可以获取到锁了。然后又执行了删除 `key` 的操作，这样就会把其它线程的锁给释放了。

Redis 也没有提供类似于 `delifequals` 这样的指令，这就需要使用 Lua 脚本来处理了，因为 Lua 脚本可以保证连续多个指令的原子性执行。

```lua
# delifequals
if redis.call("get",KEYS[1]) == ARGV[1] then
    return redis.call("del",KEYS[1])
else
    return 0
end
```

这段 Lua 脚本在执行的时候要把前面的 `tag` 作为 `ARGV[1]` 的值传进去，把 `key` 作为 `KEYS[1]` 的值传进去。

### 锁续命（Watchdog）方案

上面的方案，只是相对安全一点，因为如果真的超时了，当前线程的逻辑没有执行完，其它线程也会乘虚而入。

为了解决这个问题，我们可以在获取锁之后，开启一个守护线程，用来给快要过期的锁“续命”，也就是不断的延长锁的过期时间。现在已经有很成熟的方案，例如 redisson。

redisson 是一个在 Redis 的基础上提供了许多分布式服务。其中就包含了各种分布式锁的实现。

![redisson](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redisson.png)

redisson 自旋尝试加锁的逻辑，如果加锁失败，会拿到当前锁的剩余时间 ttl，然后让出 CPU 让其它线程执行，等待 ttl 时间后再继续尝试加锁。加锁失败的同时还会去订阅一个 Redis channel，监听锁释放的消息，当锁释放后会收到消息，然后重新尝试加锁。

### Go 实现锁续命


核心设计思路

- 后台定时续期：获取锁成功后启动一个 goroutine 定期续期
- 线程(协程)标识验证：续期时验证锁是否仍由当前协程持有
- 自动停止机制：锁释放或协程退出时自动停止续期

```go
package redistlock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	defaultWatchdogInterval = 10 * time.Second // 默认续期间隔
	defaultLockTimeout     = 30 * time.Second // 默认锁超时时间
)

type DistLock struct {
	client         *redis.Client
	key            string
	value          string // 唯一标识，格式: UUID:goroutineID
	watchdogActive bool
	stopWatchdog   chan struct{}
	mutex          sync.Mutex
}

// NewDistLock 创建一个新的分布式锁实例
func NewDistLock(client *redis.Client, key string) *DistLock {
	return &DistLock{
		client:       client,
		key:          key,
		value:        generateLockValue(),
		stopWatchdog: make(chan struct{}),
	}
}

// generateLockValue 生成锁的唯一标识值
func generateLockValue() string {
	// 生成随机 UUID 部分
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	uuid := hex.EncodeToString(buf)
	
	// 获取当前 goroutine ID
	goid := getGoroutineID()
	
	return fmt.Sprintf("%s:%d", uuid, goid)
}

// 获取 goroutine ID (简化实现)
func getGoroutineID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, _ := strconv.ParseUint(idField, 10, 64)
	return id
}

// Lock 获取分布式锁
func (dl *DistLock) Lock(ctx context.Context, timeout time.Duration) error {
	dl.mutex.Lock()
	defer dl.mutex.Unlock()

	// 尝试获取锁
	acquired, err := dl.client.SetNX(ctx, dl.key, dl.value, defaultLockTimeout).Result()
	if err != nil {
		return err
	}

	if acquired {
		// 启动看门狗
		dl.startWatchdog(ctx)
		return nil
	}

	// 等待锁释放或超时
	if timeout > 0 {
		expire := time.Now().Add(timeout)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				acquired, err := dl.client.SetNX(ctx, dl.key, dl.value, defaultLockTimeout).Result()
				if err != nil {
					return err
				}
				if acquired {
					dl.startWatchdog(ctx)
					return nil
				}
				if time.Now().After(expire) {
					return errors.New("lock timeout")
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return errors.New("lock acquisition failed")
}

// startWatchdog 启动看门狗续期机制
func (dl *DistLock) startWatchdog(ctx context.Context) {
	if dl.watchdogActive {
		return
	}

	dl.watchdogActive = true
	go func() {
		ticker := time.NewTicker(defaultWatchdogInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 续期操作
				renewed, err := dl.renewLock(ctx)
				if err != nil || !renewed {
					// 续期失败，可能是锁已释放或已失去所有权
					dl.mutex.Lock()
					dl.watchdogActive = false
					dl.mutex.Unlock()
					return
				}
			case <-dl.stopWatchdog:
				// 收到停止信号
				dl.mutex.Lock()
				dl.watchdogActive = false
				dl.mutex.Unlock()
				return
			case <-ctx.Done():
				// 上下文取消
				dl.mutex.Lock()
				dl.watchdogActive = false
				dl.mutex.Unlock()
				return
			}
		}
	}()
}

// renewLock 续期锁
func (dl *DistLock) renewLock(ctx context.Context) (bool, error) {
	// 使用 Lua 脚本保证原子性
	script := `
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("pexpire", KEYS[1], ARGV[2])
	else
		return 0
	end
	`
	result, err := dl.client.Eval(ctx, script, []string{dl.key}, dl.value, defaultLockTimeout.Milliseconds()).Result()
	if err != nil {
		return false, err
	}

	if val, ok := result.(int64); ok {
		return val == 1, nil
	}
	return false, nil
}

// Unlock 释放锁
func (dl *DistLock) Unlock(ctx context.Context) error {
	dl.mutex.Lock()
	defer dl.mutex.Unlock()

	// 先停止看门狗
	if dl.watchdogActive {
		close(dl.stopWatchdog)
		dl.watchdogActive = false
		dl.stopWatchdog = make(chan struct{})
	}

	// 使用 Lua 脚本保证原子性
	script := `
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("del", KEYS[1])
	else
		return 0
	end
	`
	_, err := dl.client.Eval(ctx, script, []string{dl.key}, dl.value).Result()
	return err
}

// IsLocked 检查锁是否仍被当前实例持有
func (dl *DistLock) IsLocked(ctx context.Context) (bool, error) {
	val, err := dl.client.Get(ctx, dl.key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == dl.value, nil
}
```

## RedLock

Redis 一般都是集群架构，很少有使用单机部署的。但是分布式锁在集群架构中是存在问题的。

比如在 Sentinel 集群中，主节点挂掉时，从节点会取而代之，客户端上却并没有明显感知。原先第一个客户端在主节点中申请成功了一把锁，但是这把锁还没有来得及同步到从节点，主节点突然挂掉了。然后从节点变成了主节点，这个新的节点内部没有这个锁，所以当另一个客户端过来请求加锁时，立即就批准了。这样就会导致系统中同样一把锁被两个客户端同时持有，不安全性由此产生。

为了解决这个问题，Redis 作者 antirez 提出了 RedLock 算法。

为了使用 Redlock，需要提供多个 Redis 实例，这些实例之前**相互独立没有主从关系**。同很多分布式算法一样，redlock 也使用**大多数机制**。

加锁时，它会向过半节点发送 `set(key, value, nx=True, ex=xxx)` 指令，只要过半节点 `set` 成功，那就认为加锁成功。释放锁时，需要向所有节点发送 `del` 指令。不过 Redlock 算法还需要考虑出错重试、时钟漂移等很多细节问题，同时**因为 Redlock 需要向多个节点进行读写，意味着相比单实例 Redis 性能会下降**一些。

![redlock](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redlock.png)

但是 RedLock 并不是一个推荐的方案，因为 RedLock 还存在一些问题：

1. 主从同步：如果主节点还没来得及把锁同步到从节点，主节点就挂掉了，那么这个锁就会丢失。那就又回到了 Redlock 最初要解决的问题上。
2. 当然也可以不部署主从节点，但是如果主节点挂了超过一半的节点，就会导致无法加锁。而且如果持久化机制是设置的每秒执行一次，如果正好在执行持久化时，主节点挂掉了，那么这个锁就会丢失。
3. 如果主节点太多，那么加锁和释放锁的时间就会比较长。


如果**非要这种高一致性的锁，那么可以使用 Zookeeper 来实现**。

## 可重入锁

如果一个锁支持**同一个线程的多次加锁**，那么这个锁就是可重入的。

## Redis Lua 脚本

Redis在 2.6 推出了脚本功能，允许开发者使用Lua  语言编写脚本传到 Redis 中执行。使用脚本的好处如下:

1. **减少网络开销**：本来 5 次网络请求的操作，可以用一个请求完成，原先 5 次请求的逻辑放在 Redis 服务器上完成。使用脚本，减少了网络往返时延。**这点跟管道类似**。
2. **原子操作**：Redis 会将整个脚本作为一个整体执行，中间不会被其他命令插入。**管道不是原子的，不过 Redis 的批量操作命令(类似 mset )是原子的**。
3. **替代 Redis 的事务功能**：Redis 自带的事务功能很鸡肋，而 Redis 的 lua 脚本几乎实现了常规的事务功能，官方推荐用 Redis lua 替代 Redis 的事务功能。

Redis 2.6 版本开始，通过内置的 Lua 解释器，可以使用 `EVAL` 命令对 Lua 脚本进行求值。

```
EVAL script numkeys key [key ...] arg [arg ...]
```

- `script` 参数是一段 Lua 脚本程序。Redis 使用 `EVAL` 命令的第一个参数来传递脚本程序。这段**脚本不必(也不应该)定义为一个 Lua 函数**。
- `numkeys` 参数用于指定键名参数的个数。
- 键名参数 `key [key ...]` 从 `EVAL` 的第三个参数开始算起，表示在脚本中所用到的那些 Redis 键 (key)，这些键名参数可以在 Lua 中通过全局变量 `KEYS` 数组，用 1 为基址的形式访问( `KEYS[1]` ， `KEYS[2]` ，以此类推)。

```bash
127.0.0.1:6379> eval "return {KEYS[1],KEYS[2],ARGV[1],ARGV[2]}" 2 key1 key2 first second
1) "key1"
2) "key2"
3) "first"
4) "second"
```

在 Lua 脚本中，可以使用 `redis.call()` 函数来执行 Redis 命令：

```lua
jedis.set("product_stock_10016", "15");  // 初始化商品10016的库存
String script = " local count = redis.call('get', KEYS[1]) " +
                " local a = tonumber(count) " +
                " local b = tonumber(ARGV[1]) " +
                " if a >= b then " +
                "   redis.call('set', KEYS[1], a-b) " +
                "   return 1 " +
                " end " +
                " return 0 ";
Object obj = jedis.eval(script, Arrays.asList("product_stock_10016"), Arrays.asList("10"));
System.out.println(obj);
```

{{< callout type="info" >}}
**不要在 Lua 脚本中出现死循环和耗时的运算，否则 Redis 会阻塞**，将不接受其他的命令，所以使用时要注意不能出现死循环、耗时的运算。Redis 是单进程、单线程执行脚本。管道不会阻塞 Redis。
{{< /callout >}}


## 优化

1. 分布式锁的粒度要尽量小，不需要被锁住的代码尽量并发执行。
2. 分段锁：比如一个商品（`product:10011:stock`）有 1000 个库存，那么可以把库存分成 10 段（`product:10011:stock1`、`product:10011:stock2` 等等），每一段都有一个锁，这样就可以有 10 个线程并发执行了。