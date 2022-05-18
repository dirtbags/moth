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
var ErrAlreadyRegistered = errors.New("team ID has already been registered")
var NoMatchingPointEntry = errors.New("Unable to find matching point entry")

// State defines the current state of a MOTH instance.
// We use the filesystem for synchronization between threads.
// The only thing State methods need to know is the path to the state directory.
type State struct {
	afero.Fs

	// Enabled tracks whether the current State system is processing updates
	Enabled bool

	refreshNow      chan bool
	eventStream     chan []string
	eventWriter     *csv.Writer
	eventWriterFile afero.File

	// Caches, so we're not hammering NFS with metadata operations
	teamNames map[string]string
	pointsLog award.List
	messages  string
	teamIDLock		sync.RWMutex
	teamIDFileLock	sync.RWMutex
	teamNameLock	sync.RWMutex
	teamNameFileLock	sync.RWMutex
	pointsLock	sync.RWMutex
	pointsLogFileLock	sync.RWMutex  // Sometimes, we need to fiddle with the file, while leaving the internal state alone
	messageFileLock	sync.RWMutex
	teamNamesLastChange time.Time
}

// NewState returns a new State struct backed by the given Fs
func NewState(fs afero.Fs) *State {
	s := &State{
		Fs:          fs,
		Enabled:     true,
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
		nextEnabled = false
		why = "`state/enabled` missing"
	}

	if nextEnabled != s.Enabled {
		s.Enabled = nextEnabled
		log.Printf("Setting enabled=%v: %s", s.Enabled, why)
		if s.Enabled {
			s.LogEvent("enabled", "", "", "", 0, why)
		} else {
			s.LogEvent("disabled", "", "", "", 0, why)
		}
	}
}

/* ****************** Team ID functions ****************** */

func (s *State) TeamIDs() ([]string, error) {
	var teamIDs []string

	s.teamIDFileLock.RLock()
	defer s.teamIDFileLock.RUnlock()

	idsFile, err := s.Open("teamids.txt")
	if err != nil {
		return teamIDs, fmt.Errorf("team IDs file does not exist")
	}
	defer idsFile.Close()

	scanner := bufio.NewScanner(idsFile)
	for scanner.Scan() {
		teamIDs = append(teamIDs, scanner.Text())
	}

	return teamIDs, nil
}

func (s *State) writeTeamIDs(teamIDs []string) error {
	s.teamIDFileLock.Lock()
	defer s.teamIDFileLock.Unlock()

	if f, err := s.OpenFile("teamids.txt", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644); err == nil {
		defer f.Close()

		for _, teamID := range teamIDs {
			fmt.Fprintln(f, string(teamID))
		}
	} else {
		return err
	}

	return nil
}

func (s *State) SetTeamIDs(teamIDs []string) error {
	s.teamIDLock.Lock()
	defer s.teamIDLock.Unlock()

	return s.writeTeamIDs(teamIDs)
}

func (s *State) AddTeamID(newTeamID string) error {
	s.teamIDLock.Lock()
	defer s.teamIDLock.Unlock()

	teamIDs, err := s.TeamIDs()

	if err != nil {
		return err
	}

	for _, teamID := range teamIDs {
		if newTeamID == teamID {
			return fmt.Errorf("Team ID already exists")
		}
	}

	teamIDs = append(teamIDs, newTeamID)

	return s.writeTeamIDs(teamIDs)
}

func (s *State) RemoveTeamID(removeTeamID string) error {
	s.teamIDLock.Lock()
	defer s.teamIDLock.Unlock()

	teamIDs, err := s.TeamIDs()

	if err != nil {
		return err
	}

	for _, teamID := range teamIDs {
		if removeTeamID != teamID {
			teamIDs = append(teamIDs, teamID)
		}
	}

	return s.writeTeamIDs(teamIDs)
}

func (s *State) TeamIDExists(teamID string) (bool, error) {
	s.teamIDLock.RLock()
	defer s.teamIDLock.RUnlock()

	teamIDs, err := s.TeamIDs()

	if err != nil {
		return false, err
	}

	for _, candidateTeamID := range teamIDs {
		if teamID == candidateTeamID {
			return true, nil
		}
	}

	return false, nil
}

/* ********************* Team Name functions ********* */

// TeamName returns team name given a team ID.
func (s *State) TeamName(teamID string) (string, error) {
	s.teamNameLock.RLock()
	defer s.teamNameLock.RUnlock()

	name, ok := s.teamNames[teamID]

	if !ok {
		return "", fmt.Errorf("unregistered team ID: %s", teamID)
	}
	return name, nil
}

func (s *State) TeamNames() map[string]string {
	s.teamNameLock.RLock()
	defer s.teamNameLock.RUnlock()
	return s.teamNames
}

// SetTeamName writes out team name.
// This can only be done once per team.
func (s *State) SetTeamName(teamID, teamName string) error {
	s.teamNameFileLock.Lock()
	defer s.teamNameFileLock.Unlock()

	teamIDs, err := s.TeamIDs()
	if err != nil {
		return err
	}

	found := false
	for _, validTeamID := range teamIDs {
		if validTeamID == teamID {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("team ID not found in list of valid team IDs")
	}

	s.teamNameLock.RLock()
	_, ok := s.teamNames[teamID]
	s.teamNameLock.RUnlock()
	if ok {
		return ErrAlreadyRegistered
	}

	teamFilename := filepath.Join("teams", teamID)
	teamFile, err := s.Fs.OpenFile(teamFilename, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
	if os.IsExist(err) {  // This shouldn't ever hit, since we just checked, but strange things happen
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

func (s *State) SetTeamNames(teams map[string]string) error {
	return s.writeTeamNames(teams)
}

func (s *State) TeamIDFromName(teamName string) (string, error) {
	for name, id := range s.TeamNames() {
		if name == teamName {
			return id, nil
		}
	}

	return "", fmt.Errorf("team name not found")
}

func (s *State) DeleteTeamName(teamID string) error {
	newTeams := s.TeamNames()

	_, ok := newTeams[teamID];
    if ok {
        delete(newTeams, teamID)
    } else {
		return fmt.Errorf("team not found")
	}
	
	return s.writeTeamNames(newTeams)
}

func (s *State) writeTeamNames(teams map[string]string) error {
	s.teamNameFileLock.Lock()
	defer s.teamNameFileLock.Unlock()

	s.RemoveAll("teams")
	s.Mkdir("teams", 0755)

	// Write out all of the new team names
	for teamID, teamName := range teams {
		teamFilename := filepath.Join("teams", teamID)
		teamFile, err := s.Fs.OpenFile(teamFilename, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)

		if err != nil {
			return err
		}
		defer teamFile.Close()

		log.Printf("Setting team name [%s] in file %s", teamName, teamFilename)
		fmt.Fprintln(teamFile, teamName)
		teamFile.Close()
	}

	s.refreshNow <- true

	return nil
}

/* **************** Point log functions ************ */

// PointsLog retrieves the current points log.
func (s *State) PointsLog() award.List {
	s.pointsLock.RLock()
	ret := make(award.List, len(s.pointsLog))
	copy(ret, s.pointsLog)
	s.pointsLock.RUnlock()
	return ret
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

func (s *State) AwardPointsAtTime(teamID, category string, points int, when int64) error {
	return s.awardPointsAtTime(when, teamID, category, points)
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

func (s *State) PointExists(teamID string, cat string, points int) bool {
	s.pointsLock.RLock()
	defer s.pointsLock.RUnlock()
	for _, pointEntry := range s.pointsLog {
		if (pointEntry.TeamID == teamID) && (pointEntry.Category == cat) && (pointEntry.Points == points) {
			return true
		}
	}

	return false
}

func (s *State) PointExistsAtTime(teamID string, cat string, points int, when int64) bool {
	s.pointsLock.RLock()
	defer s.pointsLock.RUnlock()

	for _, pointEntry := range s.pointsLog {
		if (pointEntry.TeamID == teamID) && (pointEntry.Category == cat) && (pointEntry.Points == points) && (pointEntry.When == when) {
			return true
		}

		if (pointEntry.When > when) {  // Since the points log is sorted, we can bail out earlier, if we see that current points are from later than our event
			return false
		}
	}

	return false
}

func (s *State) flushPointsLog(newPoints award.List) error {
	s.pointsLogFileLock.Lock()
	defer s.pointsLogFileLock.Unlock()

	logf, err := s.OpenFile("points.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	defer logf.Close()

	if err != nil {
		return fmt.Errorf("Can't write to points log: %s", err)
	}
	
	for _, pointEntry := range newPoints {
		fmt.Fprintln(logf, pointEntry.String())
	}
	
	return nil
}

func (s *State) RemovePoints(teamID string, cat string, points int) error {
	s.pointsLock.Lock()
	defer s.pointsLock.Unlock()

	var newPoints award.List
	removed := false

	for _, pointEntry := range s.pointsLog {
		if (pointEntry.TeamID == teamID) && (pointEntry.Category == cat) && (pointEntry.Points == points) {
			removed = true
		} else {
			newPoints = append(newPoints, pointEntry)
		}
	}

	if (! removed) {
		return NoMatchingPointEntry
	}

	err := s.flushPointsLog(newPoints)

	if err != nil {
		return err
	}

	s.refreshNow <- true

	return nil
}

func (s *State) RemovePointsAtTime(teamID string, cat string, points int, when int64) error {
	s.pointsLock.Lock()
	defer s.pointsLock.Unlock()

	var newPoints award.List
	removed := false

	for _, pointEntry := range s.pointsLog {
		if (pointEntry.TeamID == teamID) && (pointEntry.Category == cat) && (pointEntry.Points == points) && (pointEntry.When == when) {
			removed = true
		} else {
			newPoints = append(newPoints, pointEntry)
		}
	}

	if (! removed) {
		return NoMatchingPointEntry
	}

	err := s.flushPointsLog(newPoints)

	if err != nil {
		return err
	}

	s.refreshNow <- true

	return nil
}

func (s *State) SetPoints(newPoints award.List) error {
	err := s.flushPointsLog(newPoints)

	if err != nil {
		return err
	}

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
		s.pointsLock.RLock()
		for _, e := range s.pointsLog {
			if awd.Equal(e) {
				duplicate = true
				break
			}
		}
		s.pointsLock.RUnlock()

		if duplicate {
			log.Print("Skipping duplicate points: ", awd.String())
		} else {
			log.Print("Award: ", awd.String())

			{
				s.pointsLogFileLock.Lock()
				
				logf, err := s.OpenFile("points.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					log.Print("Can't append to points log: ", err)
					s.pointsLogFileLock.Unlock()
					return
				}
				fmt.Fprintln(logf, awd.String())
				logf.Close()
				s.pointsLogFileLock.Unlock()
			}

			// Stick this on the cache too
			s.pointsLock.Lock()
			s.pointsLog = append(s.pointsLog, awd)
			s.pointsLock.Unlock()
		}

		if err := s.Remove(filename); err != nil {
			log.Print("Unable to remove new points file: ", err)
		}
	}
}

/* ******************* Message functions *********** */

// Messages retrieves the current messages.
func (s *State) Messages() string {
	return s.messages
}

// SetMessages sets the current message
func (s *State) SetMessages(message string) error {
	s.messageFileLock.Lock()
	defer s.messageFileLock.Unlock()

	err := afero.WriteFile(s, "messages.html", []byte(message), 0600)

	s.refreshNow <- true

	return err
}

/* ***************** Other utilitity functions ******* */

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

	s.pointsLogFileLock.Lock()
	s.Remove("points.log")
	s.pointsLogFileLock.Unlock()

	s.Remove("events.csv")

	s.messageFileLock.Lock()
	s.Remove("messages.html")
	s.messageFileLock.Unlock()

	s.Remove("mothd.log")
	s.RemoveAll("points.tmp")
	s.RemoveAll("points.new")

	s.teamNameFileLock.Lock()
	s.RemoveAll("teams")
	s.Mkdir("teams", 0755)
	s.teamNameFileLock.Unlock()

	// Open log file
	if err := s.reopenEventLog(); err != nil {
		log.Fatal(err)
	}
	s.LogEvent("init", "", "", "", 0)

	// Make sure various subdirectories exist
	s.Mkdir("points.tmp", 0755)
	s.Mkdir("points.new", 0755)

	// Preseed available team ids if file doesn't exist
	var teamIDs []string
	id := make([]byte, 8)
	for i := 0; i < 100; i++ {
		for i := range id {
			char := rand.Intn(len(DistinguishableChars))
			id[i] = DistinguishableChars[char]
		}
		teamIDs = append(teamIDs, string(id))
	}
	s.SetTeamIDs(teamIDs)

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
		fmt.Fprintln(f, "- 2519-10-31T00:00:00Z")
		f.Close()
	}

	s.messageFileLock.Lock()
	if f, err := s.Create("messages.html"); err == nil {
		fmt.Fprintln(f, "<!-- messages.html: put client broadcast messages here. -->")
		f.Close()
	}
	s.messageFileLock.Unlock()

	if f, err := s.Create("points.log"); err == nil {
		f.Close()
	}
}

// LogEvent writes to the event log
func (s *State) LogEvent(event, participantID, teamID, cat string, points int, extra ...string) {
	s.eventStream <- append(
		[]string{
			strconv.FormatInt(time.Now().Unix(), 10),
			event,
			participantID,
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
	

	// Re-read the points log
	{
		s.pointsLogFileLock.RLock()
		defer s.pointsLogFileLock.RUnlock()

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
			s.pointsLock.Lock()
			s.pointsLog = pointsLog
			s.pointsLock.Unlock()
		}
	}

	// Only do this if the teams directory has a newer mtime; directories with
	// hundreds of team names can cause NFS I/O storms
	{
		s.teamNameLock.Lock()
		defer s.teamNameLock.Unlock()
		s.teamNameFileLock.RLock()
		defer s.teamNameFileLock.RUnlock()

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

	// Re-read the messages file
	{
		s.messageFileLock.RLock()
		defer s.messageFileLock.RUnlock()

		if bMessages, err := afero.ReadFile(s, "messages.html"); err == nil {
			s.messages = string(bMessages)
		}
	}
}

func (s *State) refresh() {
	s.maybeInitialize()
	s.updateEnabled()
	if s.Enabled {
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
