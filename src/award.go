package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Award struct {
	When     time.Time
	TeamId   string
	Category string
	Points   int
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

func ParseAward(s string) (*Award, error) {
	ret := Award{}

	s = strings.Trim(s, " \t\n")

	parts := strings.SplitN(s, " ", 5)
	if len(parts) < 4 {
		return nil, fmt.Errorf("Malformed award string")
	}

	whenEpoch, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Malformed timestamp: %s", parts[0])
	}
	ret.When = time.Unix(whenEpoch, 0)

	ret.TeamId = parts[1]
	ret.Category = parts[2]

	points, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, fmt.Errorf("Malformed Points: %s: %v", parts[3], err)
	}
	ret.Points = points

	return &ret, nil
}
