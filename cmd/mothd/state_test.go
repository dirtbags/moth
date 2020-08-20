package main

import (
	"bytes"
	"os"
	"strings"
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
	s.LogEvent("moo 2")

	if msg := <-s.eventStream; msg != "moo" {
		t.Error("Wrong message from event stream", msg)
	}
	if msg := <-s.eventStream; msg != "moo 2" {
		t.Error("Formatted event is wrong:", msg)
	}
}

func TestStateMaintainer(t *testing.T) {
	updateInterval := 10 * time.Millisecond

	s := NewTestState()
	go s.Maintain(updateInterval)

	if _, err := s.Stat("initialized"); err != nil {
		t.Error(err)
	}
	teamIDLines, err := afero.ReadFile(s, "teamids.txt")
	if err != nil {
		t.Error(err)
	}
	teamIDList := strings.Split(string(teamIDLines), "\n")
	if len(teamIDList) != 101 {
		t.Error("TeamIDList length is", len(teamIDList))
	}
	teamID := teamIDList[0]
	if len(teamID) < 6 {
		t.Error("Team ID too short:", teamID)
	}

	s.LogEvent("Hello!")

	if len(s.PointsLog()) != 0 {
		t.Error("Points log is not empty")
	}
	if err := s.SetTeamName(teamID, "The Patricks"); err != nil {
		t.Error(err)
	}
	if err := s.AwardPoints(teamID, "pategory", 31337); err != nil {
		t.Error(err)
	}
	time.Sleep(updateInterval)
	pl := s.PointsLog()
	if len(pl) != 1 {
		t.Error("Points log should have one entry")
	}
	if (pl[0].Category != "pategory") || (pl[0].TeamID != teamID) {
		t.Error("Wrong points event was recorded")
	}

	time.Sleep(updateInterval)

	eventLog, err := afero.ReadFile(s.Fs, "event.log")
	if err != nil {
		t.Error(err)
	} else if len(eventLog) != 18 {
		t.Error("Wrong event log length:", len(eventLog))
	}
}
