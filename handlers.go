package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"strconv"
)

func registerHandler(w http.ResponseWriter, req *http.Request) {
	teamname := req.FormValue("n")
	teamid := req.FormValue("h")
	
	if matched, _ := regexp.MatchString("[^0-9a-z]", teamid); matched {
		teamid = ""
	}
	
	if (teamid == "") || (teamname == "") {
		showPage(w, "Invalid Entry", "Oops! Are you sure you got that right?")
		return
	}
	
	if ! anchoredSearch(statePath("teamids.txt"), teamid, 0) {
		showPage(w, "Invalid Team ID", "I don't have a record of that team ID. Maybe you used capital letters accidentally?")
		return
	}
	
	f, err := os.OpenFile(statePath("state", teamid), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		showPage(
			w,
			"Registration failed",
			"Unable to register. Perhaps a teammate has already registered?",
		)
		return
	}
	defer f.Close()
	fmt.Fprintln(f, teamname)
	showPage(w, "Success", "Okay, your team has been named and you may begin using your team ID!")
}

func tokenHandler(w http.ResponseWriter, req *http.Request) {
	teamid := req.FormValue("t")
	token := req.FormValue("k")

	// Check answer
	if ! anchoredSearch(token, statePath("tokens.txt"), 0) {
		showPage(w, "Unrecognized token", "I don't recognize that token. Did you type in the whole thing?")
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
		showPage(w, "Unrecognized token", "Something doesn't look right about that token")
		return
	}
	
	if err := awardPoints(teamid, category, points); err != nil {
		showPage(w, "Error awarding points", err.Error())
		return
	}
	showPage(w, "Points awarded", fmt.Sprintf("%d points for %s!", points, teamid))
}

func answerHandler(w http.ResponseWriter, req *http.Request) {
	teamid := req.FormValue("t")
	category := req.FormValue("c")
	pointstr := req.FormValue("p")
	answer := req.FormValue("a")

	points, err := strconv.Atoi(pointstr)
	if err != nil {
		points = 0
	}
	
	// Defang category name; prevent directory traversal
	if matched, _ := regexp.MatchString("^[A-Za-z0-9_-]", category); matched {
		category = ""
	}

	// Check answer
	needle := fmt.Sprintf("%s %s", points, answer)
	haystack := cachePath(category, "answers.txt")
	if ! anchoredSearch(haystack, needle, 0) {
		showPage(w, "Wrong answer", err.Error())
	}

	if err := awardPoints(teamid, category, points); err != nil {
		showPage(w, "Error awarding points", err.Error())
		return
	}
	showPage(w, "Points awarded", fmt.Sprintf("%d points for %s!", points, teamid))
}

// staticHandler serves up static files.
func rootHandler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/" {
		showPage(
			w,
			"Welcome",
			`
			  <h2>Register your team</h2>
			  
			  <form action="register" action="post">
			    Team ID: <input name="h"> <br>
			    Team name: <input name="n">
			    <input type="submit" value="Register">
			  </form>
			  
			  <div>
			    If someone on your team has already registered,
			    proceed to the
			    <a href="puzzles">puzzles overview</a>.
			  </div>
			`,
		)
		return
	}
	
	http.NotFound(w, req)
}
