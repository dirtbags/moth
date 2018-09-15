package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Award struct {
	When time.Time
	TeamId string
	Category string
	Points int
}

func (a *Award) String() string {
	return fmt.Sprintf("%d %s %s %d", a.When.Unix(), a.TeamId, a.Category, a.Points)
}

func ParseAward(s string) (*Award, error) {
	ret := Award{}
	
	parts := strings.SplitN(s, " ", 5)
	if len(parts) < 4 {
		return nil, fmt.Errorf("Malformed award string")
	}
	
	whenEpoch, err := strconv.ParseInt(parts[0], 10, 64)
	if (err != nil) {
		return nil, fmt.Errorf("Malformed timestamp: %s", parts[0])
	}
	ret.When = time.Unix(whenEpoch, 0)
	
	ret.TeamId = parts[1]
	ret.Category = parts[2]
	
	points, err := strconv.Atoi(parts[3])
	if (err != nil) {
		return nil, fmt.Errorf("Malformed Points: %s", parts[3])
	}
	ret.Points = points

	return &ret, nil
}

