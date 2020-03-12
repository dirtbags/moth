package main

import (
	"testing"
	"sort"
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

	if _, err := ParseAward("bad bad bad 1"); err == nil {
		t.Error("Not throwing error on bad timestamp")
	}
	if _, err := ParseAward("1 bad bad bad"); err == nil {
		t.Error("Not throwing error on bad points")
	}
}

func TestAwardList(t *testing.T) {
  a, _ := ParseAward("1536958399 1a2b3c4d counting 1")
  b, _ := ParseAward("1536958400 1a2b3c4d counting 1")
  c, _ := ParseAward("1536958300 1a2b3c4d counting 1")
  list := AwardList{a, b, c}
  
  if sort.IsSorted(list) {
    t.Error("Unsorted list thinks it's sorted")
  }
  
  sort.Stable(list)
  if (list[0] != c) || (list[1] != a) || (list[2] != b) {
    t.Error("Sorting didn't")
  }
  
  if ! sort.IsSorted(list) {
    t.Error("Sorted list thinks it isn't")
  }
}
