package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestTheme(t *testing.T) {
	s := NewTheme("testdata/theme")

	if f, timestamp, err := s.Open("/index.html"); err != nil {
		t.Error(err)
	} else if buf, err := ioutil.ReadAll(f); err != nil {
		t.Error(err)
	} else if string(buf) != index {
		t.Error("Read wrong value from index")
	} else if fi, err := os.Stat("testdata/theme/index.html"); err != nil {
		t.Error(err)
	} else if !timestamp.Equal(fi.ModTime()) {
		t.Error("Timestamp compared wrong")
	}

	if f, _, err := s.Open("/foo/bar/index.html"); err == nil {
		f.Close()
		t.Error("Path is ignored")
	}

	if f, _, err := s.Open("nofile"); err == nil {
		f.Close()
		t.Error("Opening non-existent file didn't return an error")
	}
}
