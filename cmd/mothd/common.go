package main

import (
	"io"
	"time"
)

type Category struct {
	Name string
	Puzzles []int
}

type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

type PuzzleProvider interface {
	Metadata(cat string, points int) (io.ReadCloser, error)
	Open(cat string, points int, path string) (io.ReadCloser, error)
	Inventory() []Category
}

type ThemeProvider interface {
	Open(path string) (ReadSeekCloser, error)
	ModTime(path string) (time.Time, error)
}

type StateProvider interface {
	
}

type Component interface {
	Update()
}
