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
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Args   []string
	BaseFs afero.Fs
	fs     afero.Fs
}

// Command is a function invoked by the user
type Command func() error

func nothing() error {
	return nil
}

func usage(w io.Writer) {
	fmt.Fprintln(w, " Usage: transpile mothball [FLAGS] [MOTHBALL]")
	fmt.Fprintln(w, "        Compile a mothball")
	fmt.Fprintln(w, " Usage: inventory [FLAGS]")
	fmt.Fprintln(w, "        Show category inventory")
	fmt.Fprintln(w, " Usage: puzzle [FLAGS]")
	fmt.Fprintln(w, "        Print puzzle JSON")
	fmt.Fprintln(w, " Usage: file [FLAGS] FILENAME")
	fmt.Fprintln(w, "        Open a file for a puzzle")
	fmt.Fprintln(w, " Usage: answer [FLAGS] ANSWER")
	fmt.Fprintln(w, "        Check correctness of an answer")
	fmt.Fprintln(w, " Usage: markdown [FLAGS]")
	fmt.Fprintln(w, "        Format stdin with markdown")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "-dir DIRECTORY")
	fmt.Fprintln(w, "        Use puzzle in DIRECTORY")
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
	case "inventory":
		cmd = t.PrintInventory
	case "puzzle":
		cmd = t.DumpPuzzle
	case "file":
		cmd = t.DumpFile
	case "answer":
		cmd = t.CheckAnswer
	case "markdown":
		cmd = t.Markdown
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
	t.Args = flags.Args()

	return cmd, nil
}

// PrintInventory prints a puzzle inventory to stdout
func (t *T) PrintInventory() error {
	c := transpile.NewFsCategory(t.fs, "")

	inv, err := c.Inventory()
	if err != nil {
		return err
	}
	sort.Ints(inv)
	jinv, err := json.Marshal(
		transpile.InventoryResponse{
			Puzzles: inv,
		},
	)
	if err != nil {
		return err
	}

	t.Stdout.Write(jinv)
	return nil
}

// DumpPuzzle writes a puzzle's JSON to the writer.
func (t *T) DumpPuzzle() error {
	log.Println("Hello!")
	puzzle := transpile.NewFsPuzzle(t.fs)
	log.Println("Hello!")

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
	filename := "puzzle.json"
	if len(t.Args) > 0 {
		filename = t.Args[0]
	}

	puzzle := transpile.NewFsPuzzle(t.fs)

	f, err := puzzle.Open(filename)
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

	filename := ""
	if len(t.Args) == 0 {
		w = t.Stdout
	} else {
		filename = t.Args[0]
		outf, err := t.BaseFs.Create(filename)
		if err != nil {
			return err
		}
		defer outf.Close()
		w = outf
		log.Println("Writing mothball to", filename)
	}

	if err := transpile.Mothball(c, w); err != nil {
		if filename != "" {
			t.BaseFs.Remove(filename)
		}
		return err
	}
	return nil
}

// CheckAnswer prints whether an answer is correct.
func (t *T) CheckAnswer() error {
	answer := ""
	if len(t.Args) > 0 {
		answer = t.Args[0]
	}
	c := transpile.NewFsPuzzle(t.fs)
	_, err := fmt.Fprintf(t.Stdout, `{"Correct":%v}`, c.Answer(answer))
	return err
}

// Markdown runs stdin through a Markdown engine
func (t *T) Markdown() error {
	return transpile.Markdown(t.Stdin, t.Stdout)
}

func main() {
	t := &T{
		Stdin:  os.Stdin,
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
