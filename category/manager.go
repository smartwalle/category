package category

import (
	"github.com/smartwalle/dbs"
)

type Manager struct {
	db    dbs.SQLExecutor
	table string
}

func NewManager(db dbs.SQLExecutor, table string) *Manager {
	var m = &Manager{}
	m.db = db
	m.table = table
	return m
}