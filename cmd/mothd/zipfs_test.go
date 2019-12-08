package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestZipPerformance(t *testing.T) {
	// I get 4.8s for 10,000 reads
	if os.Getenv("BENCHMARK") == "" {
		return
	}
	
	rng := rand.New(rand.NewSource(rand.Int63()))
	
	tf, err := ioutil.TempFile("", "zipfs")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(tf.Name())

	w := zip.NewWriter(tf)
	for i := 0; i < 100; i += 1 {
		fsize := 1000
		switch {
		case i % 10 == 0:
			fsize = 400000
		case i % 20 == 6:
			fsize  = 5000000
		case i == 80:
			fsize = 1000000000
		}
		
		f, err := w.Create(fmt.Sprintf("%d.bin", i))
		if err != nil {
			t.Fatal(err)
			return
		}
		if _, err := io.CopyN(f, rng, int64(fsize)); err != nil {
			t.Error(err)
		}
	}
	w.Close()
	
	tfsize, err := tf.Seek(0, 2)
	if err != nil {
		t.Fatal(err)
	}
	
	startTime := time.Now()
	nReads := 10000
	for i := 0; i < 10000; i += 1 {
		r, err := zip.NewReader(tf, tfsize)
		if err != nil {
			t.Error(err)
			return
		}
		filenum := rng.Intn(len(r.File))
		f, err := r.File[filenum].Open()
		if err != nil {
			t.Error(err)
			continue
		}
		buf, err := ioutil.ReadAll(f)
		if err != nil {
			t.Error(err)
		}
		t.Log("Read file of size", len(buf))
		f.Close()
	}
	t.Log(nReads, "reads took", time.Since(startTime))
	t.Error("moo")
}

func TestZipfs(t *testing.T) {
	tf, err := ioutil.TempFile("", "zipfs")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(tf.Name())

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
	mb, err := OpenZipfs(tf.Name())
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
