package main

import (
	"flag"
	"fmt"
	"log"
	"mime"
	"os"
	"strconv"
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

	stateEngine := flag.String(
		"state-engine",
		"legacy",
		"Specifiy a state engine",
	)

	redis_url := flag.String(
		"redis-url",
		"",
		"URL for Redis state instance",
	)

	redis_db := flag.Uint64(
		"redis-db",
		^uint64(0),
		"Database number for Redis state instance",
	)

	redis_instance_id := flag.String(
		"redis-instance-id",
		"",
		"Unique (per-cluster) instance ID",
	)

	flag.Parse()

	osfs := afero.NewOsFs()
	theme := NewTheme(afero.NewBasePathFs(osfs, *themePath))

	config := Configuration{}

	var provider PuzzleProvider
	provider = NewMothballs(afero.NewBasePathFs(osfs, *mothballPath))
	if *puzzlePath != "" {
		provider = NewTranspilerProvider(afero.NewBasePathFs(osfs, *puzzlePath))
		config.Devel = true
		log.Println("-=- You are in development mode, champ! -=-")
	}

	var state StateProvider

	switch engine := *stateEngine; engine {
	case "redis":
		log.Println("Moth is running with Redis")
		redis_url_parsed := *redis_url
		if redis_url_parsed == "" {
			redis_url_parsed = os.Getenv("REDIS_URL")
			if redis_url_parsed == "" {
				log.Fatal("Redis mode was selected, but --redis-url or REDIS_URL were not set")
			}
		}

		redis_db_parsed := *redis_db
		if redis_db_parsed == ^uint64(0) {
			redis_db_parsed_inner, err := strconv.ParseUint(os.Getenv("REDIS_DB"), 10, 64)
			redis_db_parsed = redis_db_parsed_inner

			if err != nil {
				redis_db_parsed = 0
			}
		}

		redis_instance_id_parsed := *redis_instance_id
		if redis_instance_id_parsed == "" {
			redis_instance_id_parsed = os.Getenv("REDIS_INSTANCE_ID")
			if redis_instance_id_parsed == "" {
				log.Fatal("Redis mode was selected, but --redis-instance-id or REDIS_INSTANCE_ID were not set")
			}
		}

		state = NewRedisState(redis_url_parsed, int(redis_db_parsed), redis_instance_id_parsed)
	default:
	case "legacy":
	    log.Println("Moth is running with the legacy state engine")
		state = NewState(afero.NewBasePathFs(osfs, *statePath))
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
