package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/mail"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/russross/blackfriday"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

// NewPuzzleDir returns a new PuzzleDir for points.
func NewPuzzleDir(fs afero.Fs, points int) *PuzzleDir {
	pd := &PuzzleDir{
		fs: NewBasePathFs(fs, strconv.Itoa(points)),
	}
	// BUG(neale): Doesn't yet handle "puzzle.py" or "mkpuzzle"

	return pd
}

// PuzzleDir is a single puzzle's directory.
type PuzzleDir struct {
	fs afero.Fs
	mkpuzzle bool
}

// Open returns a newly-opened file.
func (pd *PuzzleDir) Open(name string) (io.ReadCloser, error) {
	// BUG(neale): You cannot open generated files in puzzles, only files actually on the disk
	if _, err := pd.fs.Stat(""
	return pd.fs.Open(name)
}

// Export returns a Puzzle struct for the current puzzle.
func (pd *PuzzleDir) Export() (Puzzle, error) {
	p, staticErr := pd.exportStatic()
	if staticErr == nil {
		return p, nil
	}

	// Only fall through if the static files don't exist. Otherwise, report the error.
	if !os.IsNotExist(staticErr) {
		return p, staticErr
	}

	if p, cmdErr := pd.exportCommand(); cmdErr == nil {
		return p, nil
	} else if os.IsNotExist(cmdErr) {
		// If the command doesn't exist either, report the non-existence of the static file instead.
		return p, staticErr
	} else {
		return p, cmdErr
	}
}

func (pd *PuzzleDir) exportStatic() (Puzzle, error) {
	r, err := pd.fs.Open("puzzle.md")
	if err != nil {
		var err2 error
		if r, err2 = pd.fs.Open("puzzle.moth"); err2 != nil {
			return Puzzle{}, err
		}
	}
	defer r.Close()

	headerBuf := new(bytes.Buffer)
	headerParser := rfc822HeaderParser
	headerEnd := ""

	scanner := bufio.NewScanner(r)
	lineNo := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNo++
		if lineNo == 1 {
			if line == "---" {
				headerParser = yamlHeaderParser
				headerEnd = "---"
				continue
			}
		}
		if line == headerEnd {
			headerBuf.WriteRune('\n')
			break
		}
		headerBuf.WriteString(line)
		headerBuf.WriteRune('\n')
	}

	bodyBuf := new(bytes.Buffer)
	for scanner.Scan() {
		line := scanner.Text()
		lineNo++
		bodyBuf.WriteString(line)
		bodyBuf.WriteRune('\n')
	}

	puzzle, err := headerParser(headerBuf)
	if err != nil {
		return puzzle, err
	}

	// Markdownify the body
	if puzzle.Pre.Body != "" {
		if bodyBuf.Len() > 0 {
			return puzzle, fmt.Errorf("Puzzle body present in header and in moth body")
		}
	} else {
		puzzle.Pre.Body = string(blackfriday.Run(bodyBuf.Bytes()))
	}

	return puzzle, nil
}

func (pd *PuzzleDir) exportCommand() (Puzzle, error) {
	bfs, ok := pd.fs.(*BasePathFs)
	if !ok {
		return Puzzle{}, fmt.Errorf("Fs won't resolve real paths for %v", pd)
	}
	mkpuzzlePath, err := bfs.RealPath("mkpuzzle")
	if err != nil {
		return Puzzle{}, err
	}
	log.Print(mkpuzzlePath)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, mkpuzzlePath)
	stdout, err := cmd.Output()
	if err != nil {
		return Puzzle{}, err
	}

	jsdec := json.NewDecoder(bytes.NewReader(stdout))
	jsdec.DisallowUnknownFields()
	puzzle := Puzzle{}
	if err := jsdec.Decode(&puzzle); err != nil {
		return Puzzle{}, err
	}

	return puzzle, nil
}

func legacyAttachmentParser(val []string) []Attachment {
	ret := make([]Attachment, len(val))
	for idx, txt := range val {
		parts := strings.SplitN(txt, " ", 3)
		cur := Attachment{}
		cur.FilesystemPath = parts[0]
		if len(parts) > 1 {
			cur.Filename = parts[1]
		} else {
			cur.Filename = cur.FilesystemPath
		}
		if (len(parts) > 2) && (parts[2] == "hidden") {
			cur.Listed = false
		} else {
			cur.Listed = true
		}
		ret[idx] = cur
	}
	return ret
}

// Puzzle contains everything about a puzzle.
type Puzzle struct {
	Pre struct {
		Authors       []string
		Attachments   []Attachment
		Scripts       []Attachment
		AnswerPattern string
		Body          string
	}
	Post struct {
		Objective string
		Success   struct {
			Acceptable string
			Mastery    string
		}
		KSAs []string
	}
	Debug struct {
		Log     []string
		Errors  []string
		Hints   []string
		Summary string
	}
	Answers []string
}

// Attachment carries information about an attached file.
type Attachment struct {
	Filename       string // Filename presented as part of puzzle
	FilesystemPath string // Filename in backing FS (URL, mothball, or local FS)
	Listed         bool   // Whether this file is listed as an attachment
}

func yamlHeaderParser(r io.Reader) (Puzzle, error) {
	p := Puzzle{}
	decoder := yaml.NewDecoder(r)
	decoder.SetStrict(true)
	err := decoder.Decode(&p)
	return p, err
}

func rfc822HeaderParser(r io.Reader) (Puzzle, error) {
	p := Puzzle{}
	m, err := mail.ReadMessage(r)
	if err != nil {
		return p, fmt.Errorf("Parsing RFC822 headers: %v", err)
	}

	for key, val := range m.Header {
		key = strings.ToLower(key)
		switch key {
		case "author":
			p.Pre.Authors = val
		case "pattern":
			p.Pre.AnswerPattern = val[0]
		case "script":
			p.Pre.Scripts = legacyAttachmentParser(val)
		case "file":
			p.Pre.Attachments = legacyAttachmentParser(val)
		case "answer":
			p.Answers = val
		case "summary":
			p.Debug.Summary = val[0]
		case "hint":
			p.Debug.Hints = val
		case "ksa":
			p.Post.KSAs = val
		default:
			return p, fmt.Errorf("Unknown header field: %s", key)
		}
	}

	return p, nil
}
