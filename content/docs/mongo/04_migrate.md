---
title: 数据迁移
---

## 简单迁移

如何将一台 linux A 上的 mongodb 数据迁移到另外一台 linux B 上？

两个命令即可完成任务：

- 数据的导出：`mongoexport`
- 数据的导入：`mongoimport`

具体步骤：

1. 找到 A 的 mongodb 的`mongoexport`所在目录，一般在`/usr/bin`或者`/usr/local/mongodb/bin`下。

```bash
cd /usr/local/mongodb/bin
```

2. 将数据导出，执行命令：`./mongoexport -d dbname -c collectionname -o xxx.dat`
`dbname`为数据库名称，`collectionname`为集合名称，`xxx.dat`为导出后的数据的名称。导出后的`xxx.dat`在`mongoexport`所在的目录下。

```bash
./mongoexport -d moregold -c logs -o logdata.dat
```

将数据库`moregold`下的集合`logs`导出到`mongoexport`所在的目录下，并将其命名为`logdata.dat`

4. 将导出的集合数据移动到 B 服务器上。
5. 找到 B 的`mongoimport`所在的目录：`cd /db/mongo/bin`
6. 将数据导入，执行命令`./mongoimport -h 127.0.0.1:port -u xxx -p xxx -d dbname -c collectionname /path/xxx.dat`：

```bash
./mongoimport -h 127.0.0.1:27017 -u zhangsan -p zhangsan -d moregold -c /root/logdata.dat
```

- `-h 127.0.0.1:27017`：连接到本地，端口号为`27017`
- `-u zhangsan`：用户名为`zhangsan`
- `-p zhangsan`：密码为`zhangsan`

迁移完毕。

## 停机迁移

对于生产环境下，上面的方式可能并不适用。

采用停机迁移的好处是流程操作简单，工具成本低；然而缺点也很明显，
迁移过程中业务是无法访问的，因此只适合于规格小、允许停服的场景。

1. 准备迁移工具
2. 原系统下线
3. 全量迁移
4. 新系统上线

## 业务双写

业务双写是指对现有系统先进行改造升级，支持同时对新库和旧库进行写入。
之后再通过数据迁移工具对旧数据做全量迁移，待所有数据迁移转换完成后切换到新系统。

业务双写的方案是平滑的，对线上业务影响极小；在出现问题的情况下可重新来过，操作压力也会比较小。
但实现该方案比较复杂，需要对现有的代码进行改造并完成新数据的转换及写入，对于开发人员的要求较高。
在业务逻辑清晰、团队对系统有足够的把控能力的场景下适用。

## 增量迁移

增量迁移的基本思路是先进行全量的迁移转换，待完成后持续进行增量数据的处理，直到数据追平后切换系统。

关键点：

- 要求系统支持增量数据的记录。
对于MongoDB可以利用oplog实现这点，为避免全量迁移过程中oplog被冲掉，
在开始迁移前就必须开始监听oplog，并将变更全部记录下来。
如果没有办法，需要从应用层上考虑，比如为所有的表(集合)记录下updateTime这样的时间戳，
或者升级应用并支持将修改操作单独记录下来。

- 增量数据的回放是持续的。
在所有的增量数据回放转换过程中，系统仍然会产生新的增量数据，这要求迁移工具
能做到将增量数据持续回放并将之追平，之后才能做系统切换。

> MongoDB 3.6版本开始便提供了Change Stream功能，支持对数据变更记录做监听。
这为实现数据同步及转换处理提供了更大的便利。
