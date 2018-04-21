package category

import (
	"github.com/smartwalle/dbs"
)

type manager struct {
	db    dbs.DB
	table string
}
