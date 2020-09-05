package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

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
func NewFsCategory(fs afero.Fs, cat string) Category {
	bfs := NewRecursiveBasePathFs(fs, cat)
	if info, err := bfs.Stat("mkcategory"); (err == nil) && (info.Mode()&0100 != 0) {
		if command, err := bfs.RealPath(info.Name()); err != nil {
			log.Println("Unable to resolve full path to", info.Name(), bfs)
		} else {
			return FsCommandCategory{
				fs:      bfs,
				command: command,
				timeout: 2 * time.Second,
			}
		}
	}
	return FsCategory{fs: bfs}
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

// Answer checks whether an answer is correct.
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
	fs      afero.Fs
	command string
	timeout time.Duration
}

// Inventory returns a list of point values for this category.
func (c FsCommandCategory) Inventory() ([]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.command, "inventory")
	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	ret := make([]int, 0)
	if err := json.Unmarshal(stdout, &ret); err != nil {
		return nil, err
	}

	return ret, nil
}

// Puzzle returns a Puzzle structure for the given point value.
func (c FsCommandCategory) Puzzle(points int) (Puzzle, error) {
	var p Puzzle

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.command, "puzzle", strconv.Itoa(points))
	stdout, err := cmd.Output()
	if err != nil {
		return p, err
	}

	if err := json.Unmarshal(stdout, &p); err != nil {
		return p, err
	}

	p.computeAnswerHashes()

	return p, nil
}

// Open returns an io.ReadCloser for the given filename.
func (c FsCommandCategory) Open(points int, filename string) (io.ReadCloser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.command, "file", strconv.Itoa(points), filename)
	stdout, err := cmd.Output()
	buf := ioutil.NopCloser(bytes.NewBuffer(stdout))
	if err != nil {
		return buf, err
	}

	return buf, nil
}

// Answer checks whether an answer is correct.
func (c FsCommandCategory) Answer(points int, answer string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.command, "answer", strconv.Itoa(points), answer)
	stdout, err := cmd.Output()
	if err != nil {
		log.Printf("ERROR: Answering %d points: %s", points, err)
		return false
	}

	switch strings.TrimSpace(string(stdout)) {
	case "correct":
		return true
	}

	return false
}
