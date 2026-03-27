package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dirtbags/moth/v5/pkg/writecounter"
)

// PuzzleID uniquely identifies a puzzle.
// MOTHv2 and v3 used a long hash for the ID.
// MOTHv1 and v4 used an integer.
type PuzzleID string

func (id PuzzleID) String() string {
	return string(id)
}

// CategoryID uinquely identifies a category.
type CategoryID string

func (id CategoryID) String() string {
	return string(id)
}

// Award represents a single award event.
type Award struct {
	// Unix epoch time of this event
	When     time.Time
	Team     TeamID
	Category CategoryID
	Puzzle   PuzzleID
}

// New returns a new Award.
func NewAward(when time.Time, team TeamID, cat CategoryID, puzzle PuzzleID) Award {
	return Award{when, team, cat, puzzle}
}

// ParseAward parses a string log entry into an award.
func ParseAward(s string) (Award, error) {
	ret := Award{}

	s = strings.TrimSpace(s)
	var when int64

	if n, err := fmt.Sscanf(s, "%d %s %s %s", &when, &ret.Team, &ret.Category, &ret.Puzzle); err == io.EOF {
		return ret, fmt.Errorf("malformed award string: only parsed %d fields", n)
	} else if err != nil {
		return ret, err
	}
	ret.When = time.Unix(when, 0)

	return ret, nil
}

// Record returns this award as a list of strings.
// You could use this in a CSV writer, for instance.
func (a Award) Record() []string {
	when := a.When.Unix()
	return []string{
		strconv.FormatInt(when, 10),
		a.Team.String(),
		a.Category.String(),
		a.Puzzle.String(),
	}
}

// String returns a log entry string for an award.
func (a Award) String() string {
	return strings.Join(a.Record(), " ")
}

// Bytes returns a log entry byte slice for an award.
func (a Award) Bytes() []byte {
	return []byte(a.String())
}

// Filename returns a string version of an award suitable for a filesystem
func (a Award) Filename() string {
	return fmt.Sprintf(
		"%d-%s-%s-%s.award",
		a.When.Unix(),
		url.PathEscape(a.Team.String()),
		url.PathEscape(a.Category.String()),
		url.PathEscape(a.Puzzle.String()),
	)
}

// MarshalJSON returns the award event, encoded as a list.
func (a Award) MarshalJSON() ([]byte, error) {
	ao := []interface{}{
		a.When.Unix(),
		a.Team,
		a.Category,
		a.Puzzle,
	}

	return json.Marshal(ao)
}

// UnmarshalJSON decodes the JSON string b.
func (a *Award) UnmarshalJSON(b []byte) error {
	var vals []json.RawMessage

	if err := json.Unmarshal(b, &vals); err != nil {
		return err
	}
	if len(vals) != 4 {
		return errors.New("wrong number of fields")
	}

	var epoch int64
	if err := json.Unmarshal(vals[0], &epoch); err != nil {
		return err
	} else {
		a.When = time.Unix(epoch, 0)
	}

	if err := json.Unmarshal(vals[1], &a.Team); err != nil {
		return err
	}
	if err := json.Unmarshal(vals[2], &a.Category); err != nil {
		return err
	}
	if err := json.Unmarshal(vals[3], &a.Puzzle); err != nil {
		return err
	}

	return nil
}

// Equal returns true if two award events represent the same award.
// Timestamps are ignored in this comparison!
func (a Award) Equal(o Award) bool {
	switch {
	case a.Team != o.Team:
		return false
	case a.Category != o.Category:
		return false
	case a.Puzzle != o.Puzzle:
		return false
	}
	return true
}

// AwardList is a collection of award events.
type AwardList struct {
	awards []Award
}

// NewAwardList creates an AwardList of the provide awards.
func NewAwardList(awards ...Award) *AwardList {
	return &AwardList{awards}
}

// ReadFrom reads al from r, expecting one string award per line.
//
// al will be sorted, and uniqued.
//
// Returns n: the number of entries read.
// Because duplicated entries are removed, n may not be the number of entries stored!
func (al *AwardList) ReadFrom(r io.Reader) (int64, error) {
	counter := new(writecounter.WriteCounter)
	tr := io.TeeReader(r, counter)
	cr := csv.NewReader(tr)
	cr.FieldsPerRecord = 4

	for {
		record, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return counter.Len(), err
		}

		when, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			return counter.Len(), err
		} else {
			award := Award{
				When:     time.Unix(when, 0),
				Team:     TeamID(record[1]),
				Category: CategoryID(record[2]),
				Puzzle:   PuzzleID(record[3]),
			}
			al.awards = append(al.awards, award)
		}
	}

	sort.Stable(al)
	al.awards = al.Unique().awards
	return counter.Len(), nil
}

// WriteTo writes al as strings to w, one per line.
func (al *AwardList) WriteTo(w io.Writer) (int64, error) {
	counter := new(writecounter.WriteCounter)
	tw := io.MultiWriter(w, counter)
	cw := csv.NewWriter(tw)

	for _, a := range al.awards {
		if err := cw.Write(a.Record()); err != nil {
			cw.Flush()
			return counter.Len(), err
		}
	}

	cw.Flush()
	return counter.Len(), nil
}

// Values returns a list of Awards.
func (al *AwardList) Values() []Award {
	if al == nil {
		return nil
	}
	return al.awards
}

// Len returns the number of awards in al.
func (al *AwardList) Len() int {
	return len(al.Values())
}

// Less returns true if i was awarded before j.
func (al *AwardList) Less(i, j int) bool {
	return al.awards[i].When.Before(al.awards[j].When)
}

// Swap positions i and j.
func (al *AwardList) Swap(i, j int) {
	tmp := al.awards[i]
	al.awards[i] = al.awards[j]
	al.awards[j] = tmp
}

// Append to al.
func (al *AwardList) Append(awards ...Award) {
	al.awards = append(al.awards, awards...)
}

// Contains returns true if a is present in al.
func (al *AwardList) Contains(a Award) bool {
	return slices.ContainsFunc(al.awards, a.Equal)
}

// Unique returns a AwardList with duplicates removed.
func (al *AwardList) Unique() *AwardList {
	ret := NewAwardList()
	for _, a := range al.awards {
		if ret.Contains(a) {
			continue
		}
		ret.Append(a)
	}
	return ret
}

func (al *AwardList) MarshalJSON() ([]byte, error) {
	return json.Marshal(al.awards)
}

func (al *AwardList) UnmarshalJSON(p []byte) error {
	return json.Unmarshal(p, &al.awards)
}
