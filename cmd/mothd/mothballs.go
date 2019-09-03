package main

import (
	"time"
	"io/ioutil"
	"strings"
	"log"
)

type Mothballs struct {
	Component
	categories map[string]*Zipfs
}

func NewMothballs(baseDir string) *Mothballs {
	return &Mothballs{
		Component: Component{
			baseDir: baseDir,
		},
		categories: make(map[string]*Zipfs),
	}
}

func (m *Mothballs) update() {
	// Any new categories?
	files, err := ioutil.ReadDir(m.path())
	if err != nil {
		log.Print("Error listing mothballs: ", err)
		return
	}
	for _, f := range files {
		filename := f.Name()
		filepath := m.path(filename)
		if !strings.HasSuffix(filename, ".mb") {
			continue
		}
		categoryName := strings.TrimSuffix(filename, ".mb")

		if _, ok := m.categories[categoryName]; !ok {
			zfs, err := OpenZipfs(filepath)
			if err != nil {
				log.Print("Error opening ", filepath, ": ", err)
				continue
			}
			log.Print("New mothball: ", filename)
			m.categories[categoryName] = zfs
		}
	}
}

func (m *Mothballs) Run(updateInterval time.Duration) {
	ticker := time.NewTicker(updateInterval)
	m.update()
	for {
		select {
		case when := <-ticker.C:
			log.Print("Tick: ", when)
			m.update()
		}
	}
}
