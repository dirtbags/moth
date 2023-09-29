package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dirtbags/moth/v4/pkg/award"
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
var ErrAlreadyRegistered = errors.New("team ID has already been registered")

// State defines the current state of a MOTH instance.
// We use the filesystem for synchronization between threads.
// The only thing State methods need to know is the path to the state directory.
type State struct {
	afero.Fs

	// Enabled tracks whether the current State system is processing updates
	enabled bool

	enabledWhy      string
	refreshNow      chan bool
	eventStream     chan []string
	eventWriter     *csv.Writer
	eventWriterFile afero.File

	// Caches, so we're not hammering NFS with metadata operations
	teamNamesLastChange time.Time
	teamNames           map[string]string
	pointsLog           award.List
	lock                sync.RWMutex
}

// NewState returns a new State struct backed by the given Fs
func NewState(fs afero.Fs) *State {
	s := &State{
		Fs:          fs,
		enabled:     true,
		refreshNow:  make(chan bool, 5),
		eventStream: make(chan []string, 80),

		teamNames: make(map[string]string),
	}
	if err := s.reopenEventLog(); err != nil {
		log.Fatal(err)
	}
	return s
}

// updateEnabled checks a few things to see if this state directory is "enabled".
func (s *State) updateEnabled() {
	nextEnabled := true
	why := "state/hours.txt has no timestamps before now"

	if untilFile, err := s.Open("hours.txt"); err == nil {
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
				log.Println("state/hours.txt has bad line:", line)
			}
			line, _, _ = strings.Cut(line, "#") // Remove inline comments
			line = strings.TrimSpace(line)
			until := time.Time{}
			if len(line) == 0 {
				// Let it stay as zero time, so it's always before now
			} else if until, err = time.Parse(time.RFC3339, line); err == nil {
				// Great, it was RFC 3339
			} else if until, err = time.Parse(RFC3339Space, line); err == nil {
				// Great, it was RFC 3339 with a space instead of a 'T'
			} else {
				log.Println("state/hours.txt has bad timestamp:", line)
				continue
			}
			if until.Before(time.Now()) {
				nextEnabled = thisEnabled
				why = fmt.Sprint("state/hours.txt most recent timestamp:", line)
			}
		}
	}

	if (nextEnabled != s.enabled) || (why != s.enabledWhy) {
		s.enabled = nextEnabled
		s.enabledWhy = why
		log.Printf("Setting enabled=%v: %s", s.enabled, s.enabledWhy)
		if s.enabled {
			s.LogEvent("enabled", "", "", 0, s.enabledWhy)
		} else {
			s.LogEvent("disabled", "", "", 0, s.enabledWhy)
		}
	}
}

// TeamName returns team name given a team ID.
func (s *State) TeamName(teamID string) (string, error) {
	s.lock.RLock()
	name, ok := s.teamNames[teamID]
	s.lock.RUnlock()
	if !ok {
		return "", fmt.Errorf("unregistered team ID: %s", teamID)
	}
	return name, nil
}

// SetTeamName writes out team name.
// This can only be done once per team.
func (s *State) SetTeamName(teamID, teamName string) error {
	s.lock.RLock()
	_, ok := s.teamNames[teamID]
	s.lock.RUnlock()
	if ok {
		return ErrAlreadyRegistered
	}

	idsFile, err := s.Open("teamids.txt")
	if err != nil {
		return fmt.Errorf("team IDs file does not exist")
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
		return fmt.Errorf("team ID not found in list of valid team IDs")
	}

	teamFilename := filepath.Join("teams", teamID)
	teamFile, err := s.Fs.OpenFile(teamFilename, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
	if os.IsExist(err) {
		return ErrAlreadyRegistered
	} else if err != nil {
		return err
	}
	defer teamFile.Close()
	log.Printf("Setting team name [%s] in file %s", teamName, teamFilename)
	fmt.Fprintln(teamFile, teamName)
	teamFile.Close()

	s.refreshNow <- true

	return nil
}

// PointsLog retrieves the current points log.
func (s *State) PointsLog() award.List {
	s.lock.RLock()
	ret := make(award.List, len(s.pointsLog))
	copy(ret, s.pointsLog)
	s.lock.RUnlock()
	return ret
}

// Enabled returns true if the server is in "enabled" state
func (s *State) Enabled() bool {
	return s.enabled
}

// AwardPoints gives points to teamID in category.
// This doesn't attempt to ensure the teamID has been registered.
// It first checks to make sure these are not duplicate points.
// This is not a perfect check, you can trigger a race condition here.
// It's just a courtesy to the user.
// The update task makes sure we never have duplicate points in the log.
func (s *State) AwardPoints(teamID, category string, points int) error {
	return s.awardPointsAtTime(time.Now().Unix(), teamID, category, points)
}

func (s *State) awardPointsAtTime(when int64, teamID string, category string, points int) error {
	a := award.T{
		When:     when,
		TeamID:   teamID,
		Category: category,
		Points:   points,
	}

	for _, e := range s.PointsLog() {
		if a.Equal(e) {
			return fmt.Errorf("points already awarded to this team in this category")
		}
	}

	//fn := fmt.Sprintf("%s-%s-%d", a.TeamID, a.Category, a.Points)
	fn := a.Filename()
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
		s.lock.RLock()
		for _, e := range s.pointsLog {
			if awd.Equal(e) {
				duplicate = true
				break
			}
		}
		s.lock.RUnlock()

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

			// Stick this on the cache too
			s.lock.Lock()
			s.pointsLog = append(s.pointsLog, awd)
			s.lock.Unlock()
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
	s.Remove("events.csv")
	s.Remove("mothd.log")
	s.RemoveAll("points.tmp")
	s.RemoveAll("points.new")
	s.RemoveAll("teams")

	// Open log file
	if err := s.reopenEventLog(); err != nil {
		log.Fatal(err)
	}
	s.LogEvent("init", "", "", 0)

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

	if f, err := s.Create("hours.txt"); err == nil {
		fmt.Fprintln(f, "# hours.txt: when the contest is enabled")
		fmt.Fprintln(f, "#")
		fmt.Fprintln(f, "# Enable:  + [timestamp]")
		fmt.Fprintln(f, "# Disable: - [timestamp]")
		fmt.Fprintln(f, "#")
		fmt.Fprintln(f, "# This file, and all files in this directory, are re-read periodically.")
		fmt.Fprintln(f, "# Default is enabled.")
		fmt.Fprintln(f, "# Rules with only '-' or '+' are also allowed.")
		fmt.Fprintln(f, "# Rules apply from the top down.")
		fmt.Fprintln(f, "# If you put something in out of order, it's going to be bonkers.")
		fmt.Fprintln(f)
		fmt.Fprintln(f, "- 1970-01-01T00:00:00Z")
		fmt.Fprintln(f, "+", now)
		fmt.Fprintln(f, "- 2519-10-31T00:00:00Z")
		f.Close()
	}

	if f, err := s.Create("points.log"); err == nil {
		f.Close()
	}
}

// LogEvent writes to the event log
func (s *State) LogEvent(event, teamID, cat string, points int, extra ...string) {
	s.eventStream <- append(
		[]string{
			strconv.FormatInt(time.Now().Unix(), 10),
			event,
			teamID,
			cat,
			strconv.Itoa(points),
		},
		extra...,
	)
}

func (s *State) reopenEventLog() error {
	if s.eventWriter != nil {
		s.eventWriter.Flush()
	}
	if s.eventWriterFile != nil {
		if err := s.eventWriterFile.Close(); err != nil {
			// We're going to soldier on if Close returns error
			log.Print(err)
		}
	}
	eventWriterFile, err := s.OpenFile("events.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	s.eventWriterFile = eventWriterFile
	s.eventWriter = csv.NewWriter(s.eventWriterFile)
	return nil
}

func (s *State) updateCaches() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if f, err := s.Open("points.log"); err != nil {
		log.Println(err)
	} else {
		defer f.Close()

		pointsLog := make(award.List, 0, 200)
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			cur, err := award.Parse(line)
			if err != nil {
				log.Printf("Skipping malformed award line %s: %s", line, err)
				continue
			}
			pointsLog = append(pointsLog, cur)
		}
		s.pointsLog = pointsLog
	}

	// Only do this if the teams directory has a newer mtime; directories with
	// hundreds of team names can cause NFS I/O storms
	{
		_, ismmfs := s.Fs.(*afero.MemMapFs) // Tests run so quickly that the time check isn't precise enough
		if fi, err := s.Fs.Stat("teams"); err != nil {
			log.Printf("Getting modification time of teams directory: %v", err)
		} else if ismmfs || s.teamNamesLastChange.Before(fi.ModTime()) {
			s.teamNamesLastChange = fi.ModTime()

			// The compiler recognizes this as an optimization case
			for k := range s.teamNames {
				delete(s.teamNames, k)
			}

			teamsFs := afero.NewBasePathFs(s.Fs, "teams")
			if dirents, err := afero.ReadDir(teamsFs, "."); err != nil {
				log.Printf("Reading team ids: %v", err)
			} else {
				for _, dirent := range dirents {
					teamID := dirent.Name()
					if teamNameBytes, err := afero.ReadFile(teamsFs, teamID); err != nil {
						log.Printf("Reading team %s: %v", teamID, err)
					} else {
						teamName := strings.TrimSpace(string(teamNameBytes))
						s.teamNames[teamID] = teamName
					}
				}
			}
		}
	}
}

func (s *State) refresh() {
	s.maybeInitialize()
	s.updateEnabled()
	if s.enabled {
		s.collectPoints()
	}
	s.updateCaches()
}

// Maintain performs housekeeping on a State struct.
func (s *State) Maintain(updateInterval time.Duration) {
	ticker := time.NewTicker(updateInterval)
	s.refresh()
	for {
		select {
		case msg := <-s.eventStream:
			s.eventWriter.Write(msg)
			s.eventWriter.Flush()
			s.eventWriterFile.Sync()
		case <-ticker.C:
			s.refresh()
		case <-s.refreshNow:
			s.refresh()
		}
	}
}

// DevelState is a StateProvider for use by development servers
type DevelState struct {
	StateProvider
}

// NewDevelState returns a new state object that can be used by the development server.
//
// The main thing this provides is the ability to register a team with any team ID.
// If a team ID is provided that wasn't recognized by the underlying StateProvider,
// it is associated with a team named "<devel:$ID>".
//
// This makes it possible to use the server without having to register a team.
func NewDevelState(sp StateProvider) *DevelState {
	return &DevelState{sp}
}

// TeamName returns a valid team name for any teamID
//
// If one's registered, it will use it.
// Otherwise, it returns "<devel:$ID>"
func (ds *DevelState) TeamName(teamID string) (string, error) {
	if name, err := ds.StateProvider.TeamName(teamID); err == nil {
		return name, nil
	}
	if teamID == "" {
		return "", fmt.Errorf("empty team ID")
	}
	return fmt.Sprintf("«devel:%s»", teamID), nil
}

// SetTeamName associates a team name with any teamID
//
// If the underlying StateProvider returns any sort of error,
// this returns ErrAlreadyRegistered,
// so the user can join a pre-existing team for whatever ID the provide.
func (ds *DevelState) SetTeamName(teamID, teamName string) error {
	if err := ds.StateProvider.SetTeamName(teamID, teamName); err != nil {
		return ErrAlreadyRegistered
	}
	return nil
}
