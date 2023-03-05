package main

import (
	"fmt"
	"time"

	"github.com/steverusso/lockbook-x/go-lockbook"
)

// quickly record brief thoughts
type jotCmd struct {
	// prepend the date and time to the message
	//
	// clap:opt dateit,d
	dateIt bool
	// append the date and time to the message
	//
	// clap:opt dateit-after,D
	dateItAfter bool
	// the target file (defaults to "/scratch.md")
	//
	// clap:opt target,t
	target string
	// the text you would like to jot down
	//
	// clap:arg_required
	message string
}

func (j *jotCmd) run(core lockbook.Core) error {
	var targetID lockbook.FileID
	if j.target == "" {
		// Create a doc named "scratch.md" in root if it doesn't exist.
		f, ok, err := lockbook.MaybeFileByPath(core, "/scratch.md")
		if err != nil {
			return fmt.Errorf("getting scratch file by path: %w", err)
		}
		if !ok {
			f, err = core.CreateFileAtPath("/scratch.md")
			if err != nil {
				return fmt.Errorf("creating '/scratch.md': %w", err)
			}
			fmt.Println("created a new '/scratch.md' file!")
		}
		targetID = f.ID
	} else {
		id, err := idFromSomething(core, j.target)
		if err != nil {
			return fmt.Errorf("trying to get id from '%s': %w", j.target, err)
		}
		targetID = id
	}

	// Read the target doc's current content.
	scratchContent, err := core.ReadDocument(targetID)
	if err != nil {
		return fmt.Errorf("reading scratch file: %w", err)
	}

	if j.dateIt || j.dateItAfter {
		dateTime := time.Now().Format("Mon, 2 Jan 2006 15:04")
		switch {
		case j.dateIt:
			j.message = "(" + dateTime + ") " + j.message
		case j.dateItAfter:
			j.message += " (" + dateTime + ")"
		}
	}

	// Prepend two new lines if the last chars aren't new lines already.
	n := len(scratchContent)
	if n > 0 && scratchContent[n-1] != '\n' {
		j.message = "\n" + j.message
	}
	if n > 1 && scratchContent[n-2] != '\n' {
		j.message = "\n" + j.message
	}

	j.message += "\n"
	scratchContent = append(scratchContent, j.message...)

	// Write the new content back.
	if err = core.WriteDocument(targetID, scratchContent); err != nil {
		return fmt.Errorf("writing new scratch content: %w", err)
	}
	return nil
}
