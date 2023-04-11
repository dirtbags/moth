package main

import (
	"bytes"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dirtbags/moth/v4/pkg/jsend"
)

// HTTPServer is a MOTH HTTP server
type HTTPServer struct {
	*http.ServeMux
	server *MothServer
	base   string
}

// NewHTTPServer creates a MOTH HTTP server, with handler functions registered
func NewHTTPServer(base string, server *MothServer) *HTTPServer {
	base = strings.TrimRight(base, "/")
	h := &HTTPServer{
		ServeMux: http.NewServeMux(),
		server:   server,
		base:     base,
	}
	h.HandleMothFunc("/", h.ThemeHandler)
	h.HandleMothFunc("/state", h.StateHandler)
	h.HandleMothFunc("/register", h.RegisterHandler)
	h.HandleMothFunc("/answer", h.AnswerHandler)
	h.HandleMothFunc("/content/", h.ContentHandler)

	if server.Config.Devel {
		h.HandleMothFunc("/mothballer/", h.MothballerHandler)
	}
	return h
}

// HandleMothFunc binds a new handler function which creates a new MothServer with every request
func (h *HTTPServer) HandleMothFunc(
	pattern string,
	mothHandler func(MothRequestHandler, http.ResponseWriter, *http.Request),
) {
	handler := func(w http.ResponseWriter, req *http.Request) {
		teamID := req.FormValue("id")
		mh := h.server.NewHandler(teamID)
		mothHandler(mh, w, req)
	}
	h.HandleFunc(h.base+pattern, handler)
}

// ServeHTTP provides the http.Handler interface
func (h *HTTPServer) ServeHTTP(wOrig http.ResponseWriter, r *http.Request) {
	w := StatusResponseWriter{
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

// StatusResponseWriter provides a ResponseWriter that remembers what the status code was
type StatusResponseWriter struct {
	statusCode *int
	http.ResponseWriter
}

// WriteHeader sends an HTTP response header with the provided status code
func (w StatusResponseWriter) WriteHeader(statusCode int) {
	*w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Run binds to the provided bindStr, and serves incoming requests until failure
func (h *HTTPServer) Run(bindStr string) {
	log.Printf("Listening on %s", bindStr)
	log.Fatal(http.ListenAndServe(bindStr, h))
}

// ThemeHandler serves up static content from the theme directory
func (h *HTTPServer) ThemeHandler(mh MothRequestHandler, w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	f, mtime, err := mh.ThemeOpen(path)
	if err != nil {
		http.NotFound(w, req)
		return
	}
	defer f.Close()
	http.ServeContent(w, req, path, mtime, f)
}

// StateHandler returns the full JSON-encoded state of the event
func (h *HTTPServer) StateHandler(mh MothRequestHandler, w http.ResponseWriter, req *http.Request) {
	jsend.JSONWrite(w, mh.ExportState())
}

// RegisterHandler handles attempts to register a team
func (h *HTTPServer) RegisterHandler(mh MothRequestHandler, w http.ResponseWriter, req *http.Request) {
	teamName := req.FormValue("name")
	teamName = strings.TrimSpace(teamName)
	if teamName == "" {
		jsend.Sendf(w, jsend.Fail, "empty name", "Team name may not be empty")
		return
	}

	if err := mh.Register(teamName); err == ErrAlreadyRegistered {
		jsend.Sendf(w, jsend.Success, "already registered", "team ID has already been registered")
	} else if err != nil {
		jsend.Sendf(w, jsend.Fail, "not registered", err.Error())
	} else {
		jsend.Sendf(w, jsend.Success, "registered", "team ID registered")
	}
}

// AnswerHandler checks answer correctness and awards points
func (h *HTTPServer) AnswerHandler(mh MothRequestHandler, w http.ResponseWriter, req *http.Request) {
	cat := req.FormValue("cat")
	pointstr := req.FormValue("points")
	answer := req.FormValue("answer")

	points, _ := strconv.Atoi(pointstr)

	if err := mh.CheckAnswer(cat, points, answer); err != nil {
		jsend.Sendf(w, jsend.Fail, "not accepted", err.Error())
	} else {
		jsend.Sendf(w, jsend.Success, "accepted", "%d points awarded in %s", points, cat)
	}
}

// ContentHandler returns static content from a given puzzle
func (h *HTTPServer) ContentHandler(mh MothRequestHandler, w http.ResponseWriter, req *http.Request) {
	parts := strings.SplitN(req.URL.Path[len(h.base)+1:], "/", 4)
	if len(parts) < 4 {
		http.NotFound(w, req)
		return
	}

	// parts[0] == "content"
	cat := parts[1]
	pointsStr := parts[2]
	filename := parts[3]

	if filename == "" {
		filename = "puzzle.json"
	}

	points, _ := strconv.Atoi(pointsStr)

	mf, mtime, err := mh.PuzzlesOpen(cat, points, filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer mf.Close()

	http.ServeContent(w, req, filename, mtime, mf)
}

// MothballerHandler returns a mothball
func (h *HTTPServer) MothballerHandler(mh MothRequestHandler, w http.ResponseWriter, req *http.Request) {
	parts := strings.SplitN(req.URL.Path[len(h.base)+1:], "/", 2)
	if len(parts) < 2 {
		http.NotFound(w, req)
		return
	}

	// parts[0] == "mothballer"
	filename := parts[1]
	cat := strings.TrimSuffix(filename, ".mb")
	mb := new(bytes.Buffer)
	if err := mh.Mothball(cat, mb); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mbReader := bytes.NewReader(mb.Bytes())
	http.ServeContent(w, req, filename, time.Now(), mbReader)
}
