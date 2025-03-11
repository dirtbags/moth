package main

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
)

// TeamID uniquely identifies a team.
type TeamID string

func ParseTeamID(filename string) TeamID {
	s, err := url.PathUnescape(filename)
	if err != nil {
		return TeamID(filename)
	}
	return TeamID(s)
}
func (id TeamID) String() string {
	return string(id)
}
func (id TeamID) Filename() string {
	return url.PathEscape(id.String())
}


type TeamIDList struct {
	teams []TeamID
}

func (tl *TeamIDList) WriteTo(w io.Writer) (int64, error) {
	counter := new(WriteCounter)
	tw := io.MultiWriter(w, counter)
	for _, id := range tl.teams {
		fmt.Fprintln(tw, id)
	}
	return counter.Len(), nil
}

func (tl *TeamIDList) ReadFrom(r io.Reader) (int64, error) {
	counter := new(WriteCounter)
	tr := io.TeeReader(r, counter)
	scanner := bufio.NewScanner(tr)
	for scanner.Scan() {
		line := scanner.Text()
		id := TeamID(line)
		tl.teams = append(tl.teams, id)
	}
	return counter.Len(), nil
}

func (tl *TeamIDList) Values() []TeamID {
	return tl.teams
}

func (tl *TeamIDList) Append(id TeamID) {
	tl.teams = append(tl.teams, id)
}

func (tl *TeamIDList) Len() int {
	return len(tl.teams)
}
