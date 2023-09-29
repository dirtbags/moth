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

func slurp(c chan bool) {
	for range c {
		// Nothing
	}
}

func TestState(t *testing.T) {
	s := NewTestState()
	defer close(s.refreshNow)
	go slurp(s.refreshNow)

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
		t.Errorf("Setting team name: %v", err)
	}
	if err := s.SetTeamName(teamID, "wat"); err == nil {
		t.Errorf("Registering team a second time didn't fail")
	}
	s.refresh()
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
	// Flex duplicate detection with different timestamp
	if f, err := s.Create("points.new/moo"); err != nil {
		t.Error("Creating duplicate points file:", err)
	} else {
		fmt.Fprintln(f, time.Now().Unix()+1, teamID, category, points)
		f.Close()
	}

	s.AwardPoints(teamID, category, points)
	s.refresh()
	pl = s.PointsLog()
	if len(pl) != 1 {
		for i, award := range pl {
			t.Logf("pl[%d] == %s", i, award.String())
		}
		t.Errorf("After awarding duplicate points, points log has length %d", len(pl))
	} else if (pl[0].TeamID != teamID) || (pl[0].Category != category) || (pl[0].Points != points) {
		t.Errorf("Incorrect logged award %v", pl)
	}

	if err := s.AwardPoints(teamID, category, points); err == nil {
		t.Error("Duplicate points award after refresh didn't fail")
	}

	if err := s.AwardPoints(teamID, category, points+1); err != nil {
		t.Error("Awarding more points:", err)
	}

	s.refresh()
	if len(s.PointsLog()) != 2 {
		t.Errorf("There should be two awards")
	}

	afero.WriteFile(s, "points.log", []byte("intentional parse error\n"), 0644)
	s.refresh()
	if len(s.PointsLog()) != 0 {
		t.Errorf("Intentional parse error breaks pointslog")
	}
	if err := s.AwardPoints(teamID, category, points); err != nil {
		t.Error(err)
	}
	s.refresh()
	if len(s.PointsLog()) != 1 {
		t.Log(s.PointsLog())
		t.Error("Intentional parse error screws up all parsing")
	}

	s.Fs.Remove("initialized")
	s.refresh()

	pl = s.PointsLog()
	if len(pl) != 0 {
		t.Errorf("After reinitialization, points log has length %d", len(pl))
	}

}

// Out of order points insertion, issue #168
func TestStateOutOfOrderAward(t *testing.T) {
	s := NewTestState()

	category := "meow"
	points := 100

	now := time.Now().Unix()
	if err := s.awardPointsAtTime(now+20, "AA", category, points); err != nil {
		t.Error("Awarding points to team ZZ:", err)
	}
	if err := s.awardPointsAtTime(now+10, "ZZ", category, points); err != nil {
		t.Error("Awarding points to team AA:", err)
	}
	s.refresh()
	pl := s.PointsLog()
	if len(pl) != 2 {
		t.Error("Wrong length for points log")
	}
	if pl[0].TeamID != "ZZ" {
		t.Error("Out of order points insertion not properly sorted in points log")
	}
}

func TestStateEvents(t *testing.T) {
	s := NewTestState()
	s.LogEvent("moo", "", "", 0)
	s.LogEvent("moo 2", "", "", 0)

	if msg := <-s.eventStream; strings.Join(msg[1:], ":") != "init:::0" {
		t.Error("Wrong message from event stream:", msg)
	}
	if msg := <-s.eventStream; !strings.HasPrefix(msg[5], "state/hours.txt") {
		t.Error("Wrong message from event stream:", msg[5])
	}
	if msg := <-s.eventStream; strings.Join(msg[1:], ":") != "moo:::0" {
		t.Error("Wrong message from event stream:", msg)
	}
	if msg := <-s.eventStream; strings.Join(msg[1:], ":") != "moo 2:::0" {
		t.Error("Wrong message from event stream:", msg)
	}
}

func TestStateDisabled(t *testing.T) {
	s := NewTestState()
	s.refresh()

	if !s.Enabled() {
		t.Error("Brand new state is disabled")
	}

	hoursFile, err := s.Create("hours.txt")
	if err != nil {
		t.Error(err)
	}
	defer hoursFile.Close()
	s.refresh()
	if !s.Enabled() {
		t.Error("Empty hours.txt not enabled")
	}

	fmt.Fprintln(hoursFile, "- 1970-01-01T01:01:01Z")
	hoursFile.Sync()
	s.refresh()
	if s.Enabled() {
		t.Error("1970-01-01")
	}

	fmt.Fprintln(hoursFile, "+ 1970-01-02 01:01:01+05:00")
	hoursFile.Sync()
	s.refresh()
	if !s.Enabled() {
		t.Error("1970-01-02")
	}

	fmt.Fprintln(hoursFile, "-")
	hoursFile.Sync()
	s.refresh()
	if s.Enabled() {
		t.Error("bare -")
	}

	fmt.Fprintln(hoursFile, "+")
	hoursFile.Sync()
	s.refresh()
	if !s.Enabled() {
		t.Error("bare +")
	}

	fmt.Fprintln(hoursFile, "")
	fmt.Fprintln(hoursFile, "# Comment")
	hoursFile.Sync()
	s.refresh()
	if !s.Enabled() {
		t.Error("Comment")
	}

	fmt.Fprintln(hoursFile, "intentional parse error")
	hoursFile.Sync()
	s.refresh()
	if !s.Enabled() {
		t.Error("intentional parse error")
	}

	fmt.Fprintln(hoursFile, "- 1980-01-01T01:01:01Z")
	hoursFile.Sync()
	s.refresh()
	if s.Enabled() {
		t.Error("1980-01-01")
	}

	if err := s.Remove("hours.txt"); err != nil {
		t.Error(err)
	}
	s.refresh()
	if !s.Enabled() {
		t.Error("Removing `hours.txt` disabled event")
	}

	s.Remove("initialized")
	s.refresh()
	if !s.Enabled() {
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

	s.LogEvent("Hello!", "", "", 0)

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

	eventLog, err := afero.ReadFile(s.Fs, "events.csv")
	if err != nil {
		t.Error(err)
	} else if events := strings.Split(string(eventLog), "\n"); len(events) != 4 {
		t.Log("Events:", events)
		t.Error("Wrong event log length:", len(events))
	} else if events[3] != "" {
		t.Error("Event log didn't end with newline", events)
	}
}

func TestDevelState(t *testing.T) {
	s := NewTestState()
	ds := NewDevelState(s)
	if err := ds.SetTeamName("boog", "The Boog Team"); err != ErrAlreadyRegistered {
		t.Error("Registering a team that doesn't exist", err)
	} else if err == nil {
		t.Error("Registering a team that doesn't exist didn't return ErrAlreadyRegistered")
	}
	if n, err := ds.TeamName("boog"); err != nil {
		t.Error("Devel State returned error on team name lookup")
	} else if n != "«devel:boog»" {
		t.Error("Wrong team name", n)
	}

	if err := ds.AwardPoints("blerg", "dog", 82); err != nil {
		t.Error("Devel State AwardPoints returned an error", err)
	}
}
