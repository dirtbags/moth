package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

var testMothYaml = []byte(`---
answers:
  - YAML answer
pre:
  authors:
    - Arthur
    - Buster
    - DW
---
YAML body
`)
var testMothRfc822 = []byte(`author: test
Author: Arthur
author: Fred Flintstone
answer: RFC822 answer

RFC822 body
`)

func newTestFs() afero.Fs {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "cat0/1/puzzle.md", testMothYaml, 0644)
	afero.WriteFile(fs, "cat0/1/moo.txt", []byte("Moo."), 0644)
	afero.WriteFile(fs, "cat0/2/puzzle.md", testMothRfc822, 0644)
	afero.WriteFile(fs, "cat0/3/puzzle.moth", testMothYaml, 0644)
	afero.WriteFile(fs, "cat0/4/puzzle.md", testMothYaml, 0644)
	afero.WriteFile(fs, "cat0/5/puzzle.md", testMothYaml, 0644)
	afero.WriteFile(fs, "cat0/10/puzzle.md", []byte(`---
Answers:
  - moo
Authors:
  - bad field
---
body
`), 0644)
	afero.WriteFile(fs, "cat0/20/puzzle.md", []byte("Answer: no\nBadField: yes\n\nbody\n"), 0644)
	afero.WriteFile(fs, "cat0/21/puzzle.md", []byte("Answer: broken\nSpooon\n"), 0644)
	afero.WriteFile(fs, "cat0/22/puzzle.md", []byte("---\nanswers:\n  - pencil\npre:\n body: Spooon\n---\nSpoon?\n"), 0644)
	afero.WriteFile(fs, "cat1/93/puzzle.md", []byte("Answer: no\n\nbody"), 0644)
	afero.WriteFile(fs, "cat1/barney/puzzle.md", testMothYaml, 0644)
	return fs
}

func TestEverything(t *testing.T) {
	stdout := new(bytes.Buffer)
	tp := T{
		w:  stdout,
		Fs: newTestFs(),
	}

	if err := tp.Handle("inventory"); err != nil {
		t.Error(err)
	}
	if strings.TrimSpace(stdout.String()) != `{"cat0":[1,2,3,4,5,10,20,21,22],"cat1":[93]}` {
		t.Errorf("Bad inventory: %#v", stdout.String())
	}

	stdout.Reset()
	tp.Cat = "cat0"
	tp.Points = 1
	if err := tp.Handle("open"); err != nil {
		t.Error(err)
	}

	p := Puzzle{}
	if err := json.Unmarshal(stdout.Bytes(), &p); err != nil {
		t.Error(err)
	}
	if (len(p.Answers) != 1) || (p.Answers[0] != "YAML answer") {
		t.Error("Didn't return the right object", p)
	}

	stdout.Reset()
	tp.Filename = "moo.txt"
	if err := tp.Handle("open"); err != nil {
		t.Error(err)
	}
	if stdout.String() != "Moo." {
		t.Error("Wrong file pulled")
	}
}
