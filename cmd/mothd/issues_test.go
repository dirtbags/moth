package main

import (
	"archive/zip"
	"bytes"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

// TestIssue156: if a team has an entry in teams/, but isn't in teamids.txt, we should still let them log in.
func TestIssue156(t *testing.T) {
	h := NewTestContext(t, "/")

	if err := h.WriteStateFile("test/teams/bloop", []byte{}); err != nil {
		t.Fatal(err)
	}
	h.Refresh()

	h.TestRequestAuth("GET", "http://example.com/test/puzzles/cat0/1/", nil, "", "bloop", 200)
}

// TestIssue222: v5 didn't recognize when I uploaded a newer .mb file
func TestIssue222(t *testing.T) {
	clock := NewTestClock()

	cat0mb, err := ZipDirectory("testdata/categories/cat0")
	if err != nil {
		t.Fatal(err)
	}
	mapfs := fstest.MapFS{}
	mapfs["test.mb"] = &fstest.MapFile{
		Data:    cat0mb,
		Mode:    0644,
		ModTime: clock.Now(),
		Sys:     nil,
	}

	s := &TestMothballStore{
		NewMothballStore(mapfs),
	}
	s.Refresh(clock.Now())

	cat, err := s.Category("test")
	if err != nil {
		t.Fatal(err)
	}
	if puzz := cat.Puzzles(); puzz[1] != "two" {
		t.Error("Second puzzle wrong")
	}

	// Place new contents into the same mothball name:
	// the next refresh should begin using the new contents.
	pategorymb, err := ZipDirectory("testdata/categories/pategory")
	if err != nil {
		t.Fatal(err)
	}
	mapfs["test.mb"] = &fstest.MapFile{
		Data:    pategorymb,
		Mode:    0644,
		ModTime: clock.Now(),
		Sys:     nil,
	}
	s.Refresh(clock.Now())

	{
		zfs, err := zip.NewReader(bytes.NewReader(pategorymb), int64(len(pategorymb)))
		if err != nil {
			t.Fatal(err)
		}
		if err := fstest.TestFS(zfs, "__MOTH/answers.txt", "__MOTH/puzzles.txt", "1/index.html", "2/index.html"); err != nil {
			t.Fatal(err)
		}

	}

	cat2, err := s.Category("test")
	if err != nil {
		t.Fatal(err)
	}
	if puzz := cat2.Puzzles(); puzz[1] != "2" {
		t.Error("Second puzzle wrong", puzz)
	}
}

// TestIssue239: seeker can't seek
func TestIssue239(t *testing.T) {
	h := NewTestContext(t, "/")
	request := httptest.NewRequest("GET", "http://example.com/test/puzzles/cat0/1/", nil)
	request.SetBasicAuth("test-team", "test-team")
	request.Header.Set("Range", "bytes=10-1023")
	h.TestServe(request, 206)
}
