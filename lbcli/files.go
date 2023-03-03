package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/steverusso/lockbook-x/go-lockbook"
)

// print one or more documents to stdout
type catCmd struct {
	// lockbook file path or id
	//
	// clap:arg_required
	target string
}

// create a directory or do nothing if it exists
type mkdirCmd struct {
	// a path at which to create the directory
	//
	// clap:arg_required
	path string
}

// create a document or do nothing if it exists
type mkdocCmd struct {
	// a path at which to create the document
	//
	// clap:arg_required
	path string
}

// move a file to another parent
type mvCmd struct {
	// the file to move
	//
	// clap:arg_required
	src string
	// the destination directory
	//
	// clap:arg_required
	dest string
}

// rename a file
//
// clap:cmd_usage [-f] <target> [new-name]
type renameCmd struct {
	// non-interactive (fail instead of prompting for corrections)
	//
	// clap:opt force,f
	force bool
	// the file to rename
	//
	// clap:arg_required
	target string
	// the desired new name
	//
	// clap:arg_required
	newName string
}

// delete a file
type rmCmd struct {
	// don't prompt for confirmation
	//
	// clap:opt force,f
	force bool
	// lockbook path or id to delete
	//
	// clap:arg_required
	target string
}

// write data from stdin to a lockbook document
//
// clap:cmd_usage [--trunc] <target>
type writeCmd struct {
	// truncate the file instead of appending to it
	//
	// clap:opt trunc
	trunc bool
	// lockbook path or id to write
	//
	// clap:arg_required
	target string
}

func (c *catCmd) run(core lockbook.Core) error {
	id, err := idFromSomething(core, c.target)
	if err != nil {
		return fmt.Errorf("trying to get id from %q: %w", c.target, err)
	}
	data, err := core.ReadDocument(id)
	if err != nil {
		return fmt.Errorf("reading doc %q: %w", c.target, err)
	}
	fmt.Printf("%s", data)
	return nil
}

func (c *mkdirCmd) run(core lockbook.Core) error {
	v := c.path
	if v != "/" && v[len(v)-1] != '/' {
		v += "/"
	}
	return mk(core, v)
}

func (c *mkdocCmd) run(core lockbook.Core) error {
	v := c.path
	if v != "/" && v[len(v)-1] == '/' {
		v = v[:len(v)-1]
	}
	return mk(core, v)
}

func mk(core lockbook.Core, fpath string) error {
	_, err := core.CreateFileAtPath(fpath)
	if err != nil {
		return fmt.Errorf("creating file at path %q: %w", fpath, err)
	}
	return nil
}

func (c *renameCmd) run(core lockbook.Core) error {
	if c.newName == "" && c.force {
		fmt.Fprintf(os.Stderr, "error: must provide new name if --force\n")
		os.Exit(1)
	}
	id, err := idFromSomething(core, c.target)
	if err != nil {
		return fmt.Errorf("trying to get id from %q: %w", c.target, err)
	}
	// If we don't have a name, loop until we get one.
	for {
		if c.newName != "" {
			break
		}
		fmt.Print("choose a new name: ")
		fmt.Scanln(&c.newName)
	}
	// Loop until we find an available file name. If this is forced, it'll just return the
	// first error.
	for {
		err := core.RenameFile(id, c.newName)
		// If this is a forced rename, we're not going to inspect the error and guide the
		// user on finding an available file name.
		if err == nil || c.force {
			return err
		}
		if err, ok := err.(*lockbook.Error); !ok && err.Code != lockbook.CodePathTaken {
			return err
		}
		// Loop until we get non-empty new input.
		prompt := fmt.Sprintf("the name %q is taken, please try another: ", c.newName)
		c.newName = ""
		for {
			fmt.Print(prompt)
			fmt.Scanln(&c.newName)
			if c.newName != "" {
				break
			}
			prompt = "choose a new name: "
		}
	}
}

func (c *mvCmd) run(core lockbook.Core) error {
	srcID, err := idFromSomething(core, c.src)
	if err != nil {
		return fmt.Errorf("trying to get src id from %q: %w", c.src, err)
	}
	destID, err := idFromSomething(core, c.dest)
	if err != nil {
		return fmt.Errorf("trying to get dest id from %q: %w", c.dest, err)
	}
	err = core.MoveFile(srcID, destID)
	if err != nil {
		return fmt.Errorf("moving %s -> %s: %w", srcID, destID, err)
	}
	return nil
}

func (c *rmCmd) run(core lockbook.Core) error {
	targets := []string{c.target} // todo(steve): support multiple targets in the command
	ids := make([]lockbook.FileID, len(targets))
	for i, t := range targets {
		id, err := idFromSomething(core, t)
		if err != nil {
			return fmt.Errorf("trying to get id from %q: %w", t, err)
		}
		ids[i] = id
	}
	for i, id := range ids {
		f, err := core.FileByID(id)
		if err != nil {
			return fmt.Errorf("file by id %q: %w", id, err)
		}
		if !c.force {
			phrase := fmt.Sprintf("delete %q", id)
			if t := targets[i]; t != id.String() {
				phrase += " (target: " + t + ")"
			}
			if f.IsDir() {
				children, err := core.GetAndGetChildrenRecursively(id)
				if err != nil {
					return fmt.Errorf("getting all children in order to count: %w", err)
				}
				phrase += fmt.Sprintf(" and its %d children", len(children))
			}
			answer := ""
			fmt.Printf("are you sure you want to %s? [y/N]: ", phrase)
			fmt.Scanln(&answer)
			if answer != "y" && answer != "Y" {
				fmt.Println("aborted.")
				continue
			}
		}
		err = core.DeleteFile(id)
		if err != nil {
			return fmt.Errorf("deleting file %q: %w", id, err)
		}
	}
	return nil
}

func (c *writeCmd) run(core lockbook.Core) error {
	fi, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	if fi.Mode()&os.ModeNamedPipe == 0 {
		return errors.New("to write data to a lockbook document, pipe it into this command, e.g.:\necho 'hi' | lockbook write my-doc.txt")
	}
	id, err := idFromSomething(core, c.target)
	if err != nil {
		return fmt.Errorf("trying to get an id from %q: %w", c.target, err)
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("trying to read from stdin: %w", err)
	}
	if !c.trunc {
		content, err := core.ReadDocument(id)
		if err != nil {
			return fmt.Errorf("reading doc %q: %w", id, err)
		}
		data = append(content, data...)
	}
	if err := core.WriteDocument(id, data); err != nil {
		return fmt.Errorf("writing to doc: %w", err)
	}
	return nil
}
