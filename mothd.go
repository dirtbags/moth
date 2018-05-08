package main

import (
	"bufio"
	"github.com/namsral/flag"
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
	fmt.Fprintf(w, "<link rel=\"stylesheet\" href=\"static/style.css\">")
	fmt.Fprintf(w, "<meta name=\"viewport\" content=\"width=device-width\"></head>")
	fmt.Fprintf(w, "<link rel=\"icon\" href=\"res/luna-moth.svg\" type=\"image/svg+xml\">")
  fmt.Fprintf(w, "<link rel=\"icon\" href=\"res/luna-moth.png\" type=\"image/png\">")
	fmt.Fprintf(w, "<body><h1>%s</h1>", title)
	fmt.Fprintf(w, "<section>%s</section>", body)
	fmt.Fprintf(w, "<nav>")
	fmt.Fprintf(w, "<ul>")
	fmt.Fprintf(w, "<li><a href=\"static/puzzles.html\">Puzzles</a></li>")
	fmt.Fprintf(w, "<li><a href=\"static/scoreboard.html\">Scoreboard</a></li>")
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

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("HTTP %s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
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

func main() {
	var maintenanceInterval time.Duration
	var listen string
	
	fs := flag.NewFlagSetWithEnvPrefix(os.Args[0], "MOTH", flag.ExitOnError)
	fs.StringVar(
		&moduleDir,
		"modules",
		"/moth/modules",
		"Path where your moth modules live",
	)
	fs.StringVar(
		&stateDir,
		"state",
		"/moth/state",
		"Path where state should be written",
	)
	fs.StringVar(
		&cacheDir,
		"cache",
		"/moth/cache",
		"Path for ephemeral cache",
	)
	fs.DurationVar(
		&maintenanceInterval,
		"maint",
		20 * time.Second,
		"Maintenance interval",
	)
	fs.StringVar(
		&listen,
		"listen",
		":8080",
		"[host]:port to bind and listen",
	)
	fs.Parse(os.Args[1:])
	
	if err := setup(); err != nil {
		log.Fatal(err)
	}
	go maintenance(maintenanceInterval)

	fileserver := http.FileServer(http.Dir(cacheDir))
	http.HandleFunc("/", rootHandler)
	http.Handle("/static/", http.StripPrefix("/static", fileserver))

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/token", tokenHandler)
	http.HandleFunc("/answer", answerHandler)
	
	http.HandleFunc("/puzzles.json", puzzlesHandler)
	http.HandleFunc("/points.json", pointsHandler)

	log.Printf("Listening on %s", listen)
	log.Fatal(http.ListenAndServe(listen, logRequest(http.DefaultServeMux)))
}
