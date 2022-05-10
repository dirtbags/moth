package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type zipCategory struct {
	zip.Reader
	io.Closer
	mtime time.Time
}

// Mothballs provides a collection of active mothball files (puzzle categories)
type Mothballs struct {
	fs.FS
	categories   map[string]zipCategory
	categoryLock *sync.RWMutex
}

// NewMothballs returns a new Mothballs structure backed by the provided directory
func NewMothballs(fsys fs.FS) *Mothballs {
	return &Mothballs{
		FS:           fsys,
		categories:   make(map[string]zipCategory),
		categoryLock: new(sync.RWMutex),
	}
}

func (m *Mothballs) getCat(cat string) (zipCategory, bool) {
	m.categoryLock.RLock()
	defer m.categoryLock.RUnlock()
	ret, ok := m.categories[cat]
	return ret, ok
}

// Open returns an fs.File corresponding to the filename in a puzzle's category and points
func (m *Mothballs) Open(cat string, points int, filename string) (fs.File, time.Time, error) {
	zc, ok := m.getCat(cat)
	if !ok {
		return nil, time.Time{}, fmt.Errorf("no such category: %s", cat)
	}

	f, err := zc.Open(fmt.Sprintf("%d/%s", points, filename))
	if err != nil {
		return nil, time.Time{}, err
	}

	fInfo, err := f.Stat()
	return f, fInfo.ModTime(), err
}

// Inventory returns the list of current categories
func (m *Mothballs) Inventory() []Category {
	m.categoryLock.RLock()
	defer m.categoryLock.RUnlock()
	categories := make([]Category, 0, 20)
	for cat, zfs := range m.categories {
		pointsList := make([]int, 0, 20)
		pf, err := zfs.Open("puzzles.txt")
		if err != nil {
			// No puzzles = no category
			continue
		}
		scanner := bufio.NewScanner(pf)
		for scanner.Scan() {
			line := scanner.Text()
			if pointval, err := strconv.Atoi(line); err != nil {
				log.Printf("Reading points for %s: %s", cat, err.Error())
			} else {
				pointsList = append(pointsList, pointval)
			}
		}
		sort.Ints(pointsList)
		categories = append(categories, Category{cat, pointsList})
	}
	return categories
}

// CheckAnswer returns an error if the provided answer is in any way incorrect for the given category and points
func (m *Mothballs) CheckAnswer(cat string, points int, answer string) (bool, error) {
	zfs, ok := m.getCat(cat)
	if !ok {
		return false, fmt.Errorf("no such category: %s", cat)
	}

	af, err := zfs.Open("answers.txt")
	if err != nil {
		return false, fmt.Errorf("no answers.txt file")
	}
	defer af.Close()

	needle := fmt.Sprintf("%d %s", points, answer)
	scanner := bufio.NewScanner(af)
	for scanner.Scan() {
		if scanner.Text() == needle {
			return true, nil
		}
	}

	return false, nil
}

func (m *Mothballs) newZipCategory(f fs.File) (zipCategory, error) {
	var zrc *zip.Reader
	var err error
	var closer io.ReadCloser = f
	var zipCat zipCategory

	fi, err := f.Stat()
	if err != nil {
		return zipCat, err
	}
	zipCat.mtime = fi.ModTime()

	switch r := f.(type) {
	case io.ReaderAt:
		zrc, err = zip.NewReader(r, fi.Size())
	default:
		log.Println("Does not implement io.ReaderAt, buffering in RAM:", r)
		buf := new(bytes.Buffer)
		size, err := io.Copy(buf, f)
		if err != nil {
			return zipCat, err
		}
		f.Close()
		reader := bytes.NewReader(buf.Bytes())
		zrc, err = zip.NewReader(reader, size)
		closer = io.NopCloser(reader)
	}
	if err != nil {
		return zipCat, err
	}
	zipCat.Reader = *zrc
	zipCat.Closer = closer
	return zipCat, nil
}

// refresh refreshes internal state.
// It looks for changes to the directory listing, and caches any new mothballs.
func (m *Mothballs) refresh() {
	m.categoryLock.Lock()
	defer m.categoryLock.Unlock()

	// Any new categories?
	files, err := fs.ReadDir(m.FS, "/")
	if err != nil {
		log.Println("Error listing mothballs:", err)
		return
	}
	found := make(map[string]bool)
	for _, f := range files {
		filename := f.Name()
		if !strings.HasSuffix(filename, ".mb") {
			continue
		}
		categoryName := strings.TrimSuffix(filename, ".mb")
		found[categoryName] = true

		reopen := false
		if existingMothball, ok := m.categories[categoryName]; !ok {
			reopen = true
		} else if si, err := fs.Stat(m.FS, filename); err != nil {
			log.Println(err)
		} else if si.ModTime().After(existingMothball.mtime) {
			existingMothball.Close()
			delete(m.categories, categoryName)
			reopen = true
		}

		if reopen {
			if f, err := m.FS.Open(filename); err != nil {
				log.Println(err)
			} else if zipCat, err := m.newZipCategory(f); err != nil {
				log.Println(err)
			} else {
				m.categories[categoryName] = zipCat
				log.Println("Adding category:", categoryName)
			}
		}
	}

	// Delete anything in the list that wasn't found
	for categoryName, zc := range m.categories {
		if !found[categoryName] {
			zc.Close()
			delete(m.categories, categoryName)
			log.Println("Removing category:", categoryName)
		}
	}
}

// Mothball just returns an error
func (m *Mothballs) Mothball(cat string, w io.Writer) error {
	return fmt.Errorf("refusing to repackage a compiled mothball")
}

// Maintain performs housekeeping for Mothballs.
func (m *Mothballs) Maintain(updateInterval time.Duration) {
	m.refresh()
	for range time.NewTicker(updateInterval).C {
		m.refresh()
	}
}
