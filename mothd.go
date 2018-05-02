package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var basePath = "."
var maintenanceInterval = 20 * time.Second
var categories = []string{}

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

func mothPath(filename string) string {
	return path.Join(basePath, filename)
}

func statePath(filename string) string {
	return path.Join(basePath, "state", filename)
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
