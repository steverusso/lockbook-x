package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/steverusso/lockbook-x/go-lockbook"
)

type splashScreen struct {
	updates chan<- legitUpdate
	status  string
	errMsg  string
}

type splashWorkStep int

const (
	splashInit splashWorkStep = iota
	splashOpenCore
	splashCheckForAccount
	splashSync
	splashGetRootFiles
	splashGetLastSynced
)

func (s *splashScreen) layout(gtx C, th *material.Theme) D {
	var lbl material.LabelStyle
	if s.errMsg != "" {
		lbl = material.Body1(th, s.errMsg)
		lbl.Color = color.NRGBA{255, 10, 10, 255}
	} else {
		lbl = material.Body1(th, s.status)
	}
	return layout.Center.Layout(gtx, lbl.Layout)
}

func (s *splashScreen) setStep(step splashWorkStep) {
	switch step {
	case splashInit:
		s.status = "Initializing..."
	case splashOpenCore:
		s.status = "Loading core..."
	case splashCheckForAccount:
		s.status = "Checking for account..."
	case splashSync:
		s.status = "Syncing..."
	case splashGetRootFiles:
		s.status = "Getting root children..."
	case splashGetLastSynced:
		s.status = "Getting last synced..."
	default:
		s.status = "splashWorkStep(" + strconv.FormatInt(int64(step), 10) + ")"
	}
	s.invalidate()
}

func (s *splashScreen) setError(ctx string, err error) {
	s.errMsg = fmt.Sprintf("error: %s: %s", ctx, err.Error())
	s.invalidate()
}

func (s *splashScreen) invalidate() {
	s.updates <- nil
}

func (s *splashScreen) doStartupWork() {
	s.setStep(splashOpenCore)
	dir := getDataDir()
	core, err := lockbook.NewCore(dir)
	if err != nil {
		s.setError("initializing lockbook-core", err)
		return
	}
	// Determine whether we're going to the onboard screen or the workspace by checking
	// for an account.
	s.setStep(splashCheckForAccount)
	if _, err = core.GetAccount(); err != nil {
		if err, ok := err.(*lockbook.Error); ok && err.Code == lockbook.CodeNoAccount {
			s.updates <- handoffToOnboard{core: core}
			return
		}
		s.setError("getting account", err)
		return
	}
	// Gather the errors that shouldn't prohibit the user from getting to their workspace.
	errs := []error{}

	s.setStep(splashSync)
	if err = core.SyncAll(nil); err != nil {
		errs = append(errs, fmt.Errorf("performing sync on open: %s", err))
	}

	s.setStep(splashGetRootFiles)
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

	s.setStep(splashGetLastSynced)
	lastSynced, err := core.GetLastSyncedHumanString()
	if err != nil {
		errs = append(errs, fmt.Errorf("getting last synced: %s", err))
	}

	s.updates <- handoffToWorkspace{
		core:       core,
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
