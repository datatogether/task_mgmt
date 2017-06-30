package ipfs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/datatogether/task-mgmt/tasks"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

type TaskAdd struct {
	Url              string `json:"url"`              // url to resource to be added
	Checksum         string `json:"checksum"`         // optional checksum to check resp against
	IpfsApiServerUrl string `json:"ipfsApiServerUrl"` // url of IPFS api server
}

func NewTaskAdd() tasks.Taskable {
	return &TaskAdd{}
}

func (t *TaskAdd) Valid() error {
	if t.Url == "" {
		return fmt.Errorf("url param is required")
	}
	if t.IpfsApiServerUrl == "" {
		return fmt.Errorf("no ipfs server url provided")
	}
	return nil
}

func (t *TaskAdd) Do(pch chan tasks.Progress) {
	p := tasks.Progress{Step: 1, Steps: 4, Status: "fetching resource"}

	// 1. Get the Url
	pch <- p
	res, err := http.Get(t.Url)
	if err != nil {
		p.Error = fmt.Errorf("Error getting url: %s: %s", t.Url, err.Error())
		pch <- p
		return
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		p.Error = fmt.Errorf("Error reading resource fetch response: %s", err.Error())
		pch <- p
		return
	}
	if len(data) == 0 {
		p.Error = fmt.Errorf("fetch url %s returned empty response", t.Url)
		pch <- p
		return
	}
	// close immideately, next steps could take a while
	res.Body.Close()

	// 2. run checksum
	p.Status = "running checksum"
	p.Percent = 0.33
	p.Step++

	if t.Checksum != "" {
		pch <- p
		// TODO - run checksum
	}

	// 3. prepare upload
	p.Status = "preparing upload"
	p.Percent = 0.5
	p.Step++
	pch <- p

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	f, err := w.CreateFormFile("path", filepath.Base(t.Url))
	if err != nil {
		p.Error = fmt.Errorf("error creating form file: %s", err.Error())
		pch <- p
		return
	}

	// TODO - handle errors
	if _, err := f.Write(data); err != nil {
		p.Error = fmt.Errorf("error creating form file from response data: %s", err.Error())
		pch <- p
		return
	}

	if err := w.Close(); err != nil {
		p.Error = fmt.Errorf("error closing form file: %s", err.Error())
		pch <- p
		return
	}

	// add to IPFS
	p.Status = "adding file to IPFS node"
	p.Percent = 0.75
	p.Step++
	pch <- p

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/add", t.IpfsApiServerUrl), body)
	if err != nil {
		p.Error = fmt.Errorf("error creating request: %s", err.Error())
		pch <- p
		return
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	ipfsRes, err := http.DefaultClient.Do(req)
	if err != nil {
		p.Error = fmt.Errorf("error sending request: %s", err.Error())
		pch <- p
		return
	}
	defer ipfsRes.Body.Close()
	if ipfsRes.StatusCode != http.StatusOK {
		p.Error = fmt.Errorf("IPFS Server returned non-200 status code: %d", ipfsRes.StatusCode)
		pch <- p
		return
	}

	reply := struct {
		Reply string
		Hash  string
	}{}
	if err := json.NewDecoder(ipfsRes.Body).Decode(&reply); err != nil {
		p.Error = fmt.Errorf("error reading response body: %s", err.Error())
		pch <- p
		return
	}

	p.Percent = 1.0
	p.Done = true
	pch <- p
	return
}
