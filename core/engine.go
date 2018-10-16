package core

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/yittg/ving/addons"
	"github.com/yittg/ving/errors"
	"github.com/yittg/ving/net"
	"github.com/yittg/ving/net/protocol"
	"github.com/yittg/ving/options"
	"github.com/yittg/ving/statistic"
	"github.com/yittg/ving/types"
	"github.com/yittg/ving/ui"
)

const (
	defaultLoopPeriodic = time.Millisecond * 10
)

// Engine of this utility
type Engine struct {
	opt *options.Option

	targets []*protocol.NetworkTarget

	ping *net.NPing

	statistic map[int]*statistic.Detail
	stSlice   []*statistic.Detail
	records   chan types.Record

	console *ui.Console

	addOns []addons.AddOn

	stop chan bool
}

// NewEngine new a engine instance
func NewEngine(opt *options.Option, targets []string) (*Engine, error) {
	networkTargets := make([]*protocol.NetworkTarget, 0, len(targets))
	for _, t := range targets {
		networkTargets = append(networkTargets, protocol.ResolveTarget(t))
	}
	if opt.Gateway {
		networkTargets = append(networkTargets, protocol.DiscoverGatewayTarget())
	}
	if len(networkTargets) == 0 {
		networkTargets = append(networkTargets, protocol.ResolveTarget("localhost"))
	}
	nTargets := len(networkTargets)

	stop := make(chan bool, 2)
	records := make(chan types.Record, nTargets)
	nPing := net.NewPing(stop)

	addOns := addons.All
	var addOnUIs []addons.UI
	envoy := &addons.Envoy{
		Targets: networkTargets,
		Opt:     opt,
		Ping:    nPing,
	}
	for _, addOn := range addOns {
		addOn.Init(envoy)
		addOnUIs = append(addOnUIs, addOn.GetUI())
	}
	return &Engine{
		opt:       opt,
		targets:   networkTargets,
		ping:      nPing,
		statistic: make(map[int]*statistic.Detail, nTargets),
		stSlice:   make([]*statistic.Detail, 0, nTargets),
		records:   records,

		addOns:  addOns,
		console: ui.NewConsole(nTargets, addOnUIs),

		stop: stop,
	}, nil
}

// Run the engine
func (e *Engine) Run() {
	if err := e.ping.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "start ping error, %v\n", err)
		os.Exit(2)
	}

	for idx, target := range e.targets {
		header := types.RecordHeader{
			ID:     idx,
			Target: target,
		}
		go e.pingTarget(header)
	}
	go e.loop()
	for _, addOn := range e.addOns {
		addOn.Start()
		defer addOn.Stop()
	}
	e.console.Run(e.stop)
}

func (e *Engine) pingTarget(header types.RecordHeader) {
	if header.Target.Typ == protocol.Unknown {
		e.records <- types.Record{
			RecordHeader: header,
			Successful:   false,
			ErrMsg:       header.Target.Target.(error).Error(),
			IsFatal:      true,
		}
		return
	}
	t := time.NewTicker(e.opt.Interval)

	f := func() bool {
		duration, err := e.ping.PingOnce(header.Target, e.opt.Timeout)
		header.Rounds++
		if err != nil {
			_, isTimeout := err.(*errors.ErrTimeout)
			e.records <- types.Record{
				RecordHeader: header,
				Successful:   false,
				ErrMsg:       err.Error(),
				IsFatal:      !isTimeout,
			}
			if !isTimeout {
				return true
			}
		} else {
			e.records <- types.Record{
				RecordHeader: header,
				Successful:   true,
				Cost:         duration,
			}
		}
		return false
	}

	if f() {
		return
	}

	for {
		select {
		case <-e.stop:
			return
		case <-t.C:
			if f() {
				return
			}
		}
	}
}

func (e *Engine) retireRecords(t time.Time) {
	for _, st := range e.statistic {
		if st.Dead {
			continue
		}
		st.RetireRecord(t)
	}
}

func (e *Engine) getStatistic(header types.RecordHeader) *statistic.Detail {
	target, ok := e.statistic[header.ID]
	if !ok {
		target = &statistic.Detail{
			ID:    header.ID,
			Title: header.Target.Raw,
			Total: header.Rounds,
			Cost:  make([]int, 1),
		}
		e.statistic[header.ID] = target
		e.stSlice = append(e.stSlice, target)
	}
	return target
}

func (e *Engine) sortedStatistic() {
	sort.Sort(statistic.StSlice{
		Details:      e.stSlice,
		SortStrategy: statistic.Default,
	})
}

func (e *Engine) loop() {
	ticker := time.NewTicker(defaultLoopPeriodic)
	lastSort := time.Now()
	for t := range ticker.C {
		func() {
			e.retireRecords(t)
			for _, addOn := range e.addOns {
				addOn.Schedule()
			}
			for {
				select {
				case res := <-e.records:
					st := e.getStatistic(res.RecordHeader)
					st.DealRecord(t, res)
				default:
					if e.opt.Sort && lastSort.Add(5*time.Second).Before(t) {
						e.sortedStatistic()
						lastSort = t
					}

					e.console.Render(t, e.stSlice)
					return
				}
			}
		}()
	}
}
