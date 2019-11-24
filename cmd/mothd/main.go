package main

import (
	"github.com/namsral/flag"
	"log"
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

	theme := NewTheme(*themePath)
	state := NewState(*statePath)
	puzzles := NewMothballs(*puzzlePath)

	go theme.Run(*refreshInterval)
	go state.Run(*refreshInterval)
	go puzzles.Run(*refreshInterval)

  log.Println("I would be binding to", *bindStr)
	time.Sleep(1 * time.Second)
	log.Print(state.Export(""))
	time.Sleep(19 * time.Second)
}
