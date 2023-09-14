package main

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/dirtbags/moth/v4/pkg/award"
)

// Category represents a puzzle category.
type Category struct {
	Name    string
	Puzzles []int
}

// ReadSeekCloser defines a struct that can read, seek, and close.
type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

// Configuration stores information about server configuration.
type Configuration struct {
	Devel bool
}

// StateExport is given to clients requesting the current state.
type StateExport struct {
	Config    Configuration
	Messages  string
	TeamNames map[string]string
	PointsLog award.List
	Puzzles   map[string][]int
}

// PuzzleProvider defines what's required to provide puzzles.
type PuzzleProvider interface {
	Open(cat string, points int, path string) (ReadSeekCloser, time.Time, error)
	Inventory() []Category
	CheckAnswer(cat string, points int, answer string) (bool, error)
	Mothball(cat string, w io.Writer) error
	Maintainer
}

// ThemeProvider defines what's required to provide a theme.
type ThemeProvider interface {
	Open(path string) (ReadSeekCloser, time.Time, error)
	Maintainer
}

// StateProvider defines what's required to provide MOTH state.
type StateProvider interface {
	Messages() string
	PointsLog() award.List
	TeamName(teamID string) (string, error)
	SetTeamName(teamID, teamName string) error
	AwardPoints(teamID string, cat string, points int) error
	LogEvent(event, teamID, cat string, points int, extra ...string)
	Maintainer
}

// Maintainer is something that can be maintained.
type Maintainer interface {
	// Maintain is the maintenance loop.
	// It will only be called once, when execution begins.
	// It's okay to just exit if there's no maintenance to be done.
	Maintain(updateInterval time.Duration)

	// refresh is a shortcut used internally for testing
	refresh()
}

// MothServer gathers together the providers that make up a MOTH server.
type MothServer struct {
	PuzzleProviders []PuzzleProvider
	Theme           ThemeProvider
	State           StateProvider
	Config          Configuration
}

// NewMothServer returns a new MothServer.
func NewMothServer(config Configuration, theme ThemeProvider, state StateProvider, puzzleProviders ...PuzzleProvider) *MothServer {
	return &MothServer{
		Config:          config,
		PuzzleProviders: puzzleProviders,
		Theme:           theme,
		State:           state,
	}
}

// NewHandler returns a new http.RequestHandler for the provided teamID.
func (s *MothServer) NewHandler(teamID string) MothRequestHandler {
	return MothRequestHandler{
		MothServer: s,
		teamID:     teamID,
	}
}

// MothRequestHandler provides http.RequestHandler for a MothServer.
type MothRequestHandler struct {
	*MothServer
	teamID string
}

// PuzzlesOpen opens a file associated with a puzzle.
// BUG(neale): Multiple providers with the same category name are not detected or handled well.
func (mh *MothRequestHandler) PuzzlesOpen(cat string, points int, path string) (r ReadSeekCloser, ts time.Time, err error) {
	export := mh.exportStateIfRegistered(true)
	found := false
	for _, p := range export.Puzzles[cat] {
		if p == points {
			found = true
		}
	}
	if !found {
		return nil, time.Time{}, fmt.Errorf("puzzle does not exist or is locked")
	}

	// Try every provider until someone doesn't return an error
	for _, provider := range mh.PuzzleProviders {
		r, ts, err = provider.Open(cat, points, path)
		if err != nil {
			return r, ts, err
		}
	}

	// Log puzzle.json loads
	if path == "puzzle.json" {
		mh.State.LogEvent("load", mh.teamID, cat, points)
	}

	return
}

// CheckAnswer returns an error if answer is not a correct answer for puzzle points in category cat
func (mh *MothRequestHandler) CheckAnswer(cat string, points int, answer string) error {
	correct := false
	for _, provider := range mh.PuzzleProviders {
		if ok, err := provider.CheckAnswer(cat, points, answer); err != nil {
			return err
		} else if ok {
			correct = true
		}
	}
	if !correct {
		mh.State.LogEvent("wrong", mh.teamID, cat, points)
		return fmt.Errorf("incorrect answer")
	}

	mh.State.LogEvent("correct", mh.teamID, cat, points)

	if _, err := mh.State.TeamName(mh.teamID); err != nil {
		return fmt.Errorf("invalid team ID")
	}
	if err := mh.State.AwardPoints(mh.teamID, cat, points); err != nil {
		return fmt.Errorf("error awarding points: %s", err)
	}

	return nil
}

// ThemeOpen opens a file from a theme.
func (mh *MothRequestHandler) ThemeOpen(path string) (ReadSeekCloser, time.Time, error) {
	return mh.Theme.Open(path)
}

// Register associates a team name with a team ID.
func (mh *MothRequestHandler) Register(teamName string) error {
	if teamName == "" {
		return fmt.Errorf("empty team name")
	}
	mh.State.LogEvent("register", mh.teamID, "", 0)
	return mh.State.SetTeamName(mh.teamID, teamName)
}

// ExportState anonymizes team IDs and returns StateExport.
// If a teamID has been specified for this MothRequestHandler,
// the anonymized team name for this teamID has the special value "self".
// If not, the puzzles list is empty.
func (mh *MothRequestHandler) ExportState() *StateExport {
	return mh.exportStateIfRegistered(false)
}

// Export state, replacing the team ID with "self" if the team is registered.
//
// If forceRegistered is true, go ahead and export it anyway
func (mh *MothRequestHandler) exportStateIfRegistered(forceRegistered bool) *StateExport {
	export := StateExport{}
	export.Config = mh.Config

	teamName, err := mh.State.TeamName(mh.teamID)
	registered := forceRegistered || mh.Config.Devel || (err == nil)

	export.Messages = mh.State.Messages()
	export.TeamNames = make(map[string]string)

	// Anonymize team IDs in points log, and write out team names
	pointsLog := mh.State.PointsLog()
	exportIDs := make(map[string]string)
	maxSolved := make(map[string]int)
	export.PointsLog = make(award.List, len(pointsLog))

	if registered {
		export.TeamNames["self"] = teamName
		exportIDs[mh.teamID] = "self"
	}
	for logno, awd := range pointsLog {
		if id, ok := exportIDs[awd.TeamID]; ok {
			awd.TeamID = id
		} else {
			exportID := strconv.Itoa(logno)
			name, _ := mh.State.TeamName(awd.TeamID)
			exportIDs[awd.TeamID] = exportID
			awd.TeamID = exportID
			export.TeamNames[exportID] = name
		}
		export.PointsLog[logno] = awd

		// Record the highest-value unlocked puzzle in each category
		if awd.Points > maxSolved[awd.Category] {
			maxSolved[awd.Category] = awd.Points
		}
	}

	export.Puzzles = make(map[string][]int)
	if registered {
		// We used to hand this out to everyone,
		// but then we got a bad reputation on some secretive blacklist,
		// and now the Navy can't register for events.
		for _, provider := range mh.PuzzleProviders {
			for _, category := range provider.Inventory() {
				// Append sentry (end of puzzles)
				allPuzzles := append(category.Puzzles, 0)

				max := maxSolved[category.Name]

				puzzles := make([]int, 0, len(allPuzzles))
				for i, val := range allPuzzles {
					puzzles = allPuzzles[:i+1]
					if !mh.Config.Devel && (val > max) {
						break
					}
				}
				export.Puzzles[category.Name] = puzzles
			}
		}
	}

	return &export
}

// Mothball generates a mothball for the given category.
func (mh *MothRequestHandler) Mothball(cat string, w io.Writer) error {
	var err error

	if !mh.Config.Devel {
		return fmt.Errorf("cannot mothball in production mode")
	}
	for _, provider := range mh.PuzzleProviders {
		if err = provider.Mothball(cat, w); err == nil {
			return nil
		}
	}
	return err
}
