package award

import (
	"sort"
	"testing"
)

func TestAward(t *testing.T) {
	entry := "1536958399 1a2b3c4d counting 10"
	a, err := Parse(entry)
	if err != nil {
		t.Error(err)
		return
	}
	if a.TeamID != "1a2b3c4d" {
		t.Error("TeamID parsed wrong")
	}
	if a.Category != "counting" {
		t.Error("Category parsed wrong")
	}
	if a.Points != 10 {
		t.Error("Points parsed wrong")
	}

	if a.String() != entry {
		t.Error("String conversion wonky")
	}

	b, err := Parse(entry[2:])
	if err != nil {
		t.Error(err)
	}
	if !a.Equal(b) {
		t.Error("Different timestamp events do not compare equal")
	}

	c, err := Parse(entry[:len(entry)-1])
	if err != nil {
		t.Error(err)
	}
	if a.Equal(c) {
		t.Error("Different pount values compare equal")
	}

	ja, err := a.MarshalJSON()
	if err != nil {
		t.Error(err)
	} else if string(ja) != `[1536958399,"1a2b3c4d","counting",10]` {
		t.Error("JSON wrong")
	}

	if _, err := Parse("bad bad bad 1"); err == nil {
		t.Error("Not throwing error on bad timestamp")
	}
	if _, err := Parse("1 bad bad bad"); err == nil {
		t.Error("Not throwing error on bad points")
	}

	if err := b.UnmarshalJSON(ja); err != nil {
		t.Error(err)
	} else if !b.Equal(a) {
		t.Error("UnmarshalJSON didn't work")
	}

	for _, s := range []string{`12`, `"moo"`, `{"a":1}`, `[1 2 3 4]`, `[]`, `[1,"a"]`, `[1,"a","b",4, 5]`} {
		buf := []byte(s)
		if err := a.UnmarshalJSON(buf); err == nil {
			t.Error("Bad unmarshal didn't return error:", s)
		}
	}

}

func TestAwardList(t *testing.T) {
	a, _ := Parse("1536958399 1a2b3c4d counting 1")
	b, _ := Parse("1536958400 1a2b3c4d counting 1")
	c, _ := Parse("1536958300 1a2b3c4d counting 1")
	list := List{a, b, c}

	if sort.IsSorted(list) {
		t.Error("Unsorted list thinks it's sorted")
	}

	sort.Stable(list)
	if (list[0] != c) || (list[1] != a) || (list[2] != b) {
		t.Error("Sorting didn't")
	}

	if !sort.IsSorted(list) {
		t.Error("Sorted list thinks it isn't")
	}
}
