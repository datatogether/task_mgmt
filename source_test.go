package main

import (
	"fmt"
	"testing"
)

func TestSourceStorage(t *testing.T) {
	defer resetTestData(appDB, "sources", "repo_sources")

	s := &Source{
		Title:    "test source title",
		Url:      "test_source_url",
		Checksum: "test_source_checksum",
		Meta: map[string]interface{}{
			"key": "value",
		},
	}
	if err := s.Save(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	s2 := &Source{Id: s.Id}
	if err := s2.Read(appDB); err != nil {
		t.Error(err.Error())
		return
	}

	if err := CompareSources(s, s2); err != nil {
		t.Error(err)
		return
	}

	if err := s.Delete(appDB); err != nil {
		t.Error(err.Error())
		return
	}
}

func CompareSources(a, b *Source) error {
	if a.Id != b.Id {
		return fmt.Errorf("Id mismatch: %s != %s", a.Id, b.Id)
	}
	if !a.Created.Equal(b.Created) {
		return fmt.Errorf("Created mismatch: %s != %s", a.Created, b.Created)
	}
	if !a.Updated.Equal(b.Updated) {
		return fmt.Errorf("Updated mismatch: %s != %s", a.Updated, b.Updated)
	}
	if a.Title != b.Title {
		return fmt.Errorf("Title mismatch: %s != %s", a.Title, b.Title)
	}
	if a.Url != b.Url {
		return fmt.Errorf("Url mismatch: %s != %s", a.Url, b.Url)
	}
	if a.Checksum != b.Checksum {
		return fmt.Errorf("Checksum mismatch: %s != %s", a.Checksum, b.Checksum)
	}
	if a.Meta == nil && b.Meta != nil || a.Meta != nil && b.Meta == nil {
		return fmt.Errorf("Meta mismatch: %s != %s", a.Meta, b.Meta)
	}

	return nil
}
