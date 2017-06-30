package tasks

import (
	"database/sql"
	"fmt"
	"github.com/datatogether/sql_datastore"
	"github.com/datatogether/sqlutil"
	"github.com/ipfs/go-datastore"
	"github.com/pborman/uuid"
	"time"
)

// Task represents the storable state of a task
// TODO - this needs heavy generalization
type Task struct {
	// uuid identifier for task
	Id string `json:"id"`
	// created date rounded to secounds
	Created time.Time `json:"created"`
	// updated date rounded to secounds
	Updated time.Time `json:"updated"`
	// human-readable title for the task
	Title string `json:"name"`
	// timstamp for when request was submitted for completion
	// nil if request hasn't been sent
	Request *time.Time `json:"request"`
	// timstamp for when request succeeded
	// nil if task hasn't sicceeded
	Success *time.Time `json:"success"`
	// timstamp for when request failed
	// nil if task hasn't failed
	Fail *time.Time `json:"fail"`
	// url to where the code to execute lives
	// example: https://github.com/ipfs/ipfs-wiki/mirror
	RepoUrl string `json:"repoCommit"`
	// version control repoCommit to execute code from
	RepoCommit string `json:"repoCommit"`
	// url this code is to run against
	SourceUrl string `json:"sourceUrl"`
	// checksum of source resource
	SourceChecksum string `json:"sourceChecksum"`
	// url of output
	ResultUrl string `json:"resultUrl"`
	// multihash of output
	ResultHash string `json:"resultHash"`
	// any message associated with this task (failure, info, etc.)
	Message string `json:"message"`
}

func (t Task) DatastoreType() string {
	return "Task"
}

func (t Task) GetId() string {
	return t.Id
}

func (t Task) Key() datastore.Key {
	return datastore.NewKey(fmt.Sprintf("%s:%s", t.DatastoreType(), t.GetId()))
}

func (t *Task) StatusString() string {
	if t.Request == nil {
		return "ready"
	} else if t.Success != nil {
		return "finished"
	} else if t.Fail != nil {
		return "failed"
	} else {
		return "running"
	}
}

func (t *Task) NextActionUrl() (url string, err error) {
	switch t.StatusString() {
	case "ready":
		return fmt.Sprintf("/tasks/run/%s", t.Id), nil
	case "running":
		return fmt.Sprintf("/tasks/cancel/%s", t.Id), nil
	case "failed":
		return fmt.Sprintf("/tasks/run/%s", t.Id), nil
	default:
		return "", fmt.Errorf("no next action")
	}
}

func (t *Task) NextActionTitle() (title string, err error) {
	switch t.StatusString() {
	case "ready":
		return "run", nil
	case "running":
		return "cancel", nil
	case "failed":
		return "re-run", nil
	default:
		return "", fmt.Errorf("no next action")
	}
}

// Run marks the task as "running", this should probably not be called run
// TODO - naming refactor / negotiate relationship between task que & task model
func (t *Task) Run(store datastore.Datastore) error {
	now := time.Now()
	t.Request = &now
	t.Fail = nil
	t.Success = nil

	// if err := SendTaskRequestEmail(t); err != nil {
	// 	return err
	// }
	return t.Save(store)
}

func (t *Task) Cancel(store datastore.Datastore) error {
	now := time.Now()
	t.Fail = &now
	t.Success = nil
	t.Message = "Task Cancelled"

	// if err := SendTaskCancelEmail(t); err != nil {
	// 	return err
	// }

	return t.Save(store)
}

func (t *Task) Errored(store datastore.Datastore, message string) error {
	now := time.Now()
	t.Fail = &now
	t.Message = message
	return t.Save(store)
}

func (t *Task) Succeeded(store datastore.Datastore, url, hash string) error {
	now := time.Now()
	t.Success = &now
	t.ResultUrl = url
	t.ResultHash = hash
	t.Message = ""
	return t.Save(store)
}

func (t *Task) Read(store datastore.Datastore) error {
	if t.Id == "" {
		return datastore.ErrNotFound
	}

	ti, err := store.Get(t.Key())
	if err != nil {
		return err
	}

	got, ok := ti.(*Task)
	if !ok {
		return fmt.Errorf("Invalid Response")
	}
	*t = *got
	return nil
}

func (t *Task) Save(store datastore.Datastore) (err error) {
	var exists bool
	exists, err = store.Has(t.Key())
	if err != nil {
		return err
	}

	if !exists {
		t.Id = uuid.New()
		t.Created = time.Now().Round(time.Second).In(time.UTC)
		t.Updated = t.Created
	} else {
		t.Updated = time.Now().Round(time.Second).In(time.UTC)
	}

	return store.Put(t.Key(), t)
}

func (t *Task) Delete(store datastore.Datastore) error {
	return store.Delete(t.Key())
}

func (t *Task) NewSQLModel(id string) sql_datastore.Model {
	return &Task{Id: id}
}

func (t *Task) SQLQuery(cmd sql_datastore.Cmd) string {
	switch cmd {
	case sql_datastore.CmdCreateTable:
		return qTaskCreateTable
	case sql_datastore.CmdExistsOne:
		return qTaskExists
	case sql_datastore.CmdSelectOne:
		return qTaskReadById
	case sql_datastore.CmdInsertOne:
		return qTaskInsert
	case sql_datastore.CmdUpdateOne:
		return qTaskUpdate
	case sql_datastore.CmdDeleteOne:
		return qTaskDelete
	case sql_datastore.CmdList:
		return qTasks
	default:
		return ""
	}
}

func (t *Task) UnmarshalSQL(row sqlutil.Scannable) error {
	var (
		id, title, repoUrl, repoCommit, source, sourceChecksum, message, result, resultHash string
		created, updated                                                                    time.Time
		request, success, fail                                                              *time.Time
	)
	err := row.Scan(
		&id, &created, &updated, &title, &request, &success, &fail,
		&repoUrl, &repoCommit, &source, &sourceChecksum, &result, &resultHash, &message,
	)
	if err == sql.ErrNoRows {
		return datastore.ErrNotFound
	}

	*t = Task{
		Id:             id,
		Created:        created,
		Updated:        updated,
		Title:          title,
		Request:        request,
		Success:        success,
		Fail:           fail,
		RepoUrl:        repoUrl,
		RepoCommit:     repoCommit,
		SourceUrl:      source,
		SourceChecksum: sourceChecksum,
		ResultUrl:      result,
		ResultHash:     resultHash,
	}

	return nil
}

func (t *Task) SQLParams(cmd sql_datastore.Cmd) []interface{} {
	switch cmd {
	case sql_datastore.CmdSelectOne, sql_datastore.CmdExistsOne, sql_datastore.CmdDeleteOne:
		return []interface{}{t.Id}
	case sql_datastore.CmdList:
		return []interface{}{}
	default:
		return []interface{}{
			t.Id,
			t.Created,
			t.Updated,
			t.Title,
			t.Request,
			t.Success,
			t.Fail,
			t.RepoUrl,
			t.RepoCommit,
			t.SourceUrl,
			t.SourceChecksum,
			t.ResultUrl,
			t.ResultHash,
			t.Message,
		}
	}
}
