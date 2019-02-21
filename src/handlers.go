package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type JSend struct {
	Status string    `json:"status"`
	Data   JSendData `json:"data"`
}
type JSendData struct {
	Short       string `json:"short"`
	Description string `json:"description"`
}

type Status int

const (
	Success = iota
	Fail
	Error
)

func respond(w http.ResponseWriter, req *http.Request, status Status, short string, format string, a ...interface{}) {
	resp := JSend{
		Status: "success",
		Data: JSendData{
			Short:       short,
			Description: fmt.Sprintf(format, a...),
		},
	}
	switch status {
	case Success:
		resp.Status = "success"
	case Fail:
		resp.Status = "fail"
	default:
		resp.Status = "error"
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // RFC2616 makes it pretty clear that 4xx codes are for the user-agent
	w.Write(respBytes)
}

// hasLine returns true if line appears in r.
// The entire line must match.
func hasLine(r io.Reader, line string) bool {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if scanner.Text() == line {
			return true
		}
	}
	return false
}

func (ctx *Instance) registerHandler(w http.ResponseWriter, req *http.Request) {
	teamname := req.FormValue("name")
	teamid := req.FormValue("id")

	// Keep foolish operators from shooting themselves in the foot
	// You would have to add a pathname to your list of Team IDs to open this vulnerability,
	// but I have learned not to overestimate people.
	if strings.Contains(teamid, "../") {
		teamid = "rodney"
	}

	if (teamid == "") || (teamname == "") {
		respond(
			w, req, Fail,
			"Invalid Entry",
			"Either `id` or `name` was missing from this request.",
		)
		return
	}

	teamids, err := os.Open(ctx.StatePath("teamids.txt"))
	if err != nil {
		respond(
			w, req, Fail,
			"Cannot read valid team IDs",
			"An error was encountered trying to read valid teams IDs: %v", err,
		)
		return
	}
	defer teamids.Close()
	if !hasLine(teamids, teamid) {
		respond(
			w, req, Fail,
			"Invalid Team ID",
			"I don't have a record of that team ID. Maybe you used capital letters accidentally?",
		)
		return
	}

	f, err := os.OpenFile(ctx.StatePath("teams", teamid), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		log.Print(err)
		respond(
			w, req, Fail,
			"Registration failed",
			"Unable to register. Perhaps a teammate has already registered?",
		)
		return
	}
	defer f.Close()
	fmt.Fprintln(f, teamname)
	respond(
		w, req, Success,
		"Team registered",
		"Okay, your team has been named and you may begin using your team ID!",
	)
}

func (ctx *Instance) answerHandler(w http.ResponseWriter, req *http.Request) {
	teamid := req.FormValue("id")
	category := req.FormValue("cat")
	pointstr := req.FormValue("points")
	answer := req.FormValue("answer")

	points, err := strconv.Atoi(pointstr)
	if err != nil {
		respond(
			w, req, Fail,
			"Cannot parse point value",
			"This doesn't look like an integer: %s", pointstr,
		)
		return
	}

	haystack, err := ctx.OpenCategoryFile(category, "answers.txt")
	if err != nil {
		respond(
			w, req, Fail,
			"Cannot list answers",
			"Unable to read the list of answers for this category.",
		)
		return
	}
	defer haystack.Close()

	// Look for the answer
	needle := fmt.Sprintf("%d %s", points, answer)
	if !hasLine(haystack, needle) {
		respond(
			w, req, Fail,
			"Wrong answer",
			"That is not the correct answer for %s %d.", category, points,
		)
		return
	}

	if err := ctx.AwardPoints(teamid, category, points); err != nil {
		respond(
			w, req, Error,
			"Cannot award points",
			"The answer is correct, but there was an error awarding points: %v", err.Error(),
		)
		return
	}
	respond(
		w, req, Success,
		"Points awarded",
		fmt.Sprintf("%d points for %s!", points, teamid),
	)
}

func (ctx *Instance) puzzlesHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ctx.jPuzzleList)
}

func (ctx *Instance) pointsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ctx.jPointsLog)
}

func (ctx *Instance) contentHandler(w http.ResponseWriter, req *http.Request) {
	// Prevent directory traversal
	if strings.Contains(req.URL.Path, "/.") {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Be clever: use only the last three parts of the path. This may prove to be a bad idea.
	parts := strings.Split(req.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	fileName := parts[len(parts)-1]
	puzzleId := parts[len(parts)-2]
	categoryName := parts[len(parts)-3]

	mb, ok := ctx.Categories[categoryName]
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	mbFilename := fmt.Sprintf("content/%s/%s", puzzleId, fileName)
	mf, err := mb.Open(mbFilename)
	if err != nil {
		log.Print(err)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	defer mf.Close()

	http.ServeContent(w, req, fileName, mf.ModTime(), mf)
}

func (ctx *Instance) staticHandler(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	if strings.Contains(path, "..") {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	if path == "/" {
		path = "/index.html"
	}

	f, err := os.Open(ctx.ResourcePath(path))
	if err != nil {
		http.NotFound(w, req)
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		http.NotFound(w, req)
		return
	}

	http.ServeContent(w, req, path, d.ModTime(), f)
}

type FurtiveResponseWriter struct {
	w http.ResponseWriter
	statusCode *int
}

func (w FurtiveResponseWriter) WriteHeader(statusCode int) {
	*w.statusCode = statusCode
	w.w.WriteHeader(statusCode)
}

func (w FurtiveResponseWriter) Write(buf []byte) (n int, err error) {
	n, err = w.w.Write(buf)
	return
}

func (w FurtiveResponseWriter) Header() http.Header {
	return w.w.Header()
}

// This gives Instances the signature of http.Handler
func (ctx *Instance) ServeHTTP(wOrig http.ResponseWriter, r *http.Request) {
	w := FurtiveResponseWriter{
		w: wOrig,
		statusCode: new(int),
	}
	w.Header().Set("WWW-Authenticate", "Basic")
	_, password, _ := r.BasicAuth()
	if password != ctx.Password {
		http.Error(w, "Authentication Required", 401)
	} else {
		ctx.mux.ServeHTTP(w, r)
	}
	log.Printf(
		"%s %s %s %d\n",
		r.RemoteAddr,
		r.Method,
		r.URL,
		*w.statusCode,
	)
}

func (ctx *Instance) BindHandlers() {
	ctx.mux.HandleFunc(ctx.Base+"/", ctx.staticHandler)
	ctx.mux.HandleFunc(ctx.Base+"/register", ctx.registerHandler)
	ctx.mux.HandleFunc(ctx.Base+"/answer", ctx.answerHandler)
	ctx.mux.HandleFunc(ctx.Base+"/content/", ctx.contentHandler)
	ctx.mux.HandleFunc(ctx.Base+"/puzzles.json", ctx.puzzlesHandler)
	ctx.mux.HandleFunc(ctx.Base+"/points.json", ctx.pointsHandler)
}

