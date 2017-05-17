package main

import (
	"fmt"
	"testing"
)

func TestTaskStorage(t *testing.T) {
	defer resetTestData(appDB, "tasks")

	task := &Task{
		RepoUrl:        "test_repo_url",
		RepoCommit:     "test_commit",
		SourceChecksum: "test_source_checksum",
		SourceUrl:      "test_source_url",
	}
	if err := task.Save(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	if err := task.Run(appDB); err != nil {
		t.Error(err.Error())
		return
	}
	if task.Request == nil {
		t.Errorf("task run didn't set request datestamp")
		return
	}

	if err := task.Cancel(appDB); err != nil {
		t.Error(err.Error())
		return
	}
	if task.Fail == nil {
		t.Errorf("task cancel didn't set fail datestamp")
		return
	}

	if err := task.Run(appDB); err != nil {
		t.Error(err.Error())
		return
	}
	if task.Request == nil {
		t.Errorf("task run didn't set request datestamp")
		return
	}

	if err := task.Errored(appDB, "test failure message"); err != nil {
		t.Error(err.Error())
		return
	}
	if task.Fail == nil {
		t.Errorf("task run didn't set fail datestamp")
		return
	}

	if err := task.Run(appDB); err != nil {
		t.Error(err.Error())
		return
	}
	if task.Request == nil {
		t.Errorf("task run didn't set request datestamp")
		return
	}

	if err := task.Succeeded(appDB, "test_success_url", "test_success_hash"); err != nil {
		t.Error(err.Error())
		return
	}
	if task.Success == nil {
		t.Errorf("task run didn't set success datestamp")
		return
	}

	task2 := &Task{Id: task.Id}
	if err := task2.Read(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	if err := CompareTasks(task, task2); err != nil {
		t.Error(err)
		return
	}

	if err := task.Delete(appDB); err != nil {
		t.Error(err.Error())
		return
	}
}

func CompareTasks(a, b *Task) error {
	if a.Id != b.Id {
		return fmt.Errorf("Id mismatch: %s != %s", a.Id, b.Id)
	}
	if !a.Created.Equal(b.Created) {
		return fmt.Errorf("Created mismatch: %s != %s", a.Created, b.Created)
	}
	if !a.Updated.Equal(b.Updated) {
		return fmt.Errorf("Updated mismatch: %s != %s", a.Updated, b.Updated)
	}
	if a.RepoUrl != b.RepoUrl {
		return fmt.Errorf("RepoUrl mismatch %s != %s", a.RepoUrl, b.RepoUrl)
	}
	if a.RepoCommit != b.RepoCommit {
		return fmt.Errorf("RepoCommit mismatch %s != %s", a.RepoCommit, b.RepoCommit)
	}
	if a.SourceUrl != b.SourceUrl {
		return fmt.Errorf("SourceUrl mismatch %s != %s", a.SourceUrl, b.SourceUrl)
	}
	if a.SourceChecksum != b.SourceChecksum {
		return fmt.Errorf("SourceChecksum mismatch %s != %s", a.SourceChecksum, b.SourceChecksum)
	}
	if a.ResultUrl != b.ResultUrl {
		return fmt.Errorf("ResultUrl mismatch %s != %s", a.ResultUrl, b.ResultUrl)
	}
	if a.ResultHash != b.ResultHash {
		return fmt.Errorf("ResultHash mismatch %s != %s", a.ResultHash, b.ResultHash)
	}
	if a.Message != b.Message {
		return fmt.Errorf("Message mismatch %s != %s", a.Message, b.Message)
	}
	return nil
}
