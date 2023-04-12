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
	// Recursively list children of the target directory.
	//
	// clap:opt recursive,r
	recursive bool
	// Recursively list children of the target directory in a tree format.
	//
	// clap:opt tree,t
	tree bool
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
	tree      bool
	paths     bool
	onlyDirs  bool
	onlyDocs  bool
	fullIDs   bool
}

type fileNode struct {
	id         lockbook.FileID
	treePrefix string
	dirName    string
	name       string
	isDir      bool
	shared     lsShareInfo
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
	if cfg.tree {
		nameOrPath = fn.treePrefix + nameOrPath
	}
	fmt.Fprintf(&s, "%-*s", cfg.nameWidth, nameOrPath)
	if !cfg.short {
		if fn.shared.by != "" {
			fmt.Fprintf(&s, "   @%s ", fn.shared.by)
		} else {
			s.WriteString("   ")
		}
		if fn.shared.with != "" {
			s.WriteString("-> ")
			s.WriteString(fn.shared.with)
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
		// Determine column widths.
		nameLen := len(name)
		if cfg.paths {
			nameLen += len(dirName)
		}
		if nameLen > cfg.nameWidth {
			cfg.nameWidth = nameLen
		}
		child := fileNode{
			id:      f.ID,
			dirName: dirName,
			name:    name,
			isDir:   f.IsDir(),
			shared:  getShareInfo(f.Shares, cfg.myName),
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

type branch int

const (
	branchDefault   branch = iota // Any child node except the last.
	branchLastChild               // Last sibling node in list of children.
	branchNone                    // Empty space.
)

// treeifyNode sets the given file node's tree prefix and makes the adjustments for the
// maximum name length config.
func treeifyNode(cfg *lsConfig, fn *fileNode, branchSlots []branch) {
	fn.treePrefix = getTreePrefix(branchSlots)
	// Adjust for longest name length, accounting for the tree prefixes now.
	{
		nameLen := len(fn.name)
		if cfg.paths {
			nameLen += len(fn.dirName)
		}
		if w := nameLen + len(fn.treePrefix); w > cfg.nameWidth {
			cfg.nameWidth = w
		}
	}
	// If this node was the last child of its parent, then we no longer show its branch
	// for the rest of the grandchildren.
	if n := len(branchSlots); n > 0 && branchSlots[n-1] == branchLastChild {
		branchSlots[n-1] = branchNone
	}
	for i := range fn.children {
		b := branchDefault
		if i == len(fn.children)-1 {
			b = branchLastChild
		}
		treeifyNode(cfg, &fn.children[i], append(branchSlots, b))
	}
}

func getTreePrefix(slots []branch) string {
	var s strings.Builder
	for i, b := range slots {
		switch b {
		case branchDefault:
			if i == len(slots)-1 {
				s.WriteString("├── ")
			} else {
				s.WriteString("│   ")
			}
		case branchLastChild:
			s.WriteString("└── ")
		case branchNone:
			s.WriteString("    ")
		}
	}
	return s.String()
}

type lsShareInfo struct {
	by   string
	with string
}

func getShareInfo(shares []lockbook.Share, myName string) lsShareInfo {
	by := ""
	withs := make([]string, 0, len(shares))
	for _, sh := range shares {
		if sh.SharedWith == myName {
			by = sh.SharedBy
			withs = append(withs, "me")
		} else {
			withs = append(withs, "@"+sh.SharedWith)
		}
	}
	sort.SliceStable(withs, func(i, j int) bool {
		return len(withs[i]) < len(withs[j])
	})
	with := ""
	if n := len(withs); n == 1 {
		with = withs[0]
	} else if n == 2 {
		with = withs[0] + " and " + withs[1]
	} else if n != 0 {
		with = fmt.Sprintf("%s, %s, and %d more", withs[0], withs[1], n-2)
	}
	return lsShareInfo{by: by, with: with}
}

func (ls *lsCmd) run(core lockbook.Core) error {
	if ls.tree {
		ls.recursive = true
	}
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
		tree:     ls.tree,
		paths:    ls.paths,
		onlyDirs: ls.onlyDirs,
		onlyDocs: ls.onlyDocs,
		fullIDs:  ls.fullIDs,
	}
	infos, err := getChildren(core, files, f.ID, &cfg)
	if err != nil {
		return fmt.Errorf("getting child nodes: %w", err)
	}
	if ls.tree {
		for i := range infos {
			treeifyNode(&cfg, &infos[i], []branch{})
		}
	}
	for i := range infos {
		infos[i].printOut(&cfg)
	}
	return nil
}
