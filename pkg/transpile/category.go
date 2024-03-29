package transpile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/afero"
)

// InventoryResponse is what's handed back when we ask for an inventory.
type InventoryResponse struct {
	Puzzles []int
}

// Category defines the functionality required to be a puzzle category.
type Category interface {
	// Inventory lists every puzzle in the category.
	Inventory() ([]int, error)

	// Puzzle provides a Puzzle structure for the given point value.
	Puzzle(points int) (Puzzle, error)

	// Open returns an io.ReadCloser for the given filename.
	Open(points int, filename string) (ReadSeekCloser, error)

	// Answer returns whether the given answer is correct.
	Answer(points int, answer string) bool
}

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
			log.Println("Unable to resolve full path to", info.Name())
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
	return NewFsPuzzlePoints(c.fs, points).Puzzle()
}

// Open returns an io.ReadCloser for the given filename.
func (c FsCategory) Open(points int, filename string) (ReadSeekCloser, error) {
	return NewFsPuzzlePoints(c.fs, points).Open(filename)
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

func (c FsCommandCategory) run(command string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	cmdargs := append([]string{command}, args...)
	cmd := exec.CommandContext(ctx, "./"+path.Base(c.command), cmdargs...)
	cmd.Dir = path.Dir(c.command)
	out, err := cmd.Output()
	if err, ok := err.(*exec.ExitError); ok {
		stderr := strings.TrimSpace(string(err.Stderr))
		return nil, fmt.Errorf("%s (%s)", stderr, err.String())
	}
	return out, err
}

// Inventory returns a list of point values for this category.
func (c FsCommandCategory) Inventory() ([]int, error) {
	stdout, err := c.run("inventory")
	if exerr, ok := err.(*exec.ExitError); ok {
		return nil, fmt.Errorf("inventory: %s: %s", err, string(exerr.Stderr))
	} else if err != nil {
		return nil, err
	}

	inv := InventoryResponse{}
	if err := json.Unmarshal(stdout, &inv); err != nil {
		return nil, err
	}

	return inv.Puzzles, nil
}

// Puzzle returns a Puzzle structure for the given point value.
func (c FsCommandCategory) Puzzle(points int) (Puzzle, error) {
	var p Puzzle

	stdout, err := c.run("puzzle", strconv.Itoa(points))
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
func (c FsCommandCategory) Open(points int, filename string) (ReadSeekCloser, error) {
	stdout, err := c.run("file", strconv.Itoa(points), filename)
	return nopCloser{bytes.NewReader(stdout)}, err
}

// Answer checks whether an answer is correct.
func (c FsCommandCategory) Answer(points int, answer string) bool {
	stdout, err := c.run("answer", strconv.Itoa(points), answer)
	if err != nil {
		log.Printf("ERROR: Answering %d points: %s", points, err)
		return false
	}

	ans := AnswerResponse{}
	if err := json.Unmarshal(stdout, &ans); err != nil {
		log.Printf("ERROR: Answering %d points: %s", points, err)
		return false
	}

	return ans.Correct
}
