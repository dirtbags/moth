package main

import (
	"time"

	"github.com/spf13/afero"
)

// Theme defines a filesystem-backed ThemeProvider.
type Theme struct {
	afero.Fs
}

// NewTheme returns a new Theme, backed by Fs.
func NewTheme(fs afero.Fs) *Theme {
	return &Theme{
		Fs: fs,
	}
}

// Open returns a new opened file.
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

// Maintain performs housekeeping for a Theme.
func (t *Theme) Maintain(i time.Duration) {
	// No periodic tasks for a theme
}
