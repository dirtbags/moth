package main

import (
	"net/http"
	"log"
	"strings"
	"github.com/dirtbags/moth/jsend"
)

type HTTPServer struct {
	PuzzleProvider
	ThemeProvider
	StateProvider
	*http.ServeMux
}

func NewHTTPServer(base string, theme ThemeProvider, state StateProvider, puzzles PuzzleProvider) (*HTTPServer)  {
	h := &HTTPServer{
		ThemeProvider: theme,
		StateProvider: state,
		PuzzleProvider: puzzles,
		ServeMux: http.NewServeMux(),
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
		statusCode: new(int),
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
	if strings.Contains(path, "..") {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	if path == "/" {
		path = "/index.html"
	}

	f, err := h.ThemeProvider.Open(path)
	if err != nil {
		http.NotFound(w, req)
		return
	}
	defer f.Close()
	mtime, _ := h.ThemeProvider.ModTime(path)
	http.ServeContent(w, req, path, mtime, f)
}


func (h *HTTPServer) StateHandler(w http.ResponseWriter, req *http.Request) {
	jsend.Write(w, jsend.Fail, "unimplemented", "I haven't written this yet")
}

func (h *HTTPServer) RegisterHandler(w http.ResponseWriter, req *http.Request) {
	jsend.Write(w, jsend.Fail, "unimplemented", "I haven't written this yet")
}

func (h *HTTPServer) AnswerHandler(w http.ResponseWriter, req *http.Request) {
	jsend.Write(w, jsend.Fail, "unimplemented", "I haven't written this yet")
}

func (h *HTTPServer) ContentHandler(w http.ResponseWriter, req *http.Request) {
	jsend.Write(w, jsend.Fail, "unimplemented", "I haven't written this yet")
}
