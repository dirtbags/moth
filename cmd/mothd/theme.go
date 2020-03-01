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
func (t *Theme) Open(name string) (ReadSeekCloser, time.Time, error) {
	f, err := t.Fs.Open(name)
	if err != nil {
		return nil, time.Time{}, err
	}
	
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, time.Time{}, err
	}
	
	return f, fi.ModTime(), nil
}

func (t *Theme) Update() {
	// No periodic tasks for a theme
}
