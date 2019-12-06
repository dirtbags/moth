package main

import (
	"github.com/spf13/afero"
	"net/http"
	"strings"
)

type Theme struct {
	fs afero.Fs
}

func NewTheme(fs afero.Fs) *Theme {
	return &Theme{
		fs: fs,
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

	f, err := t.fs.Open(path)
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
