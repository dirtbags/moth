package docs

import (
	"embed"

	"github.com/dirtbags/moth/v5/pkg/changelog"
)

// FS contains built-in documentation.
//
//go:embed *
var FS embed.FS

// Version returns the topmost version and date in CHANGELOG.md.
func Version() (version string, date string) {
	return changelog.ReadVersionFS(FS)
}
