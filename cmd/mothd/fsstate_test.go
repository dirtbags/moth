package main

import (
	"errors"
	"testing"

	"github.com/dirtbags/moth/v5/pkg/wfs"
)

type TestFSState struct {
	*FSState
	*TestClock
	*testing.T
}

func NewTestFSState(t *testing.T) *TestFSState {
	fsys := wfs.DirFS(t.TempDir())
	state, err := NewFSState(fsys, "test-consolidate")
	if err != nil {
		t.Fatal(err)
	}
	s := &TestFSState{
		FSState:   state,
		TestClock: NewTestClock(),
		T:         t,
	}
	s.Refresh()
	return s
}

func (s *TestFSState) Refresh() {
	s.T.Helper()
	errs := s.FSState.Refresh(s.Now())
	for _, err := range errs {
		if err != nil {
			s.T.Error(err)
		}
	}
}

func TestFSStateBasic(t *testing.T) {
	state := NewTestFSState(t)
	if err := state.configFS.WriteFile("teamids.txt", []byte("*\n"), 0644); err != nil {
		t.Fatal(err)
	}
	state.Refresh()

	state.SetTeamData("team1", "bloop")
	state.Refresh()

	if data, ok := state.TeamData("team1"); !ok {
		t.Error("Didn't find team1")
	} else if data != "bloop" {
		t.Error("Wrong team data")
	}

	if state.Enabled(state.Now()) {
		t.Error("New instance should default disabled")
	}
	if err := state.configFS.WriteFile("schedule.txt", []byte("enable -\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if errs := state.FSState.Refresh(state.Now()); errs != nil {
		t.Error(errs)
	}

	if !state.Enabled(state.Now()) {
		t.Error("Event should be enabled")
	}
	if err := state.configFS.Remove("schedule.txt"); err != nil {
		t.Error(err)
	}
	if errs := state.FSState.Refresh(state.Now()); errs == nil {
		t.Error("Removing schedule should trigger a refresh error")
	}
	if !state.Enabled(state.Now()) {
		t.Error("Removing shedule.txt should fall back to previous schedule")
	}
}

func TestFSStateCompact(t *testing.T) {
	state := NewTestFSState(t)

	al := NewAwardList()

	state.awardFS.WriteFile("moo", []byte("100,team,cat,puzzle1"), 0644)
	if awardFilenames, readErrs := state.ReadRecords(al, state.awardFS); readErrs != nil {
		t.Error(errors.Join(readErrs...))
	} else if err := state.Compact(state.awardFS, al, awardFilenames); err != nil {
		t.Error(err)
	}

	state.awardFS.WriteFile("moo2", []byte("101,team,cat,puzzle2"), 0644)
	if awardFilenames, readErrs := state.ReadRecords(al, state.awardFS); readErrs != nil {
		t.Error(errors.Join(readErrs...))
	} else if err := state.Compact(state.awardFS, al, awardFilenames); err != nil {
		t.Error(err)
	}

	state.awardFS.WriteFile("moo3", []byte("102,team,cat,puzzle3"), 0644)
	if awardFilenames, readErrs := state.ReadRecords(al, state.awardFS); readErrs != nil {
		t.Error(errors.Join(readErrs...))
	} else if err := state.Compact(state.awardFS, al, awardFilenames); err != nil {
		t.Error(err)
	} else if len(awardFilenames) != 2 {
		t.Error("Unexpected length of compacted state directory:", len(awardFilenames))
	}
	if al.Len() != 3 {
		t.Error("Wrong length:", al.Len())
	}

	for i := 0; i < 6; i++ {
		if awardFilenames, readErrs := state.ReadRecords(al, state.awardFS); readErrs != nil {
			t.Error(errors.Join(readErrs...))
		} else if err := state.Compact(state.awardFS, al, awardFilenames); err != nil {
			t.Error(err)
		} else if len(awardFilenames) != 1 {
			t.Error("Unexpected length of compacted state directory:", len(awardFilenames))
		}
		if al.Len() != 3 {
			t.Error("Wrong length:", al.Len())
		}
	}

}

type TestFSStateStore struct {
	*FSStateStore
	*TestClock
	*testing.T
}

// NewTestFSStateStore returns a TestFSStateStore with a "test" instance.
func NewTestFSStateStore(t *testing.T) *TestFSStateStore {
	fsys := wfs.DirFS(t.TempDir())
	store := &TestFSStateStore{
		FSStateStore: NewFSStateStore(fsys, "testhost", 10, 1),
		TestClock:    NewTestClock(),
		T:            t,
	}
	store.MkInstance("test")
	return store
}

func (store *TestFSStateStore) Refresh() {
	store.T.Helper()
	if err := store.FSStateStore.Refresh(store.Now()); err != nil {
		store.T.Error(err)
	}
}

func (store *TestFSStateStore) MkInstance(name string) {
	if err := store.fsys.Mkdir(name, 0755); err != nil {
		store.T.Fatal(err)
	}
	store.Refresh()
	if err := store.fsys.WriteFile(name+"/schedule.txt", []byte("enable -\nrollout - cat0\n"), 0644); err != nil {
		store.T.Fatal(err)
	}
	if err := store.fsys.WriteFile(name+"/teams/test-team", []byte("test team"), 0644); err != nil {
		store.T.Fatal(err)
	}
	store.Refresh()
}

func TestFSStateStoreBasic(t *testing.T) {
	store := NewTestFSStateStore(t)

	store.MkInstance("new")

	if _, found := store.State("undefined"); found {
		t.Error("non-existent state lookup didn't fail")
	}
	if _, found := store.State("new"); !found {
		t.Error("didn't find new state")
	}

	store.MkInstance("fred")
	if _, found := store.State("fred"); !found {
		t.Error("didn't find fred! where is fred?!")
	}

	clock := NewTestClock()
	cat, found := store.State("test")
	if !found {
		t.Error("didn't find test state")
	}
	if !cat.Enabled(clock.Now()) {
		t.Error("not enabled")
	}

	categories := cat.Categories(clock.Now())
	if len(categories) != 1 {
		t.Error("wrong number of categories:", categories)
	}
}

func TestFSStateStoreMulitpleRollout(t *testing.T) {
	store := NewTestFSStateStore(t)
	if err := store.fsys.WriteFile("test/schedule.txt", []byte("enable -\nrollout - cat0 pategory\n"), 0644); err != nil {
		t.Fatal(err)
	}
	store.Refresh()

	cat, found := store.State("test")
	if !found {
		t.Error("didn't find test state")
	}
	categories := cat.Categories(store.Now())
	if len(categories) != 2 {
		t.Error("wrong number of categories")
	}
}

func TestNilList(t *testing.T) {
	var l []int

	if len(l) != 0 {
		t.Fatal("uninitialized list had length", len(l))
	}
}
