// Provides a Puzzle interface that runs a command for each request
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ProviderCommand specifies a command to run for the puzzle API
type ProviderCommand struct {
	Path string
	Args []string
}

// Inventory runs with "action=inventory", and parses the output into a category list.
func (pc ProviderCommand) Inventory() (inv []Category) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, pc.Path, pc.Args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "ACTION=inventory")

	stdout, err := cmd.Output()
	if err != nil {
		log.Print(err)
		return
	}

	for _, line := range strings.Split(string(stdout), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, " ")
		if len(parts) < 2 {
			log.Println("Skipping misformatted line:", line)
			continue
		}
		name := parts[0]
		puzzles := make([]int, 0, 10)
		for _, pointsString := range parts[1:] {
			points, err := strconv.Atoi(pointsString)
			if err != nil {
				log.Println(err)
				continue
			}
			puzzles = append(puzzles, points)
		}
		sort.Ints(puzzles)
		inv = append(inv, Category{name, puzzles})
	}
	return
}

// NullReadSeekCloser wraps a no-op Close method around an io.ReadSeeker.
type NullReadSeekCloser struct {
	io.ReadSeeker
}

// Close does nothing.
func (f NullReadSeekCloser) Close() error {
	return nil
}

// Open passes its arguments to the command with "action=open".
func (pc ProviderCommand) Open(cat string, points int, path string) (ReadSeekCloser, time.Time, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, pc.Path, pc.Args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "ACTION=open")
	cmd.Env = append(cmd.Env, "CAT="+cat)
	cmd.Env = append(cmd.Env, "POINTS="+strconv.Itoa(points))
	cmd.Env = append(cmd.Env, "FILENAME="+path)

	stdoutBytes, err := cmd.Output()
	stdout := NullReadSeekCloser{bytes.NewReader(stdoutBytes)}
	now := time.Now()
	return stdout, now, err
}

// CheckAnswer passes its arguments to the command with "action=answer".
// If the command exits successfully and sends "correct" to stdout,
// nil is returned.
func (pc ProviderCommand) CheckAnswer(cat string, points int, answer string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, pc.Path, pc.Args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "ACTION=answer")
	cmd.Env = append(cmd.Env, "CAT="+cat)
	cmd.Env = append(cmd.Env, "POINTS="+strconv.Itoa(points))
	cmd.Env = append(cmd.Env, "ANSWER="+answer)

	stdout, err := cmd.Output()
	if ee, ok := err.(*exec.ExitError); ok {
		log.Printf("%s: %s", pc.Path, string(ee.Stderr))
		return false, err
	} else if err != nil {
		return false, err
	}
	result := strings.TrimSpace(string(stdout))

	if result != "correct" {
		if result == "" {
			result = "Nothing written to stdout"
		}
		return false, nil
	}

	return true, nil
}

// Mothball just returns an error
func (pc ProviderCommand) Mothball(cat string) (*bytes.Reader, error) {
	return nil, fmt.Errorf("Can't package a command-generated category")
}

// Maintain does nothing: a command puzzle ProviderCommand has no housekeeping
func (pc ProviderCommand) Maintain(updateInterval time.Duration) {
}
