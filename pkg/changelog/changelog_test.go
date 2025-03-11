package changelog

import (
	"testing"
	"testing/fstest"
)

var noChangelogFS = fstest.MapFS{}

var badChangelogFS = fstest.MapFS{
	"CHANGELOG.md": &fstest.MapFile{
		Data: []byte("bloop"),
	},
}

var badSyntaxChangelogFS = fstest.MapFS{
	"CHANGELOG.md": &fstest.MapFile{
		Data: []byte(`A changelog!
## 1.0.0 - unreleased
`),
	},
}

var unreleasedChangelogFS = fstest.MapFS{
	"CHANGELOG.md": &fstest.MapFile{
		Data: []byte(`A changelog!
## [unreleased]
- bloop
`),
	},
}

var releasedChangelogFS = fstest.MapFS{
	"CHANGELOG.md": &fstest.MapFile{
		Data: []byte(`A changelog!
## [1.0.0] - 2017-06-20
- bloop

## [0.9.0] - 2017-06-19
- bloop
`),
	},
}

func TestNoChangelog(t *testing.T) {
	v, d := ReadVersionFS(noChangelogFS)
	if (v != "unknown-nochangelog") || (d != "???") {
		t.Fatal("Wrong version:", v, d)
	}
}

func TestBadChangelog(t *testing.T) {
	v, d := ReadVersionFS(badChangelogFS)
	if (v != "unknown-changelog-noversion") || (d != "???") {
		t.Fatal("Wrong version:", v, d)
	}
}

func TestBadSyntaxChangelog(t *testing.T) {
	v, d := ReadVersionFS(badSyntaxChangelogFS)
	if (v != "unknown-changelog-noversion") || (d != "???") {
		t.Fatal("Wrong version:", v, d)
	}
}

func TestUnreleasedChangelog(t *testing.T) {
	v, d := ReadVersionFS(unreleasedChangelogFS)
	if (v != "unreleased") || (d != "???") {
		t.Fatal("Wrong version:", v, d)
	}
}

func TestReleasedChangelog(t *testing.T) {
	v, d := ReadVersionFS(releasedChangelogFS)
	if (v != "1.0.0") || (d != "2017-06-20") {
		t.Fatal("Wrong version:", v, d)
	}
}

