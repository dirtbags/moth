package main

import (
	"bytes"
	"io"
	"testing"

	"github.com/spf13/afero"
)

func TestFsCategory(t *testing.T) {
	c := NewFsCategory(newTestFs(), "cat0")

	if inv, err := c.Inventory(); err != nil {
		t.Error(err)
	} else if len(inv) != 9 {
		t.Error("Inventory wrong length", inv)
	}

	if p, err := c.Puzzle(1); err != nil {
		t.Error(err)
	} else if len(p.Answers) != 1 {
		t.Error("Wrong length for answers", p.Answers)
	} else if p.Answers[0] != "YAML answer" {
		t.Error("Wrong answer list", p.Answers)
	} else if !c.Answer(1, p.Answers[0]) {
		t.Error("Correct answer not accepted")
	}

	if c.Answer(1, "incorrect answer") {
		t.Error("Incorrect answer accepted as correct")
	}

	if r, err := c.Open(1, "moo.txt"); err != nil {
		t.Log(c.Puzzle(1))
		t.Error(err)
	} else {
		defer r.Close()
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, r); err != nil {
			t.Error(err)
		}
		if buf.String() != "Moo." {
			t.Error("Opened file contents wrong")
		}
	}

	if r, err := c.Open(1, "error"); err == nil {
		r.Close()
		t.Error("File wasn't supposed to exist")
	}
}

func TestOsFsCategory(t *testing.T) {
	fs := NewRecursiveBasePathFs(afero.NewOsFs(), "testdata")
	static := NewFsCategory(fs, "static")

	if p, err := static.Puzzle(1); err != nil {
		t.Error(err)
	} else if len(p.Pre.Authors) != 1 {
		t.Error("Wrong authors list", p.Pre.Authors)
	} else if p.Pre.Authors[0] != "neale" {
		t.Error("Wrong authors", p.Pre.Authors)
	}

	generated := NewFsCategory(fs, "generated")

	if inv, err := generated.Inventory(); err != nil {
		t.Error(err)
	} else if len(inv) != 5 {
		t.Error("Wrong inventory", inv)
	}

	if p, err := generated.Puzzle(1); err != nil {
		t.Error(err)
	} else if len(p.Answers) != 1 {
		t.Error("Wrong answers", p.Answers)
	} else if p.Answers[0] != "answer1.0" {
		t.Error("Wrong answers:", p.Answers)
	}
	if _, err := generated.Puzzle(20); err == nil {
		t.Error("Puzzle shouldn't exist")
	}

	if r, err := generated.Open(1, "moo.txt"); err != nil {
		t.Error(err)
	} else {
		defer r.Close()
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, r); err != nil {
			t.Error(err)
		}
		if buf.String() != "Moo.\n" {
			t.Errorf("Wrong body: %#v", buf.String())
		}
	}
	if r, err := generated.Open(1, "fail"); err == nil {
		r.Close()
		t.Error("File shouldn't exist")
	}

	if !generated.Answer(1, "answer1.0") {
		t.Error("Correct answer failed")
	}
	if generated.Answer(1, "wrong") {
		t.Error("Incorrect answer didn't fail")
	}
	if generated.Answer(2, "error") {
		t.Error("Error answer didn't fail")
	}
}
