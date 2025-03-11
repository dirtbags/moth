package mdir

import (
	"fmt"
	"io/fs"
	"runtime"
	"testing"

	"github.com/dirtbags/moth/v5/pkg/wfs"
)

func TestCreate(t *testing.T) {
	tempDir := t.TempDir()
	fsys := wfs.DirFS(tempDir)

	md, err := NewDefault(fsys)
	if err != nil {
		t.Error(err)
	}

	name := "bloop.txt"
	body := "bloop\n"
	if f, err := md.Create(name); err != nil {
		t.Error(err)
	} else if _, err := fmt.Fprint(f, body); err != nil {
		f.Close()
		t.Fatal(err)
	} else if _, err := fs.ReadFile(fsys, wfs.Join(md.TmpDir, name)); err != nil {
		f.Close()
		t.Fatal(err)
	} else if err := f.Close(); err != nil {
		t.Error(err)
	}

	if _, err := fs.ReadFile(fsys, wfs.Join(md.NewDir, name)); err != nil {
		t.Fatal(err)
	}

	t.Log("filename (fsys):", fsys.OSPath(wfs.Join(md.NewDir, name)))
	t.Log("filename (md):", md.OSPath(name))
	if data, err := fs.ReadFile(md, name); err != nil {
		t.Error(err)
	} else if string(data) != body {
		t.Error("Body isn't the same")
	}
}

func TestReset(t *testing.T) {
	tempDir := t.TempDir()
	fsys := wfs.DirFS(tempDir)

	md, err := NewDefault(fsys)
	if err != nil {
		t.Fatal(err)
	}

	if err := md.WriteFile("test.txt", []byte("bloop"), 0644); err != nil {
		t.Fatal(err)
	}
	if ents, err := fs.ReadDir(md, "."); err != nil {
		t.Error(err)
	} else if len(ents) != 1 {
		t.Error("Wrong number of files")
	}

	if err := md.Reset(); err != nil {
		t.Error(err)
	}

	if ents, err := fs.ReadDir(md, "."); err != nil {
		t.Error(err)
	} else if len(ents) != 0 {
		t.Error("Wrong number of files")
	}
}

func TestOverwrite(t *testing.T) {
	tempDir := t.TempDir()
	fsys := wfs.DirFS(tempDir)
	md, err := NewDefault(fsys)
	if err != nil {
		t.Fatal(err)
	}

	if err := md.WriteFile("a", []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	if err := md.WriteFile("a", []byte("a"), 0644); err != nil {
		t.Error(err)
	}
	if buf, err := fs.ReadFile(md, "a"); err != nil {
		t.Error(err)
	} else if string(buf) != "a" {
		t.Error("Wrong contents on overwrite")
	}

	// Trigger MDir.rename() bug
	if runtime.GOOS != "windows" {
		f, err := md.Open("a")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := md.WriteFile("a", []byte{}, 0644); err != nil {
			t.Error(err)
		}
	}
}
