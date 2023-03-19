package main

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/steverusso/lockbook-x/go-lockbook"
)

// List files in a directory.
type lsCmd struct {
	// Just display the name (or file path).
	//
	// clap:opt short,s
	short bool
	// Recursively include all children of the target directory.
	//
	// clap:opt recursive,r
	recursive bool
	// Show absolute file paths instead of file names.
	//
	// clap:opt paths
	paths bool
	// Only show folders.
	//
	// clap:opt dirs
	onlyDirs bool
	// Only show documents.
	//
	// clap:opt docs
	onlyDocs bool
	// Show full UUIDs instead of prefixes.
	//
	// clap:opt ids
	fullIDs bool
	// Path or ID of the target directory (defaults to root).
	target string
}

type lsConfig struct {
	myName    string
	idWidth   int
	nameWidth int
	short     bool
	paths     bool
	onlyDirs  bool
	onlyDocs  bool
	fullIDs   bool
}

type fileNode struct {
	id         lockbook.FileID
	dirName    string
	name       string
	isDir      bool
	sharedWith string
	sharedBy   string
	children   []fileNode
}

func (fn *fileNode) text(cfg *lsConfig) string {
	var s strings.Builder
	if !cfg.short {
		fmt.Fprintf(&s, "%-*s  ", cfg.idWidth, fn.id.String()[:cfg.idWidth])
	}
	nameOrPath := fn.name
	if cfg.paths {
		nameOrPath = fn.dirName + fn.name
	}
	fmt.Fprintf(&s, "%-*s", cfg.nameWidth, nameOrPath)
	if !cfg.short {
		if fn.sharedBy != "" {
			fmt.Fprintf(&s, "   @%s ", fn.sharedBy)
		} else {
			s.WriteString("   ")
		}
		if fn.sharedWith != "" {
			s.WriteString("-> ")
			s.WriteString(fn.sharedWith)
		}
	}
	return s.String()
}

func (fn *fileNode) printOut(cfg *lsConfig) {
	// Print if there are no type filters or if this file is the desired file type.
	if (!cfg.onlyDirs && !cfg.onlyDocs) ||
		(cfg.onlyDirs && fn.isDir) ||
		(cfg.onlyDocs && !fn.isDir) {
		fmt.Println(fn.text(cfg))
	}
	for i := range fn.children {
		fn.children[i].printOut(cfg)
	}
}

func getChildren(core lockbook.Core, files []lockbook.File, parent lockbook.FileID, cfg *lsConfig) ([]fileNode, error) {
	children := []fileNode{}
	for i := range files {
		f := &files[i]
		if f.Parent != parent {
			continue
		}
		// File name.
		name := f.Name
		if f.IsDir() {
			name += "/"
		}
		// Parent directory.
		dirName := ""
		if cfg.paths {
			fpath, err := core.PathByID(f.ID)
			if err != nil {
				return nil, fmt.Errorf("getting path for %q: %w", f.ID, err)
			}
			dirName = path.Dir(path.Clean(fpath))
			if dirName != "/" {
				dirName += "/"
			}
		}
		// Share info.
		sharedWiths := []string{}
		sharedBy := ""
		for _, sh := range f.Shares {
			if sh.SharedWith == cfg.myName {
				sharedBy = sh.SharedBy
			}
			if sh.SharedWith == cfg.myName {
				sharedWiths = append(sharedWiths, "me")
			} else {
				sharedWiths = append(sharedWiths, "@"+sh.SharedWith)
			}
		}
		sort.SliceStable(sharedWiths, func(i, j int) bool {
			return len(sharedWiths[i]) < len(sharedWiths[j])
		})
		sharedWith := ""
		if n := len(sharedWiths); n == 1 {
			sharedWith = sharedWiths[0]
		} else if n == 2 {
			sharedWith = sharedWiths[0] + " and " + sharedWiths[1]
		} else if n != 0 {
			sharedWith = fmt.Sprintf("%s, %s, and %d more", sharedWiths[0], sharedWiths[1], n-2)
		}
		// Determine column widths.
		n := len(name)
		if cfg.paths {
			n += len(dirName)
		}
		if n > cfg.nameWidth {
			cfg.nameWidth = n
		}
		child := fileNode{
			id:         f.ID,
			dirName:    dirName,
			name:       name,
			isDir:      f.IsDir(),
			sharedWith: sharedWith,
			sharedBy:   sharedBy,
		}
		childsChildren, err := getChildren(core, files, f.ID, cfg)
		if err != nil {
			return nil, fmt.Errorf("getting children for %q: %w", f.ID, err)
		}
		child.children = childsChildren
		children = append(children, child)
	}
	return children, nil
}

func (ls *lsCmd) run(core lockbook.Core) error {
	if ls.target == "" {
		ls.target = "/"
	}
	id, err := idFromSomething(core, ls.target)
	if err != nil {
		return fmt.Errorf("trying to get target from %q: %w", ls.target, err)
	}
	f, err := core.FileByID(id)
	if err != nil {
		return fmt.Errorf("getting file by id %q: %w", id, err)
	}
	var files []lockbook.File
	{
		if ls.recursive {
			files, err = core.GetAndGetChildrenRecursively(f.ID)
			if err != nil {
				return fmt.Errorf("getting children recursively for %q: %w", f.ID, err)
			}
		} else {
			files, err = core.GetChildren(f.ID)
			if err != nil {
				return fmt.Errorf("getting children for %q: %w", f.ID, err)
			}
		}
	}
	lockbook.SortFiles(files)
	if f.IsRoot() {
		for i := range files {
			if files[i].IsRoot() {
				files = append(files[:i], files[i+1:]...)
				break
			}
		}
	}
	acct, err := core.GetAccount()
	if err != nil {
		return fmt.Errorf("getting account: %v", err)
	}
	idWidth := idPrefixLen
	if ls.fullIDs {
		idWidth = len(f.ID.String())
	}
	cfg := lsConfig{
		myName:   acct.Username,
		idWidth:  idWidth,
		short:    ls.short,
		paths:    ls.paths,
		onlyDirs: ls.onlyDirs,
		onlyDocs: ls.onlyDocs,
		fullIDs:  ls.fullIDs,
	}
	infos, err := getChildren(core, files, f.ID, &cfg)
	if err != nil {
		return fmt.Errorf("getting child nodes: %w", err)
	}
	for i := range infos {
		infos[i].printOut(&cfg)
	}
	return nil
}
