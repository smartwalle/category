package category

import (
	"github.com/smartwalle/dbs"
)

type Manager struct {
	db    dbs.SQLExecutor
	table string
}

func NewManager(db dbs.SQLExecutor, table string) *Manager {
	var cm = &Manager{}
	cm.db = db
	cm.table = table
	return cm
}