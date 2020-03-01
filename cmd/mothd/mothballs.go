package main

import (
	"github.com/spf13/afero"
	"log"
	"strings"
	"bufio"
	"strconv"
	"time"
	"fmt"
)

type Mothballs struct {
	categories map[string]*Zipfs
	afero.Fs
}

func NewMothballs(fs afero.Fs) *Mothballs {
	return &Mothballs{
		Fs:         fs,
		categories: make(map[string]*Zipfs),
	}
}

func (m *Mothballs) Open(cat string, points int, filename string) (ReadSeekCloser, error) {
	mb, ok := m.categories[cat]
	if ! ok {
		return nil, fmt.Errorf("No such category: %s", cat)
	}

	path := fmt.Sprintf("content/%d/%s", points, filename)
	return mb.Open(path)
}

func (m *Mothballs) ModTime(cat string, points int, filename string) (mt time.Time, err error) {
	mb, ok := m.categories[cat]
	if ! ok {
		return mt, fmt.Errorf("No such category: %s", cat)
	}
	mt = mb.ModTime()
	return
}

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
