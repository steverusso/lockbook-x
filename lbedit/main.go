package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/font/opentype"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/text"
	"gioui.org/widget/material"
	"github.com/steverusso/gio-fonts/inconsolata/inconsolatabold"
	"github.com/steverusso/gio-fonts/inconsolata/inconsolataregular"
	"github.com/steverusso/gio-fonts/nunito/nunitobold"
	"github.com/steverusso/gio-fonts/nunito/nunitobolditalic"
	"github.com/steverusso/gio-fonts/nunito/nunitoitalic"
	"github.com/steverusso/gio-fonts/nunito/nunitoregular"
	"github.com/steverusso/lockbook-x/go-lockbook"
	"github.com/steverusso/mdedit"
)

type lockbookFS struct {
	core        lockbook.Core
	root        string
	lastOpenDir string
}

func newLockbookFS(core lockbook.Core) *lockbookFS {
	return &lockbookFS{
		core: core,
		root: "/",
	}
}

func (lbfs *lockbookFS) HomeDir() string {
	return lbfs.root
}

func (lbfs *lockbookFS) WorkingDir() string {
	return lbfs.lastOpenDir
}

func (lbfs *lockbookFS) ReadDir(fpath string) ([]fs.FileInfo, error) {
	dir, err := lbfs.core.FileByPath(fpath)
	if err != nil {
		return nil, fmt.Errorf("file by path %q: %w", fpath, err)
	}
	files, err := lbfs.core.GetChildren(dir.ID)
	if err != nil {
		return nil, fmt.Errorf("getting children of %q: %w", dir.ID, err)
	}
	infos := make([]fs.FileInfo, len(files))
	for i, f := range files {
		infos[i] = &lbFileInfo{f}
	}
	return infos, nil
}

func (lbfs *lockbookFS) ReadFile(fpath string) ([]byte, error) {
	f, err := lbfs.core.FileByPath(fpath)
	if err != nil {
		return nil, fmt.Errorf("file by path %q: %w", fpath, err)
	}
	data, err := lbfs.core.ReadDocument(f.ID)
	if err != nil {
		return nil, fmt.Errorf("reading doc %q: %w", f.ID, err)
	}
	parent := path.Dir(fpath)
	if parent == "" {
		parent = "/"
	}
	lbfs.lastOpenDir = parent
	return data, nil
}

func (lbfs *lockbookFS) WriteFile(fpath string, data []byte) error {
	f, err := lbfs.core.FileByPath(fpath)
	if err != nil {
		return fmt.Errorf("file by path %q: %w", fpath, err)
	}
	if err = lbfs.core.WriteDocument(f.ID, data); err != nil {
		return fmt.Errorf("writing doc %q: %w", f.ID, err)
	}
	go func() {
		if err = lbfs.core.SyncAll(nil); err != nil {
			log.Printf("sync after doc write: %v", err)
		}
	}()
	return nil
}

type lbFileInfo struct {
	lockbook.File
}

func (li *lbFileInfo) Name() string {
	return li.File.Name
}

func (li *lbFileInfo) Mode() fs.FileMode {
	if li.IsDir() {
		return fs.ModeDir
	}
	return fs.ModePerm
}

func (li *lbFileInfo) ModTime() time.Time {
	return time.UnixMilli(li.File.Lastmod)
}

func (li *lbFileInfo) Size() int64 {
	return -1
}

func (li *lbFileInfo) Sys() any {
	return nil
}

const topLevelKeySet = "Ctrl-[O,W," + key.NameTab + "]" +
	"|Ctrl-Shift-[" + key.NamePageUp + "," + key.NamePageDown + "," + key.NameTab + "]" +
	"|Alt-[1,2,3,4,5,6,7,8,9]"

var printFrameTimes = flag.Bool("print-frame-times", false, "Print how long each frame takes.")

func run() error {
	win := app.NewWindow(
		app.Size(1500, 900),
		app.Title("LbEdit"),
	)
	win.Perform(system.ActionCenter)

	th := material.NewTheme([]text.FontFace{
		// Proportionals.
		mustFont(text.Font{}, nunitoregular.TTF),
		mustFont(text.Font{Weight: text.Bold}, nunitobold.TTF),
		mustFont(text.Font{Weight: text.Bold, Style: text.Italic}, nunitobolditalic.TTF),
		mustFont(text.Font{Style: text.Italic}, nunitoitalic.TTF),
		// Monos.
		mustFont(text.Font{Variant: "Mono"}, inconsolataregular.TTF),
		mustFont(text.Font{Variant: "Mono", Weight: text.Bold}, inconsolatabold.TTF),
	})
	th.TextSize = 16
	th.Palette = material.Palette{
		Bg:         color.NRGBA{17, 21, 24, 255},
		Fg:         color.NRGBA{235, 235, 235, 255},
		ContrastFg: color.NRGBA{10, 180, 230, 255},
		ContrastBg: color.NRGBA{220, 220, 220, 255},
	}

	// Figure out data directory.
	dataDir := os.Getenv("LOCKBOOK_PATH")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("getting user home dir: %v", err)
		}
		dataDir = filepath.Join(home, ".lockbook/cli")
	}
	core, err := lockbook.NewCore(dataDir)
	if err != nil {
		log.Fatalf("initializing core: %v", err)
	}
	if err = core.SyncAll(nil); err != nil {
		log.Fatalf("performing opening sync: %v", err)
	}

	fs := newLockbookFS(core)

	s := mdedit.NewSession(fs, win)
	for _, fpath := range flag.Args() {
		s.OpenFile(fpath)
	}
	s.FocusActiveTab()

	var ops op.Ops
	for {
		e := <-win.Events()
		switch e := e.(type) {
		case system.FrameEvent:
			start := time.Now()
			gtx := layout.NewContext(&ops, e)
			// Process any key events since the previous frame.
			for _, ke := range gtx.Events(win) {
				if ke, ok := ke.(key.Event); ok {
					s.HandleKeyEvent(ke)
				}
			}
			// Gather key input on the entire window area.
			areaStack := clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Push(gtx.Ops)
			key.InputOp{Tag: win, Keys: topLevelKeySet}.Add(gtx.Ops)
			s.Layout(gtx, th)
			areaStack.Pop()

			e.Frame(gtx.Ops)
			if *printFrameTimes {
				log.Println(time.Now().Sub(start))
			}
		case key.Event:
			if e.State != key.Press {
				break
			}
			switch e.Modifiers {
			case key.ModCtrl:
				switch e.Name {
				case "O":
					s.OpenFileExplorerTab()
				case "W":
					s.CloseActiveTab()
				case key.NameTab:
					s.NextTab()
				}
			case key.ModCtrl | key.ModShift:
				switch e.Name {
				case key.NamePageUp:
					s.SwapTabUp()
				case key.NamePageDown:
					s.SwapTabDown()
				case key.NameTab:
					s.PrevTab()
				}
			case key.ModAlt:
				if strings.Contains("123456789", e.Name) {
					if n, err := strconv.Atoi(e.Name); err == nil {
						s.SelectTab(n - 1)
					}
				}
			}
			win.Invalidate()
		case system.DestroyEvent:
			return e.Err
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
