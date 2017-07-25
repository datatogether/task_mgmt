package ipfs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/datatogether/warc"
	"github.com/pborman/uuid"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"
)

// Should be set by implementers
var IpfsApiServerUrl = ""

// GetUrl grabs a url, return
// currently a big 'ol work in progress, and will probably be moved into it's own
// package. for now the request bytes aren't to be trusted
func GetUrlBytes(urlstr string) (request, response []byte, err error) {
	req := &warc.Request{
		WARCRecordId:  uuid.New(),
		WARCDate:      time.Now(),
		ContentLength: 0,
		WARCTargetURI: urlstr,
	}

	buf := bytes.NewBuffer(nil)
	req.Write(buf)
	request = buf.Bytes()

	cli := http.Client{
		Timeout: time.Second * 20,
	}

	res, err := cli.Get(urlstr)
	if err != nil {
		return
	}
	// close immideately, next steps could take a while
	defer res.Body.Close()

	response, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	if len(response) == 0 {
		err = fmt.Errorf("Empty Response")
	}

	// TODO - generate response as a WARC record
	// resrec := &warc.Response{
	// 	WARCRecordId:  uuid.New(),
	// 	WARCDate:      time.Now(),
	// 	ContentLength: len(response),
	// 	ContentType:   res.Header.Get("Content-Type"),
	// }

	return
}

func WriteToIpfs(ipfsurl, filename string, data []byte) (hash string, err error) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	var (
		f       io.Writer
		ipfsReq *http.Request
		ipfsRes *http.Response
	)

	f, err = w.CreateFormFile("path", filename)
	if err != nil {
		err = fmt.Errorf("error creating form file: %s", err.Error())
		return
	}

	// TODO - handle errors
	if _, err = f.Write(data); err != nil {
		err = fmt.Errorf("error creating form file data: %s", err.Error())
		return
	}

	if err = w.Close(); err != nil {
		err = fmt.Errorf("error closing form file: %s", err.Error())
		return
	}

	// add to IPFS
	ipfsReq, err = http.NewRequest("POST", fmt.Sprintf("%s/add", ipfsurl), body)
	if err != nil {
		err = fmt.Errorf("error creating request: %s", err.Error())
		return
	}
	ipfsReq.Header.Set("Content-Type", w.FormDataContentType())

	ipfsRes, err = http.DefaultClient.Do(ipfsReq)
	if err != nil {
		err = fmt.Errorf("error sending request: %s", err.Error())
		return
	}
	defer ipfsRes.Body.Close()

	if ipfsRes.StatusCode != http.StatusOK {
		err = fmt.Errorf("IPFS Server returned non-200 status code: %d", ipfsRes.StatusCode)
		return
	}

	reply := struct {
		Reply string
		Hash  string
	}{}
	if err = json.NewDecoder(ipfsRes.Body).Decode(&reply); err != nil {
		err = fmt.Errorf("error reading response body: %s", err.Error())
		return
	}

	return reply.Hash, nil
}
