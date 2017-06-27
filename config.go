package main

import (
	"fmt"
	conf "github.com/archivers-space/config"
	"os"
	"path/filepath"
)

// server modes
const (
	DEVELOP_MODE    = "develop"
	PRODUCTION_MODE = "production"
	TEST_MODE       = "test"
)

// config holds all configuration for the server. It pulls from three places (in order):
// 		1. environment variables
// 		2. .[MODE].env OR .env
//
// globally-set env variables win.
// it's totally fine to not have, say, .env.develop defined, and just
// rely on a base ".env" file. But if you're in production mode & ".env.production"
// exists, that will be read *instead* of .env
//
// configuration is read at startup and cannot be alterd without restarting the server.
type config struct {
	// site title
	Title string

	// port to listen on, will be read from PORT env variable if present.
	Port string

	// root url for service
	UrlRoot string

	// url of postgres app db
	PostgresDbUrl string

	// url of message que server
	AmqpUrl string

	// Public Key to use for signing. required.
	PublicKey string

	// TLS (HTTPS) enable support via LetsEncrypt, default false
	// not needed if operating behind a TLS proxy
	TLS bool
	// if true, requests that have X-Forwarded-Proto: http will be redirected
	// to their https variant, useful if operating behind a TLS proxy
	ProxyForceHttps bool

	// key for sending emails
	PostmarkKey string
	// list of email addresses that should get notifications
	EmailNotificationRecipients []string

	// url to kick off github oauth process
	GithubLoginUrl string
	// owner of github repo. required
	GithubRepoOwner string
	// name of github repo. required.
	GithubRepoName string

	// location of identity server
	IdentityServerUrl string
	// cookie to check for user credentials to forward to identity server.
	UserCookieKey string

	// CertbotResponse is only for doing manual SSL certificate generation via LetsEncrypt.
	CertbotResponse string
}

// initConfig pulls configuration from config.json
func initConfig(mode string) (cfg *config, err error) {
	cfg = &config{}

	if path := configFilePath(mode, cfg); path != "" {
		log.Infof("loading config file: %s", filepath.Base(path))
		if err := conf.Load(cfg, path); err != nil {
			log.Info("error loading config:", err)
		}
	} else {
		if err := conf.Load(cfg); err != nil {
			log.Info("error loading config:", err)
		}
	}

	// make sure port is set
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	err = requireConfigStrings(map[string]string{
		"PORT":                cfg.Port,
		"POSTGRES_DB_URL":     cfg.PostgresDbUrl,
		"GITHUB_REPO_OWNER":   cfg.GithubRepoOwner,
		"GITHUB_REPO_NAME":    cfg.GithubRepoName,
		"IDENTITY_SERVER_URL": cfg.IdentityServerUrl,
	})

	return
}

func packagePath(path string) string {
	return filepath.Join(os.Getenv("GOPATH"), "src/github.com/archivers-space/task-mgmt", path)
}

// requireConfigStrings panics if any of the passed in values aren't set
func requireConfigStrings(values map[string]string) error {
	for key, value := range values {
		if value == "" {
			return fmt.Errorf("%s env variable or config key must be set", key)
		}
	}

	return nil
}

// checks for .[mode].env file to read configuration from if the file exists
// defaults to .env, returns "" if no file is present
func configFilePath(mode string, cfg *config) string {
	fileName := packagePath(fmt.Sprintf(".%s.env", mode))
	if !fileExists(fileName) {
		fileName = packagePath(".env")
		if !fileExists(fileName) {
			return ""
		}
	}
	return fileName
}

// Does this file exist?
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// outputs any notable settings to stdout
func printConfigInfo() {
	// TODO
}
