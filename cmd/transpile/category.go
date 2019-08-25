package main

import (
	"os"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"hash/fnv"
	"encoding/binary"
	"encoding/json"
	"encoding/hex"
	"strconv"
	"math/rand"
	"context"
	"time"
	"os/exec"
	"bytes"
)


type PuzzleEntry struct {
	Id string
	Points int
	Puzzle Puzzle
}

func PrngOfStrings(input ...string) (*rand.Rand) {
	hasher := fnv.New64()
	for _, s := range input {
		fmt.Fprint(hasher, s, "\n")
	}
	seed := binary.BigEndian.Uint64(hasher.Sum(nil))
	source := rand.NewSource(int64(seed))
	return rand.New(source)
}


func runPuzzleGen(puzzlePath string, seed string) (*Puzzle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, puzzlePath)
	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("MOTH_PUZZLE_SEED=%s", seed),
	)
	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	jsdec := json.NewDecoder(bytes.NewReader(stdout))
	jsdec.DisallowUnknownFields()
	puzzle := new(Puzzle)
	err = jsdec.Decode(puzzle)
	if err != nil {
		return nil, err
	}
	
	return puzzle, nil
}


func ParseCategory(categoryPath string, seed string) ([]PuzzleEntry, error) {
	categoryFd, err := os.Open(categoryPath)
	if err != nil {
		return nil, err
	}
	defer categoryFd.Close()
	
	puzzleDirs, err := categoryFd.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	
	puzzleEntries := make([]PuzzleEntry, 0, len(puzzleDirs))
	for _, puzzleDir := range puzzleDirs {
		var puzzle *Puzzle
		
		puzzlePath := filepath.Join(categoryPath, puzzleDir)
		
		// Determine point value from directory name
		points, err := strconv.Atoi(puzzleDir)
		if err != nil {
			log.Printf("Skipping %s: %v", puzzlePath, err)
			continue
		}

		// Try the .moth file first
		puzzleMothPath := filepath.Join(puzzlePath, "puzzle.moth")
		puzzleFd, err := os.Open(puzzleMothPath)
		if err == nil {
			defer puzzleFd.Close()
			puzzle, err = ParseMoth(puzzleFd)
			if err != nil {
				log.Printf("Skipping %s: %v", puzzleMothPath, err)
				continue
			}
		} else if os.IsNotExist(err) {
			var genErr error
			puzzleGenPath := filepath.Join(puzzlePath, "mkpuzzle")
			puzzle, genErr = runPuzzleGen(puzzleGenPath, puzzlePath)
			if genErr != nil {
				log.Printf("Skipping %20s: %v", puzzleMothPath, err)
				log.Printf("Skipping %20s: %v", puzzleGenPath, genErr)
				continue
			}
		} else {
			log.Printf("Skipping %s: %v", puzzleMothPath, err)
			continue
		}
		
		// Create a category entry for this
		prng := PrngOfStrings(puzzlePath)
		idBytes := make([]byte, 16)
		prng.Read(idBytes)
		id := hex.EncodeToString(idBytes)
		puzzleEntry := PuzzleEntry{
			Id: id,
			Puzzle: *puzzle,
			Points: points,
		}
		puzzleEntries = append(puzzleEntries, puzzleEntry)
	}
	
	return puzzleEntries, nil
}


func main() {
	// XXX: We need a way to pass in "only run this one point value puzzle"
	// XXX: Convert puzzle.py to standalone thingies
	flag.Parse()
	baseSeedString := os.Getenv("SEED")
	
	jsenc := json.NewEncoder(os.Stdout)
	jsenc.SetEscapeHTML(false)
	jsenc.SetIndent("", "  ")

	for _, dirname := range flag.Args() {
		categoryName := filepath.Base(dirname)
		categorySeed := fmt.Sprintf("%s/%s", baseSeedString, categoryName)
		puzzles, err := ParseCategory(dirname, categorySeed)
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
