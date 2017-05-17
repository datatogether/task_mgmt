package main

import (
	"testing"
)

func TestReadSources(t *testing.T) {
	sources, err := ReadSources(appDB, "created DESC", 10, 0)
	if err != nil {
		t.Error(err)
		return
	}
	if len(sources) == 0 {
		t.Error("no sources returned")
	}
}
