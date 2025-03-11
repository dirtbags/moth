package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dirtbags/moth/v5/docs"
	"github.com/dirtbags/moth/v5/pkg/wfs"
)

var (
	// Version of this program
	Version = "[devel]"

	// Date this version came out
	VersionDate = "????-??-??"
)

type Context struct {
	CategoryStore
	StateStore
	WebFS http.FileSystem

	*http.ServeMux
}

func (ctx *Context) Refresh(now time.Time) error {
	var errs []error
	if err := ctx.CategoryStore.Refresh(now); err != nil {
		errs = append(errs, err)
	}
	if err := ctx.StateStore.Refresh(now); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [FLAGS] PUZZLEDIR [PUZZLEDIR...]:\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintln(flag.CommandLine.Output(), "Version", Version, "-", VersionDate)
}

func run() error {
	webPath := flag.String(
		"web",
		"web",
		"Path to static web files (the user interface)",
	)
	categoriesPath := flag.String(
		"categories",
		"categories",
		"Path to categories tree",
	)
	statePath := flag.String(
		"state",
		"state",
		"Path to event state directories",
	)
	refreshInterval := flag.Duration(
		"refresh",
		2*time.Second,
		"Duration between maintenance tasks",
	)
	bindStr := flag.String(
		"bind",
		":8080",
		"Bind [host]:port for HTTP service",
	)
	base := flag.String(
		"base",
		"/",
		"Base URL of this instance",
	)

	Version, VersionDate = docs.Version()

	flag.Usage = usage
	flag.Parse()

	ctx := &Context{}

	if p, err := filepath.Abs(*webPath); err != nil {
		return err
	} else if _, err := os.Stat(p); err != nil {
		return err
	} else {
		ctx.WebFS = http.Dir(p)
	}

	if p, err := filepath.Abs(*categoriesPath); err != nil {
		return err
	} else if _, err := os.Stat(p); err != nil {
		return err
	} else {
		ctx.CategoryStore = NewFSCategoryStore(os.DirFS(p))
	}

	if p, err := filepath.Abs(*statePath); err != nil {
		return err
	} else if hostname, err := os.Hostname(); err != nil {
		return err
	} else if _, err := os.Stat(p); err != nil {
		return err
	} else {
		ctx.StateStore = NewFSStateStore(wfs.DirFS(p), hostname, os.Getpid(), 1)
	}

	if err := ctx.Refresh(time.Now()); err != nil {
		slog.Error(err.Error())
	}

	ln, err := net.Listen("tcp", *bindStr)
	if err != nil {
		return err
	}
	ctx.InstallHandlers(*base)
	go http.Serve(ln, ctx)

	slog.Info("started", "addr", *bindStr)

	for now := range time.Tick(*refreshInterval) {
		if err := ctx.Refresh(now); err != nil {
			slog.Error(err.Error())
		}
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
