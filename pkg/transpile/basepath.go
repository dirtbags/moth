package transpile

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/afero"
)

// RecursiveBasePathFs is an overloaded afero.BasePathFs that has a recursive RealPath().
type RecursiveBasePathFs struct {
	afero.Fs
	source afero.Fs
	path   string
}

// NewRecursiveBasePathFs returns a new RecursiveBasePathFs.
func NewRecursiveBasePathFs(source afero.Fs, path string) *RecursiveBasePathFs {
	ret := &RecursiveBasePathFs{
		source: source,
		path:   path,
	}
	if path == "" {
		ret.Fs = source
	} else {
		ret.Fs = afero.NewBasePathFs(source, path)
	}
	return ret
}

// RealPath returns the real path to a file, "breaking out" of the RecursiveBasePathFs.
func (b *RecursiveBasePathFs) RealPath(name string) (path string, err error) {
	if err := validateBasePathName(name); err != nil {
		return name, err
	}

	bpath := filepath.Clean(b.path)
	path = filepath.Clean(filepath.Join(bpath, name))

	if parentRecursiveBasePathFs, ok := b.source.(*RecursiveBasePathFs); ok {
		return parentRecursiveBasePathFs.RealPath(path)
	} else if parentRecursiveBasePathFs, ok := b.source.(*afero.BasePathFs); ok {
		return parentRecursiveBasePathFs.RealPath(path)
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
