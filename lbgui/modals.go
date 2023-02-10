package main

import (
	"gioui.org/widget"
	"github.com/steverusso/lockbook-x/go-lockbook"
)

type modal interface {
	implsModal()
}

type createFilePrompt struct {
	input widget.Editor
	typ   lockbook.FileType
	err   error
}

func newCreateFilePrompt(typ lockbook.FileType) *createFilePrompt {
	return &createFilePrompt{
		input: widget.Editor{SingleLine: true, Submit: true},
		typ:   typ,
	}
}

func (createFilePrompt) implsModal() {}
