package transpile

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/mail"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/afero"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"gopkg.in/yaml.v2"
)

// AnswerResponse is handed back when we ask for an answer to be checked.
type AnswerResponse struct {
	Correct bool
}

// Puzzle contains everything about a puzzle that a client would see.
type Puzzle struct {
	Pre struct {
		Authors       []string
		Attachments   []string
		Scripts       []string
		Body          string
		AnswerPattern string
		AnswerHashes  []string
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

func (puzzle *Puzzle) computeAnswerHashes() {
	if len(puzzle.Answers) == 0 {
		return
	}
	puzzle.Pre.AnswerHashes = make([]string, len(puzzle.Answers))
	for i, answer := range puzzle.Answers {
		sum := sha256.Sum256([]byte(answer))
		hexsum := fmt.Sprintf("%x", sum)
		puzzle.Pre.AnswerHashes[i] = hexsum
	}
}

// StaticPuzzle contains everything a static puzzle might tell us.
type StaticPuzzle struct {
	Pre struct {
		Authors       []string
		Attachments   []StaticAttachment
		Scripts       []StaticAttachment
		AnswerPattern string
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

// StaticAttachment carries information about an attached file.
type StaticAttachment struct {
	Filename       string // Filename presented as part of puzzle
	FilesystemPath string // Filename in backing FS (URL, mothball, or local FS)
}

// UnmarshalYAML allows a StaticAttachment to be specified as a single string.
// The way the yaml library works is weird.
func (sa *StaticAttachment) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(&sa.Filename); err == nil {
		sa.FilesystemPath = sa.Filename
		return nil
	}

	parts := new(struct {
		Filename       string
		FilesystemPath string
	})
	if err := unmarshal(parts); err != nil {
		return err
	}
	sa.Filename = parts.Filename
	sa.FilesystemPath = parts.FilesystemPath
	return nil
}

// ReadSeekCloser provides io.Reader, io.Seeker, and io.Closer.
type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

// PuzzleProvider establishes the functionality required to provide one puzzle.
type PuzzleProvider interface {
	// Puzzle returns a Puzzle struct for the current puzzle.
	Puzzle() (Puzzle, error)

	// Open returns a newly-opened file.
	Open(filename string) (ReadSeekCloser, error)

	// Answer returns whether the provided answer is correct.
	Answer(answer string) bool
}

// NewFsPuzzle returns a new FsPuzzle.
func NewFsPuzzle(fs afero.Fs) PuzzleProvider {
	var command string

	if info, err := fs.Stat("mkpuzzle"); !os.IsNotExist(err) {
		if (info.Mode() & 0100) != 0 {
			// Try to get the actual path to the executable
			if pfs, ok := fs.(*RecursiveBasePathFs); ok {
				if command, err = pfs.RealPath(info.Name()); err != nil {
					log.Println("WARN: Unable to resolve full path to", info.Name(), pfs)
				}
			} else if pfs, ok := fs.(*afero.BasePathFs); ok {
				if command, err = pfs.RealPath(info.Name()); err != nil {
					log.Println("WARN: Unable to resolve full path to", info.Name(), pfs)
				}
			}
		} else {
			log.Println("WARN: mkpuzzle exists, but isn't executable.")
		}
	}

	if command != "" {
		return FsCommandPuzzle{
			fs:      fs,
			command: command,
			timeout: 2 * time.Second,
		}
	}

	return FsPuzzle{
		fs: fs,
	}

}

// NewFsPuzzlePoints returns a new FsPuzzle for points.
func NewFsPuzzlePoints(fs afero.Fs, points int) PuzzleProvider {
	return NewFsPuzzle(NewRecursiveBasePathFs(fs, strconv.Itoa(points)))
}

// FsPuzzle is a single puzzle's directory.
type FsPuzzle struct {
	fs       afero.Fs
	mkpuzzle bool
}

// Puzzle returns a Puzzle struct for the current puzzle.
func (fp FsPuzzle) Puzzle() (Puzzle, error) {
	var puzzle Puzzle

	static, body, err := fp.staticPuzzle()
	if err != nil {
		return puzzle, err
	}

	// Convert to an exportable Puzzle
	puzzle.Post = static.Post
	puzzle.Debug = static.Debug
	puzzle.Answers = static.Answers
	puzzle.Pre.Authors = static.Pre.Authors
	puzzle.Pre.Body = string(body)
	puzzle.Pre.AnswerPattern = static.Pre.AnswerPattern
	puzzle.Pre.Attachments = make([]string, len(static.Pre.Attachments))
	for i, attachment := range static.Pre.Attachments {
		puzzle.Pre.Attachments[i] = attachment.Filename
	}
	puzzle.Pre.Scripts = make([]string, len(static.Pre.Scripts))
	for i, script := range static.Pre.Scripts {
		puzzle.Pre.Scripts[i] = script.Filename
	}
	puzzle.computeAnswerHashes()

	return puzzle, nil
}

// Open returns a newly-opened file.
func (fp FsPuzzle) Open(name string) (ReadSeekCloser, error) {
	empty := nopCloser{new(bytes.Reader)}
	static, _, err := fp.staticPuzzle()
	if err != nil {
		return empty, err
	}

	var fsPath string
	for _, attachment := range append(static.Pre.Attachments, static.Pre.Scripts...) {
		if attachment.Filename == name {
			if attachment.FilesystemPath == "" {
				fsPath = attachment.Filename
			} else {
				fsPath = attachment.FilesystemPath
			}
		}
	}
	if fsPath == "" {
		return empty, fmt.Errorf("Not listed in attachments or scripts: %s", name)
	}

	return fp.fs.Open(fsPath)
}

func (fp FsPuzzle) staticPuzzle() (StaticPuzzle, []byte, error) {
	r, err := fp.fs.Open("puzzle.md")
	if err != nil {
		var err2 error
		if r, err2 = fp.fs.Open("puzzle.moth"); err2 != nil {
			return StaticPuzzle{}, nil, err
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

	static, err := headerParser(headerBuf)
	if err != nil {
		return static, nil, err
	}

	body := new(bytes.Buffer)

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
			extension.DefinitionList,
		),
	)
	md.Convert(bodyBuf.Bytes(), body)

	return static, body.Bytes(), err
}

func legacyAttachmentParser(val []string) []StaticAttachment {
	ret := make([]StaticAttachment, len(val))
	for idx, txt := range val {
		parts := strings.SplitN(txt, " ", 3)
		cur := StaticAttachment{}
		cur.FilesystemPath = parts[0]
		if len(parts) > 1 {
			cur.Filename = parts[1]
		} else {
			cur.Filename = cur.FilesystemPath
		}
		ret[idx] = cur
	}
	return ret
}

func yamlHeaderParser(r io.Reader) (StaticPuzzle, error) {
	p := StaticPuzzle{}
	decoder := yaml.NewDecoder(r)
	decoder.SetStrict(true)
	err := decoder.Decode(&p)
	return p, err
}

func rfc822HeaderParser(r io.Reader) (StaticPuzzle, error) {
	p := StaticPuzzle{}
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
		case "solution":
			p.Debug.Hints = val
		case "ksa":
			p.Post.KSAs = val
		case "objective":
			p.Post.Objective = val[0]
		case "success.acceptable":
			p.Post.Success.Acceptable = val[0]
		case "success.mastery":
			p.Post.Success.Mastery = val[0]
		default:
			return p, fmt.Errorf("Unknown header field: %s", key)
		}
	}

	return p, nil
}

// Answer checks whether the given answer is correct.
func (fp FsPuzzle) Answer(answer string) bool {
	p, _, err := fp.staticPuzzle()
	if err != nil {
		return false
	}
	for _, ans := range p.Answers {
		if ans == answer {
			return true
		}
	}
	return false
}

// FsCommandPuzzle provides an FsPuzzle backed by running a command.
type FsCommandPuzzle struct {
	fs      afero.Fs
	command string
	timeout time.Duration
}

func (fp FsCommandPuzzle) run(command string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fp.timeout)
	defer cancel()

	cmdargs := append([]string{command}, args...)
	cmd := exec.CommandContext(ctx, "./"+path.Base(fp.command), cmdargs...)
	cmd.Dir = path.Dir(fp.command)
	out, err := cmd.Output()
	if err, ok := err.(*exec.ExitError); ok {
		stderr := strings.TrimSpace(string(err.Stderr))
		return nil, fmt.Errorf("%s (%s)", stderr, err.String())
	}
	return out, err
}

// Puzzle returns a Puzzle struct for the current puzzle.
func (fp FsCommandPuzzle) Puzzle() (Puzzle, error) {
	stdout, err := fp.run("puzzle")
	if exiterr, ok := err.(*exec.ExitError); ok {
		return Puzzle{}, errors.New(string(exiterr.Stderr))
	} else if err != nil {
		return Puzzle{}, err
	}

	jsdec := json.NewDecoder(bytes.NewReader(stdout))
	jsdec.DisallowUnknownFields()
	puzzle := Puzzle{}
	if err := jsdec.Decode(&puzzle); err != nil {
		return Puzzle{}, err
	}

	puzzle.computeAnswerHashes()

	return puzzle, nil
}

type nopCloser struct {
	io.ReadSeeker
}

func (c nopCloser) Close() error {
	return nil
}

// Open returns a newly-opened file.
// BUG(neale): FsCommandPuzzle.Open() reads everything into memory, and will suck for large files.
func (fp FsCommandPuzzle) Open(filename string) (ReadSeekCloser, error) {
	stdout, err := fp.run("file", filename)
	buf := nopCloser{bytes.NewReader(stdout)}
	if err != nil {
		return buf, err
	}

	return buf, nil
}

// Answer checks whether the given answer is correct.
func (fp FsCommandPuzzle) Answer(answer string) bool {
	stdout, err := fp.run("answer", answer)
	if err != nil {
		log.Printf("ERROR: checking answer: %s", err)
		return false
	}

	ans := AnswerResponse{}
	if err := json.Unmarshal(stdout, &ans); err != nil {
		log.Printf("ERROR: checking answer: %s", err)
		return false
	}

	return ans.Correct
}
