package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

type PuzzleMap struct {
        Points int
        Path   string
}


type Mothball struct {
	zf       *zip.ReadCloser
	filename string
	puzzlemap	[]PuzzleMap
	mtime    time.Time
}

type MothballFile struct {
	f   io.ReadCloser
	pos int64
	zf  *zip.File
	io.Reader
	io.Seeker
	io.Closer
}

func NewMothballFile(zf *zip.File) (*MothballFile, error) {
	mf := &MothballFile{
		zf:  zf,
		pos: 0,
		f:   nil,
	}
	if err := mf.reopen(); err != nil {
		return nil, err
	}
	return mf, nil
}

func (mf *MothballFile) reopen() error {
	if mf.f != nil {
		if err := mf.f.Close(); err != nil {
			return err
		}
	}
	f, err := mf.zf.Open()
	if err != nil {
		return err
	}
	mf.f = f
	mf.pos = 0
	return nil
}

func (mf *MothballFile) ModTime() time.Time {
	return mf.zf.Modified
}

func (mf *MothballFile) Read(p []byte) (int, error) {
	n, err := mf.f.Read(p)
	mf.pos += int64(n)
	return n, err
}

func (mf *MothballFile) Seek(offset int64, whence int) (int64, error) {
	var pos int64
	switch whence {
	case io.SeekStart:
		pos = offset
	case io.SeekCurrent:
		pos = mf.pos + int64(offset)
	case io.SeekEnd:
		pos = int64(mf.zf.UncompressedSize64) - int64(offset)
	}

	if pos < 0 {
		return mf.pos, fmt.Errorf("Tried to seek %d before start of file", pos)
	}
	if pos >= int64(mf.zf.UncompressedSize64) {
		// We don't need to decompress anything, we're at the end of the file
		mf.f.Close()
		mf.f = ioutil.NopCloser(strings.NewReader(""))
		mf.pos = int64(mf.zf.UncompressedSize64)
		return mf.pos, nil
	}
	if pos < mf.pos {
		if err := mf.reopen(); err != nil {
			return mf.pos, err
		}
	}

	buf := make([]byte, 32*1024)
	for pos > mf.pos {
		l := pos - mf.pos
		if l > int64(cap(buf)) {
			l = int64(cap(buf)) - 1
		}
		p := buf[0:int(l)]
		n, err := mf.Read(p)
		if err != nil {
			return mf.pos, err
		} else if n <= 0 {
			return mf.pos, fmt.Errorf("Short read (%d bytes)", n)
		}
	}

	return mf.pos, nil
}

func (mf *MothballFile) Close() error {
	return mf.f.Close()
}

func OpenMothball(filename string) (*Mothball, error) {
	var m Mothball

	m.filename = filename

	err := m.Refresh()
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *Mothball) Close() error {
	return m.zf.Close()
}

func (m *Mothball) Refresh() error {
	info, err := os.Stat(m.filename)
	if err != nil {
		return err
	}
	mtime := info.ModTime()

	if !mtime.After(m.mtime) {
		return nil
	}

	zf, err := zip.OpenReader(m.filename)
	if err != nil {
		return err
	}

	if m.zf != nil {
		m.zf.Close()
	}
	m.zf = zf
	m.mtime = mtime

	mf, err := m.Open("map.txt")
	if err != nil {
	        // File isn't in there
	} else {
		defer mf.Close()

		pm := make([]PuzzleMap, 0, 30)
		scanner := bufio.NewScanner(mf)

		for scanner.Scan() {
			line := scanner.Text()

			var pointval int
			var dir string

			n, err := fmt.Sscanf(line, "%d %s", &pointval, &dir)
			if err != nil {
				log.Printf("Parsing map for %v", err)
			} else if n != 2 {
				log.Printf("Parsing map: short read")
			}

			pm = append(pm, PuzzleMap{pointval, dir})

		 }

		 m.puzzlemap = pm
	 }


	return nil
}

func (m *Mothball) get(filename string) (*zip.File, error) {
	for _, f := range m.zf.File {
		if filename == f.Name {
			return f, nil
		}
	}
	return nil, fmt.Errorf("File not found: %s %s", m.filename, filename)
}

func (m *Mothball) Header(filename string) (*zip.FileHeader, error) {
	f, err := m.get(filename)
	if err != nil {
		return nil, err
	}
	return &f.FileHeader, nil
}

func (m *Mothball) Open(filename string) (*MothballFile, error) {
	f, err := m.get(filename)
	if err != nil {
		return nil, err
	}
	mf, err := NewMothballFile(f)
	return mf, err
}

func (m *Mothball) ReadFile(filename string) ([]byte, error) {
	f, err := m.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	bytes, err := ioutil.ReadAll(f)
	return bytes, err
}
