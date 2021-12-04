package transpile

import (
	"io/fs"
	"os"
	"testing"
)

func TestSubFS(t *testing.T) {
	testdata := os.DirFS("testdata")
	if static, err := Sub(testdata, "static"); err != nil {
		t.Error(err)
	} else if buf, err := fs.ReadFile(static, "moo.txt"); err != nil {
		t.Error(err)
	} else if string(buf) != "moo.\n" {
		t.Error("Wrong file contents")
	} else if subdir, err := static.Sub("subdir"); err != nil {
		t.Error(err)
	} else if buf, err := fs.ReadFile(subdir, "moo2.txt"); err != nil {
		t.Error(err)
	} else if string(buf) != "moo too.\n" {
		t.Error("Wrong file contents too")
	} else if subdir.FullName("glue") != "static/subdir/glue" {
		t.Error("Wrong full name")
	}
}
