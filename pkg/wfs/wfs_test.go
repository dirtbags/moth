package wfs

import (
	"io/fs"
	"os"
	"testing"
)

func TestUnderlyingFsys(t *testing.T) {
	tempDir := t.TempDir()
	fsys := DirFS(tempDir)

	t.Log("tempDir is", tempDir)
	if files, err := os.ReadDir(tempDir); err != nil {
		t.Error(err)
	} else {
		t.Log("Directory contents:", files)
	}

	if _, err := fs.ReadDir(fsys, "."); err != nil {
		t.Error(err)
	}
}

func TestBasic(t *testing.T) {
	tempDir := t.TempDir()
	fsys := DirFS(tempDir)

	if err := fsys.Mkdir("dir", 0755); err != nil {
		t.Error(err)
	}

	if err := fsys.WriteFile("dir/booger.txt", []byte("booger"), 0644); err != nil {
		t.Error(err)
	}

	if contents, err := fs.ReadFile(fsys, "dir/booger.txt"); err != nil {
		t.Error(err)
	} else if string(contents) != "booger" {
		t.Error("wrong contents")
	}

	if _, err := fsys.Stat("dir/booger.txt"); err != nil {
		t.Error(err)
	}

	if _, err := fs.ReadDir(fsys, "."); err != nil {
		t.Error(err)
	}

	if err := fsys.Rename("dir/booger.txt", "dir/goober.txt"); err != nil {
		t.Error(err)
	}

	if f, err := fsys.OpenFile("dir/goober.txt", os.O_RDONLY, 0644); err != nil {
		t.Error(err)
	} else if err := f.Close(); err != nil {
		t.Error(err)
	}

	if err := fsys.Remove("dir/goober.txt"); err != nil {
		t.Error(err)
	}

	if err := fsys.RemoveAll("dir"); err != nil {
		t.Error(err)
	}

	if _, err := fsys.Stat("dir"); err == nil {
		t.Error("dir was not removed")
	}
}

func TestSub(t *testing.T) {
	tempDir := t.TempDir()
	fsys := DirFS(tempDir)

	if err := fsys.Mkdir("teams", 0755); err != nil {
		t.Error(err)
	}
	if err := fsys.WriteFile("teams/t3dro", []byte("team name"), 0644); err != nil {
		t.Error(err)
	}
	if ents, err := fs.ReadDir(fsys, "teams"); err != nil {
		t.Error(err)
	} else if len(ents) != 1 {
		t.Error("Wrong number of files in teams directory:", ents)
	}

	if subfs, err := Sub(fsys, "teams"); err != nil {
		t.Fatal(err)
	} else if _, err := fs.ReadFile(subfs, "t3dro"); err != nil {
		t.Fatal(err)
	}
}

func TestJoin(t *testing.T) {
	if a := Join("a", "b", "c"); a != "a/b/c" {
		t.Error("Join didn't work right")
	}

	if a := Join(); a != "" {
		t.Error("Join nothing made something")
	}
}
