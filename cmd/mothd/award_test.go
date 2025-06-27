package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"
)

// TestAward is a sample Award.
var TestAward = Award{
	time.Unix(1536958399, 0),
	"1a2b3c4d",
	"counting",
	"10",
}

// TestAwardStr is the string representation of TestAward.
const TestAwardStr = "1536958399 1a2b3c4d counting 10"

func TestAwardGeneral(t *testing.T) {
	a, err := ParseAward(TestAwardStr)
	if err != nil {
		t.Error(err)
		return
	}
	if !a.Equal(TestAward) {
		t.Error("Equality wonky")
	}
	if a.Team != TestAward.Team {
		t.Error("Team parsed wrong")
	}
	if a.Category != TestAward.Category {
		t.Error("Category parsed wrong")
	}
	if a.Puzzle != TestAward.Puzzle {
		t.Error("Puzzle ID parsed wrong")
	}

	if a.String() != TestAwardStr {
		t.Error("String conversion wonky")
	}
	if string(a.Bytes()) != TestAwardStr {
		t.Error("Bytes conversion wonky")
	}

	if fn := a.Filename(); fn != "1536958399-1a2b3c4d-counting-10.award" {
		t.Error("Bad filename:", fn)
	}

	b, err := ParseAward(TestAwardStr[2:])
	if err != nil {
		t.Error(err)
	}
	if !a.Equal(b) {
		t.Error("Different timestamp events do not compare equal")
	}

	if c, err := ParseAward(TestAwardStr[:len(TestAwardStr)-1]); err != nil {
		t.Error(err)
	} else if a.Equal(c) {
		t.Error("Different puzzle IDs compare equal")
	}

	if c, err := ParseAward(fmt.Sprintf("%d nope %s %s", TestAward.When.Unix(), TestAward.Category, TestAward.Puzzle)); err != nil {
		t.Error(err)
	} else if a.Equal(c) {
		t.Error("Different teams compare equal")
	}
}

func TestAwardBad(t *testing.T) {
	if _, err := ParseAward("bad bad bad 1"); err == nil {
		t.Error("Not throwing error on bad timestamp")
	}
	if _, err := ParseAward("1 bad bad"); err == nil {
		t.Error("Not throwing error on too few fields")
	}

	var b Award
	if err := b.UnmarshalJSON([]byte{}); err == nil {
		t.Error("Unmarshal nil should have errored")
	}
	if err := b.UnmarshalJSON([]byte("")); err == nil {
		t.Error("Unmarshal empty string should have errored")
	}

	if err := b.UnmarshalJSON([]byte(`[false,"team","cat","puzzle"]`)); err == nil {
		t.Error("Bad JSON didn't fail")
	}
	if err := b.UnmarshalJSON([]byte(`[0,"team",false,"puzzle"]`)); err == nil {
		t.Error("Bad JSON didn't fail", b)
	}
	if err := b.UnmarshalJSON([]byte(`[0,"team",false,12]`)); err == nil {
		t.Error("Bad JSON didn't fail", b)
	}

	if err := b.UnmarshalJSON([]byte(`[0,"team","cat"]`)); err == nil {
		t.Error("Short JSON didn't fail", b)
	}
	if err := b.UnmarshalJSON([]byte(`[0,"team","cat","puzzle","toomany"]`)); err == nil {
		t.Error("Long JSON didn't fail", b)
	}
}

func TestAwardJson(t *testing.T) {
	ja, err := TestAward.MarshalJSON()
	if err != nil {
		t.Error(err)
	} else if string(ja) != `[1536958399,"1a2b3c4d","counting","10"]` {
		t.Error("JSON wrong", string(ja))
	}

	var b Award
	if err := b.UnmarshalJSON(ja); err != nil {
		t.Error(err)
	} else if !b.Equal(TestAward) {
		t.Errorf("UnmarshalJSON didn't work: %#q: %#v != %#v", ja, b.String(), TestAward.String())
	}

	var c Award
	if err := json.Unmarshal(ja, &c); err != nil {
		t.Error(err)
	} else if !c.Equal(TestAward) {
		t.Errorf("json.Unmarshal didn't work: %#q: %#v != %#v", ja, c.String(), TestAward.String())
	}

	for _, s := range []string{`12`, `"moo"`, `{"a":1}`, `[1 2 3 4]`, `[]`, `[1,"a"]`, `[1,"a","b",10]`, `[1,"a","b",4, 5]`} {
		buf := []byte(s)
		if err := TestAward.UnmarshalJSON(buf); err == nil {
			t.Error("Bad unmarshal didn't return error:", s)
		}
	}

}

func TestAwardListBasic(t *testing.T) {
	a, _ := ParseAward("1536958399 1a2b3c4d counting 1")
	b, _ := ParseAward("1536958400 1a2b3c4d counting 1")
	c, _ := ParseAward("1536958300 1a2b3c4d counting 1")
	z, _ := ParseAward("1536958300 1a2b3c4d blert 1")
	list := NewAwardList(a, b, c)

	if sort.IsSorted(list) {
		t.Error("Unsorted list thinks it's sorted")
	}

	sort.Stable(list)
	if (list.awards[0] != c) || (list.awards[1] != a) || (list.awards[2] != b) {
		t.Error("Sort didn't")
	}

	if !sort.IsSorted(list) {
		t.Error("Sorted list thinks it isn't")
	}

	if !list.Contains(a) {
		t.Error("Contains didn't find member")
	}
	if list.Contains(z) {
		t.Error("Contains found membership for non-member")
	}

	u := list.Unique()
	if u.Len() != 1 {
		t.Error("Unique list has wrong number of elements:", u)
	}
}

func TestAwardListIo(t *testing.T) {
	r := strings.NewReader(
		strings.Join(
			[]string{
				"1536958400,1a2b3c4d,counting,1",
				"1536958399,1a2b3c4d,counting,1", // Earliest
				"1536958401,1a2b3c4d,counting,1",
				"1536958301,1a2b3c4d,blert,1",
			},
			"\n",
		),
	)

	rlen := r.Len()
	var awards AwardList
	n, err := awards.ReadFrom(r)
	if err != nil {
		t.Error(err)
	} else if n != int64(rlen) {
		t.Error("Wrong length:", n)
	}
	if awards.Len() != 2 {
		t.Error("Read awards wrong length:", awards)
	}

	f := new(strings.Builder)
	sortedUniquedString := "1536958301,1a2b3c4d,blert,1\n1536958399,1a2b3c4d,counting,1\n"
	if n, err := awards.WriteTo(f); err != nil {
		t.Error(err)
	} else if n != int64(len(sortedUniquedString)) {
		t.Error("Wrong length:", n)
	}
	if f.String() != sortedUniquedString {
		t.Error("WriteAwards output incorrect")
		t.Log("got:", f.String())
		t.Log("wanted:", sortedUniquedString)
	}

	tmp := new(bytes.Buffer)
	if n, err := awards.WriteTo(tmp); err != nil {
		t.Error(err)
	} else if n != int64(len(sortedUniquedString)) {
		t.Error("Wrong length:", n)
	} else if tmp.String() != sortedUniquedString {
		t.Error("WriteTo screwed up")
		t.Log(tmp.String())
	}
}

func TestAwardListParseFailure(t *testing.T) {
	f := bytes.NewBufferString("bleargh merf fleart")

	awards := NewAwardList()
	if _, err := awards.ReadFrom(f); err == nil {
		t.Error("Shouldn't have parsed")
	}
}
