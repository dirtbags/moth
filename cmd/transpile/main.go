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
	fmt.Fprintf(out, "Usage: %s [OPTION]... CATEGORY [PUZZLE [FILENAME]]\n", name)
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "Transpile CATEGORY, or provide individual category components.\n")
	fmt.Fprintf(out, "If PUZZLE is provided, only transpile the given puzzle.\n")
	fmt.Fprintf(out, "If FILENAME is provided, output provided file.\n")
	flag.PrintDefaults()
}

func main() {
	// XXX: Convert puzzle.py to standalone thingies
	
	flag.Usage = usage
	
	points := flag.Int("points", 0, "Transpile only this point value puzzle")
	mothball := flag.Bool("mothball", false, "Generate a mothball")
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
