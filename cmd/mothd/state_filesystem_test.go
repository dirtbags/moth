package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	
	"github.com/dirtbags/moth/pkg/award"
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

	if (! s.PointExists(teamID, category, points)) {
		t.Errorf("Unable to find points %s/%d for team %s", category, points, teamID)
	}

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

func TestStatePointsRemoval(t *testing.T) {
	s := NewTestState()
	s.refresh()

	team := "team1"
	category := "meow"
	points1 := 100
	points2 := points1 + 1

	// Add points into our log
	if err := s.AwardPoints(team, category, points1); err != nil {
		t.Errorf("Received unexpected error when awarding points: %s", err)
	}
	s.refresh()

	pointsLogLength := len(s.PointsLog())
	if pointsLogLength != 1 {
		t.Errorf("Expected 1 point in the log after awarding, got %d", pointsLogLength)
	}

	if err := s.AwardPoints(team, category, points2); err != nil {
		t.Errorf("Received unexpected error when awarding points: %s", err)
	}
	s.refresh()

	pointsLogLength = len(s.PointsLog())
	if pointsLogLength != 2 {
		t.Errorf("Expected 2 points in the log after awarding, got %d", pointsLogLength)
	}

	// Remove a point
	if err := s.RemovePoints(team, category, points1); err != nil {
		t.Errorf("Received unexpected error when removing points1: %s", err)
	}
	s.refresh()

	pointsLog := s.PointsLog()
	pointsLogLength = len(pointsLog)
	if pointsLogLength != 1 {
		t.Errorf("Expected 1 point in the log after removal, got %d", pointsLogLength)
	}

	if ((pointsLog[0].TeamID != team) || (pointsLog[0].Category != category) || (pointsLog[0].Points != points2)) {
		t.Errorf("Found unexpected points log entry after removal: %s", pointsLog[0])
	}

	// Remove a duplicate point
	if err := s.RemovePoints(team, category, points1); err != NoMatchingPointEntry {
		t.Errorf("Expected to receive NoMatchingPointEntry, received error '%s', instead", err)
	}
	s.refresh()

	pointsLog = s.PointsLog()
	pointsLogLength = len(pointsLog)
	if pointsLogLength != 1 {
		t.Errorf("Expected 1 point in the log after duplicate removal, got %d", pointsLogLength)
	}

	// Remove the second point
	if err := s.RemovePoints(team, category, points2); err != nil {
		t.Errorf("Received unexpected error when removing points2: %s", err)
	}
	s.refresh()

	pointsLog = s.PointsLog()
	pointsLogLength = len(pointsLog)
	if pointsLogLength != 0 {
		t.Errorf("Expected 0 point in the log after last removal, got %d", pointsLogLength)
	}
}

func TestStatePointsRemovalAtTime(t *testing.T) {
	s := NewTestState()
	s.refresh()

	team := "team1"
	category := "meow"
	points1 := 100
	points2 := points1 + 1
	now := time.Now().Unix()
	time1 := now
	time2 := now+10


	s.AwardPointsAtTime(team, category, points1, time1)
	s.refresh()

	pointsLogLength := len(s.PointsLog())
	if pointsLogLength != 1 {
		t.Errorf("Expected 1 point in the log, got %d", pointsLogLength)
	}

	pointsLog := s.PointsLog()
	if ((pointsLog[0].When != time1) || (pointsLog[0].TeamID != team) || (pointsLog[0].Category != category) || (pointsLog[0].Points != points1)) {
		t.Errorf("Received unexpected points entry: %s", pointsLog[0])
	}

	s.AwardPointsAtTime(team, category, points2, time2)
	s.refresh()

	pointsLogLength = len(s.PointsLog())
	if pointsLogLength != 2 {
		t.Errorf("Expected 2 point in the log, got %d", pointsLogLength)
	}

	// Remove valid points, but at wrong time
	s.RemovePointsAtTime(team, category, points1, time2)
	s.refresh()

	pointsLogLength = len(s.PointsLog())
	if pointsLogLength != 2 {
		t.Errorf("Expected 2 point in the log, got %d", pointsLogLength)
	}

	s.RemovePointsAtTime(team, category, points1, time1)
	s.refresh()

	pointsLog = s.PointsLog()
	pointsLogLength = len(pointsLog)
	if pointsLogLength != 1 {
		t.Errorf("Expected 1 point in the log, got %d", pointsLogLength)
	}

	if ((pointsLog[0].When != time2) || (pointsLog[0].TeamID != team) || (pointsLog[0].Category != category) || (pointsLog[0].Points != points2)) {
		t.Errorf("Found unexpected points log entry: %s", pointsLog[0])
	}

	s.RemovePointsAtTime(team, category, points1, time1)
	s.refresh()

	pointsLog = s.PointsLog()
	pointsLogLength = len(pointsLog)
	if pointsLogLength != 1 {
		t.Errorf("Expected 1 point in the log, got %d", pointsLogLength)
	}

	s.RemovePointsAtTime(team, category, points2, time2)
	s.refresh()

	pointsLog = s.PointsLog()
	pointsLogLength = len(pointsLog)
	if pointsLogLength != 0 {
		t.Errorf("Expected 0 point in the log, got %d", pointsLogLength)
	}
}

func TestStateSetPoints(t *testing.T) {
	s := NewTestState()
	s.refresh()

	team := "team1"
	category := "meow"
	points := 100
	time := time.Now().Unix()


	// Add points into our log
	if err := s.AwardPointsAtTime(team, category, points, time); err != nil {
		t.Errorf("Received unexpected error when awarding points: %s", err)
	}
	s.refresh()

	pointsLog := s.PointsLog()
	pointsLogLength := len(pointsLog)
	if pointsLogLength != 1 {
		t.Errorf("Expected 1 point in the log after awarding, got %d", pointsLogLength)
	}

	expectedPoints := make(award.List, pointsLogLength)
	copy( expectedPoints, pointsLog)

	s.SetPoints(make(award.List, 0))
	s.refresh()

	pointsLog = s.PointsLog()
	pointsLogLength = len(pointsLog)
	if pointsLogLength != 0 {
		t.Errorf("Expected 0 point in the log after awarding, got %d", pointsLogLength)
	}

	if err := s.SetPoints(expectedPoints); err != nil {
		t.Errorf("Received unexpected error when awarding points: %s", err)
	}
	s.refresh()

	pointsLog = s.PointsLog()
	pointsLogLength = len(pointsLog)
	if pointsLogLength != 1 {
		t.Errorf("Expected 1 point in the log after awarding, got %d", pointsLogLength)
	} else if (expectedPoints[0] != pointsLog[0]) {
		t.Errorf("Expected first point '%s', received '%s', instead", expectedPoints[0], pointsLog[0])
	}
}

// Out of order points insertion, issue #168
func TestStateOutOfOrderAward(t *testing.T) {
	s := NewTestState()

	category := "meow"
	points := 100

	now := time.Now().Unix()
	if err := s.AwardPointsAtTime("AA", category, points, now+20); err != nil {
		t.Error("Awarding points to team ZZ:", err)
	}
	if err := s.AwardPointsAtTime("ZZ", category, points, now+10); err != nil {
		t.Error("Awarding points to team AA:", err)
	}
	s.refresh()

	if (! s.PointExistsAtTime("AA", category, points, now+20)) {
		t.Errorf("Unable to find points awarded to team AA for %s/%d at %d", category, points, now+20)
	}

	if (! s.PointExistsAtTime("ZZ", category, points, now+10)) {
		t.Errorf("Unable to find points awarded to team ZZ for %s/%d at %d", category, points, now+10)
	}

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
	s.LogEvent("moo", "", "", "", 0)
	s.LogEvent("moo 2", "", "", "", 0)

	if msg := <-s.eventStream; strings.Join(msg[1:], ":") != "init::::0" {
		t.Error("Wrong message from event stream:", msg)
	}
	if msg := <-s.eventStream; strings.Join(msg[1:], ":") != "moo::::0" {
		t.Error("Wrong message from event stream:", msg)
	}
	if msg := <-s.eventStream; strings.Join(msg[1:], ":") != "moo 2::::0" {
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

	eventLog, err := afero.ReadFile(s.Fs, "events.csv")
	if err != nil {
		t.Error(err)
	} else if events := strings.Split(string(eventLog), "\n"); len(events) != 3 {
		t.Log("Events:", events)
		t.Error("Wrong event log length:", len(events))
	} else if events[2] != "" {
		t.Error("Event log didn't end with newline")
	}
}

func TestMessage(t *testing.T) {
	s := NewTestState()
	s.refresh()

	message := "foobar"

	if err:= s.SetMessages(message); err != nil {
		t.Error(err)
	}

	s.refresh()

	retrievedMessage := s.Messages()

	if (retrievedMessage != message) {
		t.Errorf("Expected message '%s', received '%s', instead", message, retrievedMessage)
	}
}

func TestStateTeamIDs(t *testing.T) {
	s := NewTestState()
	s.refresh()

	emptyTeams := make([]string, 0)
	teamID1 := "foobar"
	teamID2 := "foobaz"

	// Verify we can pull the initial list without error
	if _, err := s.TeamIDs(); err != nil {
		t.Errorf("Received unexpected error %s", err)
	}

	// Can we set team IDs to an empty list?
	if err := s.SetTeamIDs(emptyTeams); err != nil {
		t.Errorf("Received unexpected error %s", err)
	}

	if teamIDs, err := s.TeamIDs(); err != nil {
		t.Errorf("Received unexpected error %s", err)
	} else {
		if len(teamIDs) != 0 {
			t.Errorf("Expected to find 0 team IDs, found %d (%s), instead", len(teamIDs), teamIDs)
		}
	}

	// Check if an ID exists in an empty list
	if teamIDExists, err := s.TeamIDExists(teamID1); err != nil {
		t.Errorf("Received unexpected error %s", err)
		
		if teamIDExists {
			t.Errorf("Expected to receive false, since team ID list should be empty, but received true, instead")
		}
	}

	

	// Add a team ID
	if err := s.AddTeamID(teamID1); err != nil {
		t.Errorf("Received unexpected error %s", err)
	}
	s.refresh()

	if teamIDs, err := s.TeamIDs(); err != nil {
		t.Errorf("Received unexpected error %s", err)
	} else {
		if len(teamIDs) != 1 {
			t.Errorf("Expected to find 1 team ID, found %d (%s), instead", len(teamIDs), teamIDs)
		} else {
			if teamIDs[0] != teamID1 {
				t.Errorf("Expected to find team ID '%s', found '%s', instead", teamID1, teamIDs[0])
			}
		}
	}

	// Add another team ID
	if err := s.AddTeamID(teamID2); err != nil {
		t.Errorf("Received unexpected error %s", err)
	}
	s.refresh()

	if teamIDs, err := s.TeamIDs(); err != nil {
		t.Errorf("Received unexpected error %s", err)
	} else {
		if len(teamIDs) != 2 {
			t.Errorf("Expected to find 2 team IDs, found %d (%s), instead", len(teamIDs), teamIDs)
		} else {
			if exists1, err1 := s.TeamIDExists(teamID1); err1 != nil {
				t.Errorf("Received unexpected error %s", err)
			} else {
				if ! exists1 {
					t.Errorf("Expected to find team ID '%s', but didn't find it", teamID1)
				}
			}
		}
	}

	// Add a duplicate team ID
	if err := s.AddTeamID(teamID2); err == nil {
		t.Errorf("Expected to raise error, received nil, instead")
	}
	s.refresh()

	if teamIDs, err := s.TeamIDs(); err != nil {
		t.Errorf("Received unexpected error %s", err)
	} else {
		if len(teamIDs) != 2 {
			t.Errorf("Expected to find 2 team IDs, found %d (%s), instead", len(teamIDs), teamIDs)
		} else {
			if exists1, err1 := s.TeamIDExists(teamID1); err1 != nil {
				t.Errorf("Received unexpected error %s", err)
			} else {
				if ! exists1 {
					t.Errorf("Expected to find team ID '%s', but didn't find it", teamID1)
				}
			}
		}
	}

	// Remove a team ID
	if err := s.RemoveTeamID(teamID1); err != nil {
		t.Errorf("Received unexpected error %s", err)
	}
	s.refresh()

	if teamIDs, err := s.TeamIDs(); err != nil {
		t.Errorf("Received unexpected error %s", err)
	} else {
		if len(teamIDs) != 1 {
			t.Errorf("Expected to find 1 team ID, found %d (%s), instead", len(teamIDs), teamIDs)
		} else {
			if exists2, err2 := s.TeamIDExists(teamID2); err2 != nil {
				t.Errorf("Received unexpected error: %s", err2)
			} else if (! exists2) {
				t.Errorf("Expected to find team ID '%s', but didn't find it", teamID2)
			}
		}
	}

	// Remove the last team ID
	if err := s.RemoveTeamID(teamID2); err != nil {
		t.Errorf("Received unexpected error %s", err)
	}
	s.refresh()

	if teamIDs, err := s.TeamIDs(); err != nil {
		t.Errorf("Received unexpected error %s", err)
	} else {
		if len(teamIDs) != 0 {
			t.Errorf("Expected to find 0 team ID, found %d (%s), instead", len(teamIDs), teamIDs)
		} 
	}
}

func TestStateDeleteTeamIDList(t *testing.T) {
	s := NewTestState()
	s.refresh()

	s.Fs.Remove("teamids.txt")

	teamIDs, err := s.TeamIDs()

	if len(teamIDs) != 0 {
		t.Errorf("Expected to find 0 team IDs, found %d (%s), instead", len(teamIDs), teamIDs)
	}

	if err == nil {
		t.Errorf("Did not receive expected error for non-existent teamids.txt")
	}
}

func TestStateTeamNames(t *testing.T) {
	s := NewTestState()
	s.refresh()

	if teamNames := s.TeamNames(); len(teamNames) != 0 {
		t.Errorf("Expected to find 0 registered teams, found %d (%s), instead", len(teamNames), teamNames)
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
