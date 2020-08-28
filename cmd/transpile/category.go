package main

import (
	"log"
	"strconv"

	"github.com/spf13/afero"
)

// NewCategory returns a new category for the given path in the given fs.
func NewCategory(fs afero.Fs, cat string) Category {
	return Category{
		Fs: afero.NewBasePathFs(fs, cat),
	}
}

// Category represents an on-disk category.
type Category struct {
	afero.Fs
}

// Puzzles returns a list of puzzle values.
func (c Category) Puzzles() ([]int, error) {
	puzzleEntries, err := afero.ReadDir(c, ".")
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

// Puzzle returns the Puzzle associated with points.
func (c Category) Puzzle(points int) (*Puzzle, error) {
	return NewPuzzle(c.Fs, points)
}
