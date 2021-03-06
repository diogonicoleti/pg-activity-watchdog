package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/diogonicoleti/pg-activity-watchdog/watchdog"
	_ "github.com/lib/pq"

	"gopkg.in/robfig/cron.v2"
)

var (
	app            = kingpin.New("Postgres Activity Watchdog", "An app that monitors some metrics and takes a snapshot if it exceeds a configured threshold")
	version        = "dev"
	dataSourceName = app.Flag("datasource", "Database connection string").
			Short('d').
			Default("user=postgres dbname=postgres sslmode=disable").String()
	threshold = app.Flag("threshold", "Threshold to take a snapshot").
			Short('t').
			Default("30").Int()
	interval = app.Flag("interval", "Interval to execute the watchdog").
			Short('i').
			Default("1s").String()
)

func main() {
	app.Version(version)
	app.DefaultEnvars()
	kingpin.MustParse(app.Parse(os.Args[1:]))

	log.Infof("Starting PostgreSQL activity watchdog %s", version)
	log.Infof("%v", *dataSourceName)
	watchdog := watchdog.NewWatchdog(
		*dataSourceName,
		*threshold,
	)

	c := cron.New()
	if _, err := c.AddFunc("@every "+*interval, watchdog.Execute); err != nil {
		log.WithError(err).Fatal("Failed to schedule watchdog")
	}
	c.Start()

	waitInterruptSignal()
	log.Info("Stopping PostgreSQL activity watchdog")
	c.Stop()
}

func waitInterruptSignal() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println()
		log.Infof("Received signal: %s", sig)
		done <- true
	}()

	<-done
}
