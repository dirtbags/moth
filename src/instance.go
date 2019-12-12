package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

type RuntimeConfig struct {
	export_manifest bool
}

type Instance struct {
	Base            string
	MothballDir     string
	StateDir        string
	ThemeDir        string

	State		MOTHState

	AttemptInterval time.Duration

	Runtime RuntimeConfig

	categories        map[string]*Mothball
	MaxPointsUnlocked map[string]int
	update            chan bool
	jPuzzleList       []byte
	jPointsLog        []byte
	nextAttempt       map[string]time.Time
	nextAttemptMutex  *sync.RWMutex
	mux               *http.ServeMux
}

func (ctx *Instance) Initialize() error {
	// Roll over and die if directories aren't even set up
	if _, err := os.Stat(ctx.MothballDir); err != nil {
		return err
	}

	if _, err := ctx.State.Initialize(); err != nil {
		return err
	}

	ctx.Base = strings.TrimRight(ctx.Base, "/")
	ctx.categories = map[string]*Mothball{}
	ctx.update = make(chan bool, 10)
	ctx.nextAttempt = map[string]time.Time{}
	ctx.nextAttemptMutex = new(sync.RWMutex)
	ctx.mux = http.NewServeMux()

	ctx.BindHandlers()

	return nil
}

// Stuff people with mediocre handwriting could write down unambiguously, and can be entered without holding down shift
const distinguishableChars = "234678abcdefhijkmnpqrtwxyz="

func mktoken() string {
	a := make([]byte, 8)
	for i := range a {
		char := rand.Intn(len(distinguishableChars))
		a[i] = distinguishableChars[char]
	}
	return string(a)
}

func pathCleanse(parts []string) string {
	clean := make([]string, len(parts))
	for i := range parts {
		part := parts[i]
		part = strings.TrimLeft(part, ".")
		if p := strings.LastIndex(part, "/"); p >= 0 {
			part = part[p+1:]
		}
		clean[i] = part
	}
	return path.Join(clean...)
}

func (ctx Instance) MothballPath(parts ...string) string {
	tail := pathCleanse(parts)
	return path.Join(ctx.MothballDir, tail)
}

func (ctx *Instance) ThemePath(parts ...string) string {
	tail := pathCleanse(parts)
	return path.Join(ctx.ThemeDir, tail)
}

func (ctx *Instance) TooFast(teamId string) bool {
	now := time.Now()

	ctx.nextAttemptMutex.RLock()
	next, _ := ctx.nextAttempt[teamId]
	ctx.nextAttemptMutex.RUnlock()

	ctx.nextAttemptMutex.Lock()
	ctx.nextAttempt[teamId] = now.Add(ctx.AttemptInterval)
	ctx.nextAttemptMutex.Unlock()

	return now.Before(next)
}

func (ctx *Instance) OpenCategoryFile(category string, parts ...string) (io.ReadCloser, error) {
	mb, ok := ctx.categories[category]
	if !ok {
		return nil, fmt.Errorf("No such category: %s", category)
	}

	filename := path.Join(parts...)
	f, err := mb.Open(filename)
	return f, err
}

func (ctx *Instance) ValidTeamId(teamId string) bool {
	ctx.nextAttemptMutex.RLock()
	_, ok := ctx.nextAttempt[teamId]
	ctx.nextAttemptMutex.RUnlock()

	return ok
}
