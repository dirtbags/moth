package main

import (
	"bufio"
	"bytes"
	"fmt"
	"gopkg.in/russross/blackfriday.v2"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net/mail"
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

type HeaderParser func([]byte) (*Puzzle, error)

func YamlParser(input []byte) (*Puzzle, error) {
	puzzle := new(Puzzle)

	err := yaml.Unmarshal(input, puzzle)
	if err != nil {
		return nil, err
	}
	return puzzle, nil
}

func AttachmentParser(val []string) ([]Attachment) {
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
		case "script":
			puzzle.Pre.Scripts = AttachmentParser(val)
		case "file":
			puzzle.Pre.Attachments = AttachmentParser(val)
		case "answer":
			puzzle.Answers = val
		case "summary":
			puzzle.Debug.Summary = val[0]
		case "hint":
			puzzle.Debug.Hints = val
		case "ksa":
			puzzle.Post.KSAs = val
		default:
			return nil, fmt.Errorf("Unknown header field: %s", key)
		}
	}

	return puzzle, nil
}

func ParseMoth(r io.Reader) (*Puzzle, error) {
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
		return nil, err
	}
	
	// Markdownify the body
	bodyB := blackfriday.Run(bodyBuf.Bytes())
	if (puzzle.Pre.Body != "") && (len(bodyB) > 0) {
		log.Print("Body specified in header; overwriting...")
	}
	puzzle.Pre.Body = string(bodyB)

	return puzzle, nil
}
