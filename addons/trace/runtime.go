package trace

import (
	"context"
	"sync"
	"time"

	"github.com/yittg/ving/addons"
	"github.com/yittg/ving/errors"
	"github.com/yittg/ving/net"
	"github.com/yittg/ving/net/protocol"
	"github.com/yittg/ving/options"
	"github.com/yittg/ving/types"
)

type runtime struct {
	targets    []*protocol.NetworkTarget
	rawTargets []string
	ping       *net.NPing
	opt        *options.Option
	active     bool

	traceSelected chan int
	traceManually chan bool
	traceRecords  chan types.Record
	traceResult   *St

	ui         *ui
	initUILock sync.Once
}

// NewTrace new trace runtime
func NewTrace() addons.AddOn {
	return &runtime{
		traceSelected: make(chan int, 1),
		traceManually: make(chan bool, 1),
		traceRecords:  make(chan types.Record, 10),
	}
}

// Desc of this trace add-on
func (*runtime) Desc() string {
	return "traceroute the target"
}

// Init see `AddOn.Init`
func (tr *runtime) Init(envoy *addons.Envoy) {
	tr.targets = envoy.Targets
	tr.opt = envoy.Opt
	tr.ping = envoy.Ping

	for _, t := range tr.targets {
		tr.rawTargets = append(tr.rawTargets, t.Raw)
	}
}

func (tr *runtime) updateStatus(active bool) {
	tr.active = active
}

// GetUI new a runtime unit instance
func (tr *runtime) GetUI() addons.UI {
	if tr.ui == nil {
		tr.initUILock.Do(func() {
			tr.ui = newUI(tr)
		})
	}
	return tr.ui
}

func (tr *runtime) Start(ctx context.Context) {
	go tr.traceTarget(ctx)
}

func (tr *runtime) traceTarget(ctx context.Context) {
	ticker := time.NewTicker(time.Millisecond * 500)
	var header *types.RecordHeader
	ttl := 1
	gap := 0 // display the final state for gap * ticker
	manually := false
	for {
		select {
		case <-ctx.Done():
			return
		case selected := <-tr.traceSelected:
			header = &types.RecordHeader{
				ID:     selected,
				Target: tr.targets[selected],
			}
			ttl = 1
		case manually = <-tr.traceManually:
			if !manually {
				break
			}
			ttl = tr.doTraceTarget(header, ttl)
		case <-ticker.C:
			if manually {
				gap = 0
				break
			}
			if gap > 0 {
				gap--
				break
			}
			if tr.active && header != nil {
				ttl = tr.doTraceTarget(header, ttl)
				if ttl == 1 {
					gap = 4
				}
			}
		}
	}
}

func (tr *runtime) doTraceTarget(header *types.RecordHeader, ttl int) int {
	latency, from, err := tr.ping.Trace(header.Target, ttl, 2*time.Second)
	if err != nil {
		if _, ok := err.(*errors.ErrTTLExceed); ok {
			tr.traceRecords <- types.Record{
				RecordHeader: *header,
				Successful:   true,
				Cost:         latency,
				From:         from,
				IsTarget:     false,
				TTL:          ttl,
			}
			return ttl + 1
		}
		tr.traceRecords <- types.Record{
			RecordHeader: *header,
			Successful:   false,
			TTL:          ttl,
			ErrMsg:       err.Error(),
		}
		return ttl + 1
	}
	tr.traceRecords <- types.Record{
		RecordHeader: *header,
		Successful:   true,
		Cost:         latency,
		From:         from,
		IsTarget:     true,
		TTL:          ttl,
	}
	return 1
}

func (tr *runtime) Schedule() {
	for {
		select {
		case res := <-tr.traceRecords:
			if tr.traceResult == nil || tr.traceResult.ID != res.ID {
				tr.traceResult = &St{ID: res.ID}
			}
			tr.traceResult.DealRecord(res)
		default:
			return
		}
	}
}

func (tr *runtime) State() interface{} {
	return tr.traceResult
}
