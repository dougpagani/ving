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
	nItem  int
	colors map[int]termui.Attribute

	activeAddOn addons.UI
	addOns      []addons.UI

	maxColumnN   int
	chartColumnN int
	chartRowN    int
	active       int
	dead         int
}

// NewConsole init console
func NewConsole(nTargets int, addOns []addons.UI) *Console {
	maxColumnN := 1

	rand.Seed(time.Now().Unix())
	color := rand.Intn(termui.NumberofColors - 2)
	colors := make(map[int]termui.Attribute, nTargets)
	for i := 0; i < nTargets; i++ {
		colors[i] = termui.Attribute((color+i)%(termui.NumberofColors-2) + 2)
	}

	return &Console{
		nItem:      nTargets,
		maxColumnN: maxColumnN,
		colors:     colors,
		addOns:     addOns,
	}
}

func (c *Console) emptySpGroup() *termui.Sparklines {
	g := termui.NewSparklines()
	g.Border = false
	return g
}

func (c *Console) emptyList() *termui.List {
	l := termui.NewList()
	l.Border = false
	return l
}

func (c *Console) alignMainBlock(active, dead int) {
	if c.active == active && c.dead >= dead {
		return
	}
	if active == 0 {
		c.chartColumnN = 0
		c.chartRowN = 0
	} else {
		if active < c.maxColumnN {
			c.chartColumnN = active
		} else {
			c.chartColumnN = c.maxColumnN
		}
		c.chartRowN = (active + c.chartColumnN - 1) / c.chartColumnN
	}
	col := c.chartColumnN
	activeSpan, deadSpan := 12, 0
	if dead > 0 {
		activeSpan, deadSpan = 9, 3
		col++
	}
	cols := make([]*termui.Row, 0, col)
	for i := 0; i < c.chartColumnN; i++ {
		cols = append(cols, termui.NewCol(activeSpan/c.chartColumnN, 0, c.emptySpGroup()))
	}
	if dead > 0 {
		cols = append(cols, termui.NewCol(deadSpan, 0, c.emptyList()))
	}
	termui.Body.Rows[0].Cols = cols
	termui.Clear()
	termui.Body.Align()
}

func (c *Console) renderOneSp(sp *termui.Sparkline, width int, s *statistic.Detail) {
	lastRecord := s.LastRecord()
	if lastRecord == nil {
		return
	}

	var flag string
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

	title := fmt.Sprintf("%s %s", flag, s.Title)
	res := fmt.Sprintf("%v #%d[#%d]", lastRecord.View(), s.Total, s.ErrCount)
	textLen := width - 1
	format := fmt.Sprintf("%%-%ds%%%dv", textLen/2, textLen-textLen/2-1)
	sp.Title = fmt.Sprintf(format, title, res)
	sp.Data = s.Cost
	sp.LineColor = c.colors[s.ID]
	s.ResizeViewWindow(width - 1)
}

func (c *Console) adjustSpGroup(group *termui.Sparklines, unitSize int) {
	crtSize := len(group.Lines)
	if crtSize > unitSize {
		group.Lines = group.Lines[:unitSize]
		return
	}
	for i := crtSize; i < unitSize; i++ {
		sp := termui.Sparkline{}
		sp.Height = chartHeight
		sp.Title = "*"
		sp.TitleColor = termui.ColorWhite
		group.Lines = append(group.Lines, sp)
	}
}

func (c *Console) renderOneSpGroup(ord int, unit []*statistic.Detail) {
	group := termui.Body.Rows[0].Cols[ord].Widget.(*termui.Sparklines)
	c.adjustSpGroup(group, len(unit))
	height := 1
	for i := range group.Lines {
		sp := &(group.Lines[i])
		height += sp.Height + 1
		c.renderOneSp(sp, group.Width, unit[i])
	}
	group.Height = height
}

func (c *Console) renderDeads(ord int, deads []*statistic.Detail) {
	list := termui.Body.Rows[0].Cols[ord].Widget.(*termui.List)
	var items []string
	for _, dead := range deads {
		items = append(items,
			fmt.Sprintf("❌ [%s](fg-bold)", dead.Title),
			fmt.Sprintf(" %s", dead.LastRecord().View()))
	}
	list.Items = items
	list.Height = len(items)
}

// Render statistics
func (c *Console) Render(t time.Time, sts []*statistic.Detail) {
	total := len(sts)

	activeTargetSet := make(map[int]bool)
	var activeTargets []*statistic.Detail
	var deads []*statistic.Detail
	for _, st := range sts {
		if st.Dead {
			deads = append(deads, st)
			continue
		}
		activeTargetSet[st.ID] = true
		activeTargets = append(activeTargets, st)
	}
	activeTotal := len(activeTargets)
	c.alignMainBlock(activeTotal, total-activeTotal)
	ord := 0
	for i := 0; i < activeTotal; i += c.chartRowN {
		if i+c.chartColumnN >= activeTotal {
			c.renderOneSpGroup(ord, activeTargets[i:])
		} else {
			c.renderOneSpGroup(ord, activeTargets[i:i+c.chartRowN])
		}
		ord++
	}
	if len(deads) > 0 {
		c.renderDeads(ord, deads)
	}

	if c.activeAddOn != nil {
		c.activeAddOn.UpdateState(activeTargetSet)
	}
	termui.Body.Align()
	termui.Render(termui.Body)
}

func (c *Console) setAddOn(addOn addons.UI) {
	c.activeAddOn = addOn
	termui.Body.AddRows(addOn.Render())
	termui.Clear()
	termui.Body.Align()
	addOn.Activate()
}

func (c *Console) removeAddOn() {
	if c.activeAddOn == nil {
		return
	}
	c.activeAddOn.Deactivate()
	c.activeAddOn = nil
	termui.Body.Rows = termui.Body.Rows[:1]
	termui.Clear()
	termui.Body.Align()
}

func (c *Console) toggleAddOn(addOn addons.UI) {
	if c.activeAddOn == addOn {
		c.removeAddOn()
	} else {
		if c.activeAddOn != nil {
			c.removeAddOn()
			// render the canvas to avoid residual shadow
			termui.Render(termui.Body)
		}
		c.setAddOn(addOn)
	}
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
		for _, em := range addOn.RespondEvents() {
			keys = append(keys, em.Keys...)
		}
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

	termui.Body.AddRows(
		termui.NewRow(),
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
		termui.Render(termui.Body)
	})
	c.registerAddOnEvents(systemKeys)
	termui.Loop()
}
