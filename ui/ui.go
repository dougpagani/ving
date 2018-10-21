package ui

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gizak/termui"
	"github.com/yittg/ving/addons"
	"github.com/yittg/ving/config"
	"github.com/yittg/ving/statistic"
	"github.com/yittg/ving/types"
	"github.com/yittg/ving/utils/slices"
)

// Console display
type Console struct {
	colorSeed int

	activeAddOn addons.UI
	addOns      []addons.UI

	chartColumnN int
	chartRowN    int
	active       int
	dead         int
	collapseDead bool

	maxRowN         int
	sparklineHeight int
}

// NewConsole init console
func NewConsole(addOns []addons.UI) *Console {
	uiConfig := config.GetConfig().UI
	rand.Seed(time.Now().Unix())
	return &Console{
		colorSeed:       rand.Intn(termui.NumberofColors - 2),
		addOns:          addOns,
		maxRowN:         uiConfig.MaxRow,
		sparklineHeight: uiConfig.SparklineHeight,
	}
}

func (c *Console) color(id int) termui.Attribute {
	return termui.Attribute((c.colorSeed+id)%(termui.NumberofColors-2) + 2)
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
	c.active, c.dead = active, dead
	if active == 0 {
		c.chartColumnN = 0
		c.chartRowN = 0
	} else {
		if active < c.maxRowN {
			c.chartRowN = active
		} else {
			c.chartRowN = c.maxRowN
		}
		c.chartColumnN = (active + c.chartRowN - 1) / c.chartRowN
	}
	col := c.chartColumnN
	activeSpan, deadSpan := 12, 0
	if dead > 0 {
		if c.collapseDead {
			deadSpan = 1
		} else {
			deadSpan = 3
		}
		activeSpan = 12 - deadSpan
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

	errRateFlag := []string{"🐸", "🦁", "🙈"}
	maxLevel := len(errRateFlag) - 1
	errRateLevel := s.LastErrRateLevel()
	if errRateLevel > maxLevel {
		errRateLevel = maxLevel
	}
	flag := errRateFlag[errRateLevel]
	if s.LastAverageCost() < int64(5*time.Millisecond) {
		flag += " ⚡️"
	}

	title := fmt.Sprintf("%s %s", flag, s.Title)
	res := fmt.Sprintf("%v #%d[#%d]", lastRecord.View(), s.Total, s.ErrCount)
	textLen := width - 1
	format := fmt.Sprintf("%%-%ds%%%dv", textLen/2, textLen-textLen/2-1)
	sp.Title = fmt.Sprintf(format, title, res)
	sp.Data = s.Cost
	sp.LineColor = c.color(s.ID)
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
		sp.Height = c.sparklineHeight
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
	if c.collapseDead {
		item := fmt.Sprintf("❌ #%d", len(deads))
		format := fmt.Sprintf("%%%ds", list.Width-2)
		items = []string{fmt.Sprintf(format, item)}
	} else {
		for _, dead := range deads {
			items = append(items,
				fmt.Sprintf("❌ [%s](fg-bold)", dead.Title),
				fmt.Sprintf(" %s", dead.LastRecord().View()))
		}
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
		if i+c.chartRowN >= activeTotal {
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

func (c *Console) prepareGlobalKeys(stopChan chan bool) (systemKeys []string) {
	quitKey := types.EventMeta{
		Keys:        []string{"q", "<C-c>"},
		Description: "quit",
	}
	systemKeys = append(systemKeys, quitKey.Keys...)
	GlobalKeys = append(GlobalKeys, quitKey)
	termui.Handle(quitKey.Keys, func(termui.Event) {
		close(stopChan)
		termui.StopLoop()
	})

	collapseDeadKey := types.EventMeta{
		Keys:        []string{"E"},
		Description: "collapse dead targets if exist",
	}
	systemKeys = append(systemKeys, collapseDeadKey.Keys...)
	GlobalKeys = append(GlobalKeys, collapseDeadKey)
	termui.Handle(collapseDeadKey.Keys, func(termui.Event) {
		c.dead = 0 // trigger re-align main block
		c.collapseDead = !c.collapseDead
	})
	return
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
	systemKeys := c.prepareGlobalKeys(stopChan)
	termui.Handle("<Resize>", func(termui.Event) {
		termui.Body.Width = termui.TermWidth()
		termui.Body.Align()
		termui.Clear()
		termui.Render(termui.Body)
	})
	c.registerAddOnEvents(systemKeys)
	termui.Loop()
}

// GlobalKeys represents system global keys
var GlobalKeys []types.EventMeta
