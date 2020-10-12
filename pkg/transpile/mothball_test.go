package transpile

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/afero/zipfs"
)

func TestMothballsMemFs(t *testing.T) {
	static := NewFsCategory(newTestFs(), "cat1")
	mb := new(bytes.Buffer)
	if err := Mothball(static, mb); err != nil {
		t.Error(err)
	}
}

func TestMothballsOsFs(t *testing.T) {
	_, testfn, _, _ := runtime.Caller(0)
	os.Chdir(path.Dir(testfn))

	fs := NewRecursiveBasePathFs(afero.NewOsFs(), "testdata")
	static := NewFsCategory(fs, "static")
	mb := new(bytes.Buffer)
	err := Mothball(static, mb)
	if err != nil {
		t.Error(err)
		return
	}

	mbReader := bytes.NewReader(mb.Bytes())
	mbr, err := zip.NewReader(mbReader, int64(mb.Len()))
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
		} else if string(buf) != "1\n2\n3\n" {
			t.Error("Bad puzzles.txt", string(buf))
		}
	}
}
