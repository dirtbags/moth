package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

type LegacyMOTHState struct {
	StateDir	string
	update	chan bool
	maintenanceInterval time.Duration
}

func (state *LegacyMOTHState) Initialize() (bool, error) {

	if _, err := os.Stat(state.StateDir); err != nil {
		return false, err
	}

	state.MaybeInitialize()

	if state.update == nil {
		state.update = make(chan bool, 10)
		go state.Maintenance(state.maintenanceInterval)
	}
	return true, nil
}

func (state *LegacyMOTHState) login(teamName string, token string) (bool, error) {
	for a, _ := range state.getTeams() {
		if a == token {
			f, err := os.OpenFile(state.StatePath("teams", token), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		        if err != nil {
		                if os.IsExist(err) {
					return true, ErrAlreadyRegistered
		                } else {
					return false, ErrRegistrationError
				}
		        }
		        defer f.Close()

			fmt.Fprintln(f, teamName)
			return true, nil
		}
	}

	return false, ErrInvalidTeamID
}

func (state *LegacyMOTHState) StatePath(parts ...string) string {
	tail := pathCleanse(parts)
	return path.Join(state.StateDir, tail)
}

func (state *LegacyMOTHState) TeamName(teamId string) (string, error) {
        teamNameBytes, err := ioutil.ReadFile(state.StatePath("teams", teamId))
        teamName := strings.TrimSpace(string(teamNameBytes))
        return teamName, err
}

func (state *LegacyMOTHState) isEnabled() bool {
	if _, err := os.Stat(state.StatePath("disabled")); err == nil {
		return false
	}

	untilspec, err := ioutil.ReadFile(state.StatePath("until"))
	if err == nil {
		untilspecs := strings.TrimSpace(string(untilspec))
		until, err := time.Parse(time.RFC3339, untilspecs)
		if err != nil {
			log.Printf("Suspended: Unparseable until date: %s", untilspec)
			return false
		}
		if until.Before(time.Now()) {
			log.Print("Suspended: until time reached, suspending maintenance")
			return false
		}
	}

	return true
}

// AwardPoints gives points to teamId in category.
// It first checks to make sure these are not duplicate points.
// This is not a perfect check, you can trigger a race condition here.
// It's just a courtesy to the user.
// The maintenance task makes sure we never have duplicate points in the log.
func (state *LegacyMOTHState) AwardPoints(teamId, category string, points int) error {
        a := Award{
                When:     time.Now(),
                TeamId:   teamId,
                Category: category,
                Points:   points,
        }

        _, err := state.TeamName(teamId)
        if err != nil {
                return fmt.Errorf("No registered team with this hash")
        }

        for _, e := range state.PointsLog("") {
                if a.Same(e) {
                        return fmt.Errorf("Points already awarded to this team in this category")
                }
        }

        fn := fmt.Sprintf("%s-%s-%d", teamId, category, points)
        tmpfn := state.StatePath("points.tmp", fn)
        newfn := state.StatePath("points.new", fn)

        if err := ioutil.WriteFile(tmpfn, []byte(a.String()), 0644); err != nil {
                return err
        }

        if err := os.Rename(tmpfn, newfn); err != nil {
                return err
        }

        state.update <- true
        log.Printf("Award %s %s %d", teamId, category, points)
        return nil
}

func (state *LegacyMOTHState) PointsLog(teamId string) []*Award {
        var ret []*Award

        fn := state.StatePath("points.log")
        f, err := os.Open(fn)
        if err != nil {
                log.Printf("Unable to open %s: %s", fn, err)
                return ret
        }
        defer f.Close()

        scanner := bufio.NewScanner(f)
        for scanner.Scan() {
                line := scanner.Text()
                cur, err := ParseAward(line)
                if err != nil {
                        log.Printf("Skipping malformed award line %s: %s", line, err)
                        continue
                }
                if len(teamId) > 0 && cur.TeamId != teamId {
                        continue
                }
                ret = append(ret, cur)
        }

        return ret
}

func (state *LegacyMOTHState) getConfig(configName string) (string, error) {
	fn := state.StatePath(configName)
	data, err := ioutil.ReadFile(fn)

	if err != nil {
		log.Printf("Unable to open %s: %s", fn, err)
		return "", err
	}

	return string(data), nil
}

func (state *LegacyMOTHState) Maintenance(maintenanceInterval time.Duration) {
	for {
		if state.isEnabled() {
			state.collectPoints()
		}
		select {
		case <-state.update:
			// log.Print("Forced update")
		case <-time.After(maintenanceInterval):
			// log.Print("Housekeeping")
		}
	}
}

func (state *LegacyMOTHState) getTeams() map[string]struct{} {
        filepath := state.StatePath("teamids.txt")
        teamids, err := os.Open(filepath)
	teams := make(map[string]struct{})
        if err != nil {
                log.Printf("Error openining %s: %s", filepath, err)
                return teams
        }
        defer teamids.Close()

        // List out team IDs
        scanner := bufio.NewScanner(teamids)
        for scanner.Scan() {
                teamId := scanner.Text()
                if (teamId == "..") || strings.ContainsAny(teamId, "/") {
                        log.Printf("Dangerous team ID dropped: %s", teamId)
                        continue
                }
		teams[scanner.Text()] = struct{}{}
		//newList = append(newList, scanner.Text())
        }
	return teams
	//return newList
}

// collectPoints gathers up files in points.new/ and appends their contents to points.log,
// removing each points.new/ file as it goes.
func (state *LegacyMOTHState) collectPoints() {
        logf, err := os.OpenFile(state.StatePath("points.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
                log.Printf("Can't append to points log: %s", err)
                return
        }
        defer logf.Close()

        files, err := ioutil.ReadDir(state.StatePath("points.new"))
        if err != nil {
                log.Printf("Error reading packages: %s", err)
        }
        for _, f := range files {
                filename := state.StatePath("points.new", f.Name())
                s, err := ioutil.ReadFile(filename)
                if err != nil {
                        log.Printf("Can't read points file %s: %s", filename, err)
                        continue
                }
                award, err := ParseAward(string(s))
                if err != nil {
                        log.Printf("Can't parse award file %s: %s", filename, err)
                        continue
                }

                duplicate := false
                for _, e := range state.PointsLog("") {
                        if award.Same(e) {
                                duplicate = true
                                break
                        }
                }

                if duplicate {
                        log.Printf("Skipping duplicate points: %s", award.String())
                } else {
                        fmt.Fprintf(logf, "%s\n", award.String())
                }

                logf.Sync()
                if err := os.Remove(filename); err != nil {
                        log.Printf("Unable to remove %s: %s", filename, err)
                }
        }
}

func (state *LegacyMOTHState) MaybeInitialize() {
	// Only do this if it hasn't already been done
	if _, err := os.Stat(state.StatePath("initialized")); err == nil {
		return
	}
	log.Printf("initialized file missing, re-initializing")

	// Remove any extant control and state files
        os.Remove(state.StatePath("until"))
        os.Remove(state.StatePath("disabled"))
        os.Remove(state.StatePath("points.log"))
        os.RemoveAll(state.StatePath("points.tmp"))
        os.RemoveAll(state.StatePath("points.new"))
        os.RemoveAll(state.StatePath("teams"))

        // Make sure various subdirectories exist
        os.Mkdir(state.StatePath("points.tmp"), 0755)
        os.Mkdir(state.StatePath("points.new"), 0755)
        os.Mkdir(state.StatePath("teams"), 0755)

        // Preseed available team ids if file doesn't exist
        if f, err := os.OpenFile(state.StatePath("teamids.txt"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644); err == nil {
                defer f.Close()
                for i := 0; i <= 100; i += 1 {
                        fmt.Fprintln(f, mktoken())
                }
        }

        // Create initialized file that signals whether we're set up
        f, err := os.OpenFile(state.StatePath("initialized"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
        if err != nil {
                log.Print(err)
        }
        defer f.Close()
        fmt.Fprintln(f, "Remove this file to reinitialize the contest")
}
