package main

import (
	"io/ioutil"
	"testing"

	"github.com/spf13/afero"
)

func TestTheme(t *testing.T) {
	filename := "/index.html"
	fs := new(afero.MemMapFs)
	index := "this is the index"
	afero.WriteFile(fs, filename, []byte(index), 0644)
	fileInfo, err := fs.Stat(filename)
	if err != nil {
		t.Error(err)
	}

	s := NewTheme(fs)

	if f, timestamp, err := s.Open("/index.html"); err != nil {
		t.Error(err)
	} else if buf, err := ioutil.ReadAll(f); err != nil {
		t.Error(err)
	} else if string(buf) != index {
		t.Error("Read wrong value from index")
	} else if !timestamp.Equal(fileInfo.ModTime()) {
		t.Error("Timestamp compared wrong")
	}
}
