package common

import (
	"fmt"

	"github.com/gizak/termui"
)

// TargetList common view for add-ons
type TargetList struct {
	list *termui.List

	opt *TargetListOpt

	idxMap     map[int]int
	selectID   int
	lastCbID   int
	selectedCb func(int)
}

// TargetListOpt for target list
type TargetListOpt struct {
	// SelectOnMove no need to extra confirm
	SelectOnMove bool

	// InitSelected initial selected id
	InitSelected int

	// CallBackImmediate after init
	CallBackImmediate bool
}

// SelectedCallback for handler to handle the selected item
type SelectedCallback func(int)

// NewTargetList new a target list instance
func NewTargetList(selectedCb SelectedCallback, opt *TargetListOpt) *TargetList {
	return &TargetList{
		selectedCb: selectedCb,
		opt:        opt,
		idxMap:     make(map[int]int),
		selectID:   opt.InitSelected,
		lastCbID:   -1,
	}
}

// Init init list
func (tl *TargetList) Init(height int) {
	tl.list = termui.NewList()
	tl.list.BorderTop = true
	tl.list.BorderLeft = false
	tl.list.BorderBottom = false
	tl.list.BorderRight = false
	tl.list.Height = height
}

// CurrentSelected return current selected target ID
func (tl *TargetList) CurrentSelected() int {
	sid, ok := tl.idxMap[tl.selectID]
	if !ok {
		return -1
	}
	return sid
}

func (tl *TargetList) callBackSelected() {
	sid, ok := tl.idxMap[tl.selectID]
	if !ok {
		return
	}
	if sid == tl.lastCbID {
		return
	}
	tl.lastCbID = sid
	tl.selectedCb(sid)
}

// OnEnter see `ConfirmAware`
func (tl *TargetList) OnEnter() {
	if tl.selectID < 0 {
		return
	}
	tl.callBackSelected()
}

// OnUp see `VerticalDirectionAware`
func (tl *TargetList) OnUp() {
	if len(tl.list.Items) == 0 {
		return
	}
	if tl.selectID < 0 {
		tl.selectID = 0
	} else {
		tl.selectID = (tl.selectID - 1 + len(tl.list.Items)) % len(tl.list.Items)
	}
	if tl.opt.SelectOnMove {
		tl.callBackSelected()
	}
}

// OnDown see `VerticalDirectionAware`
func (tl *TargetList) OnDown() {
	if len(tl.list.Items) == 0 {
		return
	}
	tl.selectID = (tl.selectID + 1) % len(tl.list.Items)
	if tl.opt.SelectOnMove {
		tl.callBackSelected()
	}
}

// Render list render
func (tl *TargetList) Render() termui.GridBufferer {
	return tl.list
}

// UpdateState of the target list component
func (tl *TargetList) UpdateState(targets []string, actives map[int]bool) {
	if len(actives) == 0 {
		tl.list.Items = nil
		return
	}

	selectID := tl.selectID
	oldSid, ok := tl.idxMap[selectID]
	if !ok {
		oldSid = -1
	}
	items := make([]string, 0, len(actives))
	idxMap := make(map[int]int, len(actives))
	listIdx := 0

	for id, target := range targets {
		if _, ok := actives[id]; !ok {
			continue
		}
		idxMap[listIdx] = id
		var item string
		if oldSid == id {
			selectID = listIdx
			item = fmt.Sprintf("[* %s](fg-yellow)", target)
		} else {
			item = "  " + target
		}
		items = append(items, item)
		listIdx++
	}
	tl.idxMap = idxMap
	tl.list.Items = items
	tl.selectID = selectID
	if _, ok := tl.idxMap[tl.selectID]; !ok {
		tl.selectID = -1
	}
	if tl.opt.CallBackImmediate {
		tl.callBackSelected()
		tl.opt.CallBackImmediate = false
	}
}
