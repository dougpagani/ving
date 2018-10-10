package common

import (
	"fmt"

	"github.com/gizak/termui"
	"github.com/yittg/ving/statistic"
)

// TargetList common view for add-ons
type TargetList struct {
	list *termui.List

	opt *TargetListOpt

	selectID     int
	lastSelectID int
	selectedCb   func(int)
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

// NewTargetList new a target list instance
func NewTargetList(selectedCb func(int), opt *TargetListOpt) *TargetList {
	return &TargetList{
		selectedCb:   selectedCb,
		opt:          opt,
		selectID:     opt.InitSelected,
		lastSelectID: -1,
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
	if tl.opt.CallBackImmediate {
		tl.callBackSelected()
	}
}

// CurrentSelected return current selected target ID
func (tl *TargetList) CurrentSelected() int {
	return tl.selectID
}

func (tl *TargetList) callBackSelected() {
	if tl.selectID == tl.lastSelectID {
		return
	}
	tl.lastSelectID = tl.selectID
	tl.selectedCb(tl.selectID)
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
func (tl *TargetList) UpdateState(sts []*statistic.Detail) {
	maxID := 0
	for _, st := range sts {
		if maxID < st.ID {
			maxID = st.ID
		}
	}

	if tl.selectID > maxID {
		tl.selectID = -1
	}
	items := make([]string, maxID+1)
	for _, st := range sts {
		if tl.selectID == st.ID {
			items[st.ID] = fmt.Sprintf("[* %s](fg-yellow)", st.Title)
		} else {
			items[st.ID] = "  " + st.Title
		}
	}
	tl.list.Items = items
}
