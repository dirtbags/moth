package docs

import (
	"testing"
	"testing/fstest"
)

func TestVersion(t *testing.T) {
	// Make sure nothing dies
	Version()
}

// TestFS ensures important documentation files exist.
func TestFS(t *testing.T) {
	fstest.TestFS(FS,
		"CHANGELOG.md",
		"LICENSE.md",
	)
}
