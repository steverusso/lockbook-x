package main

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"gioui.org/gesture"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/gofrs/uuid"
	"github.com/steverusso/lockbook-x/go-lockbook"
)

const (
	inset     = 8
	insetHalf = inset / 2
)

type fileExplorer struct {
	targetID    lockbook.FileID
	homeBtn     widget.Clickable
	mkdirBtn    widget.Clickable
	mkdocBtn    widget.Clickable
	bcrumbs     []breadcrumb
	entries     []fileEntry
	entryList   widget.List
	lastClicked int

	maxWidth        int
	textSize        unit.Sp
	scrollBarWidth  int
	nameColWidth    int
	lastmodColWidth int
}

type nameAndID struct {
	name string
	id   lockbook.FileID
}

type breadcrumb struct {
	btn  widget.Clickable
	id   lockbook.FileID
	name string
}

func (ws *workspace) layFileExplorer(gtx C, th *material.Theme) D {
	ex := &ws.expl
	gtx.Constraints.Min = image.Point{}
	listStyle := material.List(th, &ex.entryList)

	// Ensure the cached spacing-related values are accurate.
	if ex.maxWidth != gtx.Constraints.Max.X || ex.textSize != th.TextSize || ex.scrollBarWidth != int(listStyle.Width()) {
		ex.maxWidth = gtx.Constraints.Max.X
		ex.textSize = th.TextSize
		ex.scrollBarWidth = int(listStyle.Width())
		ex.lastmodColWidth = 0
		for i := range ws.expl.entries {
			m := op.Record(gtx.Ops)
			dims := material.Body2(th, ws.expl.entries[i].lastmod).Layout(gtx)
			_ = m.Stop()
			if dims.Size.X > ex.lastmodColWidth {
				ex.lastmodColWidth = dims.Size.X
			}
		}
		ex.nameColWidth = ex.maxWidth - ex.lastmodColWidth - ex.scrollBarWidth
	}

	topBarDims := ex.layTopbar(gtx, th)
	yOffsetInc := topBarDims.Size.Y
	yOffsetInc += offsetAndLayHR(gtx, th, yOffsetInc)
	// entry column headers
	{
		gtx.Constraints.Max.X -= inset * 2
		offOp := op.Offset(image.Pt(inset, yOffsetInc+inset)).Push(gtx.Ops)
		height := 0
		{
			gtx1 := gtx
			gtx1.Constraints.Max.X = ex.nameColWidth - ex.scrollBarWidth
			dims := explHeaderLbl(th, "Name").Layout(gtx1)
			height = dims.Size.Y
		}
		{
			gtx2 := gtx
			gtx2.Constraints.Min.X = ex.lastmodColWidth
			gtx2.Constraints.Max.X = ex.lastmodColWidth
			op.Offset(image.Pt(ex.nameColWidth-ex.scrollBarWidth-inset, 0)).Add(gtx2.Ops)
			lbl := explHeaderLbl(th, "Last Modified")
			lbl.Alignment = text.End
			_ = lbl.Layout(gtx2)
		}
		offOp.Pop()
		yOffsetInc += height + inset*2
		gtx.Constraints.Max.X = ex.maxWidth // restore full max width
	}
	yOffsetInc += offsetAndLayHR(gtx, th, yOffsetInc)
	// file entry list
	{
		remainingHeight := gtx.Constraints.Max.Y - yOffsetInc
		gtx.Constraints.Max.Y = remainingHeight
		gtx.Constraints.Min.Y = remainingHeight

		offOp := op.Offset(image.Pt(0, yOffsetInc)).Push(gtx.Ops)
		if len(ex.entries) == 0 {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			_ = layout.Center.Layout(gtx, material.Body1(th, "No files here!").Layout)
		} else {
			_ = listStyle.Layout(gtx, len(ex.entries), func(gtx C, i int) D {
				return ws.layFileEntry(gtx, th, i)
			})
		}
		offOp.Pop()
	}
	return D{Size: gtx.Constraints.Max}
}

func (ex *fileExplorer) layTopbar(gtx C, th *material.Theme) D {
	// todo(steve): groupButton is widget state and should be taken out of frame
	bcrumbBtns := make([]groupButton, len(ex.bcrumbs)+1)
	bcrumbBtns[0] = groupButton{
		click: &ex.homeBtn,
		icon:  &iconHome,
	}
	for i, bc := range ex.bcrumbs {
		bcrumbBtns[i+1] = groupButton{
			click: &ex.bcrumbs[i].btn,
			text:  bc.name,
		}
	}

	btnGrpStyle := buttonGroupStyle{
		bg:       th.Bg,
		fg:       th.Fg,
		shaper:   th.Shaper,
		textSize: th.TextSize * 0.9,
	}

	m := op.Record(gtx.Ops)
	gtx.Constraints.Min = image.Point{}
	endBtnsDims := btnGrpStyle.layout(gtx, []groupButton{
		{click: &ex.mkdocBtn, icon: &iconNewDoc},
		{click: &ex.mkdirBtn, icon: &iconNewFolder},
	})
	drawEndBtns := m.Stop()

	gtx1 := gtx
	gtx1.Constraints.Max.X -= endBtnsDims.Size.X + (inset * 2)
	offOp1 := op.Offset(image.Pt(inset, inset)).Push(gtx.Ops)
	_ = btnGrpStyle.layout(gtx1, bcrumbBtns)
	offOp1.Pop()

	offOp2 := op.Offset(image.Pt(gtx1.Constraints.Max.X+inset, inset)).Push(gtx.Ops)
	drawEndBtns.Add(gtx.Ops)
	offOp2.Pop()
	return D{Size: image.Pt(gtx.Constraints.Max.X, endBtnsDims.Size.Y+(inset*2))}
}

func (ws *workspace) layFileEntry(gtx C, th *material.Theme, i int) D {
	ex := &ws.expl
	en := &ex.entries[i]
	for _, e := range en.click.Events(gtx) {
		if e.Type != gesture.TypePress {
			continue
		}
		switch e.NumClicks {
		case 2:
			if en.isDir() {
				ws.openDir(en.id)
			} else {
				ws.openFiles([]nameAndID{{name: en.name, id: en.id}})
			}
			ex.makeSelection(i, false, false)
		case 1:
			isCtrl := e.Modifiers.Contain(key.ModCtrl)
			isShift := e.Modifiers.Contain(key.ModShift)
			ex.makeSelection(i, isCtrl, isShift)
		}
		op.InvalidateOp{}.Add(gtx.Ops)
	}

	macro := op.Record(gtx.Ops)
	entryDims := ex.drawFileEntry(gtx, th, en)
	call := macro.Stop()

	bg := color.NRGBA{}
	switch {
	case en.isSelected():
		bg = lighten(th.ContrastFg, 0.05)
	case en.click.Hovered():
		bg = darken(th.ContrastFg, 0.1)
	}

	rrOp := clip.UniformRRect(image.Rectangle{Max: entryDims.Size}, 6).Push(gtx.Ops)
	paint.FillShape(gtx.Ops, bg, clip.Rect{Max: entryDims.Size}.Op())
	en.click.Add(gtx.Ops)
	call.Add(gtx.Ops)
	rrOp.Pop()
	return entryDims
}

func (ex *fileExplorer) drawFileEntry(gtx C, th *material.Theme, en *fileEntry) D {
	fullWidth := gtx.Constraints.Max.X
	gtx.Constraints.Max.X -= inset * 2

	topInsetOp := op.Offset(image.Pt(inset, insetHalf)).Push(gtx.Ops)
	height := 0
	{
		// icon
		iconDims := en.icon.Layout(gtx, th.Fg)
		height = iconDims.Size.Y
		// name label
		nameColWidth := ex.nameColWidth - iconDims.Size.X - insetHalf
		gtx1 := gtx
		gtx1.Constraints.Min.X = nameColWidth
		gtx1.Constraints.Max.X = nameColWidth
		offOp := op.Offset(image.Pt(iconDims.Size.X+insetHalf, 0)).Push(gtx1.Ops)
		vertCenter(gtx1, height, material.Body2(th, en.name).Layout)
		offOp.Pop()
	}
	{
		// last modified label
		lastmod := material.Body2(th, en.lastmod)
		lastmod.Alignment = text.End
		if !en.isSelected() && !en.click.Hovered() {
			lastmod.Color.A /= 4
		}
		gtx2 := gtx
		gtx2.Constraints.Min.X = ex.lastmodColWidth
		gtx2.Constraints.Max.X = ex.lastmodColWidth
		offOp := op.Offset(image.Pt(ex.nameColWidth-inset*2, 0)).Push(gtx2.Ops)
		vertCenter(gtx2, height, lastmod.Layout)
		offOp.Pop()
	}
	topInsetOp.Pop()
	return D{Size: image.Pt(fullWidth, height+inset)}
}

func (ws *workspace) openDir(id lockbook.FileID) {
	ws.expl.targetID = id
	go func() {
		u := openDirResult{id: id}
		defer func() { ws.updates <- u }()

		if id == uuid.Nil {
			f, err := ws.core.GetRoot()
			if err != nil {
				u.err = fmt.Errorf("getting root: %w", err)
				return
			}
			id = f.ID
		}
		parents, err := getParents(ws.core, id)
		if err != nil {
			u.err = fmt.Errorf("getting parents of %q: %w", id, err)
			return
		}
		u.parents = parents

		files, err := ws.core.GetChildren(id)
		if err != nil {
			u.err = fmt.Errorf("getting children of %q: %w", id, err)
			return
		}
		lockbook.SortFiles(files)
		u.files = files
	}()
}

func (ws *workspace) openFiles(namesAndIDs []nameAndID) {
	ws.animStage = wsExplClosing
	for _, v := range namesAndIDs {
		v := v
		ws.insertTab(v.id, v.name)
		go openFile(ws.core, ws.updates, v.id)
	}
}

func openFile(core lockbook.Core, updates chan<- legitUpdate, id lockbook.FileID) {
	u := openFileResult{id: id}

	u.data, u.err = core.ReadDocument(id)
	if u.err != nil {
		u.err = fmt.Errorf("reading doc %q: %w", id, u.err)
	}
	updates <- u
}

func (ws *workspace) deleteSelectedFiles() {
	// todo(steve): impl
}

func (ex *fileExplorer) makeSelection(i int, isCtrl, isShift bool) {
	en := &ex.entries[i]
	switch {
	case isCtrl:
		en.flags ^= entryIsSelected
	case isShift:
		if i > ex.lastClicked {
			for z := ex.lastClicked + 1; z <= i; z++ {
				ex.entries[z].flags ^= entryIsSelected
			}
		}
		if i < ex.lastClicked {
			for z := ex.lastClicked - 1; z >= i; z-- {
				ex.entries[z].flags ^= entryIsSelected
			}
		}
	default:
		ex.deselectAll()
		en.flags |= entryIsSelected
	}
	ex.lastClicked = i
}

func (ex *fileExplorer) selectAll() {
	for i := range ex.entries {
		ex.entries[i].flags |= entryIsSelected
	}
}

func (ex *fileExplorer) deselectAll() {
	for i := range ex.entries {
		en := &ex.entries[i]
		en.flags &^= entryIsSelected
	}
}

func (ex *fileExplorer) add(f *lockbook.File) {
	en := newFileEntry(f)
	for i := range ex.entries {
		if en.isLessThan(&ex.entries[i]) {
			ex.entries = append(ex.entries[:i], append([]fileEntry{en}, ex.entries[i:]...)...)
			return
		}
	}
	ex.entries = append(ex.entries, en)
}

func (ex *fileExplorer) removeEntries(del []lockbook.File) {
	for _, d := range del {
		for i := range ex.entries {
			en := &ex.entries[i]
			if en.name == d.Name {
				ex.entries = append(ex.entries[:i], ex.entries[i+1:]...)
				break
			}
		}
	}
}

func (ex *fileExplorer) populate(parents []nameAndID, files []lockbook.File) {
	if cap(ex.bcrumbs) < len(parents) {
		ex.bcrumbs = make([]breadcrumb, 0, len(parents))
	} else {
		ex.bcrumbs = ex.bcrumbs[:0]
	}
	for i := range parents {
		ex.bcrumbs = append(ex.bcrumbs, breadcrumb{
			name: strings.Clone(parents[i].name),
			id:   parents[i].id,
		})
	}

	if cap(ex.entries) < len(files) {
		ex.entries = make([]fileEntry, 0, len(files))
	} else {
		ex.entries = ex.entries[:0]
	}
	for i := range files {
		ex.entries = append(ex.entries, newFileEntry(&files[i]))
	}
}

const (
	entryIsDir uint8 = 1 << iota
	entryIsSelected
)

type fileEntry struct {
	id      lockbook.FileID
	name    string
	lastmod string
	flags   uint8
	click   gesture.Click
	icon    *widget.Icon
}

func newFileEntry(f *lockbook.File) fileEntry {
	var flags uint8
	icon := &iconRegFile
	if f.IsDir() {
		flags |= entryIsDir
		icon = &iconDirectory
	}
	return fileEntry{
		id:      f.ID,
		name:    strings.Clone(f.Name),
		lastmod: f.Lastmod.Format("2 Jan 2006 15:04"),
		flags:   flags,
		icon:    icon,
	}
}

func (en *fileEntry) isDir() bool {
	return en.flags&entryIsDir == entryIsDir
}

func (en *fileEntry) isSelected() bool {
	return en.flags&entryIsSelected == entryIsSelected
}

func (en *fileEntry) isLessThan(other *fileEntry) bool {
	if en.isDir() == other.isDir() {
		return en.name < other.name
	}
	return en.isDir()
}

func explHeaderLbl(th *material.Theme, txt string) material.LabelStyle {
	l := material.Body2(th, txt)
	l.Color.A /= 3
	l.Font.Weight = text.Bold
	return l
}
