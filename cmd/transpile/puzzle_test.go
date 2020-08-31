package main

import (
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestPuzzle(t *testing.T) {
	puzzleFs := newTestFs()
	catFs := afero.NewBasePathFs(puzzleFs, "cat0")

	{
		p, err := NewPuzzle(catFs, 1)
		if err != nil {
			t.Error(err)
		}
		t.Log(p)
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
		p, err := NewPuzzle(catFs, 2)
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

	if _, err := NewPuzzle(catFs, 3); err != nil {
		t.Error("Legacy `puzzle.moth` file:", err)
	}

	if _, err := NewPuzzle(catFs, 99); err == nil {
		t.Error("Non-existent puzzle", err)
	}

	if _, err := NewPuzzle(catFs, 10); err == nil {
		t.Error("Broken YAML")
	}
	if _, err := NewPuzzle(catFs, 20); err == nil {
		t.Error("Bad RFC822 header")
	}
	if _, err := NewPuzzle(catFs, 21); err == nil {
		t.Error("Boken RFC822 header")
	}
	if _, err := NewPuzzle(catFs, 22); err == nil {
		t.Error("Duplicate bodies")
	} else if !strings.HasPrefix(err.Error(), "Puzzle body present") {
		t.Error("Wrong error for duplicate body:", err)
	}
}

func TestFsPuzzle(t *testing.T) {
	catFs := afero.NewBasePathFs(afero.NewOsFs(), "testdata")

	if _, err := NewPuzzle(catFs, 1); err != nil {
		t.Error(err)
	}

	if _, err := NewPuzzle(catFs, 2); err != nil {
		t.Error(err)
	}
}
