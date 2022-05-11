package main

import (
	"archive/zip"
	"bufio"
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
	mtime time.Time
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
		log.Println("There's no such category")
		return false, fmt.Errorf("no such category: %s", cat)
	}

	log.Println("Opening answers.txt")
	af, err := zfs.Open("answers.txt")
	if err != nil {
		log.Println("I did not find an answer")
		return false, fmt.Errorf("no answers.txt file")
	}
	defer af.Close()

	log.Println("I'm going to start looking for an answer")
	needle := fmt.Sprintf("%d %s", points, answer)
	scanner := bufio.NewScanner(af)
	for scanner.Scan() {
		log.Println("testing equality between", scanner.Text(), needle)
		if scanner.Text() == needle {
			return true, nil
		}
	}

	log.Println("I did not find the answer", answer)

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

		reopen := false
		if existingMothball, ok := m.categories[categoryName]; !ok {
			reopen = true
		} else if si, err := m.Fs.Stat(filename); err != nil {
			log.Println(err)
		} else if si.ModTime().After(existingMothball.mtime) {
			existingMothball.Close()
			delete(m.categories, categoryName)
			reopen = true
		}

		if reopen {
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
				mtime:  fi.ModTime(),
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
