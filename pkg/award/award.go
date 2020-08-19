// Package award defines a MOTH award, and provides tools to use them.
package award

import (
	"encoding/json"
	"fmt"
	"strings"
)

// T represents a single award event.
type T struct {
	// Unix epoch time of this event
	When     int64
	TeamID   string
	Category string
	Points   int
}

// List is a collection of award events.
type List []T

// Len returns the length of the awards list.
func (awards List) Len() int {
	return len(awards)
}

// Less returns true if i was awarded before j.
func (awards List) Less(i, j int) bool {
	return awards[i].When < awards[j].When
}

// Swap exchanges the awards in positions i and j.
func (awards List) Swap(i, j int) {
	tmp := awards[i]
	awards[i] = awards[j]
	awards[j] = tmp
}

// Parse parses a string log entry into an award.T.
func Parse(s string) (T, error) {
	ret := T{}

	s = strings.TrimSpace(s)

	n, err := fmt.Sscanf(s, "%d %s %s %d", &ret.When, &ret.TeamID, &ret.Category, &ret.Points)
	if err != nil {
		return ret, err
	} else if n != 4 {
		return ret, fmt.Errorf("Malformed award string: only parsed %d fields", n)
	}

	return ret, nil
}

// String returns a log entry string for an award.T.
func (a T) String() string {
	return fmt.Sprintf("%d %s %s %d", a.When, a.TeamID, a.Category, a.Points)
}

// MarshalJSON returns the award event, encoded as a list.
func (a T) MarshalJSON() ([]byte, error) {
	ao := []interface{}{
		a.When,
		a.TeamID,
		a.Category,
		a.Points,
	}

	return json.Marshal(ao)
}

// Equal returns true if two award events represent the same award.
// Timestamps are ignored in this comparison!
func (a T) Equal(o T) bool {
	switch {
	case a.TeamID != o.TeamID:
		return false
	case a.Category != o.Category:
		return false
	case a.Points != o.Points:
		return false
	}
	return true
}
