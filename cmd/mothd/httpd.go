package main

import (
	"log"
	"net/http"
	"strings"
)

type HTTPServer struct {
	*http.ServeMux
	Puzzles PuzzleProvider
	Theme   ThemeProvider
	State   StateProvider
}

func NewHTTPServer(base string, theme ThemeProvider, state StateProvider, puzzles PuzzleProvider) *HTTPServer {
	h := &HTTPServer{
		ServeMux: http.NewServeMux(),
		Puzzles:  puzzles,
		Theme:    theme,
		State:    state,
	}
	base = strings.TrimRight(base, "/")
	h.HandleFunc(base+"/", h.ThemeHandler)
	h.HandleFunc(base+"/state", h.StateHandler)
	h.HandleFunc(base+"/register", h.RegisterHandler)
	h.HandleFunc(base+"/answer", h.AnswerHandler)
	h.HandleFunc(base+"/content/", h.ContentHandler)
	return h
}

func (h *HTTPServer) Run(bindStr string) {
	log.Printf("Listening on %s", bindStr)
	log.Fatal(http.ListenAndServe(bindStr, h))
}

type MothResponseWriter struct {
	statusCode *int
	http.ResponseWriter
}

func (w MothResponseWriter) WriteHeader(statusCode int) {
	*w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// This gives Instances the signature of http.Handler
func (h *HTTPServer) ServeHTTP(wOrig http.ResponseWriter, r *http.Request) {
	w := MothResponseWriter{
		statusCode:     new(int),
		ResponseWriter: wOrig,
	}
	h.ServeMux.ServeHTTP(w, r)
	log.Printf(
		"%s %s %s %d\n",
		r.RemoteAddr,
		r.Method,
		r.URL,
		*w.statusCode,
	)
}

func (h *HTTPServer) ThemeHandler(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	f, err := h.Theme.Open(path)
	if err != nil {
		http.NotFound(w, req)
		return
	}
	defer f.Close()
	mtime, _ := h.Theme.ModTime(path)
	http.ServeContent(w, req, path, mtime, f)
}

func (h *HTTPServer) StateHandler(w http.ResponseWriter, req *http.Request) {
	var state struct {
		Config struct {
			Devel bool
		}
		Messages  string
		TeamNames map[string]string
		PointsLog []Award
		Puzzles   map[string][]int
	}

	teamId := req.FormValue("id")
	export := h.State.Export(teamId)

	state.Messages = export.Messages
	state.TeamNames = export.TeamNames
	state.PointsLog = export.PointsLog

	state.Puzzles = make(map[string][]int)
	log.Println("Puzzles")
	for category := range h.Puzzles.Inventory() {
		log.Println(category)
	}

	JSONWrite(w, state)
}

func (h *HTTPServer) RegisterHandler(w http.ResponseWriter, req *http.Request) {
	teamId := req.FormValue("id")
	teamName := req.FormValue("name")
	if err := h.State.SetTeamName(teamId, teamName); err != nil {
		JSendf(w, JSendFail, "not registered", err.Error())
	} else {
		JSendf(w, JSendSuccess, "registered", "Team ID registered")
	}
}

func (h *HTTPServer) AnswerHandler(w http.ResponseWriter, req *http.Request) {
	JSendf(w, JSendFail, "unimplemented", "I haven't written this yet")
}

func (h *HTTPServer) ContentHandler(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	if path == "/" {
		path = "/puzzle.json"
	}
	JSendf(w, JSendFail, "unimplemented", "I haven't written this yet")
}
