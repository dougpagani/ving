package common

import (
	"fmt"

	"github.com/gizak/termui"
	"github.com/yittg/ving/statistic"
)

// TargetList common view for add-ons
type TargetList struct {
	selectID int
	list     *termui.List

	selectedCb func(int)
}

// NewTargetList new a target list instance
func NewTargetList(selectedCb func(int)) *TargetList {
	return &TargetList{
		selectedCb: selectedCb,
	}
}

// Init init list
func (pu *TargetList) Init(height int) {
	pu.list = termui.NewList()
	pu.list.BorderTop = true
	pu.list.BorderLeft = false
	pu.list.BorderBottom = false
	pu.list.BorderRight = false
	pu.list.Height = height
}

// Reset list
func (pu *TargetList) Reset() {
	pu.selectID = 0
}

// CurrentSelected return current selected target ID
func (pu *TargetList) CurrentSelected() int {
	return pu.selectID
}

// OnEnter see `ConfirmAware`
func (pu *TargetList) OnEnter() {
	if pu.selectID < 0 {
		return
	}
	pu.selectedCb(pu.selectID)
}

// OnUp see `VerticalDirectionAware`
func (pu *TargetList) OnUp() {
	if len(pu.list.Items) == 0 {
		return
	}
	if pu.selectID < 0 {
		pu.selectID = 0
	} else {
		pu.selectID = (pu.selectID - 1 + len(pu.list.Items)) % len(pu.list.Items)
	}
}

// OnDown see `VerticalDirectionAware`
func (pu *TargetList) OnDown() {
	if len(pu.list.Items) == 0 {
		return
	}
	pu.selectID = (pu.selectID + 1) % len(pu.list.Items)
}

// Render list render
func (pu *TargetList) Render() termui.GridBufferer {
	return pu.list
}

// UpdateState of the target list component
func (pu *TargetList) UpdateState(sts []*statistic.Detail) {
	maxID := 0
	for _, st := range sts {
		if maxID < st.ID {
			maxID = st.ID
		}
	}

	if pu.selectID > maxID {
		pu.selectID = -1
	}
	items := make([]string, maxID+1)
	for _, st := range sts {
		if pu.selectID == st.ID {
			items[st.ID] = fmt.Sprintf("[* %s](fg-yellow)", st.Title)
		} else {
			items[st.ID] = "  " + st.Title
		}
	}
	pu.list.Items = items
}
