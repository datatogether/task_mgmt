package main

import (
	"database/sql"
)

func ReadTasks(db sqlQueryable, orderby string, limit, offset int) ([]*Task, error) {
	rows, err := db.Query(qTasksBySourceUrl, "created DESC", limit, offset)
	if err != nil {
		return nil, err
	}

	return unmarshalTasks(rows, limit)
}

func unmarshalTasks(rows *sql.Rows, limit int) ([]*Task, error) {
	defer rows.Close()
	tasks := make([]*Task, limit)
	i := 0
	for rows.Next() {
		t := &Task{}
		if err := t.UnmarshalSQL(rows); err != nil {
			return nil, err
		}
		i++
	}

	tasks = tasks[:i]
	return tasks, nil
}
