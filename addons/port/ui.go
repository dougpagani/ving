package port

import (
	"strconv"
	"strings"

	"github.com/gizak/termui"
	"github.com/yittg/ving/addons/common"
	"github.com/yittg/ving/statistic"
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

type ui struct {
	*common.TargetList

	selectChan chan int
	par        *termui.Par
	start      bool

	view viewEnum

	source *runtime
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
	pu.TargetList = common.NewTargetList(func(selectedID int) {
		pu.selectChan <- selectedID
	})
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
	pu.source.Activate()
}

// Deactivate this add-on
func (pu *ui) Deactivate() {
	pu.source.Deactivate()
}

// ToggleKey represents key to toggle
func (pu *ui) ToggleKey() string {
	return "p"
}

// RespondEvents return all keys this add-on can handle
func (pu *ui) RespondEvents() []string {
	return []string{"v", "r"}
}

// HandleKeyEvent do handle key event
func (pu *ui) HandleKeyEvent(ev termui.Event) {
	switch ev.ID {
	case "v":
		pu.handleV()
	case "r":
		pu.handleR()
	}
}

func (pu *ui) handleV() {
	pu.view++
	if pu.view == viewEnd {
		pu.view = 0
	}
}

func (pu *ui) handleR() {
	pu.source.resetTargetIter(pu.TargetList.CurrentSelected())
}

// ActivateAfterStart see `addons.ActivateAfterStart`
func (pu *ui) ActivateAfterStart() bool {
	return pu.start
}

func (pu *ui) buildPortView(p port) string {
	switch pu.view {
	case viewName:
		return p.name
	case viewPort:
		return strconv.Itoa(p.port)
	case viewAll:
		if strings.Index(p.name, strconv.Itoa(p.port)) >= 0 {
			return p.name
		}
		return p.name + ":" + strconv.Itoa(p.port)
	default:
		return p.name
	}
}

// UpdateState of this add-on
func (pu *ui) UpdateState(sts []*statistic.Detail) {
	pu.TargetList.UpdateState(sts)

	st, ok := pu.source.RenderState().(map[int][]touchResultWrapper)
	if !ok {
		return
	}
	selected := pu.CurrentSelected()
	thisSt := st[selected]
	text := ""
	if pu.source.checkDone(selected) {
		text = "[✔](fg-green)  "
	}
	for i, trw := range thisSt {
		if i > 0 {
			text += " | "
		}
		if trw.res == nil {
			text += "[•](fg-grey)"
		} else if trw.res.connected {
			text += "[•](fg-green)"
		} else {
			text += "[•](fg-red)"
		}
		text += " " + pu.buildPortView(trw.port)
	}
	if text == "" {
		text = "<enter> to start/continue"
	}
	pu.par.Text = text
}
