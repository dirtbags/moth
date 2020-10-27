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
	Stdout io.Writer
	Stderr io.Writer
	Args   []string
	BaseFs afero.Fs
	fs     afero.Fs

	// Arguments
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
	fmt.Fprintln(w, " puzzle: Print puzzle JSON")
	fmt.Fprintln(w, " file: Open a file for a puzzle")
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
	flags.SetOutput(t.Stderr)
	directory := flags.String("dir", "", "Work directory")

	switch t.Args[1] {
	case "mothball":
		cmd = t.DumpMothball
		flags.StringVar(&t.filename, "out", "", "Path to create mothball (empty for stdout)")
	case "inventory":
		cmd = t.PrintInventory
	case "puzzle":
		cmd = t.DumpPuzzle
	case "file":
		cmd = t.DumpFile
		flags.StringVar(&t.filename, "file", "puzzle.json", "Filename to open")
	case "answer":
		cmd = t.CheckAnswer
		flags.StringVar(&t.answer, "answer", "", "Answer to check")
	case "help":
		usage(t.Stderr)
		return nothing, nil
	default:
		fmt.Fprintln(t.Stderr, "ERROR:", t.Args[1], "is not a valid command")
		usage(t.Stderr)
		return nothing, fmt.Errorf("Invalid command")
	}

	if err := flags.Parse(t.Args[2:]); err != nil {
		return nothing, err
	}
	if *directory != "" {
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

// DumpPuzzle writes a puzzle's JSON to the writer.
func (t *T) DumpPuzzle() error {
	puzzle := transpile.NewFsPuzzle(t.fs)

	p, err := puzzle.Puzzle()
	if err != nil {
		return err
	}
	jp, err := json.Marshal(p)
	if err != nil {
		return err
	}
	t.Stdout.Write(jp)
	return nil
}

// DumpFile writes a file to the writer.
func (t *T) DumpFile() error {
	puzzle := transpile.NewFsPuzzle(t.fs)

	f, err := puzzle.Open(t.filename)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(t.Stdout, f); err != nil {
		return err
	}

	return nil
}

// DumpMothball writes a mothball to the writer, or an output file if specified.
func (t *T) DumpMothball() error {
	var w io.Writer

	c := transpile.NewFsCategory(t.fs, "")

	removeOnError := false
	switch t.filename {
	case "", "-":
		w = t.Stdout
	default:
		removeOnError = true
		log.Println("Writing mothball to", t.filename)
		outf, err := t.BaseFs.Create(t.filename)
		if err != nil {
			return err
		}
		defer outf.Close()
		w = outf
	}
	if err := transpile.Mothball(c, w); err != nil {
		if removeOnError {
			t.BaseFs.Remove(t.filename)
		}
		return err
	}
	return nil
}

// CheckAnswer prints whether an answer is correct.
func (t *T) CheckAnswer() error {
	c := transpile.NewFsPuzzle(t.fs)
	log.Print(c.Puzzle())
	log.Print(t.answer)
	_, err := fmt.Fprintf(t.Stdout, `{"Correct":%v}`, c.Answer(t.answer))
	return err
}

func main() {
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
