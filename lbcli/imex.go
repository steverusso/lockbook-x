package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/steverusso/lockbook-x/go-lockbook"
)

// export a lockbook drawing as an image written to stdout
//
// clap:cmd_usage <target> [png|jpeg|pnm|tga|farbfeld|bmp]
type drawingCmd struct {
	// the drawing to export
	//
	// clap:arg_required
	target string
	// the format to convert the drawing into
	imgFmt string
}

// copy a lockbook file to your file system
//
// clap:cmd_usage <target> [dest-dir]
type exportCmd struct {
	// print out each file as it's being exported
	//
	// clap:opt verbose,v
	verbose bool
	// lockbook file path or id
	//
	// clap:arg_required
	target string
	// disk file path (defaults to working dir)
	dest string
}

func (c *exportCmd) run(core lockbook.Core) error {
	id, err := idFromSomething(core, c.target)
	if err != nil {
		return fmt.Errorf("trying to get an id from %q: %w", c.target, err)
	}
	f, err := core.FileByID(id)
	if err != nil {
		return fmt.Errorf("file by id %q: %w", id, err)
	}
	// If no destination path is provided, it'll be a file with the target name in the
	// current directory. If it's root, it'll be the account's username.
	if c.dest == "" {
		c.dest = "."
	}
	if f.IsRoot() {
		acct, err := core.GetAccount()
		if err != nil {
			return fmt.Errorf("getting account: %w", err)
		}
		c.dest = filepath.Join(c.dest, acct.Username)
	}
	var forEach func(lockbook.ImportExportFileInfo)
	if c.verbose {
		forEach = func(info lockbook.ImportExportFileInfo) {
			fmt.Printf("%s\n", info.LbPath)
		}
	}
	// Ensure the destination's parent directory exists.
	if !f.IsDir() && path.Base(c.dest) == f.Name {
		c.dest = path.Dir(c.dest)
	}
	if err := os.MkdirAll(c.dest, os.ModeDir); err != nil {
		return fmt.Errorf("creating directory %s: %w", c.dest, err)
	}
	if err := core.ExportFile(id, c.dest, forEach); err != nil {
		return fmt.Errorf("exporting: %w", err)
	}
	return nil
}

func (c *drawingCmd) run(core lockbook.Core) error {
	id, err := idFromSomething(core, c.target)
	if err != nil {
		return fmt.Errorf("trying to get id from %q: %w", c.target, err)
	}

	imgFmtCode := lockbook.ImgFmtPNG
	switch c.imgFmt {
	case "jpg", "jpeg":
		imgFmtCode = lockbook.ImgFmtJPEG
	case "pnm":
		imgFmtCode = lockbook.ImgFmtPNM
	case "tga":
		imgFmtCode = lockbook.ImgFmtTGA
	case "farbfeld":
		imgFmtCode = lockbook.ImgFmtFarbfeld
	case "bmp":
		imgFmtCode = lockbook.ImgFmtBMP
	default:
		fmt.Printf("unknown image format %q, defaulting to png...", c.imgFmt)
	}

	data, err := core.ExportDrawing(id, imgFmtCode)
	if err != nil {
		return fmt.Errorf("exporting drawing: %w", err)
	}
	fmt.Print(data)
	return nil
}
