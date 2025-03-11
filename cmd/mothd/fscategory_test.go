package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

type TestFSCategoryStore struct {
	*FSCategoryStore
}

func TestdataCategoryStore(t *testing.T) *TestFSCategoryStore {
	s := &TestFSCategoryStore{
		&FSCategoryStore{
			fsys: os.DirFS("testdata/categories"),
		},
	}
	if err := s.Refresh(TimeZero); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestFSPuzzlesLeakage(t *testing.T) {
	s := TestdataCategoryStore(t)
	c, err := s.Category("cat0")
	if err != nil {
		t.Fatal(err)
	}

	name := "/__MOTH/answers.txt"
	req := httptest.NewRequest("GET", name, nil)
	w := httptest.NewRecorder()
	c.ServeFile(w, req, name)

	if w.Result().StatusCode != 404 {
		t.Error("should have been 404")
	}
	body, _ := io.ReadAll(w.Body)
	if strings.Contains(string(body), "two") {
		t.Error("404 is leaking the page")
	}
}

func (s *TestFSCategoryStore) Should(t *testing.T, exists bool, cat Category, names ...string) {
	t.Helper()
	for _, name := range names {
		name = path.Clean("/" + name)
		req := httptest.NewRequest("GET", name, nil)
		w := httptest.NewRecorder()
		cat.ServeFile(w, req, name)

		var found bool
		switch w.Result().StatusCode {
		case 200:
			found = true
		case 404:
			found = false
		default:
			body, _ := io.ReadAll(w.Body)
			t.Errorf("%s: unexpected return code: %v", name, w.Result().Status)
			t.Log("  body:", string(body))
			continue
		}

		if found != exists {
			s := "should"
			if !exists {
				s = s + " not"
			}
			t.Errorf("%s: %s exist", name, s)
		}
	}
}

func testExistence(t *testing.T, name CategoryID, s *TestFSCategoryStore) {
	t.Helper()
	cat0, err := s.Category(name)
	if err != nil {
		t.Fatal(err)
	}

	s.Should(t, true, cat0,
		"1/index.html",
		"two/index.html",
	)

	s.Should(t, false, cat0,
		"__MOTH/answers.txt",
		"__MOTH//answers.txt",
		"/__MOTH/answers.txt",
		"//__MOTH/answers.txt",
		"__MOTH/answers.txt",
		"__MOTH\\answers.txt",
		"1/../__MOTH/answers.txt",
	)

	if !slices.Contains(cat0.Answers("1"), "hello") {
		t.Error("correct answer not listed")
	}

	if slices.Contains(cat0.Answers("0"), "hello") {
		t.Error("answer listed for wrong puzzle")
	}
}

func TestFSPuzzlesUnlocked(t *testing.T) {
	s := TestdataCategoryStore(t)
	cat, err := s.Category("cat0")
	if err != nil {
		t.Fatal(err)
	}

	if cat.Unlocked() {
		t.Error("Should not be unlocked")
	}
	puzzles := cat.Puzzles()
	if len(puzzles) != 2 {
		t.Error("Puzzles wrong length:", puzzles)
	}

}
func TestFSPuzzlesDir(t *testing.T) {
	s := TestdataCategoryStore(t)

	if err := fstest.TestFS(s.fsys, "cat0/1/index.html"); err != nil {
		t.Fatal(err)
	}

	testExistence(t, "cat0", s)
}

func TestFSPuzzlesZip(t *testing.T) {
	zf := new(bytes.Buffer)
	zw := zip.NewWriter(zf)
	if err := zw.AddFS(os.DirFS("testdata/categories/cat0")); err != nil {
		t.Fatal(err)
	}
	zw.Close()

	s := &TestFSCategoryStore{
		&FSCategoryStore{
			fsys: fstest.MapFS{
				"cat0.mb": &fstest.MapFile{
					Data:    zf.Bytes(),
					Mode:    0644,
					ModTime: TimeZero,
					Sys:     nil,
				},
			},
		},
	}
	if err := s.Refresh(TimeZero); err != nil {
		t.Fatal(err)
	}

	testExistence(t, "cat0.mb", s)
}

func TestFSPuzzlesTraversal(t *testing.T) {
	s := TestdataCategoryStore(t)

	cat0, err := s.Category("cat0")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := s.Category("."); err == nil {
		t.Error("should be an error")
	}
	if _, err := s.Category(".."); err == nil {
		t.Error("should be an error")
	}

	s.Should(t, false, cat0,
		"../README.md",
		"1/../../README.md",
	)
}

func TestFSPuzzlesBasicMetadata(t *testing.T) {
	s := TestdataCategoryStore(t)
	cat0, err := s.Category("cat0")
	if err != nil {
		t.Fatal(err)
	}
	s.Should(t, false, cat0, "__MOTH/answers.txt")
}

func TestFSPuzzlesBasicBad(t *testing.T) {
	s := &FSCategoryStore{
		fsys: os.DirFS("testdata/bad-categories"),
	}

	if err := s.Refresh(TimeZero); err == nil {
		t.Error("Bad categories refresh shouldn't be okay")
	}

}

func ZipFSCategory(path string) ([]byte, error) {
	zf := new(bytes.Buffer)
	zw := zip.NewWriter(zf)
	if err := zw.AddFS(os.DirFS(path)); err != nil {
		return nil, err
	}
	zw.Close()
	return zf.Bytes(), nil

}

func TestFSStateStoreZipBasic(t *testing.T) {
	clock := NewTestClock()

	cat0mb, err := ZipFSCategory("testdata/categories/cat0")
	if err != nil {
		t.Fatal(err)
	}
	mapfs := fstest.MapFS{}
	mapfs["test.mb"] =  &fstest.MapFile{
		Data: cat0mb,
		Mode:    0644,
		ModTime: clock.Now(),
		Sys:     nil,
	}

	s := &TestFSCategoryStore{
		&FSCategoryStore{
			fsys: mapfs,
		},
	}
	s.Refresh(clock.Now())

	cat, err := s.Category("test.mb")
	if err != nil {
		t.Fatal(err)
	}
	if puzz := cat.Puzzles(); puzz[1] != "two" {
		t.Error("Second puzzle wrong")
	}

	// Now drop a new mothball
	pategorymb, err := ZipFSCategory("testdata/categories/pategory")
	if err != nil {
		t.Fatal(err)
	}
	mapfs["test.mb"] = &fstest.MapFile{
		Data: pategorymb,
		Mode: 0644,
		ModTime: clock.Now(),
		Sys: nil,
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

	cat2, err := s.Category("test.mb")
	if err != nil {
		t.Fatal(err)
	}
	if puzz := cat2.Puzzles(); puzz[1] != "2" {
		t.Error("Second puzzle wrong", puzz)
	}
}

func readAnswers(fsys fs.FS) error {
	buf, err := fs.ReadFile(fsys, "__MOTH/answers.txt")
	if err != nil {
		return err
	}
	if len(buf) == 0 {
		return fmt.Errorf("zero-length answers.txt")
	}
	return nil
}

/*
* Below are some benchmarks I made to test out whether our current strategy of caching open ZipFiles is actually a performance improvement over just opening them every time, or using the bare filesystem.
*
* It is.
*
*     BenchmarkZipfileOperations-12          	   25963	     39200 ns/op
*     BenchmarkZipfileCachedOperations-12    	  157394	      7601 ns/op
*     BenchmarkOsfileOperations-12           	   56937	     20131 ns/op
*/

func BenchmarkZipfileOperations(b *testing.B) {
	cat0mb, err := ZipFSCategory("testdata/categories/cat0")
	if err != nil {
		b.Fatal(err)
	}
	catdir := b.TempDir()
	cat0fn := filepath.Join(catdir, "cat0.mb")
	if err := os.WriteFile(cat0fn, cat0mb, 0644); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for range b.N {
		zfs, err := zip.OpenReader(cat0fn)
		if err != nil {
			b.Fatal(err)
		}
		err = readAnswers(zfs)
		zfs.Close()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkZipfileCachedOperations(b *testing.B) {
	cat0mb, err := ZipFSCategory("testdata/categories/cat0")
	if err != nil {
		b.Fatal(err)
	}
	catdir := b.TempDir()
	cat0fn := filepath.Join(catdir, "cat0.mb")
	if err := os.WriteFile(cat0fn, cat0mb, 0644); err != nil {
		b.Fatal(err)
	}

	zfs, err := zip.OpenReader(cat0fn)
	if err != nil {
		b.Fatal(err)
	}
	defer zfs.Close()

	b.ResetTimer()

	for range b.N {
		if err := readAnswers(zfs); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOsfileOperations(b *testing.B) {
	for range b.N {
		osfs := os.DirFS("testdata/categories/cat0")
		err := readAnswers(osfs)
		if err != nil {
			b.Fatal(err)
		}
	}
}
