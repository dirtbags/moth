package main

import (
  "bytes"
	"github.com/spf13/afero"
	"os"
	"testing"
)

func TestState(t *testing.T) {
	fs := new(afero.MemMapFs)

	mustExist := func(path string) {
		_, err := fs.Stat(path)
		if os.IsNotExist(err) {
			t.Errorf("File %s does not exist", path)
		}
	}

	s := NewState(fs)
	s.Cleanup()

  pl := s.PointsLog()
  if len(pl) != 0 {
    t.Errorf("Empty points log is not empty")
  }

	mustExist("initialized")
	mustExist("enabled")
	mustExist("hours")
	
	teamidsBuf, err := afero.ReadFile(fs, "teamids.txt")
	if err != nil {
	  t.Errorf("Reading teamids.txt: %v", err)
	}
	
	teamids := bytes.Split(teamidsBuf, []byte("\n"))
	if (len(teamids) != 101) || (len(teamids[100]) > 0) {
  	t.Errorf("There weren't 100 teamids, there were %d", len(teamids))
	}
	myTeam := string(teamids[0])
	
	if err := s.SetTeamName("bad team ID", "bad team name"); err == nil {
	  t.Errorf("Setting bad team ID didn't raise an error")
	}
	
	if err := s.SetTeamName(myTeam, "My Team"); err != nil {
	  t.Errorf("Setting team name: %v", err)
	}
}
