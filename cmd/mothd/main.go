package main

import (
	"github.com/namsral/flag"
	"github.com/spf13/afero"
	"log"
	"mime"
	"net/http"
	"time"
)

func main() {
	log.Print("Started")

	themePath := flag.String(
		"theme",
		"theme",
		"Path to theme files",
	)
	statePath := flag.String(
		"state",
		"state",
		"Path to state files",
	)
	puzzlePath := flag.String(
		"mothballs",
		"mothballs",
		"Path to mothballs to host",
	)
	refreshInterval := flag.Duration(
		"refresh",
		2*time.Second,
		"Duration between maintenance tasks",
	)
	bindStr := flag.String(
		"bind",
		":8000",
		"Bind [host]:port for HTTP service",
	)

	stateFs := afero.NewBasePathFs(afero.NewOsFs(), *statePath)
	themeFs := afero.NewBasePathFs(afero.NewOsFs(), *themePath)
	mothballFs := afero.NewBasePathFs(afero.NewOsFs(), *mothballPath)

	theme := NewTheme(themeFs)
	state := NewState(stateFs)
	puzzles := NewMothballs(mothballFs)

	go state.Run(*refreshInterval)
	go puzzles.Run(*refreshInterval)

	// Add some MIME extensions
	// Doing this avoids decompressing a mothball entry twice per request
	mime.AddExtensionType(".json", "application/json")
	mime.AddExtensionType(".zip", "application/zip")

	http.HandleFunc("/", theme.staticHandler)

	log.Printf("Listening on %s", *bindStr)
	log.Fatal(http.ListenAndServe(*bindStr, nil))
}
