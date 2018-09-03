package ui

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/gizak/termui"
	"github.com/yittg/ving/types"
)

type console struct {
	width  int
	height int

	errStatistics map[string]errStatistic

	spGroup  *termui.Sparklines
	errGroup *termui.List
}

type errStatistic struct {
	id    string
	order int
	count int
	last  string
}

type errStatisticSlice []errStatistic

func (s errStatisticSlice) Len() int {
	return len(s)
}

func (s errStatisticSlice) Less(i, j int) bool {
	return s[i].order < s[j].order
}

func (s errStatisticSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (c *console) spUpdate(t uint64, sps []types.SpItem) {
	for _, d := range sps {
		line := &(c.spGroup.Lines[d.Order])
		line.Data = append(line.Data[1:], d.Value)
		format := fmt.Sprintf("%%s%%%dv", c.width-2-len(d.Id))
		line.Title = fmt.Sprintf(format, d.Id, d.Display)
	}
	c.spGroup.BorderLabel = fmt.Sprintf("Ping(%d)", t)
}

func (c *console) errUpdateHistory(errs []types.ErrItem) map[string]bool {
	newErr := make(map[string]bool, len(errs))
	for _, e := range errs {
		if old, ok := c.errStatistics[e.Id]; ok {
			c.errStatistics[e.Id] = errStatistic{
				id:    e.Id,
				order: e.Order,
				count: old.count + 1,
				last:  e.Err,
			}
		} else {
			c.errStatistics[e.Id] = errStatistic{
				id:    e.Id,
				order: e.Order,
				count: 1,
				last:  e.Err,
			}
		}
		newErr[e.Id] = true
	}
	return newErr
}

func (c *console) errUpdate(_ uint64, errs []types.ErrItem) {
	newErr := c.errUpdateHistory(errs)

	errStatistics := make(errStatisticSlice, 0, len(c.errStatistics))
	for _, e := range c.errStatistics {
		errStatistics = append(errStatistics, e)
	}
	sort.Sort(errStatistics)

	if c.errGroup.Height < len(errStatistics) {
		c.errGroup.Height = len(errStatistics)
	}
	display := make([]string, 0, len(errStatistics))
	for _, e := range errStatistics {
		title := fmt.Sprintf("* %s:%s", e.id, e.last)
		format := fmt.Sprintf("%%s%%%dv", c.width-1-len(title))
		if _, ok := newErr[e.id]; ok {
			format = fmt.Sprintf("[%s](fg-red)", format)
		}
		display = append(display, fmt.Sprintf(format, title, e.count))
	}
	c.errGroup.Items = display
}

func prepareConsole(targets []string) *console {
	consoleWidth := 80

	spHeight := 3
	sparkLines := make([]termui.Sparkline, 0, len(targets))
	rand.Seed(time.Now().Unix())
	color := rand.Intn(termui.NumberofColors - 2)
	for i, target := range targets {
		sp := termui.Sparkline{}
		sp.Height = spHeight
		sp.Title = target
		sp.Data = make([]int, consoleWidth-2)
		sp.LineColor = termui.Attribute((color+i)%(termui.NumberofColors-2) + 2)
		sp.TitleColor = termui.ColorWhite
		sparkLines = append(sparkLines, sp)
	}

	group := termui.NewSparklines(sparkLines...)
	group.Width = consoleWidth
	group.Height = len(sparkLines)*(spHeight+1) + 2
	group.BorderLabel = "Ping"
	group.BorderLabelFg = termui.ColorCyan

	errGroup := termui.NewList()
	errGroup.Y = group.Height
	errGroup.Border = false
	errGroup.Height = 1
	errGroup.Width = consoleWidth

	return &console{
		width:         consoleWidth,
		height:        group.Height + errGroup.Height,
		spGroup:       group,
		errGroup:      errGroup,
		errStatistics: make(map[string]errStatistic, len(targets)),
	}
}

// Run a spark line ui
func Run(targets []string, dataSource func() (types.DataSet), onExit func()) {
	if err := termui.Init(); err != nil {
		panic(err)
	}
	defer termui.Close()

	console := prepareConsole(targets)

	draw := func(e termui.Event) {
		t := e.Data.(termui.EvtTimer)
		source := dataSource()
		console.spUpdate(t.Count, source.SpItems)
		console.errUpdate(t.Count, source.ErrItems)
		termui.Render(console.spGroup, console.errGroup)
	}

	stop := func() {
		onExit()
		termui.StopLoop()
	}

	termui.Handle("/sys/kbd/q", func(termui.Event) {
		stop()
	})
	termui.Handle("/sys/kbd/C-c", func(termui.Event) {
		stop()
	})
	termui.Handle("/timer/1s", func(e termui.Event) {
		draw(e)
	})
	termui.Loop()
}
