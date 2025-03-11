package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/dirtbags/moth/v5/pkg/period"
)

var ErrNoSuchCategory = errors.New("no such category")

// A State provides the functionality needed for MOTH to store state.
//
// States are expected to implement their own caching, if needed.
type State interface {
	// Enabled returns true if the server should be enabled.
	Enabled(time.Time) bool

	// OpenRegistration returns true if SetTeamData will accept any TeamID.
	OpenRegistration() bool

	// ClientConfig is passed verbatim to the client.
	ClientConfig() map[string]string

	// Errors displayed in the user interface.
	//
	// States should use this to communicate problems to the administrator,
	// since nobody ever looks at the log files.
	Errors() []string

	// Hours when this instance expects to be enabled.
	// The operator can manually disable at any time.
	Hours() []period.Period

	// Categories present at a given time.
	Categories(time.Time) []CategoryID

	// Teams returns all registered team IDs.
	Teams() []TeamID

	// TeamData returns data for a team.
	TeamData(TeamID) (data string, ok bool)

	// SetTeamData records public team data.
	//
	// This is equivalent to registering a team.
	// It returns fs.ErrInvalid if the team ID is not allowed to register,
	// and fs.ErrExist if the team ID is already registered.
	SetTeamData(TeamID, string) error

	// Awards returns a time-ordered list of recorded awards.
	Awards() *AwardList

	// RecordAward commits this award to the award log.
	RecordAward(Award) error

	// Log an event.
	Log(Event) error
}

// StateStore provides states by name.
type StateStore interface {
	// State retrieves the named state.
	State(name string) (state State, found bool)

	// Refresh and re-sync with underlying storage.
	// now is provided to make unit tests simpler.
	Refresh(now time.Time) error
}

type Category interface {
	// ServeFile provides static files to clients.
	// MOTH will ensure a path is properly unlocked.
	//
	// It is the responsibility of a Category implementation to hide
	// metadata, such as Answers and Puzzles.
	// "index.html" should render content directly,
	// and not redirect to the directory:
	// this means you cannot call http.ServeFile.
	ServeFile(w http.ResponseWriter, r *http.Request, name string)

	// Unlocked will return true if every puzzle in this category should be open.
	//
	// If false, puzzles will "unlock" in the order returned by Puzzles().
	Unlocked() bool

	// Puzzles in this category, ordered by when they should be unlocked.
	Puzzles() []PuzzleID

	// Answers for a puzzle.
	Answers(PuzzleID) []string
}

// CategoryStore provides categories by name.
type CategoryStore interface {
	// Category retrieves the named category.
	Category(cat CategoryID) (Category, error)

	// Refresh and re-sync with underlying storage.
	// now is provided to make unit tests simpler.
	Refresh(now time.Time) error
}
