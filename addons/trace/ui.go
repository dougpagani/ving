package trace

import (
	"fmt"

	"github.com/gizak/termui"
	"github.com/yittg/ving/statistic"
)

const (
	traceHeight = 8
)

type ui struct {
	selectID     int
	selectChan   chan int
	manuallyChan chan bool
	list         *termui.List
	lc           *termui.LineChart
	from         *termui.List
	start        bool
	source       *runtime
}

// Activate see `ui.Activate`
func (tu *ui) Activate() {
	tu.source.Activate()
}

// Deactivate see `ui.AddOn`
func (tu *ui) Deactivate() {
	tu.source.Deactivate()
}

// Init see `AddOn`
func (tu *ui) Init() {
	tu.list = termui.NewList()
	tu.list.BorderTop = true
	tu.list.BorderLeft = false
	tu.list.BorderBottom = false
	tu.list.BorderRight = false
	tu.list.Height = traceHeight

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

	tu.Reset()
}

// Reset see `AddOn`
func (tu *ui) Reset() {
	tu.selectID = 0
	tu.from.Items = []string{}
	tu.lc.Data = map[string][]float64{}
}

// Render see `AddOn`
func (tu *ui) Render() *termui.Row {
	return termui.NewRow(
		termui.NewCol(3, 0, tu.list),
		termui.NewCol(6, 0, tu.lc),
		termui.NewCol(3, 0, tu.from),
	)
}

// UpdateState see `AddOn`
func (tu *ui) UpdateState(sts []*statistic.Detail) {
	maxID := 0
	for _, st := range sts {
		if maxID < st.ID {
			maxID = st.ID
		}
	}

	items := make([]string, maxID+1)
	for _, st := range sts {
		items[st.ID] = st.Title
	}
	tu.list.Items = items
	if tu.selectID >= len(tu.list.Items) {
		tu.selectID = -1
	}
	if tu.selectID >= 0 {
		tu.list.Items[tu.selectID] =
			fmt.Sprintf("[%s](bg-red)", tu.list.Items[tu.selectID])
	}

	st, ok := tu.source.RenderState().(*statistic.TraceSt)
	if !ok {
		return
	}
	if st != nil {
		shift := len(st.From) - tu.from.Height + 2
		if shift < 0 {
			shift = 0
		}
		tu.from.Items = st.From[shift:]
		tu.lc.Data = map[string][]float64{"default": st.Cost}
	}
}

// ToggleKey activate/deactivate this add-on
func (tu *ui) ToggleKey() string {
	return "t"
}

// RespondEvents see `AddOn`
func (tu *ui) RespondEvents() []string {
	return []string{"n", "c"}
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

// OnEnter see `ConfirmAware`
func (tu *ui) OnEnter() {
	if tu.selectID < 0 {
		return
	}
	tu.selectChan <- tu.selectID
}

func (tu *ui) handleC() {
	if tu.selectID < 0 {
		return
	}
	tu.manuallyChan <- false
}

func (tu *ui) handleN() {
	if tu.selectID < 0 {
		return
	}
	tu.manuallyChan <- true
}

// OnUp see `VerticalDirectionAware`
func (tu *ui) OnUp() {
	if len(tu.list.Items) == 0 {
		return
	}
	if tu.selectID < 0 {
		tu.selectID = 0
	} else {
		tu.selectID = (tu.selectID - 1 + len(tu.list.Items)) % len(tu.list.Items)
	}
}

// OnDown see `VerticalDirectionAware`
func (tu *ui) OnDown() {
	if len(tu.list.Items) == 0 {
		return
	}
	tu.selectID = (tu.selectID + 1) % len(tu.list.Items)
}
