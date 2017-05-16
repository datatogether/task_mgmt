package main

import (
	"database/sql"
	"fmt"
)

func ReadTasks(db sqlQueryable, orderby string, limit, offset int) ([]*Task, error) {
	rows, err := db.Query(qTasksBySourceUrl, orderby, limit, offset)
	if err != nil {
		return nil, err
	}

	return unmarshalTasks(rows, limit)
}

func GenerateAvailableTasks(db sqlQueryExecable) ([]*Task, error) {
	row, err := db.Query(qAvailableTasks)
	if err != nil {
		return nil, err
	}

	tasks := []*Task{}
	for row.Next() {
		var (
			repoUrl, repoCommit, sourceTitle, sourceUrl, sourceChecksum string
		)
		if err := row.Scan(&repoUrl, &repoCommit, &sourceTitle, &sourceUrl, &sourceChecksum); err != nil {
			return nil, err
		}

		t := &Task{
			Title:          fmt.Sprintf("injest %s to ipfs", sourceTitle),
			RepoUrl:        repoUrl,
			RepoCommit:     repoCommit,
			SourceUrl:      sourceUrl,
			SourceChecksum: sourceChecksum,
		}

		if err := t.Save(db); err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
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
		tasks[i] = t
		i++
	}

	tasks = tasks[:i]
	return tasks, nil
}
