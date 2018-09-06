package ui

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gizak/termui"
	"github.com/yittg/ving/types"
)

const (
	defaultLoopPeriodic = time.Millisecond * 10
	errStatisticWindow  = 1000
	errHighlightWindow  = 50

	chartHeight = 3
)

// Console display
type Console struct {
	nItem      int
	statistics map[int]*statistic

	chartColumnN int
	chartRowN    int
	spGroup      []*termui.Sparklines
	errGroup     *termui.List

	loopPeriodic time.Duration
}

type statistic struct {
	id    int
	title string

	total         int
	errCount      int
	iter          uint64
	lastDisplay   interface{}
	spValue       []int
	lastErr       string
	lastNIterErrs []uint64

	block *termui.Sparkline
	group *termui.Sparklines
}

// NewConsole init console
func NewConsole(targets []string) *Console {
	nTargets := len(targets)
	chartColumn := 1
	chartRow := (nTargets + chartColumn - 1) / chartColumn
	sparkLines := make([]termui.Sparkline, 0, nTargets)
	rand.Seed(time.Now().Unix())
	color := rand.Intn(termui.NumberofColors - 2)
	for i, target := range targets {
		sp := termui.Sparkline{}
		sp.Height = chartHeight
		sp.Title = target
		sp.LineColor = termui.Attribute((color+i)%(termui.NumberofColors-2) + 2)
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

	errGroup := termui.NewList()
	errGroup.Border = false
	errGroup.Height = 2

	return &Console{
		spGroup:      groups,
		errGroup:     errGroup,
		nItem:        nTargets,
		chartColumnN: chartColumn,
		chartRowN:    chartRow,
		statistics:   make(map[int]*statistic, nTargets),
		loopPeriodic: defaultLoopPeriodic,
	}
}

func (c *Console) handleSpItem(s *statistic, item types.SpItem) {
	size := c.dataLen(s)
	if len(s.spValue) == 0 {
		s.spValue = make([]int, size)
	}
	s.total = item.Iter
	s.spValue = append(s.spValue[1:], item.Value)
	s.lastDisplay = item.Display

}

func (c *Console) resizeSpGroup() {
	for _, s := range c.statistics {
		crtSize := len(s.spValue)
		targetSize := c.dataLen(s)
		if crtSize == 0 || crtSize == targetSize {
			continue
		}
		if crtSize < targetSize {
			s.spValue = append(make([]int, targetSize-crtSize), s.spValue...)
		} else {
			s.spValue = s.spValue[crtSize-targetSize:]
		}
	}
}

func (c *Console) handleErr(s *statistic, e types.ErrItem) {
	s.errCount++
	s.lastErr = e.Err
	s.lastNIterErrs = append(s.lastNIterErrs, s.iter)
}

func (c *Console) renderSp(iter uint64) {
	for _, s := range c.statistics {
		res := fmt.Sprintf("%v #%d(#%d)", s.lastDisplay, s.total, s.errCount)
		format := fmt.Sprintf("%%s%%%dv", c.dataLen(s)-len(s.title))
		s.block.Title = fmt.Sprintf(format, s.title, res)
		s.block.Data = s.spValue
	}
}

func (c *Console) renderErr(iter uint64) {
	display := make([]string, 0, len(c.statistics))
	for i := 0; i < c.nItem; i++ {
		e, ok := c.statistics[i]
		if !ok || len(e.lastNIterErrs) == 0 || e.lastErr == "" {
			continue
		}
		lastErrIter := e.lastNIterErrs[len(e.lastNIterErrs)-1]

		title := fmt.Sprintf("* %s:%s", e.title, e.lastErr)
		format := "%s"
		if lastErrIter+errHighlightWindow >= iter {
			format = "[%s](fg-red)"
		}
		display = append(display, fmt.Sprintf(format, title))
	}
	if c.errGroup.Height < len(display) {
		c.errGroup.Height = len(display)
	}
	c.errGroup.Items = display
}

func (c *Console) getTarget(header types.ItemHeader) *statistic {
	target, ok := c.statistics[header.ID]
	if !ok {
		group, block := c.allocatedBlock(header.ID)
		target = &statistic{
			id:    header.ID,
			title: header.Target,
			total: header.Iter,
			block: block,
			group: group,
		}
		c.statistics[header.ID] = target
	}
	return target
}

func (c *Console) handleRes(iter uint64, res interface{}) {

	var target *statistic
	switch res.(type) {
	case types.SpItem:
		spItem := res.(types.SpItem)
		target = c.getTarget(spItem.ItemHeader)
		target.iter = iter
		c.handleSpItem(target, spItem)
	case types.ErrItem:
		errItem := res.(types.ErrItem)
		target = c.getTarget(errItem.ItemHeader)
		target.iter = iter
		c.handleErr(target, errItem)
		c.handleSpItem(target, types.SpItem{
			ItemHeader: errItem.ItemHeader,
			Value:      0,
			Display:    "E",
		})
	default:
		// ignore
	}
	c.renderSp(iter)
}

func (c *Console) width() int {
	return termui.Body.Width
}

func (c *Console) dataLen(s *statistic) int {
	return s.group.Width - 1
}

func (c *Console) errTextLen() int {
	return c.width() - 1
}

func (c *Console) allocatedBlock(idx int) (*termui.Sparklines, *termui.Sparkline) {
	groupID := idx / c.chartRowN
	subID := idx % c.chartRowN
	group := termui.Body.Rows[0].Cols[groupID].Widget.(*termui.Sparklines)
	sp := &(group.Lines[subID])
	return group, sp
}

// Run a spark line ui
func (c *Console) Run(resChan chan interface{}, onExit func()) {
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
		termui.NewRow(
			termui.NewCol(12, 0, c.errGroup),
		),
	)
	termui.Body.Align()

	termui.DefaultEvtStream.Merge("timer", termui.NewTimerCh(c.loopPeriodic))
	termui.Handle(fmt.Sprintf("/timer/%v", c.loopPeriodic), func(e termui.Event) {
		t := e.Data.(termui.EvtTimer)

		for _, s := range c.statistics {
			for i := 0; i < len(s.lastNIterErrs); i++ {
				if s.lastNIterErrs[i]+errStatisticWindow < t.Count {
					continue
				}
				s.lastNIterErrs = s.lastNIterErrs[i:]
				break
			}
		}
		for {
			select {
			case res := <-resChan:
				c.handleRes(t.Count, res)
			default:
				c.renderErr(t.Count)
				termui.Render(termui.Body)
				return
			}
		}
	})

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
	termui.Handle("/sys/wnd/resize", func(termui.Event) {
		termui.Body.Width = termui.TermWidth()
		termui.Body.Align()
		termui.Clear()
		c.resizeSpGroup()
		termui.Render(termui.Body)
	})

	termui.Loop()
}
