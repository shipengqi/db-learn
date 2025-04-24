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