package main

import (
	"log"
	"io/ioutil"
	"time"
	"os"
	"strings"
)

func cacheMothball(filepath string, categoryName string) {
	log.Printf("I'm exploding a mothball %s %s", filepath, categoryName)
}

// maintenance runs
func tidy() {
	// Skip if we've been disabled
	if _, err := os.Stat(statePath("disabled")); err == nil {
		log.Print("disabled file found, suspending maintenance")
		return
	}
	
	// Skip if we've expired
	untilspec, err := ioutil.ReadFile(statePath("until"))
	if err == nil {
		until, err := time.Parse(time.RFC3339, string(untilspec))
		if err != nil {
			log.Print("Unparseable date in until file: %s", until)
		} else {
			if until.Before(time.Now()) {
				log.Print("until file time reached, suspending maintenance")
				return
			}
		}
	}
	
	// Get current list of categories
	newCategories := []string{}
	files, err := ioutil.ReadDir(modulesPath())
	if err != nil {
		log.Printf("Error reading packages: %s", err)
	}
	for _, f := range files {
		filename := f.Name()
		filepath := modulesPath(filename)
		if ! strings.HasSuffix(filename, ".mb") {
			continue
		}
		
		categoryName := strings.TrimSuffix(filename, ".mb")
		newCategories = append(newCategories, categoryName)
		
		// Uncompress into cache directory
		cacheMothball(filepath, categoryName)
	}
	categories = newCategories
	
	collectPoints()
}

// maintenance is the goroutine that runs a periodic maintenance task
func maintenance(maintenanceInterval time.Duration) {
	for ;; time.Sleep(maintenanceInterval) {
		tidy()
	}
}
