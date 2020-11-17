package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

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

func (ctx *Instance) generatePuzzleList() {
	maxByCategory := map[string]int{}
	for _, a := range ctx.PointsLog("") {
		if a.Points > maxByCategory[a.Category] {
			maxByCategory[a.Category] = a.Points
		}
	}

	ret := map[string][]PuzzleMap{}
	for catName, mb := range ctx.categories {
		filtered_puzzlemap := make([]PuzzleMap, 0, 30)
		completed := true

		for _, pm := range mb.puzzlemap {
			filtered_puzzlemap = append(filtered_puzzlemap, pm)

			if pm.Points > maxByCategory[catName] {
				completed = false
				maxByCategory[catName] = pm.Points
				break
			}
		}

		if completed {
			filtered_puzzlemap = append(filtered_puzzlemap, PuzzleMap{0, ""})
		}

		ret[catName] = filtered_puzzlemap
	}

	// Cache the unlocked points for use in other functions
	ctx.MaxPointsUnlocked = maxByCategory

	jpl, err := json.Marshal(ret)
	if err != nil {
		log.Printf("Marshalling puzzles.js: %v", err)
		return
	}
	ctx.jPuzzleList = jpl
}

func (ctx *Instance) generatePointsLog(teamId string) []byte {
	var ret struct {
		Teams  map[string]string `json:"teams"`
		Points []*Award          `json:"points"`
	}
	ret.Teams = map[string]string{}
	ret.Points = ctx.PointsLog(teamId)

	teamNumbersById := map[string]int{}
	for nr, a := range ret.Points {
		teamNumber, ok := teamNumbersById[a.TeamId]
		if !ok {
			teamName, err := ctx.TeamName(a.TeamId)
			if err != nil {
				teamName = "Rodney" // https://en.wikipedia.org/wiki/Rogue_(video_game)#Gameplay
			}
			teamNumber = nr
			teamNumbersById[a.TeamId] = teamNumber
			ret.Teams[strconv.FormatInt(int64(teamNumber), 16)] = teamName
		}
		a.TeamId = strconv.FormatInt(int64(teamNumber), 16)
	}

	jpl, err := json.Marshal(ret)
	if err != nil {
		log.Printf("Marshalling points.js: %v", err)
		return nil
	}

	if len(teamId) == 0 {
		ctx.jPointsLog = jpl
	}
	return jpl
}

// maintenance runs
func (ctx *Instance) tidy() {
	// Do they want to reset everything?
	ctx.MaybeInitialize()

	// Check set config
	ctx.UpdateConfig()

	// Refresh all current categories
	for categoryName, mb := range ctx.categories {
		if err := mb.Refresh(); err != nil {
			// Backing file vanished: remove this category
			log.Printf("Removing category: %s: %s", categoryName, err)
			mb.Close()
			delete(ctx.categories, categoryName)
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

		if _, ok := ctx.categories[categoryName]; !ok {
			mb, err := OpenMothball(filepath)
			if err != nil {
				log.Printf("Error opening %s: %s", filepath, err)
				continue
			}
			log.Printf("New category: %s", filename)
			ctx.categories[categoryName] = mb
		}
	}
}

// readTeams reads in the list of team IDs,
// so we can quickly validate them.
func (ctx *Instance) readTeams() {
	filepath := ctx.StatePath("teamids.txt")
	teamids, err := os.Open(filepath)
	if err != nil {
		log.Printf("Error openining %s: %s", filepath, err)
		return
	}
	defer teamids.Close()

	// List out team IDs
	newList := map[string]bool{}
	scanner := bufio.NewScanner(teamids)
	for scanner.Scan() {
		teamId := scanner.Text()
		if (teamId == "..") || strings.ContainsAny(teamId, "/") {
			log.Printf("Dangerous team ID dropped: %s", teamId)
			continue
		}
		newList[scanner.Text()] = true
	}

	// For any new team IDs, set their next attempt time to right now
	now := time.Now()
	added := 0
	for k, _ := range newList {
		ctx.nextAttemptMutex.RLock()
		_, ok := ctx.nextAttempt[k]
		ctx.nextAttemptMutex.RUnlock()

		if !ok {
			ctx.nextAttemptMutex.Lock()
			ctx.nextAttempt[k] = now
			ctx.nextAttemptMutex.Unlock()

			added += 1
		}
	}

	// For any removed team IDs, remove them
	removed := 0
	ctx.nextAttemptMutex.Lock() // XXX: This could be less of a cludgel
	for k, _ := range ctx.nextAttempt {
		if _, ok := newList[k]; !ok {
			delete(ctx.nextAttempt, k)
		}
	}
	ctx.nextAttemptMutex.Unlock()

	if (added > 0) || (removed > 0) {
		log.Printf("Team IDs updated: %d added, %d removed", added, removed)
	}
}

// collectPoints gathers up files in points.new/ and appends their contents to points.log,
// removing each points.new/ file as it goes.
func (ctx *Instance) collectPoints() {
	points := ctx.PointsLog("")

	pointsFilename := ctx.StatePath("points.log")
	pointsNewFilename := ctx.StatePath("points.log.new")

	// Yo, this is delicate.
	// If we have to return early, we must remove this file.
	// If the file's written and we move it successfully,
	// we need to remove all the little points files that built it.
	newPoints, err := os.OpenFile(pointsNewFilename, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
	if err != nil {
		log.Printf("Can't append to points log: %s", err)
		return
	}

	files, err := ioutil.ReadDir(ctx.StatePath("points.new"))
	if err != nil {
		log.Printf("Error reading packages: %s", err)
	}
	removearino := make([]string, 0, len(files))
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
		for _, e := range points {
			if award.Same(e) {
				duplicate = true
				break
			}
		}

		if duplicate {
			log.Printf("Skipping duplicate points: %s", award.String())
		} else {
			points = append(points, award)
		}
		removearino = append(removearino, filename)
	}

	sort.Stable(points)
	for _, point := range points {
		fmt.Fprintln(newPoints, point.String())
	}

	newPoints.Close()

	if err := os.Rename(pointsNewFilename, pointsFilename); err != nil {
		log.Printf("Unable to move %s to %s: %s", pointsFilename, pointsNewFilename, err)
		if err := os.Remove(pointsNewFilename); err != nil {
			log.Printf("Also couldn't remove %s: %s", pointsNewFilename, err)
		}
		return
	}

	for _, filename := range removearino {
		if err := os.Remove(filename); err != nil {
			log.Printf("Unable to remove %s: %s", filename, err)
		}
	}
}

func (ctx *Instance) isEnabled() bool {
	// Skip if we've been disabled
	if _, err := os.Stat(ctx.StatePath("disabled")); err == nil {
		log.Print("Suspended: disabled file found")
		return false
	}

	untilspec, err := ioutil.ReadFile(ctx.StatePath("until"))
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

func (ctx *Instance) UpdateConfig() {
	// Handle export manifest
	if _, err := os.Stat(ctx.StatePath("export_manifest")); err == nil {
		if !ctx.Runtime.export_manifest {
			log.Print("Enabling manifest export")
			ctx.Runtime.export_manifest = true
		}
	} else if ctx.Runtime.export_manifest {
		log.Print("Disabling manifest export")
		ctx.Runtime.export_manifest = false
	}

}

// maintenance is the goroutine that runs a periodic maintenance task
func (ctx *Instance) Maintenance(maintenanceInterval time.Duration) {
	for {
		if ctx.isEnabled() {
			ctx.tidy()
			ctx.readTeams()
			ctx.collectPoints()
			ctx.generatePuzzleList()
			ctx.generatePointsLog("")
		}
		select {
		case <-ctx.update:
			// log.Print("Forced update")
		case msg := <-ctx.eventStream:
			fmt.Fprintln(ctx.eventLogWriter, msg)
		case <-time.After(maintenanceInterval):
			// log.Print("Housekeeping...")
		}
	}
}
