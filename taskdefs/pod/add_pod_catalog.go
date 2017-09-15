package pod

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/datatogether/archive"
	"github.com/datatogether/cdxj"
	"github.com/datatogether/linked_data/pod"
	"github.com/datatogether/sql_datastore"
	"github.com/datatogether/task-mgmt/taskdefs/ipfs"
	"github.com/datatogether/task-mgmt/tasks"
	"github.com/ipfs/go-datastore"
	"path/filepath"
	"time"
)

var IpfsApiServerUrl = ""

const pageSize = 100

// AddCatalog injests a a collection to IPFS,
// it iterates through each setting hashes on collection urls
// and, eventually, generates a cdxj index of the archive
type AddCatalog struct {
	// title for collection if no collection present
	CollectionTitle string
	// url that points to catalog
	Url string `json:"url"`
	// paginate into dataset list, zero is no pagination / offset
	Limit int
	// offset to start archiving at
	Offset int
	// how many fetching goroutines to spin up. max 5
	Parallelism int
	// how long to sleep between requests (inside of parallel routines)
	CrawDelay time.Duration
	// url of IPFS api server, should be set internally
	ipfsApiServerUrl string
	// internal datastore pointer
	store datastore.Datastore
}

func NewAddCatalog() tasks.Taskable {
	return &AddCatalog{
		CrawDelay:        time.Second / 2,
		Parallelism:      2,
		ipfsApiServerUrl: IpfsApiServerUrl,
	}
}

// AddCatalog task needs to talk to an underlying database
// it's expected that the task executor will call this method
// before calling Do
func (t *AddCatalog) SetDatastore(store datastore.Datastore) {
	if sqlds, ok := store.(*sql_datastore.Datastore); ok {
		// if we're passed an sql datastore
		// make sure our collection model is registered
		sqlds.Register(
			&archive.Url{},
			&archive.Collection{},
			&archive.CollectionItem{},
		)
	}

	t.store = store
}

func (t *AddCatalog) Valid() error {
	if t.Url == "" {
		return fmt.Errorf("url is required")
	}
	if t.Parallelism > 5 {
		t.Parallelism = 5
	}
	if t.ipfsApiServerUrl == "" {
		return fmt.Errorf("no ipfs server url provided, please configure the ipfs tasks package")
	}
	return nil
}

func (t *AddCatalog) Do(pch chan tasks.Progress) {
	p := tasks.Progress{Step: 1, Steps: 4, Status: "loading collection"}
	pch <- p

	collection := &archive.Collection{Url: t.Url}
	if err := collection.Read(t.store); err != nil && err != archive.ErrNotFound {
		p.Error = fmt.Errorf("Error reading collection: %s", err.Error())
		pch <- p
		return
	} else if err == archive.ErrNotFound {
		collection.Title = t.CollectionTitle
		if err = collection.Save(t.store); err != nil {
			p.Error = fmt.Errorf("Error saving colleciton: %s", err.Error())
			pch <- p
			return
		}
	}
	p.Step++

	u := archive.Url{Url: t.Url}
	body, _, err := u.Get(t.store)
	if err != nil {
		p.Error = fmt.Errorf("error getting url: %s", err.Error())
		pch <- p
		return
	}

	cat := &pod.Catalog{}
	err = json.Unmarshal(body, cat)
	if err != nil {
		p.Error = fmt.Errorf("error parsing data catalog: %s", err.Error())
		pch <- p
		return
	}

	if t.Limit == 0 {
		t.Limit = len(cat.Dataset)
	}

	if t.Offset > len(cat.Dataset) {
		p.Error = fmt.Errorf("offset of %d greater than %d datasets", len(cat.Dataset), t.Offset)
		pch <- p
		return
	}

	// t.Limit := t.Limit + t.Offset

	// pctAdd := 1.0 / float32(len(cat.Dataset))
	indexBuf := bytes.NewBuffer(nil)
	index := cdxj.NewWriter(indexBuf)

	// TODO - refactor done chan to report progress, possibly sending the number
	// of indexes *remaining* with each iteration
	archiveIndexes := func(cat *pod.Catalog, start, stop int, done chan int) {
		for i := start; i <= stop; i++ {
			ds := cat.Dataset[i]

			for j, dist := range ds.Distribution {
				if dist.DownloadURL != "" {
					u := &archive.Url{Url: dist.DownloadURL}

					// _, _, err := u.Get(t.store)
					headerHash, bodyHash, err := ipfs.ArchiveUrl(t.store, t.ipfsApiServerUrl, u)
					if err != nil {
						fmt.Println("error archiving u url:", err.Error())
						// p.Error = err
						// pch <- p
						continue
					}

					// write metadata
					go func() {
						data, err := json.Marshal(ds)
						if err != nil {
							fmt.Println("error marshaling dataset to json:", err.Error())
							return
						}

						meta := map[string]interface{}{}
						if err := json.Unmarshal(data, &meta); err != nil {
							fmt.Println("error unmarshaling dataset to generic metadata:", err.Error())
							return
						}

						md := &archive.Metadata{
							Subject: bodyHash,
							Meta:    meta,
						}

						if err := md.Write(t.store); err != nil {
							fmt.Println("error writing metadata to store:", err.Error())
							return
						}
					}()

					// TODO - demo content for now, this is going to need lots of refinement
					indexRec := &cdxj.Record{
						Uri:        u.Url,
						Timestamp:  *u.LastGet,
						RecordType: "", // TODO set record type?
						JSON: map[string]interface{}{
							"locator": fmt.Sprintf("urn:ipfs/%s/%s", headerHash, bodyHash),
						},
					}

					if err := index.Write(indexRec); err != nil {
						p.Error = fmt.Errorf("Error writing %s body to ipfs: %s", filepath.Base(u.Url), err.Error())
						pch <- p
						return
					}

					if err = collection.SaveItems(t.store, []*archive.CollectionItem{
						&archive.CollectionItem{Url: *u},
					}); err != nil {
						p.Error = fmt.Errorf("error saving dataset %d dist %d to collection: %s", i, j, err.Error())
						pch <- p
						return
					}

					time.Sleep(t.CrawDelay)
				}
			}
		}
		done <- 1
	}

	c := make(chan int, t.Parallelism)
	sectionSize := t.Limit / t.Parallelism
	for i := 0; i < t.Parallelism; i++ {
		start := t.Offset + (sectionSize * i)
		stop := t.Offset + (sectionSize * (i + 1)) - 1
		go archiveIndexes(cat, start, stop, c)
	}
	// Drain the channel.
	for i := 0; i < t.Parallelism; i++ {
		<-c // wait for one task to complete
	}

	p.Step++
	p.Status = "writing index to IPFS"
	pch <- p
	// close & sort the index
	if err := index.Close(); err != nil {
		p.Error = fmt.Errorf("Error closing index %s", err.Error())
		pch <- p
		return
	}
	indexhash, err := ipfs.WriteToIpfs(t.ipfsApiServerUrl, fmt.Sprintf("%s.cdxj", collection.Id), indexBuf.Bytes())
	if err != nil {
		p.Error = fmt.Errorf("Error writing index to ipfs: %s", err.Error())
		pch <- p
		return
	}
	fmt.Printf("collection %s index hash: %s\n", collection.Id, indexhash)

	p.Step++
	p.Status = "saving collection results"
	pch <- p
	if err := collection.Save(t.store); err != nil {
		p.Error = fmt.Errorf("Error saving collection: %s", err.Error())
		pch <- p
		return
	}

	p.Percent = 1.0
	p.Done = true
	pch <- p
	return
}
