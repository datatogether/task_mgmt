package sciencebase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/datatogether/archive"
	"github.com/datatogether/cdxj"
	sb "github.com/datatogether/linked_data/sciencebase"
	"github.com/datatogether/sql_datastore"
	"github.com/datatogether/task-mgmt/taskdefs/ipfs"
	"github.com/datatogether/task-mgmt/tasks"
	"github.com/ipfs/go-datastore"
	"net/http"
	"path/filepath"
	"time"
)

var IpfsApiServerUrl = ""

const defaultCrawlDelay = time.Second / 2
const pageSize = 100

// AddCatalog injests a collection to IPFS,
// it iterates through each setting hashes on collection urls
// and, eventually, generates a cdxj index of the archive
type AddCatalog struct {
	// title for collection if no collection present
	CollectionTitle string
	// url that points to catalog
	Url string `json:"url"`
	// how many items deep to crawl at most, -1 == no max
	MaxDepth int
	// how many fetching goroutines to spin up. max 5
	Parallelism int
	// skip items that already have a hash value
	SkipArchived bool
	// how long to sleep between requests in seconds(inside of parallel routines)
	CrawDelay time.Duration
	// url of IPFS api server, should be set internally
	ipfsApiServerUrl string
	// internal datastore pointer
	store datastore.Datastore
}

func NewAddCatalog() tasks.Taskable {
	return &AddCatalog{
		CrawDelay:        defaultCrawlDelay,
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
	if t.Parallelism > 20 {
		t.Parallelism = 20
	}
	// super rough coercion to seconds if non-default crawldelay
	// is specified
	if t.CrawDelay != defaultCrawlDelay {
		if t.CrawDelay > 10 {
			t.CrawDelay = 10
		}
		t.CrawDelay = time.Second * t.CrawDelay
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

	parent := &sb.Item{}
	err = json.Unmarshal(body, cat)
	if err != nil {
		p.Error = fmt.Errorf("error parsing data catalog: %s", err.Error())
		pch <- p
		return
	}

	indexBuf := bytes.NewBuffer(nil)
	index := cdxj.NewWriter(indexBuf)

	// TODO - refactor done chan to report progress, possibly sending the number
	// of indexes *remaining* with each iteration

	if err := ArchiveCatalog(col, t.Url, t.MaxDepth, t.Parallelism); err != nil {

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

func ArchiveCatalog(col *archive.Collection, index *cdxj.Writer, rooturl string, maxDepth, parallelism int) error {
	ch := make(chan childItem)
	tracks := make([]chan childItem, parallelism)
	done := make(chan bool)

	for _, track := range tracks {
		track = make(chan childItem, 5)
		go func() {
			for child := range track {
				if err := ArchiveChild(col, index, child, ch); err != nil {
					fmt.Println("error archiving url:", err.Error())
					// TODO - collect errored urls, or flag as errored?
				}
				time.Sleep(t.CrawDelay)
			}
		}()
	}

	go func() {
		t := 0
		for child := range ch {

			if maxDepth == -1 || child.depth < maxDepth {
				track[t] <- child
				t++
				if t == len(tracks) {
					t = 0
				}
				stack++
			}

			stack--
			if stack == 0 {
				for _, track := range tracks {
					close(track)
				}
				close(ch)
				done <- true
			}
		}
	}()

	// add root item to kick off
	ch <- childItem{0, rootUrl}
	<-done
	return
}

type childItem struct {
	depth int
	url   string
}

func ArchiveChild(collection *archive.Collection, index *cdxj.Writer, child childItem, ch chan childItem) error {
	// archive
	hh, bh, err := ipfs.ArchiveUrl(store, ipfsApiUrl, child.url)
	if err != nil {
		return err
	}

	item := &sb.Item{}
	if err := json.NewDecoder(res.Body).Decode(item); err != nil {
		return err
	}

	if err := ArchiveFiles(col, index, item); err != nil {

	}

	if u := item.ChildrenJsonUrl(); item.HasChildren && u != "" {
		// Get URl, archive, add to collection
		res, err := http.Get(u)
		if err != nil {
			return err
		}
		// parse json
		defer res.Close()
		c := &sb.Catalog{}
		if err := json.NewDecoder(res.Body).Decode(c); err != nil {
			return err
		}

		// iterate through "items", forward through channel
		for _, item := range c.Items {
			if item.Link != nil {
				ch <- childItem{child.depth + 1, item.Link.JsonUrl(), nil}
			}
		}
		return nil
	}

	return nil
}

func ArchiveFiles(col *archive.Collection, index *cdxj.Writer, item *sb.Item) error {
	for _, f := range item.Files {
		// TODO - archive file

	}
	return nil
}
