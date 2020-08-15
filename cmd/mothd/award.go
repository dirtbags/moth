package main

import (
	"encoding/json"
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

type AwardList []*Award

// Implement sort.Interface on AwardList
func (awards AwardList) Len() int {
  return len(awards)
}

func (awards AwardList) Less(i, j int) bool {
  return awards[i].When.Before(awards[j].When)
}

func (awards AwardList) Swap(i, j int) {
  tmp := awards[i]
  awards[i] = awards[j]
  awards[j] = tmp
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

func (a *Award) MarshalJSON() ([]byte, error) {
	if a == nil {
		return []byte("null"), nil
	}
	ao := []interface{}{
		a.When,
		a.TeamId,
		a.Category,
		a.Points,
	}

	return json.Marshal(ao)
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
