package main

import (
	"flag"
	"image"
	"image/color"
	"log"
	"os"
	"strconv"
	"time"

	"gioui.org/app"
	"gioui.org/font/opentype"
	"gioui.org/gesture"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget/material"
	"github.com/steverusso/gio-fonts/inconsolata/inconsolatabold"
	"github.com/steverusso/gio-fonts/inconsolata/inconsolataregular"
	"github.com/steverusso/gio-fonts/nunito/nunitobold"
	"github.com/steverusso/gio-fonts/nunito/nunitobolditalic"
	"github.com/steverusso/gio-fonts/nunito/nunitoitalic"
	"github.com/steverusso/gio-fonts/nunito/nunitoregular"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

type legitbook struct {
	win     *app.Window
	th      *material.Theme
	updates chan legitUpdate
	screen  screenState
	click   gesture.Click
	splash  splashScreen
	work    workspace
}

type legitUpdate any

type screenState int

const (
	showSplash screenState = iota
	showOnboard
	showWorkspace
)

func (lb *legitbook) frame(gtx C) {
	const topLevelKeySet = "Alt-[F]" +
		"|Alt-[1,2,3,4,5,6,7,8,9]" +
		"|Ctrl-[A,O,W," + key.NameTab + "]" +
		"|Ctrl-Shift-[" + key.NamePageUp + "," + key.NamePageDown + "," + key.NameTab + "]" +
		"|" + key.NameDeleteForward +
		"|" + key.NameEscape

	// Process any key events since the previous frame.
	hadActivity := false
	for _, e := range gtx.Events(lb.win) {
		hadActivity = true
		if e, ok := e.(key.Event); ok && e.State == key.Press {
			lb.handleKeyEvent(gtx, e)
		}
	}
	if hadActivity && lb.screen == showWorkspace {
		lb.work.setLastActionAt(gtx.Now)
	}
	// Gather key and pointer input on the entire window area.
	defer clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Push(gtx.Ops).Pop()
	key.InputOp{Tag: lb.win, Keys: topLevelKeySet}.Add(gtx.Ops)
	pointer.InputOp{
		Tag:   lb.win,
		Types: pointer.Press | pointer.Release | pointer.Move | pointer.Enter | pointer.Leave,
	}.Add(gtx.Ops)
	// Draw the active screen content.
	paint.Fill(gtx.Ops, lb.th.Bg)
	switch lb.screen {
	case showSplash:
		lb.splash.layout(gtx, lb.th)
	case showWorkspace:
		lb.work.layout(gtx, lb.th)
	}
}

func (lb *legitbook) handleKeyEvent(gtx C, e key.Event) {
	if lb.screen != showWorkspace {
		return
	}
	if e.Modifiers == key.ModAlt && e.Name == "F" && len(lb.work.tabs) > 0 {
		lb.work.animStage.reverse()
		return
	}
	switch lb.work.animStage {
	case wsExplOpen:
		lb.handleExplEvent(gtx, e)
	case wsExplClosed:
		lb.handleTabsEvent(gtx, e)
	}
}

func (lb *legitbook) handleExplEvent(gtx C, e key.Event) {
	switch e.Modifiers {
	case 0:
		switch e.Name {
		case key.NameDeleteForward:
			lb.work.deleteSelectedFiles()
		case key.NameEscape:
			lb.work.expl.deselectAll()
		}
	case key.ModCtrl:
		switch e.Name {
		case "A":
			lb.work.expl.selectAll()
		}
	}
}

func (lb *legitbook) handleTabsEvent(gtx C, e key.Event) {
	switch e.Modifiers {
	case key.ModAlt:
		switch e.Name {
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			n, _ := strconv.Atoi(e.Name)
			lb.work.selectTab(n - 1)
		}
	}
}

func run() error {
	win := app.NewWindow(
		app.Size(1050, 680),
		app.Title("Legitbook"),
	)

	th := material.NewTheme([]text.FontFace{
		// proportionals
		mustFont(text.Font{}, nunitoregular.TTF),
		mustFont(text.Font{Weight: text.Bold}, nunitobold.TTF),
		mustFont(text.Font{Weight: text.Bold, Style: text.Italic}, nunitobolditalic.TTF),
		mustFont(text.Font{Style: text.Italic}, nunitoitalic.TTF),
		// monos
		mustFont(text.Font{Variant: "Mono"}, inconsolataregular.TTF),
		mustFont(text.Font{Variant: "Mono", Weight: text.Bold}, inconsolatabold.TTF),
	})
	th.TextSize = 18
	th.Palette = material.Palette{
		Bg:         color.NRGBA{17, 21, 24, 255},
		Fg:         color.NRGBA{235, 235, 235, 255},
		ContrastFg: color.NRGBA{10, 180, 230, 255},
		ContrastBg: color.NRGBA{220, 220, 220, 255},
	}

	updates := make(chan legitUpdate)

	lb := legitbook{
		win:     win,
		th:      th,
		updates: updates,
		splash:  splashScreen{updates: updates},
	}
	go lb.splash.doStartupWork()

	var ops op.Ops
	for {
		select {
		case u := <-lb.updates:
			switch u := u.(type) {
			case setSplashErr:
				lb.splash.errMsg = u.msg
			case handoffToWorkspace:
				lb.work = newWorkspace(lb.updates, u)
				lb.screen = showWorkspace
				lb.splash = splashScreen{}
				go lb.work.manageSyncs()
				go lb.work.manageSaves()
			case handoffToOnboard:
				// todo(steve): design and impl the "onboard" screen
			case wsUpdate:
				lb.work.handleUpdate(u)
			}
			lb.win.Invalidate()
		case e := <-lb.win.Events():
			switch e := e.(type) {
			case system.FrameEvent:
				start := time.Now()
				gtx := layout.NewContext(&ops, e)
				lb.frame(gtx)
				e.Frame(gtx.Ops)
				if *printFrameTimes {
					log.Println(time.Now().Sub(start))
				}
			case system.DestroyEvent:
				return e.Err
			}
		}
	}
}

func mustFont(fnt text.Font, data []byte) text.FontFace {
	face, err := opentype.Parse(data)
	if err != nil {
		panic("failed to parse font: " + err.Error())
	}
	return text.FontFace{Font: fnt, Face: face}
}

var printFrameTimes = flag.Bool("print-frame-times", false, "Print how long each frame takes.")

func main() {
	flag.Parse()

	go func() {
		if err := run(); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	app.Main()
}
