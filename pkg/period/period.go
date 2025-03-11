// Package period provides a RFC3339-compliant period parser.
package period

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/peterhellberg/duration"
)

var ErrDecode = errors.New("unparseable duration")

// Period has a start and end time.
type Period struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Forever contains every time.Time.
var Forever = Period{
	Start: time.Time{},
	End:   time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC),
}

// Parse an RFC3339 period string into a Period.
func Parse(str string) (Period, error) {
	p := Period{}

	parts := strings.Split(str, "/")
	if len(parts) != 2 {
		return p, ErrDecode
	}

	p.Start = time.Time{}
	p.End = time.Time{}

	var startErr, endErr error
	p.Start, startErr = time.Parse(time.RFC3339, parts[0])
	p.End, endErr = time.Parse(time.RFC3339, parts[1])

	if (startErr != nil) && (endErr != nil) {
		return p, ErrDecode
	} else if startErr != nil {
		if dur, err := duration.Parse(parts[0]); err != nil {
			return p, err
		} else {
			p.Start = p.End.Add(-dur)
		}
	} else if endErr != nil {
		if dur, err := duration.Parse(parts[1]); err != nil {
			return p, err
		} else {
			p.End = p.Start.Add(dur)
		}
	}

	return p, nil
}

// Equal returns whether p and q represent the same period.
func (p Period) Equal(q Period) bool {
	return p.Start.Equal(q.Start) && p.End.Equal(q.End)
}

// String returns the RFC3339 "time/time" representation of p.
func (p Period) String() string {
	start := p.Start.Format(time.RFC3339)
	end := p.End.Format(time.RFC3339)
	return fmt.Sprintf("%s/%s", start, end)
}

// Duration returns the computed duration of the period.
func (p Period) Duration() time.Duration {
	return p.End.Sub(p.Start)
}

// Contains returns true if t is contained in the interval described by p.
//
// The interval includes p.Start, but excludes p.End.
func (p Period) Contains(t time.Time) bool {
	if p.Start.Equal(t) {
		return true
	}
	if p.Start.Before(t) && p.End.After(t) {
		return true
	}
	return false
}
