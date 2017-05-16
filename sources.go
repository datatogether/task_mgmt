package main

import (
	"database/sql"
)

func ReadSources(db sqlQueryable, orderby string, limit, offset int) ([]*Source, error) {
	rows, err := db.Query(qSourcesBySourceUrl, orderby, limit, offset)
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
		s := &Source{}
		if err := s.UnmarshalSQL(rows); err != nil {
			return nil, err
		}

		sources[i] = s
		i++
	}

	sources = sources[:i]
	return sources, nil
}
