package main

import (
	"flag"
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

	var state_path string
	var state_engine_choice string
	var state_engine MOTHState

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
		&state_engine_choice,
		"state-engine",
		"legacy",
		"State engine to use (default: legacy, alt: sqlite)",
	)
	flag.StringVar(
		&state_path,
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


	if (state_engine_choice == "legacy") {
		lm_engine := &LegacyMOTHState{}
		lm_engine.StateDir = state_path
		lm_engine.maintenanceInterval = *maintenanceInterval
		state_engine = lm_engine
	} else {
		log.Fatal("Unrecognized state engine '", state_engine_choice, "'")
	}

	if err := setup(); err != nil {
		log.Fatal(err)
	}

	ctx.State = state_engine

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
