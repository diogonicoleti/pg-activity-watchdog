package watchdog

import (
	"github.com/apex/log"
	"github.com/jmoiron/sqlx"
)

func connect(dataSourceName string) *sqlx.DB {
	db, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		log.WithError(err).Fatal("Failed to open connection to database")
	}
	db.SetMaxOpenConns(2)
	return db
}
