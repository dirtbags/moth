package main

import (
	"fmt"
	"strings"
)

type Award struct {
	// Unix epoch time of this event
	When     int64
	TeamId   string
	Category string
	Points   int
}

func ParseAward(s string) (*Award, error) {
	ret := Award{}

	s = strings.TrimSpace(s)

	n, err := fmt.Sscanf(s, "%d %s %s %d", &ret.When, &ret.TeamId, &ret.Category, &ret.Points)
	if err != nil {
		return nil, err
	} else if n != 4 {
		return nil, fmt.Errorf("Malformed award string: only parsed %d fields", n)
	}

	return &ret, nil
}

func (a *Award) String() string {
	return fmt.Sprintf("%d %s %s %d", a.When, a.TeamId, a.Category, a.Points)
}

func (a *Award) Same(o *Award) bool {
	switch {
	case a.TeamId != o.TeamId:
		return false
	case a.Category != o.Category:
		return false
	case a.Points != o.Points:
		return false
	}
	return true
}
