package main

import (
	"testing"
)

func TestAward(t *testing.T) {
	entry := "1536958399 1a2b3c4d counting 1"
	a, err := ParseAward(entry)
	if err != nil {
		t.Error(err)
		return
	}
	if a.TeamId != "1a2b3c4d" {
		t.Error("TeamID parsed wrong")
	}
	if a.Category != "counting" {
		t.Error("Category parsed wrong")
	}
	if a.Points != 1 {
		t.Error("Points parsed wrong")
	}

	if a.String() != entry {
		t.Error("String conversion wonky")
	}

	if ja, err := a.MarshalJSON(); err != nil {
		t.Error(err)
	} else if string(ja) != `[1536958399,"1a2b3c4d","counting",1]` {
		t.Error("JSON wrong")
	}
	
	
	if _, err := ParseAward("bad bad bad 1"); err == nil {
		t.Error("Not throwing error on bad timestamp")
	}
	if _, err := ParseAward("1 bad bad bad"); err == nil {
		t.Error("Not throwing error on bad points")
	}
}
