package main

import (
	"bytes"
	"fmt"
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
	mustExist("hours.txt")

	teamIDsBuf, err := afero.ReadFile(s.Fs, "teamids.txt")
	if err != nil {
		t.Errorf("Reading teamids.txt: %v", err)
	}

	teamIDs := bytes.Split(teamIDsBuf, []byte("\n"))
	if (len(teamIDs) != 101) || (len(teamIDs[100]) > 0) {
		t.Errorf("There weren't 100 teamIDs, there were %d", len(teamIDs))
	}
	teamID := string(teamIDs[0])

	if _, err := s.TeamName(teamID); err == nil {
		t.Errorf("Bad team ID lookup didn't return error")
	}

	if err := s.SetTeamName("bad team ID", "bad team name"); err == nil {
		t.Errorf("Setting bad team ID didn't raise an error")
	}

	teamName := "My Team"
	if err := s.SetTeamName(teamID, teamName); err != nil {
		t.Errorf("Setting team name: %w", err)
	}
	if err := s.SetTeamName(teamID, "wat"); err == nil {
		t.Errorf("Registering team a second time didn't fail")
	}
	if name, err := s.TeamName(teamID); err != nil {
		t.Error(err)
	} else if name != teamName {
		t.Error("Incorrect team name:", name)
	}

	category := "poot"
	points := 3928
	if err := s.AwardPoints(teamID, category, points); err != nil {
		t.Error(err)
	}
	if err := s.AwardPoints(teamID, category, points); err != nil {
		t.Error("Two awards before refresh:", err)
	}
	// Flex duplicate detection with different timestamp
	if f, err := s.Create("points.new/moo"); err != nil {
		t.Error("Creating duplicate points file:", err)
	} else {
		fmt.Fprintln(f, time.Now().Unix()+1, teamID, category, points)
		f.Close()
	}
	s.refresh()

	if err := s.AwardPoints(teamID, category, points); err == nil {
		t.Error("Duplicate points award didn't fail")
	}

	if err := s.AwardPoints(teamID, category, points+1); err != nil {
		t.Error("Awarding more points:", err)
	}

	pl = s.PointsLog()
	if len(pl) != 1 {
		t.Errorf("After awarding points, points log has length %d", len(pl))
	} else if (pl[0].TeamID != teamID) || (pl[0].Category != category) || (pl[0].Points != points) {
		t.Errorf("Incorrect logged award %v", pl)
	}

	afero.WriteFile(s, "points.log", []byte("intentional parse error\n"), 0644)
	if len(s.PointsLog()) != 0 {
		t.Errorf("Intentional parse error breaks pointslog")
	}
	if err := s.AwardPoints(teamID, category, points); err != nil {
		t.Error(err)
	}
	s.refresh()
	if len(s.PointsLog()) != 2 {
		t.Error("Intentional parse error screws up all parsing")
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
	s.LogEvent("moo", "", "", "", 0)
	s.LogEvent("moo 2", "", "", "", 0)

	if msg := <-s.eventStream; msg != "init - - - 0" {
		t.Error("Wrong message from event stream:", msg)
	}
	if msg := <-s.eventStream; msg != "moo - - - 0" {
		t.Error("Wrong message from event stream:", msg)
	}
	if msg := <-s.eventStream; msg != "moo-2 - - - 0" {
		t.Error("Wrong message from event stream:", msg)
	}
}

func TestStateDisabled(t *testing.T) {
	s := NewTestState()
	s.refresh()

	if !s.Enabled {
		t.Error("Brand new state is disabled")
	}

	hoursFile, err := s.Create("hours.txt")
	if err != nil {
		t.Error(err)
	}
	defer hoursFile.Close()

	fmt.Fprintln(hoursFile, "- 1970-01-01T01:01:01Z")
	hoursFile.Sync()
	s.refresh()
	if s.Enabled {
		t.Error("Disabling 1970-01-01")
	}

	fmt.Fprintln(hoursFile, "+ 1970-01-01 01:01:01+05:00")
	hoursFile.Sync()
	s.refresh()
	if !s.Enabled {
		t.Error("Enabling 1970-01-02")
	}

	fmt.Fprintln(hoursFile, "")
	fmt.Fprintln(hoursFile, "# Comment")
	hoursFile.Sync()
	s.refresh()
	if !s.Enabled {
		t.Error("Comments")
	}

	fmt.Fprintln(hoursFile, "intentional parse error")
	hoursFile.Sync()
	s.refresh()
	if !s.Enabled {
		t.Error("intentional parse error")
	}

	fmt.Fprintln(hoursFile, "- 1980-01-01T01:01:01Z")
	hoursFile.Sync()
	s.refresh()
	if s.Enabled {
		t.Error("Disabling 1980-01-01")
	}

	if err := s.Remove("hours.txt"); err != nil {
		t.Error(err)
	}
	s.refresh()
	if !s.Enabled {
		t.Error("Removing `hours.txt` disabled event")
	}

	if err := s.Remove("enabled"); err != nil {
		t.Error(err)
	}
	s.refresh()
	if s.Enabled {
		t.Error("Removing `enabled` didn't disable")
	}

	s.Remove("initialized")
	s.refresh()
	if !s.Enabled {
		t.Error("Re-initializing didn't start event")
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

	s.LogEvent("Hello!", "", "", "", 0)

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

	eventLog, err := afero.ReadFile(s.Fs, "events.log")
	if err != nil {
		t.Error(err)
	} else if events := strings.Split(string(eventLog), "\n"); len(events) != 3 {
		t.Log("Events:", events)
		t.Error("Wrong event log length:", len(events))
	} else if events[2] != "" {
		t.Error("Event log didn't end with newline")
	}
}
