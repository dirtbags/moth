package main

import (
	"bufio"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/afero"
)

// Mothballs provides a collection of active mothball files (puzzle categories)
type Mothballs struct {
	categories map[string]*Zipfs
	afero.Fs
}

// NewMothballs returns a new Mothballs structure backed by the provided directory
func NewMothballs(fs afero.Fs) *Mothballs {
	return &Mothballs{
		Fs:         fs,
		categories: make(map[string]*Zipfs),
	}
}

// Open returns a ReadSeekCloser corresponding to the filename in a puzzle's category and points
func (m *Mothballs) Open(cat string, points int, filename string) (ReadSeekCloser, time.Time, error) {
	mb, ok := m.categories[cat]
	if !ok {
		return nil, time.Time{}, fmt.Errorf("No such category: %s", cat)
	}

	f, err := mb.Open(fmt.Sprintf("content/%d/%s", points, filename))
	return f, mb.ModTime(), err
}

// Inventory returns the list of current categories
func (m *Mothballs) Inventory() []Category {
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
		categories = append(categories, Category{cat, pointsList})
	}
	return categories
}

// CheckAnswer returns an error if the provided answer is in any way incorrect for the given category and points
func (m *Mothballs) CheckAnswer(cat string, points int, answer string) error {
	zfs, ok := m.categories[cat]
	if !ok {
		return fmt.Errorf("No such category: %s", cat)
	}

	af, err := zfs.Open("answers.txt")
	if err != nil {
		return fmt.Errorf("No answers.txt file")
	}
	defer af.Close()

	needle := fmt.Sprintf("%d %s", points, answer)
	scanner := bufio.NewScanner(af)
	for scanner.Scan() {
		if scanner.Text() == needle {
			return nil
		}
	}

	return fmt.Errorf("Invalid answer")
}

// Update refreshes internal state.
// It looks for changes to the directory listing, and caches any new mothballs.
func (m *Mothballs) Update() {
	// Any new categories?
	files, err := afero.ReadDir(m.Fs, "/")
	if err != nil {
		log.Print("Error listing mothballs: ", err)
		return
	}
	for _, f := range files {
		filename := f.Name()
		if !strings.HasSuffix(filename, ".mb") {
			continue
		}
		categoryName := strings.TrimSuffix(filename, ".mb")

		if _, ok := m.categories[categoryName]; !ok {
			zfs, err := OpenZipfs(m.Fs, filename)
			if err != nil {
				log.Print("Error opening ", filename, ": ", err)
				continue
			}
			log.Print("New mothball: ", filename)
			m.categories[categoryName] = zfs
		}
	}
}
