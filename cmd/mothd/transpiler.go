package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/dirtbags/moth/pkg/transpile"
	"github.com/spf13/afero"
)

// NewTranspilerProvider returns a new TranspilerProvider.
func NewTranspilerProvider(fs afero.Fs) TranspilerProvider {
	return TranspilerProvider{fs}
}

// TranspilerProvider provides puzzles generated from source files on disk
type TranspilerProvider struct {
	fs afero.Fs
}

// Inventory returns a Category list for this provider.
func (p TranspilerProvider) Inventory() []Category {
	ret := make([]Category, 0)
	inv, err := transpile.FsInventory(p.fs)
	if err != nil {
		log.Print(err)
		return ret
	}
	for name, points := range inv {
		ret = append(ret, Category{name, points})
	}
	return ret
}

type nopCloser struct {
	io.ReadSeeker
}

func (c nopCloser) Close() error {
	return nil
}

// Open returns a file associated with the given category and point value.
func (p TranspilerProvider) Open(cat string, points int, filename string) (ReadSeekCloser, time.Time, error) {
	c := transpile.NewFsCategory(p.fs, cat)
	switch filename {
	case "", "puzzle.json":
		p, err := c.Puzzle(points)
		if err != nil {
			return nopCloser{new(bytes.Reader)}, time.Time{}, err
		}
		jp, err := json.Marshal(p)
		if err != nil {
			return nopCloser{new(bytes.Reader)}, time.Time{}, err
		}
		return nopCloser{bytes.NewReader(jp)}, time.Now(), nil
	default:
		r, err := c.Open(points, filename)
		return r, time.Now(), err
	}
}

// CheckAnswer checks whether an answer si correct.
func (p TranspilerProvider) CheckAnswer(cat string, points int, answer string) (bool, error) {
	c := transpile.NewFsCategory(p.fs, cat)
	return c.Answer(points, answer), nil
}

// Mothball packages up a category into a mothball.
func (p TranspilerProvider) Mothball(cat string) (*bytes.Reader, error) {
	c := transpile.NewFsCategory(p.fs, cat)
	return transpile.Mothball(c)
}

// Maintain performs housekeeping.
func (p TranspilerProvider) Maintain(updateInterval time.Duration) {
	// Nothing to do here.
}
