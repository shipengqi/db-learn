---
title: 配置选项
weight: 8
---

MySQL Server 的配置一般都有默认值，如默认存储引擎是 `InnoDB`，这些配置选项叫做**启动选项**，可以在命令行中指定启动参数，也可以通过配置文件指定。

## 在命令行上使用选项

示例：

```bash
mysqld --default-storage-engine=MyISAM
```

- `--skip-networking` （或 `skip_networking`）选项禁止各客户端使用 TCP/IP 网络进行通信。（多个单词之间可以用 `-` 连接，
也可以用 `_` 连接）
- `--default-storage-engine=<engine>` 改变默认存储引擎。

查看更多启动选项：

```bash
mysqld --verbose --help

mysqld_safe --help
```

## 配置文件中使用选项

### 配置文件的路径

#### Windows 操作系统的配置文件

MySQL 会按照下列路径来寻找配置文件：

| 路径 | 描述 |
| --- | --- |
| `%WINDIR%\my.ini`， `%WINDIR%\my.cnf` | |
| `C:\my.ini`， `C:\my.cnf` | |
| `BASEDIR\my.ini`， `BASEDIR\my.cnf` | |
| `defaults-extra-file` | 命令行指定的额外配置文件路径 |
| `%APPDATA%\MySQL\.mylogin.cnf` | 登录路径选项（仅限客户端） |

- `%WINDIR%` 一般是 `C:\WINDOWS`，可以用 `echo %WINDIR%` 来查看。
- `BASEDIR` 指的是 MySQL 安装目录的路径
- `defaults-extra-file` 在命令行上可以这么写 `mysqld --defaults-extra-file=C:\Users\xiaohaizi\my_extra_file.txt`
- `%APPDATA%` 表示 Windows 应用程序数据目录的值
- `.mylogin.cnf` 中只能包含一些用于启动客户端连接服务端的一些选项，包括 `host`、`user`、`password`、`port` 和 `socket`。

#### UNIX 操作系统的配置文件

MySQL 会按照下列路径来寻找配置文件：

| 路径 | 描述 |
| --- | --- |
| `/etc/my.cnf` | |
| `/etc/mysql/my.cnf` | |
| `SYSCONFDIR/my.cnf` | |
| `$MYSQL_HOME/my.cnf` |  特定于服务器的选项（仅限服务器） |
| `defaults-extra-file` | 命令行指定的额外配置文件路径 |
| `~/.my.cnf` | 用户特定选项 |
| `~/.mylogin.cnf` | 登录路径选项（仅限客户端） |

- `SYSCONFDIR` 表示在使用 CMake 构建 MySQL 时使用 `SYSCONFDIR` 选项指定的目录。默认情况下，这是位于编译安装目录下的 `etc` 目录。
- `MYSQL_HOME` 变量的值是我们自己设置的，也可以不设置。

### 配置文件的内容

配置文件中的启动选项分为多个组，每个组有一个组名，用中括号 `[]` 扩起来：

```cnf
[server]
(具体的启动选项...)

[mysqld]
(具体的启动选项...)

[mysqld_safe]
(具体的启动选项...)

[client]
(具体的启动选项...)

[mysql]
(具体的启动选项...)

[mysqladmin]
(具体的启动选项...)
```

每个组下边可以定义各自的启动选项：

```cnf
[server]
option1               # option1，该选项不需要选项值
option2 = value2      # option2，该选项需要选项值
...
```

**配置文件中只能使用长形式的选项**。上面的文件转成命令行格式就是 `--option1 --option2=value2`。

不同的选项组是给不同的启动命令使用的，如 `[mysqld]` 和 `[mysql]` 组分别应用于 `mysqld` 服务端程序和 `mysql` 客户端程序。注意：

- `[server]` 组下边的启动选项将作用于所有的服务器程序。
- `[client]` 组下边的启动选项将作用于所有的客户端程序。
- `mysqld_safe` 和 `mysql.server` 这两个程序在启动时都会读取 `[mysqld]` 选项组中的选项。

#### 特定版本的选项组

可以在选项组的名称后加上特定的 MySQL 版本号，比如对于 `[mysqld]` 选项组来说，我们可以定义一个 `[mysqld-5.7]` 的选项组，它的含义
和 `[mysqld]` 一样，只不过只有版本号为 `5.7` 的 `mysqld` 程序才能使用这个选项组中的选项。

#### 配置文件的优先级

**如果在多个配置文件中设置了相同的启动选项，那以最后一个配置文件中的为准**。文件的顺序按照上面的表格从上到下的顺序。

#### 同一个配置文件中多个组的优先级

比如 `mysqld` 可以访问 `[mysqld]`、`[server]` 组，如果在同一个配置文件中，比如 `~/.my.cnf`，在这些组里出现了同样的配置项，比如这样：

```cnf
[server]
default-storage-engine=InnoDB

[mysqld]
default-storage-engine=MyISAM
```

**以最后一组中的启动选项为准**，比如上面的文件 `default-storage-engine` 就以 `[mysqld]` 组中的配置项为准。

#### defaults-file

如果不想让 MySQL 到默认的路径下搜索配置文件（就是上表中列出的那些），可以在命令行指定 `defaults-file` 选项 `mysqld --defaults-file=/tmp/myconfig.txt`。这样，在程序启动的时候将只在 `/tmp/myconfig.txt` 路径下搜索配置文件。如果文件不存在或无法访问，则会发生错误。

`defaults-file` 和 `defaults-extra-file` 的区别，使用 `defaults-extra-file` 可以指定额外的配置文件搜索路径。

#### 命令行和配置文件的优先级

**如果同一个启动选项既出现在命令行中，又出现在配置文件中，那么以命令行中的启动选项为准**。

## 系统变量

`SHOW VARIABLES [LIKE 匹配的模式];` 查看 MySQL 服务器程序支持的系统变量以及它们的当前值。

**大部分系统变量的值可以在服务端程序运行过程中进行动态修改而无需停止并重启服务器**。

### 设置系统变量

#### 通过启动选项设置

通过命令行添加启动选项和配置文件。例如：

```sh
mysqld --default-storage-engine=MyISAM --max-connections=10
```

或者

```cnf
[server]
default-storage-engine=MyISAM
max-connections=10
```

`max-connections` 表示允许同时连入的客户端数量。

#### 服务器程序运行过程中设置

##### 设置不同作用范围的系统变量

了针对不同的客户端设置不同的系统变量：

- `GLOBAL`：全局变量，影响服务器的整体操作。
- `SESSION`：会话变量，影响某个客户端连接的操作。（`SESSION` 有个别名叫 `LOCAL`）

**通过启动选项设置的系统变量的作用范围都是 `GLOBAL` 的**，也就是对所有客户端都有效的，因为在系统启动的时候还没有客户端程序连接进来。

通过客户端程序设置系统变量的语法：

```bash
SET [GLOBAL|SESSION] 系统变量名 = 值;

# 或者
SET [@@(GLOBAL|SESSION).]var_name = XXX;
```

比如在服务端进程运行过程中把作用范围为 `GLOBAL` 的系统变量 `default_storage_engine` 的值修改为 `MyISAM`，也就是想让之后新连接到
服务器的客户端都用 `MyISAM` 作为默认的存储引擎：

```bash
语句一：SET GLOBAL default_storage_engine = MyISAM;
语句二：SET @@GLOBAL.default_storage_engine = MyISAM;
```

只对本客户端生效：

```bash
语句一：SET SESSION default_storage_engine = MyISAM;
语句二：SET @@SESSION.default_storage_engine = MyISAM;
语句三：SET default_storage_engine = MyISAM;
```

**`SESSION` 是默认的作用范围**。

> 如果某个客户端改变了某个系统变量在 `GLOBAL` 作用范围的值，并不会影响该系统变量在当前已经连接的客户端作用范围为 `SESSION` 的值，只会影响
后续连入的客户端在作用范围为 `SESSION` 的值。

注意：

- 有一些系统变量只具有 `GLOBAL` 作用范围，如 `max_connections`。
- 有一些系统变量只具有 `SESSION` 作用范围，如 `insert_id`，表示在对某个包含 `AUTO_INCREMENT` 列的表进行插入时，该列初始的值。
- 有些系统变量是只读的，并不能设置值。如 version，表示当前 MySQL 的版本。

## 状态变量

MySQL 服务器程序中维护了很多关于程序运行状态的变量，它们被称为**状态变量**。**它们的值只能由服务器程序自己来设置**。状态变量也有 `GLOBAL`
和 `SESSION` 两个作用范围的，所以查看状态变量的语句可以这么写：

```sql
SHOW [GLOBAL|SESSION] STATUS [LIKE 匹配的模式];
```
