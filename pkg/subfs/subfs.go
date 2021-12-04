package transpile

import (
	"io/fs"
	"path"
)

func Sub(fsys fs.FS, dir string) (*SubFS, error) {
	return &SubFS{fsys, dir}, nil
}

type SubFS struct {
	fs.FS
	dir string
}

func (f *SubFS) FullName(name string) string {
	return path.Join(f.dir, name)
}

func (f *SubFS) Sub(dir string) (*SubFS, error) {
	newFS, err := fs.Sub(f, dir)
	newSubFS := SubFS{
		FS:  newFS,
		dir: f.FullName(dir),
	}
	return &newSubFS, err
}
