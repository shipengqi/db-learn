# Mysql

MySQL服务器的进程也被称为**MySQL数据库实例**，简称**数据库实例**。MySQL 服务器进程的默认名称为`mysqld`， 而 MySQL 客户端进程的默认名称为`mysql`。

## 安装

> `MySQL`的大部分安装包都包含了服务器程序和客户端程序，不过**在Linux下使用RPM包时会有单独的服务器RPM包和客户端RPM包，需要分别安装**。

**一定要记住MySQL的安装目录**。

一般Linux 下的安装目录为`/usr/local/mysql/`。

### bin目录
打开`/usr/local/mysql/bin`（注意这里要改成你自己的安装目录），执行`tree`：
```
.
├── mysql
├── mysql.server -> ../support-files/mysql.server
├── mysqladmin
├── mysqlbinlog
├── mysqlcheck
├── mysqld
├── mysqld_multi
├── mysqld_safe
├── mysqldump
├── mysqlimport
├── mysqlpump
... (省略其他文件)
0 directories, 40 files
```