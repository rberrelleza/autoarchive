package main

import (
	"flag"
	"os"

	_ "github.com/lib/pq"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("main.go")
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func main() {
	var (
		mode       = flag.String("mode", "all", "mode: all|web|worker|scheduler")
		baseURL    = flag.String("baseurl", os.Getenv("BASE_URL"), "local base url")
		schedule   = flag.String("schedule", "24h", "How often to evaluate idleness")
		loglevel   = flag.String("loglevel", "DEBUG", "Log level")
		pghost     = flag.String("pghost", "localhost", "PG Host")
		pgdatabase = flag.String("pgdatabase", "hiparchiver", "PG Database")
		pguser     = flag.String("pguser", "postgres", "PG User")
		pgpass     = flag.String("pgpass", "postgres", "PG Password")
	)

	flag.Parse()

	backend := logging.NewLogBackend(os.Stdout, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	backendLeveled := logging.AddModuleLevel(backendFormatter)
	parsedLogLevel, error := logging.LogLevel(*loglevel)
	checkErr(error)
	backendLeveled.SetLevel(parsedLogLevel, "")
	logging.SetBackend(backendLeveled)

	context := &Context{baseURL: *baseURL, pghost: *pghost, pguser: *pguser, pgpass: *pgpass, pgdatabase: *pgdatabase, nworkers: 4}

	log.Infof("Starting hiparchiver with role %s", *mode)
	switch *mode {
	case "all":
		context.RunDispatcher()
		context.RunScheduler(schedule)
		context.RunWeb()
	case "web":
		context.RunWeb()
	case "worker":
		context.RunDispatcher()
	case "scheduler":
		context.RunScheduler(schedule)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
