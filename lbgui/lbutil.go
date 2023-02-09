package main

import (
	"fmt"
	"strings"

	"github.com/steverusso/lockbook-x/go-lockbook"
)

// Gets all parents except root in descending order from root.
func getParents(core lockbook.Core, id lockbook.FileID) ([]nameAndID, error) {
	r := []nameAndID{}
	for {
		f, err := core.FileByID(id)
		if err != nil {
			return nil, fmt.Errorf("file by id %q: %w", id, err)
		}
		if f.ID == f.Parent {
			break
		}
		id = f.Parent
		r = append([]nameAndID{{
			name: strings.Clone(f.Name),
			id:   f.ID,
		}}, r...)
	}
	return r, nil
}
