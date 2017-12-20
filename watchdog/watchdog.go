package watchdog

import (
	"database/sql"
	"io/ioutil"
	"os"
	"time"

	"github.com/apex/log"
	"github.com/jmoiron/sqlx"
	yaml "gopkg.in/yaml.v2"
)

const (
	outputDir = "snapshots"
	fileMode  = 0777
)

// Watchdog is a watchdog that monitors PostgreSQL activity
type Watchdog struct {
	db        *sqlx.DB
	threshold int
}

type pgClientActivity struct {
	Total      int            `db:"total"`
	ClientAddr sql.NullString `db:"client_addr"`
}

type pgActivity struct {
	ProcessID    *int    `db:"pid" yaml:"pid"`
	Database     *string `db:"datname" yaml:"database"`
	User         *string `db:"usename" yaml:"username"`
	ClientAddr   *string `db:"client_addr" yaml:"client_addr"`
	BackendStart *string `db:"backend_start" yaml:"backend_start"`
	XactStart    *string `db:"xact_start" yaml:"xact_start"`
	QueryStart   *string `db:"query_start" yaml:"query_start"`
	StateChange  *string `db:"state_change" yaml:"state_change"`
	State        *string `db:"state" yaml:"state"`
	Query        *string `db:"query" yaml:"query"`
}

// NewWatchdog returns a new Watchdog
func NewWatchdog(dataSourceName string, threshold int) *Watchdog {
	if err := os.Mkdir(outputDir, fileMode); err != nil && !os.IsExist(err) {
		log.WithError(err).Fatal("Failed to create snapshots directory")
	}

	return &Watchdog{
		db:        connect(dataSourceName),
		threshold: threshold,
	}
}

// Execute gets PostgreSQL activities per client and if the
// connections counts exceeds the threshold takes a snapshot as
// a YAML file
func (w *Watchdog) Execute() {
	var clientActivities []pgClientActivity
	err := w.db.Select(&clientActivities,
		`select
				count(*) as total,
				client_addr
			from
				pg_stat_activity
			where
				state in ('idle in transaction','active')
			group by
				client_addr`)

	if err != nil {
		log.WithError(err).Error("Failed to get PostgreSQL activity")
	}

	for _, ca := range clientActivities {
		if ca.Total > w.threshold && ca.ClientAddr.Valid {
			err := w.snapshotActivities(ca.ClientAddr.String)
			if err != nil {
				log.WithError(err).Error("Failed to take PostgreSQL activity snapshot")
			}
		}
	}
}

func (w *Watchdog) snapshotActivities(clientAddr string) error {
	var activities []pgActivity
	err := w.db.Select(&activities,
		`select 
			pid, datname, usename, client_addr, backend_start, xact_start, query_start, state_change, state, query
		from pg_stat_activity
		where 
			client_addr = $1`,
		clientAddr)
	if err != nil {
		return err
	}

	bts, err := yaml.Marshal(activities)
	if err != nil {
		return err
	}

	log.Infof("Generating snapshot for client %s", clientAddr)
	return ioutil.WriteFile(
		outputDir+"/"+clientAddr+"_"+time.Now().Format(time.RFC3339)+".yaml",
		bts,
		fileMode,
	)
}
