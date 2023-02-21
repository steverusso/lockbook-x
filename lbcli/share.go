package main

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/steverusso/lockbook-x/go-lockbook"
)

// sharing related commands
type shareCmd struct {
	create  *shareCreateCmd
	pending *sharePendingCmd
	accept  *shareAcceptCmd
	reject  *shareRejectCmd
}

// share a file with another lockbook user
type shareCreateCmd struct {
	// the other user will not be able to edit the file
	//
	// clap:opt ro
	readOnly bool
	// the path or id of the lockbook file you'd like to share
	//
	// clap:arg_required
	target string
	// the username of the other lockbook user
	//
	// clap:arg_required
	username string
}

// list pending shares
type sharePendingCmd struct {
	// print full uuids instead of prefixes
	//
	// clap:opt ids
	fullIDs bool
}

// accept a pending share
type shareAcceptCmd struct {
	// id or id prefix of the pending share to accept
	//
	// clap:arg_required
	target string
	// where to place this in your file tree
	//
	// clap:arg_required
	dest string
	// name this file something else
	newName string
}

// reject a pending share
type shareRejectCmd struct {
	// id or id prefix of a pending share
	//
	// clap:arg_required
	target string
}

func (c *shareCreateCmd) run(core lockbook.Core) error {
	id, err := idFromSomething(core, c.target)
	if err != nil {
		return fmt.Errorf("trying to get id from %q: %w", c.target, err)
	}
	mode := lockbook.ShareModeWrite
	if c.readOnly {
		mode = lockbook.ShareModeRead
	}
	if err = core.ShareFile(id, c.username, mode); err != nil {
		return fmt.Errorf("sharing file: %w", err)
	}
	fmt.Printf("\033[1;32mdone!\033[0m file %q will be shared next time you sync.\n", id)
	return nil
}

func (c *sharePendingCmd) run(core lockbook.Core) error {
	pendingShares, err := core.GetPendingShares()
	if err != nil {
		return fmt.Errorf("getting pending shares: %w", err)
	}
	if len(pendingShares) == 0 {
		fmt.Println("no pending shares")
		return nil
	}
	shareInfos := filesToShareInfos(pendingShares)
	fmt.Println(shareInfoTable(shareInfos, c.fullIDs))
	return nil
}

func (c *shareAcceptCmd) run(core lockbook.Core) error {
	pendingShares, err := core.GetPendingShares()
	if err != nil {
		return fmt.Errorf("getting pending shares: %w", err)
	}
	if len(pendingShares) == 0 {
		return errors.New("no pending shares to accept")
	}
	match, err := getOnePendingShareMatch(pendingShares, c.target)
	if err != nil {
		return err
	}
	parentID := lockbook.FileID{}
	// If the destination is a valid UUID, it must be of an existing directory.
	if destID := uuid.FromStringOrNil(c.dest); !destID.IsNil() {
		f, err := core.FileByID(destID)
		if err != nil {
			return fmt.Errorf("file by id %q: %w", destID, err)
		}
		if !f.IsDir() {
			return fmt.Errorf("destination id %q isn't a folder", destID)
		}
		parentID = destID
	} else {
		// If the destination path exists, it must be a directory. The link will be
		// dropped in it.
		f, exists, err := maybeFileByPath(core, c.dest)
		if err != nil {
			return fmt.Errorf("file by path %q: %w", c.dest, err)
		}
		if !exists {
			// If the destination path doesn't exist, then it's just treated as a
			// non-existent directory path. The user can choose a name with the `--rename`
			// flag.
			fPath := c.dest
			if fPath[len(fPath)-1] != '/' {
				fPath += "/"
			}
			newFile, err := core.CreateFileAtPath(fPath)
			if err != nil {
				return fmt.Errorf("creating file at path %q: %w", fPath, err)
			}
			parentID = newFile.ID
		} else {
			if !f.IsDir() {
				return fmt.Errorf("existing destination path %q is a doc, must be a folder", c.dest)
			}
			parentID = f.ID
		}
	}

	name := c.newName
	if name == "" {
		name = match.Name
	}
	if name[len(name)-1] == '/' {
		name = name[:len(name)-1] // Prevent "name contains slash" error.
	}

	_, err = core.CreateFile(name, parentID, lockbook.FileTypeLink{Target: match.ID})
	if err != nil {
		return fmt.Errorf("creating link: %w", err)
	}
	return nil
}

func (c *shareRejectCmd) run(core lockbook.Core) error {
	pendingShares, err := core.GetPendingShares()
	if err != nil {
		return fmt.Errorf("getting pending shares: %w", err)
	}
	if len(pendingShares) == 0 {
		return errors.New("no pending shares to delete")
	}
	match, err := getOnePendingShareMatch(pendingShares, c.target)
	if err != nil {
		return err
	}
	if err = core.DeletePendingShare(match.ID); err != nil {
		return fmt.Errorf("delete pending share: %w", err)
	}
	return nil
}

func getOnePendingShareMatch(shares []lockbook.File, id string) (lockbook.File, error) {
	matches := []lockbook.File{}
	for _, f := range shares {
		if strings.HasPrefix(f.ID.String(), id) {
			matches = append(matches, f)
		}
	}
	n := len(matches)
	if len(matches) == 0 {
		desc := ""
		if t := uuid.FromStringOrNil(id); !t.IsNil() {
			desc = " prefix"
		}
		return lockbook.File{}, fmt.Errorf("no pending share found with id%s %s", desc, id)
	}
	if n > 1 {
		return lockbook.File{}, fmt.Errorf(
			"id prefix %q matched %d pending shares:\n%s",
			id, n, shareInfoTable(filesToShareInfos(matches), true),
		)
	}
	return matches[0], nil
}

type shareInfo struct {
	id   lockbook.FileID
	from string
	name string
	mode string
}

func filesToShareInfos(files []lockbook.File) []shareInfo {
	infos := make([]shareInfo, len(files))
	for i, f := range files {
		from, mode := "", ""
		if len(f.Shares) > 0 {
			from = f.Shares[0].SharedBy
			mode = strings.ToLower(f.Shares[0].Mode.String())
		}
		infos[i] = shareInfo{
			id:   f.ID,
			from: from,
			name: f.Name,
			mode: mode,
		}
	}
	sort.SliceStable(infos, func(i, j int) bool {
		a, b := infos[i], infos[j]
		if a.from != b.from {
			return a.from < b.from
		}
		if a.name != b.name {
			return a.name < b.name
		}
		return bytes.Compare(a.id[:], b.id[:]) < 0
	})
	return infos
}

func shareInfoTable(infos []shareInfo, isFullIDs bool) (ret string) {
	// Determine each column's max width.
	wID := idPrefixLen
	if isFullIDs && len(infos) > 0 {
		wID = len(infos[0].id)
	}
	wFrom := 0
	wName := 0
	wMode := 0
	for i := range infos {
		if n := len(infos[i].mode); n > wMode {
			wMode = n
		}
		if n := len(infos[i].from); n > wFrom {
			wFrom = n
		}
		if n := len(infos[i].name); n > wName {
			wName = n
		}
	}
	// Print the table column headers.
	ret += fmt.Sprintf(" %-*s | %-*s | %-*s | file\n", wID, "id", wMode, "mode", wFrom, "from")
	ret += fmt.Sprintf("-%s-+-%s-+-%s-+-%s-\n",
		strings.Repeat("-", wID),
		strings.Repeat("-", wMode),
		strings.Repeat("-", wFrom),
		strings.Repeat("-", wName),
	)
	// Print the table rows of pending share infos.
	for i, info := range infos {
		ret += fmt.Sprintf(
			" %-*s | %-*s | %-*s | %s",
			wID, info.id[:wID],
			wMode, info.mode,
			wFrom, info.from,
			info.name,
		)
		if i != len(infos)-1 {
			ret += "\n"
		}
	}
	return
}
