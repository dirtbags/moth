package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

// Stuff people with mediocre handwriting could write down unambiguously, and can be entered without holding down shift
const distinguishableChars = "234678abcdefhijkmnpqrtwxyz="

func mktoken() string {
	a := make([]byte, 8)
	for i := range a {
		char := rand.Intn(len(distinguishableChars))
		a[i] = distinguishableChars[char]
	}
	return string(a)
}

type StateExport struct {
	TeamNames map[string]string
	PointsLog []Award
	Messages  []string
}

// We use the filesystem for synchronization between threads.
// The only thing State methods need to know is the path to the state directory.
type State struct {
	Component
	Enabled bool
	update  chan bool
}

func NewState(baseDir string) *State {
	return &State{
		Component: Component{
			baseDir: baseDir,
		},
		Enabled: true,
		update:  make(chan bool, 10),
	}
}

// Check a few things to see if this state directory is "enabled".
func (s *State) UpdateEnabled() {
	if _, err := os.Stat(s.path("enabled")); os.IsNotExist(err) {
		s.Enabled = false
		log.Print("Suspended: enabled file missing")
		return
	}

	nextEnabled := true
	untilFile, err := os.Open(s.path("hours"))
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
			log.Printf("Misformatted line in hours file")
		}
		line = strings.TrimSpace(line)
		until, err := time.Parse(time.RFC3339, line)
		if err != nil {
			log.Printf("Suspended: Unparseable until date: %s", line)
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
	teamFile := s.path("teams", teamId)
	teamNameBytes, err := ioutil.ReadFile(teamFile)
	teamName := strings.TrimSpace(string(teamNameBytes))

	if os.IsNotExist(err) {
		return "", fmt.Errorf("Unregistered team ID: %s", teamId)
	} else if err != nil {
		return "", fmt.Errorf("Unregistered team ID: %s (%s)", teamId, err)
	}

	return teamName, nil
}

// Write out team name. This can only be done once.
func (s *State) SetTeamName(teamId string, teamName string) error {
	teamFile := s.path("teams", teamId)
	err := ioutil.WriteFile(teamFile, []byte(teamName), os.ModeExclusive|0644)
	return err
}

// Retrieve the current points log
func (s *State) PointsLog() []*Award {
	pointsFile := s.path("points.log")
	f, err := os.Open(pointsFile)
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

// Return an exportable points log,
// This anonymizes teamId with either an integer, or the string "self"
// for the requesting teamId.
func (s *State) Export(teamId string) *StateExport {
	teamName, _ := s.TeamName(teamId)

	pointsLog := s.PointsLog()

	export := StateExport{
		PointsLog: make([]Award, len(pointsLog)),
		Messages:  make([]string, 0, 10),
		TeamNames: map[string]string{"self": teamName},
	}

	// Read in messages
	messagesFile := s.path("messages.txt")
	if f, err := os.Open(messagesFile); err != nil {
		log.Print(err)
	} else {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			message := scanner.Text()
			if strings.HasPrefix(message, "#") {
				continue
			}
			export.Messages = append(export.Messages, message)
		}
	}

	// Read in points
	exportIds := map[string]string{teamId: "self"}
	for logno, award := range pointsLog {
		exportAward := award
		if id, ok := exportIds[award.TeamId]; ok {
			exportAward.TeamId = id
		} else {
			exportId := strconv.Itoa(logno)
			exportAward.TeamId = exportId
			exportIds[award.TeamId] = exportAward.TeamId

			name, err := s.TeamName(award.TeamId)
			if err != nil {
				name = "Rodney" // https://en.wikipedia.org/wiki/Rogue_(video_game)#Gameplay
			}
			export.TeamNames[exportId] = name
		}
		export.PointsLog[logno] = *exportAward
	}

	return &export
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
	tmpfn := s.path("points.tmp", fn)
	newfn := s.path("points.new", fn)

	if err := ioutil.WriteFile(tmpfn, []byte(a.String()), 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpfn, newfn); err != nil {
		return err
	}

	s.update <- true
	return nil
}

// collectPoints gathers up files in points.new/ and appends their contents to points.log,
// removing each points.new/ file as it goes.
func (s *State) collectPoints() {
	files, err := ioutil.ReadDir(s.path("points.new"))
	if err != nil {
		log.Print(err)
		return
	}
	for _, f := range files {
		filename := s.path("points.new", f.Name())
		awardstr, err := ioutil.ReadFile(filename)
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

			logf, err := os.OpenFile(s.path("points.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Print("Can't append to points log: ", err)
				return
			}
			fmt.Fprintln(logf, award.String())
			logf.Close()
		}

		if err := os.Remove(filename); err != nil {
			log.Print("Unable to remove new points file: ", err)
		}
	}
}

func (s *State) maybeInitialize() {
	// Are we supposed to re-initialize?
	if _, err := os.Stat(s.path("initialized")); !os.IsNotExist(err) {
		return
	}

	log.Print("initialized file missing, re-initializing")

	// Remove any extant control and state files
	os.Remove(s.path("enabled"))
	os.Remove(s.path("until"))
	os.Remove(s.path("points.log"))
	os.Remove(s.path("messages.txt"))
	os.RemoveAll(s.path("points.tmp"))
	os.RemoveAll(s.path("points.new"))
	os.RemoveAll(s.path("teams"))

	// Make sure various subdirectories exist
	os.Mkdir(s.path("points.tmp"), 0755)
	os.Mkdir(s.path("points.new"), 0755)
	os.Mkdir(s.path("teams"), 0755)

	// Preseed available team ids if file doesn't exist
	if f, err := os.OpenFile(s.path("teamids.txt"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644); err == nil {
		defer f.Close()
		for i := 0; i <= 100; i += 1 {
			fmt.Fprintln(f, mktoken())
		}
	}

	// Create some files
	ioutil.WriteFile(
		s.path("initialized"),
		[]byte("state/initialized: remove to re-initialize the contest\n"),
		0644,
	)
	ioutil.WriteFile(
		s.path("enabled"),
		[]byte("state/enabled: remove to suspend the contest\n"),
		0644,
	)
	ioutil.WriteFile(
		s.path("hours"),
		[]byte(
			"# state/hours: when the contest is enabled\n"+
				"# Lines starting with + enable, with - disable.\n"+
				"\n"+
				"+ 1970-01-01T00:00:00Z\n"+
				"- 3019-10-31T00:00:00Z\n",
		),
		0644,
	)
	ioutil.WriteFile(
		s.path("messages.txt"),
		[]byte(fmt.Sprintf("[%s] Initialized.\n", time.Now().UTC().Format(time.RFC3339))),
		0644,
	)
	ioutil.WriteFile(
		s.path("points.log"),
		[]byte(""),
		0644,
	)
}

func (s *State) Run(updateInterval time.Duration) {
	for {
		s.maybeInitialize()
		s.UpdateEnabled()
		if s.Enabled {
			s.collectPoints()
		}

		// Wait for something to happen
		select {
		case <-s.update:
		case <-time.After(updateInterval):
		}
	}
}
