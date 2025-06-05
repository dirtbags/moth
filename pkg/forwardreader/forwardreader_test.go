package forwardreader

import (
	"errors"
	"io"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

func TestForwardReaderAll(t *testing.T) {
	r := strings.NewReader("0123456789abcdef")
	fr := NewForwardReadSeeker(r, r.Size())
	buf, err := io.ReadAll(fr)
	if err != nil {
		t.Fatal(err)
	}
	if string(buf) != "0123456789abcdef" {
		t.Errorf("Read the wrong thing: %#v", buf)
	}
}

func TestForwardReaderSeek(t *testing.T) {
	r := strings.NewReader("0123456789abcdef")
	fr := NewForwardReadSeeker(r, r.Size())

	// Seek to the end, to get file size
	if n, err := fr.Seek(0, io.SeekEnd); err != nil {
		t.Error(err)
	} else if n != 16 {
		t.Errorf("Seek to end returned %d", n)
	}

	// Rewind
	if n, err := fr.Seek(5, io.SeekStart); err != nil {
		t.Error(err)
	} else if n != 5 {
		t.Errorf("Seek start: %d", n)
	}

	// Read 2 bytes
	buf := make([]byte, 2)
	if n, err := fr.Read(buf); err != nil {
		t.Error(err)
	} else if n != 2 {
		t.Errorf("Read: %d", n)
	} else if string(buf) != "56" {
		t.Errorf("Read bytes: %#v", buf)
	}

	// Rewind 2 bytes
	if n, err := fr.Seek(-2, io.SeekCurrent); err != nil {
		t.Error(err)
	} else if n != 5 {
		t.Errorf("Seek: %d", n)
	}

	// Try to read again, should fail
	if _, err := fr.Read(buf); !errors.Is(err, fs.ErrInvalid) {
		t.Error("Did not return invalid error")
	}

	// Seek to the last byte
	if n, err := fr.Seek(-1, io.SeekEnd); err != nil {
		t.Error(err)
	} else if n != 15 {
		t.Errorf("Seek to last byte returned %d", n)
	}

	// Read it
	if n, err := fr.Read(buf); err != nil {
		t.Error(err)
	} else if n != 1 {
		t.Errorf("Read: %d", n)
	} else if buf[0] != 'f' {
		t.Errorf("Read bytes: %#v", buf)
	}

	// Do another read, past end
	if _, err := fr.Read(buf); !errors.Is(err, io.EOF) {
		t.Error("Did not return EOF", err)
	}
}

func TestForwardReaderAtEOF(t *testing.T) {
	r := strings.NewReader("0123456789abcdef")
	fra := NewForwardReaderAt(r, r.Size())
	buf := make([]byte, 2)

	// ReadAt to get better coverage of code
	if _, err := fra.ReadAt(buf, 58); !errors.Is(err, io.EOF) {
		t.Error("Did not return EOF", err)
	}
}

func TestForwardFile(t *testing.T) {
	fs := fstest.MapFS{
		"foo": &fstest.MapFile{Data: []byte("0123456789abcdef"), Mode: 0x644},
	}
	f, err := fs.Open("foo")
	if err != nil {
		t.Fatal(err)
	}
	ff, err := NewForwardFile(f)
	if err != nil {
		t.Fatal(err)
	}
	defer ff.Close()

	buf := make([]byte, 2)
	if n, err := ff.Read(buf); err != nil {
		t.Error(err)
	} else if n != 2 {
		t.Error("Wrong n", n)
	}
}
