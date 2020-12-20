# Note

## 覆盖索引

为了彻底告别回表操作带来的性能损耗：**最好在查询列表里只包含索引列**

```sql
SELECT name, birthday, phone_number FROM person_info WHERE name > 'Asa' AND name < 'Barlow'
```

只查询name, birthday, phone_number这三个索引列的值，所以在通过idx_name_birthday_phone_number索引得到结果后就不必到聚簇索引中再查找记录的剩余列，也就是country列的值了，这样就省去了回表操作带来的性能损耗。

**不鼓励用 `*` 号作为查询列表，最好把我们需要查询的列依次标明**。

## join buffer

`join buffer` 就是执行连接查询前申请的一块固定大小的内存，先把**若干条驱动表结果集中的记录装在这个 `join buffer` 中**，然后开始扫描被驱动表，**每一条被驱动表的记录一次性和 join buffer 中的多条驱动表记录做匹配**，因为匹配的过程都是在内存中完成的，所以这样可以显著减少被驱动表的 I/O 代价。

最好的情况是 join buffer 足够大，能容纳驱动表结果集中的所有记录，这样只需要访问一次被驱动表就可以完成连接操作了。这种加入了 join buffer 的嵌套循环连接算法称之为**基于块的嵌套连接**（Block Nested-Loop Join）算法。

这个join buffer的大小是可以通过启动参数或者系统变量join_buffer_size进行配置，默认大小为262144字节（也就是256KB），最小可以设置为128字节。当然，对于优化被驱动表的查询来说，最好是为被驱动表加上效率高的索引，如果实在不能使用索引，并且自己的机器的内存也比较大可以尝试调大join_buffer_size的值来对连接查询进行优化。

另外需要注意的是，驱动表的记录并不是所有列都会被放到join buffer中，只有查询列表中的列和过滤条件中的列才会被放到join buffer中，所以再次提醒我们，**最好不要把 `*` 作为查询列表**，只需要把我们关心的列放到查询列表就好了，这样还可以在 `join buffer` 中放置更多的记录呢哈。
