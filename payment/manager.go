package payment

import "github.com/smartwalle/dbs"

type Manager struct {
	db dbs.DB
}

func NewManager(db dbs.DB) *Manager {
	var m = &Manager{}
	m.db = db
	return m
}
