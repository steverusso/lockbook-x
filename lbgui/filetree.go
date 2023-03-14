package main

import (
	"fmt"
	"image"
	"time"

	"gioui.org/gesture"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/gofrs/uuid"
	"github.com/steverusso/lockbook-x/go-lockbook"
)

const doubleClickDuration = 200 * time.Millisecond

type fileTree struct {
	root     treeEntry
	list     widget.List
	toSelect lockbook.FileID
}

type treeEntry struct {
	file       lockbook.File
	isLoading  bool
	isExpanded bool
	isSelected bool
	children   []treeEntry
	arrowClick gesture.Click
	click      gesture.Click
}

func (ws *workspace) layFileTree(gtx C, th *material.Theme) D {
	return material.List(th, &ws.tree.list).Layout(gtx, 1, func(gtx C, i int) D {
		return ws.layEntry(gtx, th, &ws.tree.root, 0)
	})
}

func (ws *workspace) layEntry(gtx C, th *material.Theme, en *treeEntry, lvl int) D {
	t := &ws.tree
	if !t.toSelect.IsNil() {
		if t.toSelect == en.file.ID {
			en.isSelected = true
			t.toSelect = uuid.Nil
		} else if en.isSelected {
			en.isSelected = false
		}
	}
	layIcon := func(gtx C) D {
		icon := iconDocument
		if en.file.IsDir() {
			icon = iconDirectory
		}
		return icon.Layout(gtx, th.Fg)
	}
	layName := material.Body2(th, en.file.Name).Layout
	layArrow := func(gtx C) D {
		arrow := iconArrowRight
		if en.isExpanded {
			arrow = iconArrowDown
		}
		arrMacro := op.Record(gtx.Ops)
		arrDims := arrow.Layout(gtx, th.Fg)
		arrCall := arrMacro.Stop()
		if !en.file.IsDir() {
			return arrDims
		}
		defer clip.Rect(image.Rectangle{Max: arrDims.Size}).Push(gtx.Ops).Pop()
		en.arrowClick.Add(gtx.Ops)
		arrCall.Add(gtx.Ops)
		return arrDims
	}
	m := op.Record(gtx.Ops)
	dims := layout.Inset{}.Layout(gtx, func(gtx C) D {
		flx := make([]layout.FlexChild, 0, len(en.children))
		flx = append(flx, layout.Rigid(func(gtx C) D {
			singleMacro := op.Record(gtx.Ops)
			singleDims := layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(layout.Spacer{Width: unit.Dp(16 * lvl)}.Layout),
				layout.Rigid(layArrow),
				layout.Rigid(layIcon),
				layout.Rigid(layout.Spacer{Width: 10}.Layout),
				layout.Flexed(1, layName),
			)
			singleCall := singleMacro.Stop()
			defer clip.Rect(image.Rectangle{Max: singleDims.Size}).Push(gtx.Ops).Pop()
			en.click.Add(gtx.Ops)

			if en.isSelected {
				rect := clip.Rect{Max: singleDims.Size}.Op()
				paint.FillShape(gtx.Ops, th.ContrastFg, rect)
			}
			singleCall.Add(gtx.Ops)
			return singleDims
		}))
		if en.isExpanded {
			for i := range en.children {
				chEntry := &en.children[i]
				flx = append(flx, layout.Rigid(func(gtx C) D {
					return ws.layEntry(gtx, th, chEntry, lvl+1)
				}))
			}
		}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, flx...)
	})
	call := m.Stop()

	for _, e := range en.click.Events(gtx) {
		if e.Type == gesture.TypePress {
			switch e.NumClicks {
			case 1:
				if e.Modifiers == key.ModCtrl {
					en.isSelected = !en.isSelected
				} else {
					t.toSelect = en.file.ID
				}
			case 2:
				if en.file.IsDir() {
					ws.openDirTree(en.file.ID)
				} else {
					ws.openFiles([]nameAndID{{name: en.file.Name, id: en.file.ID}})
				}
			}
		}
	}
	for _, e := range en.arrowClick.Events(gtx) {
		if e.Type == gesture.TypePress {
			if !en.isExpanded {
				if en.file.IsDir() {
					ws.openDirTree(en.file.ID)
				}
			} else {
				en.isExpanded = false
			}
		}
	}

	call.Add(gtx.Ops)
	return dims
}

func (t *fileTree) find(id lockbook.FileID) *treeEntry {
	return t.root.find(id)
}

func (t *fileTree) populate(id lockbook.FileID, files []lockbook.File) {
	en := t.find(id)
	if en == nil {
		return
	}
	en.setChildren(files)
	en.isExpanded = true
}

func newFileTreeEntry(f lockbook.File, files []lockbook.File) treeEntry {
	en := treeEntry{file: f}
	en.setChildren(files)
	return en
}

func (en *treeEntry) setChildren(files []lockbook.File) {
	entries := make([]treeEntry, len(files))
	for i, f := range files {
		entries[i] = newFileTreeEntry(f, nil)
	}
	en.children = entries
}

func (en *treeEntry) find(id lockbook.FileID) *treeEntry {
	if en.file.ID == id {
		return en
	}
	for i := range en.children {
		if ee := en.children[i].find(id); ee != nil {
			return ee
		}
	}
	return nil
}

func (ws *workspace) openDirTree(id lockbook.FileID) {
	go func() {
		u := openDirTreeResult{id: id}
		defer func() { ws.updates <- u }()

		if id == uuid.Nil {
			f, err := ws.core.GetRoot()
			if err != nil {
				u.err = fmt.Errorf("getting root: %w", err)
				return
			}
			id = f.ID
		}

		files, err := ws.core.GetChildren(id)
		if err != nil {
			u.err = fmt.Errorf("getting children of %q: %w", id, err)
			return
		}
		lockbook.SortFiles(files)
		u.files = files
	}()
}
