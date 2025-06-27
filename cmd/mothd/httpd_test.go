package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"testing/fstest"
)

func TestHttpdStatic(t *testing.T) {
	h := NewTestContext(t, "/")

	if resp := h.TestRequestAnonymous("GET", "http://example.com/", nil, 200); resp.Body.String() != "mothd\n" {
		t.Error("Wrong body:", resp)
	}

	if resp := h.TestRequestAnonymous("GET", "http://example.com/index.html", nil, 301); false {
	} else if loc, err := resp.Result().Location(); err != nil {
		t.Error(err)
	} else if loc.String() != "./" {
		t.Error("Expected redirect:", resp)
	}

	h.TestRequestAnonymous("GET", "http://example.com/notfound.html", nil, 404)
}

func TestHttpdStaticBase(t *testing.T) {
	h := NewTestContext(t, "/base")
	h.TestRequestAnonymous("GET", "http://example.com/", nil, 404)
	h.TestRequestAnonymous("GET", "http://example.com/base/", nil, 200)
}

func TestHttpdRegistration(t *testing.T) {
	h := NewTestContext(t, "/")

	teamIDs := []TeamID{
		TeamID("team0"),
		TeamID("team1"),
	}

	// If you change something about current.json,
	// replace wanted's value with the reported content from the error message
	// (after verifying it's correct, of course)
	wanted := "{\"MothVersion\":\"[devel]\",\"Running\":true,\"OpenRegistration\":false,\"ClientConfig\":{},\"Errors\":null,\"Hours\":[{\"start\":\"0001-01-01T00:00:00Z\",\"end\":\"9999-12-31T23:59:59Z\"}],\"TeamData\":{\"0\":\"test-team\"},\"Awards\":null,\"Puzzles\":null}\n"
	if resp := h.TestRequestAnonymous("GET", "http://example.com/test/state/current.json", nil, 200); resp.Body.String() != wanted {
		t.Errorf("Unexpected state\n   got: %#v\nwanted: %#v\n", resp.Body.String(), wanted)
	}

	// Team doesn't exist yet
	h.TestRequestAuth("PUT", "http://example.com/test/state/teamdata", teamIDs[0].Bytes(), teamIDs[0].String(), "", 401)

	// Write team IDs to teamids.txt
	f := new(bytes.Buffer)
	fmt.Fprintln(f, teamIDs[0])
	fmt.Fprintln(f, teamIDs[1])
	if err := h.WriteStateFile("test/teamids.txt", f.Bytes()); err != nil {
		t.Fatal(err)
	}
	h.Refresh()

	stateTest := func(method string, target string, data []byte, team int, code int) {
		t.Helper()
		teamID := teamIDs[team]
		resp := h.TestRequestAuth(method, target, data, "", teamID.String(), code)
		var obj StateExport
		if err := json.Unmarshal(resp.Body.Bytes(), &obj); err != nil {
			t.Error(err)
		}
		if obj.TeamData["self"] != string(data) {
			t.Error("Team data wrong", resp.Body.String())
		}
	}

	// Try to register a team not listed
	h.TestRequestAuth("PUT", "http://example.com/test/state/teamdata", []byte("team1500"), "team1500", "", 401)

	// Try a couple wrong registrations, then register team 0
	h.TestRequestAuth("PUT", "http://example.com/test/state/teamdata", nil, "", teamIDs[0].String(), http.StatusLengthRequired)
	h.TestRequestAuth("PUT", "http://example.com/test/state/teamdata", make([]byte, 5000), teamIDs[0].String(), "", http.StatusRequestEntityTooLarge)
	stateTest("PUT", "http://example.com/test/state/teamdata", teamIDs[0].Bytes(), 0, 200)
	h.Refresh()
	stateTest("PUT", "http://example.com/test/state/teamdata", teamIDs[0].Bytes(), 0, 409)

	// Register team 1
	h.TestRequestAuth("PUT", "http://example.com/test/state/teamdata", teamIDs[0].Bytes(), "", teamIDs[1].String(), 200)
	stateTest("PUT", "http://example.com/test/state/teamdata", teamIDs[1].Bytes(), 1, 200)
	h.Refresh()
	stateTest("PUT", "http://example.com/test/state/teamdata", teamIDs[1].Bytes(), 1, 409)
}

func TestHttpdUnlocked(t *testing.T) {
	h := NewTestContext(t, "/")

	state, found := h.Context.State("test")
	if !found {
		t.Fatal("not found")
	}
	env := &Environment{
		State: state,
		ctx:   h.Context,
	}
	categories := state.Categories(h.Now())
	if len(categories) == 0 {
		t.Fatal("no categories in this test requiring them")
	}
	for _, catid := range state.Categories(h.Now()) {
		cat, err := h.Category(catid)
		if err != nil {
			t.Error(err)
			continue
		}
		puzzles := env.Unlocked(catid, cat)
		if len(puzzles) != 1 {
			t.Error("Wrong length puzzles", puzzles)
			t.Log(state.Hours(), h.Now())
			t.Log(state.Categories(h.Now()))
		}
	}
}

func TestHttpdPuzzles(t *testing.T) {
	h := NewTestContext(t, "/")

	if _, err := h.Category("cat0"); err != nil {
		t.Fatal(err)
	}
	h.Refresh()
	if _, err := h.Category("cat0"); err != nil {
		t.Fatal(err)
	}

	h.TestRequestAnonymous("GET", "http://example.com/test/puzzles/cat0/index.html", nil, 401)
	h.TestRequestAnonymous("GET", "http://example.com/test/puzzles/cat0/1/not-here", nil, 401)
	h.TestRequestAnonymous("GET", "http://example.com/test/puzzles/cat0/1/index.html", nil, 401)

	h.TestRequest("GET", "http://example.com/test/puzzles/cat0/", nil, 200)
	h.TestRequest("GET", "http://example.com/test/puzzles/cat0/index.html", nil, 301)
	h.TestRequest("GET", "http://example.com/test/puzzles/cat0/1/", nil, 200)
	h.TestRequest("GET", "http://example.com/test/puzzles/cat0/1/index.html", nil, 301)
	h.TestRequest("GET", "http://example.com/test/puzzles/cat0/1/not-here-but-authenticated", nil, 404)
	h.TestRequest("GET", "http://example.com/test/puzzles/cat0/two/index.html", nil, 404)
	h.TestRequest("GET", "http://example.com/test/puzzles/cat0/two/moo.txt", nil, 404)
	if resp := h.TestRequest("GET", "http://example.com/test/puzzles/cat0/1/moo.txt", nil, 200); resp.Body.String() != "moo.\n" {
		t.Error("Unexpected body", resp.Body.String())
	}

	h.TestRequest("GET", "http://example.com/test/puzzles/cat0/__MOTH/answers.txt", nil, 404)
	h.TestRequest("GET", "http://example.com/test/puzzles/cat0/1/__MOTH/answers.txt", nil, 404)

	h.TestRequest("PUT", "http://example.com/test/puzzles/cat0/19324/answer", []byte("wrong"), 404)
	h.TestRequest("PUT", "http://example.com/test/puzzles/cat0/1/answer", []byte("wrong"), 409)
	h.TestRequest("PUT", "http://example.com/test/puzzles/cat0/1/answer", make([]byte, 5000), 413)
	h.TestRequest("PUT", "http://example.com/test/puzzles/cat0/1/answer", []byte("answer123"), 200)
	h.TestRequest("PUT", "http://example.com/test/puzzles/cat0/1/answer", []byte("answer123"), 200)
	h.Refresh()
	h.TestRequest("PUT", "http://example.com/test/puzzles/cat0/1/answer", []byte("answer123"), 202)
	h.Refresh()
	h.TestRequest("GET", "http://example.com/test/puzzles/cat0/two/", nil, 200)
	h.TestRequest("GET", "http://example.com/test/puzzles/cat0/two/index.html", nil, 301)
	h.TestRequest("GET", "http://example.com/test/puzzles/cat0/two/moo.txt", nil, 404)

	state := StateExport{}
	if resp := h.TestRequest("GET", "http://example.com/test/state/current.json", nil, 200); len(resp.Body.String()) == 0 {
		t.Error("Short state", resp)
	} else if err := json.Unmarshal(resp.Body.Bytes(), &state); err != nil {
		t.Error(err)
	} else if state.Awards.Len() != 1 {
		t.Errorf("Points log wrong length. Wanted 1, got %v (length %d)", state.Awards, state.Awards.Len())
	} else if len(state.Puzzles["cat0"]) != 2 {
		t.Error("Didn't unlock next puzzle")
		t.Log(state.Puzzles)
	}

}

func TestErrors(t *testing.T) {
	s := &TestMothballStore{
		NewMothballStore(
			fstest.MapFS{
				"cat0.mb": &fstest.MapFile{
					Data:    []byte{},
					Mode:    0644,
					ModTime: TimeZero,
					Sys:     nil,
				},
			},
		),
	}

	h := NewTestContext(t, "/")
	h.Context.CategoryStore = s

	if err := h.Context.Refresh(h.Now()); err == nil {
		t.Fatal("Expecting error, none found")
	}

	if resp := h.TestRequest("GET", "http://example.com/test/state/current.json", nil, 200); !strings.Contains(resp.Body.String(), "not a valid zip file") {
		t.Error("Wrong body:", resp.Body.String())
	}
}

func TestDisable(t *testing.T) {
	h := NewTestContext(t, "/")

	if err := h.WriteStateFile("test/schedule.txt", []byte("rollout - cat0\n")); err != nil {
		t.Fatal(err)
	}
	h.Refresh()

	h.TestRequest("PUT", "http://example.com/test/puzzles/cat0/1/answer", []byte("answer123"), 503)

	if err := h.WriteStateFile("test/schedule.txt", []byte("enable -\nrollout - cat0\n")); err != nil {
		t.Fatal(err)
	}
	h.Refresh()

	h.TestRequest("PUT", "http://example.com/test/puzzles/cat0/1/answer", []byte("answer123"), 200)
}
