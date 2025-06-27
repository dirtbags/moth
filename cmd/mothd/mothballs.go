package main

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dirtbags/moth/v5/pkg/forwardreader"
)

// Mothball contains an open zip.Reader and some cached information
type Mothball struct {
	*zip.Reader

	unlocked bool
	answers  map[PuzzleID][]string
	puzzles  []PuzzleID
}

// NewMothball returns a Mothball for rat
func NewMothball(rat io.ReaderAt, size int64) (*Mothball, error) {
	zr, err := zip.NewReader(rat, size)
	if err != nil {
		return nil, err
	}

	mb := &Mothball{
		Reader: zr,
	}

	for _, zf := range mb.File {
		switch zf.Name {
		case "__MOTH/unlocked":
			mb.unlocked = true
		case "__MOTH/answers.txt":
			rc, err := zf.Open()
			if err != nil {
				return nil, err
			}
			// There may be multiple zip file entries with the same name
			mb.answers = make(map[PuzzleID][]string)
			scanner := bufio.NewScanner(rc)
			for scanner.Scan() {
				line := scanner.Text()
				parts := strings.SplitN(line, " ", 2)
				puzzle := PuzzleID(parts[0])
				answer := parts[1]

				mb.answers[puzzle] = append(mb.answers[puzzle], answer)
			}
			rc.Close()
		case "__MOTH/puzzles.txt":
			rc, err := zf.Open()
			if err != nil {
				return nil, err
			}
			// There may be multiple zip file entries with the same name
			mb.puzzles = mb.puzzles[:0]
			scanner := bufio.NewScanner(rc)
			for scanner.Scan() {
				puzzle := PuzzleID(scanner.Text())
				mb.puzzles = append(mb.puzzles, puzzle)
			}
			rc.Close()
		}
	}

	return mb, nil
}

func (mb *Mothball) Unlocked() bool {
	return mb.unlocked
}

func (mb *Mothball) Puzzles() []PuzzleID {
	return mb.puzzles
}

func (mb *Mothball) Answers(puzzle PuzzleID) []string {
	return mb.answers[puzzle]
}

func (mb *Mothball) Open(name string) (fs.File, error) {
	name = path.Clean("/" + name)
	name = strings.TrimPrefix(name, "/")
	if strings.HasPrefix(name, "__MOTH/") {
		return nil, fs.ErrNotExist
	}

	f, err := mb.Reader.Open(name)
	if err != nil {
		return nil, err
	}

	// Go's HTTP server demands (at runtime) more functionality
	// than is provided by mb.Reader.Open(name)
	//
	// see https://mahesh-hegde.github.io/posts/go-interface-smuggling/
	return forwardreader.NewForwardFile(f)
}

type MothballStoreCacheEntry struct {
	io.Closer
	fs.FileInfo
	error
}

// BUG(neale): Windows won't let you remove, rename, or overwrite an open file. Once a mothball has been opened, you're stuck with it until the program exits.
type MothballStore struct {
	fsys      fs.FS
	mothballs atomic.Pointer[map[CategoryID]*Mothball]
	cache     map[CategoryID]*MothballStoreCacheEntry
}

// NewMothballStore returns a new MothballStore wrapping fsys.
func NewMothballStore(fsys fs.FS) *MothballStore {
	return &MothballStore{
		fsys:  fsys,
		cache: make(map[CategoryID]*MothballStoreCacheEntry),
	}
}

// Category returns the mothball which provides cat.
func (s *MothballStore) Category(cat CategoryID) (Category, error) {
	mothballsPtr := s.mothballs.Load()
	if mothballsPtr == nil {
		return nil, ErrNoSuchCategory
	}
	mothballs := *mothballsPtr
	mb, ok := mothballs[cat]
	if !ok {
		return nil, ErrNoSuchCategory
	}
	return mb, nil
}

func (s *MothballStore) Refresh(now time.Time) error {
	var mothballs map[CategoryID]*Mothball
	mothballsPtr := s.mothballs.Load()
	if mothballsPtr != nil {
		mothballs = *mothballsPtr
	}
	newMothballs := make(map[CategoryID]*Mothball)

	dents, err := fs.ReadDir(s.fsys, ".")
	if err != nil {
		return err
	}

	errs := make([]error, len(dents))
	for i, dent := range dents {
		name := dent.Name()
		catname := strings.TrimSuffix(name, ".mb")
		catid := CategoryID(catname)
		if name == catname {
			continue
		}

		info, err := dent.Info()
		if err != nil {
			errs[i] = err
			continue
		}

		cacheEntry, cached := s.cache[catid]
		if cached {
			if cacheEntry.ModTime().Equal(info.ModTime()) &&
				(cacheEntry.Size() == info.Size()) {
				newMothballs[catid] = mothballs[catid]
				errs[i] = cacheEntry.error
				continue
			}

			// If the file failed to open, Closer will be nil
			if cacheEntry.Closer != nil {
				cacheEntry.Close()
			}
			cacheEntry.Closer = nil
		} else {
			cacheEntry = new(MothballStoreCacheEntry)
			s.cache[catid] = cacheEntry
		}
		cacheEntry.FileInfo = info

		r, err := s.fsys.Open(name)
		if err != nil {
			errs[i] = err
			cacheEntry.error = err
			continue
		}
		cacheEntry.Closer = r

		rat, ok := r.(io.ReaderAt)
		if !ok {
			r.Close()
			err = fmt.Errorf("%s: ReaderAt not implemented", name)
			errs[i] = err
			cacheEntry.error = err
			continue
		}

		mb, err := NewMothball(rat, info.Size())
		if err != nil {
			r.Close()
			errs[i] = err
			cacheEntry.error = err
			continue
		}

		newMothballs[catid] = mb
	}

	s.mothballs.Store(&newMothballs)

	// Close any mothballs no longer in our list
	// BUG(neale): Race condition: we close the mothball file while a member file is being read
	for catid, entry := range s.cache {
		if _, ok := newMothballs[catid]; !ok {
			entry.Close()
			delete(s.cache, catid)
		}
	}

	return errors.Join(errs...)
}
