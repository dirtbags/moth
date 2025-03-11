// Package mdir provides Maildir-like directories for lockless concurrency.
//
// These directories consist of small files which are written once at creation,
// and never modified after creation.
//
// Files are created and written in a "temporary" directory,
// then closed and renamed into a "new" directory.
// Files in the "new" directory can be read and deleted.
//
// md relies on atomic renames in the filesystem.
// Whether this is the case or not on Windows appears to be an open question.
package mdir

import (
	"io/fs"
	"path/filepath"

	"github.com/dirtbags/moth/v5/pkg/wfs"
)

// MDir is a Maildir-like directory.
//
// MDir acts as a WriteFS in NewDir,
// allowing developers to use an MDir like a WriteFS that magically supports concurrency.
type MDir struct {
	wfs.WriteFS

	// TmpDir is the directory where new files will be created and written.
	TmpDir string

	// NewDir is the directory where completed files will be moved.
	NewDir string

	// BaseFS is the WriteFS in which TmpDir and NewDir exist.
	BaseFS wfs.WriteFS
}

// New creates an MDir and ensures tmpdir and newdir exist within it.
func New(baseFS wfs.WriteFS, tmpdir string, newdir string) (*MDir, error) {
	md := MDir{
		TmpDir: tmpdir,
		NewDir: newdir,
		BaseFS: baseFS,
	}
	if err := md.MakeDirectories(); err != nil {
		return nil, err
	}
	if fsys, err := wfs.Sub(baseFS, newdir); err != nil {
		return nil, err
	} else {
		md.WriteFS = fsys
	}
	return &md, nil
}

// MakeDirectories ensures the required TmpDir and NewDir directories exist.
func (md *MDir) MakeDirectories() error {
	if err := md.BaseFS.MkdirAll(md.TmpDir, 0755); err != nil {
		return err
	}
	if err := md.BaseFS.MkdirAll(md.NewDir, 0755); err != nil {
		return err
	}
	return nil
}

// NewDefault is the same as New(fsys, "tmp", "cur")
func NewDefault(fsys wfs.WriteFS) (*MDir, error) {
	return New(fsys, "tmp", "cur")
}

func (md *MDir) tmpname(name string) string {
	return filepath.Join(md.TmpDir, name)
}

func (md *MDir) newname(name string) string {
	return filepath.Join(md.NewDir, name)
}

// BUG(neale): Windows will not overwrite an open file: MDir on multi-node Windows environments will fail.
func (md *MDir) rename(name string) error {
	return md.BaseFS.Rename(md.tmpname(name), md.newname(name))
}

// WriteFile behaves the same as os.WriteFile.
func (md *MDir) WriteFile(name string, data []byte, perm fs.FileMode) error {
	f := File{
		md:   md,
		name: name,
	}
	if err := md.BaseFS.WriteFile(md.tmpname(name), data, perm); err != nil {
		return err
	}
	return f.md.rename(f.name)
}

// Reset removes all files in TmpDir and NewDir
func (md *MDir) Reset() error {
	if err := md.BaseFS.RemoveAll(md.TmpDir); err != nil {
		return err
	}
	if err := md.BaseFS.RemoveAll(md.NewDir); err != nil {
		return err
	}
	return md.MakeDirectories()
}

// Create behaves the same as os.Create.
func (md *MDir) Create(name string) (File, error) {
	osf, err := md.BaseFS.Create(md.tmpname(name))
	f := File{
		osf,
		md,
		name,
	}
	return f, err
}
