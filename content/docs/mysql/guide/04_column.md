---
title: 列
weight: 4
---

## 默认值

一个列如果没有显示的指定默认值，那么在插入记录时，该列的值就是 `NULL`。如果要指定默认值，那么可以使用 `DEFAULT` 关键字，语法如下：

```sql
列名 列的类型 DEFAULT 默认值
```

示例：

```bash
mysql> CREATE TABLE first_table (
    ->     first_column INT,
    ->     second_column VARCHAR(100) DEFAULT 'abc'
    -> );
Query OK, 0 rows affected (0.02 sec)
```

## NOT NULL

如果列中必须有值，那么可以在创建列时使用 `NOT NULL` 关键字，这样在插入记录时，如果该列没有值，那么就会报错：

```sql
列名 列的类型 NOT NULL
```

示例：

```bash
mysql> CREATE TABLE first_table (
    ->     first_column INT NOT NULL,
    ->     second_column VARCHAR(100) DEFAULT 'abc'
    -> );
Query OK, 0 rows affected (0.02 sec)
```

## 主键

一个表最多只能有一个主键，主键的值不能重复，并且不能为 `NULL`，通过主键可以找到唯一的一条记录。如果一个列被指定为主键，那么在插入记录时，如果该列没有值，那么就会报错。

**主键可以是一个列，也可以是多个列的组合**。

如果主键只是单个列的话，可以直接在该列后声明 `PRIMARY KEY`：

```sql
CREATE TABLE student_info (
    number INT PRIMARY KEY,
    name VARCHAR(5),
    sex ENUM('男', '女'),
    id_number CHAR(18),
    department VARCHAR(30),
    major VARCHAR(30),
    enrollment_time DATE
);
```

也可以把主键的声明单独提取出来：

```sql
CREATE TABLE student_info (
    number INT,
    name VARCHAR(5),
    sex ENUM('男', '女'),
    id_number CHAR(18),
    department VARCHAR(30),
    major VARCHAR(30),
    enrollment_time DATE,
    PRIMARY KEY (number)
);
```

对于多个列的组合作为主键的情况，必须使用这种单独声明的形式：

```sql
CREATE TABLE student_score (
    number INT,
    subject VARCHAR(30),
    score TINYINT,
    PRIMARY KEY (number, subject)
);
```

## UNIQUE

一个表中可以有多个 `UNIQUE` 约束，`UNIQUE` 约束的值不能重复，但是可以为 `NULL`。

单个列声明 `UNIQUE` 属性：

```sql
CREATE TABLE student_info (
    number INT PRIMARY KEY,
    name VARCHAR(5),
    sex ENUM('男', '女'),
    id_number CHAR(18) UNIQUE,
    department VARCHAR(30),
    major VARCHAR(30),
    enrollment_time DATE
);
```

也可以把 `UNIQUE` 属性的声明单独提取出来：

```sql
CREATE TABLE student_info (
    number INT PRIMARY KEY,
    name VARCHAR(5),
    sex ENUM('男', '女'),
    id_number CHAR(18),
    department VARCHAR(30),
    major VARCHAR(30),
    enrollment_time DATE,
    UNIQUE KEY uk_id_number (id_number)
);
```

`UNIQUE [约束名称] (列名1, 列名2, ...)` 中的 `UNIQUE` 也可以使用 `UNIQUE KEY` 关键字来声明。

对于多个列的组合具有 `UNIQUE` 属性的情况，必须使用这种单独声明的形式。

## AUTO_INCREMENT

`AUTO_INCREMENT` 自增属性,如果一个表中的某个列的数据类型是**整数类型或者浮点数类型，就可以设置 `AUTO_INCREMENT` 属性**。

设置了 `AUTO_INCREMENT` 属性之后，如果在插入新记录的时候不指定该列的值，那么新插入的记录的该列的值就会自动递增。

```bash
mysql> CREATE TABLE first_table (
    ->     id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    ->     first_column INT,
    ->     second_column VARCHAR(100) DEFAULT 'abc'
    -> );
Query OK, 0 rows affected (0.01 sec)
```

## 注释

每一个列**末尾**可以添加 `COMMENT` 语句来为列来添加注释：

```sql
CREATE TABLE first_table (
    id int UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '自增主键',
    first_column INT COMMENT '第一列',
    second_column VARCHAR(100) DEFAULT 'abc' COMMENT '第二列'
) COMMENT '第一个表';
```

## ZEROFILL

`ZEROFILL` 属性配合显示宽度，可以实现**数字的左边补 0**。