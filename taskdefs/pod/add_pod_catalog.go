package pod

import (
	"encoding/json"
	"fmt"
	"github.com/datatogether/archive"
	// "github.com/datatogether/cdxj"
	"github.com/datatogether/linked_data/pod"
	"github.com/datatogether/sql_datastore"
	"github.com/datatogether/task-mgmt/tasks"
	"github.com/ipfs/go-datastore"
)

const pageSize = 100

// AddCatalog injests a a collection to IPFS,
// it iterates through each setting hashes on collection urls
// and, eventually, generates a cdxj index of the archive
type AddCatalog struct {
	CollectionTitle string
	Url             string `json:"url"` // url to resource to be added
	// paginate into dataset list, zero is no pagination / offset
	Limit  int
	Offset int
	// ipfsApiServerUrl string              `json:"ipfsApiServerUrl"` // url of IPFS api server
	store datastore.Datastore // internal datastore pointer
}

func NewAddCatalog() tasks.Taskable {
	return &AddCatalog{
	// ipfsApiServerUrl: IpfsApiServerUrl,
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
	// if t.ipfsApiServerUrl == "" {
	// 	return fmt.Errorf("no ipfs server url provided, please configure the ipfs tasks package")
	// }
	return nil
}

func (t *AddCatalog) Do(pch chan tasks.Progress) {
	p := tasks.Progress{Step: 1, Steps: 4, Status: "loading collection"}
	// 1. Get the Collection & Item Count
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

	pctAdd := 1.0 / float32(len(cat.Dataset))
	// indexBuf := bytes.NewBuffer(nil)
	// index := cdxj.NewWriter(indexBuf)

	// TODO - parallelize a lil bit
	for i, ds := range cat.Dataset {
		p.Status = fmt.Sprintf("archiving dataset %d/%d", i, len(cat.Dataset))
		p.Percent += pctAdd
		pch <- p

		if t.Offset > 0 && i < t.Offset {
			continue
		}
		if t.Limit > 0 && i > t.Limit {
			break
		}

		for j, dist := range ds.Distribution {
			if dist.DownloadURL != "" {
				// headerHash, bodyHash, err := ArchiveUrl(t.ipfsApiServerUrl, &item.Url)
				u := &archive.Url{Url: dist.DownloadURL}

				_, _, err := u.Get(t.store)
				if err != nil {
					fmt.Println("error getting url:", err.Error())
					// p.Error = err
					// pch <- p
					continue
				}

				// TODO - hash url & add metadata record

				// TODO - demo content for now, this is going to need lots of refinement
				// indexRec := &cdxj.Record{
				//   Uri:        urlstr,
				//   Timestamp:  start,
				//   RecordType: "", // TODO set record type?
				//   JSON: map[string]interface{}{
				//     "locator": fmt.Sprintf("urn:ipfs/%s/%s", headerHash, bodyHash),
				//   },
				// }

				// if err := index.Write(indexRec); err != nil {
				//   p.Error = fmt.Errorf("Error writing %s body to ipfs: %s", filepath.Base(urlstr), err.Error())
				//   pch <- p
				//   return
				// }

				// if err := item.Save(t.store); err != nil {
				//   p.Error = fmt.Errorf("Error saving item: %s: %s", item.Id, err.Error())
				//   pch <- p
				//   return
				// }
				fmt.Println(u.Url)

				if err = collection.SaveItems(t.store, []*archive.CollectionItem{
					&archive.CollectionItem{Url: *u},
				}); err != nil {
					p.Error = fmt.Errorf("error saving dataset %d dist %d to collection: %s", i, j, err.Error())
					pch <- p
					return
				}
			}
		}
	}

	p.Step++
	p.Status = "writing index to IPFS"
	pch <- p
	// close & sort the index
	// if err := index.Close(); err != nil {
	// 	p.Error = fmt.Errorf("Error closing index %s", err.Error())
	// 	pch <- p
	// 	return
	// }
	// indexhash, err := WriteToIpfs(t.ipfsApiServerUrl, fmt.Sprintf("%s.cdxj", collection.Id), indexBuf.Bytes())
	// if err != nil {
	// 	p.Error = fmt.Errorf("Error writing index to ipfs: %s", err.Error())
	// 	pch <- p
	// 	return
	// }
	// fmt.Printf("collection %s index hash: %s\n", collection.Id, indexhash)

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
