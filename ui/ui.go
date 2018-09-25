package ui

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gizak/termui"
	"github.com/yittg/ving/addons"
	"github.com/yittg/ving/statistic"
	"github.com/yittg/ving/utils/slices"
)

const (
	chartHeight = 3
)

// Console display
type Console struct {
	nItem       int
	renderUnits []*renderUnit
	colors      map[int]termui.Attribute

	activeAddOn addons.UI
	addOns      []addons.UI

	chartColumnN int
	chartRowN    int
	spGroup      []*termui.Sparklines
}

type renderUnit struct {
	statistic *statistic.Detail
	block     *termui.Sparkline
	group     *termui.Sparklines
}

// NewConsole init console
func NewConsole(nTargets int, addOns []addons.UI) *Console {
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
		addOns:       addOns,
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

// Render statistics
func (c *Console) Render(t time.Time, sts []*statistic.Detail) {
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

	if c.activeAddOn != nil {
		c.activeAddOn.UpdateState(sts)
	}

	termui.Render(termui.Body)
}

func (c *Console) setAddOn(addOn addons.UI) {
	c.activeAddOn = addOn
	if len(termui.Body.Rows) == 1 {
		termui.Body.AddRows(addOn.Render())
	} else {
		termui.Body.Rows[1] = addOn.Render()
	}
	termui.Clear()
	termui.Body.Align()
	addOn.Activate()
}

func (c *Console) removeAddOn() {
	c.activeAddOn.Deactivate()
	c.activeAddOn = nil
	if len(termui.Body.Rows) == 1 {
		return
	}
	termui.Body.Rows = termui.Body.Rows[:1]
	termui.Clear()
	termui.Body.Align()
}

func (c *Console) toggleAddOn(addOn addons.UI) {
	if c.activeAddOn == addOn {
		c.removeAddOn()
	} else {
		c.setAddOn(addOn)
	}
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

func (c *Console) registerAddOnEvents(systemKeys []string) {
	onAddOnActive := func(f func(termui.Event)) func(termui.Event) {
		return func(event termui.Event) {
			if c.activeAddOn == nil {
				return
			}
			f(event)
		}
	}

	termui.Handle("<Enter>", onAddOnActive(func(event termui.Event) {
		if cAwareAddOn, ok := c.activeAddOn.(addons.ConfirmAware); ok {
			cAwareAddOn.OnEnter()
		}
	}))

	systemKeys = append(systemKeys, "j", "k")
	termui.Handle("<Up>", "<Down>", "j", "k", onAddOnActive(func(event termui.Event) {
		if vdAwareAddOn, ok := c.activeAddOn.(addons.VerticalDirectionAware); ok {
			switch event.ID {
			case "<Up>", "k":
				vdAwareAddOn.OnUp()
			case "<Down>", "j":
				vdAwareAddOn.OnDown()
			}
		}
	}))

	var keys []string
	for _, addOn := range c.addOns {
		keys = append(keys, addOn.RespondEvents()...)
		addOn.Init()
		if addOn.ActivateAfterStart() {
			c.setAddOn(addOn)
		}
		termui.Handle(addOn.ToggleKey(), func(a addons.UI) func(termui.Event) {
			return func(termui.Event) {
				c.toggleAddOn(a)
			}
		}(addOn))
	}
	for _, key := range keys {
		if len(key) == 0 || key[0] == '<' || slices.ContainStr(systemKeys, key) {
			continue
		}
		termui.Handle(key, onAddOnActive(func(event termui.Event) {
			c.activeAddOn.HandleKeyEvent(event)
		}))
	}
}

// Run a spark line ui
func (c *Console) Run(stopChan chan bool) {
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

	systemKeys := []string{"q"}
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
	c.registerAddOnEvents(systemKeys)
	termui.Loop()
}

// EventHandler for register handler for event key
type EventHandler struct {
	Key          string
	EmitAfterRun bool
	Handler      func()
}
