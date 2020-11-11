package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

type RuntimeConfig struct {
	export_manifest bool
}

type Instance struct {
	Base            string
	MothballDir     string
	StateDir        string
	ThemeDir        string
	AttemptInterval time.Duration
	UseXForwarded   bool

	Runtime RuntimeConfig

	categories        map[string]*Mothball
	MaxPointsUnlocked map[string]int
	update            chan bool
	jPuzzleList       []byte
	jPointsLog        []byte
	eventStream       chan string
	eventLogWriter    io.WriteCloser
	nextAttempt       map[string]time.Time
	nextAttemptMutex  *sync.RWMutex
	mux               *http.ServeMux
}

func (ctx *Instance) Initialize() error {
	// Roll over and die if directories aren't even set up
	if _, err := os.Stat(ctx.MothballDir); err != nil {
		return err
	}
	if _, err := os.Stat(ctx.StateDir); err != nil {
		return err
	}
	if f, err := os.OpenFile(ctx.StatePath("events.log"), os.O_RDWR|os.O_CREATE, 0644); err != nil {
		return err
	} else {
		// This stays open for the life of the process
		ctx.eventLogWriter = f
	}

	ctx.Base = strings.TrimRight(ctx.Base, "/")
	ctx.categories = map[string]*Mothball{}
	ctx.update = make(chan bool, 10)
	ctx.eventStream = make(chan string, 80)
	ctx.nextAttempt = map[string]time.Time{}
	ctx.nextAttemptMutex = new(sync.RWMutex)
	ctx.mux = http.NewServeMux()

	ctx.BindHandlers()
	ctx.MaybeInitialize()

	return nil
}

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

func (ctx *Instance) MaybeInitialize() {
	// Only do this if it hasn't already been done
	if _, err := os.Stat(ctx.StatePath("initialized")); err == nil {
		return
	}
	log.Print("initialized file missing, re-initializing")

	// Remove any extant control and state files
	os.Remove(ctx.StatePath("until"))
	os.Remove(ctx.StatePath("disabled"))
	os.Remove(ctx.StatePath("points.log"))
	os.Remove(ctx.StatePath("events.log"))

	os.RemoveAll(ctx.StatePath("points.tmp"))
	os.RemoveAll(ctx.StatePath("points.new"))
	os.RemoveAll(ctx.StatePath("teams"))

	// Make sure various subdirectories exist
	os.Mkdir(ctx.StatePath("points.tmp"), 0755)
	os.Mkdir(ctx.StatePath("points.new"), 0755)
	os.Mkdir(ctx.StatePath("teams"), 0755)

	// Preseed available team ids if file doesn't exist
	if f, err := os.OpenFile(ctx.StatePath("teamids.txt"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644); err == nil {
		defer f.Close()
		for i := 0; i <= 100; i += 1 {
			fmt.Fprintln(f, mktoken())
		}
	}

	// Record that we did all this
	ctx.LogEvent("init", "", "", "", 0)

	// Create initialized file that signals whether we're set up
	f, err := os.OpenFile(ctx.StatePath("initialized"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		log.Print(err)
	}
	defer f.Close()
	fmt.Fprintln(f, "Remove this file to reinitialize the contest")
}

func logstr(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// LogEvent writes to the event log
func (ctx *Instance) LogEvent(event, participantID, teamID, cat string, points int, extra ...string) {
	event = strings.ReplaceAll(event, " ", "-")

	msg := fmt.Sprintf(
		"%s %s %s %s %d",
		logstr(event),
		logstr(participantID),
		logstr(teamID),
		logstr(cat),
		points,
	)
	for _, x := range extra {
		msg = msg + " " + strings.ReplaceAll(x, " ", "-")
	}
	ctx.eventStream <- msg
}
func pathCleanse(parts []string) string {
	clean := make([]string, len(parts))
	for i := range parts {
		part := parts[i]
		part = strings.TrimLeft(part, ".")
		if p := strings.LastIndex(part, "/"); p >= 0 {
			part = part[p+1:]
		}
		clean[i] = part
	}
	return path.Join(clean...)
}

func (ctx Instance) MothballPath(parts ...string) string {
	tail := pathCleanse(parts)
	return path.Join(ctx.MothballDir, tail)
}

func (ctx *Instance) StatePath(parts ...string) string {
	tail := pathCleanse(parts)
	return path.Join(ctx.StateDir, tail)
}

func (ctx *Instance) ThemePath(parts ...string) string {
	tail := pathCleanse(parts)
	return path.Join(ctx.ThemeDir, tail)
}

func (ctx *Instance) TooFast(teamId string) bool {
	now := time.Now()

	ctx.nextAttemptMutex.RLock()
	next, _ := ctx.nextAttempt[teamId]
	ctx.nextAttemptMutex.RUnlock()

	ctx.nextAttemptMutex.Lock()
	ctx.nextAttempt[teamId] = now.Add(ctx.AttemptInterval)
	ctx.nextAttemptMutex.Unlock()

	return now.Before(next)
}

func (ctx *Instance) PointsLog(teamId string) AwardList {
	awardlist := AwardList{}

	fn := ctx.StatePath("points.log")

	f, err := os.Open(fn)
	if err != nil {
		log.Printf("Unable to open %s: %s", fn, err)
		return awardlist
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
		awardlist = append(awardlist, cur)
	}

	return awardlist
}

// AwardPoints gives points to teamId in category.
// It first checks to make sure these are not duplicate points.
// This is not a perfect check, you can trigger a race condition here.
// It's just a courtesy to the user.
// The maintenance task makes sure we never have duplicate points in the log.
func (ctx *Instance) AwardPoints(teamId, category string, points int) error {
	a := Award{
		When:     time.Now(),
		TeamId:   teamId,
		Category: category,
		Points:   points,
	}

	_, err := ctx.TeamName(teamId)
	if err != nil {
		return fmt.Errorf("No registered team with this hash")
	}

	for _, e := range ctx.PointsLog("") {
		if a.Same(e) {
			return fmt.Errorf("Points already awarded to this team in this category")
		}
	}

	fn := fmt.Sprintf("%s-%s-%d", teamId, category, points)
	tmpfn := ctx.StatePath("points.tmp", fn)
	newfn := ctx.StatePath("points.new", fn)

	if err := ioutil.WriteFile(tmpfn, []byte(a.String()), 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpfn, newfn); err != nil {
		return err
	}

	ctx.update <- true
	log.Printf("Award %s %s %d", teamId, category, points)
	return nil
}

func (ctx *Instance) OpenCategoryFile(category string, parts ...string) (io.ReadCloser, error) {
	mb, ok := ctx.categories[category]
	if !ok {
		return nil, fmt.Errorf("No such category: %s", category)
	}

	filename := path.Join(parts...)
	f, err := mb.Open(filename)
	return f, err
}

func (ctx *Instance) ValidTeamId(teamId string) bool {
	ctx.nextAttemptMutex.RLock()
	_, ok := ctx.nextAttempt[teamId]
	ctx.nextAttemptMutex.RUnlock()

	return ok
}

func (ctx *Instance) TeamName(teamId string) (string, error) {
	teamNameBytes, err := ioutil.ReadFile(ctx.StatePath("teams", teamId))
	teamName := strings.TrimSpace(string(teamNameBytes))
	return teamName, err
}
