package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"

	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/steverusso/lockbook-x/go-lockbook"
)

type splashScreen struct {
	updates chan<- legitUpdate
	status  string
	errMsg  string
}

type handoffToOnboard struct {
	core lockbook.Core
}

type handoffToWorkspace struct {
	core       lockbook.Core
	lastSynced string
	root       lockbook.File
	rootFiles  []lockbook.File
	errs       []error
}

type setSplashErr struct {
	msg string
}

func (s *splashScreen) layout(gtx C, th *material.Theme) D {
	var lbl material.LabelStyle
	if s.errMsg != "" {
		lbl = material.Body1(th, s.errMsg)
		lbl.Color = color.NRGBA{255, 10, 10, 255}
	} else {
		lbl = material.Body1(th, "Initializing...")
	}
	return layout.Center.Layout(gtx, lbl.Layout)
}

func (s *splashScreen) setError(ctx string, err error) {
	s.updates <- setSplashErr{
		msg: fmt.Sprintf("error: %s: %s", ctx, err.Error()),
	}
}

func (s *splashScreen) doStartupWork() {
	dir := getDataDir()
	core, err := lockbook.NewCore(dir)
	if err != nil {
		s.setError("initializing lockbook-core", err)
		return
	}
	// Determine whether we're going to the onboard screen or the workspace by checking
	// for an account.
	if _, err = core.GetAccount(); err != nil {
		if err, ok := err.(*lockbook.Error); ok && err.Code == lockbook.CodeAccountNonexistent {
			s.updates <- handoffToOnboard{core: core}
			return
		}
		s.setError("getting account", err)
		return
	}
	// Gather the errors that shouldn't prohibit the user from getting to their workspace.
	errs := []error{}

	if err = core.SyncAll(nil); err != nil {
		errs = append(errs, fmt.Errorf("performing sync on open: %s", err))
	}

	root, err := core.GetRoot()
	if err != nil {
		s.setError("getting root", err)
		return
	}
	rootFiles, err := core.GetChildren(root.ID)
	if err != nil {
		s.setError("getting root children", err)
		return
	}
	lockbook.SortFiles(rootFiles)

	lastSynced, err := core.GetLastSyncedHumanString()
	if err != nil {
		errs = append(errs, fmt.Errorf("getting last synced: %s", err))
	}

	s.updates <- handoffToWorkspace{
		core:       core,
		root:       root,
		rootFiles:  rootFiles,
		lastSynced: lastSynced,
		errs:       errs,
	}
}

func getDataDir() string {
	lbPath := os.Getenv("LOCKBOOK_PATH")
	if lbPath != "" {
		return lbPath
	}
	home, err := os.UserHomeDir()
	if err == nil {
		return filepath.Join(home, ".lockbook/cli")
	}
	log.Printf("getting user home dir: %v", err)
	return filepath.Join(".", ".lockbook")
}
