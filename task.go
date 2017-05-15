package main

import (
	"database/sql"
	"github.com/pborman/uuid"
	"time"
)

// a Task represents the state of work to be done
type Task struct {
	// uuid identifier for task
	Id string `json:"id"`
	// created date rounded to secounds
	Created time.Time `json:"created"`
	// updated date rounded to secounds
	Updated time.Time `json:"updated"`
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

func (t *Task) Requested(db *sql.DB) error {
	now := time.Now()
	t.Request = &now
	return t.Save(db)
}

func (t *Task) Errored(db *sql.DB, message string) error {
	now := time.Now()
	t.Fail = &now
	t.Message = message
	return t.Save(db)
}

func (t *Task) Succeeded(db *sql.DB, url, hash string) error {
	now := time.Now()
	t.Success = &now
	t.ResultUrl = url
	t.ResultHash = hash
	return t.Save(db)
}

func (t *Task) Read(db sqlQueryable) error {
	return t.UnmarshalSQL(db.QueryRow(qTaskReadById, t.Id))
}

func (t *Task) Save(db sqlQueryExecable) error {
	prev := &Task{Id: t.Id}
	if err := prev.Read(db); err == ErrNotFound {
		t.Id = uuid.New()
		t.Created = time.Now().Round(time.Second).In(time.UTC)
		t.Updated = t.Created
		_, err := db.Exec(qTaskInsert, t.sqlArgs()...)
		return err
	} else if err != nil {
		return err
	} else {
		t.Updated = time.Now().Round(time.Second).In(time.UTC)
		_, err := db.Exec(qTaskUpdate, t.sqlArgs()...)
		return err
	}

	return nil
}

func (t *Task) Delete(db sqlQueryExecable) error {
	_, err := db.Exec(qTaskDelete, t.Id)
	return err
}

func (t *Task) UnmarshalSQL(row sqlScannable) error {
	var (
		id, repoUrl, repoCommit, source, sourceChecksum, message, result, resultHash string
		created, updated                                                             time.Time
		request, success, fail                                                       *time.Time
	)
	err := row.Scan(
		&id, &created, &updated, &request, &success, &fail,
		&repoUrl, &repoCommit, &source, &sourceChecksum, &result, &resultHash, &message,
	)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}

	*t = Task{
		Id:             id,
		Created:        created,
		Updated:        updated,
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

func (t *Task) sqlArgs() []interface{} {
	return []interface{}{
		t.Id,
		t.Created,
		t.Updated,
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
