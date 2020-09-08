package transpile

import (
	"archive/zip"
	"io/ioutil"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/afero/zipfs"
)

func TestMothballs(t *testing.T) {
	fs := NewRecursiveBasePathFs(afero.NewOsFs(), "testdata")
	static := NewFsCategory(fs, "static")
	mb, err := Mothball(static)
	if err != nil {
		t.Error(err)
	}

	mbr, err := zip.NewReader(mb, int64(mb.Len()))
	if err != nil {
		t.Error(err)
	}
	zfs := zipfs.New(mbr)

	if f, err := zfs.Open("puzzles.txt"); err != nil {
		t.Error(err)
	} else {
		defer f.Close()
		if buf, err := ioutil.ReadAll(f); err != nil {
			t.Error(err)
		} else if string(buf) != "" {
			t.Error("Bad puzzles.txt", string(buf))
		}
	}
}
