package forwardreader

import (
	"io"
	"io/fs"

	"github.com/dirtbags/moth/v5/pkg/writecounter"
)

// ForwardReaderAt wraps an io.Reader to support io.ReaderAt.
// If ReadAt(p, off) is called with off before the current position in the underlying reader,
// an error is returned.
type ForwardReaderAt struct {
	// r is the underlying Reader we wrap
	r io.Reader

	// pos tracks how much has been consumed from r
	pos *writecounter.WriteCounter

	// sniffBuf keeps the first chunk read, in case it's requested a second time
	sniffBuf []byte
}

func NewForwardReaderAt(r io.Reader, length int64) *ForwardReaderAt {
	ls := new(ForwardReaderAt)
	ls.pos = new(writecounter.WriteCounter)
	ls.r = io.TeeReader(r, ls.pos)
	return ls
}

// ReadAt reads len(p) bytes into p starting at offset off in the underlying input source. It returns the number of bytes read (0 <= n <= len(p)) and any error encountered.
func (ls *ForwardReaderAt) ReadAt(p []byte, off int64) (int, error) {
	// send the cached sniff buffer, if it exists
	if (off == 0) && (ls.sniffBuf != nil) {
		copy(p, ls.sniffBuf)
		return len(ls.sniffBuf), nil
	}

	if off < ls.pos.Len() {
		return 0, fs.ErrInvalid
	}
	discard := off - ls.pos.Len()
	if _, err := io.CopyN(io.Discard, ls.r, discard); err != nil {
		return 0, err
	}

	n, err := ls.r.Read(p)
	if off == 0 {
		ls.sniffBuf = make([]byte, n)
		copy(ls.sniffBuf, p[:n])
	}

	return n, err
}

// NewForwardReadSeeker wraps r in an io.SectionReader that can only read forward.
//
// Seek() may be used as needed (for instance, to determine file size),
// but once a byte is read, no earlier byte can be read.
//
// This allows you to provide an fs.File to the http module,
// which, despite taking an io.Reader, actually requires an io.ReadSeeker.
// See https://mahesh-hegde.github.io/posts/go-interface-smuggling/
func NewForwardReadSeeker(r io.Reader, length int64) *io.SectionReader {
	fr := NewForwardReaderAt(r, length)
	return io.NewSectionReader(fr, 0, length)
}

type ForwardFile struct {
	io.ReadSeeker
	fs.File
}

func NewForwardFile(f fs.File) (*ForwardFile, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	ff := &ForwardFile{
		ReadSeeker: NewForwardReadSeeker(f, fi.Size()),
		File:       f,
	}
	return ff, nil
}

func (ff *ForwardFile) Read(p []byte) (int, error) {
	return ff.ReadSeeker.Read(p)
}
