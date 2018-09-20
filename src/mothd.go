package main

import (
	"flag"
	"log"
	"mime"
	"net/http"
	"time"
)

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
	resourcesDir := flag.String(
		"resources",
		"/moth/resources",
		"Path to static resources (HTML, images, css, ...)",
	)
	maintenanceInterval := flag.Duration(
		"maint",
		20*time.Second,
		"Maintenance interval",
	)
	listen := flag.String(
		"listen",
		":8080",
		"[host]:port to bind and listen",
	)
	flag.Parse()

	if err := setup(); err != nil {
		log.Fatal(err)
	}

	ctx, err := NewInstance(*base, *mothballDir, *stateDir, *resourcesDir)
	if err != nil {
		log.Fatal(err)
	}
	ctx.BindHandlers(http.DefaultServeMux)

	// Add some MIME extensions
	// Doing this avoids decompressing a mothball entry twice per request
	mime.AddExtensionType(".json", "application/json")
	mime.AddExtensionType(".zip", "application/zip")

	go ctx.Maintenance(*maintenanceInterval)

	log.Printf("Listening on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, logRequest(http.DefaultServeMux)))
}
