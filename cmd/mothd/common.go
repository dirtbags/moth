package main

import (
	"path/filepath"
	"strings"
	"time"
)

type Component struct {
	baseDir string
}

func (c *Component) path(parts ...string) string {
	path := filepath.Clean(filepath.Join(parts...))
	parts = filepath.SplitList(path)
	for i, part := range parts {
		part = strings.TrimLeft(part, "./\\:")
		parts[i] = part
	}
	parts = append([]string{c.baseDir}, parts...)
	path = filepath.Join(parts...)
	path = filepath.Clean(path)
	return path
}

func (c *Component) Run(updateInterval time.Duration) {
	// Stub!
}