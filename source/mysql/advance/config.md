---
title: 启动选项和配置文件
---
# 启动选项和配置文件
MySQL服务器程序设置项一般都有各自的默认值，比方说服务器允许同时连入的客户端的默认数量是`151`，表的默认存储引擎是`InnoDB`，我们可以在程序启动的时候去修改这些默认值，
对于这种在程序启动时指定的设置项也称之为**启动选项**。

不论是服务器相关的程序（比如`mysqld`、`mysqld_safe`）还是客户端相关的程序（比如`mysql`、`mysqladmin`），在启动的时候基本都可以指定启动参数。
这些启动参数可以放在命令行中指定，也可以把它们放在配置文件中指定。

## 在命令行上使用选项
- `--skip-networking`选项禁止各客户端使用TCP/IP网络进行通信
- `--default-storage-engine=<引擎>`改变默认存储引擎

## 配置文件中使用选项
一般推荐在配置文件中使用选项，不用每次在命令行中添加参数。

### 配置文件的路径
#### Windows操作系统的配置文件
MySQL会按照下列路径来寻找配置文件：
| 路径 | 描述 |
| --- | --- |
| `%WINDIR%\my.ini`， `%WINDIR%\my.cnf` | |
| `C:\my.ini`， `C:\my.cnf` | |
| `BASEDIR\my.ini`， `BASEDIR\my.cnf` | |
| `defaults-extra-file` | 命令行指定的额外配置文件路径 |
| `%APPDATA%\MySQL\.mylogin.cnf` | 登录路径选项（仅限客户端） |

- 前三个路径中，配置文件可以使用`.ini`的扩展名，也可以使用`.cnf`的扩展名。
- `%WINDIR%`指的是你机器上Windows目录的位置，通常是`C:\WINDOWS`，可以使用这个`echo %WINDIR%`来查看。
- `BASEDIR`指的是MySQL安装目录的路径
- `defaults-extra-file`在命令行上可以这么写`mysqld --defaults-extra-file=C:\Users\xiaohaizi\my_extra_file.txt`
- `%APPDATA%`表示Windows应用程序数据目录的值
- `.mylogin.cnf`中只能包含一些用于启动客户端软件时连接服务器的一些选项，包括`host`、`user`、`password`、`port`和`socket`。

#### 类UNIX操作系统的配置文件
MySQL会按照下列路径来寻找配置文件：
| 路径 | 描述 |
| --- | --- |
| `/etc/my.cnf` | |
| `/etc/mysql/my.cnf` | |
| `SYSCONFDIR/my.cnf` | |
| `$MYSQL_HOME/my.cnf` |  特定于服务器的选项（仅限服务器） |
| `defaults-extra-file` | 命令行指定的额外配置文件路径 |
| `~/.my.cnf` | 用户特定选项 |
| `~/.mylogin.cnf` | 登录路径选项（仅限客户端） |

- `SYSCONFDIR`表示在使用CMake构建MySQL时使用`SYSCONFDIR`选项指定的目录。默认情况下，这是位于编译安装目录下的`etc`目录。
- `MYSQL_HOME`变量的值是我们自己设置的，也可以不设置。

### 配置文件的内容
配置文件中的启动选项被划分为若干个组，每个组有一个组名，用中括号`[]`扩起来：
```
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
每个组下边可以定义若干个启动选项，我们以`[server]`组为例来看一下填写启动选项的形式（其他组中启动选项的形式是一样的）：
```
[server]
option1     #这是option1，该选项不需要选项值
option2 = value2      #这是option2，该选项需要选项值
...
```

**配置文件中只能使用长形式的选项**。上面的文件转成命令行格式就是`--option1 --option2=value2`。

配置文件中不同的选项组是给不同的启动命令使用的，例如， `[mysqld]`和`[mysql]`组分别应用于`mysqld`服务器程序和`mysql`客户端程序。但是注意：
- `[server]`组下边的启动选项将作用于所有的服务器程序。
- `[client]`组下边的启动选项将作用于所有的客户端程序。
- `mysqld_safe`和`mysql.server`这两个程序在启动时都会读取`[mysqld]`选项组中的选项。

#### 特定MySQL版本的专用选项组
可以在选项组的名称后加上特定的MySQL版本号，比如对于`[mysqld]`选项组来说，我们可以定义一个`[mysqld-5.7]`的选项组，它的含义和`[mysqld]`一样，
只不过只有版本号为`5.7`的`mysqld`程序才能使用这个选项组中的选项。

#### 配置文件的优先级
**如果在多个配置文件中设置了相同的启动选项，那以最后一个配置文件中的为准**。文件的顺序按照上面的表格从上到下的顺序。

#### 同一个配置文件中多个组的优先级
比如`mysqld`可以访问`[mysqld]`、`[server]`组，如果在同一个配置文件中，比如`~/.my.cnf`，在这些组里出现了同样的配置项，比如这样：
```
[server]
default-storage-engine=InnoDB

[mysqld]
default-storage-engine=MyISAM
```
**将以最后一个出现的组中的启动选项为准**，比如上面的文件`default-storage-engine`就以`[mysqld]`组中的配置项为准。

#### defaults-file
如果不想让MySQL到默认的路径下搜索配置文件（就是上表中列出的那些），可以在命令行指定`defaults-file`选项`mysqld --defaults-file=/tmp/myconfig.txt`。
这样，在程序启动的时候将只在/tmp/myconfig.txt路径下搜索配置文件。如果文件不存在或无法访问，则会发生错误。

`defaults-file`和`defaults-extra-file`的区别，使用`defaults-extra-file`可以指定额外的配置文件搜索路径（也就是说那些固定的配置文件路径也会被搜索）。

#### 命令行和配置文件的优先级
**如果同一个启动选项既出现在命令行中，又出现在配置文件中，那么以命令行中的启动选项为准**。

## 系统变量
`SHOW VARIABLES [LIKE 匹配的模式];`查看MySQL服务器程序支持的系统变量以及它们的当前值。
**对于大部分系统变量来说，它们的值可以在服务器程序运行过程中进行动态修改而无需停止并重启服务器**。
### 设置系统变量
#### 通过启动选项设置
通过命令行添加启动选项和配置文件。例如：
```sh
mysqld --default-storage-engine=MyISAM --max-connections=10
```
或者
```
[server]
default-storage-engine=MyISAM
max-connections=10
```

`max-connections`表示允许同时连入的客户端数量。

#### 服务器程序运行过程中设置
##### 设置不同作用范围的系统变量
为了针对不同的客户端设置不同的系统变量，设计MySQL的大叔提出了系统变量的作用范围的概念，具体来说作用范围分为这两种：
- `GLOBAL`：全局变量，影响服务器的整体操作。
- `SESSION`：会话变量，影响某个客户端连接的操作。（注：`SESSION`有个别名叫`LOCAL`）

很显然，**通过启动选项设置的系统变量的作用范围都是`GLOBAL`的**，也就是对所有客户端都有效的，因为在系统启动的时候还没有客户端程序连接进来。

通过客户端程序设置系统变量的语法：
```
SET [GLOBAL|SESSION] 系统变量名 = 值;

# 或者
SET [@@(GLOBAL|SESSION).]var_name = XXX;
```

比如我们想在服务器运行过程中把作用范围为`GLOBAL`的系统变量`default_storage_engine`的值修改为`MyISAM`，也就是想让之后新连接到
服务器的客户端都用`MyISAM`作为默认的存储引擎，那我们可以选择下边两条语句中的任意一条来进行设置：
```
语句一：SET GLOBAL default_storage_engine = MyISAM;
语句二：SET @@GLOBAL.default_storage_engine = MyISAM;
```
如果只想对本客户端生效，也可以选择下边三条语句中的任意一条来进行设置：
```
语句一：SET SESSION default_storage_engine = MyISAM;
语句二：SET @@SESSION.default_storage_engine = MyISAM;
语句三：SET default_storage_engine = MyISAM;
```

在设置系统变量的语句中省略了作用范围，**默认的作用范围就是`SESSION`**。

> 如果某个客户端改变了某个系统变量在`GLOBAL`作用范围的值，并不会影响该系统变量在当前已经连接的客户端作用范围为`SESSION`的值，只会影响后续连入的客户端在作用范围为`SESSION`的值。

注意：
- 有一些系统变量只具有`GLOBAL`作用范围，比方说`max_connections`，表示服务器程序支持同时最多有多少个客户端程序进行连接。
- 有一些系统变量只具有`SESSION`作用范围，比如`insert_id`，表示在对某个包含`AUTO_INCREMENT`列的表进行插入时，该列初始的值。
- 有些系统变量是只读的，并不能设置值。比方说version，表示当前MySQL的版本。

#### 启动选项和系统变量的区别
启动选项是在程序启动时我们程序员传递的一些参数，而系统变量是影响服务器程序运行行为的变量。

## 状态变量

MySQL服务器程序中维护了好多关于程序运行状态的变量，它们被称为**状态变量**。**它们的值只能由服务器程序自己来设置，我们是不能设置的**。
状态变量也有`GLOBAL`和`SESSION`两个作用范围的，所以查看状态变量的语句可以这么写：
```sql
SHOW [GLOBAL|SESSION] STATUS [LIKE 匹配的模式];
```

