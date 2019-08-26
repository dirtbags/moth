package main

import (
	"flag"
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
	"os"
	"log"
	"fmt"
)

func seedJoin(parts ...string) string {
	return strings.Join(parts, "::")
}

func usage() {
	out := flag.CommandLine.Output()
	name := flag.CommandLine.Name()
	fmt.Fprintf(out, "Usage: %s [OPTIONS] CATEGORY [CATEGORY ...]\n", name)
	flag.PrintDefaults()
}

func main() {
	// XXX: We need a way to pass in "only run this one point value puzzle"
	// XXX: Convert puzzle.py to standalone thingies
	
	flag.Usage = usage
	points := flag.Int("points", 0, "Transpile only this point value puzzle")
	flag.Parse()
	
	baseSeedString := os.Getenv("MOTH_SEED")
	
	jsenc := json.NewEncoder(os.Stdout)
	jsenc.SetEscapeHTML(false)
	jsenc.SetIndent("", "  ")

	for _, categoryPath := range flag.Args() {
		categoryName := filepath.Base(categoryPath)
		categorySeed := seedJoin(baseSeedString, categoryName)

		if *points > 0 {
			puzzleDir := strconv.Itoa(*points)
			puzzleSeed := seedJoin(categorySeed, puzzleDir)
			puzzlePath := filepath.Join(categoryPath, puzzleDir)
			puzzle, err := ParsePuzzle(puzzlePath, puzzleSeed)
			if err != nil {
				log.Print(err)
				continue
			}
			
			if err := jsenc.Encode(puzzle); err != nil {
				log.Fatal(err)
			}
		} else {
			puzzles, err := ParseCategory(categoryPath, categorySeed)
			if err != nil {
				log.Print(err)
				continue
			}
			
			if err := jsenc.Encode(puzzles); err != nil {
				log.Print(err)
				continue
			}
		}
	}
}
