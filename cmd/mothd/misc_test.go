// Data structures and functions to set up tests.
//
// This contains common initialization routines that simplify writing tests.

package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

	"github.com/dirtbags/moth/v5/pkg/wfs"
)

// TestClock provides a replayable clock, which provides consistent time for testing.
type TestClock struct {
	when time.Time
}

// Now advances the clock by one second, then returns the current time.
func (c *TestClock) Now() time.Time {
	c.when = c.when.Add(1 * time.Second)
	return c.when
}

// NewTestClock returns a new TestClock, initialized to 2024-07-05T22:34:10.00Z
func NewTestClock() *TestClock {
	return &TestClock{time.Date(2024, 7, 5, 22, 34, 10, 0, time.UTC)}
}

// TimeZero is epoch time 0.
var TimeZero = time.Unix(0, 0)

// TestTestClock ensures TestClock is working correctly.
func TestTestClock(t *testing.T) {
	c := NewTestClock()
	now := c.Now()
	if !c.Now().Equal(now.Add(1 * time.Second)) {
		t.Error("The test clock is broken")
	}
}

// TestFSState provides an FSState backed by a temporary directory.
//
// It also provides is a wrapper around FSState, providing an internal test clock.
type TestFSState struct {
	*FSState
	clock *TestClock
	t     *testing.T
}

// NewTestFSState returns a newly initialized TestFSState, backed by a temporary directory.
func NewTestFSState(t *testing.T) *TestFSState {
	fsys := wfs.DirFS(t.TempDir())
	state, err := NewFSState(fsys, "test-consolidate")
	if err != nil {
		t.Fatal(err)
	}
	s := &TestFSState{
		FSState: state,
		clock:   NewTestClock(),
		t:       t,
	}
	s.Refresh()
	return s
}

// Refresh the state and advance the clock.
func (s *TestFSState) Refresh() {
	s.t.Helper()
	errs := s.FSState.Refresh(s.clock.Now())
	for _, err := range errs {
		if err != nil {
			s.t.Error(err)
		}
	}
}

// TestFSStateStore contains an FSStateStore (a collection of FSStates),
// backed by a temporary directory on the filesystem,
// and with an associated TestClock.
type TestFSStateStore struct {
	*FSStateStore
	clock *TestClock
	t     *testing.T
}

// NewTestFSStateStore returns a TestFSStateStore with an
// uninitialized test instance.
//
// The test instance will be enabled, have cat0 rolled out,
// and contain a team with id TestTeamID.
func NewTestFSStateStore(t *testing.T) *TestFSStateStore {
	fsys := wfs.DirFS(t.TempDir())
	store := &TestFSStateStore{
		FSStateStore: NewFSStateStore(fsys, "testhost", 10, 1),
		clock:        NewTestClock(),
		t:            t,
	}
	store.MkInstance("test")
	return store
}

// Refresh all TestFSStates in store, and advance the clock.
func (store *TestFSStateStore) Refresh() {
	store.t.Helper()
	if err := store.FSStateStore.Refresh(store.clock.Now()); err != nil {
		store.t.Error(err)
	}
}

// TestTeamID is the registered team ID with most test instances.
//
// Specifically, the team ID will be registered for any test instances created with MkInstance.
const TestTeamID = TeamID("test-team")

// MkInstance creates a new TestFSState with the given name.
//
// The new instance will be enabled, have cat0 rolled out,
// and contain a team with id TestTeamID.
func (store *TestFSStateStore) MkInstance(name string) {
	if err := store.fsys.Mkdir(name, 0755); err != nil {
		store.t.Fatal(err)
	}
	store.Refresh()
	if err := store.fsys.WriteFile(name+"/schedule.txt", []byte("enable -\nrollout - cat0\n"), 0644); err != nil {
		store.t.Fatal(err)
	}
	if err := store.fsys.WriteFile(name+"/teams/"+(TestTeamID.String()), []byte(TestTeamID), 0644); err != nil {
		store.t.Fatal(err)
	}
	store.Refresh()
}

// TestContext provides a Context backed by testdata categories, an
// FSStateStore in a temporary directory, and a minimal static content
// filesystem.
type TestContext struct {
	*Context
	stateStore *TestFSStateStore
	clock      *TestClock
	t          *testing.T
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

// NewTestContext returns a TestContext with a ready-to-use "test" instance.
// The instance will have TestTeamID already registered.
func NewTestContext(t *testing.T, base string) *TestContext {
	states := NewTestFSStateStore(t)

	ctx := &Context{
		CategoryStore: TestdataCategoryStore(t),
		StateStore:    states.FSStateStore,
		WebFS:         http.FS(WebRootFS),
	}
	ctx.InstallHandlers(base)

	h := &TestContext{
		Context:    ctx,
		stateStore: states,
		clock:      NewTestClock(),
		t:          t,
	}

	return h
}

// Refresh the context and advance the clock.
func (h *TestContext) Refresh() {
	h.t.Helper()
	if err := h.Context.Refresh(h.clock.Now()); err != nil {
		h.t.Fatal(err)
	}
}

// TestServe serves req, logging an error if the returned HTTP code is not code.
//
// The returned httptest.ResponseRecorder can be further examined, if desired.
func (h *TestContext) TestServe(req *http.Request, code int) *httptest.ResponseRecorder {
	h.t.Helper()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Result().StatusCode != code {
		h.t.Errorf("Expected %d, got %s", code, w.Result().Status)
		h.t.Logf("     Req: %s %s", req.Method, req.URL.Path)
		h.t.Logf("  Header: %v", w.Result().Header)
		body, _ := io.ReadAll(w.Result().Body)
		h.t.Logf("    Body: %#v", string(body))
	}
	return w
}

// TestRequestAuth makes an authenticated HTTP request to target (a URL) using method.
//
// If usernade and password are empty strings, the request will not be authenticated.
//
// If body is non-nil, it will be sent as the body of the request.
// If the returned HTTP code does not match code,
// an error is logged.
func (h *TestContext) TestRequestAuth(method string, target string, body []byte, username string, password string, code int) *httptest.ResponseRecorder {
	h.t.Helper()
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	request := httptest.NewRequest(method, target, bodyReader)
	if (username != "") || (password != "") {
		request.SetBasicAuth(username, password)
	}
	return h.TestServe(request, code)
}

// TestRequestAnonymous makes an unathenticated HTTP request to target (a URL) using method.
//
// If body is non-nil, it will be sent as the body of the request.
// If the returned HTTP code does not match code,
// an error is logged.
func (h *TestContext) TestRequestAnonymous(method string, target string, body []byte, code int) *httptest.ResponseRecorder {
	return h.TestRequestAuth(method, target, body, "", "", code)
}

// TestRequest makes an authenticated HTTP request to target (a URL) using method.
//
// The request will be authenticated with TestTeamID.
//
// If body is non-nil, it will be sent as the body of the request.
// If the returned HTTP code does not match code,
// an error is logged.
func (hsa *TestContext) TestRequest(method string, target string, body []byte, code int) *httptest.ResponseRecorder {
	hsa.t.Helper()
	return hsa.TestRequestAuth(method, target, body, TestTeamID.String(), TestTeamID.String(), code)
}
