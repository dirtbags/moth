package mdir

import "os"

// File provides an os.File interface.
type File struct {
	*os.File
	md   *MDir
	name string
}

// Close the file and rename it into the "new" directory.
func (f File) Close() error {
	if err := f.File.Close(); err != nil {
		return err
	}
	return f.md.rename(f.name)
}
