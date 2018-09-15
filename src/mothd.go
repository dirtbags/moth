package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

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


func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("HTTP %s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func setup() error {
	return nil
}

func main() {
  base := flag.String(
  	"base",
  	"/",
  	"Base URL of this instance",
  )
	mothballDir := flag.String(
		"mothballs",
		"/moth/mothballs",
		"Path to read mothballs",
	)
	stateDir := flag.String(
		"state",
		"/moth/state",
		"Path to write state",
	)
	maintenanceInterval := flag.Duration(
		"maint",
		20 * time.Second,
		"Maintenance interval",
	)
	listen := flag.String(
		"listen",
		":80",
		"[host]:port to bind and listen",
	)
	flag.Parse()
	
	if err := setup(); err != nil {
		log.Fatal(err)
	}

	ctx, err := NewInstance(*base, *mothballDir, *stateDir)
	if err != nil {
		log.Fatal(err)
	}
	ctx.BindHandlers(http.DefaultServeMux)

	go ctx.Maintenance(*maintenanceInterval)

	log.Printf("Listening on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, logRequest(http.DefaultServeMux)))
}
