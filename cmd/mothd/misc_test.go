package main

import (
	"testing"
	"time"
)

// Clock returns consistent time for replayable tests
type TestClock struct {
	when time.Time
}

// Now advances the clock by one second, and returns the current time
func (c *TestClock) Now() time.Time {
	c.when = c.when.Add(1 * time.Second)
	return c.when
}

func NewTestClock() *TestClock {
	return &TestClock{time.Date(2024, 7, 5, 22, 34, 10, 0, time.UTC)}
}


// TimeZero is epoch time 0
var TimeZero = time.Unix(0, 0)

func TestTestClock(t *testing.T) {
	c := NewTestClock()
	now := c.Now()
	if !c.Now().Equal(now.Add(1 * time.Second)) {
		t.Error("The test clock is broken")
	}
}
