package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/dirtbags/moth/pkg/transpile"
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
  attachments:
    - filename: moo.txt
---
YAML body
`)

func newTestFs() afero.Fs {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "cat0/1/puzzle.md", testMothYaml, 0644)
	afero.WriteFile(fs, "cat0/1/moo.txt", []byte("Moo."), 0644)
	afero.WriteFile(fs, "cat0/2/puzzle.moth", testMothYaml, 0644)
	afero.WriteFile(fs, "cat0/3/puzzle.moth", testMothYaml, 0644)
	afero.WriteFile(fs, "cat0/4/puzzle.md", testMothYaml, 0644)
	afero.WriteFile(fs, "cat0/5/puzzle.md", testMothYaml, 0644)
	afero.WriteFile(fs, "cat0/10/puzzle.md", testMothYaml, 0644)
	afero.WriteFile(fs, "unbroken/1/puzzle.md", testMothYaml, 0644)
	afero.WriteFile(fs, "unbroken/1/moo.txt", []byte("Moo."), 0644)
	afero.WriteFile(fs, "unbroken/2/puzzle.md", testMothYaml, 0644)
	afero.WriteFile(fs, "unbroken/2/moo.txt", []byte("Moo."), 0644)
	return fs
}

func (tp T) Run(args ...string) error {
	tp.Args = append([]string{"transpile"}, args...)
	command, err := tp.ParseArgs()
	if err != nil {
		return err
	}
	return command()
}

func TestEverything(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	tp := T{
		Stdout: stdout,
		Stderr: stderr,
		BaseFs: newTestFs(),
	}

	if err := tp.Run("inventory"); err != nil {
		t.Error(err)
	}
	if stdout.String() != "cat0 1 2 3 4 5 10\nunbroken 1 2\n" {
		t.Errorf("Bad inventory: %#v", stdout.String())
	}

	stdout.Reset()
	if err := tp.Run("open", "-dir=cat0/1"); err != nil {
		t.Error(err)
	}
	p := transpile.Puzzle{}
	if err := json.Unmarshal(stdout.Bytes(), &p); err != nil {
		t.Error(err)
	}
	if (len(p.Answers) != 1) || (p.Answers[0] != "YAML answer") {
		t.Error("Didn't return the right object", p)
	}

	stdout.Reset()
	if err := tp.Run("open", "-dir=cat0/1", "-file=moo.txt"); err != nil {
		t.Error(err)
	}
	if stdout.String() != "Moo." {
		t.Error("Wrong file pulled", stdout.String())
	}

	stdout.Reset()
	if err := tp.Run("mothball", "-dir=unbroken"); err != nil {
		t.Log(tp.fs)
		t.Error(err)
	}
	if stdout.Len() < 200 {
		t.Error("That's way too short to be a mothball")
	}
	if stdout.String()[:2] != "PK" {
		t.Error("This mothball isn't a zip file!")
	}
}
