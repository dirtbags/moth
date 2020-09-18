package main

import (
	"io/ioutil"
	"testing"

	"github.com/spf13/afero"
)

func NewTestTheme() *Theme {
	return NewTheme(new(afero.MemMapFs))
}

func TestTheme(t *testing.T) {
	s := NewTestTheme()

	filename := "/index.html"
	index := "this is the index"
	afero.WriteFile(s.Fs, filename, []byte(index), 0644)
	fileInfo, err := s.Fs.Stat(filename)
	if err != nil {
		t.Error(err)
	}

	if f, timestamp, err := s.Open("/index.html"); err != nil {
		t.Error(err)
	} else if buf, err := ioutil.ReadAll(f); err != nil {
		t.Error(err)
	} else if string(buf) != index {
		t.Error("Read wrong value from index")
	} else if !timestamp.Equal(fileInfo.ModTime()) {
		t.Error("Timestamp compared wrong")
	}

	if f, _, err := s.Open("nofile"); err == nil {
		f.Close()
		t.Error("Opening non-existent file didn't return an error")
	}
}
