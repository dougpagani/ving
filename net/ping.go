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

// Ping provide ability to send and receive ICMP packets
type Ping struct {
	conn   *connSource
	connV6 *connSource

	bus chan *packet

	sessions sync.Map

	stopping chan bool
	stop     sync.WaitGroup
}

type connSource struct {
	proto int
	c     *icmp.PacketConn
}

type packet struct {
	source *connSource

	echoAt time.Time

	bytes []byte
	n     int
}

type session struct {
	id int
	ch chan time.Time
}

// NewPing new a ping
func NewPing() *Ping {
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

// Start listen
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

	p.stop.Add(1)
	go p.consumeBus()
	func() {
		p.stop.Add(2)
		go p.readFrom(p.conn)
		go p.readFrom(p.connV6)
	}()
	return nil
}

// Stop listen and receive
func (p *Ping) Stop() {
	close(p.stopping)
	p.stop.Wait()

	p.conn.close()
	p.connV6.close()
}

func (p *Ping) readFrom(c *connSource) {
	for {
		select {
		case <-p.stopping:
			p.stop.Done()
			return
		default:
			bytes := make([]byte, 512)
			if err := c.c.SetReadDeadline(time.Now().Add(time.Millisecond * 100)); err != nil {
				close(p.stopping)
				return
			}
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
				echoAt: time.Now(),
				source: c}
		}
	}
}

func (p *Ping) consumeBus() {
	for {
		select {
		case <-p.stopping:
			p.stop.Done()
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
		s.(*session).ch <- msg.echoAt
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
		sid = rand.Intn(1 << 16)
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

	t := time.Now()
	if _, err := c.c.WriteTo(bytes, dst); err != nil {
		return nil, nil, err
	}
	return &t, s, nil
}

// PingOnce to target with address as `addr`
func (p *Ping) PingOnce(target *NetworkTarget, timeout time.Duration) (time.Duration, error) {
	if target.Typ != IP {
		return 0, fmt.Errorf("unsupported network type, %v", target.Typ)
	}

	return p.doPing(target.Target.(*net.IPAddr), timeout)
}

func (p *Ping) finishSession(s *session) {
	p.sessions.Delete(s.id)
}

func (p *Ping) doPing(ipAddr *net.IPAddr, timeout time.Duration) (time.Duration, error) {
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
	timer := time.NewTimer(timeout)
	defer p.finishSession(session)
	select {
	case <-timer.C:
		return 0, fmt.Errorf("timeout")
	case pongAt := <-session.ch:
		return pongAt.Sub(*since), nil
	}
}

func (c *connSource) close() {
	c.c.Close()
}
