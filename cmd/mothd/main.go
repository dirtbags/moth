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
	puzzlePath := flag.String(
		"puzzles",
		"",
		"Path to puzzles tree (enables development mode)",
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

	osfs := afero.NewOsFs()
	theme := NewTheme(afero.NewBasePathFs(osfs, *themePath))
	state := NewState(afero.NewBasePathFs(osfs, *statePath))

	config := Configuration{}

	var provider PuzzleProvider
	provider = NewMothballs(afero.NewBasePathFs(osfs, *mothballPath))
	if *puzzlePath != "" {
		provider = NewTranspilerProvider(afero.NewBasePathFs(osfs, *puzzlePath))
		config.Devel = true
	}

	// Add some MIME extensions
	// Doing this avoids decompressing a mothball entry twice per request
	mime.AddExtensionType(".json", "application/json")
	mime.AddExtensionType(".zip", "application/zip")

	go theme.Maintain(*refreshInterval)
	go state.Maintain(*refreshInterval)
	go provider.Maintain(*refreshInterval)

	server := NewMothServer(config, theme, state, provider)
	httpd := NewHTTPServer(*base, server)

	httpd.Run(*bindStr)
}
