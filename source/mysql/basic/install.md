---
title: MySQL 安装
---
# MySQL 安装
`MySQL` 的大部分安装包都包含了服务器程序和客户端程序

> 注意，**在 Linux 下使用 RPM 安装，有时需要分别安装服务器 RPM 包和客户端 RPM 包**。

## MySQL 的安装目录

Linux 下的安装目录一般为 `/usr/local/mysql/`。

### bin目录
打开 `/usr/local/mysql/bin`（你的 mysql 安装目录），执行 `tree`：
```
.
├── mysql
├── mysql.server -> ../support-files/mysql.server
├── mysqladmin
├── mysqlbinlog
├── mysqlcheck
├── mysqld           # mysql 的服务端程序
├── mysqld_multi     # 运行多个 MySQL 服务器进程
├── mysqld_safe
├── mysqldump
├── mysqlimport
├── mysqlpump
... (省略其他文件)
0 directories, 40 files
```

这个目录一般需要配置到环境变量的 `PATH` 中，Linux 中各个路径以 `:` 分隔。也可以选择不配：
```bash
/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/usr/local/mysql/bin
```

#### mysqld_safe
`mysqld_safe` 是一个启动脚本，最终也是调用 `mysqld`，但是还会另外一个监控进程，在服务器进程挂了的时候，可以帮助重启服务器进程。
`mysqld_safe` 启动服务器程序时，会将服务器程序的出错信息和其他诊断信息重定向到某个文件中，产生出错日志，这样可以方便我们找
出发生错误的原因。

#### mysql.server
`mysql.server` 也是一个启动脚本，会间接的调用 `mysqld_safe`，使用 `mysql.server` 启动服务器程序：
```bash
mysql.server start
```
`mysql.server` 文件其实是一个链接文件，它的实际文件是 `../support-files/mysql.server`。

#### client
`bin` 目录下的 `mysql`、`mysqladmin`、`mysqldump`、`mysqlcheck` 都是客户端程序，启动客户端程序连接服务器进程：
```bash
mysql --host=<主机名>  --user=<用户名> --password=<密码>
```

注意：
- 最好不要在一行命令中输入密码。
- 如果使用的是类 UNIX 系统，并且省略 `-u` 参数后，会把你登陆操作系统的用户名当作 MySQL 的用户名去处理。