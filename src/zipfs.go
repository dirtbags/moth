package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type Zipfs struct {
	zf       *zip.ReadCloser
	filename string
	mtime    time.Time
}

type ZipfsFile struct {
	f   io.ReadCloser
	pos int64
	zf  *zip.File
	io.Reader
	io.Seeker
	io.Closer
}

func NewZipfsFile(zf *zip.File) (*ZipfsFile, error) {
	zfsf := &ZipfsFile{
		zf:  zf,
		pos: 0,
		f:   nil,
	}
	if err := zfsf.reopen(); err != nil {
		return nil, err
	}
	return zfsf, nil
}

func (zfsf *ZipfsFile) reopen() error {
	if zfsf.f != nil {
		if err := zfsf.f.Close(); err != nil {
			return err
		}
	}
	f, err := zfsf.zf.Open()
	if err != nil {
		return err
	}
	zfsf.f = f
	zfsf.pos = 0
	return nil
}

func (zfsf *ZipfsFile) ModTime() time.Time {
	return zfsf.zf.Modified
}

func (zfsf *ZipfsFile) Read(p []byte) (int, error) {
	n, err := zfsf.f.Read(p)
	zfsf.pos += int64(n)
	return n, err
}

func (zfsf *ZipfsFile) Seek(offset int64, whence int) (int64, error) {
	var pos int64
	switch whence {
	case io.SeekStart:
		pos = offset
	case io.SeekCurrent:
		pos = zfsf.pos + int64(offset)
	case io.SeekEnd:
		pos = int64(zfsf.zf.UncompressedSize64) - int64(offset)
	}

	if pos < 0 {
		return zfsf.pos, fmt.Errorf("Tried to seek %d before start of file", pos)
	}
	if pos >= int64(zfsf.zf.UncompressedSize64) {
		// We don't need to decompress anything, we're at the end of the file
		zfsf.f.Close()
		zfsf.f = ioutil.NopCloser(strings.NewReader(""))
		zfsf.pos = int64(zfsf.zf.UncompressedSize64)
		return zfsf.pos, nil
	}
	if pos < zfsf.pos {
		if err := zfsf.reopen(); err != nil {
			return zfsf.pos, err
		}
	}

	buf := make([]byte, 32*1024)
	for pos > zfsf.pos {
		l := pos - zfsf.pos
		if l > int64(cap(buf)) {
			l = int64(cap(buf)) - 1
		}
		p := buf[0:int(l)]
		n, err := zfsf.Read(p)
		if err != nil {
			return zfsf.pos, err
		} else if n <= 0 {
			return zfsf.pos, fmt.Errorf("Short read (%d bytes)", n)
		}
	}

	return zfsf.pos, nil
}

func (zfsf *ZipfsFile) Close() error {
	return zfsf.f.Close()
}

func OpenZipfs(filename string) (*Zipfs, error) {
	var zfs Zipfs

	zfs.filename = filename

	err := zfs.Refresh()
	if err != nil {
		return nil, err
	}

	return &zfs, nil
}

func (zfs *Zipfs) Close() error {
	return zfs.zf.Close()
}

func (zfs *Zipfs) Refresh() error {
	info, err := os.Stat(zfs.filename)
	if err != nil {
		return err
	}
	mtime := info.ModTime()

	if !mtime.After(zfs.mtime) {
		return nil
	}

	zf, err := zip.OpenReader(zfs.filename)
	if err != nil {
		return err
	}

	if zfs.zf != nil {
		zfs.zf.Close()
	}
	zfs.zf = zf
	zfs.mtime = mtime

	return nil
}

func (zfs *Zipfs) get(filename string) (*zip.File, error) {
	for _, f := range zfs.zf.File {
		if filename == f.Name {
			return f, nil
		}
	}
	return nil, fmt.Errorf("File not found: %s %s", zfs.filename, filename)
}

func (zfs *Zipfs) Header(filename string) (*zip.FileHeader, error) {
	f, err := zfs.get(filename)
	if err != nil {
		return nil, err
	}
	return &f.FileHeader, nil
}

func (zfs *Zipfs) Open(filename string) (*ZipfsFile, error) {
	f, err := zfs.get(filename)
	if err != nil {
		return nil, err
	}
	zfsf, err := NewZipfsFile(f)
	return zfsf, err
}

func (zfs *Zipfs) ReadFile(filename string) ([]byte, error) {
	f, err := zfs.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	bytes, err := ioutil.ReadAll(f)
	return bytes, err
}
