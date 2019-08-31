package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"context"

	"golang.org/x/net/webdav"
)

type StubLockSystem struct {
}

func (ls *StubLockSystem) Confirm(now time.Time, name0, name1 string, conditions ...webdav.Condition) (release func(), err error) {
	return nil, webdav.ErrConfirmationFailed
}

func (ls *StubLockSystem) Create(now time.Time, details webdav.LockDetails) (token string, err error) {
	return "", webdav.ErrLocked
}

func (ls *StubLockSystem) Refresh(now time.Time, token string, duration time.Duration) (webdav.LockDetails, error) {
	return webdav.LockDetails{}, webdav.ErrNoSuchLock
}

func (ls *StubLockSystem) Unlock(now time.Time, token string) error {
	return webdav.ErrNoSuchLock
}


type MothFS struct {
}

func (fs *MothFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return os.ErrPermission
}

func (fs *MothFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	f, err := os.Open("hello.txt")
	return f, err
}

func (fs *MothFS) RemoveAll(ctx context.Context, name string) error {
	return os.ErrPermission
}

func (fs *MothFS) Rename(ctx context.Context, oldName, newName string) error {
	return os.ErrPermission
}

func (fs *MothFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	info, err := os.Stat("hello.txt")
	return info, err
}

func main() {
	//dirFlag := flag.String("d", "./", "Directory to serve from. Default is CWD")
	httpPort := flag.Int("p", 80, "Port to serve on (Plain HTTP)")

	flag.Parse()

	srv := &webdav.Handler{
		FileSystem: new(MothFS),
		LockSystem: new(StubLockSystem),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				log.Printf("WEBDAV [%s]: %s, ERROR: %s\n", r.Method, r.URL, err)
			} else {
				log.Printf("WEBDAV [%s]: %s \n", r.Method, r.URL)
			}
		},
	}
	http.Handle("/", srv)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *httpPort), nil); err != nil {
		log.Fatalf("Error with WebDAV server: %v", err)
	}

}
