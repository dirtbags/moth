package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"strconv"
	"io"
	"log"
	"bufio"
)

// anchoredSearch looks for needle in r,
// skipping the first skip space-delimited words
func anchoredSearch(r io.Reader, needle string, skip int) bool {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", skip+1)
		if (len(parts) > skip) && (parts[skip] == needle) {
			return true
		}
	}

	return false
}

func anchoredSearchFile(filename string, needle string, skip int) bool {
	r, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer r.Close()
	
	return anchoredSearch(r, needle, skip)
}


func (ctx Instance) registerHandler(w http.ResponseWriter, req *http.Request) {
	teamname := req.FormValue("n")
	teamid := req.FormValue("h")
	
	if matched, _ := regexp.MatchString("[^0-9a-z]", teamid); matched {
		teamid = ""
	}
	
	if (teamid == "") || (teamname == "") {
		showPage(w, "Invalid Entry", "Oops! Are you sure you got that right?")
		return
	}
	
	if ! anchoredSearchFile(ctx.StatePath("teamids.txt"), teamid, 0) {
		showPage(w, "Invalid Team ID", "I don't have a record of that team ID. Maybe you used capital letters accidentally?")
		return
	}
	
	f, err := os.OpenFile(ctx.StatePath(teamid), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
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

func (ctx Instance) tokenHandler(w http.ResponseWriter, req *http.Request) {
	teamid := req.FormValue("t")
	token := req.FormValue("k")

	// Check answer
	if ! anchoredSearchFile(ctx.StatePath("tokens.txt"), token, 0) {
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
	
	if err := ctx.AwardPoints(teamid, category, points); err != nil {
		showPage(w, "Error awarding points", err.Error())
		return
	}
	showPage(w, "Points awarded", fmt.Sprintf("%d points for %s!", points, teamid))
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
	if ! ok {
		showPage(w, "Category does not exist", "The specified category does not exist. Sorry!")
		return
	}

	// Get the answers
	haystack, err := catmb.Open("answers.txt")
	if err != nil {
		showPage(w, "Answers do not exist",
			"Please tell the contest people that the mothball for this category has no answers.txt in it!")
		return
	}
	defer haystack.Close()
	
	// Look for the answer
	needle := fmt.Sprintf("%d %s", points, answer)
	if ! anchoredSearch(haystack, needle, 0) {
		showPage(w, "Wrong answer", err.Error())
		return
	}

	if err := ctx.AwardPoints(teamid, category, points); err != nil {
		showPage(w, "Error awarding points", err.Error())
		return
	}
	showPage(w, "Points awarded", fmt.Sprintf("%d points for %s!", points, teamid))
}

func (ctx Instance) puzzlesHandler(w http.ResponseWriter, req *http.Request) {
	puzzles := map[string][]interface{}{}
	// 	v := map[string][]interface{}{"Moo": {1, "0177f85ae895a33e2e7c5030c3dc484e8173e55c"}}
  // j, _ := json.Marshal(v)
	
	for _, category := range ctx.Categories {
		log.Print(puzzles, category)
	}
}

func (ctx Instance) pointsHandler(w http.ResponseWriter, req *http.Request) {
	
}

// staticHandler serves up static files.
func (ctx Instance) rootHandler(w http.ResponseWriter, req *http.Request) {
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

func (ctx Instance) BindHandlers(mux *http.ServeMux) {
	mux.HandleFunc(ctx.Base + "/", ctx.rootHandler)
	mux.HandleFunc(ctx.Base + "/register", ctx.registerHandler)
	mux.HandleFunc(ctx.Base + "/token", ctx.tokenHandler)
	mux.HandleFunc(ctx.Base + "/answer", ctx.answerHandler)
	mux.HandleFunc(ctx.Base + "/puzzles.json", ctx.puzzlesHandler)
	mux.HandleFunc(ctx.Base + "/points.json", ctx.pointsHandler)
}
