package main

import (
	"github.com/spf13/afero"
	"os"
	"testing"
)

func TestState(t *testing.T) {
	fs := new(afero.MemMapFs)

	mustExist := func(path string) {
		_, err := fs.Stat(path)
		if os.IsNotExist(err) {
			t.Errorf("File %s does not exist", path)
		}
	}

	s := NewState(fs)
	s.Cleanup()

	mustExist("initialized")
	mustExist("enabled")
	mustExist("hours")
}
