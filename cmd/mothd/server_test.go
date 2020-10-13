package main

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/spf13/afero"
)

const TestMaintenanceInterval = time.Millisecond * 1
const TestTeamID = "teamID"

func NewTestServer() *MothServer {
	puzzles := NewTestMothballs()
	go puzzles.Maintain(TestMaintenanceInterval)

	state := NewTestState()
	afero.WriteFile(state, "teamids.txt", []byte("teamID\n"), 0644)
	afero.WriteFile(state, "messages.html", []byte("messages.html"), 0644)
	go state.Maintain(TestMaintenanceInterval)

	theme := NewTestTheme()
	afero.WriteFile(theme.Fs, "/index.html", []byte("index.html"), 0644)
	go theme.Maintain(TestMaintenanceInterval)

	return NewMothServer(Configuration{}, theme, state, puzzles)
}

func TestServer(t *testing.T) {
	teamName := "OurTeam"
	participantID := "participantID"
	teamID := TestTeamID

	server := NewTestServer()
	handler := server.NewHandler(participantID, teamID)

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
	if r, _, err := handler.ThemeOpen("/index.html"); err != nil {
		t.Error(err)
	} else if contents, err := ioutil.ReadAll(r); err != nil {
		t.Error(err)
	} else if string(contents) != "index.html" {
		t.Error("index.html wrong contents", contents)
	}

	es := handler.ExportState()
	if es.Config.Devel {
		t.Error("Marked as development server", es.Config)
	}
	if len(es.Puzzles) != 1 {
		t.Error("Puzzle categories wrong length")
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

	if r, _, err := handler.PuzzlesOpen("pategory", 2, "puzzles.json"); err == nil {
		t.Error("Opening locked puzzle shouldn't work")
		r.Close()
	}

	if r, _, err := handler.PuzzlesOpen("pategory", 20, "puzzles.json"); err == nil {
		t.Error("Opening non-existent puzzle shouldn't work")
		r.Close()
	}

	if err := handler.CheckAnswer("pategory", 1, "answer123"); err != nil {
		t.Error("Right answer marked wrong", err)
	}

	time.Sleep(TestMaintenanceInterval)

	es = handler.ExportState()
	if len(es.PointsLog) != 1 {
		t.Error("I didn't get my points!")
	}

	// BUG(neale): We aren't currently testing the various ways to disable the server
}
