/*
	Coverage is a service for mapping an archiving surface area, and tracking
	the amount of that surface area that any number of archives have covered
*/
package main

import (
	"database/sql"
	"fmt"
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
)

func init() {
	log.Out = os.Stderr
	log.Level = logrus.InfoLevel
	log.Formatter = &logrus.TextFormatter{
		ForceColors: true,
	}
}

func main() {
	var err error
	cfg, err = initConfig(os.Getenv("GOLANG_ENV"))
	if err != nil {
		// panic if the server is missing a vital configuration detail
		panic(fmt.Errorf("server configuration error: %s", err.Error()))
	}

	connectToAppDb()
	go update(appDB)

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
}

// NewServerRoutes returns a Muxer that has all API routes.
// This makes for easy testing using httptest
func NewServerRoutes() *http.ServeMux {
	m := http.NewServeMux()
	m.HandleFunc("/.well-known/acme-challenge/", CertbotHandler)
	m.Handle("/", authMiddleware(HomeHandler))
	m.Handle("/tasks/run/", authMiddleware(RunTaskHandler))
	m.Handle("/tasks/cancel/", authMiddleware(CancelTaskHandler))
	m.Handle("/tasks/success/", authMiddleware(TaskSuccessHandler))
	m.Handle("/tasks/fail/", authMiddleware(TaskFailHandler))

	m.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("public/js"))))
	m.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("public/css"))))

	return m
}
