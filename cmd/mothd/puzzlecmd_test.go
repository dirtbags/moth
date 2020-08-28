package main

import (
	"io/ioutil"
	"os/exec"
	"testing"
)

func TestPuzzleCommand(t *testing.T) {
	pc := PuzzleCommand{
		Path: "testdata/testpiler.sh",
	}

	inv := pc.Inventory()
	if len(inv) != 2 {
		t.Errorf("Wrong length for inventory")
	}
	for _, cat := range inv {
		switch cat.Name {
		case "pategory":
			if len(cat.Puzzles) != 8 {
				t.Errorf("pategory wrong number of puzzles: %d", len(cat.Puzzles))
			}
			if cat.Puzzles[5] != 10 {
				t.Errorf("pategory puzzles[5] wrong value: %d", cat.Puzzles[5])
			}
		case "nealegory":
			if len(cat.Puzzles) != 3 {
				t.Errorf("nealegoy wrong number of puzzles: %d", len(cat.Puzzles))
			}
			if cat.Puzzles[2] != 3 {
				t.Errorf("out of order point values were not sorted")
			}
		}
	}

	if err := pc.CheckAnswer("pategory", 1, "answer"); err != nil {
		t.Errorf("Correct answer for pategory: %v", err)
	}
	if err := pc.CheckAnswer("pategory", 1, "wrong"); err == nil {
		t.Errorf("Wrong answer for pategory judged correct")
	}

	if err := pc.CheckAnswer("pategory", 2, "answer"); err == nil {
		t.Errorf("Internal error not returned")
	} else if ee, ok := err.(*exec.ExitError); ok {
		if string(ee.Stderr) != "Internal error\n" {
			t.Errorf("Unexpected error returned: %#v", string(ee.Stderr))
		}
	} else if err.Error() != "moo" {
		t.Error(err)
	}

	if f, _, err := pc.Open("pategory", 1, "moo.txt"); err != nil {
		t.Error(err)
	} else if buf, err := ioutil.ReadAll(f); err != nil {
		f.Close()
		t.Error(err)
	} else if string(buf) != "Moo.\n" {
		f.Close()
		t.Errorf("Wrong contents: %#v", string(buf))
	} else {
		f.Close()
	}

	if f, _, err := pc.Open("pategory", 1, "not.there"); err == nil {
		f.Close()
		t.Errorf("Non-existent file didn't return error: %#v", f)
	}
}
