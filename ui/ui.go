package ui

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gizak/termui"
	"github.com/yittg/ving/statistic"
)

const (
	chartHeight = 3
	traceHeight = 8
)

// Console display
type Console struct {
	nItem       int
	renderUnits []*renderUnit
	colors      map[int]termui.Attribute

	traceOn   bool
	traceUnit *traceUnit

	chartColumnN int
	chartRowN    int
	spGroup      []*termui.Sparklines
}

type renderUnit struct {
	statistic *statistic.Detail
	block     *termui.Sparkline
	group     *termui.Sparklines
}

type traceUnit struct {
	selectID     int
	selectChan   chan int
	manuallyChan chan bool
	statistic    *statistic.TraceSt
	list         *termui.List
	lc           *termui.LineChart
	from         *termui.List
}

// NewConsole init console
func NewConsole(nTargets int) *Console {
	chartColumn := 1
	chartRow := (nTargets + chartColumn - 1) / chartColumn
	sparkLines := make([]termui.Sparkline, 0, nTargets)
	rand.Seed(time.Now().Unix())
	color := rand.Intn(termui.NumberofColors - 2)
	colors := make(map[int]termui.Attribute, nTargets)
	for i := 0; i < nTargets; i++ {
		colors[i] = termui.Attribute((color+i)%(termui.NumberofColors-2) + 2)
		sp := termui.Sparkline{}
		sp.Height = chartHeight
		sp.Title = "*"
		sp.TitleColor = termui.ColorWhite
		sparkLines = append(sparkLines, sp)
	}

	groups := make([]*termui.Sparklines, 0, chartColumn)
	for i := 0; i < chartColumn; i++ {
		var members []termui.Sparkline
		if i == chartColumn-1 {
			members = sparkLines[i*chartRow:]
		} else {
			members = sparkLines[i*chartRow : (i+1)*chartRow]
		}

		g := termui.NewSparklines(members...)
		g.Height = chartRow*(chartHeight+1) + 1
		g.Border = false
		groups = append(groups, g)
	}

	return &Console{
		spGroup:      groups,
		nItem:        nTargets,
		chartColumnN: chartColumn,
		chartRowN:    chartRow,
		colors:       colors,
		renderUnits:  make([]*renderUnit, nTargets),
	}
}

func (c *Console) resizeSpGroup() {
	for _, s := range c.renderUnits {
		s.statistic.ResizeViewWindow(c.dataLen(s))
	}
}

func (c *Console) renderSp(t time.Time) {
	for _, g := range c.spGroup {
		g.Height = c.chartRowN*(chartHeight+1) + 1
	}
	for _, ru := range c.renderUnits {
		if ru == nil {
			continue
		}
		s := ru.statistic
		if s == nil {
			continue
		}
		lastRecord := s.LastRecord()
		if lastRecord == nil {
			continue
		}

		var flag string
		if s.Dead {
			flag = "❌"
		} else {
			rate := s.LastErrRate()
			if rate < 0.01 {
				flag = "🐸"
			} else if rate < 0.1 {
				flag = "🦁"
			} else {
				flag = "🙈"
			}
			if s.LastAverageCost() < int64(5*time.Millisecond) {
				flag += " ⚡️"
			}
		}

		title := fmt.Sprintf("%s %s", flag, s.Title)
		res := fmt.Sprintf("%v #%d[#%d]", lastRecord.View(), s.Total, s.ErrCount)
		textLen := c.dataLen(ru)
		format := fmt.Sprintf("%%-%ds%%%dv", textLen/2, textLen-textLen/2-1)
		ru.block.Title = fmt.Sprintf(format, title, res)
		ru.block.Data = s.Cost
		if s.Dead {
			ru.block.Height = 0
			ru.group.Height -= chartHeight
		} else {
			ru.block.Height = chartHeight
		}
	}
}

func (c *Console) renderTraceUnit(sts []*statistic.Detail, ts *statistic.TraceSt) {
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
	c.traceUnit.list.Items = items
	if c.traceUnit.selectID >= len(c.traceUnit.list.Items) {
		c.traceUnit.selectID = -1
	}
	if c.traceUnit.selectID >= 0 {
		c.traceUnit.list.Items[c.traceUnit.selectID] =
			fmt.Sprintf("[%s](bg-red)", c.traceUnit.list.Items[c.traceUnit.selectID])
	}
	c.traceUnit.statistic = ts
	if ts != nil {
		shift := len(ts.From) - c.traceUnit.from.Height + 2
		if shift < 0 {
			shift = 0
		}
		c.traceUnit.from.Items = ts.From[shift:]
		c.traceUnit.lc.Data = map[string][]float64{"default": ts.Cost}
	}
}

// Render statistics
func (c *Console) Render(t time.Time, sts []*statistic.Detail, ts *statistic.TraceSt) {
	for i, st := range sts {
		if c.renderUnits[i] == nil {
			sparklines, sparkline := c.allocatedBlock(i)
			c.renderUnits[i] = &renderUnit{
				statistic: st,
				group:     sparklines,
				block:     sparkline,
			}
		} else {
			c.renderUnits[i].statistic = st
		}
		c.renderUnits[i].block.LineColor = c.colors[st.ID]
		st.ResizeViewWindow(c.dataLen(c.renderUnits[i]))
	}

	c.renderSp(t)

	if c.traceOn {
		c.renderTraceUnit(sts, ts)
	}

	termui.Render(termui.Body)
}

// TraceOn check whether trace is active
func (c *Console) TraceOn() bool {
	return c.traceOn
}

// ToggleTrace hide trace block
func (c *Console) ToggleTrace(t time.Time, selectChan chan int, manuallyChan chan bool) bool {
	if c.traceOn {
		if len(termui.Body.Rows) > 1 {
			termui.Body.Rows = termui.Body.Rows[:1]
		}
	} else {
		if c.traceUnit == nil {
			listL := termui.NewList()
			listL.BorderTop = true
			listL.BorderLeft = false
			listL.BorderBottom = false
			listL.BorderRight = false
			listL.Height = traceHeight

			lc := termui.NewLineChart()
			lc.Height = traceHeight
			lc.Mode = "dot"
			lc.BorderTop = true
			lc.BorderLeft = false
			lc.BorderBottom = false
			lc.BorderRight = false
			lc.AxesColor = termui.ColorWhite
			lc.BorderLabel = " ms "
			lc.BorderLabelFg = termui.ColorWhite
			lc.PaddingRight = 1
			lc.LineColor["default"] = termui.ColorGreen | termui.AttrBold

			listR := termui.NewList()
			listR.BorderTop = true
			listR.BorderLeft = false
			listR.BorderBottom = false
			listR.BorderRight = false
			listR.Height = traceHeight
			c.traceUnit = &traceUnit{
				list: listL,
				lc:   lc,
				from: listR,
			}
		}
		c.traceUnit.selectID = 0
		c.traceUnit.selectChan = selectChan
		c.traceUnit.manuallyChan = manuallyChan
		c.traceUnit.from.Items = []string{}
		c.traceUnit.lc.Data = map[string][]float64{}

		traceRow := termui.NewRow(
			termui.NewCol(3, 0, c.traceUnit.list),
			termui.NewCol(6, 0, c.traceUnit.lc),
			termui.NewCol(3, 0, c.traceUnit.from),
		)
		if len(termui.Body.Rows) == 1 {
			termui.Body.AddRows(traceRow)
		} else {
			termui.Body.Rows[1] = traceRow
		}
	}
	termui.Clear()
	termui.Body.Align()
	c.traceOn = !c.traceOn
	return c.traceOn
}

func (c *Console) dataLen(ru *renderUnit) int {
	return ru.group.Width - 1
}

func (c *Console) allocatedBlock(idx int) (*termui.Sparklines, *termui.Sparkline) {
	groupID := idx / c.chartRowN
	subID := idx % c.chartRowN
	group := termui.Body.Rows[0].Cols[groupID].Widget.(*termui.Sparklines)
	sp := &(group.Lines[subID])
	return group, sp
}

// Run a spark line ui
func (c *Console) Run(stopChan chan bool, handlers ...EventHandler) {
	if err := termui.Init(); err != nil {
		panic(err)
	}
	defer termui.Close()

	groupCols := make([]*termui.Row, 0, len(c.spGroup))
	for _, g := range c.spGroup {
		groupCols = append(groupCols, termui.NewCol(12/c.chartColumnN, 0, g))
	}

	termui.Body.AddRows(
		termui.NewRow(groupCols...),
	)
	termui.Body.Align()

	termui.Handle("q", "<C-c>", func(termui.Event) {
		close(stopChan)
		termui.StopLoop()
	})
	termui.Handle("<Resize>", func(termui.Event) {
		termui.Body.Width = termui.TermWidth()
		termui.Body.Align()
		termui.Clear()
		c.resizeSpGroup()
		termui.Render(termui.Body)
	})
	termui.Handle("<Up>", func(event termui.Event) {
		if !c.traceOn {
			return
		}
		if c.traceUnit.selectID < 0 {
			c.traceUnit.selectID = 0
		} else {
			c.traceUnit.selectID = (c.traceUnit.selectID - 1 + len(c.traceUnit.list.Items)) % len(c.traceUnit.list.Items)
		}
	})
	termui.Handle("<Down>", func(event termui.Event) {
		if !c.traceOn {
			return
		}
		c.traceUnit.selectID = (c.traceUnit.selectID + 1) % len(c.traceUnit.list.Items)
	})

	termui.Handle("<Enter>", func(event termui.Event) {
		if !c.traceOn || c.traceUnit.selectID < 0 {
			return
		}
		c.traceUnit.selectChan <- c.traceUnit.selectID
	})

	termui.Handle("n", func(event termui.Event) {
		if !c.traceOn || c.traceUnit.selectID < 0 {
			return
		}
		c.traceUnit.manuallyChan <- true
	})

	termui.Handle("c", func(event termui.Event) {
		if !c.traceOn || c.traceUnit.selectID < 0 {
			return
		}
		c.traceUnit.manuallyChan <- false
	})

	for _, handler := range handlers {
		termui.Handle(handler.Key, func(termui.Event) {
			handler.Handler()
		})
		if handler.EmitAfterRun {
			handler.Handler()
		}
	}

	termui.Loop()
}

// EventHandler for register handler for event key
type EventHandler struct {
	Key          string
	EmitAfterRun bool
	Handler      func()
}
