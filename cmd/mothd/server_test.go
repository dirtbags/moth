package main

import (
	"io/ioutil"
	"testing"

	"github.com/spf13/afero"
)

const TestTeamID = "teamID"

type TestServer struct {
	*MothServer
}

// NewTestServer creates a new MothServer with NewTestMothballs and some initial state.
//
// See function definition for details.
func NewTestServer() TestServer {
	puzzles := NewTestMothballs()
	puzzles.refresh()

	state := NewTestState()
	afero.WriteFile(state, "teamids.txt", []byte("teamID\n"), 0644)
	afero.WriteFile(state, "messages.html", []byte("messages.html"), 0644)
	state.refresh()

	theme := NewTestTheme()
	afero.WriteFile(theme.Fs, "/index.html", []byte("index.html"), 0644)

	return TestServer{NewMothServer(Configuration{}, theme, state, puzzles)}
}

func (ts TestServer) refresh() {
	ts.State.(*State).refresh()
	for _, pp := range ts.PuzzleProviders {
		pp.(*Mothballs).refresh()
	}
	ts.Theme.(*Theme).refresh()
}

func TestDevelServer(t *testing.T) {
	server := NewTestServer()
	server.Config.Devel = true
	anonHandler := server.NewHandler("badParticipantId", "badTeamId")

	{
		es := anonHandler.ExportState()
		if !es.Config.Devel {
			t.Error("Not marked as development server")
		}
		if len(es.Puzzles) != 1 {
			t.Error("Wrong puzzles for anonymous state on devel server:", es.Puzzles)
		}
	}
}

func TestProdServer(t *testing.T) {
	teamName := "OurTeam"
	participantID := "participantID"
	teamID := TestTeamID

	server := NewTestServer()
	handler := server.NewHandler(participantID, teamID)
	anonHandler := server.NewHandler("badParticipantId", "badTeamId")

	{
		es := handler.ExportState()
		if es.Config.Devel {
			t.Error("Marked as development server", es.Config)
		}
		if len(es.Puzzles) != 0 {
			t.Log("State", es)
			t.Error("Unauthenticated state has non-empty puzzles list")
		}
	}

	if err := handler.Register(teamName); err != nil {
		t.Error(err)
	}
	if err := handler.Register(teamName); err == nil {
		t.Error("Registering twice should have raised an error")
	} else if err != ErrAlreadyRegistered {
		t.Error("Wrong error for duplicate registration:", err)
	}

	if r, _, err := handler.ThemeOpen("/index.html"); err != nil {
		t.Error(err)
	} else if contents, err := ioutil.ReadAll(r); err != nil {
		t.Error(err)
	} else if string(contents) != "index.html" {
		t.Error("index.html wrong contents", contents)
	}

	server.refresh()

	{
		es := handler.ExportState()
		if es.Config.Devel {
			t.Error("Marked as development server", es.Config)
		}
		if len(es.Puzzles) != 1 {
			t.Error("Puzzle categories wrong length", len(es.Puzzles))
		}
		if es.Messages != "messages.html" {
			t.Error("Messages has wrong contents")
		}
		if len(es.PointsLog) != 0 {
			t.Error("Points log not empty")
		}
		if len(es.TeamNames) != 1 {
			t.Error("Wrong number of team names")
		}
		if es.TeamNames["self"] != teamName {
			t.Error("TeamNames['self'] wrong")
		}
	}

	if r, _, err := handler.PuzzlesOpen("pategory", 1, "moo.txt"); err != nil {
		t.Error(err)
	} else if contents, err := ioutil.ReadAll(r); err != nil {
		r.Close()
		t.Error(err)
	} else if string(contents) != "moo" {
		r.Close()
		t.Error("moo.txt has wrong contents", contents)
	} else {
		r.Close()
	}

	if r, _, err := handler.PuzzlesOpen("pategory", 2, "puzzle.json"); err == nil {
		t.Error("Opening locked puzzle shouldn't work")
		r.Close()
	}

	if r, _, err := handler.PuzzlesOpen("pategory", 20, "puzzle.json"); err == nil {
		t.Error("Opening non-existent puzzle shouldn't work")
		r.Close()
	}

	if err := anonHandler.CheckAnswer("pategory", 1, "answer123"); err == nil {
		t.Error("Invalid team ID was able to get points with correct answer")
	}
	if err := handler.CheckAnswer("pategory", 1, "answer123"); err != nil {
		t.Error("Right answer marked wrong", err)
	}

	server.refresh()

	{
		es := handler.ExportState()
		if len(es.PointsLog) != 1 {
			t.Error("I didn't get my points!")
		}
		if len(es.Puzzles["pategory"]) != 2 {
			t.Error("The next puzzle didn't unlock!")
		} else if es.Puzzles["pategory"][1] != 2 {
			t.Error("The 2-point puzzle should have unlocked!")
		}
	}

	if r, _, err := handler.PuzzlesOpen("pategory", 2, "puzzle.json"); err != nil {
		t.Error("Opening unlocked puzzle should work")
	} else {
		r.Close()
	}
	if r, _, err := anonHandler.PuzzlesOpen("pategory", 2, "puzzle.json"); err != nil {
		t.Error("Opening unlocked puzzle anonymously should work")
	} else {
		r.Close()
	}

	if err := handler.CheckAnswer("pategory", 2, "wat"); err != nil {
		t.Error("Right answer marked wrong:", err)
	}

	server.refresh()

	{
		es := anonHandler.ExportState()
		if len(es.TeamNames) != 1 {
			t.Error("Anonymous TeamNames is wrong:", es.TeamNames)
		}
		if len(es.PointsLog) != 2 {
			t.Errorf("Points log wrong length: got %d, wanted 2", len(es.PointsLog))
		} else if es.PointsLog[1].TeamID != "0" {
			t.Error("Second point log didn't anonymize team ID correctly:", es.PointsLog[1])
		}
	}

	{
		es := handler.ExportState()
		if len(es.TeamNames) != 1 {
			t.Error("TeamNames is wrong:", es.TeamNames)
		}
	}

	// BUG(neale): We aren't currently testing the various ways to disable the server
}
