---
title: 跳表
weight: 7
---

`zset` 的内部实现是一个 `hashtable` 加一个跳跃列表 (`skiplist`)。

```c
struct zset {
    dict *dict; // all values  value=>score
    zskiplist *zsl;
}
```

- `dict` 是一个 `hashtable` 结构用来存储 `value` 和 `score` 值的映射关系。用来查找数据到分数的对应关系。
- `zsl` 是一个 `skiplist` 结构，用来根据分数查询数据，支持范围查询。

## 跳表的原理

假设使用链表来存储数据，查找一个元素的时间复杂度是 `O(n)`，因为要遍历整个链表。

![redis-skiplist1](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redis-skiplist1.png)

可以加一个索引层，索引层的元素指向原始链表中的元素，这样查找元素 79，就可以从索引层开始找。只要找到第一个比 79 大的元素，图中就是 84，然后就通过前一个元素 78 指向的指针，找到原始链表中的 79。

![redis-skiplist2](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redis-skiplist2.png)

上面的索引层是一层，相对于原始链表来说，查找的元素个数大概减少了一半。

如果再加一层索引层，那么查找的元素个数就会减少到原来的 1/4。

![redis-skiplist3](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redis-skiplist3.png)

从第一层索引层开始查找，最后一个元素是 78，比 79 小，那就往下一层索引层找。

如果再加一层索引层，就可以明显的看出来，一次减少一半，类似二分查找：

![skiplist-demo](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/skiplist-demo.png)

**时间复杂度**：

如果有 N 个元素：

- 第一层索引层的元素个数是 `N/2`。
- 第二层索引层的元素个数是 `N/4`。
- 第三层索引层的元素个数是 `N/8`。
- 第 k 层索引层的元素个数是 `N/(2^k)`。

假设 k 层索引层的元素个数是 2，那么 `N/(2^k)=2 => 2^k=N/2 => k=log2(N/2)`，索引的访问，每层大概访问两个元素，是常数级的。忽略掉常数，即 `k=logN`。

## zskiplist

![redis-zskiplist](https://raw.gitcode.com/shipengqi/illustrations/files/main/db/redis-zskiplist.png)

上图中只有三个元素，`a`、`b`、`c`，对应的 score 值分别是 100、120、200。

- `header` 指向跳表的**头节点**。头节点不存储任何数据，只用来作为跳表的**起始点**。
- `tail` 指向跳表的**尾节点**，尾节点不存储任何数据，只用来作为跳表的**结束点**。
- `level` 表示当前跳表的**最高层数**，初始化为 1。这是用来遍历跳表时，直接找到最高的一层开始遍历。图中虽然画的是 31 层，但是实际使用时，只使用了 3 层。
- `length` 表示跳表中元素的个数，初始化为 0。

#### zskiplistNode

跳表的节点结构：

```c
struct zslnode {
  string value;
  double score;
  zslnode*[] forwards;  // 多层连接指针
  zslnode* backward;
}
```

- `backward` 指针用来从后往前遍历跳表。

#### 随机层数

不停地往跳表中插入数据时，如果不更新索引，就有可能出现某 2 个索引结点之间数据非常多的情况。极端情况下，跳表还会退化成单链表。

所以对于每一个新插入的节点，都需要调用一个随机算法给它分配一个合理的层数。生成高层的索引节点的概率要小于生成低层的索引节点的概率。

```c
/* Returns a random level for the new skiplist node we are going to create.
 * The return value of this function is between 1 and ZSKIPLIST_MAXLEVEL
 * (both inclusive), with a powerlaw-alike distribution where higher
 * levels are less likely to be returned. */
int zslRandomLevel(void) {
    int level = 1;
    while ((random()&0xFFFF) < (ZSKIPLIST_P * 0xFFFF))
        level += 1;
    return (level<ZSKIPLIST_MAXLEVEL) ? level : ZSKIPLIST_MAXLEVEL;
}
```

## Go 实现一个简单的跳表

```go
const (
	MaxLevel    = 16   // 最大层级
	Probability = 0.5  // 每一层晋升的概率
)

type Skiplist struct {
    head *Node
    level int   // 当前跳表的最大层级
}

type Node struct {
    // nexts 指针是一个 slice 的形式，其长度对应为当前节点的高度
    nexts    []*Node
    key, val int
}

// NewNode 创建一个新节点
func NewNode(key, value int, level int) *node {
	return &node{
		key:   key,
		value: value,
		next:  make([]*node, level),
	}
}

// NewSkipList 初始化跳表
func NewSkipList() *SkipList {
	rand.Seed(time.Now().UnixNano())
	head := NewNode(math.MinInt32, nil, MaxLevel) // 头节点使用最小值
	return &SkipList{
		head:  head,
		level: 1,
	}
}

// randomLevel 决定新节点的层级（几率下降）
func randomLevel() int {
	level := 1
	for rand.Float64() < Probability && level < MaxLevel {
		level++
	}
	return level
}

// 根据 key 读取 val，第二个 bool flag 反映 key 在 skiplist 中是否存在
func (s *Skiplist) Get(key int) (int, bool) {
    // 根据 key 尝试检索对应的 node，如果 node 存在，则返回对应的 val
    if _node := s.search(key); _node != nil {
        return _node.val, true
    }
    return -1, false
}

// 从跳表中检索 key 对应的 Node
func (s *Skiplist) search(key int) *Node {
    // 每次检索从头部出发
    move := s.head
    // 每次检索从最大高度出发，直到来到首层
    for level := s.level - 1; level >= 0; level-- {
        // 在每一层中持续向右遍历，直到下一个节点不存在或者 key 值大于等于 key
        for move.nexts[level] != nil && move.nexts[level].key < key {
            move = move.nexts[level]
        }
        // 如果 key 值相等，则找到了目标直接返回
        if move.nexts[level] != nil && move.nexts[level].key == key {
            return move.nexts[level]
        }
        
        // 当前层没找到目标，则层数减 1，继续向下
    }
    
    // 遍历完所有层数，都没有找到目标，返回 nil
    return nil
}

// 将 key-val 对加入 skiplist
func (s *Skiplist) Put(key, val int) {
    // 假如 kv对已存在，则直接对值进行更新并返回
    if _node := s.search(key); _node != nil {
        _node.val = val
        return
    }


    // 新节点的高度
    level := randomLevel()

    // 创建出新的节点
    newNode := NewNode(key,val,level)
 
    // 从头节点的最高层出发
    move := s.head
    for level := s.level - 1; level >= 0; level-- {
        // 向右遍历，直到右侧节点不存在或者 key 值大于 key
        for move.nexts[level] != nil && move.nexts[level].key < key {
            move = move.nexts[level]
        }

        // 调整指针关系，完成新节点的插入
        newNode.nexts[level] = move.nexts[level]
        move.nexts[level] = &newNode
    }
}

// 根据 key 从跳表中删除对应的节点
func (s *Skiplist) Del(key int) {
    // 如果 kv 对不存在，则无需删除直接返回
    if _node := s.search(key); _node == nil {
        return
    }

    // 从头节点的最高层出发    
    move := s.head
    for level := s.level - 1; level >= 0; level-- {
        // 向右遍历，直到右侧节点不存在或者 key 值大于等于 key
        for move.nexts[level] != nil && move.nexts[level].key < key {
            move = move.nexts[level]
        }
        
        // 右侧节点不存在或者 key 值大于 target，则直接跳过
        if move.nexts[level] == nil || move.nexts[level].key > key{
           continue
        }
        
        // 走到此处意味着右侧节点的 key 值必然等于 key，则调整指针引用
        move.nexts[level] = move.nexts[level].nexts[level]
    }

    // 更新跳表层级
	for s.level > 1 && s.head.nexts[s.level-1] == nil {
		s.level--
	}
}
```