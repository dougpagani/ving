package ui

import (
	"fmt"
	"math/rand"

	ui "github.com/gizak/termui"
)

// Run a spark line ui
func Run(targets []string, dataSource func() ([]SpItem), onExit func()) {
	if err := ui.Init(); err != nil {
		panic(err)
	}
	defer ui.Close()

	width := 80
	height := 3

	sparkLines := make([]ui.Sparkline, 0, len(targets))
	color := rand.Intn(ui.NumberofColors - 2)
	for i, target := range targets {
		sp := ui.Sparkline{}
		sp.Height = height
		sp.Title = target
		sp.Data = make([]int, width-2)
		sp.LineColor = ui.Attribute((color+i)%(ui.NumberofColors-2) + 2)
		sp.TitleColor = ui.ColorWhite
		sparkLines = append(sparkLines, sp)
	}

	group := ui.NewSparklines(sparkLines...)
	group.Width = width
	group.Height = len(sparkLines)*(height+1) + 2
	group.BorderLabel = "Ping"

	draw := func(e ui.Event) {
		for idx, d := range dataSource() {
			line := &(group.Lines[idx])
			line.Data = append(line.Data[1:], d.Value)
			format := fmt.Sprintf("%%s%%%dv", width-2-len(targets[idx]))
			line.Title = fmt.Sprintf(format, targets[idx], d.Display)
		}
		ui.Render(group)
	}

	stop := func() {
		onExit()
		ui.StopLoop()
	}

	ui.Handle("/sys/kbd/q", func(ui.Event) {
		stop()
	})
	ui.Handle("/sys/kbd/C-c", func(ui.Event) {
		stop()
	})
	ui.Handle("/timer/1s", func(e ui.Event) {
		draw(e)
	})
	ui.Loop()
}
