---
title: 数据库操作
weight: 2
---

## 查看数据库

```bash
mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
4 rows in set (0.01 sec)

mysql>
```

## 切换数据库

`USE 数据库名称;`：

```bash
mysql> USE test;
Database changed

mysql>
```

## 创建

`CREATE DATABASE 数据库名;`：

```bash
mysql> CREATE DATABASE test;
Query OK, 1 row affected (0.00 sec)

mysql>
```

### IF NOT EXISTS

使用 `CREATE DATABASE` 去创建一个已经存在数据库会产生错误，可以使用 `IF NOT EXISTS` 来避免错误。

```bash
mysql> CREATE DATABASE IF NOT EXISTS test;
Query OK, 0 rows affected (0.00 sec)

mysql>
```

这个命令的意思是如果指定的数据库不存在的话就创建它，否则就什么也不做。

## 删除

`DROP DATABASE 数据库名;`：

```bash
mysql> DROP DATABASE test;
Query OK, 0 rows affected (0.00 sec)

mysql>
```

### IF EXISTS

使用 `DROP DATABASE` 去删除一个不存在的数据库会产生错误，可以使用 `IF EXISTS` 来避免错误。

```bash
mysql> DROP DATABASE IF EXISTS test;
Query OK, 0 rows affected (0.00 sec)

mysql>
```

这个命令的意思是如果指定的数据库存在的话就删除它，否则就什么也不做。