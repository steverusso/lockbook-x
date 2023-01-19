package main

import (
	"fmt"
	"os"
	"strings"

	lb "github.com/steverusso/lockbook-x/go-lockbook"
)

func debugFinfo(core lb.Core, target string) error {
	id, err := idFromSomething(core, target)
	if err != nil {
		return fmt.Errorf("trying to get id from %q: %w", target, err)
	}
	f, err := core.FileByID(id)
	if err != nil {
		return fmt.Errorf("getting file %q: %w", id, err)
	}
	acct, err := core.GetAccount()
	if err != nil {
		return fmt.Errorf("getting account: %w", err)
	}
	printFile(f, acct.Username)
	return nil
}

func printFile(f lb.File, myName string) {
	// Build the text that will contain share info.
	shares := ""
	for _, sh := range f.Shares {
		sharedBy := "@" + sh.SharedBy
		if sh.SharedBy == myName {
			sharedBy = "me"
		}
		sharedWith := "@" + sh.SharedWith
		if sh.SharedWith == myName {
			sharedWith = "me"
		}
		mode := strings.ToLower(sh.Mode.String())
		shares += fmt.Sprintf("\n    %s -> %s (%s)", sharedBy, sharedWith, mode)
	}
	data := [][2]string{
		{"name", f.Name},
		{"id", f.ID},
		{"parent", f.Parent},
		{"type", strings.ToLower(lb.FileTypeString(f.Type))},
		{"lastmod", fmt.Sprintf("%v", f.Lastmod)},
		{"lastmod_by", f.LastmodBy},
		{fmt.Sprintf("shares (%d)", len(f.Shares)), shares},
	}
	// Determine widest key name.
	nameWidth := 0
	for i := range data {
		n := len(data[i][0])
		if n > nameWidth {
			nameWidth = n
		}
	}
	for i := range data {
		fmt.Printf("  %-*s : %s\n", nameWidth, data[i][0], data[i][1])
	}
}

func debugValidate(core lb.Core) error {
	warnings, err := core.Validate()
	if err != nil {
		return fmt.Errorf("running validate: %w", err)
	}
	count := len(warnings)
	if count == 0 {
		return nil
	}
	suffix := ""
	if count > 1 {
		suffix = "s"
	}
	fmt.Fprintf(os.Stderr, "\033[1;33m%d warning%s found:\033[0m\n", count, suffix)
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "  %s\n", w)
	}
	return nil
}
