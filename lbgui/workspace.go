package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"time"

	"gioui.org/gesture"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/gofrs/uuid"
	"github.com/steverusso/lockbook-x/go-lockbook"
)

const (
	autoSyncInterval = time.Second * 5
	autoSaveInterval = time.Second * 3
)

type wsAnimStage int

const (
	wsExplOpen wsAnimStage = iota
	wsExplOpening
	wsExplClosed
	wsExplClosing
)

func (s *wsAnimStage) reverse() {
	if *s == wsExplOpen || *s == wsExplOpening {
		*s = wsExplClosing
	} else {
		*s = wsExplOpening
	}
}

type syncType uint8

const (
	syncTypeAuto syncType = iota
	syncTypeManual
)

type (
	wsUpdate interface{ implsWsUpdate() }

	openDirResult struct {
		id      lockbook.FileID
		parents []nameAndID
		files   []lockbook.File
		err     error
	}
	openFileResult struct {
		id   lockbook.FileID
		data []byte
		err  error
	}
	autoSaveScan  struct{}
	queuedSave    struct{ id lockbook.FileID }
	completedSave struct {
		id   lockbook.FileID
		err  error
		when time.Time
	}
	startSync  struct{ typ syncType }
	syncResult struct {
		typ       syncType
		newStatus string
		statusErr error
		syncErr   error
	}
)

func (openDirResult) implsWsUpdate()  {}
func (openFileResult) implsWsUpdate() {}
func (autoSaveScan) implsWsUpdate()   {}
func (queuedSave) implsWsUpdate()     {}
func (completedSave) implsWsUpdate()  {}
func (startSync) implsWsUpdate()      {}
func (syncResult) implsWsUpdate()     {}

type workspace struct {
	core       lockbook.Core
	updates    chan<- legitUpdate
	animStage  wsAnimStage
	animPct    float32
	tabs       []tab
	activeTab  int
	tabList    widget.List
	expl       fileExplorer
	bgErrs     []error
	botStatus  string
	modals     []modal
	modalCatch gesture.Click

	saveQueue     queue[saveRequest]
	lastActionAt  time.Time
	lastEditAt    time.Time
	nextSyncAt    time.Time
	autoSaveTimer *time.Timer
	autoSyncTimer *time.Timer
	manualSync    chan struct{}
	isSyncing     bool
}

func newWorkspace(updates chan<- legitUpdate, h handoffToWorkspace) workspace {
	ws := workspace{
		core:          h.core,
		updates:       updates,
		animPct:       1,
		modals:        make([]modal, 0, 3),
		saveQueue:     newQueueWithCapacity[saveRequest](8),
		lastActionAt:  time.Now(),
		autoSaveTimer: time.NewTimer(autoSaveInterval),
		autoSyncTimer: time.NewTimer(autoSyncInterval),
		manualSync:    make(chan struct{}),
	}
	ws.tabList.Axis = layout.Vertical
	ws.expl.entryList.Axis = layout.Vertical
	ws.expl.populate(nil, h.rootFiles)
	if h.lastSynced != "" {
		ws.botStatus = "Synced " + h.lastSynced
	}
	return ws
}

// setLastEditAt resets the auto-save timer if the duration between now and the edit
// before this one is longer than the auto-save interval.
func (ws *workspace) setLastEditAt(t time.Time) {
	sinceLastEdit := time.Now().Sub(ws.lastEditAt)
	if sinceLastEdit > autoSaveInterval {
		ws.autoSaveTimer.Reset(autoSaveInterval)
	}
	ws.lastEditAt = t
	ws.lastActionAt = t
}

// setLastActionAt triggers a sync if the duration between now and the next sync is longer
// than the auto-sync interval.
func (ws *workspace) setLastActionAt(t time.Time) {
	ws.lastActionAt = t
	if !ws.isSyncing && ws.nextSyncAt.Sub(time.Now()) > autoSyncInterval {
		ws.manualSync <- struct{}{}
	}
}

func (ws *workspace) manageSyncs() {
	for {
		select {
		case <-ws.autoSyncTimer.C:
			ws.updates <- startSync{syncTypeAuto}
		case <-ws.manualSync:
			ws.updates <- startSync{syncTypeManual}
		}
	}
}

func (ws *workspace) manageSaves() {
	// All save requests are taken out of the save queue by this single go routine.
	go func() {
		for {
			r := ws.saveQueue.popFront()
			err := ws.core.WriteDocument(r.id, r.data)
			now := time.Now()
			ws.updates <- completedSave{r.id, err, now}
		}
	}()
	for {
		<-ws.autoSaveTimer.C
		ws.updates <- autoSaveScan{}
	}
}

type saveRequest struct {
	id   lockbook.FileID
	data []byte
}

func (ws *workspace) tabByID(id lockbook.FileID) *tab {
	for i := range ws.tabs {
		if ws.tabs[i].id == id {
			return &ws.tabs[i]
		}
	}
	return nil
}

func (ws *workspace) sync(typ syncType) {
	r := syncResult{typ: typ}
	defer func() { ws.updates <- r }()

	if err := ws.core.SyncAll(nil); err != nil {
		r.syncErr = fmt.Errorf("syncing: %w", err)
		return
	}
	lastSynced, err := ws.core.GetLastSyncedHumanString()
	if err != nil {
		r.statusErr = fmt.Errorf("getting last synced: %w", err)
	} else {
		r.newStatus = "Synced " + lastSynced
	}
}

func (ws *workspace) handleUpdate(u wsUpdate) {
	switch u := u.(type) {
	case openDirResult:
		if u.err != nil {
			log.Printf("error: %v", u.err)
		} else if ws.expl.targetID == u.id {
			ws.expl.populate(u.parents, u.files)
		}
	case openFileResult:
		if u.err != nil {
			log.Printf("error: %v", u.err)
		} else {
			ws.setTabMarkdown(u.id, u.data)
		}
	case autoSaveScan:
		if ws.lastEditAt.IsZero() {
			break
		}
		for i := range ws.tabs {
			if ws.tabs[i].isDirty() {
				ws.saveQueue.pushBack(saveRequest{
					id:   ws.tabs[i].id,
					data: ws.tabs[i].view.Editor.Text(),
				})
				ws.tabs[i].numQueuedSaves++
			}
		}
		sinceLastEdit := time.Now().Sub(ws.lastEditAt)
		if sinceLastEdit < autoSaveInterval {
			ws.autoSaveTimer.Reset(autoSaveInterval)
		}
	case queuedSave:
		if t := ws.tabByID(u.id); t != nil {
			t.numQueuedSaves++
		}
	case completedSave:
		if u.err != nil {
			log.Printf("saving %s: %v", u.id, u.err) // todo(steve): needs to get to the ui
		}
		if t := ws.tabByID(u.id); t != nil {
			t.lastSaveAt = u.when
			t.numQueuedSaves--
		}
	case startSync:
		if ws.isSyncing {
			break
		}
		// If this is a manual sync, stop the auto sync timer or drain the channel if we
		// didn't stop it in time.
		if u.typ == syncTypeManual && !ws.autoSyncTimer.Stop() {
			<-ws.autoSyncTimer.C
		}
		ws.isSyncing = true
		go ws.sync(u.typ)
	case syncResult:
		ws.handleSyncResult(u)
	}
}

func (ws *workspace) handleSyncResult(sr syncResult) {
	if sr.syncErr != nil {
		ws.bgErrs = append(ws.bgErrs, sr.syncErr)
	}
	if sr.statusErr != nil {
		ws.bgErrs = append(ws.bgErrs, sr.statusErr)
	}
	if sr.newStatus != "" {
		ws.botStatus = sr.newStatus
	}
	switch sr.typ {
	case syncTypeAuto:
		now := time.Now()
		sinceLastAct := now.Sub(ws.lastActionAt)
		nextInterval := autoSyncInterval
		if sinceLastAct > autoSyncInterval {
			nextInterval = sinceLastAct
		}
		ws.autoSyncTimer.Reset(nextInterval)
		ws.nextSyncAt = now.Add(nextInterval)
	case syncTypeManual:
		ws.autoSyncTimer.Reset(autoSyncInterval)
		ws.nextSyncAt = time.Now().Add(autoSyncInterval)
	}
	ws.isSyncing = false
}

func (ws *workspace) layout(gtx C, th *material.Theme) D {
	if ws.expl.homeBtn.Clicked() {
		ws.openDir(uuid.Nil)
	}
	for i := range ws.expl.bcrumbs {
		if ws.expl.bcrumbs[i].btn.Clicked() {
			ws.openDir(ws.expl.bcrumbs[i].id)
		}
	}
	if ws.expl.mkdirBtn.Clicked() {
		ws.modals = append(ws.modals, newCreateFilePrompt(lockbook.FileTypeFolder{}))
	}

	_ = ws.layBaseLayer(gtx, th)
	ws.layModalLayer(gtx, th)
	return D{Size: gtx.Constraints.Max}
}

func (ws *workspace) layBaseLayer(gtx C, th *material.Theme) D {
	// Determine the height of the bottom bar so we can give the rest of the vertical
	// space to the explorer & tabs.
	m := op.Record(gtx.Ops)
	gtx.Constraints.Min.Y = 0
	botBarDims := ws.layBottomBar(gtx, th)
	drawBotBar := m.Stop()

	gtx.Constraints.Max.Y -= botBarDims.Size.Y
	_ = ws.layExplorerAndTabs(gtx, th)

	// Offset to after the explorer & tabs to place the bottom bar.
	offOp := op.Offset(image.Pt(0, gtx.Constraints.Max.Y)).Push(gtx.Ops)
	gtx.Constraints.Max.Y = botBarDims.Size.Y
	drawBotBar.Add(gtx.Ops)
	offOp.Pop()

	return D{Size: gtx.Constraints.Max}
}

func (ws *workspace) layBottomBar(gtx C, th *material.Theme) D {
	// background
	paint.FillShape(gtx.Ops, lighten(th.Bg, 0.1), clip.Rect{Max: gtx.Constraints.Max}.Op())

	// status circle
	height := int(th.TextSize * 1.5)
	diam := height - inset*2
	offOp := op.Offset(image.Pt(inset, height/2-diam/2)).Push(gtx.Ops)
	circle := clip.Ellipse{Max: image.Pt(diam, diam)}
	clr := color.NRGBA{0, 255, 0, 255}
	if len(ws.bgErrs) > 0 {
		clr = color.NRGBA{255, 0, 0, 255}
	}
	paint.FillShape(gtx.Ops, clr, circle.Op(gtx.Ops))
	offOp.Pop()

	m := op.Record(gtx.Ops)
	lblDims := material.Caption(th, ws.botStatus).Layout(gtx)
	lblCall := m.Stop()

	offOp = op.Offset(image.Pt(inset*2+diam, height/2-lblDims.Size.Y/2)).Push(gtx.Ops)
	lblCall.Add(gtx.Ops)
	offOp.Pop()

	return D{Size: image.Pt(gtx.Constraints.Max.X, height)}
}

func (ws *workspace) layExplorerAndTabs(gtx C, th *material.Theme) D {
	switch ws.animStage {
	case wsExplOpening:
		ws.animPct += 0.1
		if ws.animPct > 1 {
			ws.animPct = 1.0
			ws.animStage = wsExplOpen
		}
		op.InvalidateOp{}.Add(gtx.Ops)
	case wsExplClosing:
		ws.animPct -= 0.1
		if ws.animPct < 0 {
			ws.animPct = 0.0
			ws.animStage = wsExplClosed
		}
		op.InvalidateOp{}.Add(gtx.Ops)
	}

	switch ws.animStage {
	case wsExplClosed:
		return ws.layTabs(gtx, th)
	case wsExplOpen:
		return ws.layFileExplorer(gtx, th)
	}

	explVis := int(float32(gtx.Constraints.Max.X) * ws.animPct)
	offOp := op.Offset(image.Pt(0-(gtx.Constraints.Max.X-explVis), 0)).Push(gtx.Ops)
	_ = ws.layFileExplorer(gtx, th)
	offOp.Pop()

	offOp2 := op.Offset(image.Pt(explVis, 0)).Push(gtx.Ops)
	_ = ws.layTabs(gtx, th)
	offOp2.Pop()

	return D{Size: gtx.Constraints.Max}
}

func (ws *workspace) layModalLayer(gtx C, th *material.Theme) {
	if len(ws.modals) == 0 {
		return
	}
	paint.Fill(gtx.Ops, color.NRGBA{35, 35, 35, 240})
	switch m := ws.modals[len(ws.modals)-1].(type) {
	case *createFilePrompt:
		ws.layCreateFilePrompt(gtx, th, m)
	}
}

func (ws *workspace) layCreateFilePrompt(gtx C, th *material.Theme, p *createFilePrompt) D {
	for _, e := range p.input.Events() {
		if e, ok := e.(widget.SubmitEvent); ok {
			parent := ws.expl.targetID
			if parent == uuid.Nil {
				root, err := ws.core.GetRoot()
				if err != nil {
					p.err = err
					continue
				}
				parent = root.ID
			}
			f, err := ws.core.CreateFile(e.Text, parent, p.typ)
			if err != nil {
				p.err = err
				continue
			}
			ws.expl.add(&f)
			ws.modals = ws.modals[:len(ws.modals)-1]
			return D{}
		}
	}
	ws.modalCatch.Add(gtx.Ops)
	return layout.Center.Layout(gtx, func(gtx C) D {
		gtx.Constraints.Min.X = 200
		m := op.Record(gtx.Ops)
		dims := layout.UniformInset(12).Layout(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return material.Body1(th, "New "+lockbook.FileTypeString(p.typ)+":").Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: 12}.Layout),
				layout.Rigid(func(gtx C) D {
					return material.Editor(th, &p.input, "Hint...").Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					if p.err == nil {
						return D{}
					}
					return layout.Spacer{Height: 12}.Layout(gtx)
				}),
				layout.Rigid(func(gtx C) D {
					if p.err == nil {
						return D{}
					}
					return material.Body2(th, "error: "+p.err.Error()).Layout(gtx)
				}),
			)
		})
		draw := m.Stop()

		paint.FillShape(gtx.Ops, th.Bg, clip.UniformRRect(image.Rectangle{
			Max: image.Pt(dims.Size.X, dims.Size.Y),
		}, 8).Op(gtx.Ops))
		draw.Add(gtx.Ops)
		return dims
	})
}

func (ws *workspace) invalidate() {
	ws.updates <- nil
}
