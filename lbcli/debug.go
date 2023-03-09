package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/steverusso/lockbook-x/go-lockbook"
)

// Investigative commands mainly intended for devs.
type debugCmd struct {
	finfo    *debugFinfoCmd
	validate *debugValidateCmd
	whoami   *debugWhoamiCmd
}

func (d *debugCmd) run(core lockbook.Core) error {
	switch {
	case d.finfo != nil:
		return d.finfo.run(core)
	case d.validate != nil:
		return d.validate.run(core)
	case d.whoami != nil:
		return d.whoami.run(core)
	default:
		return nil
	}
}

// View info about a target file.
type debugFinfoCmd struct {
	// The target can be a file path, UUID, or UUID prefix.
	//
	// clap:arg_required
	target string
}

func (c *debugFinfoCmd) run(core lockbook.Core) error {
	id, err := idFromSomething(core, c.target)
	if err != nil {
		return fmt.Errorf("trying to get id from %q: %w", c.target, err)
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

func printFile(f lockbook.File, myName string) {
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
		{"id", f.ID.String()},
		{"parent", f.Parent.String()},
		{"type", strings.ToLower(lockbook.FileTypeString(f.Type))},
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

// Find invalid states within your lockbook.
type debugValidateCmd struct{}

func (debugValidateCmd) run(core lockbook.Core) error {
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

// Print user information for this lockbook.
type debugWhoamiCmd struct{}

func (debugWhoamiCmd) run(core lockbook.Core) error {
	acct, err := core.GetAccount()
	if err != nil {
		return fmt.Errorf("getting account: %w", err)
	}
	fmt.Printf("data-dir: %s\n", core.WriteablePath())
	fmt.Printf("username: %s\n", acct.Username)
	fmt.Printf("server:   %s\n", acct.APIURL)
	return nil
}
