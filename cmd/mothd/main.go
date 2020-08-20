package main

import (
	"flag"
	"mime"
	"time"

	"github.com/spf13/afero"
)

func main() {
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
	mothballPath := flag.String(
		"mothballs",
		"mothballs",
		"Path to mothball files",
	)
	refreshInterval := flag.Duration(
		"refresh",
		2*time.Second,
		"Duration between maintenance tasks",
	)
	bindStr := flag.String(
		"bind",
		":8080",
		"Bind [host]:port for HTTP service",
	)
	base := flag.String(
		"base",
		"/",
		"Base URL of this instance",
	)
	flag.Parse()

	theme := NewTheme(afero.NewBasePathFs(afero.NewOsFs(), *themePath))
	state := NewState(afero.NewBasePathFs(afero.NewOsFs(), *statePath))
	puzzles := NewMothballs(afero.NewBasePathFs(afero.NewOsFs(), *mothballPath))

	// Add some MIME extensions
	// Doing this avoids decompressing a mothball entry twice per request
	mime.AddExtensionType(".json", "application/json")
	mime.AddExtensionType(".zip", "application/zip")

	go theme.Maintain(*refreshInterval)
	go state.Maintain(*refreshInterval)
	go puzzles.Maintain(*refreshInterval)

	server := NewMothServer(puzzles, theme, state)
	httpd := NewHTTPServer(*base, server)

	httpd.Run(*bindStr)
}
