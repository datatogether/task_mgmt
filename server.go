// Task Management manages tasks, including tracking the state of tasks as they move through a queue
// As tasks are completed task-mgmt updates records of when tasks started, stopped, etc.
package main

import (
	"database/sql"
	"fmt"
	"github.com/datatogether/sql_datastore"
	"github.com/datatogether/task-mgmt/source"
	"github.com/datatogether/task-mgmt/taskdefs/ipfs"
	"github.com/datatogether/task-mgmt/tasks"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

var (
	// cfg is the global configuration for the server. It's read in at startup from
	// the config.json file and enviornment variables, see config.go for more info.
	cfg *config

	// When was the last alert sent out?
	// Use this value to avoid bombing alerts
	lastAlertSent *time.Time

	// log output
	log = logrus.New()

	// application database connection
	appDB *sql.DB

	// hoist default store
	store = sql_datastore.DefaultStore
)

func init() {
	log.Out = os.Stderr
	log.Level = logrus.InfoLevel
	log.Formatter = &logrus.TextFormatter{
		ForceColors: true,
	}

	tasks.RegisterTaskdef("ipfs.add", ipfs.NewTaskAdd)
}

func main() {
	var err error
	cfg, err = initConfig(os.Getenv("GOLANG_ENV"))
	if err != nil {
		// panic if the server is missing a vital configuration detail
		panic(fmt.Errorf("server configuration error: %s", err.Error()))
	}

	connectToAppDb()
	sql_datastore.SetDB(appDB)
	store.Register(
		&tasks.Task{},
		&source.Source{},
	)
	// go update(appDB)
	go listenRpc()

	stop, err := acceptTasks()
	if err != nil {
		panic(err.Error())
	}

	s := &http.Server{}
	// connect mux to server
	s.Handler = NewServerRoutes()

	// print notable config settings
	// printConfigInfo()

	// fire it up!
	log.Info("starting server on port", cfg.Port)

	// start server wrapped in a log.Fatal b/c http.ListenAndServe will not
	// return unless there's an error
	log.Fatal(StartServer(cfg, s))

	// lol will never happen, left here as a reminder
	// that we should be able to stop accepting new tasks
	// at any point without issue
	stop <- true
}

// NewServerRoutes returns a Muxer that has all API routes.
// This makes for easy testing using httptest
func NewServerRoutes() *http.ServeMux {
	m := http.NewServeMux()
	m.HandleFunc("/.well-known/acme-challenge/", CertbotHandler)
	m.Handle("/", middleware(NotFoundHandler))

	m.Handle("/tasks", middleware(TasksHandler))
	m.Handle("/tasks/", middleware(TaskHandler))
	// TODO - restore this:
	// m.Handle("/tasks/cancel/", middleware(CancelTaskHandler))

	// Example of individual task routing:
	m.HandleFunc("/ipfs/add", middleware(EnqueueIpfsAddHandler))

	m.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("public/js"))))
	m.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("public/css"))))

	return m
}
