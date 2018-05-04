package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

var basePath = "/home/neale/src/moth"
var maintenanceInterval = 20 * time.Second
var categories = []string{}

// anchoredSearch looks for needle in filename,
// skipping the first skip space-delimited words
func anchoredSearch(filename string, needle string, skip int) bool {
	f, err := os.Open(filename)
	if err != nil {
		log.Print("Can't open %s: %s", filename, err)
		return false
	}
	defer f.Close()
	
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(" ", line, skip+1)
		if parts[skip+1] == needle {
			return true
		}
	}

	return false
}


func awardPoints(teamid string, category string, points int) error {
	fn := fmt.Sprintf("%s-%s-%d", teamid, category, points)
	tmpfn := statePath("points.tmp", fn)
	newfn := statePath("points.new", fn)
	
	contents := fmt.Sprintf("%d %s %s %d\n", time.Now().Unix(), teamid, points)
	
	if err := ioutil.WriteFile(tmpfn, []byte(contents), 0644); err != nil {
		return err
	}
	
	if err := os.Rename(tmpfn, newfn); err != nil {
		return err
	}
	
	return nil
}

func showPage(w http.ResponseWriter, title string, body string) {
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "<!DOCTYPE html>")
	fmt.Fprintf(w, "<html><head>")
	fmt.Fprintf(w, "<title>%s</title>", title)
	fmt.Fprintf(w, "<link rel=\"stylesheet\" href=\"../style.css\">")
	fmt.Fprintf(w, "<meta name=\"viewport\" content=\"width=device-width\"></head>")
	fmt.Fprintf(w, "<body><h1>%s</h1>", title)
	fmt.Fprintf(w, "<section>%s</section>", body)
	fmt.Fprintf(w, "<nav>")
	fmt.Fprintf(w, "<ul>")
	fmt.Fprintf(w, "<li><a href=\"../register.html\">Register</a></li>")
	fmt.Fprintf(w, "<li><a href=\"../puzzles.html\">Puzzles</a></li>")
	fmt.Fprintf(w, "<li><a href=\"../scoreboard.html\">Scoreboard</a></li>")
	fmt.Fprintf(w, "</ul>")
	fmt.Fprintf(w, "</nav>")
	fmt.Fprintf(w, "</body></html>")
}

func mothPath(parts ...string) string {
	tail := path.Join(parts...)
	return path.Join(basePath, tail)
}

func statePath(parts ...string) string {
	tail := path.Join(parts...)
	return path.Join(basePath, "state", tail)
}

func exists(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false;
	}
	return true;
}

func main() {
	log.Print("Sup")
	go maintenance();
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/token", tokenHandler)
	http.HandleFunc("/answer", answerHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// docker run --rm -it -p 5880:8080 -v $HOME:$HOME:ro -w $(pwd) golang go run mothd.go
