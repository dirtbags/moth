package main

import (
	"log"
	"io/ioutil"
	"time"
	"strings"
)

func allfiles(dirpath) []string {
	files, err := ioutil.ReadDir(dirpath)
	if (err != nil) {
		log.Printf("Error reading directory %s: %s", dirpath, err)
		return []
	}
	return files
}


// maintenance runs
func tidy() {
	// Skip if we've been disabled
	if exists(statePath("disabled")) {
		log.Print("disabled file found, suspending maintenance")
		return
	}
	
	// Skip if we've expired
	untilspec, _ := ioutil.ReadFile(statePath("until"))
	until, err := time.Parse(time.RFC3339, string(untilspec))
	if err == nil {
		if until.Before(time.Now()) {
			log.Print("until file time reached, suspending maintenance")
			return
		}
	}
	
	log.Print("Hello, I'm maintaining!")
	
	//
	// Get current list of categories
	//
	newCategories := []string{}
	for f := range(allfiles(mothPath("packages"))) {
		filename := f.Name()
		filepath := mothPath("packages", filename)
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
func maintenance() {
	for ;; time.Sleep(maintenanceInterval) {
		tidy()
	}
}
