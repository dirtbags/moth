package main

import (
	"github.com/namsral/flag"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"time"
)

func setup() error {
	rand.Seed(time.Now().UnixNano())
	return nil
}

func main() {
	ctx := &Instance{}

	flag.StringVar(
		&ctx.Base,
		"base",
		"/",
		"Base URL of this instance",
	)
	flag.StringVar(
		&ctx.MothballDir,
		"mothballs",
		"/mothballs",
		"Path to read mothballs",
	)
	flag.StringVar(
		&ctx.StateDir,
		"state",
		"/state",
		"Path to write state",
	)
	flag.StringVar(
		&ctx.ThemeDir,
		"theme",
		"/theme",
		"Path to static theme resources (HTML, images, css, ...)",
	)
	flag.StringVar(
		&ctx.ExtractedMothballDir,
		"extractedMothballs",
		"/extractedMothballs",
		"Path to where mothballs are extracted to at runtime.",
	)
	flag.StringVar(
		&ctx.Progression,
		"progression",
		"all",
		"The type of challenge progression (team | all | dev)",
	)
	flag.DurationVar(
		&ctx.AttemptInterval,
		"attempt",
		500*time.Millisecond,
		"Per-team time required between answer attempts",
	)
	maintenanceInterval := flag.Duration(
		"maint",
		20*time.Second,
		"Time between maintenance tasks",
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

	err := ctx.Initialize()
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
