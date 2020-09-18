package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/afero/zipfs"
)

type zipCategory struct {
	afero.Fs
	io.Closer
}

// Mothballs provides a collection of active mothball files (puzzle categories)
type Mothballs struct {
	afero.Fs
	categories   map[string]zipCategory
	categoryLock *sync.RWMutex
}

// NewMothballs returns a new Mothballs structure backed by the provided directory
func NewMothballs(fs afero.Fs) *Mothballs {
	return &Mothballs{
		Fs:           fs,
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

// Open returns a ReadSeekCloser corresponding to the filename in a puzzle's category and points
func (m *Mothballs) Open(cat string, points int, filename string) (ReadSeekCloser, time.Time, error) {
	zc, ok := m.getCat(cat)
	if !ok {
		return nil, time.Time{}, fmt.Errorf("No such category: %s", cat)
	}

	f, err := zc.Open(fmt.Sprintf("content/%d/%s", points, filename))
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
		return false, fmt.Errorf("No such category: %s", cat)
	}

	af, err := zfs.Open("answers.txt")
	if err != nil {
		return false, fmt.Errorf("No answers.txt file")
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

// refresh refreshes internal state.
// It looks for changes to the directory listing, and caches any new mothballs.
func (m *Mothballs) refresh() {
	m.categoryLock.Lock()
	defer m.categoryLock.Unlock()

	// Any new categories?
	files, err := afero.ReadDir(m.Fs, "/")
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

		if _, ok := m.categories[categoryName]; !ok {
			f, err := m.Fs.Open(filename)
			if err != nil {
				log.Println(err)
				continue
			}

			fi, err := f.Stat()
			if err != nil {
				f.Close()
				log.Println(err)
				continue
			}

			zrc, err := zip.NewReader(f, fi.Size())
			if err != nil {
				f.Close()
				log.Println(err)
				continue
			}

			m.categories[categoryName] = zipCategory{
				Fs:     zipfs.New(zrc),
				Closer: f,
			}

			log.Println("Adding category:", categoryName)
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
func (m *Mothballs) Mothball(cat string) (*bytes.Reader, error) {
	return nil, fmt.Errorf("Can't repackage a compiled mothball")
}

// Maintain performs housekeeping for Mothballs.
func (m *Mothballs) Maintain(updateInterval time.Duration) {
	m.refresh()
	for range time.NewTicker(updateInterval).C {
		m.refresh()
	}
}
