package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/russross/blackfriday.v2"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net/mail"
	"os"
	"strings"
)

type Attachment struct {
	Filename       string // Filename presented as part of puzzle
	FilesystemPath string // Filename in backing FS (URL, mothball, or local FS)
	Listed         bool   // Whether this file is listed as an attachment
}

type Puzzle struct {
	Pre struct {
		Authors       []string
		Attachments   []Attachment
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

type HeaderParser func([]byte) (*Puzzle, error)

func YamlParser(input []byte) (*Puzzle, error) {
	puzzle := new(Puzzle)

	err := yaml.Unmarshal(input, puzzle)
	if err != nil {
		return nil, err
	}
	return puzzle, nil
}

func Rfc822Parser(input []byte) (*Puzzle, error) {
	msgBytes := append(input, '\n')
	r := bytes.NewReader(msgBytes)
	m, err := mail.ReadMessage(r)
	if err != nil {
		return nil, err
	}

	puzzle := new(Puzzle)
	for key, val := range m.Header {
		key = strings.ToLower(key)
		switch key {
		case "author":
			puzzle.Pre.Authors = val
		case "pattern":
			puzzle.Pre.AnswerPattern = val[0]
		case "answer":
			puzzle.Answers = val
		case "summary":
			puzzle.Debug.Summary = val[0]
		case "hint":
			puzzle.Debug.Hints = val
		case "ksa":
			puzzle.Post.KSAs = val
		case "file":
			for _, txt := range val {
				parts := strings.SplitN(txt, " ", 3)
				attachment := Attachment{}
				attachment.FilesystemPath = parts[0]
				if len(parts) > 1 {
					attachment.Filename = parts[1]
				} else {
					attachment.Filename = attachment.FilesystemPath
				}
				if (len(parts) > 2) && (parts[2] == "hidden") {
					attachment.Listed = false
				} else {
					attachment.Listed = true
				}

				puzzle.Pre.Attachments = append(puzzle.Pre.Attachments, attachment)
			}
		default:
			return nil, fmt.Errorf("Unknown header field: %s", key)
		}
	}

	return puzzle, nil
}

func parse(r io.Reader) error {
	headerEnd := ""
	headerBuf := new(bytes.Buffer)
	headerParser := Rfc822Parser

	scanner := bufio.NewScanner(r)
	lineNo := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNo += 1
		if lineNo == 1 {
			if line == "---" {
				headerParser = YamlParser
				headerEnd = "---"
				continue
			} else {
				headerParser = Rfc822Parser
			}
		}
		if line == headerEnd {
			break
		}
		headerBuf.WriteString(line)
		headerBuf.WriteRune('\n')
	}

	bodyBuf := new(bytes.Buffer)
	for scanner.Scan() {
		line := scanner.Text()
		lineNo += 1
		bodyBuf.WriteString(line)
		bodyBuf.WriteRune('\n')
	}

	puzzle, err := headerParser(headerBuf.Bytes())
	if err != nil {
		return err
	}

	bodyB := blackfriday.Run(bodyBuf.Bytes())

	if (puzzle.Pre.Body != "") && (len(bodyB) > 0) {
		log.Print("Body specified in header; overwriting...")
	}
	puzzle.Pre.Body = string(bodyB)

	puzzleB, _ := json.MarshalIndent(puzzle, "", "  ")

	fmt.Println(string(puzzleB))

	return nil
}

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(flag.CommandLine.Output(), "Error: no files to parse\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	for _, filename := range flag.Args() {
		f, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		if err := parse(f); err != nil {
			log.Fatal(err)
		}
	}
}
