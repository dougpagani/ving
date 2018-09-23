package port

import (
	"strconv"
	"time"

	"github.com/yittg/ving/addons"
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

	selected   chan int
	resultChan chan *touchResult

	targetPorts []port
	targetIter  map[int]int
	results     map[int][]touchResultWrapper
}

type touchResult struct {
	id        int
	portID    int
	connected bool
	connTime  time.Duration
}

type touchResultWrapper struct {
	port port
	res  *touchResult
}

// NewPortAddOn new port add-on
func NewPortAddOn() addons.AddOn {
	return &runtime{
		selected:    make(chan int, 1),
		resultChan:  make(chan *touchResult, 1),
		targetPorts: knownPorts,
		targetIter:  make(map[int]int),
		results:     make(map[int][]touchResultWrapper),
	}
}

func (rt *runtime) Init(targets []*protocol.NetworkTarget, stop chan bool, opt *options.Option, ping *net.NPing) {
	rt.targets = targets
	rt.stop = stop
	rt.opt = opt
	rt.ping = ping

	if len(opt.MorePorts) > 0 {
		customPorts := make([]port, 0, len(opt.MorePorts))
		for _, p := range opt.MorePorts {
			customPorts = append(customPorts, port{strconv.Itoa(p), p})
		}
		rt.targetPorts = append(customPorts, rt.targetPorts...)
		rt.opt.Ports = true
	}
}

func (rt *runtime) Start() {
	go rt.scanPorts()
	if rt.opt.Ports {
		rt.selected <- 0
	}
}

func (rt *runtime) scanPorts() {
	var selected int
	var host *protocol.NetworkTarget
	ticker := time.NewTicker(time.Millisecond * 500)
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
		case <-ticker.C:
			if !rt.active || host == nil {
				break
			}
			i, ok := rt.targetIter[selected]
			if !ok {
				i = 0
			}
			if i >= len(rt.targetPorts) {
				break
			}
			p := rt.targetPorts[i]
			connTime, err := rt.ping.PingOnce(protocol.TCPTarget(host, p.port), time.Second)
			rt.resultChan <- &touchResult{
				id:        selected,
				portID:    i,
				connected: err == nil,
				connTime:  connTime,
			}
			rt.targetIter[selected] = i + 1
		}
	}
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

func (rt *runtime) Collect() {
	for {
		select {
		case res := <-rt.resultChan:
			s, ok := rt.results[res.id]
			if !ok {
				s = rt.prepareTouchResults()
				rt.results[res.id] = s
			}
			s[res.portID].res = res
		default:
			return
		}
	}
}

func (rt *runtime) Activate() {
	rt.active = true
}

func (rt *runtime) Deactivate() {
	rt.active = false
}

func (rt *runtime) RenderState() interface{} {
	return rt.results
}

func (rt *runtime) NewUI() addons.UI {
	return &ui{
		selectChan: rt.selected,
		start:      rt.opt.Ports,
		source:     rt,
	}
}

func (rt *runtime) checkDone(idx int) bool {
	i, ok := rt.targetIter[idx]
	return ok && i >= len(rt.targetPorts)
}
