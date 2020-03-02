package main

import (
	"io"
	"time"
	"fmt"
	"strconv"
)

type Category struct {
	Name    string
	Puzzles []int
}

type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

type StateExport struct {
	Config struct {
		Devel bool
	}
	Messages  string
	TeamNames map[string]string
	PointsLog []Award
	Puzzles map[string][]int
}

type PuzzleProvider interface {
	Open(cat string, points int, path string) (ReadSeekCloser, time.Time, error)
	Inventory() []Category
	CheckAnswer(cat string, points int, answer string) error
	Component
}

type ThemeProvider interface {
	Open(path string) (ReadSeekCloser, time.Time, error)
	Component
}

type StateProvider interface {
	Messages() string
	PointsLog() []*Award
	TeamName(teamId string) (string, error)
	SetTeamName(teamId, teamName string) error
	AwardPoints(teamId string, cat string, points int) error
	Component
}


type Component interface {
	Update()
}


type MothServer struct {
	Puzzles PuzzleProvider
	Theme ThemeProvider
	State StateProvider
}

func NewMothServer(puzzles PuzzleProvider, theme ThemeProvider, state StateProvider) *MothServer {
	return &MothServer{
		Puzzles: puzzles,
		Theme: theme,
		State: state,
	}
}

func (s *MothServer) NewHandler(teamId string) MothRequestHandler {
	return MothRequestHandler{
		MothServer: s,
		teamId: teamId,
	}
}

// XXX: Come up with a better name for this.
type MothRequestHandler struct {
	*MothServer
	teamId string
}

func (mh *MothRequestHandler) PuzzlesOpen(cat string, points int, path string) (ReadSeekCloser, time.Time, error) {
	export := mh.ExportAllState()
	fmt.Println(export.Puzzles)
	for _, p := range export.Puzzles[cat] {
		fmt.Println(points, p)
		if p == points {
			return mh.Puzzles.Open(cat, points, path)
		}
	}
	
	return nil, time.Time{}, fmt.Errorf("Puzzle locked")
}

func (mh *MothRequestHandler) ThemeOpen(path string) (ReadSeekCloser, time.Time, error) {
	return mh.Theme.Open(path)
}

func (mh *MothRequestHandler) Register(teamName string) error {
	// XXX: Should we just return success if the team is already registered?
	// XXX: Should this function be renamed to Login?
	if teamName == "" {
		return fmt.Errorf("Empty team name")
	}
	return mh.State.SetTeamName(mh.teamId, teamName)
}

func (mh *MothRequestHandler) CheckAnswer(cat string, points int, answer string) error {
	if err := mh.Puzzles.CheckAnswer(cat, points, answer); err != nil {
		return err
	}
	
	if err := mh.State.AwardPoints(mh.teamId, cat, points); err != nil {
		return err
	}
	
	return nil
}

func (mh *MothRequestHandler) ExportAllState() *StateExport {
	export := StateExport{}

	teamName, _ := mh.State.TeamName(mh.teamId)
	
	export.Messages = mh.State.Messages()
	export.TeamNames = map[string]string{"self": teamName}

	// Anonymize team IDs in points log, and write out team names
	pointsLog := mh.State.PointsLog()
	exportIds := map[string]string{mh.teamId: "self"}
	maxSolved := map[string]int{}
	export.PointsLog = make([]Award, len(pointsLog))
	for logno, award := range pointsLog {
		exportAward := *award
		if id, ok := exportIds[award.TeamId]; ok {
			exportAward.TeamId = id
		} else {
			exportId := strconv.Itoa(logno)
			name, _ := mh.State.TeamName(award.TeamId)
			exportAward.TeamId = exportId
			exportIds[award.TeamId] = exportAward.TeamId
			export.TeamNames[exportId] = name
		}
		export.PointsLog[logno] = exportAward
		
		// Record the highest-value unlocked puzzle in each category
		if award.Points > maxSolved[award.Category] {
			maxSolved[award.Category] = award.Points
		}
	}


	export.Puzzles = make(map[string][]int)
	for _, category := range mh.Puzzles.Inventory() {
		// Append sentry (end of puzzles)
		allPuzzles := append(category.Puzzles, 0)

		max := maxSolved[category.Name]

		puzzles := make([]int, 0, len(allPuzzles))
		for i, val := range allPuzzles {
			puzzles = allPuzzles[:i+1]
			if val > max {
				break
			}
		}
		export.Puzzles[category.Name] = puzzles
	}

	return &export
}

func (mh *MothRequestHandler) ExportState() *StateExport {
	export := mh.ExportAllState()
	
	// We don't give this out to just anybody,
	// because back when we did,
	// we got a bad reputation on some secretive blacklist,
	// and now the Navy can't register for events.
	if export.TeamNames["self"] == "" {
		export.Puzzles = map[string][]int{}
	}
	
	return export
}
