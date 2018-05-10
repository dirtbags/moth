package mothball

import (
	"archive/zip"
	"os"
	"time"
)

type Mothball struct {
	zf *zipfile.File,
	filename string,
	mtime time.Time,
}

func Open(filename string) (*Mothball, error) {
	var m Mothball
	
	m.filename = filename
	
	err := m.Refresh()
	if err != nil {
		return err
	}

	return &m
}

func (m Mothball) Close() (error) {
	return m.zf.Close()
}

func (m Mothball) Refresh() (error) {
	mtime, err := os.Stat(m.filename)
	if err != nil {
		return err
	}
	
	if mtime == m.mtime {
		return nil
	}
	
	zf, err := zip.OpenReader(m.filename)
	if err != nil {
		return err
	}
	
	m.zf.Close()
	m.zf = zf
	m.mtime = mtime
}

func (m Mothball)