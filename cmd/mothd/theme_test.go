package main

import (
	"github.com/spf13/afero"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTheme(t *testing.T) {
	fs := new(afero.MemMapFs)
	afero.WriteFile(fs, "/index.html", []byte("index"), 0644)
	afero.WriteFile(fs, "/moo.html", []byte("moo"), 0644)

	s := NewTheme(fs)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.staticHandler)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Handler returned wrong code: %v", rr.Code)
	}

	if rr.Body.String() != "index" {
		t.Errorf("Handler returned wrong content: %v", rr.Body.String())
	}
}
