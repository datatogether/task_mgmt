package main

import (
	"testing"
)

func TestReadTasks(t *testing.T) {
	tasks, err := ReadTasks(appDB, "created DESC", 10, 0)
	if err != nil {
		t.Error(err)
		return
	}
	if len(tasks) == 0 {
		t.Error("no tasks returned")
	}
}

func TestGenerateAvailableTasks(t *testing.T) {
	defer resetTestData(appDB, "tasks")
	tasks, err := GenerateAvailableTasks(appDB)
	if err != nil {
		t.Error(err)
		return
	}
	if len(tasks) == 0 {
		t.Error("no tasks returned")
	}
}
