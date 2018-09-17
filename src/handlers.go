package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func respond(w http.ResponseWriter, req *http.Request, status Status, short string, format string, a ...interface{}) {
	long := fmt.Sprintf(format, a...)
	// This is a kludge. Do proper parsing when this causes problems.
	accept := req.Header.Get("Accept")
	if strings.Contains(accept, "application/json") {
		ShowJSend(w, status, short, long)
	} else {
		ShowHtml(w, status, short, long)
	}
}

func (ctx Instance) registerHandler(w http.ResponseWriter, req *http.Request) {
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

	if !anchoredSearchFile(ctx.StatePath("teamids.txt"), teamid, 0) {
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

func (ctx Instance) tokenHandler(w http.ResponseWriter, req *http.Request) {
	teamid := req.FormValue("id")
	token := req.FormValue("token")

	var category string
	var points int
	var fluff string

	stoken := strings.Replace(token, ":", " ", 2)
	n, err := fmt.Sscanf(stoken, "%s %d %s", &category, &points, &fluff)
	if err != nil || n != 3 {
		respond(
			w, req, Fail,
			"Malformed token",
			"That doesn't look like a token: %v.", err,
		)
		return
	}

	if (category == "") || (points <= 0) {
		respond(
			w, req, Fail,
			"Weird token",
			"That token doesn't make any sense.",
		)
		return
	}

	f, err := ctx.OpenCategoryFile(category, "tokens.txt")
	if err != nil {
		respond(
			w, req, Fail,
			"Cannot list valid tokens",
			err.Error(),
		)
		return
	}
	defer f.Close()

	// Make sure the token is in the list
	if !anchoredSearch(f, token, 0) {
		respond(
			w, req, Fail,
			"Unrecognized token",
			"I don't recognize that token. Did you type in the whole thing?",
		)
		return
	}

	if err := ctx.AwardPoints(teamid, category, points); err != nil {
		respond(
			w, req, Fail,
			"Error awarding points",
			err.Error(),
		)
		return
	}
	respond(
		w, req, Success,
		"Points awarded",
		"%d points for %s!", points, teamid,
	)
}

func (ctx Instance) answerHandler(w http.ResponseWriter, req *http.Request) {
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
	if !anchoredSearch(haystack, needle, 0) {
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
			"Error awarding points",
			err.Error(),
		)
		return
	}
	respond(
		w, req, Success,
		"Points awarded",
		fmt.Sprintf("%d points for %s!", points, teamid),
	)
}

type PuzzleMap struct {
	Points int    `json:"points"`
	Path   string `json:"path"`
}

func (ctx Instance) puzzlesHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	res := map[string][]PuzzleMap{}
	for catName, mb := range ctx.Categories {
		mf, err := mb.Open("map.txt")
		if err != nil {
			log.Print(err)
		}
		defer mf.Close()

		pm := make([]PuzzleMap, 0, 30)
		scanner := bufio.NewScanner(mf)
		for scanner.Scan() {
			line := scanner.Text()

			var pointval int
			var dir string

			n, err := fmt.Sscanf(line, "%d %s", &pointval, &dir)
			if err != nil {
				log.Printf("Parsing map for %s: %v", catName, err)
				continue
			} else if n != 2 {
				log.Printf("Parsing map for %s: short read", catName)
				continue
			}

			pm = append(pm, PuzzleMap{pointval, dir})
			log.Print(pm)
		}

		res[catName] = pm
		log.Print(res)
	}
	jres, _ := json.Marshal(res)
	w.Write(jres)
}

func (ctx Instance) pointsHandler(w http.ResponseWriter, req *http.Request) {
	log := ctx.PointsLog()
	jlog, err := json.Marshal(log)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write(jlog)
}

func (ctx Instance) staticHandler(w http.ResponseWriter, req *http.Request) {
	ServeStatic(w, req, ctx.ResourcesDir)
}

func (ctx Instance) BindHandlers(mux *http.ServeMux) {
	mux.HandleFunc(ctx.Base+"/", ctx.staticHandler)
	mux.HandleFunc(ctx.Base+"/register", ctx.registerHandler)
	mux.HandleFunc(ctx.Base+"/token", ctx.tokenHandler)
	mux.HandleFunc(ctx.Base+"/answer", ctx.answerHandler)
	mux.HandleFunc(ctx.Base+"/puzzles.json", ctx.puzzlesHandler)
	mux.HandleFunc(ctx.Base+"/points.json", ctx.pointsHandler)
}
