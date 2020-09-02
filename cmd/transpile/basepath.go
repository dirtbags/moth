package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/afero"
)

// BasePathFs is an overloaded afero.BasePathFs that has a recursive RealPath().
type BasePathFs struct {
	afero.Fs
	source afero.Fs
	path   string
}

// NewBasePathFs returns a new BasePathFs.
func NewBasePathFs(source afero.Fs, path string) afero.Fs {
	return &BasePathFs{
		Fs:     afero.NewBasePathFs(source, path),
		source: source,
		path:   path,
	}
}

// RealPath returns the real path to a file, "breaking out" of the BasePathFs.
func (b *BasePathFs) RealPath(name string) (path string, err error) {
	if err := validateBasePathName(name); err != nil {
		return name, err
	}

	bpath := filepath.Clean(b.path)
	path = filepath.Clean(filepath.Join(bpath, name))

	if parentBasePathFs, ok := b.source.(*BasePathFs); ok {
		return parentBasePathFs.RealPath(path)
	} else if parentBasePathFs, ok := b.source.(*afero.BasePathFs); ok {
		return parentBasePathFs.RealPath(path)
	}

	if !strings.HasPrefix(path, bpath) {
		return name, os.ErrNotExist
	}

	return path, nil
}

func validateBasePathName(name string) error {
	if runtime.GOOS != "windows" {
		// Not much to do here;
		// the virtual file paths all look absolute on *nix.
		return nil
	}

	// On Windows a common mistake would be to provide an absolute OS path
	// We could strip out the base part, but that would not be very portable.
	if filepath.IsAbs(name) {
		return os.ErrNotExist
	}

	return nil
}
