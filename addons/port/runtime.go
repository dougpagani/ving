package port

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/yittg/ving/addons"
	"github.com/yittg/ving/addons/port/types"
	"github.com/yittg/ving/config"
	"github.com/yittg/ving/net"
	"github.com/yittg/ving/net/protocol"
	"github.com/yittg/ving/options"
)

type runtime struct {
	targets    []*protocol.NetworkTarget
	rawTargets []string
	stop       chan bool
	ping       *net.NPing
	opt        *options.Option
	active     bool

	selected    chan int
	crtSelected int
	resultChan  chan *touchResult
	refreshChan chan int

	targetPorts []types.PortDesc
	targetDone  sync.Map
	results     map[int][]touchResultWrapper

	proberPool     sync.Map
	proberPoolSize int
	scheduling     *int32

	ui         *ui
	initUILock sync.Once
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

type prober struct {
	pipeMap sync.Map
	running sync.Once
}

type probeUnit struct {
	id     int
	portID int
	target *protocol.NetworkTarget
}

func newPortAddOn() addons.AddOn {
	portConfig := config.GetConfig().AddOns.Ports
	return &runtime{
		selected:       make(chan int, 1),
		proberPool:     sync.Map{},
		proberPoolSize: portConfig.ProbeConcurrency,
		resultChan:     make(chan *touchResult, 1024),
		targetDone:     sync.Map{},
		results:        make(map[int][]touchResultWrapper),
		refreshChan:    make(chan int, 1),
		stop:           make(chan bool, 2),
	}
}

// Desc of this port add-on
func (*runtime) Desc() string {
	return "port probe"
}

// Init ports scanner
func (rt *runtime) Init(envoy *addons.Envoy) {
	rt.targets = envoy.Targets
	rt.opt = envoy.Opt
	rt.ping = envoy.Ping
	for _, t := range rt.targets {
		rt.rawTargets = append(rt.rawTargets, t.Raw)
	}

	scheduling := int32(0)
	rt.scheduling = &scheduling
	if len(rt.opt.MorePorts) > 0 {
		for _, p := range rt.opt.MorePorts {
			rt.targetPorts = append(rt.targetPorts, types.PortDesc{Name: getNameOfPort(p), Port: p})
		}
		rt.opt.Ports = true
	} else {
		rt.targetPorts = getPredefinedPorts()
	}
}

func (rt *runtime) Start() {
	go rt.scanPorts()
}

func (rt *runtime) Stop() {
	close(rt.stop)
}

func (rt *runtime) currentSelected() int {
	return rt.crtSelected
}

func (rt *runtime) scanPorts() {
	var host *protocol.NetworkTarget
	ticker := time.NewTicker(time.Millisecond * 10)
	for {
		select {
		case <-rt.stop:
			return
		case rt.crtSelected = <-rt.selected:
			if rt.crtSelected < 0 || rt.crtSelected >= len(rt.targets) {
				host = nil
				continue
			}
			host = rt.targets[rt.crtSelected]
		case id := <-rt.refreshChan:
			rt.selected <- id
		case <-ticker.C:
			selected := rt.currentSelected()
			if !rt.active || host == nil || !rt.checkNotBegin(selected) {
				break
			}
			rt.targetDone.Store(selected, 0)
			for i, port := range rt.targetPorts {
				rt.probeTargetAsyc(selected, i, protocol.TCPTarget(host, port.Port))
			}
			host = nil
		}
	}
}

func (rt *runtime) probeTargetAsyc(idx, portID int, t *protocol.NetworkTarget) {
	bucket := portID % rt.proberPoolSize
	_p, existed := rt.proberPool.LoadOrStore(bucket, &prober{})
	p := _p.(*prober)

	chooseOrAllocatePipe := func(pipeMap *sync.Map, idx int) chan *probeUnit {
		_pipe, ok := pipeMap.Load(idx)
		if !ok {
			_pipe = make(chan *probeUnit, 100)
			pipeMap.Store(idx, _pipe)
		}
		return _pipe.(chan *probeUnit)
	}

	if !existed {
		p.running.Do(func() {
			go func(pipeMap *sync.Map) {
				for {
					select {
					case pu := <-chooseOrAllocatePipe(pipeMap, rt.currentSelected()):
						connTime, err := rt.ping.PingOnce(pu.target, time.Second)
						rt.resultChan <- &touchResult{
							id:        pu.id,
							portID:    pu.portID,
							connected: err == nil,
							connTime:  connTime,
						}
					default:
						time.Sleep(time.Millisecond * 10)
					}
				}
			}(&p.pipeMap)
		})
	}
	chooseOrAllocatePipe(&p.pipeMap, idx) <- &probeUnit{
		id:     idx,
		portID: portID,
		target: t,
	}
}

func (rt *runtime) resetTargetStatus(id int) {
	if !rt.checkDone(id) {
		return
	}

	rt.results[id] = rt.prepareTouchResults()
	rt.targetDone.Delete(id)
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

func (rt *runtime) doSchedule() {
	for {
		select {
		case res := <-rt.resultChan:
			s, ok := rt.results[res.id]
			if !ok {
				s = rt.prepareTouchResults()
				rt.results[res.id] = s
			}
			s[res.portID].res = res
			v, loaded := rt.targetDone.LoadOrStore(res.id, 1)
			if loaded {
				rt.targetDone.Store(res.id, v.(int)+1)
			}
		default:
			return
		}
	}
}

func (rt *runtime) Schedule() {
	if atomic.SwapInt32(rt.scheduling, 1) > 0 {
		return
	}
	defer atomic.StoreInt32(rt.scheduling, 0)
	rt.doSchedule()
}

func (rt *runtime) updateStatus(active bool) {
	rt.active = active
}

func (rt *runtime) State() interface{} {
	return rt.results
}

// GetUI init a ui for this add-on
func (rt *runtime) GetUI() addons.UI {
	if rt.ui == nil {
		rt.initUILock.Do(func() {
			rt.ui = newUI(rt)
		})
	}
	return rt.ui
}

func (rt *runtime) checkNotBegin(idx int) bool {
	_, ok := rt.targetDone.Load(idx)
	return !ok
}

func (rt *runtime) checkDone(idx int) bool {
	done, ok := rt.targetDone.Load(idx)
	return ok && done.(int) == len(rt.targetPorts)
}
