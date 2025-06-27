package main

import (
	"errors"
	"testing"
)

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

	if state.Enabled(state.clock.Now()) {
		t.Error("New instance should default disabled")
	}
	if err := state.configFS.WriteFile("schedule.txt", []byte("enable -\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if errs := state.FSState.Refresh(state.clock.Now()); errs != nil {
		t.Error(errs)
	}

	if !state.Enabled(state.clock.Now()) {
		t.Error("Event should be enabled")
	}
	if err := state.configFS.Remove("schedule.txt"); err != nil {
		t.Error(err)
	}
	if errs := state.FSState.Refresh(state.clock.Now()); errs == nil {
		t.Error("Removing schedule should trigger a refresh error")
	}
	if !state.Enabled(state.clock.Now()) {
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
	categories := cat.Categories(store.clock.Now())
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
