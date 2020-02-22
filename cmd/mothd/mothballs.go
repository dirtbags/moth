package main

import (
	"github.com/spf13/afero"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

type Mothballs struct {
	fs afero.Fs
	categories map[string]*Zipfs
}

func NewMothballs(fs afero.Fs) *Mothballs {
	return &Mothballs{
		fs: fs,
		categories: make(map[string]*Zipfs),
	}
}

func (m *Mothballs) update() {
	// Any new categories?
	files, err := afero.ReadDir(m.fs, "/")
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
