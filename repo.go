package main

import (
	"database/sql"
	"github.com/pborman/uuid"
	"time"
)

// a Repo represents a code repository that can be applied to sources
type Repo struct {
	// uuid identifier for task
	Id string `json:"id"`
	// created date rounded to secounds
	Created time.Time `json:"created"`
	// updated date rounded to secounds
	Updated time.Time `json:"updated"`
	// location of code. This should be a github url
	Url string
	// Branch to read From
	// For now this should always just be "master", url should match any
	// branching scheme
	Branch string `json:"branch"`
	// version control commit to execute code from
	// currently should only look at the master branch
	LatestCommit string `json:"codeCommit"`
}

func (r *Repo) Read(db sqlQueryable) error {
	return r.UnmarshalSQL(db.QueryRow(qRepoReadById, r.Id))
}

func (r *Repo) Save(db sqlQueryExecable) error {
	prev := &Repo{Id: r.Id}
	if err := prev.Read(db); err == ErrNotFound {
		r.Id = uuid.New()
		r.Created = time.Now().Round(time.Second).In(time.UTC)
		r.Updated = r.Created
		_, err := db.Exec(qRepoInsert, r.sqlArgs()...)
		return err
	} else if err != nil {
		return err
	} else {
		r.Updated = time.Now().Round(time.Second).In(time.UTC)
		_, err := db.Exec(qRepoUpdate, r.sqlArgs()...)
		return err
	}

	return nil
}

func (r *Repo) Delete(db sqlQueryExecable) error {
	_, err := db.Exec(qRepoDelete, r.Id)
	return err
}

func (t *Repo) UnmarshalSQL(row sqlScannable) error {
	var (
		id, url, branch, commit string
		created, updated        time.Time
	)

	err := row.Scan(&id, &created, &updated, &url, &branch, &commit)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}

	*t = Repo{
		Id:           id,
		Created:      created,
		Updated:      updated,
		Url:          url,
		Branch:       branch,
		LatestCommit: commit,
	}

	return nil
}

func (t *Repo) sqlArgs() []interface{} {
	return []interface{}{
		t.Id,
		t.Created,
		t.Updated,
		t.Url,
		t.Branch,
		t.LatestCommit,
	}
}
