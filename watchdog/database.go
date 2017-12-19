package watchdog

import (
	"github.com/jmoiron/sqlx"
)

func connect(dataSourceName string) *sqlx.DB {
	db := sqlx.MustConnect("postgres", dataSourceName)
	db.SetMaxOpenConns(2)
	return db
}
