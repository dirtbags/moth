package transpile

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
			return nil, err
		}

		// Record answers in answers.txt
		for _, answer := range puzzle.Answers {
			fmt.Fprintln(answersTxt, points, answer)
		}

		// Remove all answers from puzzle object
		puzzle.Answers = []string{}

		// Write out Puzzle object
		penc := json.NewEncoder(pw)
		if err := penc.Encode(puzzle); err != nil {
			return nil, err
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
			if err != nil {
				return nil, err
			}
			if _, err := io.Copy(aw, ar); err != nil {
				return nil, err
			}
		}
	}
	zf.Close()

	return bytes.NewReader(buf.Bytes()), nil
}
