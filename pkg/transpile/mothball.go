package transpile

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
)

// Mothball packages a Category up for a production server run.
func Mothball(c Category) (*bytes.Reader, error) {
	buf := new(bytes.Buffer)
	zf := zip.NewWriter(buf)

	inv, err := c.Inventory()
	if err != nil {
		return nil, err
	}

	puzzlesTxt, err := zf.Create("puzzles.txt")
	if err != nil {
		return nil, err
	}
	answersTxt, err := zf.Create("answers.txt")
	if err != nil {
		return nil, err
	}

	for _, points := range inv {
		fmt.Fprintln(puzzlesTxt, points)

		puzzlePath := fmt.Sprintf("%d/puzzle.json", points)
		pw, err := zf.Create(puzzlePath)
		if err != nil {
			return nil, err
		}
		puzzle, err := c.Puzzle(points)
		if err != nil {
			return nil, fmt.Errorf("Puzzle %d: %s", points, err)
		}

		// Record answers in answers.txt
		for _, answer := range puzzle.Answers {
			fmt.Fprintln(answersTxt, points, answer)
		}

		// Remove answers and debugging from puzzle object
		puzzle.Answers = []string{}
		puzzle.Debug.Errors = []string{}
		puzzle.Debug.Hints = []string{}
		puzzle.Debug.Log = []string{}

		// Write out Puzzle object
		penc := json.NewEncoder(pw)
		if err := penc.Encode(puzzle); err != nil {
			return nil, fmt.Errorf("Puzzle %d: %s", points, err)
		}

		// Write out all attachments and scripts
		attachments := append(puzzle.Pre.Attachments, puzzle.Pre.Scripts...)
		for _, att := range attachments {
			attPath := fmt.Sprintf("%d/%s", points, att)
			aw, err := zf.Create(attPath)
			if err != nil {
				return nil, err
			}
			ar, err := c.Open(points, att)
			if exerr, ok := err.(*exec.ExitError); ok {
				return nil, fmt.Errorf("Puzzle %d: %s: %s: %s", points, att, err, string(exerr.Stderr))
			} else if err != nil {
				return nil, fmt.Errorf("Puzzle %d: %s: %s", points, att, err)
			}
			if _, err := io.Copy(aw, ar); err != nil {
				return nil, fmt.Errorf("Puzzle %d: %s: %s", points, att, err)
			}
		}
	}
	zf.Close()

	return bytes.NewReader(buf.Bytes()), nil
}
