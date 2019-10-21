package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// https://github.com/omniti-labs/jsend
type JSend struct {
	Status string `json:"status"`
	Data   struct {
		Short       string `json:"short"`
		Description string `json:"description"`
	} `json:"data"`
}

const (
	JSendSuccess = "success"
	JSendFail    = "fail"
	JSendError   = "error"
)

func respond(w http.ResponseWriter, req *http.Request, status string, short string, format string, a ...interface{}) {
	resp := JSend{}
	resp.Status = status
	resp.Data.Short = short
	resp.Data.Description = fmt.Sprintf(format, a...)

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
	teamName := req.FormValue("name")
	teamId := req.FormValue("id")

	if !ctx.ValidTeamId(teamId) {
		respond(
			w, req, JSendFail,
			"Invalid Team ID",
			"I don't have a record of that team ID. Maybe you used capital letters accidentally?",
		)
		return
	}

	f, err := os.OpenFile(ctx.StatePath("teams", teamId), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			respond(
				w, req, JSendFail,
				"Already registered",
				"This team ID has already been registered.",
			)
		} else {
			log.Print(err)
			respond(
				w, req, JSendFail,
				"Registration failed",
				"Unable to register. Perhaps a teammate has already registered?",
			)
		}
		return
	}
	defer f.Close()

	fmt.Fprintln(f, teamName)
	respond(
		w, req, JSendSuccess,
		"Team registered",
		"Your team has been named and you may begin using your team ID!",
	)
}

func (ctx *Instance) answerHandler(w http.ResponseWriter, req *http.Request) {
	teamId := req.FormValue("id")
	category := req.FormValue("cat")
	pointstr := req.FormValue("points")
	answer := req.FormValue("answer")

	if ! ctx.ValidTeamId(teamId) {
		respond(
			w, req, JSendFail,
			"Invalid team ID",
			"That team ID is not valid for this event.",
		)
		return
	}
	if ctx.TooFast(teamId) {
		respond(
			w, req, JSendFail,
			"Submitting too quickly",
			"Your team can only submit one answer every %v", ctx.AttemptInterval,
		)
		return
	}

	points, err := strconv.Atoi(pointstr)
	if err != nil {
		respond(
			w, req, JSendFail,
			"Cannot parse point value",
			"This doesn't look like an integer: %s", pointstr,
		)
		return
	}

	haystack, err := ctx.OpenCategoryFile(category, "answers.txt")
	if err != nil {
		respond(
			w, req, JSendFail,
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
			w, req, JSendFail,
			"Wrong answer",
			"That is not the correct answer for %s %d.", category, points,
		)
		return
	}

	if err := ctx.AwardPoints(teamId, category, points); err != nil {
		respond(
			w, req, JSendError,
			"Cannot award points",
			"The answer is correct, but there was an error awarding points: %v", err.Error(),
		)
		return
	}
	respond(
		w, req, JSendSuccess,
		"Points awarded",
		fmt.Sprintf("%d points for %s!", points, teamId),
	)
}

func (ctx *Instance) puzzlesHandler(w http.ResponseWriter, req *http.Request) {
	teamId := req.FormValue("id")
	if _, err := ctx.TeamName(teamId); err != nil {
		http.Error(w, "Must provide team ID", http.StatusUnauthorized)
		return
	}

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

	mb, ok := ctx.categories[categoryName]
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

	f, err := os.Open(ctx.ThemePath(path))
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

func (ctx *Instance) dehydrateHandler(w http.ResponseWriter, req *http.Request) {
	/*
	teamId := req.FormValue("id")
	if _, err := ctx.TeamName(teamId); err != nil {
		http.Error(w, "Must provide team ID", http.StatusUnauthorized)
		return
	}
	*/

	zipBaseDir := "moth"

	buf := new(bytes.Buffer)

	zipWriter := zip.NewWriter( buf )

	writeZipContents(zipWriter, fmt.Sprintf("%s/%s", zipBaseDir, "puzzles.json"), ctx.jPuzzleList)
	writeZipContents(zipWriter, fmt.Sprintf("%s/%s", zipBaseDir, "points.json"), ctx.jPointsLog)

	for category_name, category := range ctx.categories {
		for _, file := range category.zf.File {
			parts := strings.Split(file.Name, "/")

			if (parts[0] == "content") {
				fmt.Printf("Inspecting %s/%s\n", category_name, file.Name)
				local_buf := new(bytes.Buffer)
				fh, _ := file.Open()
				defer fh.Close()
				local_buf.ReadFrom(fh)
				writeZipContents(zipWriter, fmt.Sprintf("%s/%s/%s", zipBaseDir, category_name, file.Name), local_buf.Bytes())
			}
		}
	}

	filepath.Walk(ctx.ThemeDir, func (path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if ! info.IsDir() {
			fmt.Printf("Walking %s\n", path)
			local_buf := new(bytes.Buffer)
			fh, _ := os.Open(path)
			defer fh.Close()
			local_buf.ReadFrom(fh)
			writeZipContents(zipWriter, fmt.Sprintf("%s/%s", zipBaseDir, path), local_buf.Bytes())
		}
		return nil
	})

	zipWriter.Close()

	w.Header().Set("Content-Type", "application/zip")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func writeZipContents(z *zip.Writer, path string, contents []byte) error {
	f, err := z.Create(path)
	if err != nil {
		return err
	}
	f.Write(contents)

	return nil
}

type FurtiveResponseWriter struct {
	w          http.ResponseWriter
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
		w:          wOrig,
		statusCode: new(int),
	}
	ctx.mux.ServeHTTP(w, r)
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
	ctx.mux.HandleFunc(ctx.Base+"/dehydrate", ctx.dehydrateHandler)
}
