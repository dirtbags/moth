package main

import (
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/spf13/afero"
)

// NopReadCloser provides an io.ReadCloser which does nothing.
type NopReadCloser struct {
}

// Read satisfies io.Reader.
func (n NopReadCloser) Read(b []byte) (int, error) {
	return 0, nil
}

// Close satisfies io.Closer.
func (n NopReadCloser) Close() error {
	return nil
}

// NewFsCategory returns a Category based on which files are present.
// If 'mkcategory' is present and executable, an FsCommandCategory is returned.
// Otherwise, FsCategory is returned.
func NewFsCategory(fs afero.Fs) Category {
	if info, err := fs.Stat("mkcategory"); (err == nil) && (info.Mode()&0100 != 0) {
		return FsCommandCategory{fs: fs}
	}
	return FsCategory{fs: fs}
}

// FsCategory provides a category backed by a .md file.
type FsCategory struct {
	fs afero.Fs
}

// Inventory returns a list of point values for this category.
func (c FsCategory) Inventory() ([]int, error) {
	puzzleEntries, err := afero.ReadDir(c.fs, ".")
	if err != nil {
		return nil, err
	}

	puzzles := make([]int, 0, len(puzzleEntries))
	for _, ent := range puzzleEntries {
		if !ent.IsDir() {
			continue
		}
		if points, err := strconv.Atoi(ent.Name()); err != nil {
			log.Println("Skipping non-numeric directory", ent.Name())
			continue
		} else {
			puzzles = append(puzzles, points)
		}
	}
	return puzzles, nil
}

// Puzzle returns a Puzzle structure for the given point value.
func (c FsCategory) Puzzle(points int) (Puzzle, error) {
	return NewFsPuzzle(c.fs, points).Puzzle()
}

// Open returns an io.ReadCloser for the given filename.
func (c FsCategory) Open(points int, filename string) (io.ReadCloser, error) {
	return NewFsPuzzle(c.fs, points).Open(filename)
}

// Answer check whether an answer is correct.
func (c FsCategory) Answer(points int, answer string) bool {
	// BUG(neale): FsCategory.Answer should probably always return false, to prevent you from running uncompiled puzzles with participants.
	p, err := c.Puzzle(points)
	if err != nil {
		return false
	}
	for _, a := range p.Answers {
		if a == answer {
			return true
		}
	}
	return false
}

// FsCommandCategory provides a category backed by running an external command.
type FsCommandCategory struct {
	fs afero.Fs
}

// Inventory returns a list of point values for this category.
func (c FsCommandCategory) Inventory() ([]int, error) {
	return nil, fmt.Errorf("Not implemented")
}

// Puzzle returns a Puzzle structure for the given point value.
func (c FsCommandCategory) Puzzle(points int) (Puzzle, error) {
	return Puzzle{}, fmt.Errorf("Not implemented")
}

// Open returns an io.ReadCloser for the given filename.
func (c FsCommandCategory) Open(points int, filename string) (io.ReadCloser, error) {
	return NopReadCloser{}, fmt.Errorf("Not implemented")
}

// Answer check whether an answer is correct.
func (c FsCommandCategory) Answer(points int, answer string) bool {
	return false
}
