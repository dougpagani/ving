package trace

import (
	"github.com/gizak/termui"
	"github.com/yittg/ving/addons"
	"github.com/yittg/ving/addons/common"
)

const (
	traceHeight = 8
)

type ui struct {
	*common.TargetList

	selectChan   chan int
	manuallyChan chan bool

	lc     *termui.LineChart
	from   *termui.List
	start  bool
	source *runtime
}

func newUI(tr *runtime) *ui {
	return &ui{
		selectChan:   tr.traceSelected,
		manuallyChan: tr.traceManually,
		start:        tr.opt.Trace,
		source:       tr,
	}
}

// Activate see `ui.Activate`
func (tu *ui) Activate() {
	tu.source.updateStatus(true)
}

// Deactivate see `ui.AddOn`
func (tu *ui) Deactivate() {
	tu.source.updateStatus(false)
}

// Init see `AddOn`
func (tu *ui) Init() {
	cb := func(selected int) {
		tu.selectChan <- selected
	}
	opt := &common.TargetListOpt{
		SelectOnMove:      tu.start,
		CallBackImmediate: tu.start,
	}
	tu.TargetList = common.NewTargetList(cb, opt)
	tu.TargetList.Init(traceHeight)

	tu.lc = termui.NewLineChart()
	tu.lc.Height = traceHeight
	tu.lc.Mode = "dot"
	tu.lc.BorderTop = true
	tu.lc.BorderLeft = false
	tu.lc.BorderBottom = false
	tu.lc.BorderRight = false
	tu.lc.AxesColor = termui.ColorWhite
	tu.lc.BorderLabel = " ms "
	tu.lc.BorderLabelFg = termui.ColorWhite
	tu.lc.PaddingRight = 1
	tu.lc.LineColor["default"] = termui.ColorGreen | termui.AttrBold

	tu.from = termui.NewList()
	tu.from.BorderTop = true
	tu.from.BorderLeft = false
	tu.from.BorderBottom = false
	tu.from.BorderRight = false
	tu.from.Height = traceHeight
}

// Render see `AddOn`
func (tu *ui) Render() *termui.Row {
	return termui.NewRow(
		termui.NewCol(3, 0, tu.TargetList.Render()),
		termui.NewCol(6, 0, tu.lc),
		termui.NewCol(3, 0, tu.from),
	)
}

// UpdateState see `AddOn`
func (tu *ui) UpdateState(actives map[int]bool) {
	tu.TargetList.UpdateState(tu.source.rawTargets, actives)

	st, ok := tu.source.State().(*St)
	if !ok {
		return
	}
	if st != nil && st.ID == tu.TargetList.CurrentSelected() {
		shift := len(st.From) - tu.from.Height + 2
		if shift < 0 {
			shift = 0
		}
		tu.from.Items = st.From[shift:]
		tu.lc.Data = map[string][]float64{"default": st.Cost}
	} else {
		tu.lc.Data = nil
		tu.from.Items = []string{"<enter> to start"}
	}
}

// ToggleKey activate/deactivate this add-on
func (tu *ui) ToggleKey() string {
	return "t"
}

// RespondEvents see `AddOn`
func (tu *ui) RespondEvents() []addons.EventMeta {
	return []addons.EventMeta{
		{[]string{"n"}, "enter manually step-in mode"},
		{[]string{"c"}, "exit manually mode"},
	}
}

// HandleKeyEvent see `AddOn`
func (tu *ui) HandleKeyEvent(ev termui.Event) {
	if ev.Type != termui.KeyboardEvent {
		return
	}
	switch ev.ID {
	case "n":
		tu.handleN()
	case "c":
		tu.handleC()
	default:
		// ignore
	}
}

// ActivateAfterStart see `AddOn`
func (tu *ui) ActivateAfterStart() bool {
	return tu.start
}

func (tu *ui) handleC() {
	if tu.TargetList.CurrentSelected() < 0 {
		return
	}
	tu.manuallyChan <- false
}

func (tu *ui) handleN() {
	if tu.TargetList.CurrentSelected() < 0 {
		return
	}
	tu.manuallyChan <- true
}
