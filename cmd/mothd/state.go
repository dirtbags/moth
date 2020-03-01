package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/afero"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Stuff people with mediocre handwriting could write down unambiguously, and can be entered without holding down shift
const DistinguishableChars = "234678abcdefhikmnpqrtwxyz="

// We use the filesystem for synchronization between threads.
// The only thing State methods need to know is the path to the state directory.
type State struct {
	afero.Fs
	Enabled bool
}

func NewState(fs afero.Fs) *State {
	return &State{
		Fs:      fs,
		Enabled: true,
	}
}

// Check a few things to see if this state directory is "enabled".
func (s *State) UpdateEnabled() {
	if _, err := s.Stat("enabled"); os.IsNotExist(err) {
		s.Enabled = false
		log.Println("Suspended: enabled file missing")
		return
	}

	nextEnabled := true
	untilFile, err := s.Open("hours")
	if err != nil {
		return
	}
	defer untilFile.Close()

	scanner := bufio.NewScanner(untilFile)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 1 {
			continue
		}

		thisEnabled := true
		switch line[0] {
		case '+':
			thisEnabled = true
			line = line[1:]
		case '-':
			thisEnabled = false
			line = line[1:]
		case '#':
			continue
		default:
			log.Println("Misformatted line in hours file")
		}
		line = strings.TrimSpace(line)
		until, err := time.Parse(time.RFC3339, line)
		if err != nil {
			log.Println("Suspended: Unparseable until date:", line)
			continue
		}
		if until.Before(time.Now()) {
			nextEnabled = thisEnabled
		}
	}
	if nextEnabled != s.Enabled {
		s.Enabled = nextEnabled
		log.Println("Setting enabled to", s.Enabled, "based on hours file")
	}
}

// Returns team name given a team ID.
func (s *State) TeamName(teamId string) (string, error) {
	teamFile := filepath.Join("teams", teamId)
	teamNameBytes, err := afero.ReadFile(s, teamFile)
	teamName := strings.TrimSpace(string(teamNameBytes))

	if os.IsNotExist(err) {
		return "", fmt.Errorf("Unregistered team ID: %s", teamId)
	} else if err != nil {
		return "", fmt.Errorf("Unregistered team ID: %s (%s)", teamId, err)
	}

	return teamName, nil
}

// Write out team name. This can only be done once.
func (s *State) SetTeamName(teamId, teamName string) error {
	if f, err := s.Open("teamids.txt"); err != nil {
		return fmt.Errorf("Team IDs file does not exist")
	} else {
		found := false
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if scanner.Text() == teamId {
				found = true
				break
			}
		}
		f.Close()
		if !found {
			return fmt.Errorf("Team ID not found in list of valid Team IDs")
		}
	}

	teamFile := filepath.Join("teams", teamId)
	err := afero.WriteFile(s, teamFile, []byte(teamName), os.ModeExclusive|0644)
	if os.IsExist(err) {
		return fmt.Errorf("Team ID is already registered")
	}
	return err
}

// Retrieve the current points log
func (s *State) PointsLog() []*Award {
	f, err := s.Open("points.log")
	if err != nil {
		log.Println(err)
		return nil
	}
	defer f.Close()

	pointsLog := make([]*Award, 0, 200)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		cur, err := ParseAward(line)
		if err != nil {
			log.Printf("Skipping malformed award line %s: %s", line, err)
			continue
		}
		pointsLog = append(pointsLog, cur)
	}
	return pointsLog
}

// Retrieve current messages
func (s *State) Messages() string {
	bMessages, _ := afero.ReadFile(s, "messages.html")
	return string(bMessages)
}

// AwardPoints gives points to teamId in category.
// It first checks to make sure these are not duplicate points.
// This is not a perfect check, you can trigger a race condition here.
// It's just a courtesy to the user.
// The update task makes sure we never have duplicate points in the log.
func (s *State) AwardPoints(teamId, category string, points int) error {
	a := Award{
		When:     time.Now().Unix(),
		TeamId:   teamId,
		Category: category,
		Points:   points,
	}

	_, err := s.TeamName(teamId)
	if err != nil {
		return err
	}

	for _, e := range s.PointsLog() {
		if a.Same(e) {
			return fmt.Errorf("Points already awarded to this team in this category")
		}
	}

	fn := fmt.Sprintf("%s-%s-%d", teamId, category, points)
	tmpfn := filepath.Join("points.tmp", fn)
	newfn := filepath.Join("points.new", fn)

	if err := afero.WriteFile(s, tmpfn, []byte(a.String()), 0644); err != nil {
		return err
	}

	if err := s.Rename(tmpfn, newfn); err != nil {
		return err
	}

	// XXX: update everything
	return nil
}

// collectPoints gathers up files in points.new/ and appends their contents to points.log,
// removing each points.new/ file as it goes.
func (s *State) collectPoints() {
	files, err := afero.ReadDir(s, "points.new")
	if err != nil {
		log.Print(err)
		return
	}
	for _, f := range files {
		filename := filepath.Join("points.new", f.Name())
		awardstr, err := afero.ReadFile(s, filename)
		if err != nil {
			log.Print("Opening new points: ", err)
			continue
		}
		award, err := ParseAward(string(awardstr))
		if err != nil {
			log.Print("Can't parse award file ", filename, ": ", err)
			continue
		}

		duplicate := false
		for _, e := range s.PointsLog() {
			if award.Same(e) {
				duplicate = true
				break
			}
		}

		if duplicate {
			log.Print("Skipping duplicate points: ", award.String())
		} else {
			log.Print("Award: ", award.String())

			logf, err := s.OpenFile("points.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Print("Can't append to points log: ", err)
				return
			}
			fmt.Fprintln(logf, award.String())
			logf.Close()
		}

		if err := s.Remove(filename); err != nil {
			log.Print("Unable to remove new points file: ", err)
		}
	}
}

func (s *State) maybeInitialize() {
	// Are we supposed to re-initialize?
	if _, err := s.Stat("initialized"); !os.IsNotExist(err) {
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	log.Print("initialized file missing, re-initializing")

	// Remove any extant control and state files
	s.Remove("enabled")
	s.Remove("hours")
	s.Remove("points.log")
	s.Remove("messages.html")
	s.RemoveAll("points.tmp")
	s.RemoveAll("points.new")
	s.RemoveAll("teams")

	// Make sure various subdirectories exist
	s.Mkdir("points.tmp", 0755)
	s.Mkdir("points.new", 0755)
	s.Mkdir("teams", 0755)

	// Preseed available team ids if file doesn't exist
	if f, err := s.OpenFile("teamids.txt", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644); err == nil {
		id := make([]byte, 8)
		for i := 0; i < 100; i += 1 {
			for i := range id {
				char := rand.Intn(len(DistinguishableChars))
				id[i] = DistinguishableChars[char]
			}
			fmt.Fprintln(f, string(id))
		}
		f.Close()
	}

	// Create some files
	if f, err := s.Create("initialized"); err == nil {
		fmt.Fprintln(f, "initialized: remove to re-initialize the contest.")
		fmt.Fprintln(f)
		fmt.Fprintln(f, "This instance was initaliazed at", now)
		f.Close()
	}

	if f, err := s.Create("enabled"); err == nil {
		fmt.Fprintln(f, "enabled: remove or rename to suspend the contest.")
		f.Close()
	}

	if f, err := s.Create("hours"); err == nil {
		fmt.Fprintln(f, "# hours: when the contest is enabled")
		fmt.Fprintln(f, "#")
		fmt.Fprintln(f, "# Enable:  + timestamp")
		fmt.Fprintln(f, "# Disable: - timestamp")
		fmt.Fprintln(f, "#")
		fmt.Fprintln(f, "# You can have multiple start/stop times.")
		fmt.Fprintln(f, "# Whatever time is the most recent, wins.")
		fmt.Fprintln(f, "# Times in the future are ignored.")
		fmt.Fprintln(f)
		fmt.Fprintln(f, "+", now)
		fmt.Fprintln(f, "- 3019-10-31T00:00:00Z")
		f.Close()
	}

	if f, err := s.Create("messages.html"); err == nil {
		fmt.Fprintln(f, "<!-- messages.html: put client broadcast messages here. -->")
		f.Close()
	}

	if f, err := s.Create("points.log"); err == nil {
		f.Close()
	}

}

func (s *State) Update() {
	s.maybeInitialize()
	s.UpdateEnabled()
	if s.Enabled {
		s.collectPoints()
	}
}
