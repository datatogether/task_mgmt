// transactional email handled by postmark
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func SendTaskRequestEmail(t *Task) error {
	if len(cfg.EmailNotificationRecipients) == 0 {
		return fmt.Errorf("no recipients are set to send email to")
	}

	body := fmt.Sprintf(`{
    "From" : "brendan@qri.io",
    "To" : "%s",
    "Tag" : "feedback",
    "Subject" : "Injest Request: %s",
    "TextBody" : "requested: %s\nsource url: %s\n"
  }`,
		strings.Join(cfg.EmailNotificationRecipients, ","),
		t.Title,
		t.Request,
		t.SourceUrl,
	)

	return sendEmail(strings.NewReader(body))
}

func SendTaskCancelEmail(t *Task) error {
	if len(cfg.EmailNotificationRecipients) == 0 {
		return fmt.Errorf("no recipients are set to send email to")
	}

	body := fmt.Sprintf(`{
    "From" : "brendan@qri.io",
    "To" : "%s",
    "Tag" : "feedback",
    "Subject" : "Request Cancelled: %s",
    "TextBody" : "requested: %s\nsource url: %s\ncancelled: %s"
  }`,
		strings.Join(cfg.EmailNotificationRecipients, ","),
		t.Title,
		t.Request,
		t.SourceUrl,
		t.Fail,
	)

	return sendEmail(strings.NewReader(body))
}

// send an email using postmark transactional email service
// postmarkapp.com
func sendEmail(jsonBody io.Reader) error {
	if cfg.PostmarkKey == "" {
		return fmt.Errorf("missing postmark key for sending email")
	}

	url := "https://api.postmarkapp.com/email/"

	req, err := http.NewRequest("POST", url, jsonBody)
	if err != nil {
		return err
	}
	req.Header.Add("X-Postmark-Server-Token", cfg.PostmarkKey)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Body = ioutil.NopCloser(jsonBody)

	res, err := http.DefaultClient.Do(req)
	// if the server responds with an error, process & log out
	if res.StatusCode == 422 {
		responseBody := map[string]interface{}{}
		json.NewDecoder(res.Body).Decode(&responseBody)
		logger.Println(responseBody)
	}

	return err
}

// send an email with a server-side template
// func sendTemplateEmail(jsonBody io.Reader) error {
//   url := "https://api.postmarkapp.com/email/withTemplate/"

//   req, err := http.NewRequest("POST", url, jsonBody)
//   if err != nil {
//     return New500Error(err.Error())
//   }
//   req.Header.Add("X-Postmark-Server-Token", config.PostmarkKey)
//   req.Header.Add("Accept", "application/json")
//   req.Header.Add("Content-Type", "application/json")
//   req.Body = ioutil.NopCloser(jsonBody)

//   res, err := http.DefaultClient.Do(req)
//   if err != nil {
//     return err
//   }
//   defer res.Body.Close()

//   // if the server responds with an error, process & log
//   if res.StatusCode == 422 {
//     responseBody := map[string]interface{}{}
//     json.NewDecoder(res.Body).Decode(&responseBody)
//     logger.Println(responseBody)
//   }

//   return Error500IfErr(err)
// }
