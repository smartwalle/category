package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/smartwalle/m4go/category"
)

func main() {
	var db, _ = sql.Open("mysql", "")
	var cm = category.NewSercie(db, "category")

	categoryList, err := cm.GetCategoryList(47, 0, 0)
	if err != nil {
		fmt.Println("err", err)
		return
	}

	for _, category := range categoryList {
		fmt.Println(category.Type, category.Id, category.IsLeafNode(), category.Name, category.Description, category.LeftValue, category.RightValue, category.Ext1, category.Ext2)
	}
}
