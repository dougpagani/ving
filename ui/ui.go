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
)

// Console display
type Console struct {
	nItem       int
	renderUnits []*renderUnit
	colors      map[int]termui.Attribute

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

	termui.Render(termui.Body)
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
