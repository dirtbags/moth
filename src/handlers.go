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
	"os/exec"
	"path"
)

type JSend struct {
	Status string    `json:"status"`
	Data   JSendData `json:"data"`
}
type JSendData struct {
	Short       string `json:"short"`
	Description string `json:"description"`
}

// ShowJSend renders a JSend response to w
func ShowJSend(w http.ResponseWriter, status Status, short string, description string) {

	resp := JSend{
		Status: "success",
		Data: JSendData{
			Short:       short,
			Description: description,
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

type Status int

const (
	Success = iota
	Fail
	Error
)

// ShowHtml delevers an HTML response to w
func ShowHtml(w http.ResponseWriter, status Status, title string, body string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	statusStr := ""
	switch status {
	case Success:
		statusStr = "Success"
	case Fail:
		statusStr = "Fail"
	default:
		statusStr = "Error"
	}

	fmt.Fprintf(w, "<!DOCTYPE html>")
	fmt.Fprintf(w, "<!-- If you put `application/json` in the `Accept` header of this request, you would have gotten a JSON object instead of HTML. -->\n")
	fmt.Fprintf(w, "<html><head>")
	fmt.Fprintf(w, "<title>%s</title>", title)
	fmt.Fprintf(w, "<link rel=\"stylesheet\" href=\"basic.css\">")
	fmt.Fprintf(w, "<meta name=\"viewport\" content=\"width=device-width\">")
	fmt.Fprintf(w, "<link rel=\"icon\" href=\"res/icon.svg\" type=\"image/svg+xml\">")
	fmt.Fprintf(w, "<link rel=\"icon\" href=\"res/icon.png\" type=\"image/png\">")
	fmt.Fprintf(w, "</head><body><h1 class=\"%s\">%s</h1>", statusStr, title)
	fmt.Fprintf(w, "<section>%s</section>", body)
	fmt.Fprintf(w, "<nav>")
	fmt.Fprintf(w, "<ul>")
	fmt.Fprintf(w, "<li><a href=\"puzzle-list.html\">Puzzles</a></li>")
	fmt.Fprintf(w, "<li><a href=\"scoreboard.html\">Scoreboard</a></li>")
	fmt.Fprintf(w, "</ul>")
	fmt.Fprintf(w, "</nav>")
	fmt.Fprintf(w, "</body></html>")
}

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

// hasSubstr returns the line where substr appears in r.
// The line must contain substr.
func hasSubstr(r io.Reader, substr string) string {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), substr) {
			return scanner.Text()
		}
	}
	return ""
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
	foundAns := false
	if err != nil {
		// We did not find an answer file, but we can still look for a dynamic answer
	} else {
		// Look for the answer
		needle := fmt.Sprintf("%d %s", points, answer)
		if !hasLine(haystack, needle) {
			// We did not find the answer, but we can still look for a dynamic answer
		} else {
			foundAns = true
		}
	}
	defer haystack.Close()

	
	if !foundAns {
		// Now we look for a dynamic answer, since no static one matched
		haystackdyn, errdyn := ctx.OpenCategoryFile(category, "answerdyn.txt")
		if errdyn != nil && err != nil {
			respond(
				w, req, Fail,
				"Cannot list answers",
				"Unable to read the list of static or dynamic answers for this category.",
			)
			return
		}
		defer haystackdyn.Close()

		// Look for the answerdyn file
		needledyn := fmt.Sprintf("%d", points)
		answerFile := hasSubstr(haystackdyn, needledyn)
		if answerFile == "" {
			// This is where the answer file is run, check to make sure needledyn is the full line
			// If this code is reached, neither answers nor answersdyn has an entry for the submission.
			respond(
				w, req, Fail,
				"Wrong answer",
				"That is not the correct answer for %s %d.", category, points,
			)
			return
		} else {
			//If this point is reached, then we have a dynamic grader to run
			splitAnswer := strings.Split(answerFile + " " + answer, " ")
			recombinedCommand := splitAnswer[1]
			splitAnswer = splitAnswer[2:]
			cmd := exec.Command(recombinedCommand, splitAnswer...)
			
			// Now we read the map to get the correct answer file directory
			haystackmap, errmap := ctx.OpenCategoryFile(category, "map.txt")
			if errmap != nil && err != nil {
				respond(
					w, req, Fail,
					"Cannot read map",
					"Unable to read the map for the category.",
				)
				return
			}
			defer haystackmap.Close()
			mappedDir := hasSubstr(haystackmap, needledyn)
			splitMap := strings.Split(mappedDir, " ")
			gradeDir, direrr := ctx.GetCategoryDir(category)
			cmd.Dir = path.Join(gradeDir, "answerdyn", splitMap[1])
			
			// Now we run the dynamic grader command with the answer as the first argument
			out, cmderr := cmd.CombinedOutput()
			theOutput := string(out)
			theOutput = strings.TrimRight(theOutput, "\n")
			if cmderr != nil || direrr != nil || theOutput != "true" {
				// The answer script must return "true" on stdout to be correct
				respond(
					w, req, Fail,
					"Wrong answer",
					"That is not the correct answer for %s %d.", category, points,
				)
				return
			}
		}
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

func (ctx *Instance) BindHandlers(mux *http.ServeMux) {
	mux.HandleFunc(ctx.Base+"/", ctx.staticHandler)
	mux.HandleFunc(ctx.Base+"/register", ctx.registerHandler)
	mux.HandleFunc(ctx.Base+"/answer", ctx.answerHandler)
	mux.HandleFunc(ctx.Base+"/content/", ctx.contentHandler)
	mux.HandleFunc(ctx.Base+"/puzzles.json", ctx.puzzlesHandler)
	mux.HandleFunc(ctx.Base+"/points.json", ctx.pointsHandler)
}
