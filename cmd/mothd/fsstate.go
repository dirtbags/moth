package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dirtbags/moth/v5/pkg/mdir"
	"github.com/dirtbags/moth/v5/pkg/period"
	"github.com/dirtbags/moth/v5/pkg/wfs"
)

// DistinguishableChars are visually unambiguous glyphs.
// People with mediocre handwriting could write these down unambiguously,
// and they can be entered without holding down shift.
const DistinguishableChars = "34678abcdefhikmnpqrtwxy="

const teamidsName = "teamids.txt"

const initializedName = "initialized"
const initializedContents = `initialized: remove to re-initialize the contest.

Initialization will clear awards and team registration,
but will preserve any files in this directory.
`

const readmeName = "README.md"
const readmeContents = `# State Directory Layout

initialized
: Delete this to reset registered teams and delete all awards.

schedule.txt
: Schedule of when event will run, and when categories are released.

teamids.txt
: List of team IDs accepted for registration.

client/
: Client configuration files: refer to your client's documentation.

teams/
: Team "public" data

events/
: Event log

awards/
: Award log
`

const scheduleName = "schedule.txt"
const initialSchedule = `# schedule.txt: when things happen.
#
# Rename or remove this file to quickly disable an instance.
#
# Times and periods are RFC3339 format.
# "-" means "forever".
# You can get RFC3339 format with the 'date' command:
#
#    $ TZ=US/Central date -d 'oct 28 08:30' -Is
#
# Examples:
#
# enable -
# enable 2024-07-25T19:35:14-06:00/2024-07-26T01:35:14-06:00
# enable 2024-07-25T19:35:14-06:00/PT6H
# enable PT6H/2024-07-26T01:35:14-06:00
# rollout - intro.mb easycat1.mb easycat2.mb
# rollout 2024-07-25T21:00:00-0600 hardcat1.mb hardcat2.mb
`

// Rollout describes when a category should roll out.
type Rollout struct {
	Category string
	When     time.Time
}

type FSState struct {
	configFS *mdir.MDir
	clientFS *mdir.MDir
	teamFS   *mdir.MDir
	awardFS  *mdir.MDir
	eventFS  *mdir.MDir

	// openRegistration, obtained from teamids.txt
	openRegistration atomic.Bool

	// hours this event is running, obtained from schedule.txt
	hours atomic.Pointer[[]period.Period]

	// rollout schedule, obtained from schedule.txt
	rollout atomic.Pointer[[]Rollout]

	// teamids considered valid, obtained from teamids.txt
	teamids atomic.Pointer[[]TeamID]

	// teamdata for every team, obtained from teamFS
	teamdata atomic.Pointer[map[TeamID]string]

	// config displayed in UI
	config atomic.Pointer[map[string]string]

	// errors displayed in UI
	errors atomic.Pointer[[]error]

	// awards listed in awardFS
	awards atomic.Pointer[AwardList]

	// compactBase is the filename to use for compaction.
	compactBase string
}

func NewFSState(fsys wfs.WriteFS, compactBase string) (*FSState, error) {
	configFS, err := mdir.New(fsys, ".tmp", ".")
	if err != nil {
		return nil, err
	}
	clientFS, err := mdir.New(fsys, ".client.tmp", "client")
	if err != nil {
		return nil, err
	}
	eventFS, err := mdir.New(fsys, ".log.tmp", "log")
	if err != nil {
		return nil, err
	}
	teamFS, err := mdir.New(fsys, ".teams.tmp", "teams")
	if err != nil {
		return nil, err
	}
	awardFS, err := mdir.New(fsys, ".awards.tmp", "awards")
	if err != nil {
		return nil, err
	}
	teamdata := make(map[TeamID]string)

	s := &FSState{
		configFS:    configFS,
		clientFS:    clientFS,
		eventFS:     eventFS,
		teamFS:      teamFS,
		awardFS:     awardFS,
		compactBase: compactBase,
	}
	s.teamdata.Store(&teamdata)
	return s, nil
}

func (s *FSState) Enabled(when time.Time) bool {
	enabled := false
	for _, p := range s.Hours() {
		if p.Contains(when) {
			enabled = true
		}
	}
	return enabled
}

func (s *FSState) OpenRegistration() bool {
	return s.openRegistration.Load()
}

func (s *FSState) Hours() []period.Period {
	return *s.hours.Load()
}

func (s *FSState) ClientConfig() map[string]string {
	return *s.config.Load()
}

func (s *FSState) Errors() []string {
	errors := s.errors.Load()
	if errors == nil {
		return nil
	}

	var strs []string
	for _, err := range *errors {
		if err == nil {
			continue
		}
		strs = append(strs, err.Error())
	}
	return strs
}

func (s *FSState) Categories(when time.Time) []CategoryID {
	var categories []CategoryID
	for _, r := range *s.rollout.Load() {
		if r.When.Compare(when) <= 0 {
			cat := CategoryID(r.Category)
			categories = append(categories, cat)
		}
	}
	return categories
}

func (s *FSState) Teams() (keys []TeamID) {
	teamdata := *s.teamdata.Load()
	for k := range teamdata {
		keys = append(keys, k)
	}
	return
}

func (s *FSState) TeamData(team TeamID) (data string, ok bool) {
	teamdata := *s.teamdata.Load()
	data, ok = teamdata[team]
	return
}

func (s *FSState) SetTeamData(team TeamID, data string) error {
	permitted := s.openRegistration.Load()
	if !permitted {
		teamidsp := s.teamids.Load()
		if teamidsp != nil {
			for _, t := range *teamidsp {
				if team == t {
					permitted = true
					break
				}
			}
		}
	}
	if !permitted {
		return fs.ErrInvalid
	}

	teamdata := *s.teamdata.Load()
	if _, found := teamdata[team]; found {
		return fs.ErrExist
	}

	return s.teamFS.WriteFile(team.Filename(), []byte(data), 0644)
}

// writeItems writes wt to a file.
func writeItems(fsys *mdir.MDir, name string, wt io.WriterTo) error {
	w, err := fsys.Create(name)
	if err != nil {
		return err
	}

	if _, err := wt.WriteTo(w); err != nil {
		w.Close()
		fsys.Remove(name)
		return err
	}

	if err := w.Close(); err != nil {
		fsys.Remove(name)
		return err
	}

	return nil
}

func (s *FSState) RecordAward(a Award) error {
	// We don't cache the new award,
	// to ensure that all instances have similar cache validity.
	return writeItems(s.awardFS, a.Filename(), NewAwardList(a))
}

func (s *FSState) Awards() *AwardList {
	return s.awards.Load()
}

func (s *FSState) Log(e Event) error {
	return writeItems(s.eventFS, e.Filename(), NewEventList(e))
}

// readlines returns a list of lines from a file, with empty lines removed
func (s *FSState) readlines(name string) (lines []string, err error) {
	r, err := s.configFS.Open(name)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}

	return lines, err
}

func (s *FSState) parseSchedule(lines []string) ([]period.Period, []Rollout, error) {
	var hours []period.Period
	var rollout []Rollout

	for lineno, line := range lines {
		fields := strings.Fields(line)

		// skip blank lines and comments
		switch {
		case len(line) == 0:
			continue
		case line[0] == '#':
			continue
		case len(fields) == 0:
			continue
		case len(fields) < 2:
			return nil, nil, fmt.Errorf("schedule line %d: parse error", lineno)
		}

		what := fields[0]
		when := fields[1]

		switch what {
		case "enable":
			var p period.Period

			if when == "-" {
				p = period.Forever
			} else {
				if pp, err := period.Parse(when); err != nil {
					return nil, nil, err
				} else {
					p = pp

				}
			}
			hours = append(hours, p)

		case "rollout":
			var t time.Time

			if when == "-" {
				t = time.Time{}
			} else {
				if tp, err := time.Parse(time.RFC3339, when); err != nil {
					return nil, nil, err
				} else {
					t = tp
				}
			}

			for _, cat := range fields[2:] {
				if cat == "" {
					continue
				}
				r := Rollout{
					When:     t,
					Category: cat,
				}
				rollout = append(rollout, r)
			}
		default:
			return nil, nil, fmt.Errorf("unparseable schedule line: %#q", line)

		}
	}

	return hours, rollout, nil
}

// ReadRecords runs r.ReadFrom for every file in fsys.
func (s *FSState) ReadRecords(r io.ReaderFrom, fsys fs.FS) (filenames []string, errs []error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, append(errs, err)
	}

	for _, entry := range entries {
		name := entry.Name()

		f, err := fsys.Open(name)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		_, readErr := r.ReadFrom(f)
		closeErr := f.Close()
		if readErr != nil {
			errs = append(errs, readErr)
			continue
		}
		if closeErr != nil {
			errs = append(errs, closeErr)
		}
		filenames = append(filenames, name)
	}

	return filenames, errs
}

// Compact records into a single file.
//
// If len(filenames) is <= 1, do nothing.
// Otherwise, write all records from w into a new file named s.compactBase,
// then remove all filenames.
func (s *FSState) Compact(fsys *mdir.MDir, w io.WriterTo, filenames []string) error {
	if len(filenames) <= 1 {
		return nil
	}

	if err := writeItems(fsys, s.compactBase, w); err != nil {
		return err
	}

	var errs []error
	for _, name := range filenames {
		if name == s.compactBase {
			continue
		}
		err := fsys.Remove(name)
		switch {
		case err == nil:
		// pass
		case errors.Is(err, fs.ErrNotExist): // not a problem
		default:
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Refresh this state.
//
// The following errors can occur here:
// - Unable to tell if `initialized` file exists
// - Error parsing `hours` file
func (s *FSState) Refresh(now time.Time) []error {
	var errs []error

	if _, err := s.configFS.Stat(initializedName); errors.Is(err, fs.ErrNotExist) {
		if err1 := s.initialize(); err1 != nil {
			errs = append(errs, err1)
		}
	} else if err != nil {
		errs = append(errs, err)
	}

	if lines, err := s.readlines(teamidsName); err != nil {
		errs = append(errs, err)
	} else if (len(lines) == 1) && (lines[0] == "*") {
		s.openRegistration.Store(true)
		s.teamids.Store(nil)
	} else {
		var teamids []TeamID

		s.openRegistration.Store(false)
		for _, line := range lines {
			if line == "" {
				continue
			}
			teamids = append(teamids, TeamID(line))
		}
		s.teamids.Store(&teamids)
	}

	if lines, err := s.readlines(scheduleName); err != nil {
		errs = append(errs, err)
	} else {
		hours, rollout, err := s.parseSchedule(lines)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %v", scheduleName, err))
		}
		s.hours.Store(&hours)
		s.rollout.Store(&rollout)
	}

	config := make(map[string]string)
	configents, err := fs.ReadDir(s.clientFS, ".")
	if err != nil {
		errs = append(errs, err)
	} else {
		for _, ent := range configents {
			name := ent.Name()
			contents, err := fs.ReadFile(s.clientFS, name)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			config[name] = string(contents)
		}
		s.config.Store(&config)
	}

	teams := make(map[TeamID]string)
	dents, err := fs.ReadDir(s.teamFS, ".")
	if err != nil {
		errs = append(errs, err)
	}
	for _, dent := range dents {
		name := dent.Name()
		if data, err := fs.ReadFile(s.teamFS, name); err != nil {
			errs = append(errs, err)
		} else {
			teams[TeamID(name)] = string(data)
		}
	}
	s.teamdata.Store(&teams)

	awards := NewAwardList()
	if awardFilenames, readErrs := s.ReadRecords(awards, s.awardFS); readErrs != nil {
		errs = append(errs, readErrs...)
	} else {
		s.awards.Store(awards)
		errs = append(errs, s.Compact(s.awardFS, awards, awardFilenames))
	}

	events := NewEventList()
	if eventFilenames, readErrs := s.ReadRecords(events, s.eventFS); readErrs != nil {
		errs = append(errs, readErrs...)
	} else {
		errs = append(errs, s.Compact(s.eventFS, events, eventFilenames))
	}

	var finalErrs []error
	for _, err := range errs {
		if err == nil {
			continue
		}
		finalErrs = append(finalErrs, err)
	}
	if len(finalErrs) == 0 {
		s.configFS.RemoveAll("errors.txt")
	} else {
		if f, err := s.configFS.Create("errors.txt"); err != nil {
			finalErrs = append(finalErrs, err)
		} else {
			defer f.Close()
			fmt.Fprintln(f, "## errors.txt - errors in this instance")
			fmt.Fprintln(f, "##")
			fmt.Fprintln(f, "## This file is written by the server when errors are encountered.")
			fmt.Fprintln(f)
			for _, err := range finalErrs {
				fmt.Fprintln(f, err.Error())
			}
		}
	}

	s.errors.Store(&finalErrs)
	return finalErrs
}

// writeConfigFileExcl writes a file to configFS iff it doesn't exist
func (s *FSState) writeConfigFileExcl(name string, contents string) error {
	w, err := s.configFS.OpenFile(
		name,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0644,
	)
	switch {
	case errors.Is(err, os.ErrExist):
		// Leave the existing file alone
		return nil
	case err != nil:
		return err
	}

	_, err = w.Write([]byte(contents))
	if err1 := w.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}

func (s *FSState) initialize() error {
	if err := s.teamFS.Reset(); err != nil {
		return err
	}
	if err := s.awardFS.Reset(); err != nil {
		return err
	}
	if err := s.eventFS.Reset(); err != nil {
		return err
	}

	teamids := new(strings.Builder)
	id := make([]byte, 8)
	for i := 0; i < 100; i++ {
		for i := range id {
			char := rand.Intn(len(DistinguishableChars))
			id[i] = DistinguishableChars[char]
		}
		teamids.Write(id)
		teamids.WriteRune('\n')
	}
	if err := s.writeConfigFileExcl(teamidsName, teamids.String()); err != nil {
		return err
	}

	if err := s.writeConfigFileExcl(scheduleName, initialSchedule); err != nil {
		return err
	}

	if err := s.writeConfigFileExcl(readmeName, readmeContents); err != nil {
		return err
	}

	if err := s.writeConfigFileExcl(initializedName, initializedContents); err != nil {
		return err
	}

	return nil
}

// FSStateStore provides StateStore for a wfs.WriteFS
type FSStateStore struct {
	fsys        wfs.WriteFS
	states      atomic.Pointer[map[string]*FSState]
	compactBase string
}

func NewFSStateStore(fsys wfs.WriteFS, hostname string, pid int, instance int) *FSStateStore {
	store := &FSStateStore{
		fsys: fsys,
		compactBase: fmt.Sprintf(
			"all-%s-%d-%d",
			hostname,
			pid,
			instance,
		),
	}
	return store
}

func (store *FSStateStore) State(name string) (state State, found bool) {
	statesPtr := store.states.Load()
	if statesPtr == nil {
		return nil, false
	}
	states := *statesPtr

	state, found = states[name]
	return state, found
}

func (store *FSStateStore) Refresh(now time.Time) error {
	var oldstates map[string]*FSState
	if sp := store.states.Load(); sp != nil {
		oldstates = *sp
	}
	newStates := make(map[string]*FSState)

	dents, err := fs.ReadDir(store.fsys, ".")
	if err != nil {
		return err
	}

	for _, dent := range dents {
		if !dent.IsDir() {
			continue
		}

		name := dent.Name()
		if existing, ok := oldstates[name]; ok {
			newStates[name] = existing
			continue
		}

		newfs, err := wfs.Sub(store.fsys, name)
		if err != nil {
			return err
		}

		newstate, err := NewFSState(newfs, store.compactBase)
		if err != nil {
			return err
		}

		newStates[name] = newstate
	}

	var errs []error
	for _, fsstate := range newStates {
		errs = append(errs, fsstate.Refresh(now)...)
	}

	// always store newStates: we don't want a goofy instance to block other changes.
	store.states.Store(&newStates)

	return errors.Join(errs...)
}
