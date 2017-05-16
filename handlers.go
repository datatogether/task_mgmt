package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
)

// templates is a collection of views for rendering with the renderTemplate function
// see homeHandler for an example
var templates = template.Must(template.ParseFiles(
	"views/accessDenied.html",
	"views/error.html",
	"views/expired.html",
	"views/index.html",
	"views/login.html",
	"views/notFound.html",
	"views/resources.html",
))

// renderTemplate renders a template with the values of cfg.TemplateData
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	if data == nil {
		data = cfg.TemplateData
	}
	err := templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderError(w http.ResponseWriter, err error) {
	renderTemplate(w, "error.html", map[string]string{
		"message": err.Error(),
	})
}

func reqParamInt(key string, r *http.Request) (int, error) {
	i, err := strconv.ParseInt(r.FormValue(key), 10, 0)
	return int(i), err
}

func reqParamBool(key string, r *http.Request) (bool, error) {
	return strconv.ParseBool(r.FormValue(key))
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := ReadTasks(appDB, "created DESC", 30, 0)
	if err != nil {
		renderError(w, err)
		return
	}

	renderTemplate(w, "resources.html", map[string]interface{}{
		"tasks": tasks,
	})
}

// HealthCheckHandler is a basic "hey I'm fine" for load balancers & co
// TODO - add Database connection & proper configuration checks here for more accurate
// health reporting
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{ "status" : 200 }`))
}

// EmptyOkHandler is an empty 200 response, often used
// for OPTIONS requests that responds with headers set in addCorsHeaders
func EmptyOkHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// CertbotHandler pipes the certbot response for manual certificate generation
func CertbotHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, cfg.CertbotResponse)
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{ "status" :  "not found" }`))
}
