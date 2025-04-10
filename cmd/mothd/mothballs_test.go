package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

type TestMothballStore struct {
	*MothballStore
}

// NewTestMothballStore creates a MothballStore from fsys, by generating mothballs in memory
func NewTestMothballStore(fsys fs.FS) (*TestMothballStore, error) {
	outfs := make(fstest.MapFS)
	dents, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}
	for _, dent := range dents {
		name := dent.Name()
		if !dent.IsDir() {
			continue
		}

		info, err := dent.Info()
		if err != nil {
			return nil, err

		}
		catFS, err := fs.Sub(fsys, name)
		if err != nil {
			return nil, err
		}

		zf := new(bytes.Buffer)
		zw := zip.NewWriter(zf)
		if err := zw.AddFS(catFS); err != nil {
			return nil, err
		}
		zw.Close()

		mbname := fmt.Sprintf("%s.mb", name)
		outfs[mbname] = &fstest.MapFile{
			Data:    zf.Bytes(),
			Mode:    0644,
			ModTime: info.ModTime(),
		}
	}

	s := &TestMothballStore{
		NewMothballStore(outfs),
	}
	return s, nil
}

func (s *TestMothballStore) Should(t *testing.T, exists bool, cat Category, names ...string) {
	t.Helper()
	for _, name := range names {
		var found bool

		name = path.Clean("/" + name)
		if _, err := fs.Stat(cat, name); errors.Is(err, fs.ErrNotExist) {
			found = false
		} else if err != nil {
			t.Error(err)
			continue
		} else {
			found = true
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

// TestdataCategoryStoreOpen zips up each directory in testdata/$name,
// and returns a TestMothballStore containing the resulting mothballs.
func TestdataCategoryStoreOpen(t *testing.T, name string) *TestMothballStore {
	t.Helper()

	categoriesPath := path.Join("testdata", name)
	categoriesFS := os.DirFS(categoriesPath)
	s, err := NewTestMothballStore(categoriesFS)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestdataCategoryStore(t *testing.T) *TestMothballStore {
	t.Helper()
	s := TestdataCategoryStoreOpen(t, "categories")
	if err := s.Refresh(TimeZero); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestMothballsPuzzlesLeakage(t *testing.T) {
	s := TestdataCategoryStore(t)
	c, err := s.Category("cat0")
	if err != nil {
		t.Fatal(err)
	}

	name := "/__MOTH/answers.txt"
	req := httptest.NewRequest("GET", name, nil)
	w := httptest.NewRecorder()
	http.ServeFileFS(w, req, c, name)

	if w.Result().StatusCode != 404 {
		t.Error("should have been 404")
	}
	body, _ := io.ReadAll(w.Body)
	if strings.Contains(string(body), "two") {
		t.Error("404 is leaking the page")
	}
}

func testExistence(t *testing.T, catid CategoryID, s *TestMothballStore) {
	t.Helper()
	cat, err := s.Category(catid)
	if err != nil {
		t.Fatal(err)
	}

	s.Should(t, true, cat,
		"1/index.html",
		"two/index.html",
	)

	s.Should(t, false, cat,
		"__MOTH/answers.txt",
		"__MOTH//answers.txt",
		"/__MOTH/answers.txt",
		"//__MOTH/answers.txt",
		"__MOTH/answers.txt",
		"__MOTH\\answers.txt",
		"1/../__MOTH/answers.txt",
	)

	if !slices.Contains(cat.Answers("1"), "hello") {
		t.Error("correct answer not listed")
	}

	if slices.Contains(cat.Answers("0"), "hello") {
		t.Error("answer listed for wrong puzzle")
	}
}

func TestMothballsPuzzlesUnlocked(t *testing.T) {
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
func TestMothballsPuzzlesTraversal(t *testing.T) {
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

func TestMothballsPuzzlesBasicMetadata(t *testing.T) {
	s := TestdataCategoryStore(t)
	cat0, err := s.Category("cat0")
	if err != nil {
		t.Fatal(err)
	}
	s.Should(t, false, cat0, "__MOTH/answers.txt")
}

// ZipDirectory returns the bytes of a zip file containing everything under path.
func ZipDirectory(path string) ([]byte, error) {
	zf := new(bytes.Buffer)
	zw := zip.NewWriter(zf)
	if err := zw.AddFS(os.DirFS(path)); err != nil {
		return nil, err
	}
	zw.Close()
	return zf.Bytes(), nil

}

func TestMothballsStateStoreZipBasic(t *testing.T) {
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
	cat0mb, err := ZipDirectory("testdata/categories/cat0")
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
	cat0mb, err := ZipDirectory("testdata/categories/cat0")
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
