package port

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gizak/termui"
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
	selectID   int
	selectChan chan int

	list  *termui.List
	par   *termui.Par
	start bool

	view viewEnum

	source *runtime
}

// OnEnter see `ConfirmAware`
func (pu *ui) OnEnter() {
	if pu.selectID < 0 {
		return
	}
	pu.selectChan <- pu.selectID
}

// OnUp see `VerticalDirectionAware`
func (pu *ui) OnUp() {
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
func (pu *ui) OnDown() {
	if len(pu.list.Items) == 0 {
		return
	}
	pu.selectID = (pu.selectID + 1) % len(pu.list.Items)
}
func (pu *ui) Render() *termui.Row {
	return termui.NewRow(
		termui.NewCol(3, 0, pu.list),
		termui.NewCol(9, 0, pu.par),
	)
}

func (pu *ui) Init() {
	pu.list = termui.NewList()
	pu.list.BorderTop = true
	pu.list.BorderLeft = false
	pu.list.BorderBottom = false
	pu.list.BorderRight = false
	pu.list.Height = portsHeight

	pu.par = termui.NewPar("")
	pu.par.Height = portsHeight
	pu.par.BorderTop = true
	pu.par.BorderLeft = false
	pu.par.BorderBottom = false
	pu.par.BorderRight = false

	pu.Reset()
}

func (pu *ui) Reset() {
	pu.selectID = 0
}

func (pu *ui) Activate() {
	pu.source.Activate()
}

func (pu *ui) Deactivate() {
	pu.source.Deactivate()
}

func (pu *ui) ToggleKey() string {
	return "p"
}

func (pu *ui) RespondEvents() []string {
	return []string{"v", "r"}
}

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
	pu.source.resetTargetIter(pu.selectID)
}

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

func (pu *ui) UpdateState(sts []*statistic.Detail) {
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

	st, ok := pu.source.RenderState().(map[int][]touchResultWrapper)
	if !ok {
		return
	}
	thisSt := st[pu.selectID]
	text := ""
	if pu.source.checkDone(pu.selectID) {
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
	pu.par.Text = text
}
