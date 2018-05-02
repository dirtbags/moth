package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var basePath = "."
var maintenanceInterval = 20 * time.Second
var categories = []string{}

type Award struct {
	when time.Time,
	team string,
	category string,
	points int,
	comment string
}

func ParseAward(s string) (*Award, error) {
	ret := Award{}
	
	parts := strings.SplitN(s, " ", 5)
	if len(parts) < 4 {
		return nil, Error("Malformed award string")
	}
	
	whenEpoch, err = strconv.Atoi(parts[0])
	if (err != nil) {
		return nil, Errorf("Malformed timestamp: %s", parts[0])
	}
	ret.when = time.Unix(whenEpoch, 0)
	
	ret.team = parts[1]
	ret.category = parts[2]
	
	points, err = strconv.Atoi(parts[3])
	if (err != nil) {
		return nil, Errorf("Malformed points: %s", parts[3])
	}
	
	if len(parts) == 5 {
		ret.comment = parts[4]
	}
	
	return &ret
}

func (a *Award) String() string {
	return fmt.Sprintf("%d %s %s %d %s", a.when.Unix(), a.team, a.category, a.points, a.comment)
}

func mooHandler(w http.ResponseWriter, req *http.Request) {
	moo := req.FormValue("moo")
	fmt.Fprintf(w, "Hello, %q. %s", html.EscapeString(req.URL.Path), html.EscapeString(moo))
}

func rootHandler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
}

func mothPath(parts ...string) string {
	return path.Join(basePath, parts...)
}

func statePath(parts ...string) string {
	return path.Join(basePath, "state", parts...)
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
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/moo/", mooHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// docker run --rm -it -p 5880:8080 -v $HOME:$HOME:ro -w $(pwd) golang go run mothd.go
