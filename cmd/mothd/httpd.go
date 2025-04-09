package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/dirtbags/moth/v5/pkg/period"
)

const MaxAnswerLength = 4096
const MaxPublicLength = 4096

// SetHandlers sets up handlers on the context.
func (ctx *Context) InstallHandlers(baseURL string) {
	ctx.ServeMux = http.NewServeMux()
	baseURL = strings.TrimRight(baseURL, "/")
	webServer := http.FileServer(ctx.WebFS)
	ctx.Handle(baseURL+"/", http.StripPrefix(baseURL, webServer))
	ctx.HandleFunc("GET "+baseURL+"/{instance}/puzzles/{cat}/{name}", ctx.GetPuzzleFile)
	ctx.HandleFunc("GET "+baseURL+"/{instance}/puzzles/{cat}/{puzzle}/{name...}", ctx.GetPuzzleFile)
	ctx.HandleFunc("PUT "+baseURL+"/{instance}/puzzles/{cat}/{puzzle}/answer", ctx.PutAnswer)
	ctx.HandleFunc("GET "+baseURL+"/{instance}/state/current.json", ctx.GetStateCurrent)
	ctx.HandleFunc("PUT "+baseURL+"/{instance}/state/public", ctx.PutStateTeamData) // Deprecated
	ctx.HandleFunc("PUT "+baseURL+"/{instance}/state/teamdata", ctx.PutStateTeamData)
	ctx.HandleFunc("OPTIONS "+baseURL+"/", ctx.Options)
}

// ServeHTTP passes the request to h.ServeMux, then logs it
func (ctx *Context) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	slog.Debug("HTTP request",
		"addr", req.RemoteAddr,
		"method", req.Method,
		"URL", req.URL,
	)
	ctx.ServeMux.ServeHTTP(w, req)
}

// ReadPutData reads the data sent with a PUT request.
//
// It performs checks to avoid memory exhaustion
// and some other problems.
func ReadPutData(w http.ResponseWriter, req *http.Request, maxlen int64) (buf string, ok bool) {
	if req.ContentLength <= 0 {
		http.Error(w, "", http.StatusLengthRequired)
		return "", false
	}
	if req.ContentLength > maxlen {
		http.Error(w, "", http.StatusRequestEntityTooLarge)
		return "", false
	}

	b := new(bytes.Buffer)
	if _, err := io.Copy(b, req.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", false
	}
	return b.String(), true
}

// GetState looks up the named state instance, tries to authenticate against it,
// and returns an HTTPState.
//
// nil is returned if no state instance is found.
func (ctx *Context) GetState(req *http.Request) (hs *HTTPState, authenticated bool) {
	instance := req.PathValue("instance")
	state, found := ctx.State(instance)
	if !found {
		return nil, false
	}
	hs = &HTTPState{
		State: state,
	}

	if username, password, ok := req.BasicAuth(); !ok {
		if c, err := req.Cookie("teamid"); err != nil {
			return hs, false
		} else {
			hs.Team = TeamID(c.Value)
		}
	} else if username != "" {
		hs.Team = TeamID(username)
	} else {
		hs.Team = TeamID(password)
	}
	hs.TeamData, authenticated = hs.State.TeamData(hs.Team)

	return hs, authenticated
}

type HTTPState struct {
	State    State
	Team     TeamID
	TeamData string
}

// Log an event to a state instance.
func (hs *HTTPState) Log(when time.Time, name string, cat CategoryID, puzzle PuzzleID, extra ...string) {
	e := NewEvent(when, name, hs.Team, cat, puzzle, extra...)
	slog.Debug("event",
		"name", name,
		"team", hs.Team,
		"cat", cat,
		"puzzle", puzzle,
		"extra", extra,
	)
	if err := hs.State.Log(e); err != nil {
		slog.Error("logging error", "error", err.Error())
	}
}

// StateExport is given to clients requesting the current state.
type StateExport struct {
	// MothVersion contains the current version of MOTH.
	MothVersion string

	// Running is true if the server is updating the Awards log.
	Running bool

	// OpenRegistration is true if all team IDs are considered valid for setting team data.
	OpenRegistration bool

	// ClientConfig is passed through to the client.
	ClientConfig map[string]string

	// Errors lists server-side errors to be shown in the client, possibly to only administrators.
	Errors []string

	// Hours when this instance is expected to be running.
	Hours []period.Period

	// TeamData maps anonymized team id to data for all registered teams.
	TeamData map[TeamID]string

	// Awards which have been granted in this instance.
	Awards *AwardList

	// Puzzles lists currently open puzzle IDs in each category.
	Puzzles map[CategoryID][]PuzzleID
}

// Unlocked returns all currently unlocked puzzles in a category.
//
// If all puzzles in a category are unlocked,
// the list for that category is terminated with a zero-value PuzzleID.
func (hs *HTTPState) Unlocked(now time.Time, categories CategoryStore) (categoriesInventory map[CategoryID][]PuzzleID, errs []error) {
	categoriesInventory = make(map[CategoryID][]PuzzleID)
	isUnlocked := make(map[CategoryID]bool)
	for _, catid := range hs.State.Categories(now) {
		cat, err := categories.Category(catid)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		categoriesInventory[catid] = cat.Puzzles()
		isUnlocked[catid] = cat.Unlocked()
	}

	maxIndex := make(map[CategoryID]int)
	for _, awd := range hs.State.Awards().Values() {
		inventory := categoriesInventory[awd.Category]
		if inventory == nil {
			continue
		}

		puzzleIndex := slices.Index(inventory, awd.Puzzle) + 1
		if puzzleIndex > maxIndex[awd.Category] {
			maxIndex[awd.Category] = puzzleIndex
		}
	}

	for cat, inventory := range categoriesInventory {
		max := maxIndex[cat]
		if isUnlocked[cat] || (max == len(inventory)) {
			inventory = append(inventory, PuzzleID(""))
		} else {
			inventory = inventory[:max+1]
		}
		categoriesInventory[cat] = inventory
	}
	return categoriesInventory, errs
}

// validate finds the named instance,
// authenticates the user,
// looks up the category,
// and checks whether the puzzle is unlocked.
func (ctx *Context) validate(w http.ResponseWriter, req *http.Request, now time.Time) (hs *HTTPState, cat Category, ok bool) {
	catid := CategoryID(req.PathValue("cat"))
	puzzleid := PuzzleID(req.PathValue("puzzle"))

	var authenticated bool
	hs, authenticated = ctx.GetState(req)
	if hs == nil {
		http.NotFound(w, req)
		return
	}
	if !authenticated {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	if c, err := ctx.CategoryStore.Category(catid); err != nil {
		http.NotFound(w, req)
		return
	} else {
		cat = c
	}

	if puzzleid != "" {
		unlocked, _ := hs.Unlocked(now, ctx.CategoryStore)
		if !slices.Contains(unlocked[catid], puzzleid) {
			http.NotFound(w, req)
			return
		}
	}

	ok = true
	return
}

func (ctx *Context) Options(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Allow", "GET, PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	w.WriteHeader(http.StatusOK)
}

// GetPuzzleFile serves static content from a Puzzle.
func (ctx *Context) GetPuzzleFile(w http.ResponseWriter, req *http.Request) {
	_, cat, ok := ctx.validate(w, req, time.Now())
	if !ok {
		return
	}

	name := req.PathValue("name")
	if puzzle := req.PathValue("puzzle"); puzzle != "" {
		name = puzzle + "/" + name
	}
	name = path.Clean(name)
	if strings.HasSuffix(name, "/") {
		http.NotFound(w, req)
		return
	}
	if name == "" {
		name = "index.html"
	}

	cat.ServeFile(w, req, name)
}

// PutAnswer checks answer correctness and records correct submissions
func (ctx *Context) PutAnswer(w http.ResponseWriter, req *http.Request) {
	now := time.Now()
	catid := CategoryID(req.PathValue("cat"))
	puzzleid := PuzzleID(req.PathValue("puzzle"))

	hs, cat, ok := ctx.validate(w, req, now)
	if !ok {
		return
	}

	// Read in what they're submitting
	answer, ok := ReadPutData(w, req, MaxAnswerLength)
	if !ok {
		return
	}

	// Is it right?
	correct := cat.Answers(puzzleid)
	if !slices.Contains(correct, answer) {
		hs.Log(now, "wrong", catid, puzzleid)
		http.Error(w, "Wrong answer", http.StatusConflict)
		return
	}
	hs.Log(now, "correct", catid, puzzleid)

	// Prepare to record the event solution.
	awd := NewAward(now, hs.Team, catid, puzzleid)

	// Do this check after correctness
	// so people can waste a lot of time by not talking to their team-mates.
	if hs.State.Awards().Contains(awd) {
		http.Error(w, "puzzle already solved by this team", http.StatusAccepted)
		return
	}

	// Now we can record the award.
	if err := hs.State.RecordAward(awd); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Correct answer")
}

func (ctx *Context) GetStateCurrent(w http.ResponseWriter, req *http.Request) {
	hs, authenticated := ctx.GetState(req)
	if hs == nil {
		http.NotFound(w, req)
		return
	}

	ctx.sendState(w, hs, authenticated, http.StatusOK)
}

func (ctx *Context) PutStateTeamData(w http.ResponseWriter, req *http.Request) {
	now := time.Now()
	hs, _ := ctx.GetState(req)
	if hs == nil {
		http.NotFound(w, req)
		return
	}

	// Read in what they're submitting
	data, ok := ReadPutData(w, req, MaxPublicLength)
	if !ok {
		return
	}

	err := hs.State.SetTeamData(hs.Team, data)
	switch {
	case errors.Is(err, fs.ErrInvalid):
		ctx.sendState(w, hs, false, http.StatusUnauthorized)
	case errors.Is(err, fs.ErrExist):
		ctx.sendState(w, hs, true, http.StatusConflict)
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	default:
		hs.TeamData = data
		ctx.sendState(w, hs, true, http.StatusOK)
		hs.Log(now, "public", "", "", data)
	}
}

func (ctx *Context) sendState(w http.ResponseWriter, hs *HTTPState, authenticated bool, code int) {
	now := time.Now()

	export := &StateExport{
		MothVersion:      Version,
		Running:          hs.State.Enabled(now),
		OpenRegistration: hs.State.OpenRegistration(),
		Hours:            hs.State.Hours(),
		ClientConfig:     hs.State.ClientConfig(),
		Errors:           hs.State.Errors(),
	}

	// Anonymize team IDs
	export.TeamData = make(map[TeamID]string)
	exportIDs := make(map[TeamID]TeamID)
	if authenticated {
		exportIDs[hs.Team] = "self"
		export.TeamData["self"] = hs.TeamData
	}
	for _, team := range hs.State.Teams() {
		if team == hs.Team {
			continue
		}
		n := len(export.TeamData)
		anonTeam := TeamID(strconv.Itoa(n))
		exportIDs[team] = anonTeam
		export.TeamData[anonTeam], _ = hs.State.TeamData(team)
	}

	// Anonymize team IDs in points log, and write out team names.
	export.Awards = NewAwardList()
	for logno, awd := range hs.State.Awards().Values() {
		if anonTeam, ok := exportIDs[awd.Team]; ok {
			awd.Team = anonTeam
		} else {
			awd.Team = TeamID(strconv.Itoa(logno))
		}
		export.Awards.Append(awd)
	}

	if authenticated {
		puzzles, errs := hs.Unlocked(now, ctx.CategoryStore)
		for _, err := range errs {
			export.Errors = append(export.Errors, err.Error())
		}
		export.Puzzles = puzzles
	}

	r := new(bytes.Buffer)
	enc := json.NewEncoder(r)
	if err := enc.Encode(export); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(r.Len()))
	w.WriteHeader(code)
	if _, err := io.Copy(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
