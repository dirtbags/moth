package main

import (
	"archive/zip"
	"fmt"
	"io"
	"github.com/spf13/afero"
	"testing"
)


func TestZipfs(t *testing.T) {
	fs := new(afero.MemMapFs)
	
	tf, err := fs.Create("/test.zip")
	if err != nil {
		t.Error(err)
		return
	}
	defer fs.Remove(tf.Name())

	w := zip.NewWriter(tf)
	f, err := w.Create("moo.txt")
	if err != nil {
		t.Error(err)
		return
	}
	// no Close method

	_, err = fmt.Fprintln(f, "The cow goes moo")
	//.Write([]byte("The cow goes moo"))
	if err != nil {
		t.Error(err)
		return
	}
	w.Close()
	tf.Close()

	// Now read it in
	mb, err := OpenZipfs(fs, tf.Name())
	if err != nil {
		t.Error(err)
		return
	}

	cow, err := mb.Open("moo.txt")
	if err != nil {
		t.Error(err)
		return
	}

	line := make([]byte, 200)
	n, err := cow.Read(line)
	if (err != nil) && (err != io.EOF) {
		t.Error(err)
		return
	}

	if string(line[:n]) != "The cow goes moo\n" {
		t.Log(line)
		t.Error("Contents didn't match")
		return
	}

}
