package main

import (
	"testing"
)

// TestIssue156: if a team has an entry in teams/, but isn't in teamids.txt, we should still let them log in.
func TestIssue156(t *testing.T) {
	h := NewTestHttpd(t, "/")

	if err := h.TestFSStateStore.fsys.WriteFile("test/teams/bloop", []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	h.Refresh()

	ha := h.Authenticate("bloop", "")
	ha.TestRequest(t, "GET", "http://example.com/test/puzzles/cat0/1/", nil, 200)
}
