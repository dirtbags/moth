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
	Open(cat string, points int, path string) (ReadSeekCloser, error)
	ModTime(cat string, points int, path string) (time.Time, error)
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
	TeamName(teamId string) (string, error)
	SetTeamName(teamId, teamName string) error
}

type Component interface {
	Update()
}
