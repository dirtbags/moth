package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/dirtbags/moth/pkg/transpile"
	"github.com/spf13/afero"
)

var testMothYaml = []byte(`---
answers:
  - YAML answer
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

func TestTranspilerEverything(t *testing.T) {
	stdin := new(bytes.Buffer)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	tp := T{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		BaseFs: newTestFs(),
	}

	if err := tp.Run("inventory", "-dir=cat0"); err != nil {
		t.Error(err)
	}
	if stdout.String() != "{\"Puzzles\":[1,2,3,4,5,10]}" {
		t.Errorf("Bad inventory: %#v", stdout.String())
	}

	stdout.Reset()
	if err := tp.Run("puzzle", "-dir=cat0/1"); err != nil {
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
	if err := tp.Run("file", "-dir=cat0/1", "moo.txt"); err != nil {
		t.Error(err)
	}
	if stdout.String() != "Moo." {
		t.Error("Wrong file pulled", stdout.String())
	}

	stdout.Reset()
	if err := tp.Run("answer", "-dir=cat0/1", "YAML answer"); err != nil {
		t.Error(err)
	}
	if stdout.String() != `{"Correct":true}` {
		t.Error("Answer validation failed", stdout.String())
	}

	stdout.Reset()
	stdin.Reset()
	stdin.WriteString("text *emphasized* text")
	if err := tp.Run("markdown"); err != nil {
		t.Error(err)
	}
	if stdout.String() != "<p>text <em>emphasized</em> text</p>\n" {
		t.Error("Markdown conversion failed", stdout.String())
	}
}

func TestMothballs(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	tp := T{
		Stdout: stdout,
		Stderr: stderr,
		BaseFs: newTestFs(),
	}

	stdout.Reset()
	if err := tp.Run("mothball", "-dir=unbroken", "unbroken.mb"); err != nil {
		t.Error(err)
		return
	}

	// afero.WriteFile(tp.BaseFs, "unbroken.mb", []byte("moo"), 0644)
	fis, err := afero.ReadDir(tp.BaseFs, "/")
	if err != nil {
		t.Error(err)
	}
	for _, fi := range fis {
		t.Log(fi.Name())
	}

	mb, err := tp.BaseFs.Open("unbroken.mb")
	if err != nil {
		t.Error(err)
		return
	}
	defer mb.Close()

	info, err := mb.Stat()
	if err != nil {
		t.Error(err)
		return
	}

	zmb, err := zip.NewReader(mb, info.Size())
	if err != nil {
		t.Error(err)
		return
	}
	for _, zf := range zmb.File {
		f, err := zf.Open()
		if err != nil {
			t.Error(err)
			continue
		}
		defer f.Close()
		buf, err := ioutil.ReadAll(f)
		if err != nil {
			t.Error(err)
			continue
		}

		switch zf.Name {
		case "answers.txt":
			if len(buf) == 0 {
				t.Error("answers.txt empty")
			}
		case "puzzles.txt":
			if len(buf) == 0 {
				t.Error("puzzles.txt empty")
			}
		}
	}
}

func TestFilesystem(t *testing.T) {
	stdin := new(bytes.Buffer)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	tp := T{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		BaseFs: afero.NewOsFs(),
	}

	stdout.Reset()
	if err := tp.Run("puzzle", "-dir=testdata/cat1/1"); err != nil {
		t.Error(err)
	}
	if !strings.Contains(stdout.String(), "moo") {
		t.Error("File not pulled from cwd", stdout.String())
	}

	stdout.Reset()
	if err := tp.Run("file", "-dir=testdata/cat1/1", "moo.txt"); err != nil {
		t.Error(err)
	}
	if !strings.Contains(stdout.String(), "Moo.") {
		t.Error("Wrong file pulled", stdout.String())
	}
}

func TestCwd(t *testing.T) {
	testwd, err := os.Getwd()
	if err != nil {
		t.Error("Can't get current working directory!")
		return
	}
	defer os.Chdir(testwd)

	stdin := new(bytes.Buffer)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	tp := T{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		BaseFs: afero.NewOsFs(),
	}

	stdout.Reset()
	os.Chdir("/")
	if err := tp.Run(
		"file",
		fmt.Sprintf("-dir=%s/testdata/cat1/1", testwd),
		"moo.txt",
	); err != nil {
		t.Error(err)
	}
}
