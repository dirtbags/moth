package main

import (
	"log"
	"net/http"
	"strings"
	"strconv"
)

type HTTPServer struct {
	*http.ServeMux
	server  *MothServer
	base    string
}

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
	return h
}

func (h *HTTPServer) HandleMothFunc(
	pattern string,
	mothHandler func(MothRequestHandler, http.ResponseWriter, *http.Request),
) {
	handler := func(w http.ResponseWriter, req *http.Request) {
		mh := h.server.NewHandler(req.FormValue("id"))
		mothHandler(mh, w, req)
	}
	h.HandleFunc(h.base + pattern, handler)
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

type MothResponseWriter struct {
	statusCode *int
	http.ResponseWriter
}

func (w MothResponseWriter) WriteHeader(statusCode int) {
	*w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (h *HTTPServer) Run(bindStr string) {
	log.Printf("Listening on %s", bindStr)
	log.Fatal(http.ListenAndServe(bindStr, h))
}

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

func (h *HTTPServer) StateHandler(mh MothRequestHandler, w http.ResponseWriter, req *http.Request) {
	JSONWrite(w, mh.ExportState())
}

func (h *HTTPServer) RegisterHandler(mh MothRequestHandler, w http.ResponseWriter, req *http.Request) {
	teamName := req.FormValue("name")
	if err := mh.Register(teamName); err != nil {
		JSendf(w, JSendFail, "not registered", err.Error())
	} else {
		JSendf(w, JSendSuccess, "registered", "Team ID registered")
	}
}

func (h *HTTPServer) AnswerHandler(mh MothRequestHandler, w http.ResponseWriter, req *http.Request) {
	cat := req.FormValue("cat")
	pointstr := req.FormValue("points")
	answer := req.FormValue("answer")
	
	points, _ := strconv.Atoi(pointstr)
	
	if err := mh.CheckAnswer(cat, points, answer); err != nil {
		JSendf(w, JSendFail, "not accepted", err.Error())
	} else {
		JSendf(w, JSendSuccess, "accepted", "%d points awarded in %s", points, cat)
	}
}

func (h *HTTPServer) ContentHandler(mh MothRequestHandler, w http.ResponseWriter, req *http.Request) {
	trimLen := len(h.base) + len("/content/")
	parts := strings.SplitN(req.URL.Path[trimLen:], "/", 3)
  if len(parts) < 3 {
    http.Error(w, "Not Found", http.StatusNotFound)
    return
  }

	cat := parts[0]
	pointsStr := parts[1]
	filename := parts[2]

	if (filename == "") {
		filename = "puzzles.json"
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
