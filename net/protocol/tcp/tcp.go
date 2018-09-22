package tcp

import (
	"net"
	"time"
)

// TPing provide ability to connect to tcp port
type TPing struct {
}

// NewPing for tcp
func NewPing(stop chan bool) *TPing {
	return &TPing{}
}

// Touch a tcp addr
func (p *TPing) Touch(addr *net.TCPAddr, timeout time.Duration) (time.Duration, error) {
	dialAt := time.Now()
	conn, err := net.DialTimeout("tcp", addr.String(), timeout)
	dialDoneAt := time.Now()
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	return dialDoneAt.Sub(dialAt), nil
}
