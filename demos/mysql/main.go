package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:spq@2029@tcp(127.0.0.1:3306)/world?charset=utf8")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	fmt.Println("connect to mysql success")

	// 设置数据库最大连接数
	db.SetConnMaxLifetime(100)
	// 设置上数据库最大闲置连接数
	db.SetMaxIdleConns(10)
	// Ping
	if err := db.Ping(); err != nil {
		fmt.Printf("ping database failed, error: [%s]", err.Error())
		return
	}
	rows, err := db.Query("select name from city limit 10")
	if err != nil {
		fmt.Printf("select failed, error: [%s]", err.Error())
		return
	}
	// 释放连接
	defer rows.Close()
	// 循环条件，遍历读取结果集
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			fmt.Printf("get data failed, error: [%s]", err.Error())
			return
		}

		fmt.Println(name)
	}
}
