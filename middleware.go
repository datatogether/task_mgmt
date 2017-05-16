package main

import (
	// "encoding/json"
	"fmt"
	"net/http"
	"time"
)

// middleware handles request logging
func middleware(handler http.HandlerFunc) http.HandlerFunc {
	// no-auth middware func
	return func(w http.ResponseWriter, r *http.Request) {
		// poor man's logging:
		fmt.Println(r.Method, r.URL.Path, time.Now())

		// If this server is operating behind a proxy, but we still want to force
		// users to use https, cfg.ProxyForceHttps == true will listen for the common
		// X-Forward-Proto & redirect to https
		if cfg.ProxyForceHttps {
			if r.Header.Get("X-Forwarded-Proto") == "http" {
				w.Header().Set("Connection", "close")
				url := "https://" + r.Host + r.URL.String()
				http.Redirect(w, r, url, http.StatusMovedPermanently)
				return
			}
		}

		addCORSHeaders(w, r)

		// TODO - Strict Transport config?
		// if cfg.TLS {
		// 	// If TLS is enabled, set 1 week strict TLS, 1 week for now to prevent catastrophic mess-ups
		// 	w.Header().Add("Strict-Transport-Security", "max-age=604800")
		// }
		handler(w, r)
	}
}

// authMiddleware adds http basic auth if configured
func authMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// token := r.FormValue("access_token")
		// c, err := r.Cookie(cfg.UserCookieKey)
		// if err != nil {
		// 	logger.Println(err.Error())
		// }

		// // we gots no login info, so login required
		// if c == nil && token == "" {
		// 	renderTemplate(w, "login.html", nil)
		// 	return
		// }

		// req, err := http.NewRequest("GET", fmt.Sprintf("%s/session/oauth/github/repoaccess?access_token=%s&owner=%s&repo=%s", cfg.IdentityServerUrl, token, cfg.GithubRepoOwner, cfg.GithubRepoName), nil)

		// if err != nil {
		// 	renderError(w, fmt.Errorf("error contacting identity server: %s", err.Error()))
		// 	logger.Println(err.Error())
		// 	return
		// }

		// req.AddCookie(c)
		// res, err := http.DefaultClient.Do(req)
		// if err != nil {
		// 	renderError(w, fmt.Errorf("error contacting identity server: %s", err.Error()))
		// 	logger.Println(err.Error())
		// 	return
		// }
		// defer res.Body.Close()

		// data := map[string]interface{}{}
		// if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		// 	renderError(w, fmt.Errorf("error contacting identity server: %s", err.Error()))
		// 	return
		// }

		// // User Needs github added to their account
		// if res.StatusCode == http.StatusUnauthorized {
		// 	// renderError(w, fmt.Errorf("%s", data["meta"]))
		// 	renderTemplate(w, "login.html", nil)
		// 	return
		// } else if res.StatusCode != http.StatusOK {
		// 	renderError(w, fmt.Errorf("%s", data["meta"].(map[string]interface{})["message"]))
		// 	return
		// }

		// perm := data["data"].(map[string]interface{})["permission"]
		// if perm != "admin" && perm != "write" {
		// 	renderTemplate(w, "accessDenied.html", nil)
		// 	return
		// }

		// no-auth middware func
		middleware(handler)(w, r)
	}
}

// addCORSHeaders adds CORS header info for whitelisted servers
func addCORSHeaders(w http.ResponseWriter, r *http.Request) {
	// origin := r.Header.Get("Origin")
	// for _, o := range cfg.AllowedOrigins {
	// 	if origin == o {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	// return
	// 	}
	// }
}
