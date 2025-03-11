package main

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// FSCategory is the interface needed by FSCategoryStore.
type FSCategory interface {
	Category
	Close() error
	OlderThan(time.Time) bool
}

// FSBaseCategory provides cached values, and mtime comparison.
//
// This is used by composed structs to provide the FSCategory interface.
type FSBaseCategory struct {
	mtime time.Time

	unlocked bool
	answers  map[PuzzleID][]string
	puzzles  []PuzzleID
}

// initialize caches metadata (unlocked, answers, puzzles).
func NewFSBaseCategory(fsys fs.FS, mtime time.Time) (*FSBaseCategory, error) {
	c := &FSBaseCategory{
		mtime: mtime,
	}

	if _, err := fs.Stat(fsys, "__MOTH/unlocked"); err == nil {
		c.unlocked = true
	}

	if r, err := fsys.Open("__MOTH/answers.txt"); err != nil {
		return nil, err
	} else {
		c.answers = make(map[PuzzleID][]string)
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.SplitN(line, " ", 2)
			puzzle := PuzzleID(parts[0])
			answer := parts[1]

			c.answers[puzzle] = append(c.answers[puzzle], answer)
		}
		r.Close()
	}

	if r, err := fsys.Open("__MOTH/puzzles.txt"); err != nil {
		return nil, err
	} else {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			puzzle := PuzzleID(scanner.Text())
			c.puzzles = append(c.puzzles, puzzle)
		}
		r.Close()
	}

	return c, nil
}

func (c *FSBaseCategory) OlderThan(t time.Time) bool {
	return c.mtime.Before(t)
}

func (c *FSBaseCategory) Unlocked() bool {
	return c.unlocked
}

func (c *FSBaseCategory) Puzzles() []PuzzleID {
	return c.puzzles
}

func (c *FSBaseCategory) Answers(puzzle PuzzleID) []string {
	return c.answers[puzzle]
}

// FSDirCategory implements a Category backed by an fs.FS.
type FSDirCategory struct {
	*FSBaseCategory
	fsys fs.FS
}

func NewFSDirCategory(fsys fs.FS, mtime time.Time) (*FSDirCategory, error) {
	bc, err := NewFSBaseCategory(fsys, mtime)
	if err != nil {
		return nil, err
	}

	c := &FSDirCategory{
		FSBaseCategory: bc,
		fsys:           fsys,
	}
	return c, nil
}

// Close does nothing.
func (c *FSDirCategory) Close() error {
	return nil
}

// ServeFile
func (c *FSDirCategory) ServeFile(w http.ResponseWriter, req *http.Request, name string) {
	name = strings.TrimPrefix(name, "/")
	if strings.HasPrefix(name, "__MOTH/") {
		http.NotFound(w, req)
		return
	}

	f, err := c.fsys.Open(name)
	if errors.Is(err, fs.ErrInvalid) {
		http.NotFound(w, req)
		return
	} else if err != nil {
		HTTPError(w, err)
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		HTTPError(w, err)
		return
	}

	// MOTH only serves files
	if info.IsDir() {
		http.NotFound(w, req)
		return
	}

	fseeker, ok := f.(io.ReadSeeker)
	if !ok {
		http.Error(w, "file does not implement io.ReadSeeker", http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, req, name, info.ModTime(), fseeker)
}

type FSZipCategory struct {
	*FSBaseCategory
	reader *zip.Reader
	closer io.Closer
}

func NewFSZipCategory(rat io.ReaderAt, closer io.Closer, mtime time.Time, size int64) (*FSZipCategory, error) {
	reader, err := zip.NewReader(rat, size)
	if err != nil {
		return nil, err
	}

	bc, err := NewFSBaseCategory(reader, mtime)
	if err != nil {
		return nil, err
	}

	c := &FSZipCategory{
		FSBaseCategory: bc,
		reader:         reader,
		closer:         closer,
	}

	return c, nil
}

func (c *FSZipCategory) Close() error {
	return c.closer.Close()
}

func (c *FSZipCategory) ServeFile(w http.ResponseWriter, req *http.Request, name string) {
	name = strings.TrimPrefix(name, "/")
	if strings.HasPrefix(name, "__MOTH/") {
		http.NotFound(w, req)
		return
	}

	w.Header().Add("Vary", "Accept-Encoding")

	var zf *zip.File
	for _, f := range c.reader.File {
		if f.Name == name {
			zf = f
			// Keep going, zip files can have multiple entries
		}
	}
	if zf == nil {
		http.NotFound(w, req)
		return
	}

	// Maybe we don't have to decompress anything?
	if (zf.Method == zip.Deflate) && strings.Contains(req.Header.Get("Accept-Encoding"), "deflate") {
		if c.ServeFileDeflate(w, req, name, zf) {
			return
		}
	}

	fseeker := NewZipFileSeeker(zf)
	http.ServeContent(w, req, name, zf.Modified, fseeker)

}

func (c *FSZipCategory) ServeFileDeflate(w http.ResponseWriter, req *http.Request, name string, zf *zip.File) bool {
	rf, err := zf.OpenRaw()
	if err != nil {
		return false
	}

	fseeker, ok := rf.(io.ReadSeeker)
	if !ok {
		slog.Error("zip.File.OpenRaw() no longer implements io.ReadSeeker")
		return false
	}

	w.Header().Add("Content-Encoding", "deflate")
	http.ServeContent(w, req, name, zf.Modified, fseeker)
	return true
}

// HTTPError sends an error code based on err
func HTTPError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, fs.ErrNotExist):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, fs.ErrPermission):
		http.Error(w, err.Error(), http.StatusForbidden)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// BUG(neale): Windows won't let you remove, rename, or overwrite an open file.
type FSCategoryStore struct {
	fsys       fs.FS
	categories atomic.Pointer[map[CategoryID]FSCategory]
}

func NewFSCategoryStore(fsys fs.FS) *FSCategoryStore {
	c := &FSCategoryStore{
		fsys: fsys,
	}
	return c
}

func (s *FSCategoryStore) Category(cat CategoryID) (Category, error) {
	categoriesPtr := s.categories.Load()
	if categoriesPtr == nil {
		return nil, ErrNoSuchCategory
	}
	categories := *categoriesPtr
	c, ok := categories[cat]
	if !ok {
		return nil, ErrNoSuchCategory
	}
	return c, nil
}

// newCategory is a helper function for Refresh.
//
// It returns the parsed category id, if it thinks this is a category.
// cat will be the existing category if the filesystem doesn't indicate a change.
func (s *FSCategoryStore) newCategory(dent fs.DirEntry, oldcategories map[CategoryID]FSCategory) (catid CategoryID, cat FSCategory, err error) {
	name := dent.Name()
	catid = CategoryID(name)

	switch name[0] {
	case '.', '_':
		return "", nil, nil
	}

	info, err := dent.Info()
	if err != nil {
		return "", nil, err
	}

	// Keep old category if modtime hasn't changed
	mtime := info.ModTime()
	if oldcat, ok := oldcategories[catid]; ok {
		cat = oldcat
		if !oldcat.OlderThan(mtime) {
			return catid, oldcat, nil
		}
	}

	switch {
	case dent.IsDir():
		fsys, ferr := fs.Sub(s.fsys, name)
		if ferr != nil {
			return catid, nil, ferr
		}

		cat, err = NewFSDirCategory(fsys, mtime)

	case strings.HasSuffix(name, ".mb"):
		r, ferr := s.fsys.Open(name)
		if ferr != nil {
			return catid, nil, ferr
		}

		rat, ok := r.(io.ReaderAt)
		if !ok {
			r.Close()
			return catid, nil, errors.New("mothball file handle does not provide ReaderAt")
		}

		cat, err = NewFSZipCategory(rat, r, mtime, info.Size())

	default:
		// Skip other files, like README
		return "", nil, nil

	}

	return catid, cat, err
}

func (s *FSCategoryStore) Refresh(now time.Time) error {
	var oldcategories map[CategoryID]FSCategory
	if cp := s.categories.Load(); cp != nil {
		oldcategories = *cp
	}
	newcategories := make(map[CategoryID]FSCategory)

	dents, err := fs.ReadDir(s.fsys, ".")
	if err != nil {
		return err
	}

	var errs []error
	for _, dent := range dents {
		catid, cat, err := s.newCategory(dent, oldcategories)
		if err != nil {
			errs = append(errs, fmt.Errorf("category %s: %v", dent.Name(), err))
			continue
		}
		if cat != nil {
			newcategories[catid] = cat
		}
	}

	s.categories.Store(&newcategories)

	// Close any categories no longer in our list
	for catid, cat := range oldcategories {
		if _, ok := newcategories[catid]; !ok {
			cat.Close()
		}
	}

	return errors.Join(errs...)
}
