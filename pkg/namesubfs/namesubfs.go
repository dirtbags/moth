package namesubfs

import (
	"io/fs"
	"log"
	"path"
)

// Sub returns a NameSubFS corresponding to the subtree rooted at fsys's dir.
func NameSub(fsys fs.FS, dir string) (*NameSubFS, error) {
	switch f := fsys.(type) {
	case *NameSubFS:
		return f.NameSub(dir)
	default:
		baseFS := &NameSubFS{fsys, ""}
		return baseFS.NameSub(dir)
	}
}

// A NameSubFS is a file system allowing the query of the full path name of entries
type NameSubFS struct {
	fs.FS
	dir string
}

// FullName returns the path to name.
//
// This is not the absolute path!
// It is relative to whatever was provided to the initial Sub call.
func (f *NameSubFS) FullName(name string) string {
	return path.Join(f.dir, name)
}

// NameSub returns a NameSubFS corresponding to the subtree rooted at dir.
func (f *NameSubFS) NameSub(dir string) (*NameSubFS, error) {
	log.Println("Sub", f.dir)
	newFS, err := fs.Sub(f.FS, dir)
	if err != nil {
		return nil, err
	}
	newNameSubFS := NameSubFS{
		FS:  newFS,
		dir: f.FullName(dir),
	}
	return &newNameSubFS, err
}

// NameSub returns an FS corresponding to the subtree rooted at dir.
func (f *NameSubFS) Sub(dir string) (fs.FS, error) {
	return f.NameSub(dir)
}
