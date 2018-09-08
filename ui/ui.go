package ui

import (
	"fmt"
	"math"
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

	total             int
	errCount          int
	iter              uint64
	cost              []int
	lastErr           string
	lastErrIter       uint64
	lastNIterRecord   []recordWithIter
	lastNIterErrCount int
	lastNIterCost     int64

	dead bool

	block *termui.Sparkline
	group *termui.Sparklines
}

func (s *statistic) lastRecord() *recordWithIter {
	n := len(s.lastNIterRecord)
	if n == 0 {
		return nil
	}
	return &s.lastNIterRecord[n-1]
}

func (s *statistic) lastErrRate() float64 {
	return float64(s.lastNIterErrCount) / float64(len(s.lastNIterRecord))
}

func (s *statistic) lastAverageCost() int64 {
	successfulCount := len(s.lastNIterRecord) - s.lastNIterErrCount
	if successfulCount <= 0 {
		return math.MaxInt64
	}
	return s.lastNIterCost / int64(successfulCount)
}

type recordWithIter struct {
	iter   uint64
	record types.Record
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

func (c *Console) resizeSpGroup() {
	for _, s := range c.statistics {
		crtSize := len(s.cost)
		targetSize := c.dataLen(s)
		if crtSize == 0 || crtSize == targetSize {
			continue
		}
		if crtSize < targetSize {
			s.cost = append(make([]int, targetSize-crtSize), s.cost...)
		} else {
			s.cost = s.cost[crtSize-targetSize:]
		}
	}
}

func (c *Console) renderSp(iter uint64) {
	for _, s := range c.statistics {
		lastRecord := s.lastRecord()
		if lastRecord == nil {
			continue
		}
		var view interface{}
		if lastRecord.record.Successful {
			view = lastRecord.record.Cost
		} else {
			view = "Err"
		}

		var flag string
		if s.dead {
			flag = "❌"
		} else {
			rate := s.lastErrRate()
			if rate < 0.01 {
				flag = "🐸"
			} else if rate < 0.1 {
				flag = "🦁"
			} else {
				flag = "🙈"
			}
			if s.lastAverageCost() < int64(5*time.Millisecond) {
				flag += " ⚡️"
			}
		}

		title := fmt.Sprintf("%s %s", flag, s.title)

		res := fmt.Sprintf("%v #%d[#%d]", view, s.total, s.errCount)

		textLen := c.dataLen(s)
		format := fmt.Sprintf("%%-%ds%%%dv", textLen/2, textLen-textLen/2-1)
		s.block.Title = fmt.Sprintf(format, title, res)
		s.block.Data = s.cost
	}
}

func (c *Console) renderErr(iter uint64) {
	display := make([]string, 0, len(c.statistics))
	for i := 0; i < c.nItem; i++ {
		s, ok := c.statistics[i]
		if !ok || s.lastErrIter == 0 {
			continue
		}
		title := fmt.Sprintf("* %s:%s", s.title, s.lastErr)
		format := "%s"
		if s.lastErrIter+errHighlightWindow >= iter {
			format = "[%s](fg-red)"
		}
		display = append(display, fmt.Sprintf(format, title))
	}
	if c.errGroup.Height < len(display) {
		c.errGroup.Height = len(display)
	}
	c.errGroup.Items = display
}

func (c *Console) render(iter uint64) {
	c.renderSp(iter)
	c.renderErr(iter)
}

func (c *Console) getStatistic(header types.RecordHeader) *statistic {
	target, ok := c.statistics[header.ID]
	if !ok {
		group, block := c.allocatedBlock(header.ID)
		target = &statistic{
			id:    header.ID,
			title: header.Target.Raw,
			total: header.Rounds,
			block: block,
			group: group,
		}
		c.statistics[header.ID] = target
	}
	return target
}

func (c *Console) handleRes(iter uint64, record types.Record) {
	var s *statistic
	s = c.getStatistic(record.RecordHeader)
	s.iter = iter
	s.total = record.Rounds
	s.lastNIterRecord = append(s.lastNIterRecord, recordWithIter{
		iter:   iter,
		record: record,
	})

	size := c.dataLen(s)
	if len(s.cost) == 0 {
		s.cost = make([]int, size)
	}
	if record.Successful {
		s.lastNIterCost += int64(record.Cost)
		s.cost = append(s.cost[1:], int(record.Cost))
	} else {
		s.errCount++
		s.lastNIterErrCount++
		s.lastErr = record.ErrMsg
		s.lastErrIter = iter
		s.cost = append(s.cost[1:], 0)
		if record.IsFatal {
			s.dead = true
		}
	}
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
func (c *Console) Run(recordChan chan types.Record, onExit func()) {
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
			if s.dead {
				continue
			}
			for i := 0; i < len(s.lastNIterRecord); i++ {
				record := s.lastNIterRecord[i]
				if record.iter+errStatisticWindow < t.Count {
					if !record.record.Successful {
						s.lastNIterErrCount--
					} else {
						s.lastNIterCost -= int64(record.record.Cost)
					}
					continue
				}
				s.lastNIterRecord = s.lastNIterRecord[i:]
				break
			}
		}
		for {
			select {
			case res := <-recordChan:
				c.handleRes(t.Count, res)
			default:
				c.render(t.Count)
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
