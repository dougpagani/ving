package net

import (
	"fmt"
	"net"
	"time"

	"github.com/yittg/ving/net/protocol"
	"github.com/yittg/ving/net/protocol/icmp"
)

// NPing network ping
type NPing struct {
	icmpPing *icmp.IPing
}

// NewPing new a ping
func NewPing(stopChan chan bool) *NPing {
	return &NPing{
		icmpPing: icmp.NewPing(stopChan),
	}
}

// Start listen
func (p *NPing) Start() (err error) {
	return p.icmpPing.Start()
}

// PingOnce to target with address as `addr`
func (p *NPing) PingOnce(target *protocol.NetworkTarget, timeout time.Duration) (time.Duration, error) {
	switch target.Typ {
	case protocol.IP:
		return p.icmpPing.Ping(target.Target.(*net.IPAddr), timeout)
	default:
		return 0, fmt.Errorf("unsupported network type, %v", target.Typ)
	}
}
