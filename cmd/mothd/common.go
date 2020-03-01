package main

import (
	"io"
	"time"
)

type Category struct {
	Name    string
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

type StateExport struct {
	Messages  string
	TeamNames map[string]string
	PointsLog []Award
}

type StateProvider interface {
	Export(teamId string) *StateExport
	SetTeamName(teamId, teamName string) error
}

type Component interface {
	Update()
}
