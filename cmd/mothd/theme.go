package main

import (
	"io"
	"os"
	"path"
	"time"
)

// Theme defines a filesystem-backed ThemeProvider.
type Theme struct {
	basedir string
}

// NewTheme returns a new Theme, backed by Fs.
func NewTheme(basedir string) *Theme {
	return &Theme{
		basedir: basedir,
	}
}

// Open returns a new opened file.
func (t *Theme) Open(name string) (io.ReadSeekCloser, time.Time, error) {
	f, err := os.Open(path.Join(t.basedir, name))
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

// Maintain performs housekeeping for a Theme.
func (t *Theme) Maintain(i time.Duration) {
	// No periodic tasks for a theme
}
