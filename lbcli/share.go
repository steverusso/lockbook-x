package main

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/gofrs/uuid"
	lb "github.com/steverusso/lockbook-x/go-lockbook"
)

func createShare(core lb.Core, target, toWho string, readOnly bool) error {
	id, err := idFromSomething(core, target)
	if err != nil {
		return fmt.Errorf("trying to get id from %q: %w", target, err)
	}
	mode := lb.ShareModeWrite
	if readOnly {
		mode = lb.ShareModeRead
	}
	if err = core.ShareFile(id, toWho, mode); err != nil {
		return fmt.Errorf("sharing file: %w", err)
	}
	fmt.Printf("\033[1;32mdone!\033[0m file %q will be shared next time you sync.\n", id)
	return nil
}

func listPendingShares(core lb.Core, isFullIDs bool) error {
	pendingShares, err := core.GetPendingShares()
	if err != nil {
		return fmt.Errorf("getting pending shares: %w", err)
	}
	if len(pendingShares) == 0 {
		fmt.Println("no pending shares")
		return nil
	}
	shareInfos := filesToShareInfos(pendingShares)
	fmt.Println(shareInfoTable(shareInfos, isFullIDs))
	return nil
}

func acceptShare(core lb.Core, id, dest, newName string) error {
	pendingShares, err := core.GetPendingShares()
	if err != nil {
		return fmt.Errorf("getting pending shares: %w", err)
	}
	if len(pendingShares) == 0 {
		return errors.New("no pending shares to accept")
	}
	match, err := getOnePendingShareMatch(pendingShares, id)
	if err != nil {
		return err
	}
	parentID := lb.FileID{}
	// If the destination is a valid UUID, it must be of an existing directory.
	if destID, err := uuid.FromString(dest); err == nil {
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
		f, exists, err := maybeFileByPath(core, dest)
		if err != nil {
			return fmt.Errorf("file by path %q: %w", dest, err)
		}
		if !exists {
			// If the destination path doesn't exist, then it's just treated as a
			// non-existent directory path. The user can choose a name with the `--rename`
			// flag.
			fPath := dest
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
				return fmt.Errorf("existing destination path %q is a doc, must be a folder", dest)
			}
			parentID = f.ID
		}
	}

	name := newName
	if name == "" {
		name = match.Name
	}
	if name[len(name)-1] == '/' {
		name = name[:len(name)-1] // Prevent "name contains slash" error.
	}

	_, err = core.CreateFile(name, parentID, lb.FileTypeLink{Target: match.ID})
	if err != nil {
		return fmt.Errorf("creating link: %w", err)
	}
	return nil
}

func deletePendingShare(core lb.Core, id string) error {
	pendingShares, err := core.GetPendingShares()
	if err != nil {
		return fmt.Errorf("getting pending shares: %w", err)
	}
	if len(pendingShares) == 0 {
		return errors.New("no pending shares to delete")
	}
	match, err := getOnePendingShareMatch(pendingShares, id)
	if err != nil {
		return err
	}
	if err = core.DeletePendingShare(match.ID); err != nil {
		return fmt.Errorf("delete pending share: %w", err)
	}
	return nil
}

func getOnePendingShareMatch(shares []lb.File, id string) (lb.File, error) {
	matches := []lb.File{}
	for _, f := range shares {
		if strings.HasPrefix(f.ID.String(), id) {
			matches = append(matches, f)
		}
	}
	n := len(matches)
	if len(matches) == 0 {
		desc := ""
		if !lb.IsUUID(id) {
			desc = " prefix"
		}
		return lb.File{}, fmt.Errorf("no pending share found with id%s %s", desc, id)
	}
	if n > 1 {
		return lb.File{}, fmt.Errorf(
			"id prefix %q matched %d pending shares:\n%s",
			id, n, shareInfoTable(filesToShareInfos(matches), true),
		)
	}
	return matches[0], nil
}

type shareInfo struct {
	id   lb.FileID
	from string
	name string
	mode string
}

func filesToShareInfos(files []lb.File) []shareInfo {
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
