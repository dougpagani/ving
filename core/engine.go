package core

import (
	"fmt"
	"os"
	"sort"
	"time"

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

	console *ui.Console

	records chan types.Record

	traceSelected chan int
	traceManually chan bool
	traceRecords  chan types.Record
	traceResult   *statistic.TraceSt

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
	nTargets := len(networkTargets)
	if nTargets == 0 {
		return nil, &errors.ErrNoTarget{}
	}

	stop := make(chan bool, 2)
	records := make(chan types.Record, nTargets)

	return &Engine{
		opt:       opt,
		targets:   networkTargets,
		ping:      net.NewPing(stop),
		statistic: make(map[int]*statistic.Detail, nTargets),
		stSlice:   make([]*statistic.Detail, 0, nTargets),
		console:   ui.NewConsole(nTargets),
		records:   records,

		traceSelected: make(chan int, 1),
		traceManually: make(chan bool, 1),
		traceRecords:  make(chan types.Record, 10),

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
	go e.traceTarget()
	if e.opt.Trace {
		e.traceSelected <- 0
	}
	e.console.Run(e.stop, ui.EventHandler{
		Key:          "t",
		EmitAfterRun: e.opt.Trace,
		Handler: func() {
			e.console.ToggleTrace(time.Now(), e.traceSelected, e.traceManually)
		},
	})
}

func (e *Engine) traceTarget() {
	ticker := time.NewTicker(time.Millisecond * 500)
	var header *types.RecordHeader
	ttl := 1
	gap := 0 // display the final state for gap * ticker
	manually := false
	for {
		select {
		case <-e.stop:
			return
		case selected := <-e.traceSelected:
			header = &types.RecordHeader{
				ID:     selected,
				Target: e.targets[selected],
			}
			ttl = 1
		case manually = <-e.traceManually:
			if !manually {
				break
			}
			ttl = e.doTraceTarget(header, ttl)
		case <-ticker.C:
			if manually {
				gap = 0
				break
			}
			if !e.console.TraceOn() {
				header = nil
				e.traceResult = nil
			}
			if gap > 0 {
				gap--
				break
			}
			if e.console.TraceOn() && header != nil {
				ttl = e.doTraceTarget(header, ttl)
				if ttl == 1 {
					gap = 4
				}
			}
		}
	}
}

func (e *Engine) doTraceTarget(header *types.RecordHeader, ttl int) int {
	latency, from, err := e.ping.Trace(header.Target, ttl, 2*time.Second)
	if err != nil {
		if _, ok := err.(*errors.ErrTTLExceed); ok {
			e.traceRecords <- types.Record{
				RecordHeader: *header,
				Successful:   true,
				Cost:         latency,
				From:         from,
				IsTarget:     false,
				TTL:          ttl,
			}
			return ttl + 1
		}
		e.traceRecords <- types.Record{
			RecordHeader: *header,
			Successful:   false,
			TTL:          ttl,
			ErrMsg:       err.Error(),
		}
		return 1
	}
	e.traceRecords <- types.Record{
		RecordHeader: *header,
		Successful:   true,
		Cost:         latency,
		From:         from,
		IsTarget:     true,
		TTL:          ttl,
	}
	return 1
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
	for {
		select {
		case <-e.stop:
			return
		case <-t.C:
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
					return
				}
			} else {
				e.records <- types.Record{
					RecordHeader: header,
					Successful:   true,
					Cost:         duration,
				}
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
	lastSort := int64(-1)
	for t := range ticker.C {
		if lastSort < 0 {
			lastSort = t.Unix()
		}
		func() {
			e.retireRecords(t)
			for {
				select {
				case res := <-e.records:
					st := e.getStatistic(res.RecordHeader)
					st.DealRecord(t, res)
				case res := <-e.traceRecords:
					if e.traceResult == nil || e.traceResult.ID != res.ID {
						e.traceResult = &statistic.TraceSt{ID: res.ID}
					}
					e.traceResult.DealRecord(t, res)
				default:
					if e.opt.Sort && t.Unix()-lastSort >= 5 {
						e.sortedStatistic()
						lastSort = t.Unix()
					}
					e.console.Render(t, e.stSlice, e.traceResult)
					return
				}
			}
		}()
	}
}
