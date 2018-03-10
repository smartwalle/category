package category

import (
	"github.com/smartwalle/dbs"
)

type CategoryManager struct {
	db    dbs.SQLExecutor
	table string
}

func NewManager(db dbs.SQLExecutor, table string) *CategoryManager {
	var cm = &CategoryManager{}
	cm.db = db
	cm.table = table
	return cm
}