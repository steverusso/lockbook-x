package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	lb "github.com/steverusso/lockbook-x/go-lockbook"
)

func importFile(core lb.Core, diskTarget, lbDest string) error {
	return nil
}

func exportFile(core lb.Core, isVerbose bool, target, dest string) error {
	id, err := idFromSomething(core, target)
	if err != nil {
		return fmt.Errorf("trying to get an id from %q: %w", target, err)
	}
	f, err := core.FileByID(id)
	if err != nil {
		return fmt.Errorf("file by id %q: %w", id, err)
	}
	// If no destination path is provided, it'll be a file with the target name in the
	// current directory. If it's root, it'll be the account's username.
	if dest == "" {
		dest = "."
	}
	if f.IsRoot() {
		acct, err := core.GetAccount()
		if err != nil {
			return fmt.Errorf("getting account: %w", err)
		}
		dest = filepath.Join(dest, acct.Username)
	}
	var forEach func(lb.ImportExportFileInfo)
	if isVerbose {
		forEach = func(info lb.ImportExportFileInfo) {
			fmt.Printf("%s\n", info.LbPath)
		}
	}
	// Ensure the destination's parent directory exists.
	if !f.IsDir() && path.Base(dest) == f.Name {
		dest = path.Dir(dest)
	}
	if err := os.MkdirAll(dest, os.ModeDir); err != nil {
		return fmt.Errorf("creating directory %s: %w", dest, err)
	}
	if err := core.ExportFile(id, dest, forEach); err != nil {
		return fmt.Errorf("exporting: %w", err)
	}
	return nil
}

func exportDrawing(core lb.Core, target, imgFmt string) error {
	id, err := idFromSomething(core, target)
	if err != nil {
		return fmt.Errorf("trying to get id from %q: %w", target, err)
	}

	imgFmtCode := lb.ImgFmtPNG
	switch imgFmt {
	case "jpg", "jpeg":
		imgFmtCode = lb.ImgFmtJPEG
	case "pnm":
		imgFmtCode = lb.ImgFmtPNM
	case "tga":
		imgFmtCode = lb.ImgFmtTGA
	case "farbfeld":
		imgFmtCode = lb.ImgFmtFarbfeld
	case "bmp":
		imgFmtCode = lb.ImgFmtBMP
	default:
		fmt.Printf("unknown image format %q, defaulting to png...", imgFmt)
	}

	data, err := core.ExportDrawing(id, imgFmtCode)
	if err != nil {
		return fmt.Errorf("exporting drawing: %w", err)
	}
	fmt.Print(data)
	return nil
}
