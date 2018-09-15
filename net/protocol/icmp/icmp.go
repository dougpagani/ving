package icmp

import (
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/yittg/ving/errors"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// IPing provide ability to send and receive ICMP packets
type IPing struct {
	conn   *connSource
	connV6 *connSource

	sessions sync.Map

	stopping chan bool
	stop     sync.WaitGroup
}

var protoMap = map[int]protoDesc{
	4: {1, ipv4.ICMPTypeEcho, ipv4.ICMPTypeEchoReply, ipv4.ICMPTypeTimeExceeded},
	6: {58, ipv6.ICMPTypeEchoRequest, ipv6.ICMPTypeEchoReply, ipv6.ICMPTypeTimeExceeded},
}

type protoDesc struct {
	proto  int
	reqTyp icmp.Type
	relTyp icmp.Type
	ttlTyp icmp.Type
}

type connSource struct {
	pd  protoDesc
	c   *icmp.PacketConn
	bus chan *packet
}

type packet struct {
	source *connSource

	echoAt   time.Time
	echoFrom net.Addr

	typ   icmp.Type
	bytes []byte
	n     int
}

type session struct {
	id int
	ch chan *packet
}

// NewPing new a ping
func NewPing(stopChan chan bool) *IPing {
	return &IPing{
		stopping: stopChan,
		sessions: sync.Map{},
	}
}

func newSession() *session {
	return &session{
		ch: make(chan *packet, 1),
	}
}

func (p *IPing) newConn(network string, version int) (*connSource, error) {
	c, err := icmp.ListenPacket(network, "")
	if err != nil {
		return nil, err
	}
	return &connSource{
		c:   c,
		pd:  protoMap[version],
		bus: make(chan *packet, 256),
	}, nil
}

func (p *IPing) newIPv4Conn() (*connSource, error) {
	return p.newConn(networkType["ipv4"], 4)
}

func (p *IPing) newIPv6Conn() (*connSource, error) {
	return p.newConn(networkType["ipv6"], 6)
}

// Start listen
func (p *IPing) Start() (err error) {
	p.conn, err = p.newIPv4Conn()
	if err != nil {
		return
	}
	p.connV6, err = p.newIPv6Conn()
	if err != nil {
		return
	}

	p.stop.Add(4)
	go p.consumeBus(p.conn)
	go p.readFrom(p.conn)
	go p.consumeBus(p.connV6)
	go p.readFrom(p.connV6)
	go p.wait()
	return nil
}

func (p *IPing) wait() {
	p.stop.Wait()

	p.conn.close()
	p.connV6.close()
}

func (p *IPing) readFrom(c *connSource) {
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
			n, addr, err := c.c.ReadFrom(bytes)
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
			c.bus <- &packet{
				bytes:    bytes,
				n:        n,
				echoAt:   time.Now(),
				echoFrom: addr,
				source:   c}
		}
	}
}

func (p *IPing) consumeBus(c *connSource) {
	for {
		select {
		case <-p.stopping:
			p.stop.Done()
			return
		case msg := <-c.bus:
			p.parseMsg(msg)
		}
	}
}

func (p *IPing) parseMsg(pkt *packet) {
	var m *icmp.Message
	var err error
	if m, err = icmp.ParseMessage(pkt.source.pd.proto, pkt.bytes[:pkt.n]); err != nil {
		return
	}
	pkt.typ = m.Type

	enSessionCh := func(sid int) {
		if s, ok := p.sessions.Load(sid); ok {
			s.(*session).ch <- pkt
		}
	}

	if echo, ok := m.Body.(*icmp.Echo); ok {
		enSessionCh(echo.ID)
	} else if tex, ok := m.Body.(*icmp.TimeExceeded); ok {
		if dat, err := ipv4.ParseHeader(tex.Data); err == nil {
			originPkt, _ := icmp.ParseMessage(pkt.source.pd.proto, tex.Data[dat.Len:])
			if echo, ok := originPkt.Body.(*icmp.Echo); ok {
				enSessionCh(echo.ID)
			}
		}
	}

}

func (p *IPing) send(ipAddr *net.IPAddr, c *connSource) (*time.Time, *session, error) {
	var sid int
	s := newSession()
	for {
		sid = rand.Intn(1 << 16)
		if _, loaded := p.sessions.LoadOrStore(sid, s); !loaded {
			s.id = sid
			break
		}
	}
	bytes, err := (&icmp.Message{
		Type: c.pd.reqTyp,
		Code: 0,
		Body: &icmp.Echo{
			ID:   sid,
			Seq:  0,
			Data: []byte{0, 1, 2},
		},
	}).Marshal(nil)
	if err != nil {
		return nil, nil, err
	}
	t := time.Now()
	if _, err := c.c.WriteTo(bytes, p.buildDst(ipAddr)); err != nil {
		return nil, nil, err
	}
	return &t, s, nil
}

func (p *IPing) finishSession(s *session) {
	p.sessions.Delete(s.id)
}

func (p *IPing) doPing(ipAddr *net.IPAddr, c *connSource, timeout time.Duration) (time.Duration, net.Addr, error) {
	since, session, e := p.send(ipAddr, c)
	if e != nil {
		return 0, nil, e
	}
	timer := time.NewTimer(timeout)
	defer p.finishSession(session)
	select {
	case <-timer.C:
		return 0, nil, &errors.ErrTimeout{}
	case pkt := <-session.ch:
		if pkt.typ != c.pd.relTyp {
			if pkt.typ == c.pd.ttlTyp {
				return pkt.echoAt.Sub(*since), pkt.echoFrom, &errors.ErrTTLExceed{}
			}
			return 0, nil, &errors.ErrTimeout{}
		}
		return pkt.echoAt.Sub(*since), pkt.echoFrom, nil
	}
}

// Ping ipAddr with timeout
func (p *IPing) Ping(ipAddr *net.IPAddr, timeout time.Duration) (latency time.Duration, err error) {
	if ipAddr.IP.To4() != nil {
		latency, _, err = p.doPing(ipAddr, p.conn, timeout)
	} else {
		latency, _, err = p.doPing(ipAddr, p.connV6, timeout)
	}
	return
}

// Trace ipAddr with timeout
func (p *IPing) Trace(ipAddr *net.IPAddr, ttl int, timeout time.Duration) (time.Duration, net.Addr, error) {
	var c *connSource
	var err error
	if ipAddr.IP.To4() != nil {
		c, err = p.newIPv4Conn()
		if err != nil {
			return 0, nil, err
		}
		err = c.c.IPv4PacketConn().SetTTL(ttl)
	} else {
		c, err = p.newIPv6Conn()
		if err != nil {
			return 0, nil, err
		}
		err = c.c.IPv6PacketConn().SetHopLimit(ttl)
	}
	if err != nil {
		return 0, nil, err
	}
	defer c.close()
	return p.doPing(ipAddr, c, timeout)
}

func (c *connSource) close() {
	c.c.Close()
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
