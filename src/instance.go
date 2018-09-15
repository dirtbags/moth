package main

import (
	"os"
	"log"
	"bufio"
	"fmt"
	"time"
	"io/ioutil"
	"path"
	"strings"
)

type Instance struct {
	Base string
	MothballDir string
	StateDir string
	Categories map[string]*Mothball
}

func NewInstance(base, mothballDir, stateDir string) (*Instance, error) {
	ctx := &Instance{
		Base: strings.TrimRight(base, "/"),
		MothballDir: mothballDir,
		StateDir: stateDir,
	}

	// Roll over and die if directories aren't even set up
	if _, err := os.Stat(mothballDir); err != nil {
		return nil, err
	}
	if _, err := os.Stat(stateDir); err != nil {
		return nil, err
	}

	ctx.Initialize()
	
	return ctx, nil
}

func (ctx *Instance) Initialize () {
	// Make sure points directories exist
	os.Mkdir(ctx.StatePath("points.tmp"), 0755)
	os.Mkdir(ctx.StatePath("points.new"), 0755)

	// Preseed available team ids if file doesn't exist
	if f, err := os.OpenFile(ctx.StatePath("teamids.txt"), os.O_WRONLY | os.O_CREATE | os.O_EXCL, 0644); err == nil {
		defer f.Close()
		for i := 0; i <= 9999; i += 1 {
			fmt.Fprintf(f, "%04d\n", i)
		}
	}
	
	if f, err := os.OpenFile(ctx.StatePath("initialized"), os.O_WRONLY | os.O_CREATE | os.O_EXCL, 0644); err == nil {
		defer f.Close()
		fmt.Println("Remove this file to reinitialize the contest")
	}
}

func (ctx Instance) MothballPath(parts ...string) string {
	tail := path.Join(parts...)
	return path.Join(ctx.MothballDir, tail)
}

func (ctx *Instance) StatePath(parts ...string) string {
	tail := path.Join(parts...)
	return path.Join(ctx.StateDir, tail)
}


func (ctx *Instance) PointsLog() []Award {
	var ret []Award

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
		ret = append(ret, *cur)
	}
	
	return ret
}

// awardPoints gives points points to team teamid in category category
func (ctx *Instance) AwardPoints(teamid string, category string, points int) error {
	fn := fmt.Sprintf("%s-%s-%d", teamid, category, points)
	tmpfn := ctx.StatePath("points.tmp", fn)
	newfn := ctx.StatePath("points.new", fn)
	
	contents := fmt.Sprintf("%d %s %s %d\n", time.Now().Unix(), teamid, category, points)
	
	if err := ioutil.WriteFile(tmpfn, []byte(contents), 0644); err != nil {
		return err
	}
	
	if err := os.Rename(tmpfn, newfn); err != nil {
		return err
	}
	
	log.Printf("Award %s %s %d", teamid, category, points)
	return nil
}
