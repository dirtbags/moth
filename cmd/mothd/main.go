package main

import (
	"flag"
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"
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
	seed := flag.String(
		"seed",
		"",
		"Random seed to use, overrides $SEED",
	)
	flag.Parse()

	var theme *Theme
	osfs := afero.NewOsFs()
	if p, err := filepath.Abs(*themePath); err != nil {
		log.Fatal(err)
	} else {
		theme = NewTheme(afero.NewBasePathFs(osfs, p))
	}

	config := Configuration{}

	var provider PuzzleProvider
	if p, err := filepath.Abs(*mothballPath); err != nil {
		log.Fatal(err)
	} else {
		provider = NewMothballs(afero.NewBasePathFs(osfs, p))
	}
	if *puzzlePath != "" {
		if p, err := filepath.Abs(*puzzlePath); err != nil {
			log.Fatal(err)
		} else {
			provider = NewTranspilerProvider(afero.NewBasePathFs(osfs, p))
		}
		config.Devel = true
		log.Println("-=- You are in development mode, champ! -=-")
	}

	var state StateProvider
	if p, err := filepath.Abs(*statePath); err != nil {
		log.Fatal(err)
	} else {
		state = NewState(afero.NewBasePathFs(osfs, p))
	}
	if config.Devel {
		state = NewDevelState(state)
	}

	// Set random seed
	if *seed == "" {
		*seed = os.Getenv("SEED")
	}
	if *seed == "" {
		*seed = fmt.Sprintf("%d%d", os.Getpid(), time.Now().Unix())
	}
	os.Setenv("SEED", *seed)
	log.Print("SEED=", *seed)

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
