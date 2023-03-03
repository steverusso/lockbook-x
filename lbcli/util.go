package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/steverusso/lockbook-x/go-lockbook"
)

func maybeFileByPath(core lockbook.Core, p string) (lockbook.File, bool, error) {
	f, err := core.FileByPath(p)
	if err != nil {
		if err, ok := err.(*lockbook.Error); ok && err.Code == lockbook.CodeFileNonexistent {
			return lockbook.File{}, false, nil
		}
		return lockbook.File{}, false, err
	}
	return f, true, nil
}

func idFromSomething(core lockbook.Core, v string) (uuid.UUID, error) {
	if id := uuid.FromStringOrNil(v); !id.IsNil() {
		return id, nil
	}
	f, err := core.FileByPath(v)
	if err == nil {
		return f.ID, nil
	}
	if err, ok := err.(*lockbook.Error); ok && err.Code != lockbook.CodeFileNonexistent {
		return uuid.Nil, fmt.Errorf("trying to get a file by path: %w", err)
	}
	// Not a full UUID and not a path, so that leaves UUID prefix.
	files, err := core.ListMetadatas()
	if err != nil {
		return uuid.Nil, fmt.Errorf("listing metadatas to check ids: %w", err)
	}
	possibs := make([]lockbook.File, 0, 5)
	for i := range files {
		if strings.HasPrefix(files[i].ID.String(), v) {
			possibs = append(possibs, files[i])
		}
	}
	n := len(possibs)
	if n == 0 {
		return uuid.Nil, fmt.Errorf("value %q is not a path, uuid, or uuid prefix", v)
	}
	if n == 1 {
		return possibs[0].ID, nil
	}
	// Multiple ID prefix matches.
	errMsg := fmt.Sprintf("value %q is not a path and matches %d file ID prefixes:\n", v, n)
	for _, f := range possibs {
		pathOrErr, err := core.PathByID(f.ID)
		if err != nil {
			pathOrErr = fmt.Sprintf("error getting path: %v", err)
		}
		errMsg += fmt.Sprintf("  %s  %s\n", f.ID, pathOrErr)
	}
	return uuid.Nil, errors.New(errMsg)
}
