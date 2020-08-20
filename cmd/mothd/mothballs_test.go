package main

import (
	"archive/zip"
	"fmt"
	"testing"

	"github.com/spf13/afero"
)

var testFiles = []struct {
	Name, Body string
}{
	{"puzzles.txt", "1"},
	{"answers.txt", "1 answer123\n1 answer456\n"},
	{"content/1/puzzle.json", `{"name": "moo"}`},
	{"content/1/moo.txt", `moo`},
}

func (m *Mothballs) createMothball(cat string) {
	f, _ := m.Create(fmt.Sprintf("%s.mb", cat))
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	for _, file := range testFiles {
		of, _ := w.Create(file.Name)
		of.Write([]byte(file.Body))
	}
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
		for _, points := range cat.Puzzles {
			f, _, err := m.Open(cat.Name, points, "puzzle.json")
			if err != nil {
				t.Error(cat.Name, err)
				continue
			}
			f.Close()
		}
	}

	if err := m.CheckAnswer("pategory", 1, "answer"); err == nil {
		t.Error("Wrong answer marked right")
	}
	if err := m.CheckAnswer("pategory", 1, "answer123"); err != nil {
		t.Error("Right answer marked wrong", err)
	}
	if err := m.CheckAnswer("pategory", 1, "answer456"); err != nil {
		t.Error("Right answer marked wrong", err)
	}

	m.createMothball("test2")
	m.Fs.Remove("pategory.mb")
	m.refresh()
	inv = m.Inventory()
	if len(inv) != 1 {
		t.Error("Deleted mothball is still around", inv)
	}

}
