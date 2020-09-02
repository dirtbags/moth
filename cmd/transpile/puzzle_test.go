package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestPuzzle(t *testing.T) {
	puzzleFs := newTestFs()
	catFs := afero.NewBasePathFs(puzzleFs, "cat0")

	{
		pd := NewPuzzleDir(catFs, 1)
		p, err := pd.Export()
		if err != nil {
			t.Error(err)
		}

		if (len(p.Answers) == 0) || (p.Answers[0] != "YAML answer") {
			t.Error("Answers are wrong", p.Answers)
		}
		if (len(p.Pre.Authors) != 3) || (p.Pre.Authors[1] != "Buster") {
			t.Error("Authors are wrong", p.Pre.Authors)
		}
		if p.Pre.Body != "<p>YAML body</p>\n" {
			t.Errorf("Body parsed wrong: %#v", p.Pre.Body)
		}
	}

	{
		p, err := NewPuzzleDir(catFs, 2).Export()
		if err != nil {
			t.Error(err)
		}
		if (len(p.Answers) == 0) || (p.Answers[0] != "RFC822 answer") {
			t.Error("Answers are wrong", p.Answers)
		}
		if (len(p.Pre.Authors) != 3) || (p.Pre.Authors[1] != "Arthur") {
			t.Error("Authors are wrong", p.Pre.Authors)
		}
		if p.Pre.Body != "<p>RFC822 body</p>\n" {
			t.Errorf("Body parsed wrong: %#v", p.Pre.Body)
		}
	}

	if _, err := NewPuzzleDir(catFs, 3).Export(); err != nil {
		t.Error("Legacy `puzzle.moth` file:", err)
	}

	if _, err := NewPuzzleDir(catFs, 99).Export(); err == nil {
		t.Error("Non-existent puzzle", err)
	}

	if _, err := NewPuzzleDir(catFs, 10).Export(); err == nil {
		t.Error("Broken YAML")
	}
	if _, err := NewPuzzleDir(catFs, 20).Export(); err == nil {
		t.Error("Bad RFC822 header")
	}
	if _, err := NewPuzzleDir(catFs, 21).Export(); err == nil {
		t.Error("Boken RFC822 header")
	}
	if p, err := NewPuzzleDir(catFs, 22).Export(); err == nil {
		t.Error("Duplicate bodies")
	} else if !strings.HasPrefix(err.Error(), "Puzzle body present") {
		t.Log(p)
		t.Error("Wrong error for duplicate body:", err)
	}
}

func TestFsPuzzle(t *testing.T) {
	catFs := afero.NewBasePathFs(afero.NewOsFs(), "testdata")

	if _, err := NewPuzzleDir(catFs, 1).Export(); err != nil {
		t.Error(err)
	}

	if _, err := NewPuzzleDir(catFs, 2).Export(); err != nil {
		t.Error(err)
	}

	mkpuzzleDir := NewPuzzleDir(catFs, 3)
	if _, err := mkpuzzleDir.Export(); err != nil {
		t.Error(err)
	}

	if body, err := mkpuzzleDir.Open("moo.txt"); err != nil {
		t.Error(err)
	} else {
		defer body.Close()
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, body); err != nil {
			t.Error(err)
		}
		if buf.String() != "Moo.\n" {
			t.Error("Wrong body")
		}
	}
}
