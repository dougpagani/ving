package tcp

import (
	"log"
	"net"
	"time"
)

func Example() {
	ping := NewPing(nil)

	addrHTTP, _ := net.ResolveTCPAddr("tcp", "example.com:80")
	addrHTTPS, _ := net.ResolveTCPAddr("tcp", "example.com:443")

	addrArr := []*net.TCPAddr{addrHTTP, addrHTTPS}
	for _, addr := range addrArr {
		duration, err := ping.Touch(addr, time.Second*2)
		if err != nil {
			log.Fatalf("touch %s error, %s", addr.String(), err)
		}
		log.Printf("touch %s in %+v", addr.String(), duration)
	}

	// Output:
}
