package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

// PuzzleProvider establishes the functionality required to provide one puzzle.
type PuzzleProvider interface {
	// Puzzle returns a Puzzle struct for the current puzzle.
	Puzzle() (Puzzle, error)

	// Open returns a newly-opened file.
	Open(filename string) (io.ReadCloser, error)

	// Answer returns whether the provided answer is correct.
	Answer(answer string) bool
}

// NewFsPuzzle returns a new FsPuzzle for points.
func NewFsPuzzle(fs afero.Fs, points int) PuzzleProvider {
	pfs := NewRecursiveBasePathFs(fs, strconv.Itoa(points))
	if info, err := pfs.Stat("mkpuzzle"); (err == nil) && (info.Mode()&0100 != 0) {
		if command, err := pfs.RealPath(info.Name()); err != nil {
			log.Println("Unable to resolve full path to", info.Name(), pfs)
		} else {
			return FsCommandPuzzle{
				fs:      pfs,
				command: command,
				timeout: 2 * time.Second,
			}
		}
	}

	return FsPuzzle{
		fs: pfs,
	}
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

// Answer checks whether the given answer is correct.
func (fp FsPuzzle) Answer(answer string) bool {
	return false
}

// FsCommandPuzzle provides an FsPuzzle backed by running a command.
type FsCommandPuzzle struct {
	fs      afero.Fs
	command string
	timeout time.Duration
}

// Puzzle returns a Puzzle struct for the current puzzle.
func (fp FsCommandPuzzle) Puzzle() (Puzzle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fp.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, fp.command)
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

// Open returns a newly-opened file.
func (fp FsCommandPuzzle) Open(filename string) (io.ReadCloser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fp.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, fp.command, "-file", filename)
	// BUG(neale): FsCommandPuzzle.Open() reads everything into memory, and will suck for large files.
	out, err := cmd.Output()
	if err != nil {
		return NopReadCloser{}, err
	}
	buf := bytes.NewBuffer(out)

	return ioutil.NopCloser(buf), nil
}

// Answer checks whether the given answer is correct.
func (fp FsCommandPuzzle) Answer(answer string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), fp.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, fp.command, "-answer", answer)
	out, err := cmd.Output()
	if err != nil {
		log.Print("ERROR", err)
		return false
	}

	switch strings.TrimSpace(string(out)) {
	case "correct":
		return true
	}
	return false
}
