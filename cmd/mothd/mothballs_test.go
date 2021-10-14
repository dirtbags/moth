package main

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/spf13/afero"
)

var testFiles = []struct {
	Name, Body string
}{
	{"puzzles.txt", "1\n3\n2\n"},
	{"answers.txt", "1 answer123\n1 answer456\n2 wat\n"},
	{"1/puzzle.json", `{"name": "moo"}`},
	{"2/puzzle.json", `{}`},
	{"2/moo.txt", `moo`},
	{"3/puzzle.json", `{}`},
	{"3/moo.txt", `moo`},
}

func (m *Mothballs) createMothballWithMoo1(cat string, moo1 string) {
	f, _ := m.Create(fmt.Sprintf("%s.mb", cat))
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	for _, file := range testFiles {
		of, _ := w.Create(file.Name)
		of.Write([]byte(file.Body))
	}

	of, _ := w.Create("1/moo.txt")
	of.Write([]byte(moo1))
}

func (m *Mothballs) createMothball(cat string) {
	m.createMothballWithMoo1(cat, "moo")
}

func NewTestMothballs() *Mothballs {
	m := NewMothballs(new(afero.MemMapFs))
	m.createMothball("pategory")
	m.refresh()
	return m
}

func TestMothballs(t *testing.T) {
	m := NewTestMothballs()
	if _, ok := m.categories["pategory"]; !ok {
		t.Error("Didn't create a new category")
	}

	inv := m.Inventory()
	if len(inv) != 1 {
		t.Error("Wrong inventory size:", inv)
	}
	for _, cat := range inv {
		switch cat.Name {
		case "pategory":
			if len(cat.Puzzles) != 3 {
				t.Error("Puzzles list wrong length")
			}
			if cat.Puzzles[1] != 2 {
				t.Error("Puzzles list not sorted")
			}
		}
		for _, points := range cat.Puzzles {
			f, _, err := m.Open(cat.Name, points, "puzzle.json")
			if err != nil {
				t.Error(cat.Name, err)
				continue
			}
			f.Close()
		}
	}

	if f, _, err := m.Open("nealegory", 1, "puzzle.json"); err == nil {
		f.Close()
		t.Error("You can't open a puzzle in a nealegory, that doesn't even rhyme!")
	}

	if f, _, err := m.Open("pategory", 1, "bozo"); err == nil {
		f.Close()
		t.Error("This file shouldn't exist")
	}

	if ok, _ := m.CheckAnswer("pategory", 1, "answer"); ok {
		t.Error("Wrong answer marked right")
	}
	if _, err := m.CheckAnswer("pategory", 1, "answer123"); err != nil {
		t.Error("Right answer marked wrong", err)
	}
	if _, err := m.CheckAnswer("pategory", 1, "answer456"); err != nil {
		t.Error("Right answer marked wrong", err)
	}
	if ok, err := m.CheckAnswer("nealegory", 1, "moo"); ok {
		t.Error("Checking answer in non-existent category should fail")
	} else if err.Error() != "no such category: nealegory" {
		t.Error("Wrong error message")
	}

	goofyText := "bozonics"
	//time.Sleep(1 * time.Second) // I don't love this, but we need the mtime to increase, and it's only accurate to 1s
	m.createMothballWithMoo1("pategory", goofyText)
	m.refresh()
	if f, _, err := m.Open("pategory", 1, "moo.txt"); err != nil {
		t.Error("pategory/1/moo.txt", err)
	} else if contents, err := ioutil.ReadAll(f); err != nil {
		t.Error("read all pategory/1/moo.txt", err)
	} else if string(contents) != goofyText {
		t.Error("read all replacement pategory/1/moo.txt contents wrong, got", string(contents))
	}

	m.createMothball("test2")
	m.Fs.Remove("pategory.mb")
	m.refresh()
	inv = m.Inventory()
	if len(inv) != 1 {
		t.Error("Deleted mothball is still around", inv)
	}

}
