package main

import (
	"github.com/datatogether/task-mgmt/tasks"
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
	"views/tasks.html",
	"views/message.html",
))

func ipfsAdd(w http.ResponseWriter, r *http.Request) {
	err := SubmitTask("ipfs.add", map[string]string{
		"url":              r.FormValue("url"),
		"ipfsApiServerUrl": cfg.IpfsApiUrl,
	})
	if err != nil {
		renderError(w, err)
		return
	}

	renderMessage(w, "dope", "message sent")
}

// renderTemplate renders a template with the values of cfg.TemplateData
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	if data == nil {
		data = map[string]string{
			"title":          cfg.Title,
			"GithubLoginUrl": cfg.GithubLoginUrl,
		}
	}
	err := templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		log.Info(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderError(w http.ResponseWriter, err error) {
	renderTemplate(w, "error.html", map[string]string{
		"message": err.Error(),
	})
}

func renderMessage(w http.ResponseWriter, title, message string) {
	renderTemplate(w, "error.html", map[string]string{
		"title":   title,
		"message": message,
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
	tsks, err := tasks.ReadTasks(store, "created DESC", 30, 0)
	if err != nil {
		log.Info(err.Error())
		renderError(w, err)
		return
	}

	perm := false
	pValue := r.Context().Value("permission")
	if pString, ok := pValue.(string); ok {
		if pString == "admin" || pString == "write" {
			perm = true
		}
	}

	renderTemplate(w, "tasks.html", map[string]interface{}{
		"writePerm": perm,
		"tasks":     tsks,
	})
}

func RunTaskHandler(w http.ResponseWriter, r *http.Request) {
	t := &tasks.Task{
		Id: r.URL.Path[len("/tasks/run/"):],
	}
	if err := t.Read(store); err != nil {
		renderError(w, err)
		return
	}

	// if err := t.Run(store); err != nil {
	// 	renderError(w, err)
	// 	return
	// }

	renderMessage(w, "Now Running Task", "We've shipped your task off for execution, check back here in 12-24 hours to see status!")
}

func CancelTaskHandler(w http.ResponseWriter, r *http.Request) {
	t := &tasks.Task{
		Id: r.URL.Path[len("/tasks/cancel/"):],
	}
	if err := t.Read(store); err != nil {
		renderError(w, err)
		return
	}

	if err := t.Cancel(store); err != nil {
		renderError(w, err)
		return
	}

	renderMessage(w, "Task Cancelled", "You've cancelled this task")
}

func TaskSuccessHandler(w http.ResponseWriter, r *http.Request) {
	t := &tasks.Task{
		Id: r.URL.Path[len("/tasks/success/"):],
	}
	if err := t.Read(store); err != nil {
		renderError(w, err)
		return
	}

	if err := t.Cancel(store); err != nil {
		renderError(w, err)
		return
	}

	renderMessage(w, "Task Completed", "We've marked this task as completed")
}

func TaskFailHandler(w http.ResponseWriter, r *http.Request) {
	t := &tasks.Task{
		Id: r.URL.Path[len("/tasks/fail/"):],
	}
	if err := t.Read(store); err != nil {
		renderError(w, err)
		return
	}

	msg := r.FormValue("message")
	if msg == "" {
		msg = "Task Failed"
	}

	if err := t.Errored(store, msg); err != nil {
		renderError(w, err)
		return
	}

	renderMessage(w, "Task Failed", "We've marked this task as failed. It can now be re-requested")
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
