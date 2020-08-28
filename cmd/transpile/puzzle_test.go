package main

import (
	"testing"

	"github.com/spf13/afero"
)

func TestPuzzle(t *testing.T) {
	puzzleFs := newTestFs()
	catFs := afero.NewBasePathFs(puzzleFs, "cat0")

	p1, err := NewPuzzle(catFs, 1)
	if err != nil {
		t.Error(err)
	}
	t.Log(p1)
	if (len(p1.Answers) == 0) || (p1.Answers[0] != "YAML answer") {
		t.Error("Answers are wrong", p1.Answers)
	}
	if (len(p1.Pre.Authors) != 3) || (p1.Pre.Authors[1] != "Buster") {
		t.Error("Authors are wrong", p1.Pre.Authors)
	}
	if p1.Pre.Body != "<p>YAML body</p>\n" {
		t.Errorf("Body parsed wrong: %#v", p1.Pre.Body)
	}

	p2, err := NewPuzzle(catFs, 2)
	if err != nil {
		t.Error(err)
	}
	if (len(p2.Answers) == 0) || (p2.Answers[0] != "RFC822 answer") {
		t.Error("Answers are wrong", p2.Answers)
	}
	if (len(p2.Pre.Authors) != 3) || (p2.Pre.Authors[1] != "Arthur") {
		t.Error("Authors are wrong", p2.Pre.Authors)
	}
	if p2.Pre.Body != "<p>RFC822 body</p>\n" {
		t.Errorf("Body parsed wrong: %#v", p2.Pre.Body)
	}

	if _, err := NewPuzzle(catFs, 10); err == nil {
		t.Error("Broken YAML didn't trigger an error")
	}
}
