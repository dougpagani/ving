package ui

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gizak/termui"
	"github.com/yittg/ving/net/protocol"
	"github.com/yittg/ving/statistic"
	"github.com/yittg/ving/types"
)

const (
	defaultLoopPeriodic = time.Millisecond * 10
	errHighlightWindow  = 50

	chartHeight = 3
)

// Console display
type Console struct {
	nItem       int
	renderUnits map[int]*renderUnit

	chartColumnN int
	chartRowN    int
	spGroup      []*termui.Sparklines
	errGroup     *termui.List

	loopPeriodic time.Duration
}

type renderUnit struct {
	statistic *statistic.Detail
	block     *termui.Sparkline
	group     *termui.Sparklines
}

// NewConsole init console
func NewConsole(targets []*protocol.NetworkTarget) *Console {
	nTargets := len(targets)
	chartColumn := 1
	chartRow := (nTargets + chartColumn - 1) / chartColumn
	sparkLines := make([]termui.Sparkline, 0, nTargets)
	rand.Seed(time.Now().Unix())
	color := rand.Intn(termui.NumberofColors - 2)
	for i, target := range targets {
		sp := termui.Sparkline{}
		sp.Height = chartHeight
		sp.Title = target.Raw
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
		renderUnits:  make(map[int]*renderUnit, nTargets),
		loopPeriodic: defaultLoopPeriodic,
	}
}

func (c *Console) resizeSpGroup() {
	for _, s := range c.renderUnits {
		s.statistic.ResizeViewWindow(c.dataLen(s))
	}
}

func (c *Console) retireRecord(t time.Time) {
	for _, ru := range c.renderUnits {
		if ru.statistic.Dead {
			continue
		}
		ru.statistic.RetireRecord(t)
	}
}

func (c *Console) renderSp(t time.Time) {
	for _, ru := range c.renderUnits {
		s := ru.statistic
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
	}
}

func (c *Console) renderErr(t time.Time) {
	display := make([]string, 0, len(c.renderUnits))
	for i := 0; i < c.nItem; i++ {
		ru, ok := c.renderUnits[i]
		if !ok {
			continue
		}
		s := ru.statistic
		lastErr := s.LastErrorRecord()
		if lastErr == nil {
			continue
		}
		title := fmt.Sprintf("* %s:%s", s.Title, lastErr.Err)
		format := "%s"
		if lastErr.T.Add(errHighlightWindow).After(t) {
			format = "[%s](fg-red)"
		}
		display = append(display, fmt.Sprintf(format, title))
	}
	if c.errGroup.Height < len(display) {
		c.errGroup.Height = len(display)
	}
	c.errGroup.Items = display
}

func (c *Console) render(t time.Time) {
	c.renderSp(t)
	c.renderErr(t)
}

func (c *Console) getRenderUnit(header types.RecordHeader) *renderUnit {
	target, ok := c.renderUnits[header.ID]
	if !ok {
		group, block := c.allocatedBlock(header.ID)
		target = &renderUnit{
			statistic: &statistic.Detail{
				ID:    header.ID,
				Title: header.Target.Raw,
				Total: header.Rounds,
			},
			block: block,
			group: group,
		}
		target.statistic.ResizeViewWindow(c.dataLen(target))
		c.renderUnits[header.ID] = target
	}
	return target
}

func (c *Console) width() int {
	return termui.Body.Width
}

func (c *Console) dataLen(ru *renderUnit) int {
	return ru.group.Width - 1
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
func (c *Console) Run(recordChan chan types.Record, stopChan chan bool) {
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

	go func() {
		ticker := time.NewTicker(c.loopPeriodic)
		for t := range ticker.C {
			func() {
				c.retireRecord(t)
				for {
					select {
					case res := <-recordChan:
						ru := c.getRenderUnit(res.RecordHeader)
						ru.statistic.DealRecord(t, res)
					default:
						c.render(t)
						termui.Render(termui.Body)
						return
					}
				}
			}()
		}
	}()

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

	termui.Loop()
}
