package main

import (
	"database/sql"
	"encoding/json"
	"github.com/pborman/uuid"
	"time"
)

// Source is an http origin of data
type Source struct {
	Id string
	// created date rounded to secounds
	Created time.Time `json:"created"`
	// updated date rounded to secounds
	Updated time.Time `json:"updated"`
	// Human-readable title for source
	Title    string
	Url      string
	Checksum string
	Meta     map[string]interface{}
}

func (s *Source) Read(db sqlQueryable) error {
	return s.UnmarshalSQL(db.QueryRow(qSourceReadById, s.Id))
}

func (s *Source) Save(db sqlQueryExecable) error {
	prev := &Source{Id: s.Id}
	if err := prev.Read(db); err != ErrNotFound {
		s.Id = uuid.New()
		s.Created = time.Now().In(time.UTC)
		s.Updated = s.Created
		_, err := db.Exec(qSourceInsert, s.sqlArgs()...)
		return err
	} else if err != nil {
		return err
	} else {
		s.Updated = time.Now().In(time.UTC)
		_, err := db.Exec(qSourceUpdate, s.sqlArgs()...)
		return err
	}
	return nil
}

func (s *Source) Delete(db sqlQueryExecable) error {
	_, err := db.Exec(qSourceDelete, s.Id)
	return err
}

func (s *Source) UnmarshalSQL(row sqlScannable) error {
	var (
		id, url, checksum, title string
		created, updated         time.Time
		meta                     []byte
	)

	if err := row.Scan(&id, &created, &updated, &title, &url, &checksum, &meta); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	*s = Source{
		Id:       id,
		Created:  created,
		Updated:  updated,
		Title:    title,
		Url:      url,
		Checksum: checksum,
	}

	if meta != nil {
		s.Meta = map[string]interface{}{}
		if err := json.Unmarshal(meta, &s.Meta); err != nil {
			return err
		}
	}

	return nil
}

func (s *Source) sqlArgs() []interface{} {
	var (
		meta []byte
		err  error
	)
	if s.Meta != nil {
		if meta, err = json.Marshal(s.Meta); err != nil {
			meta = nil
		}
	}

	return []interface{}{
		s.Id,
		s.Created,
		s.Updated,
		s.Title,
		s.Url,
		s.Checksum,
		meta,
	}
}
