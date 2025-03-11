package period

import (
	"encoding/json"
	"testing"
	"time"
)

const Canonical = "2024-07-25T19:35:14-06:00/2024-07-26T01:35:14-06:00"
const CanonicalJSON = `{"start":"2024-07-25T19:35:14-06:00","end":"2024-07-26T01:35:14-06:00"}`

var TestVectors = []string{
	Canonical,
	"2024-07-25T19:35:14-06:00/PT6H",
	"PT6H/2024-07-26T01:35:14-06:00",
}

func ExpectFail(t *testing.T, str string) {
	if _, err := Parse(str); err == nil {
		t.Error("Parse should have failed:", str)
	}
}

func TestParseErrors(t *testing.T) {
	ExpectFail(t, "")
	ExpectFail(t, "/")
	ExpectFail(t, "2024-07-25T19:35:14-06:00/")
	ExpectFail(t, "/2024-07-25T19:35:14-06:00")
	ExpectFail(t, "bargle")
	ExpectFail(t, "bargle/bargle")
	ExpectFail(t, "2024-07-25T19:35:14-06:00/3h")
	ExpectFail(t, "3h/2024-07-25T19:35:14-06:00")
	ExpectFail(t, "PT3H/PT3H")
}

func TestContains(t *testing.T) {
	p, _ := Parse(Canonical)
	if !p.Contains(p.Start) {
		t.Error("Period does not contain p.Start")
	}
	if p.Contains(p.End) {
		t.Error("Period contains p.End")
	}
	if !p.Contains(p.Start.Add(1 * time.Minute)) {
		t.Error("Period does not contain inner time")
	}
	if p.Contains(p.Start.Add(-1 * time.Minute)) {
		t.Error("Period contains earlier time")
	}
	if p.Contains(p.End.Add(1 * time.Minute)) {
		t.Error("Period contains later time")
	}
}

func TestDuration(t *testing.T) {
	p, _ := Parse(Canonical)
	if dur := p.Duration(); dur != 6*time.Hour {
		t.Error("Duration returned the wrong value:", dur.String())
	}
}

func TestString(t *testing.T) {
	for _, vec := range TestVectors {
		p, err := Parse(vec)
		if err != nil {
			t.Error(err)
			break
		}

		if s := p.String(); s != Canonical {
			t.Error("String() wrong:", s)
		}
	}
}

func TestJSON(t *testing.T) {
	p, err := Parse(Canonical)
	if err != nil {
		t.Fatal(err)
	}

	var q Period
	if buf, err := json.Marshal(p); err != nil {
		t.Error(err)
	} else if string(buf) != CanonicalJSON {
		t.Error("JSON Marshal wrong:", string(buf))
	} else if err := json.Unmarshal(buf, &q); err != nil {
		t.Error(err)
	} else if !p.Equal(q) {
		t.Error("JSON Marshal/Unmarshal not equal")
	}

}
