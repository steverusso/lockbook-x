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
	// the target file (defaults to "/jots.md")
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
		// Create a doc named "jots.md" in root if it doesn't exist.
		f, ok, err := maybeFileByPath(core, "/jots.md")
		if err != nil {
			return fmt.Errorf("getting jot file by path: %w", err)
		}
		if !ok {
			f, err = core.CreateFileAtPath("/jots.md")
			if err != nil {
				return fmt.Errorf("creating '/jots.md': %w", err)
			}
			fmt.Println("created a new '/jots.md' file!")
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
	jotsContent, err := core.ReadDocument(targetID)
	if err != nil {
		return fmt.Errorf("reading jot file: %w", err)
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

	if j.dateIt {
		j.message = "(" + time.Now().Format("Mon, 2 Jan 2006 15:04") + ") " + j.message
	}

	// Prepend two new lines if the last chars aren't new lines already.
	n := len(jotsContent)
	if n > 0 && jotsContent[n-1] != '\n' {
		j.message = "\n" + j.message
	}
	if n > 1 && jotsContent[n-2] != '\n' {
		j.message = "\n" + j.message
	}

	j.message += "\n"
	jotsContent = append(jotsContent, j.message...)

	// Write the new content back.
	if err = core.WriteDocument(targetID, jotsContent); err != nil {
		return fmt.Errorf("writing new jots content: %w", err)
	}
	return nil
}
