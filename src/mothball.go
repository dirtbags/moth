package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
)

type Mothball struct {
	zf *zip.ReadCloser
	filename string
	mtime time.Time
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

func (m *Mothball) Close() (error) {
	return m.zf.Close()
}

func (m *Mothball) Refresh() (error) {
	info, err := os.Stat(m.filename)
	if err != nil {
		return err
	}
	mtime := info.ModTime()
	
	if ! mtime.After(m.mtime) {
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
	
	return nil
}

func (m *Mothball) Open(filename string) (io.ReadCloser, error) {
	for _, f := range m.zf.File {
		if filename == f.Name {
			ret, err := f.Open()
			return ret, err
		}
	}
	return nil, fmt.Errorf("File not found: %s in %s", filename, m.filename)
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
