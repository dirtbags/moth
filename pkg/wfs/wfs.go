// Package wfs provides a writable fs.FS.
//
// It does this the cheater way, by passing everything down to the os package.
package wfs

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Joins any number of path elements into a single path, using the fs.FS path separator "/".
func Join(elem ...string) string {
	return strings.Join(elem, "/")
}

type WriteFS struct {
	fs.FS
	basepath string
}

func DirFS(basepath string) WriteFS {
	return WriteFS{
		FS:       os.DirFS(basepath),
		basepath: basepath,
	}
}

func (wfs WriteFS) OSPath(name string) string {
	clean := filepath.Clean("/" + name)
	return filepath.Join(wfs.basepath, clean)
}

func (wfs WriteFS) Remove(name string) error {
	return os.Remove(wfs.OSPath(name))
}

func (wfs WriteFS) RemoveAll(name string) error {
	return os.RemoveAll(wfs.OSPath(name))
}

func (wfs WriteFS) OpenFile(name string, flag int, perm fs.FileMode) (*os.File, error) {
	return os.OpenFile(wfs.OSPath(name), flag, perm)
}

func (wfs WriteFS) Create(name string) (*os.File, error) {
	return os.Create(wfs.OSPath(name))
}

func (wfs WriteFS) Mkdir(name string, perm fs.FileMode) error {
	return os.Mkdir(wfs.OSPath(name), perm)
}

func (wfs WriteFS) MkdirAll(name string, perm fs.FileMode) error {
	return os.MkdirAll(wfs.OSPath(name), perm)
}

func (wfs WriteFS) Rename(oldpath, newpath string) error {
	return os.Rename(wfs.OSPath(oldpath), wfs.OSPath(newpath))
}

func (wfs WriteFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(wfs.OSPath(name))
}

func (wfs WriteFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(wfs.OSPath(name), data, perm)
}

func Sub(wfs WriteFS, dir string) (WriteFS, error) {
	return DirFS(wfs.OSPath(dir)), nil
}
