package main

import (
	"database/sql"
)

func ReadSources(db sqlQueryable, orderby string, limit, offset int) ([]*Source, error) {
	rows, err := db.Query(qSourcesBySourceUrl, "created DESC", limit, offset)
	if err != nil {
		return nil, err
	}

	return unmarshalSources(rows, limit)
}

func unmarshalSources(rows *sql.Rows, limit int) ([]*Source, error) {
	defer rows.Close()
	sources := make([]*Source, limit)
	i := 0
	for rows.Next() {
		t := &Source{}
		if err := t.UnmarshalSQL(rows); err != nil {
			return nil, err
		}
		i++
	}

	sources = sources[:i]
	return sources, nil
}
