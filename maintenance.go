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
	
	// Make sure points directories exist
	os.Mkdir(statePath("points.tmp"), 0755)
	os.Mkdir(statePath("points.new"), 0755)

	// Get current list of categories
	newCategories := []string{}
	files, err := ioutil.ReadDir(mothPath("packages"))
	if err != nil {
		log.Printf("Error reading packages: %s", err)
	}
	for _, f := range files {
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
