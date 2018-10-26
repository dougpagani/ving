package port

import (
	"strconv"
	"strings"
	"time"

	"github.com/gizak/termui"
	"github.com/yittg/ving/addons/common"
	"github.com/yittg/ving/addons/port/types"
	et "github.com/yittg/ving/types"
)

const (
	portsHeight = 8
)

type viewEnum int

const (
	viewName viewEnum = iota
	viewPort
	viewAll
	viewEnd
)

type filterEnum int

const (
	all filterEnum = iota
	reached
	unReached
	unChecked
	end
)

type ui struct {
	*common.TargetList

	selectChan chan int
	par        *termui.Par
	start      bool

	view   viewEnum
	filter filterEnum

	source *runtime
}

func newUI(rt *runtime) *ui {
	return &ui{
		selectChan: rt.selected,
		start:      rt.opt.Ports,
		source:     rt,
	}
}

// Render of port add-on
func (pu *ui) Render() *termui.Row {
	return termui.NewRow(
		termui.NewCol(3, 0, pu.TargetList.Render()),
		termui.NewCol(9, 0, pu.par),
	)
}

// Init the port add-on view
func (pu *ui) Init() {
	cb := func(selectedID int) {
		pu.selectChan <- selectedID
	}
	opt := &common.TargetListOpt{
		SelectOnMove:      pu.start,
		CallBackImmediate: pu.start,
	}
	pu.TargetList = common.NewTargetList(cb, opt)
	pu.TargetList.Init(portsHeight)

	pu.par = termui.NewPar("")
	pu.par.Height = portsHeight
	pu.par.BorderTop = true
	pu.par.BorderLeft = false
	pu.par.BorderBottom = false
	pu.par.BorderRight = false
}

// Activate this add-on
func (pu *ui) Activate() {
	pu.source.updateStatus(true)
}

// Deactivate this add-on
func (pu *ui) Deactivate() {
	pu.source.updateStatus(false)
}

// ToggleKey represents key to toggle
func (pu *ui) ToggleKey() string {
	return "p"
}

// RespondEvents return all keys this add-on can handle
func (pu *ui) RespondEvents() []et.EventMeta {
	return []et.EventMeta{
		{Keys: []string{"v"}, Description: "change view mode, name, port number, or both"},
		{Keys: []string{"r"}, Description: "refresh and probe all ports again"},
		{Keys: []string{"f"}, Description: "filter ports list, reached, unreached, or all"},
	}
}

// HandleKeyEvent do handle key event
func (pu *ui) HandleKeyEvent(ev termui.Event) {
	switch ev.ID {
	case "f":
		pu.handleF()
	case "v":
		pu.handleV()
	case "r":
		pu.handleR()
	}
}

func (pu *ui) handleF() {
	pu.filter++
	if pu.filter == end {
		pu.filter = 0
	}
}

func (pu *ui) handleV() {
	pu.view++
	if pu.view == viewEnd {
		pu.view = 0
	}
}

func (pu *ui) handleR() {
	pu.source.resetTargetStatus(pu.TargetList.CurrentSelected())
}

// ActivateAfterStart see `addons.ActivateAfterStart`
func (pu *ui) ActivateAfterStart() bool {
	return pu.start
}

func (pu *ui) buildPortView(p types.PortDesc) string {
	switch pu.view {
	case viewName:
		return p.Name
	case viewPort:
		return strconv.Itoa(p.Port)
	case viewAll:
		if strings.Contains(p.Name, strconv.Itoa(p.Port)) {
			return p.Name
		}
		return p.Name + ":" + strconv.Itoa(p.Port)
	default:
		return p.Name
	}
}

func (pu *ui) filtered(res *touchResult) bool {
	switch pu.filter {
	case all:
		return true
	case reached:
		return res != nil && res.connected
	case unReached:
		return res != nil && !res.connected
	case unChecked:
		return res == nil
	default:
		return true
	}
}

// UpdateState of this add-on
func (pu *ui) UpdateState(t time.Time, actives map[int]bool) {
	pu.TargetList.UpdateState(pu.source.rawTargets, actives)

	st, ok := pu.source.State().(map[int][]touchResultWrapper)
	if !ok {
		return
	}
	selected := pu.CurrentSelected()
	thisSt, ok := st[selected]
	if !ok {
		pu.par.Text = "<enter> to start/continue"
		return
	}
	text := ""
	if pu.source.checkDone(selected) {
		text = "[✔](fg-green)  "
	}

	if pu.filter == reached {
		text += "[Reached](fg-green) "
	} else if pu.filter == unReached {
		text += "[unReached](fg-red) "
	} else if pu.filter == unChecked {
		text += "[unChecked](fg-grey) "
	}

	matched := false
	for _, trw := range thisSt {
		if !pu.filtered(trw.res) {
			continue
		}
		if matched {
			text += " | "
		} else {
			matched = true
		}
		if trw.res == nil {
			text += "[•](fg-grey)"
		} else if trw.res.connected {
			text += "[•](fg-green)"
		} else {
			text += "[•](fg-red)"
		}
		text += " " + pu.buildPortView(trw.port)
		if len(text) > 4096 {
			break
		}
	}
	if !matched {
		text += "none ports"
	}
	pu.par.Text = text
}
