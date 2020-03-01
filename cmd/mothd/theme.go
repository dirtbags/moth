package main

import (
	"github.com/spf13/afero"
	"time"
)

type Theme struct {
	afero.Fs
}

func NewTheme(fs afero.Fs) *Theme {
	return &Theme{
		Fs: fs,
	}
}

// I don't understand why I need this. The type checking system is weird here.
func (t *Theme) Open(name string) (ReadSeekCloser, error) {
	return t.Fs.Open(name)
}

func (t *Theme) ModTime(name string) (mt time.Time, err error) {
	fi, err := t.Fs.Stat(name)
	if err == nil {
		mt = fi.ModTime()
	}
	return
}

func (t *Theme) Update() {
	// No periodic tasks for a theme
}
