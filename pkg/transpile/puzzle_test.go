package transpile

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestPuzzle(t *testing.T) {
	puzzleFs := newTestFs()
	catFs := NewRecursiveBasePathFs(puzzleFs, "cat0")

	{
		pd := NewFsPuzzlePoints(catFs, 1)
		p, err := pd.Puzzle()
		if err != nil {
			t.Error(err)
		}

		if (len(p.Answers) == 0) || (p.Answers[0] != "YAML answer") {
			t.Error("Answers are wrong", p.Answers)
		}
		if (len(p.Pre.Authors) != 3) || (p.Pre.Authors[1] != "Buster") {
			t.Error("Authors are wrong", p.Pre.Authors)
		}
		if p.Pre.Body != "<p>YAML body</p>\n" {
			t.Errorf("Body parsed wrong: %#v", p.Pre.Body)
		}

		f, err := pd.Open("moo.txt")
		if err != nil {
			t.Error(err)
		}
		defer f.Close()
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, f); err != nil {
			t.Error(err)
		}
		if buf.String() != "Moo." {
			t.Error("Attachment wrong: ", buf.String())
		}
	}

	{
		p, err := NewFsPuzzlePoints(catFs, 2).Puzzle()
		if err != nil {
			t.Error(err)
		}
		if (len(p.Answers) == 0) || (p.Answers[0] != "RFC822 answer") {
			t.Error("Answers are wrong", p.Answers)
		}
		if (len(p.Pre.Authors) != 3) || (p.Pre.Authors[1] != "Arthur") {
			t.Error("Authors are wrong", p.Pre.Authors)
		}
		if p.Pre.Body != "<p>RFC822 body</p>\n" {
			t.Errorf("Body parsed wrong: %#v", p.Pre.Body)
		}
	}

	if _, err := NewFsPuzzlePoints(catFs, 3).Puzzle(); err != nil {
		t.Error("Legacy `puzzle.moth` file:", err)
	}

	if puzzle, err := NewFsPuzzlePoints(catFs, 4).Puzzle(); err != nil {
		t.Error("Markdown test file:", err)
	} else if !strings.Contains(puzzle.Pre.Body, "<table>") {
		t.Error("Markdown table extension isn't making tables")
	} else if !strings.Contains(puzzle.Pre.Body, "<dl>") {
		t.Error("Markdown dictionary extension isn't making tables")
	}

	if _, err := NewFsPuzzlePoints(catFs, 99).Puzzle(); err == nil {
		t.Error("Non-existent puzzle", err)
	}

	if _, err := NewFsPuzzlePoints(catFs, 10).Puzzle(); err == nil {
		t.Error("Broken YAML")
	}
	if _, err := NewFsPuzzlePoints(catFs, 20).Puzzle(); err == nil {
		t.Error("Bad RFC822 header")
	}
	if _, err := NewFsPuzzlePoints(catFs, 21).Puzzle(); err == nil {
		t.Error("Boken RFC822 header")
	}

	{
		fs := afero.NewMemMapFs()
		if err := afero.WriteFile(fs, "1/mkpuzzle", []byte("bleat"), 0755); err != nil {
			t.Error(err)
		}
		p := NewFsPuzzlePoints(fs, 1)
		if _, ok := p.(FsCommandPuzzle); !ok {
			t.Error("We didn't get an FsCommandPuzzle")
		}
		if _, err := p.Puzzle(); err == nil {
			t.Error("We didn't get an error trying to run a command from a MemMapFs")
		}
	}
}

func TestFsPuzzle(t *testing.T) {
	catFs := NewRecursiveBasePathFs(NewRecursiveBasePathFs(afero.NewOsFs(), "testdata"), "static")

	if _, err := NewFsPuzzlePoints(catFs, 1).Puzzle(); err != nil {
		t.Error(err)
	}

	if _, err := NewFsPuzzlePoints(catFs, 2).Puzzle(); err != nil {
		t.Error(err)
	}

	mkpuzzleDir := NewFsPuzzlePoints(catFs, 3)
	if _, err := mkpuzzleDir.Puzzle(); err != nil {
		t.Error(err)
	}

	if r, err := mkpuzzleDir.Open("moo.txt"); err != nil {
		t.Error(err)
	} else {
		defer r.Close()
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, r); err != nil {
			t.Error(err)
		}
		if buf.String() != "Moo.\n" {
			t.Errorf("Wrong body: %#v", buf.String())
		}
	}

	if r, err := mkpuzzleDir.Open("error"); err == nil {
		r.Close()
		t.Error("Error open didn't return error")
	}

	if !mkpuzzleDir.Answer("moo") {
		t.Error("Right answer marked wrong")
	}
	if mkpuzzleDir.Answer("wrong") {
		t.Error("Wrong answer marked correct")
	}
	if mkpuzzleDir.Answer("error") {
		t.Error("Error answer marked correct")
	}
}

func TestAttachment(t *testing.T) {
	buf := bytes.NewBufferString(`
pre:
  attachments: 
    - simple
    - filename: complex
      filesystempath: backingfile
`)
	p, err := yamlHeaderParser(buf)
	if err != nil {
		t.Error(err)
		return
	}

	att := p.Pre.Attachments
	if len(att) != 2 {
		t.Error("Wrong number of attachments", att)
	}
	if att[0].Filename != "simple" {
		t.Error("Attachment 0 wrong")
	}
	if att[0].Filename != att[0].FilesystemPath {
		t.Error("Attachment 0 wrong")
	}
	if att[1].Filename != "complex" {
		t.Error("Attachment 1 wrong")
	}
	if att[1].FilesystemPath != "backingfile" {
		t.Error("Attachment 2 wrong")
	}
}
