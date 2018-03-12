package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/smartwalle/dbs"
	"github.com/smartwalle/m4go/category"
	"fmt"
)

func main() {
	var pool = dbs.NewSQL("mysql", "", 10, 10)

	var cm = category.NewManager(pool.GetSession(), "category")

	categoryList, err := cm.GetCategoryList(39, 0, 0, 0)
	if err != nil {
		fmt.Println("err", err)
		return
	}

	for _, category := range categoryList {
		fmt.Println(category.Type, category.Id, category.Name, category.Description, category.LeftValue, category.RightValue)
	}
}

