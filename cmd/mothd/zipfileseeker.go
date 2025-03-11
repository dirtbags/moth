package main

import (
	"archive/zip"
	"io"
	"io/fs"
)

// ZipFileSeeker wraps a zip.File with functionality to allow seeking.
// This allows using a zip.File with http.ServeContent.
//
// Seek is implemented by uncompressing the contents.
// This could be compute-intensive.
type ZipFileSeeker struct {
	*zip.File
	offset int64
	f      io.ReadCloser
}

// NewZipFileSeeker implements io.ReadSeekCloser on a zip.File.
func NewZipFileSeeker(zf *zip.File) *ZipFileSeeker {
	return &ZipFileSeeker{
		File: zf,
	}
}

func (zf *ZipFileSeeker) open() error {
	if zf.f != nil {
		return nil
	}
	return zf.Rewind()
}

func (zf *ZipFileSeeker) Rewind() error {
	if err := zf.Close(); err != nil {
		return err
	}

	f, err := zf.Open()
	if err != nil {
		return err
	}
	zf.f = f
	zf.offset = 0
	return nil
}

func (zf *ZipFileSeeker) Read(p []byte) (n int, err error) {
	if zf.offset == int64(zf.UncompressedSize64)-1 {
		return 0, io.EOF
	}
	if err := zf.open(); err != nil {
		return 0, err
	}
	if zf.f == nil {
		if err := zf.Rewind(); err != nil {
			return 0, err
		}
	}
	n, err = zf.f.Read(p)
	zf.offset += int64(n)
	return n, err
}

// Seek is implemented by uncompressing the contents.
// This could be compute-intensive.
func (zf *ZipFileSeeker) Seek(offset int64, whence int) (int64, error) {
	var pos int64
	switch whence {
	case io.SeekStart:
		pos = offset
	case io.SeekCurrent:
		pos = zf.offset + offset
	case io.SeekEnd:
		pos = int64(zf.UncompressedSize64) - 1 + offset
	default:
		return zf.offset, fs.ErrInvalid
	}

	// Out of bounds?
	if (pos < 0) || (pos >= int64(zf.UncompressedSize64)) {
		return zf.offset, fs.ErrInvalid
	}

	// Optimize seek to end of file.
	if pos == int64(zf.UncompressedSize64)-1 {
		err := zf.Close()
		return zf.offset, err
	}

	toread := pos - zf.offset
	if toread < 0 {
		if err := zf.Rewind(); err != nil {
			return zf.offset, err
		}
		toread = pos
	} else {
		if err := zf.open(); err != nil {
			return zf.offset, err
		}
	}

	null := new(WriteCounter)
	n, err := io.CopyN(null, zf.f, toread)
	zf.offset += n
	return zf.offset, err
}

// Close the currently-open file, if any, and set the internal position to the end of the file.
func (zf *ZipFileSeeker) Close() error {
	zf.offset = int64(zf.UncompressedSize64) - 1
	if zf.f != nil {
		return zf.f.Close()
	}
	return nil
}
