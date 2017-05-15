/*
	Coverage is a service for mapping an archiving surface area, and tracking
	the amount of that surface area that any number of archives have covered
*/
package main

import (
	"database/sql"
	"fmt"
	"log"
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
	logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

	// application database connection
	appDB *sql.DB
)

func main() {
	var err error
	cfg, err = initConfig(os.Getenv("GOLANG_ENV"))
	if err != nil {
		// panic if the server is missing a vital configuration detail
		panic(fmt.Errorf("server configuration error: %s", err.Error()))
	}

	connectToAppDb()

	s := &http.Server{}
	m := http.NewServeMux()
	m.HandleFunc("/.well-known/acme-challenge/", CertbotHandler)
	m.Handle("/", authMiddleware(HomeHandler))

	m.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("public/js"))))
	m.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("public/css"))))

	// connect mux to server
	s.Handler = m

	// print notable config settings
	// printConfigInfo()

	// fire it up!
	fmt.Println("starting server on port", cfg.Port)

	// start server wrapped in a log.Fatal b/c http.ListenAndServe will not
	// return unless there's an error
	logger.Fatal(StartServer(cfg, s))
}
