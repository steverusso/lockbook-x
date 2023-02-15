package main

import (
	"fmt"
	"path"
	"sort"

	"github.com/steverusso/lockbook-x/go-lockbook"
)

// list files in a directory
type lsCmd struct {
	short     bool   `opt:"short,s" desc:"just display the name (or file path)"`
	recursive bool   `opt:"recursive,r" desc:"recursively include all children of the target directory"`
	paths     bool   `opt:"paths" desc:"show absolute file paths instead of file names"`
	onlyDirs  bool   `opt:"dirs" desc:"only show folders"`
	onlyDocs  bool   `opt:"docs" desc:"only show documents"`
	fullIDs   bool   `opt:"ids" desc:"show full uuids instead of prefixes"`
	target    string `arg:"target directory (defaults to root)"`
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

func (node *fileNode) text(cfg *lsConfig) (s string) {
	if !cfg.short {
		s += fmt.Sprintf("%-*s  ", cfg.idWidth, node.id.String()[:cfg.idWidth])
	}
	nameOrPath := node.name
	if cfg.paths {
		nameOrPath = fmt.Sprintf("%s%s", node.dirName, node.name)
	}
	s += fmt.Sprintf("%-*s", cfg.nameWidth, nameOrPath)
	if !cfg.short {
		if node.sharedBy != "" {
			s += "   @" + node.sharedBy + " "
		} else {
			s += "   "
		}
		if node.sharedWith != "" {
			s += "-> " + node.sharedWith
		}
	}
	return
}

func (node *fileNode) printOut(cfg *lsConfig) {
	if (!cfg.onlyDirs && !cfg.onlyDocs) || (cfg.onlyDirs && node.isDir) || (cfg.onlyDocs && !node.isDir) {
		fmt.Println(node.text(cfg))
	}
	for i := range node.children {
		node.children[i].printOut(cfg)
	}
}

func getChildren(core lockbook.Core, files []lockbook.File, parent lockbook.FileID, cfg *lsConfig) ([]fileNode, error) {
	lockbook.SortFiles(files)
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
		fpath, err := core.PathByID(f.ID)
		if err != nil {
			return nil, fmt.Errorf("getting path for %q: %w", f.ID, err)
		}
		dirName := path.Dir(path.Clean(fpath))
		if dirName != "/" {
			dirName += "/"
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
			n = len(fmt.Sprintf("%s%s", dirName, name))
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
	f, err := core.FileByPath(ls.target)
	if err != nil {
		return fmt.Errorf("getting file by path %q: %w", ls.target, err)
	}
	var files []lockbook.File
	{
		var err error
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
	for i := range files {
		if files[i].IsRoot() {
			files = append(files[:i], files[i+1:]...)
			break
		}
	}
	acct, err := core.GetAccount()
	if err != nil {
		return fmt.Errorf("getting account: %v", err)
	}
	idWidth := idPrefixLen
	if ls.fullIDs {
		idWidth = len(f.ID)
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
