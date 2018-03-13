package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/smartwalle/m4go/category"
	"fmt"
	"database/sql"
)

func main() {
	var db, _ = sql.Open("mysql", "root:smok2015@tcp(192.168.192.250:3306)/titan_dev?parseTime=true")
	var cm = category.NewManager(db, "category")

	categoryList, err := cm.GetCategoryList(0, 0, 0, 0, "", 0)
	if err != nil {
		fmt.Println("err", err)
		return
	}

	for _, category := range categoryList {
		fmt.Println(category.Type, category.Id, category.Name, category.Description, category.LeftValue, category.RightValue, category.Ext1, category.Ext2)
	}

	fmt.Println(cm.MoveToRoot(37))
}
