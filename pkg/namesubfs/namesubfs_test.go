package namesubfs

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestSubFS(t *testing.T) {
	testfs := fstest.MapFS{
		"static/moo.txt":         &fstest.MapFile{Data: []byte("moo.\n")},
		"static/subdir/moo2.txt": &fstest.MapFile{Data: []byte("moo too.\n")},
	}
	if static, err := NameSub(testfs, "static"); err != nil {
		t.Error(err)
	} else if buf, err := fs.ReadFile(static, "moo.txt"); err != nil {
		t.Error(err)
	} else if string(buf) != "moo.\n" {
		t.Error("Wrong file contents")
	} else if subdir, err := NameSub(static, "subdir"); err != nil {
		t.Error(err)
	} else if buf, err := fs.ReadFile(subdir, "moo2.txt"); err != nil {
		t.Error(err)
	} else if string(buf) != "moo too.\n" {
		t.Error("Wrong file contents too")
	} else if subdir.FullName("glue") != "static/subdir/glue" {
		t.Error("Wrong full name", subdir.FullName("glue"))
	}

	if a, err := NameSub(testfs, "a"); err != nil {
		t.Error(err)
	} else if b, err := fs.Sub(a, "b"); err != nil {
		t.Error(err)
	} else if c, err := NameSub(b, "c"); err != nil {
		t.Error(err)
	} else if c.FullName("d") != "a/b/c/d" {
		t.Error(c.FullName("d"))
	}
}
