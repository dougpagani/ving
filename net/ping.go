package net

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/yittg/ving/net/protocol"
	"github.com/yittg/ving/net/protocol/icmp"
	"github.com/yittg/ving/net/protocol/tcp"
)

// NPing network ping
type NPing struct {
	icmpPing *icmp.IPing
	tcpPing  *tcp.TPing
}

// NewPing new a ping
func NewPing() *NPing {
	return &NPing{
		icmpPing: icmp.NewPing(),
		tcpPing:  tcp.NewPing(),
	}
}

// Start listen
func (p *NPing) Start(ctx context.Context) (err error) {
	return p.icmpPing.Start(ctx)
}

// PingOnce to target with address as `addr`
func (p *NPing) PingOnce(target *protocol.NetworkTarget, timeout time.Duration) (time.Duration, error) {
	switch target.Typ {
	case protocol.IP:
		return p.icmpPing.Ping(target.Target.(*net.IPAddr), timeout)
	case protocol.TCP:
		return p.tcpPing.Touch(target.Target.(*net.TCPAddr), timeout)
	default:
		return 0, fmt.Errorf("unsupported network type, %v", target.Typ)
	}
}

// Trace to target with address as `addr`
func (p *NPing) Trace(target *protocol.NetworkTarget, ttl int, timeout time.Duration) (time.Duration, net.Addr, error) {
	if target.Typ != protocol.IP {
		return 0, nil, fmt.Errorf("unsupported network type, %v", target.Typ)
	}
	return p.icmpPing.Trace(target.Target.(*net.IPAddr), ttl, timeout)
}
