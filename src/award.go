package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Award struct {
	When     time.Time
	TeamId   string
	Category string
	Points   int
}

func ParseAward(s string) (*Award, error) {
	ret := Award{}

	s = strings.Trim(s, " \t\n")

	var whenEpoch int64

	n, err := fmt.Sscanf(s, "%d %s %s %d", &whenEpoch, &ret.TeamId, &ret.Category, &ret.Points)
	if err != nil {
		return nil, err
	} else if n != 4 {
		return nil, fmt.Errorf("Malformed award string: only parsed %d fields", n)
	}

	ret.When = time.Unix(whenEpoch, 0)

	return &ret, nil
}

func (a *Award) String() string {
	return fmt.Sprintf("%d %s %s %d", a.When.Unix(), a.TeamId, a.Category, a.Points)
}

func (a *Award) MarshalJSON() ([]byte, error) {
	if a == nil {
		return []byte("null"), nil
	}
	jTeamId, err := json.Marshal(a.TeamId)
	if err != nil {
		return nil, err
	}
	jCategory, err := json.Marshal(a.Category)
	if err != nil {
		return nil, err
	}
	ret := fmt.Sprintf(
		"[%d,%s,%s,%d]",
		a.When.Unix(),
		jTeamId,
		jCategory,
		a.Points,
	)
	return []byte(ret), nil
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
