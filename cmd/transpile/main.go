package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	"github.com/dirtbags/moth/pkg/transpile"

	"github.com/spf13/afero"
)

// T contains everything required for a transpilation invocation (across the nation).
type T struct {
	// What action to take
	w        io.Writer
	Cat      string
	Points   int
	Answer   string
	Filename string
	Fs       afero.Fs
}

// ParseArgs parses command-line arguments into T, returning the action to take.
// BUG(neale): CLI arguments are not related to how the CLI will be used.
func (t *T) ParseArgs() string {
	action := flag.String("action", "inventory", "Action to take: must be 'inventory', 'open', 'answer', or 'mothball'")
	flag.StringVar(&t.Cat, "cat", "", "Puzzle category")
	flag.IntVar(&t.Points, "points", 0, "Puzzle point value")
	flag.StringVar(&t.Answer, "answer", "", "Answer to check for correctness, for 'answer' action")
	flag.StringVar(&t.Filename, "filename", "", "Filename, for 'open' action")
	basedir := flag.String("basedir", ".", "Base directory containing all puzzles")
	flag.Parse()

	osfs := afero.NewOsFs()
	t.Fs = afero.NewBasePathFs(osfs, *basedir)

	return *action
}

// Handle performs the requested action
func (t *T) Handle(action string) error {
	switch action {
	case "inventory":
		return t.PrintInventory()
	case "open":
		return t.Open()
	case "mothball":
		return t.Mothball()
	default:
		return fmt.Errorf("Unimplemented action: %s", action)
	}
}

// PrintInventory prints a puzzle inventory to stdout
func (t *T) PrintInventory() error {
	inv := make(map[string][]int)

	dirEnts, err := afero.ReadDir(t.Fs, ".")
	if err != nil {
		return err
	}
	for _, ent := range dirEnts {
		if ent.IsDir() {
			c := t.NewCategory(ent.Name())
			if puzzles, err := c.Inventory(); err != nil {
				log.Print(err)
				continue
			} else {
				sort.Ints(puzzles)
				inv[ent.Name()] = puzzles
			}
		}
	}

	m := json.NewEncoder(t.w)
	if err := m.Encode(inv); err != nil {
		return err
	}
	return nil
}

// Open writes a file to the writer.
func (t *T) Open() error {
	c := t.NewCategory(t.Cat)

	switch t.Filename {
	case "puzzle.json", "":
		// BUG(neale): we need a way to tell the transpiler to strip answers
		p, err := c.Puzzle(t.Points)
		if err != nil {
			return err
		}
		jp, err := json.Marshal(p)
		if err != nil {
			return err
		}
		t.w.Write(jp)
	default:
		f, err := c.Open(t.Points, t.Filename)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := io.Copy(t.w, f); err != nil {
			return err
		}
	}

	return nil
}

// Mothball writes a mothball to the writer.
func (t *T) Mothball() error {
	c := t.NewCategory(t.Cat)
	mb, err := transpile.Mothball(c)
	if err != nil {
		return err
	}
	if _, err := io.Copy(t.w, mb); err != nil {
		return err
	}
	return nil
}

// NewCategory returns a new Fs-backed category.
func (t *T) NewCategory(name string) transpile.Category {
	return transpile.NewFsCategory(t.Fs, name)
}

func main() {
	// XXX: Convert puzzle.py to standalone thingies

	t := &T{
		w: os.Stdout,
	}
	action := t.ParseArgs()
	if err := t.Handle(action); err != nil {
		log.Fatal(err)
	}
}
