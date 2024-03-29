package main

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/steverusso/lockbook-x/go-lockbook"
	"github.com/steverusso/mdedit"
)

type tab struct {
	id             lockbook.FileID
	name           string
	btn            widget.Clickable
	view           mdedit.View
	isLoading      bool
	numQueuedSaves uint8
	lastEditAt     time.Time
	lastSaveAt     time.Time
}

func (t *tab) isDirty() bool {
	return t.lastSaveAt.Before(t.lastEditAt)
}

func (ws *workspace) layTabsNotebook(gtx C, th *material.Theme) D {
	if len(ws.tabs) == 0 {
		return layout.Center.Layout(gtx, func(gtx C) D {
			return ws.logo.Layout(gtx)
		})
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			if ws.tabList.Axis != layout.Horizontal {
				ws.tabList.Axis = layout.Horizontal
			}
			return ws.layTabList(gtx, th)
		}),
		layout.Rigid(rule{color: th.Fg}.Layout),
		layout.Flexed(1, func(gtx C) D {
			return ws.layMarkdownTab(gtx, th, &ws.tabs[ws.activeTab])
		}),
	)
}

func (ws *workspace) layTabs(gtx C, th *material.Theme) D {
	if len(ws.tabs) == 0 {
		return layout.Center.Layout(gtx, func(gtx C) D {
			return material.Body1(th, "Use Ctrl-O to open a file!").Layout(gtx)
		})
	}
	return layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			if ws.tabList.Axis != layout.Vertical {
				ws.tabList.Axis = layout.Vertical
			}
			gtx.Constraints.Min.X = 220
			gtx.Constraints.Max.X = 220
			return ws.layTabList(gtx, th)
		}),
		layout.Rigid(func(gtx C) D {
			size := image.Point{1, gtx.Constraints.Max.Y}
			rect := clip.Rect{Max: size}.Op()
			paint.FillShape(gtx.Ops, th.Fg, rect)
			return D{Size: size}
		}),
		layout.Flexed(1, func(gtx C) D {
			return ws.layMarkdownTab(gtx, th, &ws.tabs[ws.activeTab])
		}),
	)
}

func (ws *workspace) layTabList(gtx C, th *material.Theme) D {
	return material.List(th, &ws.tabList).Layout(gtx, len(ws.tabs), func(gtx C, i int) D {
		t := &ws.tabs[i]
		if t.btn.Clicked() {
			ws.selectTab(i)
			op.InvalidateOp{}.Add(gtx.Ops)
		}
		txt := t.name
		if t.isDirty() {
			txt += "*"
		}
		// If this is the active tab, emphasize the text and invert the bg & fg.
		lbl := material.Body1(th, txt)
		// lbl.Font.Variant = "Mono"
		bg := th.Bg
		if i == ws.activeTab {
			lbl.Font.Weight = text.Bold
			lbl.Color = bg
			bg = th.ContrastBg
		}
		// Record the layout in order to get the size for filling the background.
		m := op.Record(gtx.Ops)
		dims := t.btn.Layout(gtx, func(gtx C) D {
			return layout.UniformInset(5).Layout(gtx, lbl.Layout)
		})
		call := m.Stop()
		// Fill the background and draw the tab button.
		rect := clip.Rect{Max: dims.Size}
		paint.FillShape(gtx.Ops, bg, rect.Op())
		call.Add(gtx.Ops)
		return dims
	})
}

func (ws *workspace) layMarkdownTab(gtx C, th *material.Theme, t *tab) D {
	defer func() {
		if t.view.Editor.HasChanged() {
			ws.setLastEditAt(gtx.Now)
			t.lastEditAt = gtx.Now
		}
	}()
	if t.view.Editor.SaveRequested() && t.isDirty() {
		ws.saveQueue.pushBack(saveRequest{
			id:   t.id,
			data: t.view.Editor.Text(),
		})
		t.numQueuedSaves++
	}
	if t.numQueuedSaves > 0 {
		layout.NE.Layout(gtx, func(gtx C) D {
			return material.Loader(th).Layout(gtx)
		})
	}
	return mdedit.ViewStyle{
		Theme:      th,
		EditorFont: text.Font{Variant: "Mono"},
		Palette: mdedit.Palette{
			Fg:         th.Fg,
			Bg:         th.Bg,
			LineNumber: color.NRGBA{200, 180, 4, 125},
			Heading:    color.NRGBA{200, 193, 255, 255},
			ListMarker: color.NRGBA{10, 190, 240, 255},
			BlockQuote: color.NRGBA{165, 165, 165, 230},
			CodeBlock:  color.NRGBA{162, 120, 70, 255},
		},
		View: &t.view,
	}.Layout(gtx)
}

func (ws *workspace) insertTab(id lockbook.FileID, name string) {
	ws.tabs = append(ws.tabs, tab{})
	if len(ws.tabs) > 1 {
		ws.activeTab++
	}
	copy(ws.tabs[ws.activeTab+1:], ws.tabs[ws.activeTab:])
	t := tab{
		id:   id,
		name: name,
	}
	t.view.Mode = mdedit.ViewModeSingle
	t.view.SingleWidget = mdedit.SingleViewEditor
	t.view.SplitRatio = 0.5
	ws.tabs[ws.activeTab] = t
}

func (ws *workspace) selectTab(n int) {
	if len(ws.tabs) == 0 || n < 0 {
		return
	}
	if n >= len(ws.tabs) {
		n = len(ws.tabs) - 1
	}
	ws.activeTab = n
	ws.tabs[ws.activeTab].view.Editor.Focus()
}

func (ws *workspace) setTabMarkdown(id lockbook.FileID, data []byte) {
	for i := range ws.tabs {
		if ws.tabs[i].id == id {
			ws.tabs[i].view.Editor.SetText(data)
			ws.tabs[i].view.Editor.Focus()
			return
		}
	}
}

func (ws *workspace) closeActiveTab() {
	ws.tabs = append(ws.tabs[:ws.activeTab], ws.tabs[ws.activeTab+1:]...)
	if ws.activeTab >= len(ws.tabs) && ws.activeTab != 0 {
		ws.activeTab--
	}
}
