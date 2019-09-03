package main

import (
	"path/filepath"
	"strings"
)

func MothPath(base string, parts ...string) string {
	path := filepath.Clean(filepath.Join(parts...))
	parts = filepath.SplitList(path)
	for i, part := range parts {
		part = strings.TrimLeft(part, "./\\:")
		parts[i] = part
	}
	parts = append([]string{base}, parts...)
	path = filepath.Join(parts...)
	path = filepath.Clean(path)
	return path
}
