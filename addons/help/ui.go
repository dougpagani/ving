package help

import (
	"github.com/gizak/termui"
	"github.com/yittg/ving/addons"
)

type ui struct {
	msg *termui.List

	source *runtime
}

func newUI(h *runtime) *ui {
	return &ui{
		source: h,
	}
}

func (h *ui) Render() *termui.Row {
	return termui.NewRow(
		termui.NewCol(12, 0, h.msg),
	)
}

func (h *ui) Init() {
	h.msg = termui.NewList()
	h.msg.BorderTop = true
	h.msg.BorderLeft = false
	h.msg.BorderBottom = false
	h.msg.BorderRight = false
	h.msg.Height = 1
}

func (h *ui) Activate() {
}

func (h *ui) Deactivate() {
}

func (h *ui) ToggleKey() string {
	return "h"
}

func (h *ui) RespondEvents() []addons.EventMeta {
	return nil
}

func (h *ui) HandleKeyEvent(ev termui.Event) {
}

func (h *ui) ActivateAfterStart() bool {
	return false
}

func (h *ui) UpdateState(map[int]bool) {
	state, ok := h.source.State().([]string)
	if !ok {
		return
	}
	h.msg.Items = state
	h.msg.Height = len(h.msg.Items) + 1
}
