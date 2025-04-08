package mdir

import (
	"errors"
	"io"
	"io/fs"
)

// ReadCloser implements io.ReadCloser
type ReadCloser struct {
	entries []fs.DirEntry
	md      *MDir
	f       fs.File
}

// OpenAll returns a ReadCloser which concatenates every file in md.
func (md *MDir) OpenAll() (*ReadCloser, error) {
	entries, err := fs.ReadDir(md, ".")
	if err != nil {
		return nil, err
	}

	rc := ReadCloser{
		entries,
		md,
		nil,
	}
	return &rc, nil
}

func (rc *ReadCloser) Read(p []byte) (n int, err error) {
	if rc.f == nil {
		if len(rc.entries) == 0 {
			return 0, io.EOF
		}
		entry := rc.entries[0]
		rc.entries = rc.entries[1:]

		rc.f, err = rc.md.Open(entry.Name())
		if err != nil {
			return 0, err
		}
	}

	n, err = rc.f.Read(p)
	if errors.Is(err, io.EOF) {
		rc.f.Close()
		rc.f = nil
		return n, nil // will it cause problems to return 0, nil?
	}

	return n, err
}
