package main

import (
	"github.com/spf13/afero"
	"io/ioutil"
	"testing"
)

func TestTheme(t *testing.T) {
	fs := new(afero.MemMapFs)
	index := "this is the index"
	afero.WriteFile(fs, "/index.html", []byte(index), 0644)

	s := NewTheme(fs)

	if f, err := s.Open("/index.html"); err != nil {
		t.Error(err)
	} else if buf, err := ioutil.ReadAll(f); err != nil {
		t.Error(err)
	} else if string(buf) != index {
		t.Error("Read wrong value from index")
	}
}
