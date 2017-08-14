package ipfs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/datatogether/archive"
	"github.com/datatogether/sql_datastore"
	"github.com/datatogether/task-mgmt/tasks"
	"github.com/ipfs/go-datastore"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

type TaskAdd struct {
	Url              string              `json:"url"`              // url to resource to be added
	Checksum         string              `json:"checksum"`         // optional checksum to check resp against
	ipfsApiServerUrl string              `json:"ipfsApiServerUrl"` // url of IPFS api server
	store            datastore.Datastore // internal datastore pointer
}

func NewTaskAdd() tasks.Taskable {
	return &TaskAdd{
		ipfsApiServerUrl: IpfsApiServerUrl,
	}
}

// AddCollection task needs to talk to an underlying database
// it's expected that the task executor will call this method
// before calling Do
func (t *TaskAdd) SetDatastore(store datastore.Datastore) {
	if sqlds, ok := store.(*sql_datastore.Datastore); ok {
		// if we're passed an sql datastore
		// make sure our collection model is registered
		sqlds.Register(
			&archive.Url{},
		)
	}

	t.store = store
}

func (t *TaskAdd) Valid() error {
	if t.Url == "" {
		return fmt.Errorf("url param is required")
	}
	if t.ipfsApiServerUrl == "" {
		return fmt.Errorf("no ipfs server url provided, please configure the ipfs tasks package")
	}
	return nil
}

func (t *TaskAdd) Do(pch chan tasks.Progress) {
	p := tasks.Progress{Step: 1, Steps: 4, Status: "fetching resource"}

	u := &archive.Url{
		Url: t.Url,
	}

	if err := u.Read(t.store); err != nil && err != archive.ErrNotFound {
		p.Error = fmt.Errorf("Error reading url: %s", err.Error())
		pch <- p
		return
	}

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

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/add", t.ipfsApiServerUrl), body)
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
