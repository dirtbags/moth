package main

import (
	"os"
	"testing"
)

func TestTranspiler(t *testing.T) {
	p := NewTranspilerProvider(os.DirFS("testdata"))

	inv := p.Inventory()
	if len(inv) != 1 {
		t.Error("Wrong inventory:", inv)
	} else if len(inv[0].Puzzles) != 2 {
		t.Error("Wrong puzzles:", inv)
	}
}
