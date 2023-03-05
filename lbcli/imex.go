package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/steverusso/lockbook-x/go-lockbook"
)

// import files into lockbook from your system
//
// clap:cmd_name import
type importCmd struct {
	// don't output progress on each file
	//
	// clap:opt quiet,q
	quiet bool
	// the file(s) to import into lockbook
	//
	// clap:arg_required
	diskPath string
	// where to put the imported files in lockbook
	dest string
}

func (c *importCmd) run(core lockbook.Core) error {
	// Determine the destination target's ID. If no destination was provided, use root.
	var destID lockbook.FileID
	if c.dest == "" {
		root, err := core.GetRoot()
		if err != nil {
			return fmt.Errorf("getting root: %w", err)
		}
		destID = root.ID
	} else {
		id, err := idFromSomething(core, c.dest)
		if err != nil {
			return fmt.Errorf("trying to get an id from %q: %w", c.dest, err)
		}
		destID = id
	}
	dest, err := core.FileByID(destID)
	if err != nil {
		return fmt.Errorf("file by id %q: %w", destID, err)
	}
	if !dest.IsDir() {
		destID = dest.Parent
	}
	var forEach func(lockbook.ImportFileInfo)
	if !c.quiet {
		var total int
		var count int
		forEach = func(info lockbook.ImportFileInfo) {
			switch {
			case info.Total != 0:
				total = info.Total
			case info.DiskPath != "":
				count++
				fmt.Printf("(%d/%d) %s ... ", count, total, info.DiskPath)
			case info.FileDone != nil:
				fmt.Println("done")
			}
		}
	}
	err = core.ImportFile(c.diskPath, destID, forEach)
	if err != nil {
		return fmt.Errorf("importing '%s': %w", c.diskPath, err)
	}
	return nil
}

// copy a lockbook file to your file system
//
// clap:cmd_usage [--quiet] <target> [dest-dir]
// clap:cmd_usage [--img-fmt <fmt>] <drawing> [dest-dir]
type exportCmd struct {
	// format for exporting a lockbook drawing (png|jpeg|pnm|tga|farbfeld|bmp)
	//
	// clap:opt img-fmt,i
	imgFmt string
	// don't output progress on each file
	//
	// clap:opt quiet,q
	quiet bool
	// lockbook file path or id
	//
	// clap:arg_required
	target string
	// disk file path (default ".")
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
	// Check if we're explicitly exporting a lockbook drawing to an image.
	if !f.IsDir() && filepath.Ext(f.Name) == "draw" && c.imgFmt != "" {
		if err := exportDrawing(core, id, c.imgFmt); err != nil {
			return fmt.Errorf("exporting drawing: %w", err)
		}
		return nil
	}
	if c.imgFmt != "" {
		fmt.Fprintln(os.Stderr, "ignoring '--img-fmt' option because target is not a drawing")
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
	var forEach func(lockbook.ExportFileInfo)
	if !c.quiet {
		forEach = func(info lockbook.ExportFileInfo) {
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

func exportDrawing(core lockbook.Core, id lockbook.FileID, imgFmt string) error {
	var imgFmtCode lockbook.ImageFormat
	switch imgFmt {
	case "png":
		imgFmtCode = lockbook.ImgFmtPNG
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
		fmt.Printf("unknown image format %q, defaulting to png...", imgFmt)
	}

	data, err := core.ExportDrawing(id, imgFmtCode)
	if err != nil {
		return fmt.Errorf("exporting drawing: %w", err)
	}
	fmt.Print(data)
	return nil
}
