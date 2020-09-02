package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	"github.com/GoBike/envflag"
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

// NewCategory returns a new Category as specified by cat.
func (t *T) NewCategory(cat string) Category {
	return NewCategory(t.Fs, cat)
}

// ParseArgs parses command-line arguments into T, returning the action to take
func (t *T) ParseArgs() string {
	action := flag.String("action", "inventory", "Action to take: must be 'inventory', 'open', 'answer', or 'mothball'")
	flag.StringVar(&t.Cat, "cat", "", "Puzzle category")
	flag.IntVar(&t.Points, "points", 0, "Puzzle point value")
	flag.StringVar(&t.Answer, "answer", "", "Answer to check for correctness, for 'answer' action")
	flag.StringVar(&t.Filename, "filename", "", "Filename, for 'open' action")
	basedir := flag.String("basedir", ".", "Base directory containing all puzzles")
	envflag.Parse()

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
	default:
		return fmt.Errorf("Unimplemented action: %s", action)
	}
}

// PrintInventory prints a puzzle inventory to stdout
func (t *T) PrintInventory() error {
	dirEnts, err := afero.ReadDir(t.Fs, ".")
	if err != nil {
		return err
	}
	for _, ent := range dirEnts {
		if ent.IsDir() {
			c := t.NewCategory(ent.Name())
			if puzzles, err := c.Puzzles(); err != nil {
				log.Print(err)
				continue
			} else {
				fmt.Fprint(t.w, ent.Name())
				sort.Ints(puzzles)
				for _, points := range puzzles {
					fmt.Fprint(t.w, " ")
					fmt.Fprint(t.w, points)
				}
				fmt.Fprintln(t.w)
			}
		}
	}
	return nil
}

// Open writes a file to the writer.
func (t *T) Open() error {
	c := t.NewCategory(t.Cat)
	pd := c.PuzzleDir(t.Points)

	switch t.Filename {
	case "puzzle.json", "":
		p, err := pd.Export()
		if err != nil {
			return err
		}
		jp, err := json.Marshal(p)
		if err != nil {
			return err
		}
		t.w.Write(jp)
	default:
		f, err := pd.Open(t.Filename)
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
