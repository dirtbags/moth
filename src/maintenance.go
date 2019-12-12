package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
	for _, a := range ctx.State.PointsLog("") {
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
	ret.Points = ctx.State.PointsLog(teamId)

	teamNumbersById := map[string]int{}
	for nr, a := range ret.Points {
		teamNumber, ok := teamNumbersById[a.TeamId]
		if !ok {
			teamName, err := ctx.State.TeamName(a.TeamId)
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
	ctx.State.Initialize()

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
	teamList := ctx.State.getTeams()

	// For any new team IDs, set their next attempt time to right now
	now := time.Now()
	added := 0
	for teamName, _ := range teamList {
		ctx.nextAttemptMutex.RLock()
		_, ok := ctx.nextAttempt[teamName]
		ctx.nextAttemptMutex.RUnlock()

		if !ok {
			ctx.nextAttemptMutex.Lock()
			ctx.nextAttempt[teamName] = now
			ctx.nextAttemptMutex.Unlock()

			added += 1
		}
	}

	// For any removed team IDs, remove them
	removed := 0
	ctx.nextAttemptMutex.Lock() // XXX: This could be less of a cludgel
	for teamName, _ := range ctx.nextAttempt {
		if _, ok := teamList[teamName]; !ok {
			delete(ctx.nextAttempt, teamName)
		}
	}
	ctx.nextAttemptMutex.Unlock()

	if (added > 0) || (removed > 0) {
		log.Printf("Team IDs updated: %d added, %d removed", added, removed)
	}
}

func (ctx *Instance) UpdateConfig() {
	// Handle export manifest
	if _, err := ctx.State.getConfig("export_manifest"); err == nil {
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
		if ctx.State.isEnabled() {
			ctx.tidy()
			ctx.readTeams()
			ctx.generatePuzzleList()
			ctx.generatePointsLog("")
		}
		select {
		case <-ctx.update:
			// log.Print("Forced update")
		case <-time.After(maintenanceInterval):
			// log.Print("Housekeeping...")
		}
	}
}
