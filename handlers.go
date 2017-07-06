package main

import (
	"github.com/datatogether/api/apiutil"
	"github.com/datatogether/task-mgmt/tasks"
	"io"
	"net/http"
	"strconv"
)

func TasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		EnqueueTaskHandler(w, r)
	case "GET":
		ListTasksHandler(w, r)
	default:
		NotFoundHandler(w, r)
	}
}

func EnqueueTaskHandler(w http.ResponseWriter, r *http.Request) {
	t := &tasks.Task{
		Type: "ipfs.add",
		Params: map[string]interface{}{
			"url":              r.FormValue("url"),
			"ipfsApiServerUrl": cfg.IpfsApiUrl,
		},
	}

	if err := t.Enqueue(store, cfg.AmqpUrl); err != nil {
		log.Infoln(err)
		apiutil.WriteErrResponse(w, http.StatusBadRequest, err)
		return
	}

	apiutil.WriteMessageResponse(w, "successfully enqueued task", t)
}

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		ReadTaskHandler(w, r)
	case "POST":
		EnqueueTaskHandler(w, r)
	default:
		NotFoundHandler(w, r)
	}
}

func ReadTaskHandler(w http.ResponseWriter, r *http.Request) {
	t := &tasks.Task{
		Id: r.URL.Path[len("/tasks/"):],
	}
	if err := t.Read(store); err != nil {
		apiutil.WriteErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	apiutil.WriteResponse(w, t)
}

func EnqueueIpfsAddHandler(w http.ResponseWriter, r *http.Request) {
	t := &tasks.Task{
		Type: "ipfs.add",
		Params: map[string]interface{}{
			"url":              r.FormValue("url"),
			"ipfsApiServerUrl": cfg.IpfsApiUrl,
		},
	}

	if err := t.Enqueue(store, cfg.AmqpUrl); err != nil {
		log.Infoln(err)
		apiutil.WriteErrResponse(w, http.StatusBadRequest, err)
		return
	}

	apiutil.WriteMessageResponse(w, "task successfully enqueued", t)
}

func reqParamInt(key string, r *http.Request) (int, error) {
	i, err := strconv.ParseInt(r.FormValue(key), 10, 0)
	return int(i), err
}

func reqParamBool(key string, r *http.Request) (bool, error) {
	return strconv.ParseBool(r.FormValue(key))
}

func ListTasksHandler(w http.ResponseWriter, r *http.Request) {
	p := apiutil.PageFromRequest(r)
	ts, err := tasks.ReadTasks(store, "created DESC", p.Limit(), p.Offset())
	if err != nil {
		log.Infoln(err.Error())
		apiutil.WriteErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	apiutil.WritePageResponse(w, ts, r, p)
}

// func HomeHandler(w http.ResponseWriter, r *http.Request) {
// 	tsks, err := tasks.ReadTasks(store, "created DESC", 30, 0)
// 	if err != nil {
// 		log.Info(err.Error())
// 		renderError(w, err)
// 		return
// 	}

// 	perm := false
// 	pValue := r.Context().Value("permission")
// 	if pString, ok := pValue.(string); ok {
// 		if pString == "admin" || pString == "write" {
// 			perm = true
// 		}
// 	}

// 	renderTemplate(w, "tasks.html", map[string]interface{}{
// 		"writePerm": perm,
// 		"tasks":     tsks,
// 	})
// }

// func RunTaskHandler(w http.ResponseWriter, r *http.Request) {
// 	t := &tasks.Task{
// 		Id: r.URL.Path[len("/tasks/run/"):],
// 	}
// 	if err := t.Read(store); err != nil {
// 		renderError(w, err)
// 		return
// 	}

// 	// if err := t.Run(store); err != nil {
// 	// 	renderError(w, err)
// 	// 	return
// 	// }

// 	renderMessage(w, "Now Running Task", "We've shipped your task off for execution, check back here in 12-24 hours to see status!")
// }

// TODO - restore
func CancelTaskHandler(w http.ResponseWriter, r *http.Request) {
	// t := &tasks.Task{
	// 	Id: r.URL.Path[len("/tasks/cancel/"):],
	// }
	// if err := t.Read(store); err != nil {
	// 	renderError(w, err)
	// 	return
	// }

	// if err := t.Cancel(store); err != nil {
	// 	renderError(w, err)
	// 	return
	// }

	// renderMessage(w, "Task Cancelled", "You've cancelled this task")
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
