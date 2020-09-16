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

// T represents the state of things
type T struct {
	Stdout   io.Writer
	Stderr   io.Writer
	Args     []string
	BaseFs   afero.Fs
	fs       afero.Fs
	filename string
	answer   string
}

// Command is a function invoked by the user
type Command func() error

func nothing() error {
	return nil
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "Usage: transpile COMMAND [flags]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, " mothball: Compile a mothball")
	fmt.Fprintln(w, " inventory: Show category inventory")
	fmt.Fprintln(w, " open: Open a file for a puzzle")
	fmt.Fprintln(w, " answer: Check correctness of an answer")
}

// ParseArgs parses arguments and runs the appropriate action.
func (t *T) ParseArgs() (Command, error) {
	var cmd Command

	if len(t.Args) == 1 {
		usage(t.Stderr)
		return nothing, nil
	}

	flags := flag.NewFlagSet(t.Args[1], flag.ContinueOnError)
	directory := flags.String("dir", "", "Work directory")

	switch t.Args[1] {
	case "mothball":
		cmd = t.DumpMothball
	case "inventory":
		cmd = t.PrintInventory
	case "open":
		flags.StringVar(&t.filename, "file", "puzzle.json", "Filename to open")
		cmd = t.DumpFile
	case "answer":
		flags.StringVar(&t.answer, "answer", "", "Answer to check")
		cmd = t.CheckAnswer
	case "help":
		usage(t.Stderr)
		return nothing, nil
	default:
		usage(t.Stderr)
		return nothing, fmt.Errorf("%s is not a valid command", t.Args[1])
	}

	flags.SetOutput(t.Stderr)
	if err := flags.Parse(t.Args[2:]); err != nil {
		return nothing, err
	}
	if *directory != "" {
		log.Println(*directory)
		t.fs = afero.NewBasePathFs(t.BaseFs, *directory)
	} else {
		t.fs = t.BaseFs
	}

	return cmd, nil
}

// PrintInventory prints a puzzle inventory to stdout
func (t *T) PrintInventory() error {
	inv, err := transpile.FsInventory(t.fs)
	if err != nil {
		return err
	}

	cats := make([]string, 0, len(inv))
	for cat := range inv {
		cats = append(cats, cat)
	}
	sort.Strings(cats)
	for _, cat := range cats {
		puzzles := inv[cat]
		fmt.Fprint(t.Stdout, cat)
		for _, p := range puzzles {
			fmt.Fprint(t.Stdout, " ", p)
		}
		fmt.Fprintln(t.Stdout)
	}
	return nil
}

// DumpFile writes a file to the writer.
// BUG(neale): The "open" and "answer" actions don't work on categories with an "mkcategory" executable.
func (t *T) DumpFile() error {
	puzzle := transpile.NewFsPuzzle(t.fs)

	switch t.filename {
	case "puzzle.json", "":
		p, err := puzzle.Puzzle()
		if err != nil {
			return err
		}
		jp, err := json.Marshal(p)
		if err != nil {
			return err
		}
		t.Stdout.Write(jp)
	default:
		f, err := puzzle.Open(t.filename)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := io.Copy(t.Stdout, f); err != nil {
			return err
		}
	}

	return nil
}

// DumpMothball writes a mothball to the writer.
func (t *T) DumpMothball() error {
	c := transpile.NewFsCategory(t.fs, "")
	mb, err := transpile.Mothball(c)
	if err != nil {
		return err
	}
	if _, err := io.Copy(t.Stdout, mb); err != nil {
		return err
	}
	return nil
}

// CheckAnswer prints whether an answer is correct.
func (t *T) CheckAnswer() error {
	c := transpile.NewFsPuzzle(t.fs)
	if c.Answer(t.answer) {
		fmt.Fprintln(t.Stdout, "correct")
	} else {
		fmt.Fprintln(t.Stdout, "wrong")
	}
	return nil
}

func main() {
	// XXX: Convert puzzle.py to standalone thingies

	t := &T{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Args:   os.Args,
		BaseFs: afero.NewOsFs(),
	}
	cmd, err := t.ParseArgs()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd(); err != nil {
		log.Fatal(err)
	}
}
