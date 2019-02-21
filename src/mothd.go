package main

import (
	"github.com/namsral/flag"
	"log"
	"mime"
	"net/http"
	"time"
)


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
		"/mothballs",
		"Path to read mothballs",
	)
	stateDir := flag.String(
		"state",
		"/state",
		"Path to write state",
	)
	themeDir := flag.String(
		"theme",
		"/theme",
		"Path to static theme resources (HTML, images, css, ...)",
	)
	password := flag.String(
		"password",
		"sesame",
		"Pass Word (in the 1920s sense) to view the site. Not a secure passphrase.",
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

	ctx, err := NewInstance(*base, *mothballDir, *stateDir, *themeDir, *password)
	if err != nil {
		log.Fatal(err)
	}

	// Add some MIME extensions
	// Doing this avoids decompressing a mothball entry twice per request
	mime.AddExtensionType(".json", "application/json")
	mime.AddExtensionType(".zip", "application/zip")

	go ctx.Maintenance(*maintenanceInterval)

	log.Printf("Listening on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, ctx))
}
