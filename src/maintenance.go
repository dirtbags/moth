package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type PuzzleMap struct {
	Points int
	Path   string
}

func (pm *PuzzleMap) MarshalJSON() ([]byte, error) {
	if pm == nil {
		return []byte("null"), nil
	}

	jPath, err := json.Marshal(pm.Path)
	if err != nil {
		return nil, err
	}

	ret := fmt.Sprintf("[%d,%s]", pm.Points, string(jPath))
	return []byte(ret), nil
}

func (ctx *Instance) generatePuzzleList() error {
	maxByCategory := map[string]int{}
	for _, a := range ctx.PointsLog() {
		if a.Points > maxByCategory[a.Category] {
			maxByCategory[a.Category] = a.Points
		}
	}

	ret := map[string][]PuzzleMap{}
	for catName, mb := range ctx.Categories {
		mf, err := mb.Open("map.txt")
		if err != nil {
			return err
		}
		defer mf.Close()

		pm := make([]PuzzleMap, 0, 30)
		completed := true
		scanner := bufio.NewScanner(mf)
		for scanner.Scan() {
			line := scanner.Text()

			var pointval int
			var dir string

			n, err := fmt.Sscanf(line, "%d %s", &pointval, &dir)
			if err != nil {
				return err
			} else if n != 2 {
				return fmt.Errorf("Parsing map for %s: short read", catName)
			}

			pm = append(pm, PuzzleMap{pointval, dir})

			if pointval > maxByCategory[catName] {
				completed = false
				break
			}
		}
		if completed {
			pm = append(pm, PuzzleMap{0, ""})
		}

		ret[catName] = pm
	}

	jpl, err := json.Marshal(ret)
	if err == nil {
		ctx.jPuzzleList = jpl
	}
	return err
}

func (ctx *Instance) generatePointsLog() error {
	var ret struct {
		Teams  map[string]string `json:"teams"`
		Points []*Award          `json:"points"`
	}
	ret.Teams = map[string]string{}
	ret.Points = ctx.PointsLog()

	teamNumbersById := map[string]int{}
	for nr, a := range ret.Points {
		teamNumber, ok := teamNumbersById[a.TeamId]
		if !ok {
			teamName, err := ctx.TeamName(a.TeamId)
			if err != nil {
				teamName = "[unregistered]"
			}
			teamNumber = nr
			teamNumbersById[a.TeamId] = teamNumber
			ret.Teams[strconv.FormatInt(int64(teamNumber), 16)] = teamName
		}
		a.TeamId = strconv.FormatInt(int64(teamNumber), 16)
	}

	jpl, err := json.Marshal(ret)
	if err == nil {
		ctx.jPointsLog = jpl
	}
	return err
}

// maintenance runs
func (ctx *Instance) tidy() {
	// Do they want to reset everything?
	ctx.MaybeInitialize()

	// Skip if we've been disabled
	if _, err := os.Stat(ctx.StatePath("disabled")); err == nil {
		log.Print("disabled file found, suspending maintenance")
		return
	}

	// Skip if we've expired
	untilspec, err := ioutil.ReadFile(ctx.StatePath("until"))
	if err == nil {
		until, err := time.Parse(time.RFC3339, string(untilspec))
		if err != nil {
			log.Printf("Unparseable date in until file: %v", until)
		} else {
			if until.Before(time.Now()) {
				log.Print("until file time reached, suspending maintenance")
				return
			}
		}
	}

	// Refresh all current categories
	for categoryName, mb := range ctx.Categories {
		if err := mb.Refresh(); err != nil {
			// Backing file vanished: remove this category
			log.Printf("Removing category: %s: %s", categoryName, err)
			mb.Close()
			delete(ctx.Categories, categoryName)
		}
	}

	// Any new categories?
	files, err := ioutil.ReadDir(ctx.MothballPath())
	if err != nil {
		log.Printf("Error listing mothballs: %s", err)
	}
	for _, f := range files {
		filename := f.Name()
		filepath := ctx.MothballPath(filename)
		if !strings.HasSuffix(filename, ".mb") {
			continue
		}
		categoryName := strings.TrimSuffix(filename, ".mb")

		if _, ok := ctx.Categories[categoryName]; !ok {
			mb, err := OpenMothball(filepath)
			if err != nil {
				log.Printf("Error opening %s: %s", filepath, err)
				continue
			}
			log.Printf("New category: %s", filename)
			ctx.Categories[categoryName] = mb
		}
	}
}

// collectPoints gathers up files in points.new/ and appends their contents to points.log,
// removing each points.new/ file as it goes.
func (ctx *Instance) collectPoints() {
	logf, err := os.OpenFile(ctx.StatePath("points.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Can't append to points log: %s", err)
		return
	}
	defer logf.Close()

	files, err := ioutil.ReadDir(ctx.StatePath("points.new"))
	if err != nil {
		log.Printf("Error reading packages: %s", err)
	}
	for _, f := range files {
		filename := ctx.StatePath("points.new", f.Name())
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
		for _, e := range ctx.PointsLog() {
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

// maintenance is the goroutine that runs a periodic maintenance task
func (ctx *Instance) Maintenance(maintenanceInterval time.Duration) {
	for {
		ctx.tidy()
		ctx.collectPoints()
		ctx.generatePuzzleList()
		ctx.generatePointsLog()
		select {
		case <-ctx.update:
			// log.Print("Forced update")
		case <-time.After(maintenanceInterval):
			// log.Print("Housekeeping...")
		}
	}
}
