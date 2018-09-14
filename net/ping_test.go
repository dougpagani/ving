package net

import (
	"log"
	"sync"
	"time"

	"github.com/yittg/ving/net/protocol"
)

func Example() {
	stopChan := make(chan bool, 2)
	p := NewPing(stopChan)
	if err := p.Start(); err != nil {
		log.Fatalf("start nping failed, %v", err)
	}

	targets := []string{"127.0.0.1", "example.com"}

	networkTargets := make([]*protocol.NetworkTarget, 0, len(targets))
	for _, t := range targets {
		networkTargets = append(networkTargets, protocol.ResolveTarget(t))
	}
	wg := sync.WaitGroup{}
	for _, target := range networkTargets {
		wg.Add(1)
		go func(t *protocol.NetworkTarget) {
			defer wg.Done()
			rtt, err := p.PingOnce(t, time.Second)
			if err != nil {
				log.Printf("ping %s error, %v\n", t.Raw, err)
			} else {
				log.Printf("ping %s rtt = %v\n", t.Raw, rtt)
			}
		}(target)
	}
	wg.Wait()

	// Output:
}
