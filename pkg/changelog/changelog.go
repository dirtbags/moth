// Package changelog parses the most recent version from a CHANGELOG.md file.
package changelog

import (
	"bufio"
	"io"
	"io/fs"
	"strings"
)

// ReadVersionFS calls ReadVersion on the CHANGELOG.md file in fsys.
func ReadVersionFS(fs fs.FS) (version string, date string) {
	r, err := fs.Open("CHANGELOG.md")
	if err != nil {
		return "unknown-nochangelog", "???"
	}
	defer r.Close()

	return ReadVersion(r)
}

// ReadVersion returns the topmost version listed in r.
//
// r must be a changelog in "keep a changelog" format:
// https://keepachangelog.com/
// Specifically, versions must appear as one of the two formats:
//
//   - ## [Unreleased]
//   - ## [1.0.0] - 2017-06-20
func ReadVersion(r io.Reader) (version string, date string) {
	date = "???"

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		release := strings.TrimPrefix(line, "## [")
		if release == line {
			continue
		}
		v, d, found := strings.Cut(release, " - ")
		version = strings.TrimSuffix(v, "]")
		if found {
			date = d
		}
		return version, date
	}
	return "unknown-changelog-noversion", date
}
