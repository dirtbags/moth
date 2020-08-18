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
	{"content/1/puzzle.json", `{"name": "moo"}`},
	{"content/1/moo.txt", `My cow goes "moo"`},
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

func TestMothballs(t *testing.T) {
	m := NewMothballs(new(afero.MemMapFs))
	m.createMothball("test1")
	m.Update()
	if _, ok := m.categories["test1"]; !ok {
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

	m.createMothball("test2")
	m.Fs.Remove("test1.mb")
	m.Update()
	inv = m.Inventory()
	if len(inv) != 1 {
		t.Error("Deleted mothball is still around", inv)
	}
}
