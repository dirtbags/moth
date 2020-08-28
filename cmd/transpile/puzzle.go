package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/mail"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/russross/blackfriday"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

// NewPuzzle returns a new Puzzle for points.
func NewPuzzle(fs afero.Fs, points int) (*Puzzle, error) {
	p := &Puzzle{
		fs: afero.NewBasePathFs(fs, strconv.Itoa(points)),
	}

	if err := p.parseMoth(); err != nil {
		return p, err
	}
	// BUG(neale): Doesn't yet handle "puzzle.py" or "mkpuzzle"

	return p, nil
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
	fs      afero.Fs
}

// Attachment carries information about an attached file.
type Attachment struct {
	Filename       string // Filename presented as part of puzzle
	FilesystemPath string // Filename in backing FS (URL, mothball, or local FS)
	Listed         bool   // Whether this file is listed as an attachment
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

func (p *Puzzle) yamlHeaderParser(r io.Reader) error {
	decoder := yaml.NewDecoder(r)
	decoder.SetStrict(true)
	return decoder.Decode(p)
}

func (p *Puzzle) rfc822HeaderParser(r io.Reader) error {
	m, err := mail.ReadMessage(r)
	if err != nil {
		return fmt.Errorf("Parsing RFC822 headers: %v", err)
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
			return fmt.Errorf("Unknown header field: %s", key)
		}
	}

	return nil
}

func (p *Puzzle) parseMoth() error {
	r, err := p.fs.Open("puzzle.moth")
	if err != nil {
		return err
	}
	defer r.Close()

	headerBuf := new(bytes.Buffer)
	headerParser := p.rfc822HeaderParser
	headerEnd := ""

	scanner := bufio.NewScanner(r)
	lineNo := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNo++
		if lineNo == 1 {
			if line == "---" {
				headerParser = p.yamlHeaderParser
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

	if err := headerParser(headerBuf); err != nil {
		return err
	}

	// Markdownify the body
	if (p.Pre.Body != "") && (bodyBuf.Len() > 0) {
		return fmt.Errorf("Puzzle body present in header and in moth body")
	}
	p.Pre.Body = string(blackfriday.Run(bodyBuf.Bytes()))

	return nil
}

func (p *Puzzle) mkpuzzle() error {
	bfs, ok := p.fs.(*afero.BasePathFs)
	if !ok {
		return fmt.Errorf("Fs won't resolve real paths for %v", p)
	}
	mkpuzzlePath, err := bfs.RealPath("mkpuzzle")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, mkpuzzlePath)
	stdout, err := cmd.Output()
	if err != nil {
		return err
	}

	jsdec := json.NewDecoder(bytes.NewReader(stdout))
	jsdec.DisallowUnknownFields()
	puzzle := new(Puzzle)
	if err := jsdec.Decode(puzzle); err != nil {
		return err
	}

	return nil
}

// Open returns a newly-opened file.
func (p *Puzzle) Open(name string) (io.ReadCloser, error) {
	// BUG(neale): You cannot open generated files in puzzles, only files actually on the disk
	return p.fs.Open(name)
}
