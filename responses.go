package main

import (
	"encoding/json"
	"net/http"
)

func writeResponse(w http.ResponseWriter, data interface{}) error {
	env := map[string]interface{}{
		"meta": map[string]interface{}{
			"code": http.StatusOK,
		},
		"data": data,
	}
	return jsonResponse(w, env)
}

func writeMessageResponse(w http.ResponseWriter, message string, data interface{}) error {
	env := map[string]interface{}{
		"meta": map[string]interface{}{
			"code":    http.StatusOK,
			"message": message,
		},
		"data": data,
	}

	return jsonResponse(w, env)
}

func writeErrResponse(w http.ResponseWriter, code int, err error) error {
	env := map[string]interface{}{
		"meta": map[string]interface{}{
			"code":  code,
			"error": err.Error(),
		},
	}

	res, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Println(err.Error())
		return err
	}

	w.WriteHeader(code)
	_, err = w.Write(res)
	return err
}

func jsonResponse(w http.ResponseWriter, env interface{}) error {
	res, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Println(err.Error())
		return err
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(res)
	return err
}
