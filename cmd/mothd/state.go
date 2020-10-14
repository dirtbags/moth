package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dirtbags/moth/pkg/award"
	"github.com/spf13/afero"
)

// DistinguishableChars are visually unambiguous glyphs.
// People with mediocre handwriting could write these down unambiguously,
// and they can be entered without holding down shift.
const DistinguishableChars = "34678abcdefhikmnpqrtwxy="

// RFC3339Space is a time layout which replaces 'T' with a space.
// This is also a valid RFC3339 format.
const RFC3339Space = "2006-01-02 15:04:05Z07:00"

// ErrAlreadyRegistered means a team cannot be registered because it was registered previously.
var ErrAlreadyRegistered = errors.New("Team ID has already been registered")

// State defines the current state of a MOTH instance.
// We use the filesystem for synchronization between threads.
// The only thing State methods need to know is the path to the state directory.
type State struct {
	afero.Fs

	// Enabled tracks whether the current State system is processing updates
	Enabled bool

	refreshNow  chan bool
	eventStream chan string
	eventWriter afero.File
}

// NewState returns a new State struct backed by the given Fs
func NewState(fs afero.Fs) *State {
	s := &State{
		Fs:          fs,
		Enabled:     true,
		refreshNow:  make(chan bool, 5),
		eventStream: make(chan string, 80),
	}
	if err := s.reopenEventLog(); err != nil {
		log.Fatal(err)
	}
	return s
}

// updateEnabled checks a few things to see if this state directory is "enabled".
func (s *State) updateEnabled() {
	nextEnabled := true
	why := "`state/enabled` present, `state/hours.txt` missing"

	if untilFile, err := s.Open("hours.txt"); err == nil {
		defer untilFile.Close()
		why = "`state/hours.txt` present"

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
				log.Println("Misformatted line in hours.txt file")
			}
			line = strings.TrimSpace(line)
			until, err := time.Parse(time.RFC3339, line)
			if err != nil {
				until, err = time.Parse(RFC3339Space, line)
			}
			if err != nil {
				log.Println("Suspended: Unparseable until date:", line)
				continue
			}
			if until.Before(time.Now()) {
				nextEnabled = thisEnabled
			}
		}
	}

	if _, err := s.Stat("enabled"); os.IsNotExist(err) {
		dirs, _ := afero.ReadDir(s, ".")
		for _, dir := range dirs {
			log.Println(dir.Name())
		}

		log.Print(s, err)
		nextEnabled = false
		why = "`state/enabled` missing"
	}

	if nextEnabled != s.Enabled {
		s.Enabled = nextEnabled
		log.Printf("Setting enabled=%v: %s", s.Enabled, why)
	}
}

// TeamName returns team name given a team ID.
func (s *State) TeamName(teamID string) (string, error) {
	teamFs := afero.NewBasePathFs(s.Fs, "teams")
	teamNameBytes, err := afero.ReadFile(teamFs, teamID)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("Unregistered team ID: %s", teamID)
	} else if err != nil {
		return "", fmt.Errorf("Unregistered team ID: %s (%s)", teamID, err)
	}

	teamName := strings.TrimSpace(string(teamNameBytes))
	return teamName, nil
}

// SetTeamName writes out team name.
// This can only be done once per team.
func (s *State) SetTeamName(teamID, teamName string) error {
	idsFile, err := s.Open("teamids.txt")
	if err != nil {
		return fmt.Errorf("Team IDs file does not exist")
	}
	defer idsFile.Close()
	found := false
	scanner := bufio.NewScanner(idsFile)
	for scanner.Scan() {
		if scanner.Text() == teamID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("Team ID not found in list of valid Team IDs")
	}

	teamFilename := filepath.Join("teams", teamID)
	teamFile, err := s.Fs.OpenFile(teamFilename, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
	if os.IsExist(err) {
		return ErrAlreadyRegistered
	} else if err != nil {
		return err
	}
	defer teamFile.Close()
	log.Println("Setting team name to:", teamName, teamFilename, teamFile)
	fmt.Fprintln(teamFile, teamName)
	teamFile.Close()
	return nil
}

// PointsLog retrieves the current points log.
func (s *State) PointsLog() award.List {
	f, err := s.Open("points.log")
	if err != nil {
		log.Println(err)
		return nil
	}
	defer f.Close()

	pointsLog := make(award.List, 0, 200)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		log.Println(line)
		cur, err := award.Parse(line)
		if err != nil {
			log.Printf("Skipping malformed award line %s: %s", line, err)
			continue
		}
		pointsLog = append(pointsLog, cur)
	}
	return pointsLog
}

// Messages retrieves the current messages.
func (s *State) Messages() string {
	bMessages, _ := afero.ReadFile(s, "messages.html")
	return string(bMessages)
}

// AwardPoints gives points to teamID in category.
// It first checks to make sure these are not duplicate points.
// This is not a perfect check, you can trigger a race condition here.
// It's just a courtesy to the user.
// The update task makes sure we never have duplicate points in the log.
func (s *State) AwardPoints(teamID, category string, points int) error {
	a := award.T{
		When:     time.Now().Unix(),
		TeamID:   teamID,
		Category: category,
		Points:   points,
	}

	_, err := s.TeamName(teamID)
	if err != nil {
		return err
	}

	for _, e := range s.PointsLog() {
		if a.Equal(e) {
			return fmt.Errorf("Points already awarded to this team in this category")
		}
	}

	fn := fmt.Sprintf("%s-%s-%d", teamID, category, points)
	tmpfn := filepath.Join("points.tmp", fn)
	newfn := filepath.Join("points.new", fn)

	if err := afero.WriteFile(s, tmpfn, []byte(a.String()), 0644); err != nil {
		return err
	}

	if err := s.Rename(tmpfn, newfn); err != nil {
		return err
	}

	//  State should be updated immediately
	s.refreshNow <- true

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
		awd, err := award.Parse(string(awardstr))
		if err != nil {
			log.Print("Can't parse award file ", filename, ": ", err)
			continue
		}

		duplicate := false
		for _, e := range s.PointsLog() {
			if awd.Equal(e) {
				duplicate = true
				break
			}
		}

		if duplicate {
			log.Print("Skipping duplicate points: ", awd.String())
		} else {
			log.Print("Award: ", awd.String())

			logf, err := s.OpenFile("points.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Print("Can't append to points log: ", err)
				return
			}
			fmt.Fprintln(logf, awd.String())
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
	s.Remove("hours.txt")
	s.Remove("points.log")
	s.Remove("messages.html")
	s.Remove("mothd.log")
	s.RemoveAll("points.tmp")
	s.RemoveAll("points.new")
	s.RemoveAll("teams")

	// Open log file
	if err := s.reopenEventLog(); err != nil {
		log.Fatal(err)
	}

	// Make sure various subdirectories exist
	s.Mkdir("points.tmp", 0755)
	s.Mkdir("points.new", 0755)
	s.Mkdir("teams", 0755)

	// Preseed available team ids if file doesn't exist
	if f, err := s.OpenFile("teamids.txt", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644); err == nil {
		id := make([]byte, 8)
		for i := 0; i < 100; i++ {
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
		fmt.Fprintln(f, "This instance was initialized at", now)
		f.Close()
	}

	if f, err := s.Create("enabled"); err == nil {
		fmt.Fprintln(f, "enabled: remove or rename to suspend the contest.")
		f.Close()
	}

	if f, err := s.Create("hours.txt"); err == nil {
		fmt.Fprintln(f, "# hours.txt: when the contest is enabled")
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

// LogEvent writes msg to the event log
func (s *State) LogEvent(msg string) {
	s.eventStream <- msg
}

func (s *State) reopenEventLog() error {
	if s.eventWriter != nil {
		if err := s.eventWriter.Close(); err != nil {
			// We're going to soldier on if Close returns error
			log.Print(err)
		}
	}
	eventWriter, err := s.OpenFile("event.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	s.eventWriter = eventWriter
	return nil
}

func (s *State) refresh() {
	s.maybeInitialize()
	s.updateEnabled()
	if s.Enabled {
		s.collectPoints()
	}
}

// Maintain performs housekeeping on a State struct.
func (s *State) Maintain(updateInterval time.Duration) {
	ticker := time.NewTicker(updateInterval)
	s.refresh()
	for {
		select {
		case msg := <-s.eventStream:
			fmt.Fprintln(s.eventWriter, time.Now().Unix(), msg)
			s.eventWriter.Sync()
		case <-ticker.C:
			s.refresh()
		case <-s.refreshNow:
			s.refresh()
		}
	}
}
