package net

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

type Ping struct {
	conn   *connSource
	connV6 *connSource

	bus chan *packet

	sessions sync.Map

	stopping chan bool
}

type connSource struct {
	proto int
	c     *icmp.PacketConn
}

type packet struct {
	source *connSource

	recvAt time.Time

	bytes []byte
	n     int
}

type session struct {
	id int
	ch chan time.Time
}

func NewPing() (*Ping) {
	return &Ping{
		bus:      make(chan *packet, 256),
		stopping: make(chan bool),
		sessions: sync.Map{},
	}
}

func newSession() *session {
	return &session{
		ch: make(chan time.Time),
	}
}

func (p *Ping) Start() (err error) {
	c, err := icmp.ListenPacket("udp4", "")
	if err != nil {
		return
	}
	p.conn = &connSource{c: c, proto: 1}
	c, err = icmp.ListenPacket("udp6", "")
	if err != nil {
		return
	}
	p.connV6 = &connSource{c: c, proto: 58}

	go p.consumeBus()
	go p.recv()

	return nil
}

func (p *Ping) Stop() {
	close(p.stopping)
}

func (p *Ping) recv() {
	go p.recvFrom(p.conn)

	go p.recvFrom(p.connV6)
}

func (p *Ping) recvFrom(c *connSource) {
	for {
		select {
		case <-p.stopping:
			return
		default:
			bytes := make([]byte, 512)
			c.c.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
			n, _, err := c.c.ReadFrom(bytes)
			if err != nil {
				if netOpErr, ok := err.(*net.OpError); ok {
					if netOpErr.Timeout() {
						continue
					} else {
						close(p.stopping)
						return
					}
				}
			}
			p.bus <- &packet{
				bytes:  bytes,
				n:      n,
				recvAt: time.Now(),
				source: c}
		}
	}
}

func (p *Ping) consumeBus() {
	for {
		select {
		case <-p.stopping:
			return
		case msg := <-p.bus:
			p.parseMsg(msg)
		}
	}
}

func (p *Ping) parseMsg(msg *packet) {
	var m *icmp.Message
	var err error
	if m, err = icmp.ParseMessage(msg.source.proto, msg.bytes[:msg.n]); err != nil {
		return
	}

	if m.Type != ipv4.ICMPTypeEchoReply && m.Type != ipv6.ICMPTypeEchoReply {
		return
	}

	body := m.Body.(*icmp.Echo)
	if s, ok := p.sessions.Load(body.ID); ok {
		s.(*session).ch <- msg.recvAt
	}
}

func (p *Ping) send(ipAddr *net.IPAddr, c *connSource) (*time.Time, *session, error) {
	var typ icmp.Type
	if c.proto == 1 {
		typ = ipv4.ICMPTypeEcho
	} else {
		typ = ipv6.ICMPTypeEchoRequest
	}

	var sid int
	s := newSession()
	for {
		sid = rand.Intn(1<<16)
		if _, loaded := p.sessions.LoadOrStore(sid, s); !loaded {
			s.id = sid
			break
		}
	}
	dst := &net.UDPAddr{IP: ipAddr.IP, Zone: ipAddr.Zone}
	bytes, err := (&icmp.Message{
		Type: typ,
		Code: 0,
		Body: &icmp.Echo{
			ID:  sid,
			Seq: 1,
		},
	}).Marshal(nil)
	if err != nil {
		return nil, nil, err
	}

	for {
		t := time.Now()
		if _, err := c.c.WriteTo(bytes, dst); err != nil {
			return nil, nil, err
		}
		return &t, s, nil
	}
}

func (p *Ping) PingOnce(addr string) (time.Duration, error) {
	ipAddr, err := net.ResolveIPAddr("ip", addr)
	if err != nil {
		return 0, err
	}

	return p.doPing(ipAddr)
}

func (p *Ping) finishSession(s *session) {
	p.sessions.Delete(s.id)
}

func (p *Ping) doPing(ipAddr *net.IPAddr) (time.Duration, error) {
	var c *connSource
	if ipAddr.IP.To4() != nil {
		c = p.conn
	} else {
		c = p.connV6
	}
	since, session, e := p.send(ipAddr, c)
	if e != nil {
		return 0, e
	}
	timeout := time.NewTimer(time.Second * 2)
	defer p.finishSession(session)
	select {
	case <-timeout.C:
		return 0, fmt.Errorf("timeout")
	case pongAt := <-session.ch:
		return pongAt.Sub(*since), nil
	}
}
