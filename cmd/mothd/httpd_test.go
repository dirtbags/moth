package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

type TestHTTPServer struct {
	*Context
	*TestFSStateStore
	t   *testing.T
	now time.Time
}

// WebRootFS is a minimal filesystem usable as the web root.
var WebRootFS = fstest.MapFS{
	"index.html": &fstest.MapFile{
		Data:    []byte("mothd\n"),
		Mode:    0644,
		ModTime: time.Time{},
		Sys:     nil,
	},
}

// NewTestHttpd returns a TestHTTPServer with a "test" instance.
func NewTestHttpd(t *testing.T, base string) *TestHTTPServer {
	states := NewTestFSStateStore(t)

	ctx := &Context{
		CategoryStore: TestdataCategoryStore(t),
		StateStore:    states.FSStateStore,
		WebFS:         http.FS(WebRootFS),
	}
	ctx.InstallHandlers(base)

	h := &TestHTTPServer{
		Context:          ctx,
		TestFSStateStore: states,
		t:                t,
		now:              time.Date(2024, 8, 15, 9, 0, 0, 0, time.UTC),
	}

	return h
}

func (h *TestHTTPServer) Refresh() {
	h.t.Helper()
	h.now = h.now.Add(1 * time.Second)
	if err := h.Context.Refresh(h.now); err != nil {
		h.t.Fatal(err)
	}
}

func (h *TestHTTPServer) TestServe(t *testing.T, req *http.Request, code int) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Result().StatusCode != code {
		t.Errorf("Expected %d, got %s", code, w.Result().Status)
		t.Logf("     Req: %s %s", req.Method, req.URL.Path)
		t.Logf("  Header: %v", w.Result().Header)
		body, _ := io.ReadAll(w.Result().Body)
		t.Logf("    Body: %#v", string(body))
	}
	return w
}

// TestRequest makes an HTTP request to target (a URL) using method.
// If body is non-nil, it will be sent as the body of the request.
// If the returned HTTP code does not match code,
// an error is logged to t.
func (h *TestHTTPServer) TestRequest(t *testing.T, method string, target string, body []byte, code int) *httptest.ResponseRecorder {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	request := httptest.NewRequest(method, target, bodyReader)
	return h.TestServe(t, request, code)
}

func (h *TestHTTPServer) TestRequestAuth(t *testing.T, method string, target string, body []byte, username string, password string, code int) *httptest.ResponseRecorder {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	request := httptest.NewRequest(method, target, bodyReader)
	request.SetBasicAuth(username, password)
	return h.TestServe(t, request, code)
}

func (hs *TestHTTPServer) Authenticate(username string, password string) *TestHTTPServerAuthenticated {
	return &TestHTTPServerAuthenticated{hs, username, password}
}

type TestHTTPServerAuthenticated struct {
	*TestHTTPServer
	username string
	password string
}

func (hsa *TestHTTPServerAuthenticated) TestRequest(t *testing.T, method string, target string, body []byte, code int) *httptest.ResponseRecorder {
	t.Helper()
	return hsa.TestRequestAuth(t, method, target, body, hsa.username, hsa.password, code)
}

func TestHttpdStatic(t *testing.T) {
	h := NewTestHttpd(t, "/")

	if resp := h.TestRequest(t, "GET", "http://example.com/", nil, 200); resp.Body.String() != "mothd\n" {
		t.Error("Wrong body:", resp)
	}

	if resp := h.TestRequest(t, "GET", "http://example.com/index.html", nil, 301); false {
	} else if loc, err := resp.Result().Location(); err != nil {
		t.Error(err)
	} else if loc.String() != "./" {
		t.Error("Expected redirect:", resp)
	}

	h.TestRequest(t, "GET", "http://example.com/notfound.html", nil, 404)
}

func TestHttpdStaticBase(t *testing.T) {
	h := NewTestHttpd(t, "/base")
	h.TestRequest(t, "GET", "http://example.com/", nil, 404)
	h.TestRequest(t, "GET", "http://example.com/base/", nil, 200)
}

func TestHttpdRegistration(t *testing.T) {
	h := NewTestHttpd(t, "/")
	teamData := []byte("TeamData")
	team2Data := []byte("Team2Data")

	h.TestRequest(t, "GET", "http://example.com/test1/state/current.json", nil, 404)

	h.MkInstance("test1")

	// My technique so far has been to replace this with what's reported,
	// but only if I'm expecting a change.
	// Please leave the test, though, since it will catch unintentional changes.
	wanted := "{\"MothVersion\":\"[devel]\",\"Running\":true,\"OpenRegistration\":false,\"ClientConfig\":{},\"Errors\":null,\"Hours\":[{\"start\":\"0001-01-01T00:00:00Z\",\"end\":\"9999-12-31T23:59:59Z\"}],\"TeamData\":{},\"Awards\":null,\"Puzzles\":null}\n"
	if resp := h.TestRequest(t, "GET", "http://example.com/test1/state/current.json", nil, 200); resp.Body.String() != wanted {
		t.Errorf("Unexpected state\n   got: %#v\nwanted: %#v\n", resp.Body.String(), wanted)
	}

	h.TestRequestAuth(t, "PUT", "http://example.com/test1/state/teamdata", teamData, "team1", "", 401)

	if f, err := h.TestFSStateStore.fsys.Create("test1/teamids.txt"); err != nil {
		t.Error(err)
	} else {
		fmt.Fprintln(f, "team1")
		fmt.Fprintln(f, "team2")
		f.Close()
	}
	h.Refresh()

	stateTest := func(method string, target string, data []byte, teamid string, code int) {
		t.Helper()
		resp := h.TestRequestAuth(t, method, target, data, "", teamid, code)
		var obj StateExport
		if err := json.Unmarshal(resp.Body.Bytes(), &obj); err != nil {
			t.Error(err)
		}
		if obj.TeamData["self"] != string(data) {
			t.Error("Team data wrong", resp.Body.String())
		}
	}

	h.TestRequestAuth(t, "PUT", "http://example.com/test1/state/teamdata", nil, "team1", "", http.StatusLengthRequired)
	h.TestRequestAuth(t, "PUT", "http://example.com/test1/state/teamdata", make([]byte, 5000), "team1", "", http.StatusRequestEntityTooLarge)
	h.TestRequestAuth(t, "PUT", "http://example.com/test1/state/teamdata", teamData, "team1500", "", 401)
	h.TestRequestAuth(t, "PUT", "http://example.com/test1/state/teamdata", teamData, "team1", "", 200)
	h.TestRequestAuth(t, "PUT", "http://example.com/test1/state/teamdata", team2Data, "", "team2", 200)
	stateTest("PUT", "http://example.com/test1/state/teamdata", teamData, "team1", 200)
	stateTest("PUT", "http://example.com/test1/state/teamdata", team2Data, "team2", 200)
	h.Refresh()
	stateTest("PUT", "http://example.com/test1/state/teamdata", teamData, "team1", 409)
	stateTest("PUT", "http://example.com/test1/state/teamdata", team2Data, "team2", 409)
	stateTest("GET", "http://example.com/test1/state/current.json", team2Data, "team2", 200)
}

func TestHttpdUnlocked(t *testing.T) {
	h := NewTestHttpd(t, "/")
	h.MkInstance("test1")

	if err := h.TestFSStateStore.fsys.WriteFile("test1/schedule.txt", []byte("enable -\nrollout - cat0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	h.Refresh()

	state, found := h.Context.State("test1")
	if !found {
		t.Fatal("not found")
	}
	env := &Environment{
		State: state,
		ctx:   h.Context,
	}
	categories := state.Categories(h.now)
	if len(categories) == 0 {
		t.Fatal("no categories in this test requiring them")
	}
	for _, catid := range state.Categories(h.now) {
		cat, err := h.Category(catid)
		if err != nil {
			t.Error(err)
			continue
		}
		puzzles := env.Unlocked(cat)
		if len(puzzles) != 1 {
			t.Error("Wrong length puzzles", puzzles)
			t.Log(state.Hours(), h.now)
			t.Log(state.Categories(h.now))
		}
	}
}

func TestHttpdPuzzles(t *testing.T) {
	h := NewTestHttpd(t, "/")
	h.TestFSStateStore.MkInstance("test1")
	h.TestFSStateStore.fsys.WriteFile("test1/teams/team1", []byte("TeamData"), 0644)

	if _, err := h.Category("cat0"); err != nil {
		t.Fatal(err)
	}
	h.Refresh()
	if _, err := h.Category("cat0"); err != nil {
		t.Fatal(err)
	}

	h.TestRequest(t, "GET", "http://example.com/test1/puzzles/cat0/index.html", nil, 401)
	h.TestRequest(t, "GET", "http://example.com/test1/puzzles/cat0/1/not-here", nil, 401)
	h.TestRequest(t, "GET", "http://example.com/test1/puzzles/cat0/1/index.html", nil, 401)

	ha := h.Authenticate("team1", "")

	ha.TestRequest(t, "GET", "http://example.com/test1/puzzles/cat0/", nil, 200)
	ha.TestRequest(t, "GET", "http://example.com/test1/puzzles/cat0/index.html", nil, 301)
	ha.TestRequest(t, "GET", "http://example.com/test1/puzzles/cat0/1/", nil, 200)
	ha.TestRequest(t, "GET", "http://example.com/test1/puzzles/cat0/1/index.html", nil, 301)
	ha.TestRequest(t, "GET", "http://example.com/test1/puzzles/cat0/1/not-here-but-authenticated", nil, 404)
	ha.TestRequest(t, "GET", "http://example.com/test1/puzzles/cat0/two/index.html", nil, 404)
	ha.TestRequest(t, "GET", "http://example.com/test1/puzzles/cat0/two/moo.txt", nil, 404)
	if resp := ha.TestRequest(t, "GET", "http://example.com/test1/puzzles/cat0/1/moo.txt", nil, 200); resp.Body.String() != "moo.\n" {
		t.Error("Unexpected body", resp.Body.String())
	}

	ha.TestRequest(t, "GET", "http://example.com/test1/puzzles/cat0/__MOTH/answers.txt", nil, 404)
	ha.TestRequest(t, "GET", "http://example.com/test1/puzzles/cat0/1/__MOTH/answers.txt", nil, 404)

	ha.TestRequest(t, "PUT", "http://example.com/test1/puzzles/cat0/19324/answer", []byte("wrong"), 404)
	ha.TestRequest(t, "PUT", "http://example.com/test1/puzzles/cat0/1/answer", []byte("wrong"), 409)
	ha.TestRequest(t, "PUT", "http://example.com/test1/puzzles/cat0/1/answer", make([]byte, 5000), 413)
	ha.TestRequest(t, "PUT", "http://example.com/test1/puzzles/cat0/1/answer", []byte("answer123"), 200)
	ha.TestRequest(t, "PUT", "http://example.com/test1/puzzles/cat0/1/answer", []byte("answer123"), 200)
	h.Refresh()
	ha.TestRequest(t, "PUT", "http://example.com/test1/puzzles/cat0/1/answer", []byte("answer123"), 202)
	h.Refresh()
	ha.TestRequest(t, "GET", "http://example.com/test1/puzzles/cat0/two/", nil, 200)
	ha.TestRequest(t, "GET", "http://example.com/test1/puzzles/cat0/two/index.html", nil, 301)
	ha.TestRequest(t, "GET", "http://example.com/test1/puzzles/cat0/two/moo.txt", nil, 404)

	state := StateExport{}
	if resp := ha.TestRequest(t, "GET", "http://example.com/test1/state/current.json", nil, 200); len(resp.Body.String()) == 0 {
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

func TestHTTPZipPuzzles(t *testing.T) {
	s := TestdataCategoryStore(t)

	states := NewTestFSStateStore(t)

	ctx := &Context{
		CategoryStore: s,
		StateStore:    states.FSStateStore,
		WebFS:         http.FS(WebRootFS),
	}
	ctx.InstallHandlers("/")

	h := &TestHTTPServer{
		Context:          ctx,
		TestFSStateStore: states,
		t:                t,
		now:              time.Date(2024, 8, 15, 9, 0, 0, 0, time.UTC),
	}
	if err := h.TestFSStateStore.fsys.WriteFile("test/teams/team1", []byte("TeamData"), 0644); err != nil {
		t.Fatal(err)
	}
	h.Refresh()
	if err := h.TestFSStateStore.fsys.WriteFile("test/schedule.txt", []byte("enable -\nrollout - cat0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	h.Refresh()

	//
	// Now that we've set up the HTTP server, we can finally run a test
	//
	ha := h.Authenticate("team1", "")
	ha.TestRequest(t, "GET", "http://example.com/test/puzzles/cat0/1/", nil, 200)
	ha.TestRequest(t, "GET", "http://example.com/test/puzzles/cat0/__MOTH/answers.txt", nil, 404)
	ha.TestRequest(t, "GET", "http://example.com/test/puzzles/cat0/1/__MOTH/answers.txt", nil, 404)
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

	h := NewTestHttpd(t, "/")
	h.Context.CategoryStore = s

	h.now = h.now.Add(1 * time.Second)
	if err := h.Context.Refresh(h.now); err == nil {
		t.Fatal("Expecting error, none found")
	}

	if resp := h.TestRequest(t, "GET", "http://example.com/test/state/current.json", nil, 200); !strings.Contains(resp.Body.String(), "not a valid zip file") {
		t.Error("Wrong body:", resp.Body.String())
	}
}

func TestDisable(t *testing.T) {
	h := NewTestHttpd(t, "/")
	h.TestFSStateStore.fsys.WriteFile("test/teams/team1", []byte("TeamData"), 0644)
	ha := h.Authenticate("team1", "")

	if err := ha.TestFSStateStore.fsys.WriteFile("test/schedule.txt", []byte("rollout - cat0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	ha.Refresh()

	ha.TestRequest(t, "PUT", "http://example.com/test/puzzles/cat0/1/answer", []byte("answer123"), 503)

	if err := ha.TestFSStateStore.fsys.WriteFile("test/schedule.txt", []byte("enable -\nrollout - cat0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	ha.Refresh()

	ha.TestRequest(t, "PUT", "http://example.com/test/puzzles/cat0/1/answer", []byte("answer123"), 200)
}
