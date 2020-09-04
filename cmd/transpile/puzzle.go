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
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/russross/blackfriday"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

// NewFsPuzzle returns a new FsPuzzle for points.
func NewFsPuzzle(fs afero.Fs, points int) *FsPuzzle {
	fp := &FsPuzzle{
		fs: NewBasePathFs(fs, strconv.Itoa(points)),
	}

	return fp
}

// FsPuzzle is a single puzzle's directory.
type FsPuzzle struct {
	fs       afero.Fs
	mkpuzzle bool
}

// Puzzle returns a Puzzle struct for the current puzzle.
func (fp FsPuzzle) Puzzle() (Puzzle, error) {
	r, err := fp.fs.Open("puzzle.md")
	if err != nil {
		var err2 error
		if r, err2 = fp.fs.Open("puzzle.moth"); err2 != nil {
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

// Open returns a newly-opened file.
func (fp FsPuzzle) Open(name string) (io.ReadCloser, error) {
	return fp.fs.Open(name)
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

func (fp FsPuzzle) Answer(answer string) bool {
	return false
}

type FsCommandPuzzle struct {
	fs afero.Fs
}

func (fp FsCommandPuzzle) Puzzle() (Puzzle, error) {
	bfs, ok := fp.fs.(*BasePathFs)
	if !ok {
		return Puzzle{}, fmt.Errorf("Fs won't resolve real paths for %v", fp)
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

func (fp FsCommandPuzzle) Open(filename string) (io.ReadCloser, error) {
	return NopReadCloser{}, fmt.Errorf("Not implemented")
}

func (fp FsCommandPuzzle) Answer(answer string) bool {
	return false
}
