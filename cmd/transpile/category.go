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


func ParsePuzzle(puzzlePath string, seed string) (*Puzzle, error) {
	puzzleFd, err := os.Open(puzzlePath)
	if err != nil {
		return nil, err
	}
	defer puzzleFd.Close()
	
	puzzle, err := ParseMoth(puzzleFd)
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
		puzzlePath := filepath.Join(categoryPath, puzzleDir, "puzzle.moth")
		puzzleSeed := fmt.Sprintf("%s/%s", seed, puzzleDir)
		
		points, err := strconv.Atoi(puzzleDir)
		if err != nil {
			log.Printf("Skipping %s: %v", puzzlePath, err)
			continue
		}
	
		puzzle, err := ParsePuzzle(puzzlePath, puzzleSeed)
		if err != nil {
			log.Printf("Skipping %s: %v", puzzlePath, err)
			continue
		}
		
		prng := PrngOfStrings(puzzleSeed)
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
	
	for _, dirname := range flag.Args() {
		categoryName := filepath.Base(dirname)
		categorySeed := fmt.Sprintf("%s/%s", baseSeedString, categoryName)
		puzzles, err := ParseCategory(dirname, categorySeed)
		if err != nil {
			log.Print(err)
			continue
		}
		
		jpuzzles, err := json.MarshalIndent(puzzles, "", "  ")
		if err != nil {
			log.Print(err)
			continue
		}
		
		fmt.Println(string(jpuzzles))
	}
}
