package main

import (
	"image"
	"image/color"

	"gioui.org/gesture"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

var (
	iconArrowDown  = mustIcon(icons.HardwareKeyboardArrowDown)
	iconArrowRight = mustIcon(icons.HardwareKeyboardArrowRight)
	iconDirectory  = mustIcon(icons.FileFolderOpen)
	iconDocument   = mustIcon(icons.ActionDescription)
	iconHome       = mustIcon(icons.ActionHome)
	iconNewDoc     = mustIcon(icons.ActionNoteAdd)
	iconNewFolder  = mustIcon(icons.FileCreateNewFolder)
	iconRegFile    = mustIcon(icons.ActionDescription)
)

// mustIcon returns a new `widget.Icon` for the given byte slice. It panics on error.
func mustIcon(data []byte) widget.Icon {
	ic, err := widget.NewIcon(data)
	if err != nil {
		panic(err)
	}
	return *ic
}

func vertCenter(gtx C, maxHeight int, w layout.Widget) {
	m := op.Record(gtx.Ops)
	dims := w(gtx)
	call := m.Stop()

	offOp := op.Offset(image.Pt(0, maxHeight/2-dims.Size.Y/2)).Push(gtx.Ops)
	call.Add(gtx.Ops)
	offOp.Pop()
}

func offsetAndLayHR(gtx C, th *material.Theme, yOffset int) int {
	offOp := op.Offset(image.Pt(0, yOffset)).Push(gtx.Ops)
	dims := rule{width: 1, color: th.Fg}.Layout(gtx)
	offOp.Pop()
	return dims.Size.Y
}

type rule struct {
	width int
	color color.NRGBA
	axis  layout.Axis
}

func (rl rule) Layout(gtx C) D {
	if rl.width == 0 {
		rl.width = 1
	}
	size := image.Point{gtx.Constraints.Max.X, rl.width}
	if rl.axis == layout.Vertical {
		size = image.Point{rl.width, gtx.Constraints.Max.Y}
	}
	rect := clip.Rect{Max: size}.Op()
	paint.FillShape(gtx.Ops, rl.color, rect)
	return D{Size: size}
}

type buttonGroupStyle struct {
	bg       color.NRGBA
	fg       color.NRGBA
	shaper   text.Shaper
	textSize unit.Sp
}

type groupButton struct {
	click    *widget.Clickable
	icon     *widget.Icon
	text     string
	disabled bool
}

func (g buttonGroupStyle) layout(gtx C, buttons []groupButton) D {
	if len(buttons) == 0 {
		return D{}
	}
	// Determine the height of the tallest element.
	var maxHeight int
	{
		for _, b := range buttons {
			m := op.Record(gtx.Ops)
			dims := g.drawButton(gtx, b)
			_ = m.Stop()
			if h := dims.Size.Y; h > maxHeight {
				maxHeight = h
			}
		}
	}
	// Make the flex wrapped buttons with dividers in between.
	border := merge(g.fg, g.bg, 0.3)
	divider := func(gtx C) D {
		gtx.Constraints.Max.Y = maxHeight
		return rule{color: border, axis: layout.Vertical}.Layout(gtx)
	}
	flexBtns := make([]layout.FlexChild, 0, len(buttons)*2-1)
	for i, b := range buttons {
		if i != 0 {
			flexBtns = append(flexBtns, layout.Rigid(divider))
		}
		b := b
		flexBtns = append(flexBtns, layout.Rigid(func(gtx C) D {
			return g.drawButtonWithBg(gtx, b, maxHeight)
		}))
	}
	// Determine this button group's full size for the rounded rectangle clip.
	const radius = 5
	m := op.Record(gtx.Ops)
	fullDims := widget.Border{
		Color:        border,
		CornerRadius: radius,
		Width:        1,
	}.Layout(gtx, func(gtx C) D {
		return layout.Flex{}.Layout(gtx, flexBtns...)
	})
	fullDraw := m.Stop()
	// Clip the rounded rectangle area and draw the button group.
	defer clip.RRect{
		Rect: image.Rectangle{Max: image.Point{
			X: fullDims.Size.X - 1,
			Y: fullDims.Size.Y - 1,
		}},
		SE: radius, SW: radius, NW: radius, NE: radius,
	}.Push(gtx.Ops).Pop()
	fullDraw.Add(gtx.Ops)
	return fullDims
}

func (g *buttonGroupStyle) drawButtonWithBg(gtx C, b groupButton, height int) D {
	// Pre-draw in order to get the width for filling the background.
	m := op.Record(gtx.Ops)
	dims := g.drawButton(gtx, b)
	call := m.Stop()
	// Adjust background color under certain conditions.
	bg := g.bg
	switch {
	case b.disabled:
		bg = darken(bg, 0.05)
	case b.click.Pressed():
		bg = lighten(bg, 0.1)
	case b.click.Hovered():
		bg = lighten(bg, 0.4)
	case !b.disabled:
		bg = lighten(bg, 0.2)
	}
	size := image.Point{X: dims.Size.X, Y: height}
	paint.FillShape(gtx.Ops, bg, clip.Rect{Max: size}.Op())
	// Vertically center the button content.
	defer op.Offset(image.Point{Y: height/2 - dims.Size.Y/2}).Push(gtx.Ops).Pop()
	return b.click.Layout(gtx, func(gtx C) D {
		call.Add(gtx.Ops)
		return D{Size: size}
	})
}

func (g *buttonGroupStyle) drawButton(gtx C, b groupButton) D {
	fg := g.fg
	if b.disabled {
		fg.A /= 2
	}
	gtx.Constraints.Max.X -= 12
	gtx.Constraints.Max.Y -= 12
	defer op.Offset(image.Pt(6, 6)).Push(gtx.Ops).Pop()
	dims := D{}
	switch {
	case b.icon != nil:
		dims = b.icon.Layout(gtx, fg)
	case b.text != "":
		gtx.Constraints.Max.X -= 4
		paint.ColorOp{Color: fg}.Add(gtx.Ops)
		lblOffOp := op.Offset(image.Pt(2, 0)).Push(gtx.Ops)
		dims = widget.Label{MaxLines: 1}.Layout(gtx, g.shaper, text.Font{}, g.textSize, b.text)
		dims.Size.X += 4
		lblOffOp.Pop()
	}
	return D{Size: image.Pt(dims.Size.X+12, dims.Size.Y+12)}
}

func darken(c color.NRGBA, f float32) color.NRGBA {
	return color.NRGBA{
		R: uint8(float32(c.R) * (1 - f)),
		G: uint8(float32(c.G) * (1 - f)),
		B: uint8(float32(c.B) * (1 - f)),
		A: 255,
	}
}

func lighten(c color.NRGBA, f float32) color.NRGBA {
	return color.NRGBA{
		R: c.R + uint8(float32(255-c.R)*f),
		G: c.G + uint8(float32(255-c.G)*f),
		B: c.B + uint8(float32(255-c.B)*f),
		A: 255,
	}
}

func merge(c1, c2 color.NRGBA, p float32) color.NRGBA {
	return color.NRGBA{
		R: mergeCalc(c1.R, c2.R, p),
		G: mergeCalc(c1.G, c2.G, p),
		B: mergeCalc(c1.B, c2.B, p),
		A: 255,
	}
}

func mergeCalc(a, b uint8, p float32) uint8 {
	v1 := float32(a) * (1 - p)
	v2 := float32(b) * p
	return uint8(v1 + v2)
}

type popupMenuButton struct {
	click   gesture.Click
	pressed bool
}

func (b *popupMenuButton) Pressed() bool {
	c := b.pressed
	b.pressed = false
	return c
}

func layPopupMenuItem(gtx C, th *material.Theme, btn *popupMenuButton, txt string) D {
	for _, e := range btn.click.Events(gtx) {
		if e.Type == gesture.TypePress {
			btn.pressed = true
			break
		}
	}

	m := op.Record(gtx.Ops)
	dims := material.Body2(th, txt).Layout(gtx)
	call := m.Stop()

	bg := th.Bg
	if btn.click.Hovered() {
		bg = th.ContrastFg
	}

	size := image.Pt(gtx.Constraints.Max.X, dims.Size.Y)
	defer clip.Rect(image.Rectangle{Max: size}).Push(gtx.Ops).Pop()

	paint.Fill(gtx.Ops, bg)
	btn.click.Add(gtx.Ops)
	call.Add(gtx.Ops)

	return D{Size: size}
}
