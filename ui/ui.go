package ui

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gizak/termui"
	"github.com/yittg/ving/types"
)

const defaultLoopPeriodic = time.Millisecond * 10

// Console display
type Console struct {
	width  int
	height int

	nItem         int
	errStatistics map[int]*errStatistic

	spGroup  *termui.Sparklines
	errGroup *termui.List

	loopPeriodic time.Duration
}

type errStatistic struct {
	id       int
	title    string
	count    int
	last     string
	lastIter uint64
}

// NewConsole init console
func NewConsole(targets []string) *Console {
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
	nTargets := len(targets)
	return &Console{
		width:         consoleWidth,
		height:        group.Height + errGroup.Height,
		spGroup:       group,
		errGroup:      errGroup,
		nItem:         nTargets,
		errStatistics: make(map[int]*errStatistic, nTargets),
		loopPeriodic:  defaultLoopPeriodic,
	}
}

func (c *Console) handleSpItem(item types.SpItem) {
	line := &(c.spGroup.Lines[item.Id])
	line.Data = append(line.Data[1:], item.Value)
	res := fmt.Sprintf("%v #%d", item.Display, item.Iter)
	format := fmt.Sprintf("%%s%%%dv", c.width-2-len(item.Target))
	line.Title = fmt.Sprintf(format, item.Target, res)
}

func (c *Console) handleErr(iter uint64, e types.ErrItem) {
	if old, ok := c.errStatistics[e.Id]; ok {
		old.count += 1
		old.lastIter = iter
		old.last = e.Err
		return
	}
	c.errStatistics[e.Id] = &errStatistic{
		id:       e.Id,
		title:    e.Target,
		count:    1,
		last:     e.Err,
		lastIter: iter,
	}
}

func (c *Console) displayErr(iter uint64) {
	if c.errGroup.Height < len(c.errStatistics) {
		c.errGroup.Height = len(c.errStatistics)
	}
	display := make([]string, 0, len(c.errStatistics))
	for i := 0; i < c.nItem; i += 1 {
		e, ok := c.errStatistics[i]
		if !ok {
			continue
		}
		title := fmt.Sprintf("* %s:%s", e.title, e.last)
		count := fmt.Sprintf("#%d", e.count)
		format := fmt.Sprintf("%%s%%%dv", c.width-1-len(title))
		if e.lastIter+50 >= iter { // remain 10 iter, i.e. 50 * 10ms
			format = fmt.Sprintf("[%s](fg-red)", format)
		}
		display = append(display, fmt.Sprintf(format, title, count))
	}
	c.errGroup.Items = display
}

func (c *Console) handleRes(iter uint64, res interface{}) {
	switch res.(type) {
	case types.SpItem:
		c.handleSpItem(res.(types.SpItem))
	case types.ErrItem:
		errItem := res.(types.ErrItem)
		c.handleErr(iter, errItem)
		c.handleSpItem(types.SpItem{
			ItemHeader: errItem.ItemHeader,
			Value:      0,
			Display:    "E",
		})
	default:
		// ignore
	}
	c.displayErr(iter)
	termui.Render(c.spGroup, c.errGroup)
}

// Run a spark line ui
func (c *Console) Run(resChan chan interface{}, onExit func()) {
	if err := termui.Init(); err != nil {
		panic(err)
	}
	defer termui.Close()

	termui.DefaultEvtStream.Merge("timer", termui.NewTimerCh(c.loopPeriodic))

	termui.Handle(fmt.Sprintf("/timer/%v", c.loopPeriodic), func(e termui.Event) {
		t := e.Data.(termui.EvtTimer)
		for {
			select {
			case res := <-resChan:
				c.handleRes(t.Count, res)
			default:
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

	termui.Loop()
}
