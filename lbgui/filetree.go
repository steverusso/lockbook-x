package main

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"gioui.org/gesture"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
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
	popup    treePopupMenu
	toSelect lockbook.FileID
}

type treeEntry struct {
	file       lockbook.File
	isLoading  bool
	isExpanded bool
	isSelected bool
	children   []treeEntry
	arrowClick gesture.Click

	lastClickAt time.Duration
	// Contiguous means clicks within the "double click duration" from each other.
	contigClicks int
}

func (ws *workspace) layFileTree(gtx C, th *material.Theme) D {
	for _, e := range gtx.Events(&ws.tree) {
		e, ok := e.(pointer.Event)
		if !ok || e.Buttons != pointer.ButtonSecondary {
			continue
		}
		ws.tree.popup = treePopupMenu{
			state:    popupStateOpenNext,
			position: image.Pt(int(e.Position.X), int(e.Position.Y)),
		}
	}

	m := op.Record(gtx.Ops)
	dims := material.List(th, &ws.tree.list).Layout(gtx, len(ws.tree.root.children), func(gtx C, i int) D {
		return ws.layEntry(gtx, th, &ws.tree.root.children[i], 0, 0)
	})
	call := m.Stop()

	defer clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops).Pop()
	pointer.InputOp{Tag: &ws.tree, Types: pointer.Press}.Add(gtx.Ops)
	call.Add(gtx.Ops)
	return dims
}

func (ws *workspace) layEntry(gtx C, th *material.Theme, en *treeEntry, lvl, yOffset int) D {
	t := &ws.tree
	if !t.toSelect.IsNil() {
		if t.toSelect == en.file.ID {
			en.isSelected = true
			t.toSelect = uuid.Nil
		} else if en.isSelected {
			en.isSelected = false
		}
	}
	for _, e := range gtx.Events(en) {
		e, ok := e.(pointer.Event)
		if !ok || e.Type != pointer.Press {
			continue
		}

		if e.Time-en.lastClickAt < doubleClickDuration {
			en.contigClicks++
		} else {
			en.contigClicks = 1
		}
		en.lastClickAt = e.Time

		if e.Buttons == pointer.ButtonSecondary {
			if !en.isSelected {
				t.toSelect = en.file.ID
			}
			// fmt.Println(int(e.Position.Y) + yOffset)
			continue
		}

		switch en.contigClicks {
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

	height := 0
	m := op.Record(gtx.Ops)
	dims := drawEntry(gtx, th, en, lvl)
	call := m.Stop()

	rrOp := clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops)
	pointer.InputOp{
		Tag:   en,
		Types: pointer.Press | pointer.Release | pointer.Enter | pointer.Leave,
	}.Add(gtx.Ops)

	if en.isSelected {
		rect := clip.Rect{Max: dims.Size}.Op()
		paint.FillShape(gtx.Ops, th.ContrastFg, rect)
	}

	call.Add(gtx.Ops)
	rrOp.Pop()

	height += dims.Size.Y
	if en.isExpanded {
		for i := range en.children {
			chOff := op.Offset(image.Pt(0, height)).Push(gtx.Ops)
			chDims := ws.layEntry(gtx, th, &en.children[i], lvl+1, yOffset+height)
			chOff.Pop()
			height += chDims.Size.Y
		}
	}

	return D{Size: image.Pt(gtx.Constraints.Max.X, height)}
}

func drawEntry(gtx C, th *material.Theme, en *treeEntry, lvl int) D {
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
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(layout.Spacer{Width: unit.Dp(16 * lvl)}.Layout),
		layout.Rigid(layArrow),
		layout.Rigid(layIcon),
		layout.Rigid(layout.Spacer{Width: 10}.Layout),
		layout.Flexed(1, layName),
	)
}

func (t *fileTree) populate(id lockbook.FileID, files []lockbook.File) {
	en := t.find(id)
	if en == nil {
		return
	}
	en.setChildren(files)
	en.isExpanded = true
}

func (t *fileTree) find(id lockbook.FileID) *treeEntry {
	return t.root.find(id)
}

type treeSelection struct {
	entries []*treeEntry
}

func (t *fileTree) selection() treeSelection {
	return t.root.selection()
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

func (en *treeEntry) selection() treeSelection {
	var sel treeSelection
	if en.isSelected {
		sel.entries = append(sel.entries, en)
	}
	for i := range en.children {
		chSel := en.children[i].selection()
		sel.entries = append(sel.entries, chSel.entries...)
	}
	return sel
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

type popupState uint8

const (
	popupStateClosed popupState = iota
	popupStateOpenNext
	popupStateOpen
)

type treePopupMenu struct {
	state    popupState
	position image.Point

	newDoc popupMenuButton
	newDir popupMenuButton
}

func (ws *workspace) layoutTreePopup(gtx C, th *material.Theme) D {
	pm := &ws.tree.popup
	if pm.newDoc.Pressed() {
		*pm = treePopupMenu{}
		ws.modals = append(ws.modals, newCreateFilePrompt(lockbook.FileTypeDocument{}))
		return D{}
	}
	if pm.newDir.Pressed() {
		*pm = treePopupMenu{}
		ws.modals = append(ws.modals, newCreateFilePrompt(lockbook.FileTypeFolder{}))
		return D{}
	}

	width := 200
	height := 0

	m := op.Record(gtx.Ops)
	gtx.Constraints.Max.X = width
	// new doc
	{
		dims := layPopupMenuItem(gtx, th, &pm.newDoc, "New Document")
		height += dims.Size.Y
	}
	// new folder
	{
		offOp := op.Offset(image.Pt(0, height)).Push(gtx.Ops)
		dims := layPopupMenuItem(gtx, th, &pm.newDir, "New Folder")
		height += dims.Size.Y
		offOp.Pop()
	}
	// todo:
	// rename
	// export
	// delete
	call := m.Stop()

	offOp := op.Offset(pm.position).Push(gtx.Ops)
	rrOp := clip.Rect(image.Rectangle{Max: image.Pt(width, height)}).Push(gtx.Ops)

	paint.ColorOp{Color: color.NRGBA{255, 0, 0, 255}}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	call.Add(gtx.Ops)

	rrOp.Pop()
	offOp.Pop()
	return D{}
}
