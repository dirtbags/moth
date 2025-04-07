package main

import (
	"bytes"
	"encoding/json"
	"sort"
	"testing"
	"time"
)

var TestEvent1 = Event{time.Unix(120, 0), "name", "team", "cat", "puzzle", nil}
var TestEvent2 = Event{time.Unix(240, 0), "name", "team", "cat", "puzzle", []string{"extra1", "extra2"}}

func TestEventBasic(t *testing.T) {
	if fn := TestEvent1.Filename(); fn != "120-name-team-cat-puzzle.event" {
		t.Error("Filename wrong", fn)
	}
	if s := TestEvent1.String(); s != "120,name,team,cat,puzzle" {
		t.Error("String wrong", s)
	}
	if s := TestEvent2.String(); s != "240,name,team,cat,puzzle,extra1,extra2" {
		t.Error("String wrong", s)
	}

	if te2, err := ParseEvent(TestEvent2.String()); err != nil {
		t.Error(err)
	} else if !TestEvent2.Equal(te2) {
		t.Error("Equality failed")
	}

	if _, err := ParseEvent("x,x,x,x,x"); err == nil {
		t.Error("Bad parse")
	}
	if _, err := ParseEvent("x,x,x"); err == nil {
		t.Error("Bad parse")
	}
	if _, err := ParseEvent(""); err == nil {
		t.Error("Bad parse")
	}
}

func TestEventJSON(t *testing.T) {
	var e Event
	if buf, err := json.Marshal(TestEvent2); err != nil {
		t.Error(err)
	} else if string(buf) != `[240,"name","team","cat","puzzle","extra1","extra2"]` {
		t.Errorf("JSON Marshal wrong: %#q", buf)
	} else if err := json.Unmarshal(buf, &e); err != nil {
		t.Error(err)
	} else if !e.Equal(TestEvent2) {
		t.Errorf("Unmarshal wrong: %#q", e)
	}
}

func TestEventListBasic(t *testing.T) {
	list := NewEventList(TestEvent2, TestEvent1)

	sort.Stable(list)
	if !TestEvent1.Equal(list.events[0]) {
		t.Error("Sort didn't")
	}

	buf := new(bytes.Buffer)

	// Do this twice to test uniqueness
	for i := 0; i < 2; i++ {
		if n, err := list.WriteTo(buf); err != nil {
			t.Error(err)
		} else if n != 64 {
			t.Error("WriteTo wrote the wrong number of bytes:", n)
		}

		if n, err := list.ReadFrom(buf); err != nil {
			t.Error(err)
		} else if n != 64 {
			t.Error("ReadFrom read the wrong number of bytes:", n)
		}
	}
}
