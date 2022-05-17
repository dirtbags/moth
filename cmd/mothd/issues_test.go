package main

import (
	"testing"

	"github.com/spf13/afero"
)

func TestIssue156(t *testing.T) {
	puzzles := NewTestMothballs()
	state := NewTestState()
	theme := NewTestTheme()
	server := NewMothServer(Configuration{}, theme, state, puzzles)

	afero.WriteFile(state, "teams/bloop", []byte("bloop: the team"), 0644)
	afero.WriteFile(state, "teamids.txt", []byte("bloop"), 0644)
	state.refresh()

	handler := server.NewHandler("", "bloop")
	es := handler.ExportState()
	if _, ok := es.TeamNames["self"]; !ok {
		t.Log("Team 'bloop' not found in exported state")
		t.Fail()
	}

	err := handler.Register("bloop: the other team")
	if err != ErrAlreadyRegistered {
		t.Logf("Expected ErrAlreadyRegistered when registering team, received %s, instead", err)
		t.Fail()
	}
}
