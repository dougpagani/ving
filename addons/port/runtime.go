package port

import (
	"strconv"
	"sync"
	"time"

	"github.com/yittg/ving/addons"
	"github.com/yittg/ving/addons/port/types"
	"github.com/yittg/ving/net"
	"github.com/yittg/ving/net/protocol"
	"github.com/yittg/ving/options"
)

type runtime struct {
	targets []*protocol.NetworkTarget
	stop    chan bool
	ping    *net.NPing
	opt     *options.Option
	active  bool

	selected    chan int
	resultChan  chan *touchResult
	refreshChan chan int

	targetPorts []types.PortDesc
	targetDone  sync.Map
	results     map[int][]touchResultWrapper
}

type touchResult struct {
	id        int
	portID    int
	connected bool
	connTime  time.Duration
}

type touchResultWrapper struct {
	port types.PortDesc
	res  *touchResult
}

// NewPortAddOn new port add-on
func NewPortAddOn() addons.AddOn {
	return &runtime{
		selected:    make(chan int, 1),
		resultChan:  make(chan *touchResult, 1024),
		targetPorts: getPredefinedPorts(),
		targetDone:  sync.Map{},
		results:     make(map[int][]touchResultWrapper),
		refreshChan: make(chan int, 1),
	}
}

// Init ports scanner
func (rt *runtime) Init(targets []*protocol.NetworkTarget, stop chan bool, opt *options.Option, ping *net.NPing) {
	rt.targets = targets
	rt.stop = stop
	rt.opt = opt
	rt.ping = ping

	if len(opt.MorePorts) > 0 {
		customPorts := make([]types.PortDesc, 0, len(opt.MorePorts))
		for _, p := range opt.MorePorts {
			customPorts = append(customPorts, types.PortDesc{Name: strconv.Itoa(p), Port: p})
		}
		rt.targetPorts = append(customPorts, rt.targetPorts...)
		rt.opt.Ports = true
	}
}

// Start ports scanner
func (rt *runtime) Start() {
	go rt.scanPorts()
}

func (rt *runtime) scanPorts() {
	var selected int
	var host *protocol.NetworkTarget
	ticker := time.NewTicker(time.Millisecond * 10)
	for {
		select {
		case <-rt.stop:
			return
		case selected = <-rt.selected:
			if selected < 0 || selected >= len(rt.targets) {
				host = nil
				continue
			}
			host = rt.targets[selected]
		case id := <-rt.refreshChan:
			rt.selected <- id
			rt.targetDone.Store(id, false)
			rt.results[id] = rt.prepareTouchResults()
		case <-ticker.C:
			if !rt.active || host == nil || rt.checkDone(selected) {
				break
			}
			g := sync.WaitGroup{}
			for i, port := range rt.targetPorts {
				g.Add(1)
				go func(idx int, p types.PortDesc) {
					defer g.Done()
					connTime, err := rt.ping.PingOnce(protocol.TCPTarget(host, p.Port), time.Second)
					rt.resultChan <- &touchResult{
						id:        selected,
						portID:    idx,
						connected: err == nil,
						connTime:  connTime,
					}
				}(i, port)
			}
			g.Wait()
			host = nil
		}
	}
}

func (rt *runtime) resetTargetIter(id int) {
	rt.refreshChan <- id
}

func (rt *runtime) prepareTouchResults() []touchResultWrapper {
	s := make([]touchResultWrapper, len(rt.targetPorts))
	for i, port := range rt.targetPorts {
		s[i] = touchResultWrapper{
			port: port,
		}
	}
	return s
}

// Collect scan results
func (rt *runtime) Collect() {
	updated := make(map[int]bool)
	for {
		select {
		case res := <-rt.resultChan:
			s, ok := rt.results[res.id]
			if !ok {
				s = rt.prepareTouchResults()
				rt.results[res.id] = s
			}
			s[res.portID].res = res
			updated[res.id] = true
		default:
			for id := range updated {
				done := true
				for _, s := range rt.results[id] {
					if s.res == nil {
						done = false
						break
					}
				}
				rt.targetDone.Store(id, done)
			}
			return
		}
	}
}

// Activate ports scanner add-on
func (rt *runtime) Activate() {
	rt.active = true
}

// Deactivate ports scanner add-on
func (rt *runtime) Deactivate() {
	rt.active = false
}

// RenderState return state to render
func (rt *runtime) RenderState() interface{} {
	return rt.results
}

// NewUI init a ui for this add-on
func (rt *runtime) NewUI() addons.UI {
	return &ui{
		selectChan: rt.selected,
		start:      rt.opt.Ports,
		source:     rt,
	}
}

func (rt *runtime) checkDone(idx int) bool {
	done, ok := rt.targetDone.Load(idx)
	return ok && done.(bool)
}
