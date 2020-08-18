package main

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/spf13/afero"
)

func NewTestState() *State {
	s := NewState(new(afero.MemMapFs))
	s.refresh()
	return s
}

func TestState(t *testing.T) {
	s := NewTestState()

	mustExist := func(path string) {
		_, err := s.Fs.Stat(path)
		if os.IsNotExist(err) {
			t.Errorf("File %s does not exist", path)
		}
	}

	pl := s.PointsLog()
	if len(pl) != 0 {
		t.Errorf("Empty points log is not empty")
	}

	mustExist("initialized")
	mustExist("enabled")
	mustExist("hours")

	teamIDsBuf, err := afero.ReadFile(s.Fs, "teamids.txt")
	if err != nil {
		t.Errorf("Reading teamids.txt: %v", err)
	}

	teamIDs := bytes.Split(teamIDsBuf, []byte("\n"))
	if (len(teamIDs) != 101) || (len(teamIDs[100]) > 0) {
		t.Errorf("There weren't 100 teamIDs, there were %d", len(teamIDs))
	}
	teamID := string(teamIDs[0])

	if err := s.SetTeamName("bad team ID", "bad team name"); err == nil {
		t.Errorf("Setting bad team ID didn't raise an error")
	}

	if err := s.SetTeamName(teamID, "My Team"); err != nil {
		t.Errorf("Setting team name: %v", err)
	}

	category := "poot"
	points := 3928
	s.AwardPoints(teamID, category, points)
	s.refresh()

	pl = s.PointsLog()
	if len(pl) != 1 {
		t.Errorf("After awarding points, points log has length %d", len(pl))
	} else if (pl[0].TeamID != teamID) || (pl[0].Category != category) || (pl[0].Points != points) {
		t.Errorf("Incorrect logged award %v", pl)
	}

	s.Fs.Remove("initialized")
	s.refresh()

	pl = s.PointsLog()
	if len(pl) != 0 {
		t.Errorf("After reinitialization, points log has length %d", len(pl))
	}
}

func TestStateEvents(t *testing.T) {
	s := NewTestState()
	s.LogEvent("moo")
	s.LogEventf("moo %d", 2)

	if msg := <-s.eventStream; msg != "moo" {
		t.Error("Wrong message from event stream", msg)
	}
	if msg := <-s.eventStream; msg != "moo 2" {
		t.Error("Formatted event is wrong:", msg)
	}
}

func TestStateMaintainer(t *testing.T) {
	s := NewTestState()
	go s.Maintain(2 * time.Second)

	s.LogEvent("Hello!")
	eventLog, _ := afero.ReadFile(s.Fs, "event.log")
	if len(eventLog) != 12 {
		t.Error("Wrong event log length:", len(eventLog))
	}
}
