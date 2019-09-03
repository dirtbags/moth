package main

import (
	"net/http"
	"os"
	"strings"
)

type Theme struct {
	Component
}

func NewTheme(baseDir string) *Theme {
	return &Theme{
		Component: Component{
			baseDir: baseDir,
		},
	}
}

func (t *Theme) staticHandler(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	if strings.Contains(path, "/.") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	if path == "/" {
		path = "/index.html"
	}

	f, err := os.Open(t.path(path))
	if err != nil {
		http.NotFound(w, req)
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		http.NotFound(w, req)
		return
	}

	http.ServeContent(w, req, path, d.ModTime(), f)
}
