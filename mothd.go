package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var moduleDir string
var stateDir string
var cacheDir string
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

func modulesPath(parts ...string) string {
	tail := path.Join(parts...)
	return path.Join(moduleDir, tail)
}

func statePath(parts ...string) string {
	tail := path.Join(parts...)
	return path.Join(stateDir, tail)
}

func cachePath(parts ...string) string {
	tail := path.Join(parts...)
	return path.Join(cacheDir, tail)
}

func setup() error {
	// Roll over and die if directories aren't even set up
	if _, err := os.Stat(modulesPath()); os.IsNotExist(err) {
		return err
	}
	if _, err := os.Stat(statePath()); os.IsNotExist(err) {
		return err
	}
	if _, err := os.Stat(cachePath()); os.IsNotExist(err) {
		return err
	}
	
	// Make sure points directories exist
	os.Mkdir(statePath("points.tmp"), 0755)
	os.Mkdir(statePath("points.new"), 0755)

	// Preseed available team ids if file doesn't exist
	if f, err := os.OpenFile(statePath("teamids.txt"), os.O_WRONLY | os.O_CREATE | os.O_EXCL, 0644); err == nil {
		defer f.Close()
		for i := 0; i <= 9999; i += 1 {
			fmt.Fprintf(f, "%04d\n", i)
		}
	}
	
	return nil
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	flag.StringVar(
		&moduleDir,
		"modules",
		"/modules",
		"Path where your moth modules live",
	)
	flag.StringVar(
		&stateDir,
		"state",
		"/state",
		"Path where state should be written",
	)
	flag.StringVar(
		&cacheDir,
		"cache",
		"/cache",
		"Path for ephemeral cache",
	)
	maintenanceInterval := flag.Duration(
		"maint",
		20 * time.Second,
		"Maintenance interval",
	)
	listen := flag.String(
		"listen",
		":8080",
		"[host]:port to bind and listen",
	)
	
	if err := setup(); err != nil {
		log.Fatal(err)
	}
	go maintenance(*maintenanceInterval)

	http.HandleFunc("/", rootHandler)
	http.Handle("/static/", http.FileServer(http.Dir(cacheDir)))

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/token", tokenHandler)
	http.HandleFunc("/answer", answerHandler)

	log.Printf("Listening on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, logRequest(http.DefaultServeMux)))
}
