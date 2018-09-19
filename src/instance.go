package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

type Instance struct {
	Base         string
	MothballDir  string
	StateDir     string
	ResourcesDir string
	Categories   map[string]*Mothball
	update       chan bool
}

func NewInstance(base, mothballDir, stateDir, resourcesDir string) (*Instance, error) {
	ctx := &Instance{
		Base:         strings.TrimRight(base, "/"),
		MothballDir:  mothballDir,
		StateDir:     stateDir,
		ResourcesDir: resourcesDir,
		Categories:   map[string]*Mothball{},
		update:       make(chan bool, 10),
	}

	// Roll over and die if directories aren't even set up
	if _, err := os.Stat(mothballDir); err != nil {
		return nil, err
	}
	if _, err := os.Stat(stateDir); err != nil {
		return nil, err
	}

	ctx.MaybeInitialize()

	return ctx, nil
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
		for i := 0; i <= 9999; i += 1 {
			fmt.Fprintf(f, "%04d\n", i)
		}
	}

	// Create initialized file that signals whether we're set up
	f, err := os.OpenFile(ctx.StatePath("initialized"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		log.Print(err)
	}
	defer f.Close()
	fmt.Fprintln(f, "Remove this file to reinitialize the contest")
}

func (ctx Instance) MothballPath(parts ...string) string {
	tail := path.Join(parts...)
	return path.Join(ctx.MothballDir, tail)
}

func (ctx *Instance) StatePath(parts ...string) string {
	tail := path.Join(parts...)
	return path.Join(ctx.StateDir, tail)
}

func (ctx *Instance) PointsLog() []*Award {
	var ret []*Award

	fn := ctx.StatePath("points.log")
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
		ret = append(ret, cur)
	}

	return ret
}

// awardPoints gives points to teamid in category.
// It first checks to make sure these are not duplicate points.
// This is not a perfect check, you can trigger a race condition here.
// It's just a courtesy to the user.
// The maintenance task makes sure we never have duplicate points in the log.
func (ctx *Instance) AwardPoints(teamid, category string, points int) error {
	a := Award{
		When:     time.Now(),
		TeamId:   teamid,
		Category: category,
		Points:   points,
	}

	for _, e := range ctx.PointsLog() {
		if a.Same(e) {
			return fmt.Errorf("Points already awarded to this team in this category")
		}
	}

	fn := fmt.Sprintf("%s-%s-%d", teamid, category, points)
	tmpfn := ctx.StatePath("points.tmp", fn)
	newfn := ctx.StatePath("points.new", fn)

	if err := ioutil.WriteFile(tmpfn, []byte(a.String()), 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpfn, newfn); err != nil {
		return err
	}

	ctx.update <- true
	log.Printf("Award %s %s %d", teamid, category, points)
	return nil
}

func (ctx *Instance) OpenCategoryFile(category string, parts ...string) (io.ReadCloser, error) {
	mb, ok := ctx.Categories[category]
	if !ok {
		return nil, fmt.Errorf("No such category: %s", category)
	}

	filename := path.Join(parts...)
	f, err := mb.Open(filename)
	return f, err
}

func (ctx *Instance) TeamName(teamId string) (string, error) {
	teamNameBytes, err := ioutil.ReadFile(ctx.StatePath("teams", teamId))
	teamName := strings.TrimSpace(string(teamNameBytes))
	return teamName, err
}
