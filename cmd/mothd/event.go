package main

import (
	"bytes"
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
)

type Event struct {
	When   time.Time
	Name   string
	Team   TeamID
	Cat    CategoryID
	Puzzle PuzzleID
	Extra  []string
}

func NewEvent(when time.Time, name string, team TeamID, cat CategoryID, puzzle PuzzleID, extra ...string) Event {
	return Event{when, name, team, cat, puzzle, extra}
}

func ParseEvent(str string) (e Event, err error) {
	l := NewEventList()
	if n, err := l.ReadFrom(bytes.NewBufferString(str)); err != nil {
		return e, err
	} else if n == 0 {
		return e, fmt.Errorf("short read")
	} else {
		return l.events[0], nil
	}
}

func (e Event) Record() []string {
	record := []string{
		strconv.FormatInt(e.When.Unix(), 10),
		e.Name,
		e.Team.String(),
		e.Cat.String(),
		e.Puzzle.String(),
	}
	record = append(record, e.Extra...)
	return record
}

func (e Event) Filename() string {
	return url.PathEscape(strings.Join(e.Record(), "-")) + ".event"
}

func (e Event) String() string {
	l := NewEventList(e)
	f := new(strings.Builder)
	if n, err := l.WriteTo(f); err != nil {
		return "##ERROR"
	} else if n < 1 {
		return "##SHORT"
	} else {
		return strings.TrimSpace(f.String())
	}
}

func (e Event) Bytes() []byte {
	return []byte(e.String())
}

func (e Event) Equal(b Event) bool {
	return ((e.Name == b.Name) &&
		(e.Team == b.Team) &&
		(e.Cat == b.Cat) &&
		(e.Puzzle == b.Puzzle) &&
		slices.Equal(e.Extra, b.Extra))
}

func (e Event) MarshalJSON() ([]byte, error) {
	rec := []any{
		e.When.Unix(),
		e.Name,
		e.Team,
		e.Cat,
		e.Puzzle,
	}
	for _, x := range e.Extra {
		rec = append(rec, x)
	}
	return json.Marshal(rec)
}

func (e *Event) UnmarshalJSON(p []byte) error {
	rec := make([]any, 0, 7)
	r := bytes.NewReader(p)
	dec := json.NewDecoder(r)
	dec.UseNumber()

	if err := dec.Decode(&rec); err != nil {
		return err
	}
	if len(rec) < 5 {
		return fmt.Errorf("not enough fields: %#q", rec)
	}
	if when, ok := rec[0].(json.Number); !ok {
		return fmt.Errorf("invalid timestamp: %#q", rec[0])
	} else if n, err := when.Int64(); err != nil {
		return err
	} else {
		e.When = time.Unix(n, 0)
	}

	rest := make([]string, len(rec)-1)
	for i, v := range rec[1:] {
		if s, ok := v.(string); !ok {
			return fmt.Errorf("not a string: %#q", v)
		} else {
			rest[i] = s
		}
	}

	e.Name = rest[0]
	e.Team = TeamID(rest[1])
	e.Cat = CategoryID(rest[2])
	e.Puzzle = PuzzleID(rest[3])
	e.Extra = rest[4:]

	return nil
}

type EventList struct {
	events []Event
}

func NewEventList(events ...Event) *EventList {
	return &EventList{events}
}

func (el *EventList) Append(e Event) {
	el.events = append(el.events, e)
}

// Values returns a list of Events.
func (el *EventList) Values() []Event {
	if el == nil {
		return nil
	}
	return el.events
}

// Len returns the number of events in el.
func (el *EventList) Len() int {
	return len(el.Values())
}

// Less returns true if i.When comes before j.When
func (el *EventList) Less(i, j int) bool {
	return el.events[i].When.Before(el.events[j].When)
}

// Swap exchanges the items i and j.
func (el *EventList) Swap(i, j int) {
	tmp := el.events[i]
	el.events[i] = el.events[j]
	el.events[j] = tmp
}

// ReadFrom reads el from r, expecting one string award per line.
//
// el will be sorted.
//
// el will be sorted when err==nil.
func (el *EventList) ReadFrom(r io.Reader) (int64, error) {
	counter := new(WriteCounter)
	tr := io.TeeReader(r, counter)
	cr := csv.NewReader(tr)
	cr.FieldsPerRecord = -1
	for {
		var record []string
		var err error

		record, err = cr.Read()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return counter.Len(), err
		} else if len(record) < 5 {
			return counter.Len(), fmt.Errorf("incomplete record: %#v", record)
		}

		var when int64
		if when, err = strconv.ParseInt(record[0], 10, 64); err != nil {
			return counter.Len(), err
		} else {
			event := Event{
				When:   time.Unix(when, 0),
				Name:   record[1],
				Team:   TeamID(record[2]),
				Cat:    CategoryID(record[3]),
				Puzzle: PuzzleID(record[4]),
				Extra:  record[5:],
			}
			el.events = append(el.events, event)
		}
	}
	sort.Stable(el)
	el.events = el.Unique().events
	return counter.Len(), nil
}

// WriteTo writes el as strings to w, one per line.
func (el *EventList) WriteTo(w io.Writer) (int64, error) {
	counter := new(WriteCounter)
	tw := io.MultiWriter(w, counter)
	cw := csv.NewWriter(tw)

	for _, e := range el.events {
		if err := cw.Write(e.Record()); err != nil {
			cw.Flush()
			return counter.Len(), err
		}
	}
	cw.Flush()
	return counter.Len(), nil
}

// Contains returns true if e is present in al.
func (el *EventList) Contains(e Event) bool {
	return slices.ContainsFunc(el.events, e.Equal)
}

// Unique returns an EventList with duplicates removed.
func (el *EventList) Unique() *EventList {
	ret := NewEventList()
	for _, e := range el.events {
		if ret.Contains(e) {
			continue
		}
		ret.Append(e)
	}
	return ret
}

func (el *EventList) MarshalJSON() ([]byte, error) {
	return json.Marshal(el.events)
}

func (el *EventList) UnmarshalJSON(p []byte) error {
	return json.Unmarshal(p, &el.events)
}
