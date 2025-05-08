---
title: 数据类型
weight: 1
---

## 数值

### 整数类型

MySQL 支持 SQL 标准支持的整型类型：`INT`、`SMALLINT`。也支持 `TINYINT`、`MEDIUMINT` 和 `BIGINT` 整型类型。

| 类型 | 占用的存储空间（单位：字节） | 无符号数取值范围 | 有符号数取值范围 | 含义 |
| --- | --- | --- | --- | --- |
| TINYINT | 1 | 0 ~ 2⁸-1 | -2⁷ ~ 2⁷-1  | 非常小的整数 |
| SMALLINT | 2 | 0 ~ 2¹⁶-1 | -2¹⁵ ~ 2¹⁵-1 | 小的整数 |
| MEDIUMINT | 3 | 0 ~ 2²⁴-1 | -2²³ ~ 2²³-1 | 中等大小的整数 |
| INT（别名：INTEGER） | 4 | 0 ~ 2³²-1 | -2³¹ ~ 2³¹-1 | 标准的整数 |
| BIGINT | 8 | 0 ~ 2⁶⁴-1 | -2⁶³ ~ 2⁶³-1 | 大整数 |


避免使用整数的显示宽度(参看文档最后)，也就是说，不要用 `INT(10)` 类似的方法指定字段显示宽度，直接用 `INT`。

#### INT 显示宽度

创建数据表时整数类型可以指定一个长度，如下。但是，这里的长度并非是 TINYINT 类型存储的最大长度，而是显示的最大长度。
```sql
CREATE TABLE `user`(
  `id` TINYINT(2) UNSIGNED
);
```
这里表示 user 表的 id 字段的类型是 TINYINT，可以存储的最大数值是 255。所以，在存储数据时，如果存入值小于等于 255，如 200，虽然超过 2 位，但是没有超出 TINYINT 类型长度，所以可以正常保存；如果存入值大于 255，如 500，那么 MySQL 会自动保存为 TINYINT 类型的最大值255。

在查询数据时，不管查询结果为何值，都按实际输出。这里 `TINYINT(2)` 中 `2` 的作用就是，当需要在查询结果前填充 `0` 时，
命令中加 `上ZEROFILL` 就可以实现，如：
```sql
`id` TINYINT(2) UNSIGNED ZEROFILL
```

这样，查询结果如果是 `5`，那输出就是 `05`。如果指定 `TINYINT(5)`，那输出就是 `00005`，其实实际存储的值还是 `5`，而且存储的数据不会超过 `255`，只是 MySQL 输出数据时在前面填充了 `0`。


## 浮点数

## 定点数

## 无符号数值类型


```sql
SET @a = '{
  "cellphone": "188888888888",
  "wxchat": "this.",
  "QQ": "166666666"
}';

INSERT INTO UserLogin VALUES (1, @a);
```

```sql
SELECT 
    user_id,
    JSON_UNQUOTE(JSON_EXTRACT(loginInfo,"$.cellphone")) cellphone,
    JSON_UNQUOTE(JSON_EXTRACT(loginInfo,"$.wxchat")) wxchat
FROM UserLogin;
```

使用 `->>` 表达式，效果一样：

```sql
SELECT
    userId,
    loginInfo->>"$.cellphone" cellphone,
    loginInfo->>"$.wxchat" wxchat
FROM UserLogin;
```