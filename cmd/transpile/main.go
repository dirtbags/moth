package main

import (
	"gopkg.in/russross/blackfriday.v2"
	"gopkg.in/yaml.v2"
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"io"
	"os"
	"net/mail"
	"strings"
)

type Header struct {
	Pre struct {
		Authors []string
	}
	Answers []string
	Post struct {
		Objective string
	}
	Debug struct {
		Log []string
		Error string
	}
}

type HeaderParser func([]byte) (*Header, error)

func YamlParser(input []byte) (*Header, error) {
	header := new(Header)
	
	err := yaml.Unmarshal(input, header)
	if err != nil {
		return nil, err
	}
	return header, nil
}

func Rfc822Parser(input []byte) (*Header, error) {
	msgBytes := append(input, '\n')
	r := bytes.NewReader(msgBytes)
	m, err := mail.ReadMessage(r)
	if err != nil {
		return nil, err
	}
	
	header := new(Header)
	for key, val := range m.Header {
		key = strings.ToLower(key)
		switch key {
			case "author":
				header.Pre.Authors = val
			case "answer":
				header.Answers = val
			default:
				return nil, fmt.Errorf("Unknown header field: %s", key)
		}
	}

	return header, nil
}


func parse(r io.Reader) (error) {
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
	
	header, err := headerParser(headerBuf.Bytes())
	if err != nil {
		return err
	}
	
	headerB, _ := yaml.Marshal(header)
	bodyB := blackfriday.Run(bodyBuf.Bytes())
	fmt.Println(string(headerB))
	fmt.Println("")
	fmt.Println(string(bodyB))

	return nil
}

func main() {
	flag.Parse()
	
	if flag.NArg() < 1 {
		fmt.Fprintf(flag.CommandLine.Output(), "Error: no files to parse\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	
	for _,filename := range flag.Args() {
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
