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
func Mothball(c Category, w io.Writer) error {
	zf := zip.NewWriter(w)

	inv, err := c.Inventory()
	if err != nil {
		return err
	}

	puzzlesTxt := new(bytes.Buffer)
	answersTxt := new(bytes.Buffer)

	for _, points := range inv {
		fmt.Fprintln(puzzlesTxt, points)

		puzzlePath := fmt.Sprintf("%d/puzzle.json", points)
		pw, err := zf.Create(puzzlePath)
		if err != nil {
			return err
		}
		puzzle, err := c.Puzzle(points)
		if err != nil {
			return fmt.Errorf("Puzzle %d: %s", points, err)
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
			return fmt.Errorf("Puzzle %d: %s", points, err)
		}

		// Write out all attachments and scripts
		attachments := append(puzzle.Pre.Attachments, puzzle.Pre.Scripts...)
		for _, att := range attachments {
			attPath := fmt.Sprintf("%d/%s", points, att)
			aw, err := zf.Create(attPath)
			if err != nil {
				return err
			}
			ar, err := c.Open(points, att)
			if exerr, ok := err.(*exec.ExitError); ok {
				return fmt.Errorf("Puzzle %d: %s: %s: %s", points, att, err, string(exerr.Stderr))
			} else if err != nil {
				return fmt.Errorf("Puzzle %d: %s: %s", points, att, err)
			}
			if _, err := io.Copy(aw, ar); err != nil {
				return fmt.Errorf("Puzzle %d: %s: %s", points, att, err)
			}
		}
	}

	pf, err := zf.Create("puzzles.txt")
	if err != nil {
		return err
	}
	puzzlesTxt.WriteTo(pf)

	af, err := zf.Create("answers.txt")
	if err != nil {
		return err
	}
	answersTxt.WriteTo(af)

	zf.Close()

	return nil
}
