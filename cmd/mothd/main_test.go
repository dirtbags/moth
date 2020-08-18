package main

import (
	"testing"
)

func TestEverything(t *testing.T) {
	state := NewTestState()
	t.Error("No test")

	state.refresh()
}
