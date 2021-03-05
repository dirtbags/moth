package main

import (
	"testing"

	"github.com/spf13/afero"
)

func TestTranspiler(t *testing.T) {
	fs := afero.NewBasePathFs(afero.NewOsFs(), "testdata")
	p := NewTranspilerProvider(fs)

	inv := p.Inventory()
	if len(inv) != 1 {
		t.Error("Wrong inventory:", inv)
	} else if len(inv[0].Puzzles) != 2 {
		t.Error("Wrong puzzles:", inv)
	}
}
