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

// Environment contains instance state and authentication information.
type Environment struct {
	// State associated with the instance in the URL of this request.
	State State

	// TeamID of the team making the request.
	// A value here does not imply the request is authenticated:
	// this is just what the request specified as the team ID.
	TeamID TeamID

	// TeamData of the team making the request.

	TeamData string

	// Authenticated is true if the request is "logged in".
	// This happens when the HTTP request included a TeamID which the state recognizes as a registered team.
	Authenticated bool

	ctx *Context
}

// GetEnvironment builds an Environment from req.
// nil means no environemnt can be built: reply with 404 in this case.
func (ctx *Context) GetEnvironment(req *http.Request) *Environment {
	instance := req.PathValue("instance")
	state, found := ctx.State(instance)
	if !found {
		return nil
	}

	// Determine team id, if applicable
	var teamid TeamID
	if c, err := req.Cookie("teamid"); err == nil {
		teamid = TeamID(c.Value)
	}
	if username, password, ok := req.BasicAuth(); ok {
		if password != "" {
			teamid = TeamID(password)
		} else {
			// Fall back to team ID as the username, if no password is provided
			teamid = TeamID(username)
		}
	}

	// Look team data up. This is how the state tells us whether the teamid is properly authenticated.
	teamdata, authenticated := state.TeamData(teamid)

	return &Environment{
		State:         state,
		TeamID:        teamid,
		TeamData:      teamdata,
		Authenticated: authenticated,
		ctx:           ctx,
	}
}

// Log writes an event to a state's log.
func (e *Environment) Log(when time.Time, name string, cat CategoryID, puzzle PuzzleID, extra ...string) {
	evt := NewEvent(when, name, e.TeamID, cat, puzzle, extra...)
	slog.Debug("event",
		"name", name,
		"team", e.TeamID,
		"cat", cat,
		"puzzle", puzzle,
		"extra", extra,
	)
	if err := e.State.Log(evt); err != nil {
		slog.Error("logging error", "error", err.Error())
	}
}

// Unlocked returns the list of unlocked puzzles in cat.
//
// If all puzzles are unlocked,
// the list is terminated with a zero-value PuzzleID.
func (e *Environment) Unlocked(cat Category) []PuzzleID {
	if cat.Unlocked() {
		return nil
	}

	inventory := cat.Puzzles()

	last := -1
	for _, awd := range e.State.Awards().Values() {
		idx := slices.Index(inventory, awd.Puzzle)
		if idx > last {
			last = idx
		}
	}
	inventory = append(inventory, PuzzleID(""))
	return inventory[:last+2]
}

// StateExport is given to clients requesting the current state.
type StateExport struct {
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

func (ctx *Context) Options(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Allow", "GET, PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	w.WriteHeader(http.StatusOK)
}

// GetPuzzleFile serves static content from a Puzzle.
func (ctx *Context) GetPuzzleFile(w http.ResponseWriter, req *http.Request) {
	env := ctx.GetEnvironment(req)
	if env == nil {
		http.NotFound(w, req)
	}
	if !env.Authenticated {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	name := path.Clean(req.PathValue("name"))

	// Find category
	catid := CategoryID(req.PathValue("cat"))
	cat, err := ctx.Category(catid)
	if err != nil {
		http.NotFound(w, req)
		return
	}

	// If a puzzle id was provided, ensure it's unlocked
	puzzleid := PuzzleID(req.PathValue("puzzle"))
	if puzzleid != "" {
		unlocked := env.Unlocked(cat)
		if !slices.Contains(unlocked, puzzleid) {
			http.NotFound(w, req)
			return
		}
		name = path.Join(puzzleid.String(), name)
	}

	http.ServeFileFS(w, req, cat, name)
}

// PutAnswer checks answer correctness and records correct submissions
func (ctx *Context) PutAnswer(w http.ResponseWriter, req *http.Request) {
	now := time.Now()

	env := ctx.GetEnvironment(req)
	if env == nil {
		http.NotFound(w, req)
	}
	if !env.Authenticated {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	// Find category
	catid := CategoryID(req.PathValue("cat"))
	cat, err := ctx.Category(catid)
	if err != nil {
		http.NotFound(w, req)
		return
	}

	// Ensure this puzzle is unlocked
	puzzleid := PuzzleID(req.PathValue("puzzle"))
	unlocked := env.Unlocked(cat)
	if !slices.Contains(unlocked, puzzleid) {
		http.NotFound(w, req)
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
		env.Log(now, "wrong", catid, puzzleid)
		http.Error(w, "Wrong answer", http.StatusConflict)
		return
	}
	env.Log(now, "correct", catid, puzzleid)

	// Prepare to record the event solution.
	awd := NewAward(now, env.TeamID, catid, puzzleid)

	// Do this check after correctness
	// so people can waste a lot of time by not talking to their team-mates.
	if env.State.Awards().Contains(awd) {
		http.Error(w, "puzzle already solved by this team", http.StatusAccepted)
		return
	}

	// Now we can record the award.
	if err := env.State.RecordAward(awd); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Correct answer")
}

func (ctx *Context) GetStateCurrent(w http.ResponseWriter, req *http.Request) {
	now := time.Now()
	env := ctx.GetEnvironment(req)
	if env == nil {
		http.NotFound(w, req)
		return
	}

	env.sendState(w, now, http.StatusOK)
}

func (ctx *Context) PutStateTeamData(w http.ResponseWriter, req *http.Request) {
	now := time.Now()

	env := ctx.GetEnvironment(req)
	if env == nil {
		http.NotFound(w, req)
		return
	}

	// If they're already authenticated, they can't set team data again.
	if env.Authenticated {
		env.sendState(w, now, http.StatusConflict)
		return
	}

	// Read in what they're submitting
	data, ok := ReadPutData(w, req, MaxPublicLength)
	if !ok {
		// ReadPutData already sends a response
		return
	}

	if err := env.State.SetTeamData(env.TeamID, data); errors.Is(err, fs.ErrInvalid) {
		env.sendState(w, now, http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	env.TeamData = data
	env.Authenticated = true
	env.sendState(w, now, http.StatusOK)
	env.Log(now, "public", "", "", data)
}

func (env *Environment) sendState(w http.ResponseWriter, now time.Time, code int) {
	export := &StateExport{
		Running:          env.State.Enabled(now),
		OpenRegistration: env.State.OpenRegistration(),
		Hours:            env.State.Hours(),
		ClientConfig:     env.State.ClientConfig(),
		Errors:           env.State.Errors(),
	}

	// Anonymize team IDs
	export.TeamData = make(map[TeamID]string)
	exportIDs := make(map[TeamID]TeamID)
	if env.Authenticated {
		exportIDs[env.TeamID] = "self"
		export.TeamData["self"] = env.TeamData
	}
	for _, team := range env.State.Teams() {
		if team == env.TeamID {
			continue
		}
		n := len(export.TeamData)
		anonTeam := TeamID(strconv.Itoa(n))
		exportIDs[team] = anonTeam
		export.TeamData[anonTeam], _ = env.State.TeamData(team)
	}

	// Anonymize team IDs in points log, and write out team names.
	export.Awards = NewAwardList()
	for logno, awd := range env.State.Awards().Values() {
		if anonTeam, ok := exportIDs[awd.Team]; ok {
			awd.Team = anonTeam
		} else {
			awd.Team = TeamID(strconv.Itoa(logno))
		}
		export.Awards.Append(awd)
	}

	if env.Authenticated {
		export.Puzzles = make(map[CategoryID][]PuzzleID)
		for _, catid := range env.State.Categories(now) {
			cat, err := env.ctx.Category(catid)
			if err != nil {
				export.Errors = append(export.Errors, err.Error())
				continue
			}
			export.Puzzles[catid] = env.Unlocked(cat)
		}
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
