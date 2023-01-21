package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	lb "github.com/steverusso/lockbook-x/go-lockbook"
)

func cat(core lb.Core, targets []string) error {
	for _, target := range targets {
		id, err := idFromSomething(core, target)
		if err != nil {
			return fmt.Errorf("trying to get id from %q: %w", target, err)
		}
		data, err := core.ReadDocument(id)
		if err != nil {
			return fmt.Errorf("reading doc %q: %w", target, err)
		}
		fmt.Printf("%s", data)
	}
	return nil
}

func mk(core lb.Core, fpath string) error {
	_, err := core.CreateFileAtPath(fpath)
	if err != nil {
		return fmt.Errorf("creating file at path %q: %w", fpath, err)
	}
	return nil
}

func rename(core lb.Core, target, newName string, isForce bool) error {
	id, err := idFromSomething(core, target)
	if err != nil {
		return fmt.Errorf("trying to get id from %q: %w", target, err)
	}
	// If we don't have a name, loop until we get one.
	for {
		if newName != "" {
			break
		}
		fmt.Print("choose a new name: ")
		fmt.Scanln(&newName)
	}
	// Loop until we find an available file name. If this is forced, it'll just return the
	// first error.
	for {
		err := core.RenameFile(id, newName)
		// If this is a forced rename, we're not going to inspect the error and guide the
		// user on finding an available file name.
		if err == nil || isForce {
			return err
		}
		if err, ok := err.(*lb.Error); !ok && err.Code != lb.CodeFileNameUnavailable {
			return err
		}
		// Loop until we get non-empty new input.
		prompt := fmt.Sprintf("the name %q is taken, please try another: ", newName)
		newName = ""
		for {
			fmt.Print(prompt)
			fmt.Scanln(&newName)
			if newName != "" {
				break
			}
			prompt = "choose a new name: "
		}
	}
}

func moveFile(core lb.Core, srcTarget, destTarget string) error {
	srcID, err := idFromSomething(core, srcTarget)
	if err != nil {
		return fmt.Errorf("trying to get src id from %q: %w", srcTarget, err)
	}
	destID, err := idFromSomething(core, destTarget)
	if err != nil {
		return fmt.Errorf("trying to get dest id from %q: %w", destTarget, err)
	}
	err = core.MoveFile(srcID, destID)
	if err != nil {
		return fmt.Errorf("moving %s -> %s: %w", srcID, destID, err)
	}
	return nil
}

func deleteFiles(core lb.Core, targets []string, isForce bool) error {
	ids := make([]lb.FileID, len(targets))
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
		if !isForce {
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

func writeDoc(core lb.Core, target string, trunc bool) error {
	fi, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	if fi.Mode()&os.ModeNamedPipe == 0 {
		return errors.New("to write data to a lockbook document, pipe it into this command, e.g.:\necho 'hi' | lockbook write my-doc.txt")
	}
	id, err := idFromSomething(core, target)
	if err != nil {
		return fmt.Errorf("trying to get an id from %q: %w", target, err)
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("trying to read from stdin: %w", err)
	}
	if !trunc {
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
