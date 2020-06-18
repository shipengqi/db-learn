package main

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// https://gorm.io/zh_CN/

type City struct {
	ID   int // 名为 `ID` 的字段会默认作为表的主键
	Name string
}

// 表名默认就是结构体名称的复数，但这里的表名是 city
// 可以禁用默认表名的复数形式，如果置为 true，则 `City` 的默认表名是 `city`，但是会应用到所有表名
// db.SingularTable(true)
// 所以这里使用 TableName 方法
// 将 City 的表名设置为 `city`
func (City) TableName() string {
	return "city"
}

// 也可以在执行具体指令时，指定表名，如 db.Table("deleted_users").Find(&deleted_users)

func main() {
	db, err := gorm.Open("mysql", "root:spq@2029@tcp(127.0.0.1:3306)/world?charset=utf8")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	fmt.Println("connect to mysql success")

	cities := []*City{}
	err = db.Limit(10).Select("name").Find(&cities).Error
	if err != nil {
		fmt.Printf("find failed, error: [%s]", err.Error())
		return
	}
	for _, v := range cities {
		fmt.Println(v.Name)
	}
	

	fmt.Println("-------------------------------------")
	rows, err := db.Raw("select name from city limit 10").Rows()
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
