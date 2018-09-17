package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func respond(w http.ResponseWriter, req *http.Request, status Status, short string, description string) {
	// This is a kludge. Do proper parsing when this causes problems.
	accept := req.Header.Get("Accept")
	if strings.Contains(accept, "application/json") {
		ShowJSend(w, status, short, description)
	} else {
		ShowHtml(w, status, short, description)
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
	teamid := req.FormValue("t")
	token := req.FormValue("k")

	// Check answer
	if !anchoredSearchFile(ctx.StatePath("tokens.txt"), token, 0) {
		respond(
			w, req, Fail,
			"Unrecognized token",
			"I don't recognize that token. Did you type in the whole thing?",
		)
		return
	}

	parts := strings.Split(token, ":")
	category := ""
	pointstr := ""
	if len(parts) >= 2 {
		category = parts[0]
		pointstr = parts[1]
	}
	points, err := strconv.Atoi(pointstr)
	if err != nil {
		points = 0
	}
	// Defang category name; prevent directory traversal
	if matched, _ := regexp.MatchString("^[A-Za-z0-9_-]", category); matched {
		category = ""
	}

	if (category == "") || (points == 0) {
		respond(
			w, req, Fail,
			"Unrecognized token",
			"Something doesn't look right about that token",
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
		fmt.Sprintf("%d points for %s!", points, teamid),
	)
}

func (ctx Instance) answerHandler(w http.ResponseWriter, req *http.Request) {
	teamid := req.FormValue("t")
	category := req.FormValue("c")
	pointstr := req.FormValue("p")
	answer := req.FormValue("a")

	points, err := strconv.Atoi(pointstr)
	if err != nil {
		points = 0
	}

	catmb, ok := ctx.Categories[category]
	if !ok {
		respond(
			w, req, Fail,
			"Category does not exist",
			"The requested category does not exist. Sorry!",
		)
		return
	}

	// Get the answers
	haystack, err := catmb.Open("answers.txt")
	if err != nil {
		respond(
			w, req, Error,
			"Answers do not exist",
			"Please tell the contest people that the mothball for this category has no answers.txt in it!",
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
			err.Error(),
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
			parts := strings.Split(line, " ")
			if len(parts) != 2 {
				continue
			}
			pointval, err := strconv.Atoi(parts[0])
			if err != nil {
				log.Print(err)
				continue
			}
			dir := parts[1]

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
